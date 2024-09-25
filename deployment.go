package main

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
)

type deployment struct {
	config                 *HomelabConfig
	groups                 containerGroupMap
	groupsOrder            []string
	networks               networkMap
	networksOrder          []string
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
	var containerEndpoints map[ContainerReference]networkEndpointList
	d.networks, containerEndpoints, err = validateIPAMConfig(ctx, &config.IPAM)
	if err != nil {
		return nil, err
	}
	d.updateNetworksOrder()

	d.groups, err = validateGroupsConfig(config.Groups)
	if err != nil {
		return nil, err
	}
	d.updateGroupsOrder()

	err = validateContainersConfig(config.Containers, d.groups, &config.Global, containerEndpoints, d.allowedContainers)
	if err != nil {
		return nil, err
	}

	for _, g := range d.groups {
		g.updateContainersOrder()
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

func (d *deployment) updateGroupsOrder() {
	d.groupsOrder = make([]string, 0)
	for g := range d.groups {
		d.groupsOrder = append(d.groupsOrder, g)
	}
	sort.Slice(d.groupsOrder, func(i, j int) bool {
		g1 := d.groupsOrder[i]
		g2 := d.groupsOrder[j]
		return g1 < g2
	})
}

func (d *deployment) updateNetworksOrder() {
	d.networksOrder = make([]string, 0)
	for n := range d.networks {
		d.networksOrder = append(d.networksOrder, n)
	}
	sort.Slice(d.networksOrder, func(i, j int) bool {
		n1 := d.networksOrder[i]
		n2 := d.networksOrder[j]
		return n1 < n2
	})
}

func (d *deployment) String() string {
	var sb strings.Builder

	sb.WriteString("Deployment{Groups:[")
	if len(d.groupsOrder) == 0 {
		sb.WriteString("empty")
	} else {
		sb.WriteString(d.groups[d.groupsOrder[0]].String())
		for i := 1; i < len(d.groupsOrder); i++ {
			sb.WriteString(fmt.Sprintf(", %s", d.groups[d.groupsOrder[i]]))
		}
	}

	sb.WriteString("], Networks:[")
	if len(d.networksOrder) == 0 {
		sb.WriteString("empty]}")
	} else {
		sb.WriteString(d.networks[d.networksOrder[0]].String())
		for i := 1; i < len(d.networksOrder); i++ {
			sb.WriteString(fmt.Sprintf(", %s", d.networks[d.networksOrder[i]]))
		}
		sb.WriteString("]}")
	}
	return sb.String()

}
