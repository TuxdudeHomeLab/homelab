package main

import (
	"fmt"
	"io"
)

type deployment struct {
	config            *HomelabConfig
	groups            containerGroupMap
	networks          networkMap
	host              *hostInfo
	allowedContainers stringSet
}

func buildDeployment(configsPath string) (*deployment, error) {
	return buildDeploymentFromConfigsPath(configsPath, newHostInfo())
}

func buildDeploymentFromConfigsPath(configsPath string, host *hostInfo) (*deployment, error) {
	config := HomelabConfig{}
	err := config.parseConfigs(configsPath)
	if err != nil {
		return nil, err
	}
	return buildDeploymentFromConfig(&config, host)
}

func buildDeploymentFromReader(reader io.Reader, host *hostInfo) (*deployment, error) {
	config := HomelabConfig{}
	err := config.parse(reader)
	if err != nil {
		return nil, err
	}
	return buildDeploymentFromConfig(&config, host)
}

func buildDeploymentFromConfig(config *HomelabConfig, host *hostInfo) (*deployment, error) {
	d := deployment{
		config: config,
		host:   host,
	}
	// env := newConfigEnv(host)

	err := validateGlobalConfig(&config.Global)
	if err != nil {
		return nil, err
	}

	err = validateHostsConfig(config.Hosts)
	if err != nil {
		return nil, err
	}

	err = validateIPAMConfig(&config.IPAM)
	if err != nil {
		return nil, err
	}

	_, err = validateGroupsConfig(config.Groups)
	if err != nil {
		return nil, err
	}

	err = validateContainersConfig(config.Containers, config.Groups, &config.Global)
	if err != nil {
		return nil, err
	}

	// First build the networks as it will be looked up while building
	// the container groups and containers within.
	d.populateNetworks()
	d.populateGroups()
	d.populateAllowedContainers()

	return &d, nil
}

func (d *deployment) populateNetworks() {
	networks := make(networkMap)
	for _, n := range d.config.IPAM.Networks.BridgeModeNetworks {
		nt := newBridgeModeNetwork(d, &n)
		networks[nt.name()] = nt
	}
	for _, n := range d.config.IPAM.Networks.ContainerModeNetworks {
		nt := newContainerModeNetwork(d, &n)
		networks[nt.name()] = nt
	}
	d.networks = networks
}

func (d *deployment) populateGroups() {
	groups := make(containerGroupMap)
	for _, g := range d.config.Groups {
		cg := newContainerGroupDeprecated(d, &g, &d.config.Containers)
		groups[cg.name()] = cg
	}
	d.groups = groups
}

func (d *deployment) populateAllowedContainers() {
	d.allowedContainers = make(stringSet)
	for _, h := range d.config.Hosts {
		if h.Name == d.host.hostName {
			for _, c := range h.AllowedContainers {
				d.allowedContainers[containerName(c.Group, c.Container)] = true
			}
			break
		}
	}
}

func (d *deployment) queryAllContainers() containerMap {
	result := make(containerMap)
	for _, g := range d.groups {
		for _, c := range g.containers {
			result[c.name()] = c
		}
	}
	return result
}

func (d *deployment) queryAllContainersInGroup(group string) containerMap {
	result := make(containerMap)
	for _, g := range d.groups {
		if g.name() == group {
			for _, c := range g.containers {
				result[c.name()] = c
			}
			break
		}
	}
	return result
}

func (d *deployment) queryContainer(group string, container string) *container {
	cn := containerName(group, container)
	for _, g := range d.groups {
		if g.name() == group {
			for _, c := range g.containers {
				if c.name() == cn {
					return c
				}
			}
		}
	}
	return nil
}

func (d *deployment) String() string {
	return fmt.Sprintf("Deployment{Groups:%s, Networks:%s}", d.groups, d.networks)
}
