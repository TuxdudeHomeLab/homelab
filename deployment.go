package main

import (
	"fmt"
)

type deployment struct {
	config   *HomelabConfig
	groups   containerGroupMap
	networks networkMap
	host     *hostInfo
}

func buildDeployment() (*deployment, error) {
	c := HomelabConfig{}
	err := c.parse()
	if err != nil {
		return nil, err
	}

	err = c.validate()
	if err != nil {
		return nil, err
	}

	return newDeployment(&c), nil
}

func newDeployment(config *HomelabConfig) *deployment {
	d := deployment{config: config}
	// First build the networks as it will be looked up while building
	// the container groups and containers within.
	d.populateNetworks()
	d.populateGroups()
	d.populateHostInfo()
	return &d
}

func (d *deployment) populateNetworks() {
	networks := make(networkMap)
	for _, n := range d.config.IPAM.Networks.BridgeModeNetworks {
		nt := newBridgeModeNetwork(d, &n)
		networks[nt.name()] = nt
		// TODO: Remove these once these functions get used.
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
		cg := newContainerGroup(d, &g, &d.config.Containers)
		groups[cg.name()] = cg
	}
	d.groups = groups
}

func (d *deployment) populateHostInfo() {
	d.host = newHostInfo(d.config)
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
