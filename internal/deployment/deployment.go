package deployment

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/tuxdudehomelab/homelab/internal/config"
	"github.com/tuxdudehomelab/homelab/internal/config/env"
	"github.com/tuxdudehomelab/homelab/internal/host"
	"github.com/tuxdudehomelab/homelab/internal/user"
)

type Deployment struct {
	Config                 *config.HomelabConfig
	Groups                 ContainerGroupMap
	GroupsOrder            []string
	Networks               NetworkMap
	NetworksOrder          []string
	Host                   *host.HostInfo
	allowedContainers      containerSet
	containerDockerConfigs containerDockerConfigMap
}

func FromConfigsPath(ctx context.Context, configsPath string) (*Deployment, error) {
	r, err := config.MergedConfigsReader(ctx, configsPath)
	if err != nil {
		return nil, err
	}

	return FromReader(ctx, r)
}

func FromReader(ctx context.Context, reader io.Reader) (*Deployment, error) {
	config := config.HomelabConfig{}
	err := config.Parse(ctx, reader)
	if err != nil {
		return nil, err
	}
	return FromConfig(ctx, &config)
}

func FromConfig(ctx context.Context, conf *config.HomelabConfig) (*Deployment, error) {
	u, found := user.UserInfoFromContext(ctx)
	if !found {
		u = user.NewUserInfo(ctx)
		ctx = user.WithUserInfo(ctx, u)
	}
	// TODO: Actually use the userInfo.
	_ = u

	h, found := host.HostInfoFromContext(ctx)
	if !found {
		h = host.NewHostInfo(ctx)
		ctx = host.WithHostInfo(ctx, h)
	}
	d := Deployment{
		Config:                 conf,
		Host:                   h,
		containerDockerConfigs: containerDockerConfigMap{},
	}

	env := env.NewConfigEnv(ctx)
	envWithGlobal, err := validateGlobalConfig(ctx, env, &conf.Global)
	if err != nil {
		return nil, err
	}

	d.allowedContainers, err = validateHostsConfig(conf.Hosts, h)
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

func (d *Deployment) QueryContainers(ctx context.Context, allGroups bool, group, container string) (containerList, error) {
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
		ct, err := d.queryContainer(config.ContainerReference{Group: group, Container: container})
		if err != nil {
			return nil, err
		}
		return containerList{ct}, nil
	}
	log(ctx).Fatalf("Invalid scenario, possibly indicating a bug in the code")
	return nil, nil
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
