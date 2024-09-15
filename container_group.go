package main

import "fmt"

type containerGroup struct {
	config     *ContainerGroupConfig
	deployment *deployment
	containers containerMap
}

type containerGroupMap map[string]*containerGroup

func newContainerGroup(dep *deployment, groupConfig *ContainerGroupConfig, containerConfigs *[]ContainerConfig) *containerGroup {
	g := containerGroup{
		deployment: dep,
		config:     groupConfig,
	}

	containers := make(containerMap)
	for _, c := range *containerConfigs {
		if c.Info.Group == g.name() {
			ct := newContainer(&g, &c)
			containers[ct.name()] = ct
		}
	}
	g.containers = containers

	return &g
}

func (c *containerGroup) name() string {
	return c.config.Name
}

func (c *containerGroup) String() string {
	return fmt.Sprintf("Group{Name:%s Containers:%s}", c.name(), c.containers)
}

func (c containerGroupMap) String() string {
	return stringifyMap(c)
}
