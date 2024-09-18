package main

import (
	"fmt"
	"net/netip"
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
				return fmt.Errorf("mount specified by just the name %s not found in defs", m.Name)
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
			return fmt.Errorf("mount name %s specifies options in %s, that are not supported when mount type is bind", m.Name, location)
		}
	}
	return nil
}

func validateGlobalContainerConfig(config *GlobalContainerConfig, globalMountDefs []MountConfig) error {
	if config.StopTimeout < 0 {
		return fmt.Errorf("container stop timeout cannot be negative (%d) in global container config", config.StopTimeout)
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
		return fmt.Errorf("restart policy max retry count can be set in %s only when the mode is on-failure", location)
	}
	if len(config.Mode) == 0 {
		return nil
	}
	if _, err := restartPolicyModeFromString(config.Mode); err != nil {
		return fmt.Errorf("invalid restart policy mode %s in %s, valid values are %s", config.Mode, location, restartPolicyModeValidValues())
	}
	if config.MaxRetryCount < 0 {
		return fmt.Errorf("restart policy max retry count (%d) in %s cannot be negative", config.MaxRetryCount, location)
	}
	return nil
}

func validateIPAMConfig(config *IPAMConfig) error {
	networks := make(map[string]bool)
	hostInterfaces := make(map[string]bool)
	bridgeModeNetworks := config.Networks.BridgeModeNetworks
	prefixes := make(map[netip.Prefix]string)
	for _, n := range bridgeModeNetworks {
		if len(n.Name) == 0 {
			return fmt.Errorf("network name cannot be empty")
		}
		if networks[n.Name] {
			return fmt.Errorf("network %s defined more than once in the IPAM config", n.Name)
		}

		if len(n.HostInterfaceName) == 0 {
			return fmt.Errorf("host interface name of network %s cannot be empty", n.Name)
		}
		if hostInterfaces[n.HostInterfaceName] {
			return fmt.Errorf("host interface name %s of network %s is already used by another network in the IPAM config", n.HostInterfaceName, n.Name)
		}
		if n.Priority <= 0 {
			return fmt.Errorf("network %s cannot have a non-positive priority %d", n.Name, n.Priority)
		}

		networks[n.Name] = true
		hostInterfaces[n.HostInterfaceName] = true
		prefix, err := netip.ParsePrefix(n.CIDR)
		if err != nil {
			return fmt.Errorf("CIDR %s of network %s is invalid, reason: %w", n.CIDR, n.Name, err)
		}
		netAddr := prefix.Addr()
		if !netAddr.Is4() {
			return fmt.Errorf("CIDR %s of network %s is not an IPv4 subnet CIDR", n.CIDR, n.Name)
		}
		if masked := prefix.Masked(); masked.Addr() != netAddr {
			return fmt.Errorf("CIDR %s of network %s is not the same as the network address %s", n.CIDR, n.Name, masked)
		}
		if prefixLen := prefix.Bits(); prefixLen > 30 {
			return fmt.Errorf("CIDR %s of network %s (prefix length: %d) cannot have a prefix length more than 30 which makes the network unusable for container IP address allocations", n.CIDR, n.Name, prefixLen)
		}
		if !netAddr.IsPrivate() {
			return fmt.Errorf("CIDR %s of network %s is not within the RFC1918 private address space", n.CIDR, n.Name)
		}
		for pre, preNet := range prefixes {
			if prefix.Overlaps(pre) {
				return fmt.Errorf("CIDR %s of network %s overlaps with CIDR %s of network %s", n.CIDR, n.Name, pre, preNet)
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
				return fmt.Errorf("container IP config within network %s has invalid container reference, reason: %w", n.Name, err)
			}

			caddr, err := netip.ParseAddr(ip)
			if err != nil {
				return fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s has invalid IP %s, reason: %w", ct.Group, ct.Container, n.Name, ip, err)
			}
			if !prefix.Contains(caddr) {
				return fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s cannot have an IP %s that does not belong to the network CIDR %s", ct.Group, ct.Container, n.Name, ip, prefix)
			}
			if caddr.Compare(netAddr) == 0 {
				return fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s cannot have an IP %s matching the network address %s", ct.Group, ct.Container, n.Name, ip, netAddr)
			}
			if caddr.Compare(gatewayAddr) == 0 {
				return fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s cannot have an IP %s matching the gateway address %s", ct.Group, ct.Container, n.Name, ip, gatewayAddr)
			}
			if containers[ct] {
				return fmt.Errorf("container {Group:%s Container:%s} cannot have multiple endpoints in network %s", ct.Group, ct.Container, n.Name)
			}
			if containerIPs[caddr] {
				return fmt.Errorf("IP %s of container {Group:%s Container:%s} is already in use by another container in network %s", ip, ct.Group, ct.Container, n.Name)
			}

			containers[ct] = true
			containerIPs[caddr] = true
		}
	}
	containerModeNetworks := config.Networks.ContainerModeNetworks
	for _, n := range containerModeNetworks {
		if len(n.Name) == 0 {
			return fmt.Errorf("network name cannot be empty")
		}
		if networks[n.Name] {
			return fmt.Errorf("network %s defined more than once in the IPAM config", n.Name)
		}
		if n.Priority <= 0 {
			return fmt.Errorf("network %s cannot have a non-positive priority %d", n.Name, n.Priority)
		}
		networks[n.Name] = true

		containers := make(map[ContainerReference]bool)
		for _, ct := range n.Containers {
			if err := validateContainerReference(&ct); err != nil {
				return fmt.Errorf("container IP config within network %s has invalid container reference, reason: %w", n.Name, err)
			}
			if containers[ct] {
				return fmt.Errorf("container {Group:%s Container:%s} is connected to multiple container mode network stacks", ct.Group, ct.Container)
			}
			containers[ct] = true
		}
	}
	return nil
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

func validateGroupsConfig(groups []ContainerGroupConfig) error {
	groupNames := make(map[string]bool)
	for _, g := range groups {
		if len(g.Name) == 0 {
			return fmt.Errorf("group name cannot be empty in the groups config")
		}
		if groupNames[g.Name] {
			return fmt.Errorf("group %s defined more than once in the groups config", g.Name)
		}
		if g.Order < 1 {
			return fmt.Errorf("group %s cannot have a non-positive order %d", g.Name, g.Order)
		}

		groupNames[g.Name] = true
	}
	return nil
}

func validateContainersConfig(containers []ContainerConfig) error {
	// TODO: Perform the following (and more) validations:
	// Container configs:
	//     a. Parent group name is a valid group defined under group config.
	//     b. No duplicate container names within the same group.
	//     c. Order defined for all the containers.
	//     d. Image defined for all the containers.
	//     e. Validate mandatory properties of every device config.
	//     f. Validate manadatory properties of every container config mount.
	//     g. Mount pure name references are valid global config mount references.
	//     h. Validate manadatory properties of every container config env.
	//     i. Every container config env specifies exactly one of value or
	//        valueCommand, but not both.
	//     j. Validate mandatory properties of every published port config.
	//     k. Validate mandatory properties of every label config.
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
