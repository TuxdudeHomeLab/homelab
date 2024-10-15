package deployment

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	dcontainer "github.com/docker/docker/api/types/container"
	dnetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/tuxdudehomelab/homelab/internal/cmdexec"
	"github.com/tuxdudehomelab/homelab/internal/config"
	"github.com/tuxdudehomelab/homelab/internal/docker"
	"github.com/tuxdudehomelab/homelab/internal/utils"
)

const (
	// Delay between successive purge (stop and remove) kill attempts.
	purgeKillDelay = 20 * time.Millisecond
)

type Container struct {
	config        *config.Container
	globalConfig  *config.Global
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
type ContainerList []*Container
type containerSet map[config.ContainerReference]bool
type containerMap map[config.ContainerReference]*Container
type containerDockerConfigMap map[config.ContainerReference]*containerDockerConfigs

func newContainer(group *ContainerGroup, config *config.Container, globalConfig *config.Global, endpoints networkEndpointList, allowedOnHost bool) *Container {
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

func (c *Container) Start(ctx context.Context, dc *docker.Client) (bool, error) {
	log(ctx).Debugf("Starting container %s ...", c.Name())

	// Validate the container is allowed to run on the current host.
	if !c.isAllowedOnCurrentHost() {
		return false, nil
	}

	err := c.startInternal(ctx, dc)
	if err != nil {
		return false, utils.LogToErrorAndReturn(ctx, "Failed to start container %s, reason:%v", c.Name(), err)
	}

	log(ctx).Debugf("Started container %s", c.Name())
	log(ctx).InfoEmpty()
	return true, nil
}

func (c *Container) Stop(ctx context.Context, dc *docker.Client) (bool, error) {
	log(ctx).Debugf("Stopping container %s ...", c.Name())

	stopped, st, err := c.stopInternal(ctx, dc)
	if err != nil {
		return false, utils.LogToErrorAndReturn(ctx, "Failed to stop container %s, reason:%v", c.Name(), err)
	}

	if stopped {
		if st != docker.ContainerStateRunning && st != docker.ContainerStatePaused && st != docker.ContainerStateRestarting {
			log(ctx).Warnf("Container %s cannot be stopped since it is in state %s", c.Name(), st)
		} else {
			log(ctx).Debugf("Stopped container %s", c.Name())
		}
	} else {
		if st == docker.ContainerStateNotFound {
			log(ctx).Warnf("Container %s cannot be stopped since it was not found", c.Name())
		} else {
			log(ctx).Warnf("Container %s cannot be stopped since it is in state %s", c.Name(), st)
		}
	}
	log(ctx).InfoEmpty()

	return stopped, nil
}

func (c *Container) Purge(ctx context.Context, dc *docker.Client) (bool, error) {
	log(ctx).Debugf("Purging container %s ...", c.Name())

	purged, err := c.purgeInternal(ctx, dc)
	if err != nil {
		return false, utils.LogToErrorAndReturn(ctx, "Failed to purge container %s, reason:%v", c.Name(), err)
	}

	if purged {
		log(ctx).Debugf("Purged container %s", c.Name())
		log(ctx).InfoEmpty()
	}

	return purged, nil
}

func (c *Container) startInternal(ctx context.Context, dc *docker.Client) error {
	// 1. Execute start pre-hook command if specified.
	if len(c.config.Lifecycle.StartPreHook) > 0 {
		log(ctx).Infof("Output from start pre-hook for container %s >>>", c.Name())
		cmd := c.config.Lifecycle.StartPreHook
		exec := cmdexec.MustExecutor(ctx)
		out, err := exec.Run(cmd[0], cmd[1:]...)
		log(ctx).Printf("%s", strings.TrimSpace(out))
		if err != nil {
			return fmt.Errorf("encountered error while running the start pre-hook for container %s, reason: %w", c.Name(), err)
		}
	}

	// 2. Pull the container image.
	if !c.config.Image.SkipImagePull {
		err := dc.PullImage(ctx, c.imageReference())
		if err != nil {
			if !c.config.Image.IgnoreImagePullFailures {
				return err
			}
			log(ctx).Warnf("Ignoring - Image pull for container %s failed, reason: %v", c.Name(), err)
		}
	}

	// 3. Purge (i.e. stop and remove) any previously existing containers
	// under the same name.
	purged, err := c.purgeInternal(ctx, dc)
	if err != nil {
		return err
	}
	if purged {
		log(ctx).Debugf("Purged container %s", c.Name())
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
	log(ctx).Infof("Creating container %s", c.Name())
	cdc := c.generateDockerConfigs()
	err = dc.CreateContainer(ctx, c.Name(), cdc.ContainerConfig, cdc.HostConfig, cdc.NetworkConfig)
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
	log(ctx).Infof("Starting container %s", c.Name())
	err = dc.StartContainer(ctx, c.Name())
	return err
}

func (c *Container) stopInternal(ctx context.Context, dc *docker.Client) (bool, docker.ContainerState, error) {
	st, err := dc.GetContainerState(ctx, c.Name())
	if err != nil {
		return false, docker.ContainerStateUnknown, err
	}
	log(ctx).Debugf("stopInternal - Container %s current state: %s", c.Name(), st)

	switch st {
	case docker.ContainerStateNotFound:
		// Nothing to stop.
		return false, st, nil
	case docker.ContainerStateRunning, docker.ContainerStatePaused, docker.ContainerStateRestarting:
		// 2. Pull the container image before stopping if requested.
		if c.config.Image.PullImageBeforeStop {
			err := dc.PullImage(ctx, c.imageReference())
			if err != nil {
				if !c.config.Image.IgnoreImagePullFailures {
					return false, st, err
				}
				log(ctx).Warnf("Ignoring - Image pull for container %s failed, reason: %v", c.Name(), err)
			}
		}

		// Stop the container.
		log(ctx).Infof("Stopping container %s", c.Name())
		if err := dc.StopContainer(ctx, c.Name()); err != nil {
			return false, st, err
		}
		return true, st, nil
	case docker.ContainerStateCreated, docker.ContainerStateExited, docker.ContainerStateDead, docker.ContainerStateRemoving:
		// Container is already stopped in this state.
		return true, st, nil
	default:
		log(ctx).Fatalf("container %s is in an unsupported state %v, possibly indicating a bug in the code", c.Name(), st)
	}

	return false, st, fmt.Errorf("failed to stop container %s since it is in state %s", c.Name(), st)
}

func (c *Container) purgeInternal(ctx context.Context, dc *docker.Client) (bool, error) {
	// Stop the container once (if possible).
	stopped, _, err := c.stopInternal(ctx, dc)
	if err != nil {
		return false, err
	}

	if !stopped {
		// The container was not found, hence it could not be stopped.
		// Nothing to do for purge either.
		return false, nil
	}

	purged := false
	// We will attempt killing the container attemptsRemaining times
	// in the worst case before giving up. The one extra attempt is to
	// remove the container which has been terminated.
	stopAndRemoveAttempts := dc.ContainerPurgeKillAttempts() + 1
	attemptsRemaining := stopAndRemoveAttempts

	for !purged && attemptsRemaining > 0 {
		attemptsRemaining--

		st, err := dc.GetContainerState(ctx, c.Name())
		if err != nil {
			return false, err
		}
		log(ctx).Debugf("purgeInternal - Container %s current state: %s", c.Name(), st)

		switch st {
		case docker.ContainerStateNotFound:
			// Nothing further to do for purge.
			purged = true
		case docker.ContainerStateRunning, docker.ContainerStatePaused, docker.ContainerStateRestarting:
			// The container was already stopped (if possible).
			// Kill the container next as a precaution and ignore any errors.
			log(ctx).Infof("Killing container %s", c.Name())
			_ = dc.KillContainer(ctx, c.Name())
			// Add a delay before checking the container state again.
			time.Sleep(purgeKillDelay)
		case docker.ContainerStateCreated, docker.ContainerStateExited, docker.ContainerStateDead:
			// Directly remove the container.
			log(ctx).Infof("Removing container %s", c.Name())
			err = dc.RemoveContainer(ctx, c.Name())
			if err != nil {
				return false, err
			}
		case docker.ContainerStateRemoving:
			// Nothing to be done here, although this could lead to some
			// unknown handling in next steps.
			log(ctx).Warnf("container %s is in REMOVING state already, can lead to issues for any further operations including creating container with the same name", c.Name())
			// Add a delay before checking the container state again.
			time.Sleep(purgeKillDelay)
		default:
			log(ctx).Fatalf("container %s is in an unsupported state %v, possibly indicating a bug in the code", c.Name(), st)
		}
	}

	if purged {
		return true, nil
	}

	// Check the container state one final time after exhausting all purge
	// attempts, and return the final error status based on that.
	st, err := dc.GetContainerState(ctx, c.Name())
	if err != nil {
		return false, err
	}
	if st != docker.ContainerStateNotFound {
		return false, fmt.Errorf("failed to purge container %s after %d attempts", c.Name(), stopAndRemoveAttempts)
	}
	return true, nil
}

func (c *Container) generateDockerConfigs() *containerDockerConfigs {
	pMap, pSet := c.publishedPorts()
	return &containerDockerConfigs{
		ContainerConfig: c.dockerContainerConfig(pSet),
		HostConfig:      c.dockerHostConfig(pMap),
		NetworkConfig:   c.dockerNetworkConfig(),
	}
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
		Healthcheck:     c.healthCheck(),
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
		ExtraHosts:     c.extraHosts(),
		GroupAdd:       c.additionalUserGroups(),
		Privileged:     c.privileged(),
		ReadonlyRootfs: c.readOnlyRootfs(),
		Tmpfs:          c.tmpfsMounts(),
		ShmSize:        c.shmSize(),
		Sysctls:        c.sysctls(),
		Resources:      c.resources(),
	}
}

func (c *Container) dockerNetworkConfig() *dnetwork.NetworkingConfig {
	ne := c.primaryNetworkEndpoint()
	if ne == nil {
		return nil
	}
	return &dnetwork.NetworkingConfig{
		EndpointsConfig: ne,
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

func (c *Container) healthCheck() *dcontainer.HealthConfig {
	h := c.config.Health
	res := &dcontainer.HealthConfig{}
	empty := true

	if len(h.Cmd) > 0 {
		cmd := []string{"CMD"}
		cmd = append(cmd, h.Cmd...)
		res.Test = cmd
		empty = false
	}
	if len(h.Interval) > 0 {
		res.Interval = utils.MustParseDuration(h.Interval)
		empty = false
	}
	if len(h.Timeout) > 0 {
		res.Timeout = utils.MustParseDuration(h.Timeout)
		empty = false
	}
	if len(h.StartPeriod) > 0 {
		res.StartPeriod = utils.MustParseDuration(h.StartPeriod)
		empty = false
	}
	if len(h.StartInterval) > 0 {
		res.StartInterval = utils.MustParseDuration(h.StartInterval)
		empty = false
	}
	if h.Retries != 0 {
		res.Retries = h.Retries
		empty = false
	}

	if empty {
		return nil
	}
	return res
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
	if t == 0 {
		return nil
	}
	return &t
}

func (c *Container) imageReference() string {
	return c.config.Image.Image
}

func (c *Container) bindMounts() []string {
	// TODO: Do this once for the entire deployment and reuse it.
	globalMountDefs := make(map[string]string, 0)
	for _, md := range c.globalConfig.MountDefs {
		globalMountDefs[md.Name] = mountConfigToString(&md)
	}

	containerMounts := make(map[string]string, 0)
	containerMountNames := make([]string, 0)
	// TODO: Do this once for the entire deployment and reuse it.
	// Get all the global container config mounts.
	for _, mount := range c.globalConfig.Container.Mounts {
		if val, found := globalMountDefs[mount.Name]; found {
			containerMounts[mount.Name] = val
		} else {
			containerMounts[mount.Name] = mountConfigToString(&mount)
		}
		containerMountNames = append(containerMountNames, mount.Name)
	}
	// Get all the container specific mount configs and apply
	// them as overrides for the global.
	for _, mount := range c.config.Filesystem.Mounts {
		if val, found := globalMountDefs[mount.Name]; found {
			containerMounts[mount.Name] = val
		} else {
			containerMounts[mount.Name] = mountConfigToString(&mount)
		}
		containerMountNames = append(containerMountNames, mount.Name)
	}

	// Convert the result to include only the bind mount strings.
	res := make([]string, 0)
	for _, mount := range containerMountNames {
		res = append(res, containerMounts[mount])
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
		natPort := nat.Port(fmt.Sprintf("%s/%s", p.ContainerPort, p.Protocol))
		pMap[natPort] = []nat.PortBinding{
			{
				HostIP:   p.HostIP,
				HostPort: p.HostPort,
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
	if len(mode) == 0 && maxRetry == 0 {
		return dcontainer.RestartPolicy{}
	}

	return dcontainer.RestartPolicy{
		Name:              docker.MustRestartPolicyModeFromString(mode),
		MaximumRetryCount: maxRetry,
	}
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

func (c *Container) extraHosts() []string {
	return c.config.Network.ExtraHosts
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
	if len(c.config.Runtime.ShmSize) == 0 {
		return 0
	}
	return utils.MustParseRAMInBytes(c.config.Runtime.ShmSize)
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

func (c *Container) resources() dcontainer.Resources {
	devices := &c.config.Filesystem.Devices
	var devs []dcontainer.DeviceMapping
	for _, d := range slices.Concat(devices.Static, devices.Dynamic) {
		m := dcontainer.DeviceMapping{}
		m.PathOnHost = d.Src
		if len(d.Dst) > 0 {
			m.PathInContainer = d.Dst
		} else {
			m.PathInContainer = d.Src
		}
		perms := strings.Builder{}
		if !d.DisallowRead {
			perms.WriteRune('r')
		}
		if !d.DisallowWrite {
			perms.WriteRune('w')
		}
		if !d.DisallowMknod {
			perms.WriteRune('m')
		}
		m.CgroupPermissions = perms.String()
		devs = append(devs, m)
	}
	return dcontainer.Resources{
		Devices: devs,
	}
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

func containerMapToList(cm containerMap) ContainerList {
	res := make(ContainerList, 0, len(cm))
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

func mountConfigToString(v *config.Mount) string {
	if v.ReadOnly {
		return fmt.Sprintf("%s:%s:ro", v.Src, v.Dst)
	}
	return fmt.Sprintf("%s:%s", v.Src, v.Dst)
}
