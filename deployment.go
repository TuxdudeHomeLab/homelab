package main

import (
	"context"
	"fmt"
	"io"
)

type deployment struct {
	config                 *HomelabConfig
	groups                 containerGroupMap
	networks               networkMap
	host                   *hostInfo
	allowedContainers      containerSet
	containerDockerConfigs containerDockerConfigMap
}

func buildDeployment(ctx context.Context, configsPath string) (*deployment, error) {
	return buildDeploymentFromConfigsPath(ctx, configsPath, newHostInfo(ctx))
}

func buildDeploymentFromConfigsPath(ctx context.Context, configsPath string, host *hostInfo) (*deployment, error) {
	config := HomelabConfig{}
	err := config.parseConfigs(ctx, configsPath)
	if err != nil {
		return nil, err
	}
	return buildDeploymentFromConfig(ctx, &config, host)
}

func buildDeploymentFromReader(ctx context.Context, reader io.Reader, host *hostInfo) (*deployment, error) {
	config := HomelabConfig{}
	err := config.parse(ctx, reader)
	if err != nil {
		return nil, err
	}
	return buildDeploymentFromConfig(ctx, &config, host)
}

func buildDeploymentFromConfig(ctx context.Context, config *HomelabConfig, host *hostInfo) (*deployment, error) {
	d := deployment{
		config:                 config,
		host:                   host,
		containerDockerConfigs: containerDockerConfigMap{},
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
	var containerRefIPs map[ContainerReference]networkContainerIPList
	d.networks, containerRefIPs, err = validateIPAMConfig(ctx, &config.IPAM)
	if err != nil {
		return nil, err
	}

	d.groups, err = validateGroupsConfig(config.Groups)
	if err != nil {
		return nil, err
	}

	err = validateContainersConfig(config.Containers, d.groups, &config.Global, containerRefIPs, d.allowedContainers)
	if err != nil {
		return nil, err
	}

	for _, g := range d.groups {
		for _, ct := range g.containers {
			cConfig, hConfig, nConfig, err := ct.generateDockerConfigs()
			if err != nil {
				log(ctx).Fatalf("Error generating docker configs for container %s, reason: %v", ct, err)
			}
			cdc := containerDockerConfigs{
				ContainerConfig: cConfig,
				HostConfig:      hConfig,
				NetworkConfig:   nConfig,
			}
			d.containerDockerConfigs[ct.config.Info] = &cdc
		}
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

func (d *deployment) queryContainer(ct *ContainerReference) *container {
	cn := containerName(ct)
	for _, g := range d.groups {
		if g.name() == ct.Group {
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
