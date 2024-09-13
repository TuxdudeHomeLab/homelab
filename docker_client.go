package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"

	dcontainer "github.com/docker/docker/api/types/container"
	dfilters "github.com/docker/docker/api/types/filters"
	dimage "github.com/docker/docker/api/types/image"
	dnetwork "github.com/docker/docker/api/types/network"
	dclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sys/unix"

	// "github.com/docker/docker/pkg/term"
	"github.com/moby/term"
)

type dockerClient struct {
	client      *dclient.Client
	platform    string
	ociPlatform ocispec.Platform
	debug       bool
}

const (
	containerStateUnknown containerState = iota
	containerStateNotFound
	containerStateCreated
	containerStateRunning
	containerStatePaused
	containerStateRestarting
	containerStateRemoving
	containerStateExited
	containerStateDead
)

type containerState uint8

func containerStateFromString(state string) containerState {
	switch state {
	case "created":
		return containerStateCreated
	case "running":
		return containerStateRunning
	case "paused":
		return containerStatePaused
	case "restarting":
		return containerStateRestarting
	case "removing":
		return containerStateRemoving
	case "exited":
		return containerStateExited
	case "dead":
		return containerStateDead
	default:
		return containerStateUnknown
	}
}

func (c containerState) String() string {
	switch c {
	case containerStateUnknown:
		return "Unknown"
	case containerStateNotFound:
		return "NotFound"
	case containerStateCreated:
		return "Created"
	case containerStateRunning:
		return "Running"
	case containerStatePaused:
		return "Paused"
	case containerStateRestarting:
		return "Restarting"
	case containerStateRemoving:
		return "Removing"
	case containerStateExited:
		return "Exited"
	case containerStateDead:
		return "Dead"
	default:
		log.Fatalf("Invalid scenario, possibly indicating a bug in the code")
		return "invalid-container-state"
	}
}

func newDockerClient(platform, arch string) (*dockerClient, error) {
	client, err := dclient.NewClientWithOpts(dclient.FromEnv, dclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create a new docker API client, reason: %w", err)
	}
	return &dockerClient{
		client:      client,
		platform:    platform,
		ociPlatform: ocispec.Platform{Architecture: arch},
		debug:       isLogLevelDebug() || isLogLevelTrace(),
	}, nil
}

func (d *dockerClient) pullImage(ctx context.Context, imageName string) error {
	progress, err := d.client.ImagePull(ctx, imageName, dimage.PullOptions{Platform: d.platform})
	if err != nil {
		return fmt.Errorf("failed to pull the image %s, reason: %w", imageName, err)
	}
	defer progress.Close()

	// Store info about existing locally available image.
	avail, id := d.queryLocalImage(ctx, imageName)
	// Show verbose pull progress only if either in debug mode or
	// there is no existing locally available image.
	showPullProgress := d.debug || !avail

	// Perform the actual image pull.
	if showPullProgress {
		if !avail {
			log.Infof("Pulling image: %s", imageName)
		} else {
			log.Debugf("Pulling image: %s", imageName)
		}
		termFd, isTerm := term.GetFdInfo(os.Stdout)
		err = jsonmessage.DisplayJSONMessagesStream(progress, os.Stdout, termFd, isTerm, nil)
	} else {
		_, err = io.Copy(io.Discard, progress)
	}
	if err != nil {
		return fmt.Errorf("failed while pulling the image %s, reason: %w", imageName, err)
	}

	// If pull progress was already shown, no need to show the updates again.
	if showPullProgress {
		log.Debugf("Pulled image successfully: %s", imageName)
		return nil
	}

	// Otherwise, determine if the image was updated and show the updated ID
	// of the image.
	avail, newId := d.queryLocalImage(ctx, imageName)
	if !avail {
		log.Fatalf("Image is expected to be available after pull, but is unavailable possibly indicating a bug or system failure!")
	}
	if newId != id {
		log.Infof("Pulled newer version of image %s: %s", imageName, newId)
	}
	return nil
}

func (d *dockerClient) queryLocalImage(ctx context.Context, imageName string) (bool, string) {
	filter := dfilters.NewArgs()
	filter.Add("reference", imageName)
	images, err := d.client.ImageList(ctx, dimage.ListOptions{
		All:            false,
		Filters:        filter,
		SharedSize:     false,
		ContainerCount: false,
		Manifests:      false,
	})

	// Ignore errors by considering the image is not available locally in
	// case of errors.
	if err != nil || len(images) == 0 {
		return false, ""
	}

	return true, images[0].ID
}

func (d *dockerClient) createContainer(ctx context.Context, containerName string, cConfig *dcontainer.Config, hConfig *dcontainer.HostConfig, nConfig *dnetwork.NetworkingConfig) error {
	log.Debugf("Creating container %s ...", containerName)
	resp, err := d.client.ContainerCreate(ctx, cConfig, hConfig, nConfig, &d.ociPlatform, containerName)
	if err != nil {
		log.Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to create the container, reason: %w", err)
	}

	log.Debugf("Container %s created successfully - %s", containerName, resp.ID)
	if len(resp.Warnings) > 0 {
		log.Warnf("Warnings encountered while creating the container %s\n%s", containerName, prettyPrintJSON(resp.Warnings))
	}
	return nil
}

func (d *dockerClient) startContainer(ctx context.Context, containerName string) error {
	log.Debugf("Starting container %s ...", containerName)
	err := d.client.ContainerStart(ctx, containerName, dcontainer.StartOptions{})
	if err != nil {
		log.Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to start the container, reason: %w", err)
	}

	log.Debugf("Container %s started successfully", containerName)
	return nil
}

func (d *dockerClient) stopContainer(ctx context.Context, containerName string) error {
	log.Debugf("Stopping container %s ...", containerName)
	err := d.client.ContainerStop(ctx, containerName, dcontainer.StopOptions{})
	if err != nil {
		log.Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to stop the container, reason: %w", err)
	}

	log.Debugf("Container %s stopped successfully", containerName)
	return nil
}

func (d *dockerClient) killContainer(ctx context.Context, containerName string) error {
	log.Debugf("Killing container %s ...", containerName)
	err := d.client.ContainerKill(ctx, containerName, unix.SignalName(unix.SIGKILL))
	if err != nil {
		log.Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to kill the container, reason: %w", err)
	}

	log.Debugf("Container %s killed successfully", containerName)
	return nil
}

func (d *dockerClient) removeContainer(ctx context.Context, containerName string) error {
	log.Debugf("Removing container %s ...", containerName)
	err := d.client.ContainerRemove(ctx, containerName, dcontainer.RemoveOptions{Force: false})
	if err != nil {
		log.Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to remove the container, reason: %w", err)
	}

	log.Debugf("Container %s removed successfully", containerName)
	return nil
}

func (d *dockerClient) getContainerState(ctx context.Context, containerName string) (containerState, error) {
	c, err := d.client.ContainerInspect(ctx, containerName)
	if dclient.IsErrNotFound(err) {
		return containerStateNotFound, nil
	}
	if err != nil {
		return containerStateUnknown, fmt.Errorf("failed to retrieve the container state, reason: %w", err)
	}
	return containerStateFromString(c.State.Status), nil
}

func (d *dockerClient) createNetwork(ctx context.Context, n *network) error {
	// TODO: Implement this.
	return nil
}

// TODO: Remove this after this function is used.
// nolint (unused)
func (d *dockerClient) deleteNetwork(ctx context.Context, networkName string) error {
	// TODO: Implement this.
	return nil
}

func (d *dockerClient) networkExists(ctx context.Context, networkName string) bool {
	filter := dfilters.NewArgs()
	filter.Add("name", networkName)
	networks, err := d.client.NetworkList(ctx, dnetwork.ListOptions{
		Filters: filter,
	})

	// Ignore errors by considering the network is not present in case of
	// errors.
	return err == nil && len(networks) > 0
}

func (d *dockerClient) connectContainerToBridgeModeNetwork(ctx context.Context, containerName, networkName, ip string) error {
	// TODO: Implement this.
	return nil
}

// TODO: Remove this after this function is used.
// nolint (unused)
func (d *dockerClient) disconnectContainerFromNetwork(ctx context.Context, containerName, networkName string) error {
	// TODO: Implement this.
	return nil
}

func (d *dockerClient) close() {
	d.client.Close()
}
