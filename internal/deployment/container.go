package deployment

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	dcontainer "github.com/docker/docker/api/types/container"
	dnetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/tuxdudehomelab/homelab/internal/config"
	"github.com/tuxdudehomelab/homelab/internal/docker"
	"github.com/tuxdudehomelab/homelab/internal/utils"
)

const (
	// six attempts implies we will attempt killing the container
	// five times in the worst case before giving up. The one
	// remaining attempt is to remove the container which has been
	// terminated.
	stopAndRemoveAttempts = 6
)

type Container struct {
	config        *config.ContainerConfig
	globalConfig  *config.GlobalConfig
	group         *ContainerGroup
	endpoints     networkEndpointList
	allowedOnHost bool
}

type containerNetworkEndpoint struct {
	network *Network
	ip      string
}

type containerDockerConfigs struct {
	ContainerConfig *dcontainer.Config
	HostConfig      *dcontainer.HostConfig
	NetworkConfig   *dnetwork.NetworkingConfig
}

type networkEndpointList []*containerNetworkEndpoint
type containerList []*Container
type containerSet map[config.ContainerReference]bool
type containerMap map[config.ContainerReference]*Container
type containerDockerConfigMap map[config.ContainerReference]*containerDockerConfigs

func newContainer(group *ContainerGroup, config *config.ContainerConfig, globalConfig *config.GlobalConfig, endpoints networkEndpointList, allowedOnHost bool) *Container {
	return &Container{
		config:        config,
		globalConfig:  globalConfig,
		group:         group,
		endpoints:     endpoints,
		allowedOnHost: allowedOnHost,
	}
}

func (c *Container) isAllowedOnCurrentHost() bool {
	return c.allowedOnHost
}

func (c *Container) Start(ctx context.Context, dc *docker.DockerClient) (bool, error) {
	log(ctx).Debugf("Starting container %s ...", c.Name())

	// Validate the container is allowed to run on the current host.
	if !c.isAllowedOnCurrentHost() {
		return false, nil
	}

	err := c.startInternal(ctx, dc)
	if err != nil {
		return false, utils.LogToErrorAndReturn(ctx, "Failed to start container %s, reason:%v", c.Name(), err)
	}

	log(ctx).Infof("Started container %s", c.Name())
	log(ctx).InfoEmpty()
	return true, nil
}

func (c *Container) purge(ctx context.Context, dc *docker.DockerClient) error {
	purged := false
	stoppedOnceAlready := false
	attemptsRemaining := stopAndRemoveAttempts

	for !purged && attemptsRemaining > 0 {
		attemptsRemaining--

		st, err := dc.GetContainerState(ctx, c.Name())
		if err != nil {
			return err
		}
		log(ctx).Debugf("Container %s current state: %s", c.Name(), st)

		switch st {
		case docker.ContainerStateNotFound:
			// Nothing to stop and/or remove.
			purged = true
		case docker.ContainerStateRunning, docker.ContainerStatePaused, docker.ContainerStateRestarting:
			// Stop the container if not stopped already.
			if !stoppedOnceAlready {
				err = dc.StopContainer(ctx, c.Name())
				if err != nil {
					return err
				}
				stoppedOnceAlready = true
				// Reset this to attempt killing the container at least
				// stopAndRemoveAttempts -1 times prior to giving up.
				attemptsRemaining = stopAndRemoveAttempts
			} else {
				// Kill the container next as a precaution and ignore any errors.
				_ = dc.KillContainer(ctx, c.Name())
				// Add a delay before checking the container state again.
				time.Sleep(dc.ContainerStopAndRemoveKillDelay())
			}
		case docker.ContainerStateCreated, docker.ContainerStateExited, docker.ContainerStateDead:
			// Directly remove the container.
			err = dc.RemoveContainer(ctx, c.Name())
			if err != nil {
				return err
			}
		case docker.ContainerStateRemoving:
			// Nothing to be done here, although this could lead to some
			// unknown handling in next steps.
			log(ctx).Warnf("container %s is in REMOVING state already, can lead to issues while we create the container next", c.Name())
			// Add a delay before checking the container state again.
			time.Sleep(dc.ContainerStopAndRemoveKillDelay())
		default:
			log(ctx).Fatalf("container %s is in an unsupported state %v, possibly indicating a bug in the code", c.Name(), st)
		}
	}

	if purged {
		return nil
	}

	// Check the container state one final time after exhausting all purge
	// attempts, and return the final error status based on that.
	st, err := dc.GetContainerState(ctx, c.Name())
	if err != nil {
		return err
	}
	if st != docker.ContainerStateNotFound {
		return fmt.Errorf("failed to stop and remove container %s after %d attempts", c.Name(), stopAndRemoveAttempts)
	}
	return nil
}

func (c *Container) startInternal(ctx context.Context, dc *docker.DockerClient) error {
	// 1. Execute any pre-start commands.
	// TODO: Implement this.

	// 2. Pull the container image.
	err := dc.PullImage(ctx, c.imageReference())
	if err != nil {
		return err
	}

	// 3. Purge (i.e. stop and remove) any previously existing containers
	// under the same name.
	err = c.purge(ctx, dc)
	if err != nil {
		return err
	}

	// 4. For the primary network interface of the container, create
	// the network for the container prior to creating the container
	// attached to this network.
	if len(c.endpoints) > 0 {
		// network.create(...) gracefully handles the case for when the
		// network exists already.
		err := c.endpoints[0].network.create(ctx, dc)
		if err != nil {
			return err
		}
		log(ctx).Debugf("Connecting container %s to network %s with IP %s at the time of container creation ...", c.Name(), c.endpoints[0].network.name(), c.endpoints[0].ip)
	} else {
		log(ctx).Warnf("Container %s has no network endpoints configured, this is uncommon!", c.Name())
	}

	// 5. Create the container.
	cConfig, hConfig, nConfig, err := c.generateDockerConfigs()
	if err != nil {
		return err
	}
	err = dc.CreateContainer(ctx, c.Name(), cConfig, hConfig, nConfig)
	if err != nil {
		return err
	}

	// 6. For each non-primary network interface of the container, create
	// the network for the container if it doesn't exist already prior to
	// connecting the container to the network.
	for i := 1; i < len(c.endpoints); i++ {
		ip := c.endpoints[i]
		err := ip.network.create(ctx, dc)
		if err != nil {
			return err
		}
		err = ip.network.connectContainer(ctx, dc, c.Name(), ip.ip)
		if err != nil {
			return err
		}
	}

	// 7. Start the created container.
	err = dc.StartContainer(ctx, c.Name())
	return err
}

func (c *Container) generateDockerConfigs() (*dcontainer.Config, *dcontainer.HostConfig, *dnetwork.NetworkingConfig, error) {
	pMap, pSet := c.publishedPorts()
	return c.dockerContainerConfig(pSet), c.dockerHostConfig(pMap), c.dockerNetworkConfig(), nil
}

func (c *Container) dockerContainerConfig(pSet nat.PortSet) *dcontainer.Config {
	return &dcontainer.Config{
		Hostname:        c.hostName(),
		Domainname:      c.domainName(),
		User:            c.userAndGroup(),
		ExposedPorts:    pSet,
		Tty:             c.attachToTty(),
		Env:             c.envVars(),
		Cmd:             c.args(),
		Entrypoint:      c.entrypoint(),
		NetworkDisabled: c.isNetworkDisabled(),
		Labels:          c.labels(),
		StopSignal:      c.stopSignal(),
		StopTimeout:     c.stopTimeout(),
		Image:           c.imageReference(),
	}
}

func (c *Container) dockerHostConfig(pMap nat.PortMap) *dcontainer.HostConfig {
	return &dcontainer.HostConfig{
		Binds:          c.bindMounts(),
		NetworkMode:    c.networkMode(),
		PortBindings:   pMap,
		RestartPolicy:  c.restartPolicy(),
		AutoRemove:     c.autoRemove(),
		CapAdd:         c.capAddList(),
		CapDrop:        c.capDropList(),
		DNS:            c.dnsServers(),
		DNSOptions:     c.dnsOptions(),
		DNSSearch:      c.dnsSearch(),
		GroupAdd:       c.additionalUserGroups(),
		Privileged:     c.privileged(),
		ReadonlyRootfs: c.readOnlyRootfs(),
		Tmpfs:          c.tmpfsMounts(),
		ShmSize:        c.shmSize(),
		Sysctls:        c.sysctls(),
	}
}

func (c *Container) dockerNetworkConfig() *dnetwork.NetworkingConfig {
	ne := c.primaryNetworkEndpoint()
	if ne == nil {
		return nil
	}
	return &dnetwork.NetworkingConfig{
		EndpointsConfig: c.primaryNetworkEndpoint(),
	}
}

func (c *Container) Name() string {
	return containerName(&c.config.Info)
}

func (c *Container) hostName() string {
	return c.config.Network.HostName
}

func (c *Container) domainName() string {
	d := c.config.Network.DomainName
	if len(d) == 0 {
		d = c.globalConfig.Container.DomainName
	}
	return d
}

func (c *Container) userAndGroup() string {
	u := c.config.User.User
	g := c.config.User.PrimaryGroup
	if len(g) > 0 {
		return fmt.Sprintf("%s:%s", u, g)
	}
	return u
}

func (c *Container) attachToTty() bool {
	return c.config.Runtime.AttachToTty
}

func (c *Container) envVars() []string {
	env := make(map[string]string, 0)
	envKeys := make([]string, 0)
	// TODO: Substitute global config env variables in the value fields.
	// TODO: Support invoking ValueCmd for evaluating the value.
	for _, e := range c.globalConfig.Container.Env {
		env[e.Var] = e.Value
		envKeys = append(envKeys, e.Var)
	}
	for _, e := range c.config.Runtime.Env {
		if _, found := env[e.Var]; !found {
			envKeys = append(envKeys, e.Var)
		}
		env[e.Var] = e.Value
	}

	res := make([]string, 0)
	for _, k := range envKeys {
		res = append(res, fmt.Sprintf("%s=%s", k, env[k]))
	}

	if len(res) == 0 {
		return nil
	}
	return res
}

func (c *Container) args() []string {
	return c.config.Runtime.Args
}

func (c *Container) entrypoint() []string {
	return c.config.Runtime.Entrypoint
}

func (c *Container) isNetworkDisabled() bool {
	return false
}

func (c *Container) labels() map[string]string {
	res := make(map[string]string, 0)
	for _, l := range c.config.Metadata.Labels {
		res[l.Name] = l.Value
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

func (c *Container) stopSignal() string {
	return c.config.Lifecycle.StopSignal
}

func (c *Container) stopTimeout() *int {
	t := c.config.Lifecycle.StopTimeout
	if t == 0 {
		t = c.globalConfig.Container.StopTimeout
	}
	return &t
}

func (c *Container) imageReference() string {
	return c.config.Image.Image
}

func (c *Container) bindMounts() []string {
	// TODO: Do this once for the entire deployment and reuse it.
	bm := make(map[string]string, 0)
	mountNames := make([]string, 0)
	for _, md := range c.globalConfig.MountDefs {
		bm[md.Name] = mountConfigToString(&md)
		mountNames = append(mountNames, md.Name)
	}

	binds := make(map[string]string, 0)
	// Get all the global container config mounts.
	// TODO: Do this once for the entire deployment and reuse it.
	for _, mount := range c.globalConfig.Container.Mounts {
		val, found := bm[mount.Name]
		if found {
			binds[mount.Name] = val
		} else {
			binds[mount.Name] = mountConfigToString(&mount)
			mountNames = append(mountNames, mount.Name)
		}
	}
	// Get all the container specific mount configs and apply
	// them as overrides for the global.
	for _, mount := range c.config.Filesystem.Mounts {
		val, found := bm[mount.Name]
		if found {
			binds[mount.Name] = val
		} else {
			binds[mount.Name] = mountConfigToString(&mount)
			mountNames = append(mountNames, mount.Name)
		}
	}

	// Convert the result to include only the bind mount strings.
	res := make([]string, 0)
	for _, mount := range mountNames {
		res = append(res, binds[mount])
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

func (c *Container) networkMode() dcontainer.NetworkMode {
	if len(c.endpoints) == 0 {
		return "none"
	}
	n := c.endpoints[0].network
	if n.mode == networkModeContainer {
		return dcontainer.NetworkMode(fmt.Sprintf("container:%s-%s", n.containerModeInfo.container.Group, n.containerModeInfo.container.Container))
	}
	return dcontainer.NetworkMode(n.name())
}

func (c *Container) publishedPorts() (nat.PortMap, nat.PortSet) {
	pMap := make(nat.PortMap)
	pSet := make(nat.PortSet)
	for _, p := range c.config.Network.PublishedPorts {
		natPort := nat.Port(fmt.Sprintf("%d/%s", p.ContainerPort, p.Protocol))
		pMap[natPort] = []nat.PortBinding{
			{
				HostIP:   p.HostIP,
				HostPort: strconv.Itoa(p.HostPort),
			},
		}
		pSet[natPort] = struct{}{}
	}
	if len(pSet) == 0 {
		return nil, nil
	}
	return pMap, pSet
}

func (c *Container) restartPolicy() dcontainer.RestartPolicy {
	mode := c.config.Lifecycle.RestartPolicy.Mode
	maxRetry := c.config.Lifecycle.RestartPolicy.MaxRetryCount
	if len(mode) == 0 {
		mode = c.globalConfig.Container.RestartPolicy.Mode
		maxRetry = c.globalConfig.Container.RestartPolicy.MaxRetryCount
	}

	rpm, err := docker.RestartPolicyModeFromString(mode)
	if err != nil {
		panic(fmt.Sprintf("unable to convert restart policy mode %s setting for container %s, reason: %v, possibly indicating a bug in the code", mode, c.Name(), err))
	}
	return dcontainer.RestartPolicy{Name: rpm, MaximumRetryCount: maxRetry}
}

func (c *Container) autoRemove() bool {
	return c.config.Lifecycle.AutoRemove
}

func (c *Container) capAddList() []string {
	return c.config.Security.CapAdd
}

func (c *Container) capDropList() []string {
	return c.config.Security.CapDrop
}

func (c *Container) dnsServers() []string {
	return c.config.Network.DNSServers
}

func (c *Container) dnsOptions() []string {
	return c.config.Network.DNSOptions
}

func (c *Container) dnsSearch() []string {
	d := c.config.Network.DNSSearch
	if len(d) == 0 {
		d = c.globalConfig.Container.DNSSearch
	}
	return d
}

func (c *Container) additionalUserGroups() []string {
	return c.config.User.AdditionalGroups
}

func (c *Container) privileged() bool {
	return c.config.Security.Privileged
}

func (c *Container) readOnlyRootfs() bool {
	return c.config.Filesystem.ReadOnlyRootfs
}

func (c *Container) tmpfsMounts() map[string]string {
	return nil
}

func (c *Container) shmSize() int64 {
	return 0
}

func (c *Container) sysctls() map[string]string {
	res := make(map[string]string, 0)
	for _, s := range c.config.Security.Sysctls {
		res[s.Key] = s.Value
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

func (c *Container) primaryNetworkEndpoint() map[string]*dnetwork.EndpointSettings {
	res := make(map[string]*dnetwork.EndpointSettings)
	if len(c.endpoints) > 0 && c.endpoints[0].network.mode == networkModeBridge {
		res[c.endpoints[0].network.name()] = &dnetwork.EndpointSettings{
			IPAMConfig: &dnetwork.EndpointIPAMConfig{
				IPv4Address: c.endpoints[0].ip,
			},
		}
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

func (c *Container) String() string {
	return fmt.Sprintf("Container{Name:%s}", c.Name())
}

func containerName(ct *config.ContainerReference) string {
	return fmt.Sprintf("%s-%s", ct.Group, ct.Container)
}

func containerMapToList(cm containerMap) containerList {
	res := make(containerList, 0, len(cm))
	for _, c := range cm {
		res = append(res, c)
	}

	// Return containers sorted by order. Group order takes higher priority
	// over container order within the same group. If two containers still
	// have the same order at both the group and container levels, then
	// the container name is used to lexicographically sort the containers.
	sort.Slice(res, func(i, j int) bool {
		c1 := res[i]
		c2 := res[j]
		if c1.group.config.Order == c2.group.config.Order {
			if c1.config.Lifecycle.Order == c2.config.Lifecycle.Order {
				return c1.Name() < c2.Name()
			}
			return c1.config.Lifecycle.Order < c2.config.Lifecycle.Order
		} else {
			return c1.group.config.Order < c2.group.config.Order
		}
	})
	return res
}

func mountConfigToString(v *config.MountConfig) string {
	if v.ReadOnly {
		return fmt.Sprintf("%s:%s:ro", v.Src, v.Dst)
	}
	return fmt.Sprintf("%s:%s", v.Src, v.Dst)
}
