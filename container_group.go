package main

import "fmt"

type containerGroup struct {
	config     *ContainerGroupConfig
	containers containerMap
}

type containerGroupMap map[string]*containerGroup

func newContainerGroup(groupConfig *ContainerGroupConfig) *containerGroup {
	return &containerGroup{
		config:     groupConfig,
		containers: containerMap{},
	}
}

func (c *containerGroup) addContainer(config *ContainerConfig, globalConfig *GlobalConfig, networks networkMap, isAllowedOnCurrentHost bool) {
	ct := newContainer(c, config, globalConfig, networks, isAllowedOnCurrentHost)
	// TODO: Make ContainerReference the key instead.
	c.containers[config.Info] = ct
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
