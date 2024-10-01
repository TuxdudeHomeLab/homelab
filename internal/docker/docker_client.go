package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	dcontainer "github.com/docker/docker/api/types/container"
	dfilters "github.com/docker/docker/api/types/filters"
	dimage "github.com/docker/docker/api/types/image"
	dnetwork "github.com/docker/docker/api/types/network"
	dclient "github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/tuxdudehomelab/homelab/internal/host"
	"github.com/tuxdudehomelab/homelab/internal/inspect"

	"golang.org/x/sys/unix"

	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"
)

const (
	defaultContainerStopAndRemoveKillDelay = 1 * time.Second
)

type DockerClient struct {
	client                          DockerAPIClient
	platform                        string
	ociPlatform                     ocispec.Platform
	containerStopAndRemoveKillDelay time.Duration
	debug                           bool
}

func NewDockerClient(ctx context.Context) *DockerClient {
	h := host.MustHostInfo(ctx)
	return &DockerClient{
		client:                          MustDockerAPIClient(ctx),
		platform:                        h.DockerPlatform,
		ociPlatform:                     ocispec.Platform{Architecture: h.Arch},
		containerStopAndRemoveKillDelay: evalContainerStopAndRemoveKillDelay(ctx),
		debug:                           dockerDebugFromInspect(ctx),
	}
}

func (d *DockerClient) PullImage(ctx context.Context, imageName string) error {
	// Store info about existing locally available image.
	avail, id := d.QueryLocalImage(ctx, imageName)
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
	avail, newId := d.QueryLocalImage(ctx, imageName)
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

func (d *DockerClient) QueryLocalImage(ctx context.Context, imageName string) (bool, string) {
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

func (d *DockerClient) CreateContainer(ctx context.Context, containerName string, cConfig *dcontainer.Config, hConfig *dcontainer.HostConfig, nConfig *dnetwork.NetworkingConfig) error {
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

func (d *DockerClient) StartContainer(ctx context.Context, containerName string) error {
	log(ctx).Debugf("Starting container %s ...", containerName)
	err := d.client.ContainerStart(ctx, containerName, dcontainer.StartOptions{})
	if err != nil {
		log(ctx).Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to start the container, reason: %w", err)
	}

	log(ctx).Debugf("Container %s started successfully", containerName)
	return nil
}

func (d *DockerClient) StopContainer(ctx context.Context, containerName string) error {
	log(ctx).Debugf("Stopping container %s ...", containerName)
	err := d.client.ContainerStop(ctx, containerName, dcontainer.StopOptions{})
	if err != nil {
		log(ctx).Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to stop the container, reason: %w", err)
	}

	log(ctx).Debugf("Container %s stopped successfully", containerName)
	return nil
}

func (d *DockerClient) KillContainer(ctx context.Context, containerName string) error {
	log(ctx).Debugf("Killing container %s ...", containerName)
	err := d.client.ContainerKill(ctx, containerName, unix.SignalName(unix.SIGKILL))
	if err != nil {
		log(ctx).Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to kill the container, reason: %w", err)
	}

	log(ctx).Debugf("Container %s killed successfully", containerName)
	return nil
}

func (d *DockerClient) RemoveContainer(ctx context.Context, containerName string) error {
	log(ctx).Debugf("Removing container %s ...", containerName)
	err := d.client.ContainerRemove(ctx, containerName, dcontainer.RemoveOptions{Force: false})
	if err != nil {
		log(ctx).Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to remove the container, reason: %w", err)
	}

	log(ctx).Debugf("Container %s removed successfully", containerName)
	return nil
}

func (d *DockerClient) GetContainerState(ctx context.Context, containerName string) (ContainerState, error) {
	c, err := d.client.ContainerInspect(ctx, containerName)
	if dclient.IsErrNotFound(err) {
		return ContainerStateNotFound, nil
	}
	if err != nil {
		return ContainerStateUnknown, fmt.Errorf("failed to retrieve the container state, reason: %w", err)
	}
	return containerStateFromString(c.State.Status), nil
}

func (d *DockerClient) CreateNetwork(ctx context.Context, networkName string, options dnetwork.CreateOptions) error {
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
func (d *DockerClient) RemoveNetwork(ctx context.Context, networkName string) error {
	log(ctx).Debugf("Removing network %s ...", networkName)
	err := d.client.NetworkRemove(ctx, networkName)
	if err != nil {
		log(ctx).Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to remove the network, reason: %w", err)
	}

	log(ctx).Debugf("Network %s removed successfully", networkName)
	return nil
}

func (d *DockerClient) NetworkExists(ctx context.Context, networkName string) bool {
	filter := dfilters.NewArgs()
	filter.Add("name", networkName)
	networks, err := d.client.NetworkList(ctx, dnetwork.ListOptions{
		Filters: filter,
	})

	// Ignore errors by considering the network is not present in case of
	// errors.
	return err == nil && len(networks) > 0
}

func (d *DockerClient) ConnectContainerToBridgeModeNetwork(ctx context.Context, containerName, networkName, ip string) error {
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
func (d *DockerClient) DisconnectContainerFromNetwork(ctx context.Context, containerName, networkName string) error {
	log(ctx).Debugf("Disconnecting container %s from network %s ...", containerName, networkName)
	err := d.client.NetworkDisconnect(ctx, networkName, containerName, false)
	if err != nil {
		log(ctx).Errorf("err: %s", reflect.TypeOf(err))
		return fmt.Errorf("failed to disconnect container %s from network %s, reason: %w", containerName, networkName, err)
	}

	log(ctx).Debugf("Container %s disconnected from network %s successfully", containerName, networkName)
	return nil
}

func (d *DockerClient) ContainerStopAndRemoveKillDelay() time.Duration {
	return d.containerStopAndRemoveKillDelay
}

func (d *DockerClient) Close() {
	d.client.Close()
}

func dockerDebugFromInspect(ctx context.Context) bool {
	lvl := inspect.HomelabInspectLevelFromContext(ctx)
	return lvl == inspect.HomelabInspectLevelDebug || lvl == inspect.HomelabInspectLevelTrace
}

func evalContainerStopAndRemoveKillDelay(ctx context.Context) time.Duration {
	if delay, ok := getContainerStopAndRemoveKillDelay(ctx); ok {
		return delay
	}
	return defaultContainerStopAndRemoveKillDelay
}
