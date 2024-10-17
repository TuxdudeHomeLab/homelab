package deployment

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/tuxdudehomelab/homelab/internal/config"
	"github.com/tuxdudehomelab/homelab/internal/config/env"
)

type Deployment struct {
	Config            *config.Homelab
	Groups            ContainerGroupMap
	GroupsOrder       []string
	Networks          NetworkMap
	NetworksOrder     []string
	allowedContainers containerSet
	dockerConfigs     containerDockerConfigMap
}

func FromConfigsPath(ctx context.Context, configsPath string) (*Deployment, error) {
	r, err := config.MergedConfigsReader(ctx, configsPath)
	if err != nil {
		return nil, err
	}

	return FromReader(ctx, r)
}

func FromReader(ctx context.Context, reader io.Reader) (*Deployment, error) {
	conf := config.Homelab{}
	err := conf.Parse(ctx, reader)
	if err != nil {
		return nil, err
	}
	return FromConfig(ctx, &conf)
}

func FromConfig(ctx context.Context, conf *config.Homelab) (*Deployment, error) {
	d := Deployment{
		Config:        conf,
		dockerConfigs: containerDockerConfigMap{},
	}

	systemEnv := env.NewSystemConfigEnvManager(ctx)
	envWithGlobal, err := validateGlobalConfig(ctx, systemEnv, &conf.Global)
	if err != nil {
		return nil, err
	}

	d.allowedContainers, err = validateHostsConfig(ctx, conf.Hosts)
	if err != nil {
		return nil, err
	}

	// First build the networks as they will be looked up while building
	// the container groups and containers within.
	var containerEndpoints map[config.ContainerReference]networkEndpointList
	d.Networks, containerEndpoints, err = validateIPAMConfig(ctx, &conf.IPAM)
	if err != nil {
		return nil, err
	}
	d.updateNetworksOrder()

	d.Groups, err = validateGroupsConfig(conf.Groups)
	if err != nil {
		return nil, err
	}
	d.updateGroupsOrder()

	err = validateContainersConfig(ctx, envWithGlobal, conf.Containers, d.Groups, &conf.Global, containerEndpoints, d.allowedContainers)
	if err != nil {
		return nil, err
	}

	for _, g := range d.Groups {
		g.updateContainersOrder()
		for _, ct := range g.containers {
			d.dockerConfigs[ct.config.Info] = ct.generateDockerConfigs()
		}
	}

	return &d, nil
}

func (d *Deployment) queryAllContainers() containerMap {
	result := make(containerMap)
	for _, g := range d.Groups {
		for cref, c := range g.containers {
			result[cref] = c
		}
	}
	return result
}

func (d *Deployment) queryAllContainersInGroup(group string) (containerMap, error) {
	result := make(containerMap)
	g, found := d.Groups[group]
	if !found {
		return nil, fmt.Errorf("group %s not found", group)
	}
	for cref, c := range g.containers {
		result[cref] = c
	}
	return result, nil
}

func (d *Deployment) queryContainer(cRef config.ContainerReference) (*Container, error) {
	g, found := d.Groups[cRef.Group]
	if !found {
		return nil, fmt.Errorf("group %s not found", cRef.Group)
	}
	ct, found := g.containers[cRef]
	if !found {
		return nil, fmt.Errorf("container %s not found", cRef)
	}
	return ct, nil
}

func (d *Deployment) queryNetwork(networkName string) (*Network, error) {
	if n, found := d.Networks[networkName]; found {
		return n, nil
	}
	return nil, fmt.Errorf("network %s not found", networkName)
}

func (d *Deployment) QueryAllContainersInAllGroups(ctx context.Context) (ContainerList, error) {
	return containerMapToList(d.queryAllContainers()), nil
}

func (d *Deployment) QueryAllContainersInGroup(ctx context.Context, group string) (ContainerList, error) {
	ctMap, err := d.queryAllContainersInGroup(group)
	if err != nil {
		return nil, err
	}
	return containerMapToList(ctMap), nil
}

func (d *Deployment) QueryContainer(ctx context.Context, group, container string) (ContainerList, error) {
	ct, err := d.queryContainer(config.ContainerReference{Group: group, Container: container})
	if err != nil {
		return nil, err
	}
	return ContainerList{ct}, nil
}

func (d *Deployment) QueryNetwork(ctx context.Context, network string) (NetworkList, error) {
	net, err := d.queryNetwork(network)
	if err != nil {
		return nil, err
	}
	return NetworkList{net}, nil
}

func (d *Deployment) updateGroupsOrder() {
	d.GroupsOrder = make([]string, 0)
	for g := range d.Groups {
		d.GroupsOrder = append(d.GroupsOrder, g)
	}
	sort.Slice(d.GroupsOrder, func(i, j int) bool {
		g1 := d.GroupsOrder[i]
		g2 := d.GroupsOrder[j]
		return g1 < g2
	})
}

func (d *Deployment) updateNetworksOrder() {
	d.NetworksOrder = make([]string, 0)
	for n := range d.Networks {
		d.NetworksOrder = append(d.NetworksOrder, n)
	}
	sort.Slice(d.NetworksOrder, func(i, j int) bool {
		n1 := d.NetworksOrder[i]
		n2 := d.NetworksOrder[j]
		return n1 < n2
	})
}

func (d *Deployment) String() string {
	var sb strings.Builder

	sb.WriteString("Deployment{Groups:[")
	if len(d.GroupsOrder) == 0 {
		sb.WriteString("empty")
	} else {
		sb.WriteString(d.Groups[d.GroupsOrder[0]].String())
		for i := 1; i < len(d.GroupsOrder); i++ {
			sb.WriteString(fmt.Sprintf(", %s", d.Groups[d.GroupsOrder[i]]))
		}
	}

	sb.WriteString("], Networks:[")
	if len(d.NetworksOrder) == 0 {
		sb.WriteString("empty]}")
	} else {
		sb.WriteString(d.Networks[d.NetworksOrder[0]].String())
		for i := 1; i < len(d.NetworksOrder); i++ {
			sb.WriteString(fmt.Sprintf(", %s", d.Networks[d.NetworksOrder[i]]))
		}
		sb.WriteString("]}")
	}
	return sb.String()

}
