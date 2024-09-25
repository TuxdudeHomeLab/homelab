package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"

	"github.com/sasha-s/go-deadlock"

	dtypes "github.com/docker/docker/api/types"
	dcontainer "github.com/docker/docker/api/types/container"
	dimage "github.com/docker/docker/api/types/image"
	dnetwork "github.com/docker/docker/api/types/network"
	derrdefs "github.com/docker/docker/errdefs"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type fakeDockerHost struct {
	mu                  deadlock.RWMutex
	containers          fakeContainerMap
	networks            fakeNetworkMap
	images              fakeImageMap
	validImagesForPull  stringSet
	failContainerCreate stringSet
	failContainerStart  stringSet
	failNetworkCreate   stringSet
	failNetworkConnect  stringSet
}

type fakeContainerInfo struct {
	name                 string
	id                   string
	state                containerState
	pendingRequiredStops int
	pendingRequiredKills int
	failInspect          bool
	failRemove           bool
	failStop             bool
	containerConfig      *dcontainer.Config
	hostConfig           *dcontainer.HostConfig
	networkConfig        *dnetwork.NetworkingConfig
}

type fakeNetworkInfo struct {
	name    string
	id      string
	options *dnetwork.CreateOptions
}

type fakeImageInfo struct {
	name string
	id   string
}

type fakeContainerInitInfo struct {
	name               string
	image              string
	state              containerState
	requiredExtraStops int
	requiredExtraKills int
	failInspect        bool
	failRemove         bool
	failStop           bool
}

type fakeNetworkInitInfo struct {
	name string
}

type fakeContainerMap map[string]*fakeContainerInfo
type fakeNetworkMap map[string]*fakeNetworkInfo
type fakeImageMap map[string]*fakeImageInfo

type fakeDockerHostInitInfo struct {
	containers          []*fakeContainerInitInfo
	networks            []*fakeNetworkInitInfo
	existingImages      stringSet
	validImagesForPull  stringSet
	failContainerCreate stringSet
	failContainerStart  stringSet
	failNetworkCreate   stringSet
	failNetworkConnect  stringSet
}

func fakeDockerHostFromContext(ctx context.Context) *fakeDockerHost {
	dockerClient, ok := dockerAPIClientFromContext(ctx)
	if !ok {
		panic("unable to retrieve the docker API client from context in test")
	}
	fakeDockerHost, ok := dockerClient.(*fakeDockerHost)
	if !ok {
		panic("unable to convert the retrieved client to fake docker host")
	}
	return fakeDockerHost
}

func newEmptyFakeDockerHost() *fakeDockerHost {
	return newFakeDockerHost(&fakeDockerHostInitInfo{})
}

func newFakeDockerHost(initInfo *fakeDockerHostInitInfo) *fakeDockerHost {
	f := &fakeDockerHost{
		containers:          fakeContainerMap{},
		networks:            fakeNetworkMap{},
		images:              fakeImageMap{},
		validImagesForPull:  stringSet{},
		failContainerCreate: stringSet{},
		failContainerStart:  stringSet{},
		failNetworkCreate:   stringSet{},
		failNetworkConnect:  stringSet{},
	}
	if initInfo == nil {
		return f
	}

	for _, ct := range initInfo.containers {
		ctInfo := newFakeContainerInfo(
			ct.name,
			&dcontainer.Config{Image: ct.image},
			&dcontainer.HostConfig{},
			&dnetwork.NetworkingConfig{})
		ctInfo.state = ct.state
		ctInfo.pendingRequiredStops = ct.requiredExtraStops
		ctInfo.pendingRequiredKills = ct.requiredExtraKills
		ctInfo.failInspect = ct.failInspect
		ctInfo.failStop = ct.failStop
		ctInfo.failRemove = ct.failRemove
		f.containers[ct.name] = ctInfo
	}
	for _, n := range initInfo.networks {
		f.networks[n.name] = newFakeNetworkInfo(n.name)
	}
	for img := range initInfo.existingImages {
		f.images[img] = newFakeImageInfo(img)
	}
	f.validImagesForPull = initInfo.validImagesForPull
	for c := range initInfo.failContainerCreate {
		f.failContainerCreate[c] = struct{}{}
	}
	for c := range initInfo.failContainerStart {
		f.failContainerStart[c] = struct{}{}
	}
	for n := range initInfo.failNetworkCreate {
		f.failNetworkCreate[n] = struct{}{}
	}
	for n := range initInfo.failNetworkConnect {
		f.failNetworkConnect[n] = struct{}{}
	}
	return f
}

func newFakeContainerInfo(containerName string, cConfig *dcontainer.Config, hConfig *dcontainer.HostConfig, nConfig *dnetwork.NetworkingConfig) *fakeContainerInfo {
	return &fakeContainerInfo{
		name:            containerName,
		id:              randomSHA256ID(),
		state:           containerStateCreated,
		containerConfig: cConfig,
		hostConfig:      hConfig,
		networkConfig:   nConfig,
	}
}

func newFakeNetworkInfo(networkName string) *fakeNetworkInfo {
	return &fakeNetworkInfo{
		name: networkName,
		id:   randomSHA256ID(),
	}
}

func newFakeImageInfo(imageName string) *fakeImageInfo {
	return &fakeImageInfo{
		name: imageName,
		id:   randomSHA256ID(),
	}
}

func (f *fakeDockerHost) Close() error {
	return nil
}

func (f *fakeDockerHost) ContainerCreate(ctx context.Context, cConfig *dcontainer.Config, hConfig *dcontainer.HostConfig, nConfig *dnetwork.NetworkingConfig, platform *ocispec.Platform, containerName string) (dcontainer.CreateResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	resp := dcontainer.CreateResponse{}
	if _, found := f.containers[containerName]; found {
		return resp, fmt.Errorf("container %s already exists in the fake docker host", containerName)
	}

	if nConfig != nil {
		for n := range nConfig.EndpointsConfig {
			if _, found := f.networks[n]; !found {
				return resp, fmt.Errorf("container %s is attempting to connect to network %s that doesn't exist on the fake docker host", containerName, n)
			}
		}
	}

	if _, found := f.failContainerCreate[containerName]; found {
		return resp, fmt.Errorf("failed to create container %s on the fake docker host", containerName)
	}

	ct := newFakeContainerInfo(containerName, cConfig, hConfig, nConfig)
	f.containers[containerName] = ct
	resp.ID = ct.id
	return resp, nil
}

func (f *fakeDockerHost) ContainerInspect(ctx context.Context, containerName string) (dtypes.ContainerJSON, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	ct, found := f.containers[containerName]
	if !found {
		return dtypes.ContainerJSON{}, derrdefs.NotFound(fmt.Errorf("container %s not found on the fake docker host", containerName))
	}

	if ct.failInspect {
		return dtypes.ContainerJSON{}, fmt.Errorf("failed to inspect container %s on the fake docker host", containerName)
	}

	return dtypes.ContainerJSON{
		ContainerJSONBase: &dtypes.ContainerJSONBase{
			ID:    ct.id,
			State: dockerContainerState(ct.state),
			Image: ct.containerConfig.Image,
			Name:  ct.name,
		},
	}, nil
}

func (f *fakeDockerHost) ContainerKill(ctx context.Context, containerName, signal string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	ct, found := f.containers[containerName]
	if !found {
		return derrdefs.NotFound(fmt.Errorf("container %s not found on the fake docker host", containerName))
	}

	switch ct.state {
	case containerStateRunning, containerStatePaused, containerStateRestarting:
		if ct.pendingRequiredKills > 0 {
			ct.pendingRequiredKills--
		} else {
			ct.state = containerStateExited
		}
		return nil
	case containerStateCreated, containerStateExited, containerStateDead, containerStateRemoving:
		return fmt.Errorf("container in state %s on the fake docker host cannot be killed", ct.state)
	case containerStateUnknown:
		panic("ContainerKill invoked on a container in an unknown state on the fake docker host, possibly indicating a bug")
	case containerStateNotFound:
		panic("ContainerKill invoked on a container in a not found state on the fake docker host, possibly indicating a bug")
	default:
		panic(fmt.Sprintf("ContainerKill invoked on a container in %s state on the fake docker host, possibly indicating a bug", ct.state))
	}
}

func (f *fakeDockerHost) ContainerRemove(ctx context.Context, containerName string, options dcontainer.RemoveOptions) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	ct, found := f.containers[containerName]
	if !found {
		return derrdefs.NotFound(fmt.Errorf("container %s not found on the fake docker host", containerName))
	}

	switch ct.state {
	case containerStateCreated, containerStateExited, containerStateDead:
		if ct.failRemove {
			return fmt.Errorf("failed to remove container %s on the fake docker host", containerName)
		}

		delete(f.containers, containerName)
		return nil
	case containerStateRunning, containerStatePaused, containerStateRestarting, containerStateRemoving:
		return fmt.Errorf("container in state %s on the fake docker host cannot be removed", ct.state)
	case containerStateUnknown:
		panic("ContainerRemove invoked on a container in an unknown state on the fake docker host, possibly indicating a bug")
	case containerStateNotFound:
		panic("ContainerRemove invoked on a container in a not found state on the fake docker host, possibly indicating a bug")
	default:
		panic(fmt.Sprintf("ContainerRemove invoked on a container in %s state on the fake docker host, possibly indicating a bug", ct.state))
	}
}

func (f *fakeDockerHost) ContainerStart(ctx context.Context, containerName string, options dcontainer.StartOptions) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	ct, found := f.containers[containerName]
	if !found {
		return derrdefs.NotFound(fmt.Errorf("container %s not found on the fake docker host", containerName))
	}

	if ct.state != containerStateCreated {
		return fmt.Errorf("container %s is not in created state that is required to start the container, but rather in state %s on the fake docker host", containerName, ct.state)
	}

	if _, found := f.failContainerStart[containerName]; found {
		return fmt.Errorf("failed to start container %s on the fake docker host", containerName)
	}

	ct.state = containerStateRunning
	return nil
}

func (f *fakeDockerHost) ContainerStop(ctx context.Context, containerName string, options dcontainer.StopOptions) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	ct, found := f.containers[containerName]
	if !found {
		return derrdefs.NotFound(fmt.Errorf("container %s not found on the fake docker host", containerName))
	}

	switch ct.state {
	case containerStateRunning, containerStatePaused, containerStateRestarting:
		if ct.failStop {
			return fmt.Errorf("failed to stop container %s on the fake docker host", containerName)
		}

		if ct.pendingRequiredStops > 0 {
			ct.pendingRequiredStops--
		} else {
			ct.state = containerStateExited
		}
		return nil
	// Created, Exited can be no-op stopped. Possibly Dead and Removing too.
	// However our real implementation doesn't invoke stop in these cases,
	// hence keeping the fake to the same contract for added safety.
	case containerStateCreated, containerStateExited, containerStateDead, containerStateRemoving:
		return fmt.Errorf("container in state %s on the fake docker host cannot be stopped", ct.state)
	case containerStateUnknown:
		panic("ContainerStop invoked on a container in an unknown state on the fake docker host, possibly indicating a bug")
	case containerStateNotFound:
		panic("ContainerStop invoked on a container in a not found state on the fake docker host, possibly indicating a bug")
	default:
		panic(fmt.Sprintf("ContainerStop invoked on a container in %s state on the fake docker host, possibly indicating a bug", ct.state))
	}
}

func (f *fakeDockerHost) ImageList(ctx context.Context, options dimage.ListOptions) ([]dimage.Summary, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if options.All {
		return nil, fmt.Errorf("listing all images on the fake docker host is unsupported")
	}
	if options.SharedSize {
		return nil, fmt.Errorf("retieving shared size of images on the fake docker host is unsupported")
	}
	if options.ContainerCount {
		return nil, fmt.Errorf("retieving container count for images on the fake docker host is unsupported")
	}
	if options.Manifests {
		return nil, fmt.Errorf("retieving manifests for images on the fake docker host is unsupported")
	}
	if options.Filters.Len() == 0 {
		return nil, fmt.Errorf("filters cannot be empty while listing images on the fake docker host")
	}
	if options.Filters.Len() != 1 {
		return nil, fmt.Errorf("filters must have exactly one arg while listing images on the fake docker host")
	}
	refs := options.Filters.Get("reference")
	if len(refs) != 1 {
		return nil, fmt.Errorf("filters must have exactly one reference key while listing images on the fake docker host")
	}
	img, found := f.images[refs[0]]
	if !found {
		return []dimage.Summary{}, nil
	}
	return []dimage.Summary{
		{
			ID: img.id,
		},
	}, nil
}

func (f *fakeDockerHost) ImagePull(ctx context.Context, refStr string, options dimage.PullOptions) (io.ReadCloser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, found := f.validImagesForPull[refStr]; !found {
		return nil, fmt.Errorf("image %s not found or invalid and cannot be pulled by the fake docker host", refStr)
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func (f *fakeDockerHost) NetworkConnect(ctx context.Context, networkName, containerName string, config *dnetwork.EndpointSettings) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, found := f.networks[networkName]; !found {
		return derrdefs.NotFound(fmt.Errorf("network %s not found on the fake docker host", networkName))
	}

	if _, found := f.failNetworkConnect[networkName]; found {
		return fmt.Errorf("failed to connect container %s to network %s on the fake docker host", containerName, networkName)
	}

	// TODO: Perform more validations of the network endpoint within
	// the network.
	return nil
}

func (f *fakeDockerHost) NetworkCreate(ctx context.Context, networkName string, options dnetwork.CreateOptions) (dnetwork.CreateResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	resp := dnetwork.CreateResponse{}
	if _, found := f.networks[networkName]; found {
		return resp, fmt.Errorf("network %s already exists in the fake docker host", networkName)
	}

	if _, found := f.failNetworkCreate[networkName]; found {
		return resp, fmt.Errorf("failed to create network %s on the fake docker host", networkName)
	}

	n := newFakeNetworkInfo(networkName)
	n.options = &options
	f.networks[networkName] = n
	resp.ID = n.id
	return resp, nil
}

func (f *fakeDockerHost) NetworkDisconnect(ctx context.Context, networkName, containerName string, force bool) error {
	panic("NetworkDisconnect unimplemented")
}

func (f *fakeDockerHost) NetworkList(ctx context.Context, options dnetwork.ListOptions) ([]dnetwork.Summary, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if options.Filters.Len() == 0 {
		return nil, fmt.Errorf("filters cannot be empty while listing networks on the fake docker host")
	}
	if options.Filters.Len() != 1 {
		return nil, fmt.Errorf("filters must have exactly one arg while listing networks on the fake docker host")
	}
	refs := options.Filters.Get("name")
	if len(refs) != 1 {
		return nil, fmt.Errorf("filters must have exactly one name key while listing networks on the fake docker host")
	}
	n, found := f.networks[refs[0]]
	if !found {
		return []dnetwork.Summary{}, nil
	}
	return []dnetwork.Summary{
		{
			Name:  n.name,
			ID:    n.id,
			Scope: "local",
		},
	}, nil
}

func (f *fakeDockerHost) NetworkRemove(ctx context.Context, networkName string) error {
	panic("NetworkRemove unimplemented")
}

func (f *fakeDockerHost) forceRemoveContainer(containerName string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	_, found := f.containers[containerName]
	if !found {
		return derrdefs.NotFound(fmt.Errorf("container %s not found on the fake docker host", containerName))
	}

	delete(f.containers, containerName)
	return nil
}

func (f *fakeDockerHost) getContainerState(containerName string) containerState {
	f.mu.RLock()
	defer f.mu.RUnlock()

	ct, found := f.containers[containerName]
	if !found {
		return containerStateNotFound
	}
	return ct.state
}

func dockerContainerState(state containerState) *dtypes.ContainerState {
	st := &dtypes.ContainerState{}
	switch state {
	case containerStateCreated:
		st.Status = "created"
	case containerStateRunning:
		st.Status = "running"
		st.Running = true
	case containerStatePaused:
		st.Status = "paused"
		st.Paused = true
	case containerStateRestarting:
		st.Status = "restarting"
		st.Restarting = true
	case containerStateRemoving:
		st.Status = "removing"
	case containerStateExited:
		st.Status = "exited"
	case containerStateDead:
		st.Status = "dead"
	case containerStateUnknown:
		panic("Unknown container state on fake docker host")
	case containerStateNotFound:
		panic("Cannot handle Not found container state while building container state info on fake docker host, possibly indicating a bug in the code")
	default:
		panic("Invalid scenario while building container state on fake docker host, possibly indicating a bug in the code")
	}
	return st
}

func randomSHA256ID() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	h := sha256.New()
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}
