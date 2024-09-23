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

func buildDeploymentFromConfigsPath(ctx context.Context, configsPath string) (*deployment, error) {
	r, err := mergedConfigsReader(ctx, configsPath)
	if err != nil {
		return nil, err
	}

	return buildDeployment(ctx, r)
}

func buildDeployment(ctx context.Context, reader io.Reader) (*deployment, error) {
	config := HomelabConfig{}
	err := config.parse(ctx, reader)
	if err != nil {
		return nil, err
	}
	return buildDeploymentFromConfig(ctx, &config)
}

func buildDeploymentFromConfig(ctx context.Context, config *HomelabConfig) (*deployment, error) {
	host, found := hostInfoFromContext(ctx)
	if !found {
		host = newHostInfo(ctx)
	}
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

func (d *deployment) queryAllContainersInGroup(group string) (containerMap, error) {
	result := make(containerMap)
	g, found := d.groups[group]
	if !found {
		return nil, fmt.Errorf("group %s not found", group)
	}
	for cref, c := range g.containers {
		result[cref] = c
	}
	return result, nil
}

func (d *deployment) queryContainer(cRef ContainerReference) (*container, error) {
	g, found := d.groups[cRef.Group]
	if !found {
		return nil, fmt.Errorf("group %s not found", cRef.Group)
	}
	ct, found := g.containers[cRef]
	if !found {
		return nil, fmt.Errorf("container %s not found", cRef)
	}
	return ct, nil
}

func (d *deployment) queryContainers(ctx context.Context, allGroups bool, group, container string) (containerList, error) {
	if allGroups {
		return containerMapToList(d.queryAllContainers()), nil
	}
	if group != "" && container == "" {
		ctMap, err := d.queryAllContainersInGroup(group)
		if err != nil {
			return nil, err
		}
		return containerMapToList(ctMap), nil
	}
	if group != "" {
		ct, err := d.queryContainer(ContainerReference{Group: group, Container: container})
		if err != nil {
			return nil, err
		}
		return containerList{ct}, nil
	}
	log(ctx).Fatalf("Invalid scenario, possibly indicating a bug in the code")
	return nil, nil
}

func (d *deployment) String() string {
	return fmt.Sprintf("Deployment{Groups:%s, Networks:%s}", d.groups, d.networks)
}
