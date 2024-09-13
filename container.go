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
	config       *ContainerConfig
	globalConfig *GlobalConfig
	group        *containerGroup
	ips          networkContainerIPList
}

type containerIP struct {
	network *network
	ip      string
}

type containerMap map[string]*container
type containerList []*container
type networkContainerIPList []*containerIP

func newContainer(group *containerGroup, config *ContainerConfig) *container {
	ct := container{
		config:       config,
		globalConfig: &group.deployment.config.Global,
		group:        group,
	}
	cName := config.Name
	gName := group.name()

	var ips networkContainerIPList
	for _, n := range group.deployment.networks {
		if n.mode == networkModeBridge {
			for _, c := range n.bridgeModeConfig.Containers {
				if c.Container.Group == gName && c.Container.Container == cName {
					ips = append(ips, newBridgeModeContainerIP(n, c.IP))
					break
				}
			}
		} else if n.mode == networkModeContainer {
			for _, c := range n.containerModeConfig.Containers {
				if c.Group == gName && c.Container == cName {
					ips = append(ips, newContainerModeContainerIP(n))
					break
				}
			}
		}
	}

	// Sort the networks by priority (i.e. lowest priority is the primary
	// network interface for the container).
	sort.Slice(ips, func(i, j int) bool {
		n1 := ips[i].network
		n2 := ips[j].network

		if n1.mode != n2.mode {
			log.Fatalf("Container %s has networks of different types which is unsupported", ct.name())
		}
		if n1.mode == networkModeBridge {
			if n1.bridgeModeConfig.Priority == n2.bridgeModeConfig.Priority {
				log.Fatalf("Container %s is connected to two bridge mode networks of same priority %d which is unsupported", ct.name(), n1.bridgeModeConfig.Priority)
			}
			return n1.bridgeModeConfig.Priority < n2.bridgeModeConfig.Priority
		} else {
			if n1.containerModeConfig.Priority == n2.containerModeConfig.Priority {
				log.Fatalf("Container %s is connected to two container mode networks of same priority %d which is unsupported", ct.name(), n1.containerModeConfig.Priority)
			}
			return n1.containerModeConfig.Priority < n2.containerModeConfig.Priority
		}
	})

	ct.ips = ips
	return &ct
}

func (c *container) isAllowedOnCurrentHost() bool {
	return c.group.deployment.host.allowedContainers[c.name()]
}

func (c *container) start(ctx context.Context, docker *dockerClient) error {
	log.Debugf("Starting container %s ...", c.name())

	// Validate the container is allowed to run on the current host.
	if !c.isAllowedOnCurrentHost() {
		log.Warnf("Container %s not allowed to run on host '%s'", c.name(), c.group.deployment.host.humanFriendlyHostName)
		return nil
	}

	err := c.startInternal(ctx, docker)
	if err != nil {
		return logToErrorAndReturn("Failed to start container %s, reason:%v", c.name(), err)
	}

	log.Infof("Started container %s", c.name())
	log.InfoEmpty()
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
		log.Debugf("Container %s current state: %s", c.name(), st)

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
			log.Warnf("container %s is in REMOVING state already, can lead to issues while we create the container next")
			// Add a delay before checking the container state again.
			time.Sleep(stopAndRemoveKillDelay)
		default:
			log.Fatalf("container %s is in an unsupported state %v, possibly indicating a bug in the code", st)
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
		log.Debugf("Connecting container %s to network %s with IP %s at the time of container creation ...", c.name(), c.ips[0].network.name(), c.ips[0].ip)
	} else {
		log.Warnf("Container %s has no network endpoints configured, this is uncommon!", c.name())
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
	cConfig := dcontainer.Config{
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
	hConfig := dcontainer.HostConfig{
		Binds:          c.volumeBindMounts(),
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
	nConfig := dnetwork.NetworkingConfig{
		EndpointsConfig: c.primaryNetworkEndpoint(),
	}
	return &cConfig, &hConfig, &nConfig, nil
}

func (c *container) name() string {
	return containerName(c.group.name(), c.config.Name)
}

func (c *container) hostName() string {
	return c.config.HostName
}

func (c *container) domainName() string {
	d := c.config.DomainName
	if len(d) == 0 {
		d = c.globalConfig.Container.DomainName
	}
	return d
}

func (c *container) userAndGroup() string {
	u := c.config.User
	g := c.config.PrimaryUserGroup
	if len(g) > 0 {
		return fmt.Sprintf("%s:%s", u, g)
	}
	return u
}

func (c *container) attachToTty() bool {
	return c.config.AttachToTty
}

func (c *container) envVars() []string {
	env := make(map[string]string, 0)
	// TODO: Substitute global config env variables in the value fields.
	// TODO: Support invoking ValueCmd for evaluating the value.
	for _, e := range c.globalConfig.Container.Env {
		env[e.Var] = e.Value
	}
	for _, e := range c.config.Env {
		env[e.Var] = e.Value
	}

	res := make([]string, 0)
	for k, v := range env {
		res = append(res, fmt.Sprintf("%s=%s", k, v))
	}
	return res
}

func (c *container) args() []string {
	return c.config.Args
}

func (c *container) entrypoint() []string {
	return c.config.Entrypoint
}

func (c *container) isNetworkDisabled() bool {
	return false
}

func (c *container) labels() map[string]string {
	res := make(map[string]string, 0)
	for _, l := range c.config.Labels {
		res[l.Name] = l.Value
	}
	return res
}

func (c *container) stopSignal() string {
	return c.config.StopSignal
}

func (c *container) stopTimeout() *int {
	t := c.config.StopTimeout
	if t == 0 {
		t = c.globalConfig.Container.StopTimeout
	}
	return &t
}

func (c *container) imageReference() string {
	return c.config.Image
}

func (c *container) volumeBindMounts() []string {
	// TODO: Do this once for the entire deployment and reuse it.
	vd := make(map[string]string, 0)
	for _, v := range c.globalConfig.VolumeDefs {
		vd[v.Name] = volumeConfigToString(&v)
	}

	binds := make(map[string]string, 0)
	// Get all the global container config volumes.
	// TODO: Do this once for the entire deployment and reuse it.
	for _, vol := range c.globalConfig.Container.Volumes {
		val, ok := vd[vol.Name]
		if ok {
			binds[vol.Name] = val
		} else {
			binds[vol.Name] = volumeConfigToString(&vol)
		}
	}
	// Get all the container specific volume configs and apply
	// them as overrides for the global.
	for _, vol := range c.config.Volumes {
		val, ok := vd[vol.Name]
		if ok {
			binds[vol.Name] = val
		} else {
			binds[vol.Name] = volumeConfigToString(&vol)
		}
	}

	// Convert the result to include only the volume bind mount strings.
	res := make([]string, 0)
	for _, val := range binds {
		res = append(res, val)
	}
	return res
}

func (c *container) networkMode() dcontainer.NetworkMode {
	if len(c.ips) > 0 {
		if c.ips[0].network.mode == networkModeContainer {
			return dcontainer.NetworkMode(fmt.Sprintf("container:%s", c.ips[0].network))
		}
		return dcontainer.NetworkMode(c.ips[0].network.name())
	}
	return "none"
}

func (c *container) publishedPorts() (nat.PortMap, nat.PortSet) {
	pMap := make(nat.PortMap)
	pSet := make(nat.PortSet)
	for _, p := range c.config.PublishedPorts {
		natPort := nat.Port(fmt.Sprintf("%d/%s", p.ContainerPort, p.Proto))
		pMap[natPort] = []nat.PortBinding{
			{
				HostIP:   p.HostIP,
				HostPort: strconv.Itoa(p.HostPort),
			},
		}
		pSet[natPort] = struct{}{}
	}
	return pMap, pSet
}

func (c *container) restartPolicy() dcontainer.RestartPolicy {
	pol := c.config.RestartPolicy
	if len(pol) == 0 {
		pol = c.globalConfig.Container.RestartPolicy
	}

	// TODO: Perform better validation of the restart policy config setting
	// prior to directly covnerting it to RestartPolicyMode.
	return dcontainer.RestartPolicy{Name: dcontainer.RestartPolicyMode(pol)}
}

func (c *container) autoRemove() bool {
	return c.config.AutoRemove
}

func (c *container) capAddList() []string {
	return c.config.CapAdd
}

func (c *container) capDropList() []string {
	return c.config.CapDrop
}

func (c *container) dnsServers() []string {
	return c.config.DNSServers
}

func (c *container) dnsOptions() []string {
	return c.config.DNSOptions
}

func (c *container) dnsSearch() []string {
	d := c.config.DNSSearch
	if len(d) == 0 {
		d = c.globalConfig.Container.DNSSearch
	}
	return d
}

func (c *container) additionalUserGroups() []string {
	return c.config.AdditionalUserGroups
}

func (c *container) privileged() bool {
	return c.config.Privileged
}

func (c *container) readOnlyRootfs() bool {
	return c.config.ReadOnlyRootfs
}

func (c *container) tmpfsMounts() map[string]string {
	return nil
}

func (c *container) shmSize() int64 {
	return 0
}

func (c *container) sysctls() map[string]string {
	res := make(map[string]string, 0)
	for _, s := range c.config.Sysctls {
		res[s.Key] = s.Value
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
	return res
}

func (c *container) String() string {
	return fmt.Sprintf("Container{Name:%s}", c.name())
}

func (c containerMap) String() string {
	return stringifyMap(c)
}

func newBridgeModeContainerIP(network *network, ip string) *containerIP {
	return &containerIP{network: network, ip: ip}
}

func newContainerModeContainerIP(network *network) *containerIP {
	return &containerIP{network: network}
}

func containerName(group string, container string) string {
	return fmt.Sprintf("%s-%s", group, container)
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
			if c1.config.Order == c2.config.Order {
				return c1.name() < c2.name()
			}
			return c1.config.Order < c2.config.Order
		} else {
			return c1.group.config.Order < c2.group.config.Order
		}
	})
	return res
}

func volumeConfigToString(v *VolumeConfig) string {
	if v.ReadOnly {
		return fmt.Sprintf("%s:%s:ro", v.Src, v.Dst)
	}
	return fmt.Sprintf("%s:%s", v.Src, v.Dst)
}
