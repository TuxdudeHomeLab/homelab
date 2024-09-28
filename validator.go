package main

import (
	"context"
	"fmt"
	"net/netip"
	"os"
	"sort"
	"time"

	"github.com/docker/go-units"
)

func validateGlobalConfig(ctx context.Context, parentEnv *configEnv, config *GlobalConfig) (*configEnv, error) {
	if err := validateBaseDir(config.BaseDir); err != nil {
		return nil, err
	}

	newEnv, err := validateConfigEnv(ctx, parentEnv, config.Env, "global config")
	if err != nil {
		return nil, err
	}

	if err := validateMountsConfig(config.MountDefs, nil, nil, "global config mount defs"); err != nil {
		return nil, err
	}

	if err := validateGlobalContainerConfig(&config.Container, config.MountDefs); err != nil {
		return nil, err
	}

	return newEnv, nil
}

func validateBaseDir(baseDir string) error {
	if len(baseDir) == 0 {
		return fmt.Errorf("homelab base directory cannot be empty")
	}
	pathStat, err := os.Stat(baseDir)
	if err != nil {
		return fmt.Errorf("os.Stat() failed on homelab base directory path, reason: %w", err)
	}
	if !pathStat.IsDir() {
		return fmt.Errorf("homelab base directory path %s must be a directory", baseDir)
	}
	return nil
}

func validateConfigEnv(ctx context.Context, parentEnv *configEnv, config []ConfigEnv, location string) (*configEnv, error) {
	envs := envMap{}
	envOrder := make([]string, 0)
	for _, e := range config {
		if len(e.Var) == 0 {
			return nil, fmt.Errorf("empty env var in %s", location)
		}
		if _, found := envs[e.Var]; found {
			return nil, fmt.Errorf("env var %s specified more than once in %s", e.Var, location)
		}

		if len(e.Value) == 0 && len(e.ValueCommand) == 0 {
			return nil, fmt.Errorf("neither value nor valueCommand specified for env var %s in %s", e.Var, location)
		}
		if len(e.Value) > 0 && len(e.ValueCommand) > 0 {
			return nil, fmt.Errorf("exactly one of value or valueCommand must be specified for env var %s in %s", e.Var, location)
		}

		if len(e.Value) > 0 {
			envs[e.Var] = e.Value
		} else {
			envs[e.Var] = e.ValueCommand
		}
		envOrder = append(envOrder, e.Var)
	}
	return parentEnv.override(ctx, envs, envOrder), nil
}

func validateContainerEnv(config []ContainerEnv, location string) error {
	envs := stringSet{}
	for _, e := range config {
		if len(e.Var) == 0 {
			return fmt.Errorf("empty env var in %s", location)
		}
		if _, found := envs[e.Var]; found {
			return fmt.Errorf("env var %s specified more than once in %s", e.Var, location)
		}
		envs[e.Var] = struct{}{}

		if len(e.Value) == 0 && len(e.ValueCommand) == 0 {
			return fmt.Errorf("neither value nor valueCommand specified for env var %s in %s", e.Var, location)
		}
		if len(e.Value) > 0 && len(e.ValueCommand) > 0 {
			return fmt.Errorf("exactly one of value or valueCommand must be specified for env var %s in %s", e.Var, location)
		}
	}
	return nil
}

func validateLabelsConfig(config []LabelConfig, location string) error {
	labels := stringSet{}
	for _, l := range config {
		if len(l.Name) == 0 {
			return fmt.Errorf("empty label name in %s", location)
		}
		if _, found := labels[l.Name]; found {
			return fmt.Errorf("label name %s specified more than once in %s", l.Name, location)
		}
		labels[l.Name] = struct{}{}

		if len(l.Value) == 0 {
			return fmt.Errorf("empty label value for label %s in %s", l.Name, location)
		}
	}
	return nil
}

func validateMountsConfig(config, commonConfig, globalDefs []MountConfig, location string) error {
	// First build a map of the mounts from the globalDefs (which should
	// already have been validated).
	globalMountDefs := stringSet{}
	for _, m := range globalDefs {
		globalMountDefs[m.Name] = struct{}{}
	}

	// Build a map of the mounts from the commonConfig next which acts
	// as the first set of mounts to apply. These should also have been
	// validated prior and hence we don't validate them here again.
	mounts := stringSet{}
	for _, m := range commonConfig {
		mounts[m.Name] = struct{}{}
	}

	// Finally iterate and validate the mounts in the current level config.
	for _, m := range config {
		if len(m.Name) == 0 {
			return fmt.Errorf("mount name cannot be empty in %s", location)
		}
		if _, found := mounts[m.Name]; found {
			return fmt.Errorf("mount name %s defined more than once in %s", m.Name, location)
		}
		mounts[m.Name] = struct{}{}

		if len(m.Type) == 0 && len(m.Src) == 0 && len(m.Dst) == 0 && !m.ReadOnly {
			// This is a mount with just the name. Match this against the
			// global mount defs.
			if _, found := globalMountDefs[m.Name]; !found {
				return fmt.Errorf("mount specified by just the name %s not found in defs in %s", m.Name, location)
			}
			// No further validation needed for a mount referencing a def.
			return nil
		}

		if m.Type != "bind" {
			return fmt.Errorf("unsupported mount type %s for mount %s in %s", m.Type, m.Name, location)
		}
		if len(m.Src) == 0 {
			return fmt.Errorf("mount name %s cannot have an empty value for src in %s", m.Name, location)
		}
		if len(m.Dst) == 0 {
			return fmt.Errorf("mount name %s cannot have an empty value for dst in %s", m.Name, location)
		}
		if len(m.Options) > 0 {
			return fmt.Errorf("bind mount name %s cannot specify options in %s", m.Name, location)
		}
	}
	return nil
}

func validateDevicesConfig(devices []DeviceConfig, location string) error {
	for _, d := range devices {
		if len(d.Src) == 0 {
			return fmt.Errorf("device src cannot be empty in %s", location)
		}
	}
	return nil
}

func validatePublishedPortsConfig(ports []PublishedPortConfig, location string) error {
	for _, p := range ports {
		if p.ContainerPort <= 0 {
			return fmt.Errorf("published container port %d cannot be non-positive in %s", p.ContainerPort, location)
		}
		if p.Protocol != "tcp" && p.Protocol != "udp" {
			return fmt.Errorf("published container port %d specifies an invalid protocol %s in %s", p.ContainerPort, p.Protocol, location)
		}
		if len(p.HostIP) == 0 {
			return fmt.Errorf("published host IP cannot be empty for container port %d in %s", p.ContainerPort, location)
		}
		if _, err := netip.ParseAddr(p.HostIP); err != nil {
			return fmt.Errorf("published host IP %s for container port %d is invalid in %s, reason: %w", p.HostIP, p.ContainerPort, location, err)
		}
		if p.HostPort <= 0 {
			return fmt.Errorf("published host port %d cannot be non-positive in %s", p.HostPort, location)
		}
	}
	return nil
}

func validateSysctlsConfig(sysctls []SysctlConfig, location string) error {
	keys := stringSet{}
	for _, s := range sysctls {
		if len(s.Key) == 0 {
			return fmt.Errorf("empty sysctl key in %s", location)
		}
		if _, found := keys[s.Key]; found {
			return fmt.Errorf("sysctl key %s specified more than once in %s", s.Key, location)
		}
		keys[s.Key] = struct{}{}

		if len(s.Value) == 0 {
			return fmt.Errorf("empty sysctl value for sysctl %s in %s", s.Key, location)
		}
	}
	return nil
}

func validateHealthConfig(config *ContainerHealthConfig, location string) error {
	if config.Retries < 0 {
		return fmt.Errorf("health check retries %d cannot be negative in %s", config.Retries, location)
	}
	if len(config.Interval) > 0 {
		if _, err := time.ParseDuration(config.Interval); err != nil {
			return fmt.Errorf("health check interval %s is invalid in %s, reason: %w", config.Interval, location, err)
		}
	}
	if len(config.Timeout) > 0 {
		if _, err := time.ParseDuration(config.Timeout); err != nil {
			return fmt.Errorf("health check timeout %s is invalid in %s, reason: %w", config.Timeout, location, err)
		}
	}
	if len(config.StartPeriod) > 0 {
		if _, err := time.ParseDuration(config.StartPeriod); err != nil {
			return fmt.Errorf("health check start period %s is invalid in %s, reason: %w", config.StartPeriod, location, err)
		}
	}
	if len(config.StartInterval) > 0 {
		if _, err := time.ParseDuration(config.StartInterval); err != nil {
			return fmt.Errorf("health check start interval %s is invalid in %s, reason: %w", config.StartInterval, location, err)
		}
	}
	return nil
}

func validateGlobalContainerConfig(config *GlobalContainerConfig, globalMountDefs []MountConfig) error {
	if config.StopTimeout < 0 {
		return fmt.Errorf("container stop timeout %d cannot be negative in global container config", config.StopTimeout)
	}
	if err := validateContainerRestartPolicy(&config.RestartPolicy, "global container config"); err != nil {
		return err
	}
	if err := validateContainerEnv(config.Env, "global container config"); err != nil {
		return err
	}
	if err := validateMountsConfig(config.Mounts, nil, globalMountDefs, "global container config mounts"); err != nil {
		return err
	}
	if err := validateLabelsConfig(config.Labels, "global container config"); err != nil {
		return err
	}
	return nil
}

func validateContainerRestartPolicy(config *ContainerRestartPolicyConfig, location string) error {
	if config.Mode != "on-failure" && config.MaxRetryCount != 0 {
		return fmt.Errorf("restart policy max retry count can be set only when the mode is on-failure in %s", location)
	}
	if len(config.Mode) == 0 {
		return nil
	}
	if _, err := restartPolicyModeFromString(config.Mode); err != nil {
		return fmt.Errorf("invalid restart policy mode %s in %s, valid values are %s", config.Mode, location, restartPolicyModeValidValues())
	}
	if config.MaxRetryCount < 0 {
		return fmt.Errorf("restart policy max retry count %d cannot be negative in %s", config.MaxRetryCount, location)
	}
	return nil
}

func validateIPAMConfig(ctx context.Context, config *IPAMConfig) (networkMap, map[ContainerReference]networkEndpointList, error) {
	networks := networkMap{}
	hostInterfaces := stringSet{}
	bridgeModeNetworks := config.Networks.BridgeModeNetworks
	prefixes := make(map[netip.Prefix]string)
	containerEndpoints := make(map[ContainerReference]networkEndpointList)
	allBridgeModeContainers := make(map[ContainerReference]struct{})

	for _, n := range bridgeModeNetworks {
		if len(n.Name) == 0 {
			return nil, nil, fmt.Errorf("network name cannot be empty")
		}
		if _, found := networks[n.Name]; found {
			return nil, nil, fmt.Errorf("network %s defined more than once in the IPAM config", n.Name)
		}

		if len(n.HostInterfaceName) == 0 {
			return nil, nil, fmt.Errorf("host interface name of network %s cannot be empty", n.Name)
		}
		if _, found := hostInterfaces[n.HostInterfaceName]; found {
			return nil, nil, fmt.Errorf("host interface name %s of network %s is already used by another network in the IPAM config", n.HostInterfaceName, n.Name)
		}
		if n.Priority <= 0 {
			return nil, nil, fmt.Errorf("network %s cannot have a non-positive priority %d", n.Name, n.Priority)
		}

		hostInterfaces[n.HostInterfaceName] = struct{}{}
		prefix, err := netip.ParsePrefix(n.CIDR)
		if err != nil {
			return nil, nil, fmt.Errorf("CIDR %s of network %s is invalid, reason: %w", n.CIDR, n.Name, err)
		}
		netAddr := prefix.Addr()
		if !netAddr.Is4() {
			return nil, nil, fmt.Errorf("CIDR %s of network %s is not an IPv4 subnet CIDR", n.CIDR, n.Name)
		}
		if masked := prefix.Masked(); masked.Addr() != netAddr {
			return nil, nil, fmt.Errorf("CIDR %s of network %s is not the same as the network address %s", n.CIDR, n.Name, masked)
		}
		if prefixLen := prefix.Bits(); prefixLen > 30 {
			return nil, nil, fmt.Errorf("CIDR %s of network %s (prefix length: %d) cannot have a prefix length more than 30 which makes the network unusable for container IP address allocations", n.CIDR, n.Name, prefixLen)
		}
		if !netAddr.IsPrivate() {
			return nil, nil, fmt.Errorf("CIDR %s of network %s is not within the RFC1918 private address space", n.CIDR, n.Name)
		}
		for pre, preNet := range prefixes {
			if prefix.Overlaps(pre) {
				return nil, nil, fmt.Errorf("CIDR %s of network %s overlaps with CIDR %s of network %s", n.CIDR, n.Name, pre, preNet)
			}
		}
		prefixes[prefix] = n.Name
		gatewayAddr := netAddr.Next()
		bmn := newBridgeModeNetwork(n.Name, n.Priority, &bridgeModeNetworkInfo{
			priority:          n.Priority,
			hostInterfaceName: n.HostInterfaceName,
			cidr:              prefix,
			gateway:           gatewayAddr,
		})
		networks[n.Name] = bmn

		containers := make(map[ContainerReference]struct{})
		containerIPs := make(map[netip.Addr]struct{})
		for _, cip := range n.Containers {
			ip := cip.IP
			ct := cip.Container
			if err := validateContainerReference(&ct); err != nil {
				return nil, nil, fmt.Errorf("container IP config within network %s has invalid container reference, reason: %w", n.Name, err)
			}

			caddr, err := netip.ParseAddr(ip)
			if err != nil {
				return nil, nil, fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s has invalid IP %s, reason: %w", ct.Group, ct.Container, n.Name, ip, err)
			}
			if !prefix.Contains(caddr) {
				return nil, nil, fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s cannot have an IP %s that does not belong to the network CIDR %s", ct.Group, ct.Container, n.Name, ip, prefix)
			}
			if caddr.Compare(netAddr) == 0 {
				return nil, nil, fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s cannot have an IP %s matching the network address %s", ct.Group, ct.Container, n.Name, ip, netAddr)
			}
			if caddr.Compare(gatewayAddr) == 0 {
				return nil, nil, fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s cannot have an IP %s matching the gateway address %s", ct.Group, ct.Container, n.Name, ip, gatewayAddr)
			}
			if _, found := containers[ct]; found {
				return nil, nil, fmt.Errorf("container {Group:%s Container:%s} cannot have multiple endpoints in network %s", ct.Group, ct.Container, n.Name)
			}
			if _, found := containerIPs[caddr]; found {
				return nil, nil, fmt.Errorf("IP %s of container {Group:%s Container:%s} is already in use by another container in network %s", ip, ct.Group, ct.Container, n.Name)
			}

			containers[ct] = struct{}{}
			allBridgeModeContainers[ct] = struct{}{}
			containerIPs[caddr] = struct{}{}
			containerEndpoints[ct] = append(containerEndpoints[ct], newBridgeModeEndpoint(bmn, ip))
		}
	}

	containerModeNetworks := config.Networks.ContainerModeNetworks
	allContainerModeContainers := make(map[ContainerReference]struct{})
	for _, n := range containerModeNetworks {
		if len(n.Name) == 0 {
			return nil, nil, fmt.Errorf("network name cannot be empty")
		}
		if _, found := networks[n.Name]; found {
			return nil, nil, fmt.Errorf("network %s defined more than once in the IPAM config", n.Name)
		}
		if err := validateContainerReference(&n.Container); err != nil {
			return nil, nil, fmt.Errorf("container reference of container mode network %s is invalid, reason: %w", n.Name, err)
		}
		cmn := newContainerModeNetwork(n.Name, &containerModeNetworkInfo{
			container: n.Container,
		})
		networks[n.Name] = cmn

		for _, ct := range n.AttachingContainers {
			if err := validateContainerReference(&ct); err != nil {
				return nil, nil, fmt.Errorf("container IP config within network %s has invalid container reference, reason: %w", n.Name, err)
			}
			if _, found := allContainerModeContainers[ct]; found {
				return nil, nil, fmt.Errorf("container {Group:%s Container:%s} is connected to multiple container mode network stacks", ct.Group, ct.Container)
			}
			if _, found := allBridgeModeContainers[ct]; found {
				return nil, nil, fmt.Errorf("container {Group:%s Container:%s} is connected to both bridge mode and container mode network stacks", ct.Group, ct.Container)
			}
			allContainerModeContainers[ct] = struct{}{}
			containerEndpoints[ct] = append(containerEndpoints[ct], newContainerModeEndpoint(cmn))
		}
	}

	for ct, endpoints := range containerEndpoints {
		if len(endpoints) <= 1 {
			continue
		}

		priorities := make(map[int]struct{})
		for _, e := range endpoints {
			p := e.network.bridgeModeInfo.priority
			if _, found := priorities[p]; found {
				return nil, nil, fmt.Errorf("container {Group:%s Container:%s} cannot have multiple bridge mode network endpoints whose networks have the same priority %d", ct.Group, ct.Container, p)
			}
			priorities[p] = struct{}{}
		}

		// Sort the networks for each container by priority (i.e. lowest
		// priority is the primary network interface for the container).
		sort.Slice(endpoints, func(i, j int) bool {
			// These networks are all guaranteed to be bridge mode networks
			// as we have already validated that a given container connects
			// to at most one container mode network and doesn't connect
			// to both bridge and container mode networks at the same time.
			n1 := endpoints[i].network
			n2 := endpoints[j].network

			return n1.bridgeModeInfo.priority < n2.bridgeModeInfo.priority
		})
	}

	return networks, containerEndpoints, nil
}

func validateHostsConfig(hosts []HostConfig, currentHost *hostInfo) (containerSet, error) {
	hostNames := stringSet{}
	allowedContainers := containerSet{}
	for _, h := range hosts {
		if len(h.Name) == 0 {
			return nil, fmt.Errorf("host name cannot be empty in the hosts config")
		}
		if _, found := hostNames[h.Name]; found {
			return nil, fmt.Errorf("host %s defined more than once in the hosts config", h.Name)
		}
		hostNames[h.Name] = struct{}{}

		containers := make(map[ContainerReference]bool)
		for _, ct := range h.AllowedContainers {
			err := validateContainerReference(&ct)
			if err != nil {
				return nil, fmt.Errorf("allowed container config within host %s has invalid container reference, reason: %w", h.Name, err)
			}
			if containers[ct] {
				return nil, fmt.Errorf("container {Group:%s Container:%s} defined more than once in the hosts config for host %s", ct.Group, ct.Container, h.Name)
			}
			containers[ct] = true
			if h.Name == currentHost.hostName {
				allowedContainers[ct] = true
			}
		}
	}
	return allowedContainers, nil
}

func validateGroupsConfig(groups []ContainerGroupConfig) (containerGroupMap, error) {
	containerGroups := containerGroupMap{}
	for _, g := range groups {
		if len(g.Name) == 0 {
			return nil, fmt.Errorf("group name cannot be empty in the groups config")
		}
		if _, found := containerGroups[g.Name]; found {
			return nil, fmt.Errorf("group %s defined more than once in the groups config", g.Name)
		}
		if g.Order < 1 {
			return nil, fmt.Errorf("group %s cannot have a non-positive order %d", g.Name, g.Order)
		}

		containerGroups[g.Name] = newContainerGroup(&g)
	}
	return containerGroups, nil
}

func validateContainersConfig(ctx context.Context, parentEnv *configEnv, containersConfig []ContainerConfig, groups containerGroupMap, globalConfig *GlobalConfig, containerEndpoints map[ContainerReference]networkEndpointList, allowedContainers containerSet) error {
	for _, ct := range containersConfig {
		g, found := groups[ct.Info.Group]
		if !found {
			return fmt.Errorf("group definition missing in groups config for the container {Group:%s Container:%s} in the containers config", ct.Info.Group, ct.Info.Container)
		}
		if _, found := g.containers[ct.Info]; found {
			return fmt.Errorf("container {Group:%s Container:%s} defined more than once in the containers config", ct.Info.Group, ct.Info.Container)
		}

		loc := fmt.Sprintf("container {Group: %s Container:%s} config", ct.Info.Group, ct.Info.Container)
		ctConfigEnv, err := validateConfigEnv(ctx, parentEnv, ct.Config.Env, loc)
		if err != nil {
			return err
		}
		ct.applyConfigEnv(ctConfigEnv)

		if len(ct.Image.Image) == 0 {
			return fmt.Errorf("image cannot be empty in %s", loc)
		}
		if ct.Image.SkipImagePull {
			if ct.Image.IgnoreImagePullFailures {
				return fmt.Errorf("ignoreImagePullFailures cannot be true when skipImagePull is true in %s", loc)
			}
			if ct.Image.PullImageBeforeStop {
				return fmt.Errorf("pullImageBeforeStop cannot be true when skipImagePull is true in %s", loc)
			}
		}

		if err := validateLabelsConfig(ct.Metadata.Labels, loc); err != nil {
			return err
		}

		if ct.Lifecycle.Order <= 0 {
			return fmt.Errorf("container order %d cannot be non-positive in %s", ct.Lifecycle.Order, loc)
		}
		if err := validateContainerRestartPolicy(&ct.Lifecycle.RestartPolicy, loc); err != nil {
			return err
		}
		if ct.Lifecycle.StopTimeout < 0 {
			return fmt.Errorf("container stop timeout %d cannot be negative in %s", ct.Lifecycle.StopTimeout, loc)
		}

		if len(ct.User.PrimaryGroup) > 0 && len(ct.User.User) == 0 {
			return fmt.Errorf("container user primary group cannot be set without setting the user in %s", loc)
		}

		if err := validateDevicesConfig(ct.Filesystem.Devices, loc); err != nil {
			return err
		}

		if err := validateMountsConfig(ct.Filesystem.Mounts, globalConfig.Container.Mounts, globalConfig.MountDefs, fmt.Sprintf("%s mounts", loc)); err != nil {
			return err
		}

		if err := validatePublishedPortsConfig(ct.Network.PublishedPorts, loc); err != nil {
			return err
		}

		if err := validateSysctlsConfig(ct.Security.Sysctls, loc); err != nil {
			return err
		}

		if err := validateHealthConfig(&ct.Health, loc); err != nil {
			return err
		}

		if len(ct.Runtime.ShmSize) > 0 {
			if _, err := units.RAMInBytes(ct.Runtime.ShmSize); err != nil {
				return fmt.Errorf("invalid shmSize %s in %s, reason: %w", ct.Runtime.ShmSize, loc, err)
			}
		}
		if err := validateContainerEnv(ct.Runtime.Env, loc); err != nil {
			return err
		}

		g.addContainer(&ct, globalConfig, containerEndpoints[ct.Info], allowedContainers[ct.Info])
	}

	return nil
}

func validateContainerReference(ref *ContainerReference) error {
	if len(ref.Group) == 0 {
		return fmt.Errorf("container reference cannot have an empty group name")
	}
	if len(ref.Container) == 0 {
		return fmt.Errorf("container reference cannot have an empty container name")
	}
	return nil
}

func newBridgeModeEndpoint(network *network, ip string) *containerNetworkEndpoint {
	return &containerNetworkEndpoint{network: network, ip: ip}
}

func newContainerModeEndpoint(network *network) *containerNetworkEndpoint {
	return &containerNetworkEndpoint{network: network}
}
