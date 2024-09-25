package main

import (
	"fmt"
	"sort"
	"strings"
)

type containerGroup struct {
	config          *ContainerGroupConfig
	containers      containerMap
	containersOrder []ContainerReference
}

type containerGroupMap map[string]*containerGroup

func newContainerGroup(groupConfig *ContainerGroupConfig) *containerGroup {
	return &containerGroup{
		config:     groupConfig,
		containers: containerMap{},
	}
}

func (c *containerGroup) addContainer(config *ContainerConfig, globalConfig *GlobalConfig, ips networkContainerIPList, isAllowedOnCurrentHost bool) {
	ct := newContainer(c, config, globalConfig, ips, isAllowedOnCurrentHost)
	c.containers[config.Info] = ct
}

func (c *containerGroup) name() string {
	return c.config.Name
}

func (c *containerGroup) updateContainersOrder() {
	c.containersOrder = make([]ContainerReference, 0)
	for ct := range c.containers {
		c.containersOrder = append(c.containersOrder, ct)
	}
	sort.Slice(c.containersOrder, func(i, j int) bool {
		ct1 := c.containersOrder[i]
		ct2 := c.containersOrder[j]
		return ct1.Container < ct2.Container
	})
}

func (c *containerGroup) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Group{Name:%s Containers:[", c.name()))
	if len(c.containersOrder) == 0 {
		sb.WriteString("empty]}")
		return sb.String()
	}

	sb.WriteString(c.containers[c.containersOrder[0]].String())
	for i := 1; i < len(c.containersOrder); i++ {
		sb.WriteString(fmt.Sprintf(", %s", c.containers[c.containersOrder[i]]))
	}
	sb.WriteString("]}")
	return sb.String()
}
