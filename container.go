package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	dcontainer "github.com/docker/docker/api/types/container"
	dnetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

const (
	stopAndRemoveAttempts  = 5
	stopAndRemoveKillDelay = 1 * time.Second
)

type container struct {
	config        *ContainerConfig
	globalConfig  *GlobalConfig
	group         *containerGroup
	ips           networkContainerIPList
	allowedOnHost bool
}

type containerIP struct {
	network *network
	ip      string
}

type containerDockerConfigs struct {
	ContainerConfig *dcontainer.Config
	HostConfig      *dcontainer.HostConfig
	NetworkConfig   *dnetwork.NetworkingConfig
}

type networkContainerIPList []*containerIP
type containerList []*container
type containerSet map[ContainerReference]bool
type containerMap map[ContainerReference]*container
type containerDockerConfigMap map[ContainerReference]*containerDockerConfigs

func newContainer(group *containerGroup, config *ContainerConfig, globalConfig *GlobalConfig, ips networkContainerIPList, allowedOnHost bool) *container {
	return &container{
		config:        config,
		globalConfig:  globalConfig,
		group:         group,
		ips:           ips,
		allowedOnHost: allowedOnHost,
	}
}

func (c *container) isAllowedOnCurrentHost() bool {
	return c.allowedOnHost
}

func (c *container) start(ctx context.Context, docker *dockerClient, humanFriendlyHostName string) error {
	log(ctx).Debugf("Starting container %s ...", c.name())

	// Validate the container is allowed to run on the current host.
	if !c.isAllowedOnCurrentHost() {
		log(ctx).Warnf("Container %s not allowed to run on host '%s'", c.name(), humanFriendlyHostName)
		return nil
	}

	err := c.startInternal(ctx, docker)
	if err != nil {
		return logToErrorAndReturn(ctx, "Failed to start container %s, reason:%v", c.name(), err)
	}

	log(ctx).Infof("Started container %s", c.name())
	log(ctx).InfoEmpty()
	return nil
}

func (c *container) purge(ctx context.Context, docker *dockerClient) error {
	purged := false
	stoppedOnceAlready := false
	attemptsRemaining := stopAndRemoveAttempts

	for !purged && attemptsRemaining > 0 {
		attemptsRemaining--

		st, err := docker.getContainerState(ctx, c.name())
		if err != nil {
			return err
		}
		log(ctx).Debugf("Container %s current state: %s", c.name(), st)

		switch st {
		case containerStateNotFound:
			// Nothing to stop and/or remove.
			purged = true
		case containerStateRunning, containerStatePaused, containerStateRestarting:
			// Stop the container if not stopped already.
			if !stoppedOnceAlready {
				err = docker.stopContainer(ctx, c.name())
				if err != nil {
					return err
				}
				stoppedOnceAlready = true
			} else {
				// Kill the container next as a precaution and ignore any errors.
				_ = docker.killContainer(ctx, c.name())
				// Add a delay before checking the container state again.
				time.Sleep(stopAndRemoveKillDelay)
			}
		case containerStateCreated, containerStateExited, containerStateDead:
			// Directly remove the container.
			err = docker.removeContainer(ctx, c.name())
			if err != nil {
				return err
			}
		case containerStateRemoving:
			// Nothing to be done here, although this could lead to some
			// unknown handling in next steps.
			log(ctx).Warnf("container %s is in REMOVING state already, can lead to issues while we create the container next")
			// Add a delay before checking the container state again.
			time.Sleep(stopAndRemoveKillDelay)
		default:
			log(ctx).Fatalf("container %s is in an unsupported state %v, possibly indicating a bug in the code", st)
		}
	}

	if purged {
		return nil
	}

	// Check the container state one final time after exhausting all purge
	// attempts, and return the final error status based on that.
	st, err := docker.getContainerState(ctx, c.name())
	if err != nil {
		return err
	}
	if st != containerStateNotFound {
		return fmt.Errorf("failed to stop and remove container %s after %d attempts", c.name(), stopAndRemoveAttempts)
	}
	return nil
}

func (c *container) startInternal(ctx context.Context, docker *dockerClient) error {
	// 1. Execute any pre-start commands.
	// TODO: Implement this.

	// 2. Pull the container image.
	err := docker.pullImage(ctx, c.imageReference())
	if err != nil {
		return err
	}

	// 3. Purge (i.e. stop and remove) any previously existing containers
	// under the same name.
	err = c.purge(ctx, docker)
	if err != nil {
		return err
	}

	// 4. For the primary network interface of the container, create
	// the network for the container prior to creating the container
	// attached to this network.
	if len(c.ips) > 0 {
		// network.create(...) gracefully handles the case for when the
		// network exists already.
		err := c.ips[0].network.create(ctx, docker)
		if err != nil {
			return err
		}
		log(ctx).Debugf("Connecting container %s to network %s with IP %s at the time of container creation ...", c.name(), c.ips[0].network.name(), c.ips[0].ip)
	} else {
		log(ctx).Warnf("Container %s has no network endpoints configured, this is uncommon!", c.name())
	}

	// 5. Create the container.
	cConfig, hConfig, nConfig, err := c.generateDockerConfigs()
	if err != nil {
		return err
	}
	err = docker.createContainer(ctx, c.name(), cConfig, hConfig, nConfig)
	if err != nil {
		return err
	}

	// 6. For each non-primary network interface of the container, create
	// the network for the container if it doesn't exist already prior to
	// connecting the container to the network.
	for i := 1; i < len(c.ips); i++ {
		ip := c.ips[i]
		err := ip.network.create(ctx, docker)
		if err != nil {
			return err
		}
		err = ip.network.connectContainer(ctx, docker, c.name(), ip.ip)
		if err != nil {
			return err
		}
	}

	// 7. Start the created container.
	err = docker.startContainer(ctx, c.name())
	return err
}

func (c *container) generateDockerConfigs() (*dcontainer.Config, *dcontainer.HostConfig, *dnetwork.NetworkingConfig, error) {
	pMap, pSet := c.publishedPorts()
	return c.dockerContainerConfig(pSet), c.dockerHostConfig(pMap), c.dockerNetworkConfig(), nil
}

func (c *container) dockerContainerConfig(pSet nat.PortSet) *dcontainer.Config {
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

func (c *container) dockerHostConfig(pMap nat.PortMap) *dcontainer.HostConfig {
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

func (c *container) dockerNetworkConfig() *dnetwork.NetworkingConfig {
	ne := c.primaryNetworkEndpoint()
	if ne == nil {
		return nil
	}
	return &dnetwork.NetworkingConfig{
		EndpointsConfig: c.primaryNetworkEndpoint(),
	}
}

func (c *container) name() string {
	return containerName(&c.config.Info)
}

func (c *container) hostName() string {
	return c.config.Network.HostName
}

func (c *container) domainName() string {
	d := c.config.Network.DomainName
	if len(d) == 0 {
		d = c.globalConfig.Container.DomainName
	}
	return d
}

func (c *container) userAndGroup() string {
	u := c.config.User.User
	g := c.config.User.PrimaryGroup
	if len(g) > 0 {
		return fmt.Sprintf("%s:%s", u, g)
	}
	return u
}

func (c *container) attachToTty() bool {
	return c.config.Runtime.AttachToTty
}

func (c *container) envVars() []string {
	env := make(map[string]string, 0)
	envKeys := make([]string, 0)
	// TODO: Substitute global config env variables in the value fields.
	// TODO: Support invoking ValueCmd for evaluating the value.
	for _, e := range c.globalConfig.Container.Env {
		env[e.Var] = e.Value
		envKeys = append(envKeys, e.Var)
	}
	for _, e := range c.config.Runtime.Env {
		if _, ok := env[e.Var]; !ok {
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

func (c *container) args() []string {
	return c.config.Runtime.Args
}

func (c *container) entrypoint() []string {
	return c.config.Runtime.Entrypoint
}

func (c *container) isNetworkDisabled() bool {
	return false
}

func (c *container) labels() map[string]string {
	res := make(map[string]string, 0)
	for _, l := range c.config.Metadata.Labels {
		res[l.Name] = l.Value
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

func (c *container) stopSignal() string {
	return c.config.Lifecycle.StopSignal
}

func (c *container) stopTimeout() *int {
	t := c.config.Lifecycle.StopTimeout
	if t == 0 {
		t = c.globalConfig.Container.StopTimeout
	}
	return &t
}

func (c *container) imageReference() string {
	return c.config.Image.Image
}

func (c *container) bindMounts() []string {
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
		val, ok := bm[mount.Name]
		if ok {
			binds[mount.Name] = val
		} else {
			binds[mount.Name] = mountConfigToString(&mount)
			mountNames = append(mountNames, mount.Name)
		}
	}
	// Get all the container specific mount configs and apply
	// them as overrides for the global.
	for _, mount := range c.config.Filesystem.Mounts {
		val, ok := bm[mount.Name]
		if ok {
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

func (c *container) networkMode() dcontainer.NetworkMode {
	if len(c.ips) == 0 {
		return "none"
	}
	if c.ips[0].network.mode == networkModeContainer {
		return dcontainer.NetworkMode(fmt.Sprintf("container:%s", c.ips[0].network))
	}
	return dcontainer.NetworkMode(c.ips[0].network.name())
}

func (c *container) publishedPorts() (nat.PortMap, nat.PortSet) {
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

func (c *container) restartPolicy() dcontainer.RestartPolicy {
	mode := c.config.Lifecycle.RestartPolicy.Mode
	maxRetry := c.config.Lifecycle.RestartPolicy.MaxRetryCount
	if len(mode) == 0 {
		mode = c.globalConfig.Container.RestartPolicy.Mode
		maxRetry = c.globalConfig.Container.RestartPolicy.MaxRetryCount
	}

	rpm, err := restartPolicyModeFromString(mode)
	if err != nil {
		panic(fmt.Sprintf("unable to convert restart policy mode %s setting for container %s, reason: %v, possibly indicating a bug in the code", mode, c.name(), err))
	}
	return dcontainer.RestartPolicy{Name: rpm, MaximumRetryCount: maxRetry}
}

func (c *container) autoRemove() bool {
	return c.config.Lifecycle.AutoRemove
}

func (c *container) capAddList() []string {
	return c.config.Security.CapAdd
}

func (c *container) capDropList() []string {
	return c.config.Security.CapDrop
}

func (c *container) dnsServers() []string {
	return c.config.Network.DNSServers
}

func (c *container) dnsOptions() []string {
	return c.config.Network.DNSOptions
}

func (c *container) dnsSearch() []string {
	d := c.config.Network.DNSSearch
	if len(d) == 0 {
		d = c.globalConfig.Container.DNSSearch
	}
	return d
}

func (c *container) additionalUserGroups() []string {
	return c.config.User.AdditionalGroups
}

func (c *container) privileged() bool {
	return c.config.Security.Privileged
}

func (c *container) readOnlyRootfs() bool {
	return c.config.Filesystem.ReadOnlyRootfs
}

func (c *container) tmpfsMounts() map[string]string {
	return nil
}

func (c *container) shmSize() int64 {
	return 0
}

func (c *container) sysctls() map[string]string {
	res := make(map[string]string, 0)
	for _, s := range c.config.Security.Sysctls {
		res[s.Key] = s.Value
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

func (c *container) primaryNetworkEndpoint() map[string]*dnetwork.EndpointSettings {
	res := make(map[string]*dnetwork.EndpointSettings)
	if len(c.ips) > 0 && c.ips[0].network.mode == networkModeBridge {
		res[c.ips[0].network.name()] = &dnetwork.EndpointSettings{
			IPAMConfig: &dnetwork.EndpointIPAMConfig{
				IPv4Address: c.ips[0].ip,
			},
		}
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

func (c *container) String() string {
	return fmt.Sprintf("Container{Name:%s}", c.name())
}

func (c containerMap) String() string {
	return stringifyMap(c)
}

func containerName(ct *ContainerReference) string {
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
				return c1.name() < c2.name()
			}
			return c1.config.Lifecycle.Order < c2.config.Lifecycle.Order
		} else {
			return c1.group.config.Order < c2.group.config.Order
		}
	})
	return res
}

func mountConfigToString(v *MountConfig) string {
	if v.ReadOnly {
		return fmt.Sprintf("%s:%s:ro", v.Src, v.Dst)
	}
	return fmt.Sprintf("%s:%s", v.Src, v.Dst)
}
