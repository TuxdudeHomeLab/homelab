package main

import (
	"context"
	"fmt"
	"sort"
)

type container struct {
	config *ContainerConfig
	group  *containerGroup
	ips    networkContainerIPMap
}

type containerIP struct {
	network *network
	IP      string
}

type containerMap map[string]*container
type containerList []*container
type networkContainerIPMap map[string]*containerIP

func newContainer(group *containerGroup, config *ContainerConfig) *container {
	c := container{group: group, config: config}
	cName := config.Name
	gName := group.name()

	ips := make(networkContainerIPMap)
	for _, n := range group.deployment.networks {
		if n.mode == networkModeBridge {
			for _, c := range n.bridgeModeConfig.Containers {
				if c.Container.Group == gName && c.Container.Container == cName {
					ips[n.name()] = newBridgeModeContainerIP(n, c.IP)
					break
				}
			}
		} else if n.mode == networkModeContainer {
			for _, c := range n.containerModeConfig.Containers {
				if c.Group == gName && c.Container == cName {
					ips[n.name()] = newContainerModeContainerIP(n)
					break
				}
			}
		}
	}
	c.ips = ips
	return &c
}

func (c *container) isAllowedOnCurrentHost() bool {
	return c.group.deployment.host.allowedContainers[c.name()]
}

func (c *container) start(ctx context.Context, docker *dockerClient) error {
	log.Debugf("Starting container %s ...", c.name())

	// 1. Validate the container is allowed to run on the current host.
	if !c.isAllowedOnCurrentHost() {
		return logToWarnAndReturn("Container %s not allowed to run on host '%s'", c.name(), c.group.deployment.host.humanFriendlyHostName)
	}

	err := c.startInternal(ctx, docker)
	if err != nil {
		return logToErrorAndReturn("Failed to start container '%s', reason:%v", c.name(), err)
	}

	log.Infof("Started container %s", c.name())
	log.InfoEmpty()
	return nil
}

func (c *container) startInternal(ctx context.Context, docker *dockerClient) error {
	// 1. Create the network for the container if it doesn't exist already.
	// TODO: Implement this.

	// 2. Execute any pre-start commands.
	// TODO: Implement this.

	// 3. Pull the container image.
	err := docker.pullImage(ctx, c.config.Image)
	if err != nil {
		return err
	}

	// 4. Purge (i.e. stop and remove) any previously existing containers
	// under the same name.
	// TODO: Implement this.

	// 5. Create the container.
	// TODO: Implement this.

	// 6. Start the created container.
	// TODO: Implement this.

	return nil
}

func (c *container) name() string {
	return containerName(c.group.name(), c.config.Name)
}

func (c *container) String() string {
	return fmt.Sprintf("Container{Name:%s}", c.name())
}

func (c containerMap) String() string {
	return stringifyMap(c)
}

func newBridgeModeContainerIP(network *network, ip string) *containerIP {
	return &containerIP{network: network, IP: ip}
}

func newContainerModeContainerIP(network *network) *containerIP {
	return &containerIP{network: network}
}

func containerName(group string, container string) string {
	return fmt.Sprintf("%s-%s", group, container)
}

func containerMapToList(cm containerMap) containerList {
	res := make(containerList, 0, len(cm))
	for _, c := range cm {
		res = append(res, c)
	}

	// Return containers sorted by order. Group order takes higher priority
	// over container order within the same group. If two containers still
	// have the same order at both the group and container levels, then
	// the container name is used to lexicographically sort the containers.
	sort.Slice(res, func(i, j int) bool {
		c1 := res[i]
		c2 := res[j]
		if c1.group.config.Order == c2.group.config.Order {
			if c1.config.Order == c2.config.Order {
				return c1.name() < c2.name()
			}
			return c1.config.Order < c2.config.Order
		} else {
			return c1.group.config.Order < c2.group.config.Order
		}
	})
	return res
}
