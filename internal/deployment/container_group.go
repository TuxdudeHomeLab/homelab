package deployment

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tuxdudehomelab/homelab/internal/config"
)

type ContainerGroup struct {
	config          *config.ContainerGroup
	containers      containerMap
	containersOrder []config.ContainerReference
}

type ContainerGroupMap map[string]*ContainerGroup

func NewContainerGroup(groupConfig *config.ContainerGroup) *ContainerGroup {
	return &ContainerGroup{
		config:     groupConfig,
		containers: containerMap{},
	}
}

func (c *ContainerGroup) addContainer(config *config.Container, globalConfig *config.Global, endpoints networkEndpointList, isAllowedOnCurrentHost bool) {
	ct := newContainer(c, config, globalConfig, endpoints, isAllowedOnCurrentHost)
	c.containers[config.Info] = ct
}

func (c *ContainerGroup) name() string {
	return c.config.Name
}

func (c *ContainerGroup) updateContainersOrder() {
	c.containersOrder = make([]config.ContainerReference, 0)
	for ct := range c.containers {
		c.containersOrder = append(c.containersOrder, ct)
	}
	sort.Slice(c.containersOrder, func(i, j int) bool {
		ct1 := c.containersOrder[i]
		ct2 := c.containersOrder[j]
		return ct1.Container < ct2.Container
	})
}

func (c *ContainerGroup) String() string {
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
