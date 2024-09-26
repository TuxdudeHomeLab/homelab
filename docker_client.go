package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	dtypes "github.com/docker/docker/api/types"
	dcontainer "github.com/docker/docker/api/types/container"
	dfilters "github.com/docker/docker/api/types/filters"
	dimage "github.com/docker/docker/api/types/image"
	dnetwork "github.com/docker/docker/api/types/network"
	dclient "github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"golang.org/x/sys/unix"

	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"
)

type dockerClient struct {
	client      dockerAPIClient
	platform    string
	ociPlatform ocispec.Platform
	debug       bool
}

type dockerAPIClient interface {
	Close() error

	ContainerCreate(ctx context.Context, config *dcontainer.Config, hostConfig *dcontainer.HostConfig, networkingConfig *dnetwork.NetworkingConfig, platform *ocispec.Platform, containerName string) (dcontainer.CreateResponse, error)
	ContainerInspect(ctx context.Context, containerName string) (dtypes.ContainerJSON, error)
	ContainerKill(ctx context.Context, containerName, signal string) error
	ContainerRemove(ctx context.Context, containerName string, options dcontainer.RemoveOptions) error
	ContainerStart(ctx context.Context, containerName string, options dcontainer.StartOptions) error
	ContainerStop(ctx context.Context, containerName string, options dcontainer.StopOptions) error

	ImageList(ctx context.Context, options dimage.ListOptions) ([]dimage.Summary, error)
	ImagePull(ctx context.Context, refStr string, options dimage.PullOptions) (io.ReadCloser, error)

	NetworkConnect(ctx context.Context, networkName, containerName string, config *dnetwork.EndpointSettings) error
	NetworkCreate(ctx context.Context, networkName string, options dnetwork.CreateOptions) (dnetwork.CreateResponse, error)
	NetworkDisconnect(ctx context.Context, networkName, containerName string, force bool) error
	NetworkList(ctx context.Context, options dnetwork.ListOptions) ([]dnetwork.Summary, error)
	NetworkRemove(ctx context.Context, networkName string) error
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

func restartPolicyModeFromString(pol string) (dcontainer.RestartPolicyMode, error) {
	switch pol {
	case "", "no":
		return dcontainer.RestartPolicyDisabled, nil
	case "always":
		return dcontainer.RestartPolicyAlways, nil
	case "on-failure":
		return dcontainer.RestartPolicyOnFailure, nil
	case "unless-stopped":
		return dcontainer.RestartPolicyUnlessStopped, nil
	default:
		return "", fmt.Errorf("invalid restart policy mode string: %s", pol)
	}
}

func restartPolicyModeValidValues() string {
	return "[ 'no', 'always', 'on-failure', 'unless-stopped' ]"
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
		panic("Invalid scenario in containerState stringer, possibly indicating a bug in the code")
	}
}

func buildDockerAPIClient(ctx context.Context) (dockerAPIClient, error) {
	if client, found := dockerAPIClientFromContext(ctx); found {
		return client, nil
	}
	return dclient.NewClientWithOpts(dclient.FromEnv, dclient.WithAPIVersionNegotiation())
}

func newDockerClient(ctx context.Context, platform, arch string) (*dockerClient, error) {
	client, err := buildDockerAPIClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new docker API client, reason: %w", err)
	}
	lvl := homelabInspectLevelFromContext(ctx)
	return &dockerClient{
		client:      client,
		platform:    platform,
		ociPlatform: ocispec.Platform{Architecture: arch},
		debug:       lvl == homelabInspectLevelDebug || lvl == homelabInspectLevelTrace,
	}, nil
}

func (d *dockerClient) pullImage(ctx context.Context, imageName string) error {
	// Store info about existing locally available image.
	avail, id := d.queryLocalImage(ctx, imageName)
	// Show verbose pull progress only if either in debug mode or
	// there is no existing locally available image.
	showPullProgress := d.debug || !avail

	progress, err := d.client.ImagePull(ctx, imageName, dimage.PullOptions{Platform: d.platform})
	if err != nil {
		return fmt.Errorf("failed to pull the image %s, reason: %w", imageName, err)
	}
	defer progress.Close()

	// Perform the actual image pull.
	if showPullProgress {
		if !avail {
			log(ctx).Infof("Pulling image: %s", imageName)
		} else {
			log(ctx).Debugf("Pulling image: %s", imageName)
		}
		termFd, isTerm := term.GetFdInfo(os.Stdout)
		err = jsonmessage.DisplayJSONMessagesStream(progress, os.Stdout, termFd, isTerm, nil)
	} else {
		_, err = io.Copy(io.Discard, progress)
	}
	if err != nil {
		return fmt.Errorf("failed while pulling the image %s, reason: %w", imageName, err)
	}

	if showPullProgress {
		log(ctx).Debugf("Image pull for %s complete", imageName)
	}

	// Otherwise, determine if the image was updated and show the updated ID
	// of the image.
	avail, newId := d.queryLocalImage(ctx, imageName)
	if !avail {
		return fmt.Errorf("image %s not available locally after a successful pull, possibly indicating a bug or a system failure!", imageName)
	}

	// If pull progress was already shown, no need to show the updates again.
	if showPullProgress {
		log(ctx).Debugf("Pulled image successfully: %s", imageName)
		return nil
	}

	if newId != id {
		log(ctx).Infof("Pulled newer version of image %s: %s", imageName, newId)
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
	log(ctx).Debugf("Creating container %s ...", containerName)
	resp, err := d.client.ContainerCreate(ctx, cConfig, hConfig, nConfig, &d.ociPlatform, containerName)
	if err != nil {
		log(ctx).Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to create the container, reason: %w", err)
	}

	log(ctx).Debugf("Container %s created successfully - %s", containerName, resp.ID)
	if len(resp.Warnings) > 0 {
		var sb strings.Builder
		for i, w := range resp.Warnings {
			sb.WriteString(fmt.Sprintf("\n%d - %s", i+1, w))
		}
		log(ctx).Warnf("Warnings encountered while creating the container %s%s", containerName, sb.String())
	}
	return nil
}

func (d *dockerClient) startContainer(ctx context.Context, containerName string) error {
	log(ctx).Debugf("Starting container %s ...", containerName)
	err := d.client.ContainerStart(ctx, containerName, dcontainer.StartOptions{})
	if err != nil {
		log(ctx).Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to start the container, reason: %w", err)
	}

	log(ctx).Debugf("Container %s started successfully", containerName)
	return nil
}

func (d *dockerClient) stopContainer(ctx context.Context, containerName string) error {
	log(ctx).Debugf("Stopping container %s ...", containerName)
	err := d.client.ContainerStop(ctx, containerName, dcontainer.StopOptions{})
	if err != nil {
		log(ctx).Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to stop the container, reason: %w", err)
	}

	log(ctx).Debugf("Container %s stopped successfully", containerName)
	return nil
}

func (d *dockerClient) killContainer(ctx context.Context, containerName string) error {
	log(ctx).Debugf("Killing container %s ...", containerName)
	err := d.client.ContainerKill(ctx, containerName, unix.SignalName(unix.SIGKILL))
	if err != nil {
		log(ctx).Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to kill the container, reason: %w", err)
	}

	log(ctx).Debugf("Container %s killed successfully", containerName)
	return nil
}

func (d *dockerClient) removeContainer(ctx context.Context, containerName string) error {
	log(ctx).Debugf("Removing container %s ...", containerName)
	err := d.client.ContainerRemove(ctx, containerName, dcontainer.RemoveOptions{Force: false})
	if err != nil {
		log(ctx).Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to remove the container, reason: %w", err)
	}

	log(ctx).Debugf("Container %s removed successfully", containerName)
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

func (d *dockerClient) createNetwork(ctx context.Context, networkName string, options dnetwork.CreateOptions) error {
	log(ctx).Debugf("Creating network %s ...", networkName)
	resp, err := d.client.NetworkCreate(ctx, networkName, options)

	if err != nil {
		log(ctx).Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to create the network, reason: %w", err)
	}

	log(ctx).Debugf("Network %s created successfully - %s", networkName, resp.ID)
	if len(resp.Warning) > 0 {
		log(ctx).Warnf("Warning encountered while creating the network %s\n%s", networkName, resp.Warning)
	}
	return nil
}

// TODO: Remove this after this function is used.
// nolint (unused)
func (d *dockerClient) removeNetwork(ctx context.Context, networkName string) error {
	log(ctx).Debugf("Removing network %s ...", networkName)
	err := d.client.NetworkRemove(ctx, networkName)
	if err != nil {
		log(ctx).Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to remove the network, reason: %w", err)
	}

	log(ctx).Debugf("Network %s removed successfully", networkName)
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
	log(ctx).Debugf("Connecting container %s to network %s with IP %s ...", containerName, networkName, ip)
	err := d.client.NetworkConnect(ctx, networkName, containerName, &dnetwork.EndpointSettings{
		IPAMConfig: &dnetwork.EndpointIPAMConfig{
			IPv4Address: ip,
		},
	})
	if err != nil {
		log(ctx).Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to connect container %s to network %s, reason: %w", containerName, networkName, err)
	}

	log(ctx).Debugf("Container %s connected to network %s successfully", containerName, networkName)
	return nil
}

// TODO: Remove this after this function is used.
// nolint (unused)
func (d *dockerClient) disconnectContainerFromNetwork(ctx context.Context, containerName, networkName string) error {
	log(ctx).Debugf("Disconnecting container %s from network %s ...", containerName, networkName)
	err := d.client.NetworkDisconnect(ctx, networkName, containerName, false)
	if err != nil {
		log(ctx).Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to disconnect container %s from network %s, reason: %w", containerName, networkName, err)
	}

	log(ctx).Debugf("Container %s disconnected from network %s successfully", containerName, networkName)
	return nil
}

func (d *dockerClient) close() {
	d.client.Close()
}
