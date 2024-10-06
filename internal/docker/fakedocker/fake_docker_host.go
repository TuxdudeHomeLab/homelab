package fakedocker

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/sasha-s/go-deadlock"
	"github.com/tuxdudehomelab/homelab/internal/docker"
	"github.com/tuxdudehomelab/homelab/internal/utils"

	dtypes "github.com/docker/docker/api/types"
	dcontainer "github.com/docker/docker/api/types/container"
	dimage "github.com/docker/docker/api/types/image"
	dnetwork "github.com/docker/docker/api/types/network"
	derrdefs "github.com/docker/docker/errdefs"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type FakeDockerHost struct {
	mu                   deadlock.RWMutex
	containers           fakeContainerMap
	networks             fakeNetworkMap
	images               fakeImageMap
	warnContainerCreate  utils.StringSet
	failContainerCreate  utils.StringSet
	failContainerInspect utils.StringSet
	failContainerKill    utils.StringSet
	failContainerRemove  utils.StringSet
	failContainerStart   utils.StringSet
	failContainerStop    utils.StringSet
	validImagesForPull   utils.StringSet
	failImagePull        utils.StringSet
	noImageAfterPull     utils.StringSet
	warnNetworkCreate    utils.StringSet
	failNetworkCreate    utils.StringSet
	failNetworkConnect   utils.StringSet
}

type fakeContainerInfo struct {
	name                 string
	id                   string
	state                docker.ContainerState
	containerStopIssued  bool
	pendingRequiredStops int
	pendingRequiredKills int
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

type FakeContainerInitInfo struct {
	Name               string
	Image              string
	State              docker.ContainerState
	RequiredExtraStops int
	RequiredExtraKills int
}

type FakeNetworkInitInfo struct {
	Name string
}

type fakeContainerMap map[string]*fakeContainerInfo
type fakeNetworkMap map[string]*fakeNetworkInfo
type fakeImageMap map[string]*fakeImageInfo

type FakeDockerHostInitInfo struct {
	Containers           []*FakeContainerInitInfo
	Networks             []*FakeNetworkInitInfo
	ExistingImages       utils.StringSet
	WarnContainerCreate  utils.StringSet
	FailContainerCreate  utils.StringSet
	FailContainerInspect utils.StringSet
	FailContainerKill    utils.StringSet
	FailContainerRemove  utils.StringSet
	FailContainerStart   utils.StringSet
	FailContainerStop    utils.StringSet
	ValidImagesForPull   utils.StringSet
	FailImagePull        utils.StringSet
	NoImageAfterPull     utils.StringSet
	WarnNetworkCreate    utils.StringSet
	FailNetworkCreate    utils.StringSet
	FailNetworkConnect   utils.StringSet
}

type wrappedReader func(p []byte) (int, error)

func (w wrappedReader) Read(p []byte) (int, error) {
	return w(p)
}

func FakeDockerHostFromContext(ctx context.Context) *FakeDockerHost {
	if f, ok := docker.MustAPIClient(ctx).(*FakeDockerHost); ok {
		return f
	}
	panic("unable to convert the retrieved client to fake docker host")
}

func NewEmptyFakeDockerHost() *FakeDockerHost {
	return NewFakeDockerHost(nil)
}

func NewFakeDockerHost(initInfo *FakeDockerHostInitInfo) *FakeDockerHost {
	f := &FakeDockerHost{
		containers:           fakeContainerMap{},
		networks:             fakeNetworkMap{},
		images:               fakeImageMap{},
		warnContainerCreate:  utils.StringSet{},
		failContainerCreate:  utils.StringSet{},
		failContainerInspect: utils.StringSet{},
		failContainerKill:    utils.StringSet{},
		failContainerRemove:  utils.StringSet{},
		failContainerStart:   utils.StringSet{},
		failContainerStop:    utils.StringSet{},
		validImagesForPull:   utils.StringSet{},
		failImagePull:        utils.StringSet{},
		noImageAfterPull:     utils.StringSet{},
		warnNetworkCreate:    utils.StringSet{},
		failNetworkCreate:    utils.StringSet{},
		failNetworkConnect:   utils.StringSet{},
	}
	if initInfo == nil {
		return f
	}

	for _, ct := range initInfo.Containers {
		ctInfo := newFakeContainerInfo(
			ct.Name,
			&dcontainer.Config{Image: ct.Image},
			&dcontainer.HostConfig{},
			&dnetwork.NetworkingConfig{})
		ctInfo.state = ct.State
		ctInfo.pendingRequiredStops = ct.RequiredExtraStops
		ctInfo.pendingRequiredKills = ct.RequiredExtraKills
		f.containers[ct.Name] = ctInfo
	}
	for _, n := range initInfo.Networks {
		f.networks[n.Name] = newFakeNetworkInfo(n.Name)
	}
	for img := range initInfo.ExistingImages {
		f.images[img] = newFakeImageInfo(img)
	}
	for c := range initInfo.WarnContainerCreate {
		f.warnContainerCreate[c] = struct{}{}
	}
	for c := range initInfo.FailContainerCreate {
		f.failContainerCreate[c] = struct{}{}
	}
	for c := range initInfo.FailContainerInspect {
		f.failContainerInspect[c] = struct{}{}
	}
	for c := range initInfo.FailContainerKill {
		f.failContainerKill[c] = struct{}{}
	}
	for c := range initInfo.FailContainerRemove {
		f.failContainerRemove[c] = struct{}{}
	}
	for c := range initInfo.FailContainerStart {
		f.failContainerStart[c] = struct{}{}
	}
	for c := range initInfo.FailContainerStop {
		f.failContainerStop[c] = struct{}{}
	}
	for i := range initInfo.ValidImagesForPull {
		f.validImagesForPull[i] = struct{}{}
	}
	for i := range initInfo.FailImagePull {
		f.failImagePull[i] = struct{}{}
	}
	for i := range initInfo.NoImageAfterPull {
		f.noImageAfterPull[i] = struct{}{}
	}
	for n := range initInfo.WarnNetworkCreate {
		f.warnNetworkCreate[n] = struct{}{}
	}
	for n := range initInfo.FailNetworkCreate {
		f.failNetworkCreate[n] = struct{}{}
	}
	for n := range initInfo.FailNetworkConnect {
		f.failNetworkConnect[n] = struct{}{}
	}
	return f
}

func newFakeContainerInfo(containerName string, cConfig *dcontainer.Config, hConfig *dcontainer.HostConfig, nConfig *dnetwork.NetworkingConfig) *fakeContainerInfo {
	return &fakeContainerInfo{
		name:                containerName,
		id:                  randomSHA256ID(),
		state:               docker.ContainerStateCreated,
		containerStopIssued: false,
		containerConfig:     cConfig,
		hostConfig:          hConfig,
		networkConfig:       nConfig,
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

func (f *FakeDockerHost) Close() error {
	return nil
}

func (f *FakeDockerHost) ContainerCreate(ctx context.Context, cConfig *dcontainer.Config, hConfig *dcontainer.HostConfig, nConfig *dnetwork.NetworkingConfig, platform *ocispec.Platform, containerName string) (dcontainer.CreateResponse, error) {
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

	if _, found := f.warnContainerCreate[containerName]; found {
		resp.Warnings = []string{
			fmt.Sprintf("first warning generated during container create for %s on the fake docker host", containerName),
			fmt.Sprintf("second warning generated during container create for %s on the fake docker host", containerName),
			fmt.Sprintf("third warning generated during container create for %s on the fake docker host", containerName),
		}
	}

	return resp, nil
}

func (f *FakeDockerHost) ContainerInspect(ctx context.Context, containerName string) (dtypes.ContainerJSON, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	ct, found := f.containers[containerName]
	if !found {
		return dtypes.ContainerJSON{}, derrdefs.NotFound(fmt.Errorf("container %s not found on the fake docker host", containerName))
	}

	if _, found := f.failContainerInspect[containerName]; found {
		return dtypes.ContainerJSON{}, fmt.Errorf("failed to inspect container %s on the fake docker host", containerName)
	}

	return dtypes.ContainerJSON{
		ContainerJSONBase: &dtypes.ContainerJSONBase{
			ID:    ct.id,
			State: fakeDockerContainerState(ct.state),
			Image: ct.containerConfig.Image,
			Name:  ct.name,
		},
	}, nil
}

func (f *FakeDockerHost) ContainerKill(ctx context.Context, containerName, signal string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	ct, found := f.containers[containerName]
	if !found {
		return derrdefs.NotFound(fmt.Errorf("container %s not found on the fake docker host", containerName))
	}

	switch ct.state {
	case docker.ContainerStateRunning, docker.ContainerStatePaused, docker.ContainerStateRestarting:
		if ct.pendingRequiredKills > 0 {
			ct.pendingRequiredKills--
		} else {
			if _, found := f.failContainerKill[containerName]; found {
				return fmt.Errorf("failed to kill container %s on the fake docker host", containerName)
			}

			ct.state = docker.ContainerStateExited
		}
		return nil
	case docker.ContainerStateCreated, docker.ContainerStateExited, docker.ContainerStateDead, docker.ContainerStateRemoving:
		return fmt.Errorf("container in state %s on the fake docker host cannot be killed", ct.state)
	case docker.ContainerStateUnknown:
		panic("ContainerKill invoked on a container in an unknown state on the fake docker host, possibly indicating a bug")
	case docker.ContainerStateNotFound:
		panic("ContainerKill invoked on a container in a not found state on the fake docker host, possibly indicating a bug")
	default:
		panic(fmt.Sprintf("ContainerKill invoked on a container in %s state on the fake docker host, possibly indicating a bug", ct.state))
	}
}

func (f *FakeDockerHost) ContainerRemove(ctx context.Context, containerName string, options dcontainer.RemoveOptions) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	ct, found := f.containers[containerName]
	if !found {
		return derrdefs.NotFound(fmt.Errorf("container %s not found on the fake docker host", containerName))
	}

	switch ct.state {
	case docker.ContainerStateCreated, docker.ContainerStateExited, docker.ContainerStateDead:
		if _, found := f.failContainerRemove[containerName]; found {
			return fmt.Errorf("failed to remove container %s on the fake docker host", containerName)
		}

		delete(f.containers, containerName)
		return nil
	case docker.ContainerStateRunning, docker.ContainerStatePaused, docker.ContainerStateRestarting, docker.ContainerStateRemoving:
		return fmt.Errorf("container in state %s on the fake docker host cannot be removed", ct.state)
	case docker.ContainerStateUnknown:
		panic("ContainerRemove invoked on a container in an unknown state on the fake docker host, possibly indicating a bug")
	case docker.ContainerStateNotFound:
		panic("ContainerRemove invoked on a container in a not found state on the fake docker host, possibly indicating a bug")
	default:
		panic(fmt.Sprintf("ContainerRemove invoked on a container in %s state on the fake docker host, possibly indicating a bug", ct.state))
	}
}

func (f *FakeDockerHost) ContainerStart(ctx context.Context, containerName string, options dcontainer.StartOptions) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	ct, found := f.containers[containerName]
	if !found {
		return derrdefs.NotFound(fmt.Errorf("container %s not found on the fake docker host", containerName))
	}

	if ct.state != docker.ContainerStateCreated {
		return fmt.Errorf("container %s is not in created state that is required to start the container, but rather in state %s on the fake docker host", containerName, ct.state)
	}

	if _, found := f.failContainerStart[containerName]; found {
		return fmt.Errorf("failed to start container %s on the fake docker host", containerName)
	}

	ct.state = docker.ContainerStateRunning
	return nil
}

func (f *FakeDockerHost) ContainerStop(ctx context.Context, containerName string, options dcontainer.StopOptions) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	ct, found := f.containers[containerName]
	if !found {
		return derrdefs.NotFound(fmt.Errorf("container %s not found on the fake docker host", containerName))
	}
	ct.containerStopIssued = true

	switch ct.state {
	case docker.ContainerStateRunning, docker.ContainerStatePaused, docker.ContainerStateRestarting:
		if _, found := f.failContainerStop[containerName]; found {
			return fmt.Errorf("failed to stop container %s on the fake docker host", containerName)
		}

		if ct.pendingRequiredStops > 0 {
			ct.pendingRequiredStops--
		} else {
			ct.state = docker.ContainerStateExited
		}
		return nil
	// Created, Exited can be no-op stopped. Possibly Dead and Removing too.
	// However our real implementation doesn't invoke stop in these cases,
	// hence keeping the fake to the same contract for added safety.
	case docker.ContainerStateCreated, docker.ContainerStateExited, docker.ContainerStateDead, docker.ContainerStateRemoving:
		return fmt.Errorf("container in state %s on the fake docker host cannot be stopped", ct.state)
	case docker.ContainerStateUnknown:
		panic("ContainerStop invoked on a container in an unknown state on the fake docker host, possibly indicating a bug")
	case docker.ContainerStateNotFound:
		panic("ContainerStop invoked on a container in a not found state on the fake docker host, possibly indicating a bug")
	default:
		panic(fmt.Sprintf("ContainerStop invoked on a container in %s state on the fake docker host, possibly indicating a bug", ct.state))
	}
}

func (f *FakeDockerHost) ImageList(ctx context.Context, options dimage.ListOptions) ([]dimage.Summary, error) {
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

func (f *FakeDockerHost) ImagePull(ctx context.Context, imageName string, options dimage.PullOptions) (io.ReadCloser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, found := f.validImagesForPull[imageName]; !found {
		return nil, fmt.Errorf("image %s not found or invalid and cannot be pulled by the fake docker host", imageName)
	}

	return io.NopCloser(wrappedReader(func(p []byte) (int, error) {
		f.mu.Lock()
		defer f.mu.Unlock()

		if _, found := f.failImagePull[imageName]; found {
			return 0, fmt.Errorf("failed to pull image %s on the fake docker host", imageName)
		}

		if _, found := f.noImageAfterPull[imageName]; !found {
			f.images[imageName] = newFakeImageInfo(imageName)
		}
		return 0, io.EOF
	})), nil
}

func (f *FakeDockerHost) NetworkConnect(ctx context.Context, networkName, containerName string, config *dnetwork.EndpointSettings) error {
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

func (f *FakeDockerHost) NetworkCreate(ctx context.Context, networkName string, options dnetwork.CreateOptions) (dnetwork.CreateResponse, error) {
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
	if _, found := f.warnNetworkCreate[networkName]; found {
		resp.Warning = fmt.Sprintf("warning generated during network create for network %s on the fake docker host", networkName)
	}
	return resp, nil
}

func (f *FakeDockerHost) NetworkDisconnect(ctx context.Context, networkName, containerName string, force bool) error {
	panic("NetworkDisconnect unimplemented")
}

func (f *FakeDockerHost) NetworkList(ctx context.Context, options dnetwork.ListOptions) ([]dnetwork.Summary, error) {
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

func (f *FakeDockerHost) NetworkRemove(ctx context.Context, networkName string) error {
	panic("NetworkRemove unimplemented")
}

func (f *FakeDockerHost) ForceRemoveContainer(containerName string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	_, found := f.containers[containerName]
	if !found {
		return derrdefs.NotFound(fmt.Errorf("container %s not found on the fake docker host", containerName))
	}

	delete(f.containers, containerName)
	return nil
}

func (f *FakeDockerHost) GetContainerState(containerName string) docker.ContainerState {
	f.mu.RLock()
	defer f.mu.RUnlock()

	ct, found := f.containers[containerName]
	if !found {
		return docker.ContainerStateNotFound
	}
	return ct.state
}

func (f *FakeDockerHost) ContainerStopIssued(containerName string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if ct, found := f.containers[containerName]; found {
		return ct.containerStopIssued
	}
	return false
}

func fakeDockerContainerState(state docker.ContainerState) *dtypes.ContainerState {
	st := &dtypes.ContainerState{}
	switch state {
	case docker.ContainerStateCreated:
		st.Status = "created"
	case docker.ContainerStateRunning:
		st.Status = "running"
		st.Running = true
	case docker.ContainerStatePaused:
		st.Status = "paused"
		st.Paused = true
	case docker.ContainerStateRestarting:
		st.Status = "restarting"
		st.Restarting = true
	case docker.ContainerStateRemoving:
		st.Status = "removing"
	case docker.ContainerStateExited:
		st.Status = "exited"
	case docker.ContainerStateDead:
		st.Status = "dead"
	// We allow these unsupported values just because this is fake docker host
	// and we want to cover a few additional edge cases in our tests.
	case docker.ContainerStateUnknown:
		st.Status = "unknown"
	case docker.ContainerStateNotFound:
		st.Status = "notFound"
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
