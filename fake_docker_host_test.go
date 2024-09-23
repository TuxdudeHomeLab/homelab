package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"

	dtypes "github.com/docker/docker/api/types"
	dcontainer "github.com/docker/docker/api/types/container"
	dimage "github.com/docker/docker/api/types/image"
	dnetwork "github.com/docker/docker/api/types/network"
	derrdefs "github.com/docker/docker/errdefs"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type fakeDockerHost struct {
	containers         fakeContainerMap
	networks           fakeNetworkMap
	images             fakeImageMap
	validImagesForPull stringSet
}

type fakeContainerInfo struct {
	name            string
	id              string
	state           containerState
	containerConfig *dcontainer.Config
	hostConfig      *dcontainer.HostConfig
	networkConfig   *dnetwork.NetworkingConfig
}

type fakeNetworkInfo struct {
	name string
	id   string
}

type fakeImageInfo struct {
	name string
	id   string
}

type fakeContainerInitInfo struct {
	name  string
	image string
	state containerState
}

type fakeNetworkInitInfo struct {
	name string
}

type fakeContainerMap map[string]*fakeContainerInfo
type fakeNetworkMap map[string]*fakeNetworkInfo
type fakeImageMap map[string]*fakeImageInfo

type fakeDockerHostInitInfo struct {
	containers         []*fakeContainerInitInfo
	networks           []*fakeNetworkInitInfo
	existingImages     stringSet
	validImagesForPull stringSet
}

func newEmptyFakeDockerHost() *fakeDockerHost {
	return newFakeDockerHost(&fakeDockerHostInitInfo{})
}

func newFakeDockerHost(initInfo *fakeDockerHostInitInfo) *fakeDockerHost {
	f := &fakeDockerHost{
		containers:         fakeContainerMap{},
		networks:           fakeNetworkMap{},
		images:             fakeImageMap{},
		validImagesForPull: stringSet{},
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
		f.containers[ct.name] = ctInfo
	}
	for _, n := range initInfo.networks {
		f.networks[n.name] = newFakeNetworkInfo(n.name)
	}
	for img := range initInfo.existingImages {
		f.images[img] = newFakeImageInfo(img)
	}
	f.validImagesForPull = initInfo.validImagesForPull
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
	resp := dcontainer.CreateResponse{}
	if _, found := f.containers[containerName]; found {
		return resp, fmt.Errorf("container %s already exists in the fake docker host", containerName)
	}
	ct := newFakeContainerInfo(containerName, cConfig, hConfig, nConfig)
	f.containers[containerName] = ct
	resp.ID = ct.id
	return resp, nil
}

func (f *fakeDockerHost) ContainerInspect(ctx context.Context, containerName string) (dtypes.ContainerJSON, error) {
	ct, found := f.containers[containerName]
	if !found {
		return dtypes.ContainerJSON{}, derrdefs.NotFound(fmt.Errorf("container %s not found on the fake docker host", containerName))
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
	panic("ContainerKill unimplemented")
}

func (f *fakeDockerHost) ContainerRemove(ctx context.Context, containerName string, options dcontainer.RemoveOptions) error {
	panic("ContainerRemove unimplemented")
}

func (f *fakeDockerHost) ContainerStart(ctx context.Context, containerName string, options dcontainer.StartOptions) error {
	ct, found := f.containers[containerName]
	if !found {
		return derrdefs.NotFound(fmt.Errorf("container %s not found on the fake docker host", containerName))
	}
	if ct.state != containerStateCreated {
		return fmt.Errorf("container %s is not in created state that is required to start the container, but rather in state %s on the fake docker host", containerName, ct.state)
	}
	ct.state = containerStateRunning
	return nil
}

func (f *fakeDockerHost) ContainerStop(ctx context.Context, containerName string, options dcontainer.StopOptions) error {
	panic("ContainerStop unimplemented")
}

func (f *fakeDockerHost) ImageList(ctx context.Context, options dimage.ListOptions) ([]dimage.Summary, error) {
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
	if _, found := f.validImagesForPull[refStr]; !found {
		return nil, fmt.Errorf("image %s not found or invalid and cannot be pulled by the fake docker host", refStr)
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func (f *fakeDockerHost) NetworkConnect(ctx context.Context, networkName, containerName string, config *dnetwork.EndpointSettings) error {
	if _, found := f.networks[networkName]; !found {
		return derrdefs.NotFound(fmt.Errorf("network %s not found on the fake docker host", networkName))
	}
	// TODO: Perform more validations of the network endpoint within
	// the network.
	return nil
}

func (f *fakeDockerHost) NetworkDisconnect(ctx context.Context, networkName, containerName string, force bool) error {
	panic("NetworkDisconnect unimplemented")
}

func (f *fakeDockerHost) NetworkList(ctx context.Context, options dnetwork.ListOptions) ([]dnetwork.Summary, error) {
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
