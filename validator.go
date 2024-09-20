package main

import (
	"fmt"
	"net/netip"
	"time"

	"github.com/docker/go-units"
)

func validateGlobalConfig(config *GlobalConfig) error {
	if err := validateConfigEnv(config.Env, "global config"); err != nil {
		return err
	}

	if err := validateMountsConfig(config.MountDefs, nil, nil, "global config mount defs"); err != nil {
		return err
	}

	if err := validateGlobalContainerConfig(&config.Container, config.MountDefs); err != nil {
		return err
	}

	return nil
}

func validateConfigEnv(config []ConfigEnv, location string) error {
	envs := make(map[string]bool)
	for _, e := range config {
		if len(e.Var) == 0 {
			return fmt.Errorf("empty env var in %s", location)
		}
		if envs[e.Var] {
			return fmt.Errorf("env var %s specified more than once in %s", e.Var, location)
		}
		envs[e.Var] = true

		if len(e.Value) == 0 && len(e.ValueCommand) == 0 {
			return fmt.Errorf("neither value nor valueCommand specified for env var %s in %s", e.Var, location)
		}
		if len(e.Value) > 0 && len(e.ValueCommand) > 0 {
			return fmt.Errorf("exactly one of value or valueCommand must be specified for env var %s in %s", e.Var, location)
		}
	}
	return nil
}

func validateContainerEnv(config []ContainerEnv, location string) error {
	envs := make(map[string]bool)
	for _, e := range config {
		if len(e.Var) == 0 {
			return fmt.Errorf("empty env var in %s", location)
		}
		if envs[e.Var] {
			return fmt.Errorf("env var %s specified more than once in %s", e.Var, location)
		}
		envs[e.Var] = true

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
	labels := make(map[string]bool)
	for _, l := range config {
		if len(l.Name) == 0 {
			return fmt.Errorf("empty label name in %s", location)
		}
		if labels[l.Name] {
			return fmt.Errorf("label name %s specified more than once in %s", l.Name, location)
		}
		labels[l.Name] = true

		if len(l.Value) == 0 {
			return fmt.Errorf("empty label value for label %s in %s", l.Name, location)
		}
	}
	return nil
}

func validateMountsConfig(config, commonConfig, globalDefs []MountConfig, location string) error {
	// First build a map of the mounts from the globalDefs (which should
	// already have been validated).
	globalMountDefs := make(map[string]bool)
	for _, m := range globalDefs {
		globalMountDefs[m.Name] = true
	}

	// Build a map of the mounts from the commonConfig next which acts
	// as the first set of mounts to apply. These should also have been
	// validated prior and hence we don't validate them here again.
	mounts := make(map[string]bool)
	for _, m := range commonConfig {
		mounts[m.Name] = true
	}

	// Finally iterate and validate the mounts in the current level config.
	for _, m := range config {
		if len(m.Name) == 0 {
			return fmt.Errorf("mount name cannot be empty in %s", location)
		}
		if mounts[m.Name] {
			return fmt.Errorf("mount name %s defined more than once in %s", m.Name, location)
		}
		mounts[m.Name] = true

		if len(m.Type) == 0 && len(m.Src) == 0 && len(m.Dst) == 0 && !m.ReadOnly {
			// This is a mount with just the name. Match this against the
			// global mount defs.
			if !globalMountDefs[m.Name] {
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
	keys := make(map[string]bool)
	for _, s := range sysctls {
		if len(s.Key) == 0 {
			return fmt.Errorf("empty sysctl key in %s", location)
		}
		if keys[s.Key] {
			return fmt.Errorf("sysctl key %s specified more than once in %s", s.Key, location)
		}
		keys[s.Key] = true

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

func validateIPAMConfig(config *IPAMConfig) (networkMap, error) {
	networks := networkMap{}
	hostInterfaces := make(map[string]bool)
	bridgeModeNetworks := config.Networks.BridgeModeNetworks
	prefixes := make(map[netip.Prefix]string)
	for _, n := range bridgeModeNetworks {
		if len(n.Name) == 0 {
			return nil, fmt.Errorf("network name cannot be empty")
		}
		if _, ok := networks[n.Name]; ok {
			return nil, fmt.Errorf("network %s defined more than once in the IPAM config", n.Name)
		}

		if len(n.HostInterfaceName) == 0 {
			return nil, fmt.Errorf("host interface name of network %s cannot be empty", n.Name)
		}
		if hostInterfaces[n.HostInterfaceName] {
			return nil, fmt.Errorf("host interface name %s of network %s is already used by another network in the IPAM config", n.HostInterfaceName, n.Name)
		}
		if n.Priority <= 0 {
			return nil, fmt.Errorf("network %s cannot have a non-positive priority %d", n.Name, n.Priority)
		}

		networks[n.Name] = newBridgeModeNetwork(&n)
		hostInterfaces[n.HostInterfaceName] = true
		prefix, err := netip.ParsePrefix(n.CIDR)
		if err != nil {
			return nil, fmt.Errorf("CIDR %s of network %s is invalid, reason: %w", n.CIDR, n.Name, err)
		}
		netAddr := prefix.Addr()
		if !netAddr.Is4() {
			return nil, fmt.Errorf("CIDR %s of network %s is not an IPv4 subnet CIDR", n.CIDR, n.Name)
		}
		if masked := prefix.Masked(); masked.Addr() != netAddr {
			return nil, fmt.Errorf("CIDR %s of network %s is not the same as the network address %s", n.CIDR, n.Name, masked)
		}
		if prefixLen := prefix.Bits(); prefixLen > 30 {
			return nil, fmt.Errorf("CIDR %s of network %s (prefix length: %d) cannot have a prefix length more than 30 which makes the network unusable for container IP address allocations", n.CIDR, n.Name, prefixLen)
		}
		if !netAddr.IsPrivate() {
			return nil, fmt.Errorf("CIDR %s of network %s is not within the RFC1918 private address space", n.CIDR, n.Name)
		}
		for pre, preNet := range prefixes {
			if prefix.Overlaps(pre) {
				return nil, fmt.Errorf("CIDR %s of network %s overlaps with CIDR %s of network %s", n.CIDR, n.Name, pre, preNet)
			}
		}
		prefixes[prefix] = n.Name

		gatewayAddr := netAddr.Next()
		containers := make(map[ContainerReference]bool)
		containerIPs := make(map[netip.Addr]bool)
		for _, cip := range n.Containers {
			ip := cip.IP
			ct := cip.Container
			if err := validateContainerReference(&ct); err != nil {
				return nil, fmt.Errorf("container IP config within network %s has invalid container reference, reason: %w", n.Name, err)
			}

			caddr, err := netip.ParseAddr(ip)
			if err != nil {
				return nil, fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s has invalid IP %s, reason: %w", ct.Group, ct.Container, n.Name, ip, err)
			}
			if !prefix.Contains(caddr) {
				return nil, fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s cannot have an IP %s that does not belong to the network CIDR %s", ct.Group, ct.Container, n.Name, ip, prefix)
			}
			if caddr.Compare(netAddr) == 0 {
				return nil, fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s cannot have an IP %s matching the network address %s", ct.Group, ct.Container, n.Name, ip, netAddr)
			}
			if caddr.Compare(gatewayAddr) == 0 {
				return nil, fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s cannot have an IP %s matching the gateway address %s", ct.Group, ct.Container, n.Name, ip, gatewayAddr)
			}
			if containers[ct] {
				return nil, fmt.Errorf("container {Group:%s Container:%s} cannot have multiple endpoints in network %s", ct.Group, ct.Container, n.Name)
			}
			if containerIPs[caddr] {
				return nil, fmt.Errorf("IP %s of container {Group:%s Container:%s} is already in use by another container in network %s", ip, ct.Group, ct.Container, n.Name)
			}

			containers[ct] = true
			containerIPs[caddr] = true
		}
	}
	containerModeNetworks := config.Networks.ContainerModeNetworks
	for _, n := range containerModeNetworks {
		if len(n.Name) == 0 {
			return nil, fmt.Errorf("network name cannot be empty")
		}
		if _, ok := networks[n.Name]; ok {
			return nil, fmt.Errorf("network %s defined more than once in the IPAM config", n.Name)
		}
		if n.Priority <= 0 {
			return nil, fmt.Errorf("network %s cannot have a non-positive priority %d", n.Name, n.Priority)
		}
		networks[n.Name] = newContainerModeNetwork(&n)

		containers := make(map[ContainerReference]bool)
		for _, ct := range n.Containers {
			if err := validateContainerReference(&ct); err != nil {
				return nil, fmt.Errorf("container IP config within network %s has invalid container reference, reason: %w", n.Name, err)
			}
			if containers[ct] {
				return nil, fmt.Errorf("container {Group:%s Container:%s} is connected to multiple container mode network stacks", ct.Group, ct.Container)
			}
			containers[ct] = true
		}
	}
	return networks, nil
}

func validateHostsConfig(hosts []HostConfig) error {
	hostNames := make(map[string]bool)
	for _, h := range hosts {
		if len(h.Name) == 0 {
			return fmt.Errorf("host name cannot be empty in the hosts config")
		}
		if hostNames[h.Name] {
			return fmt.Errorf("host %s defined more than once in the hosts config", h.Name)
		}
		hostNames[h.Name] = true

		containers := make(map[ContainerReference]bool)
		for _, ct := range h.AllowedContainers {
			err := validateContainerReference(&ct)
			if err != nil {
				return fmt.Errorf("allowed container config within host %s has invalid container reference, reason: %w", h.Name, err)
			}
			if containers[ct] {
				return fmt.Errorf("container {Group:%s Container:%s} defined more than once in the hosts config for host %s", ct.Group, ct.Container, h.Name)
			}
			containers[ct] = true
		}
	}
	return nil
}

func validateGroupsConfig(groups []ContainerGroupConfig) (containerGroupMap, error) {
	containerGroups := containerGroupMap{}
	for _, g := range groups {
		if len(g.Name) == 0 {
			return nil, fmt.Errorf("group name cannot be empty in the groups config")
		}
		if _, ok := containerGroups[g.Name]; ok {
			return nil, fmt.Errorf("group %s defined more than once in the groups config", g.Name)
		}
		if g.Order < 1 {
			return nil, fmt.Errorf("group %s cannot have a non-positive order %d", g.Name, g.Order)
		}

		containerGroups[g.Name] = &containerGroup{
			config: &g,
		}
	}
	return containerGroups, nil
}

func validateContainersConfig(containersConfig []ContainerConfig, groupsConfig []ContainerGroupConfig, globalConfig *GlobalConfig) error {
	groups := make(map[string]bool)
	for _, g := range groupsConfig {
		groups[g.Name] = true
	}

	containers := make(map[ContainerReference]bool)
	for _, ct := range containersConfig {
		if !groups[ct.Info.Group] {
			return fmt.Errorf("group definition missing in groups config for the container {Group:%s Container:%s} in the containers config", ct.Info.Group, ct.Info.Container)
		}
		if containers[ct.Info] {
			return fmt.Errorf("container {Group:%s Container:%s} defined more than once in the containers config", ct.Info.Group, ct.Info.Container)
		}
		containers[ct.Info] = true

		loc := fmt.Sprintf("container {Group: %s Container:%s} config", ct.Info.Group, ct.Info.Container)
		if err := validateConfigEnv(ct.Config.Env, loc); err != nil {
			return err
		}

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
