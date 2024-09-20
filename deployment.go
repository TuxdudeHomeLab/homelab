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
	allowedContainers containerReferenceSet
}

type containerReferenceSet map[ContainerReference]bool

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

	d.allowedContainers, err = validateHostsConfig(config.Hosts, host)
	if err != nil {
		return nil, err
	}

	// First build the networks as it will be looked up while building
	// the container groups and containers within.
	d.networks, err = validateIPAMConfig(&config.IPAM)
	if err != nil {
		return nil, err
	}

	d.groups, err = validateGroupsConfig(config.Groups)
	if err != nil {
		return nil, err
	}

	err = validateContainersConfig(config.Containers, d.groups, &config.Global, d.networks, d.allowedContainers)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (d *deployment) queryAllContainers() containerMap {
	result := make(containerMap)
	for _, g := range d.groups {
		for cref, c := range g.containers {
			result[cref] = c
		}
	}
	return result
}

func (d *deployment) queryAllContainersInGroup(group string) containerMap {
	result := make(containerMap)
	for _, g := range d.groups {
		if g.name() == group {
			for cref, c := range g.containers {
				result[cref] = c
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
