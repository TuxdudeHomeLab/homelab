package main

import (
	"context"
	"fmt"
)

type network struct {
	deployment          *deployment
	mode                networkMode
	bridgeModeConfig    *BridgeModeNetworkConfig
	containerModeConfig *ContainerModeNetworkConfig
}

type networkMap map[string]*network

const (
	networkModeUnknown networkMode = iota
	networkModeBridge
	networkModeContainer
)

type networkMode uint8

func newBridgeModeNetwork(dep *deployment, config *BridgeModeNetworkConfig) *network {
	n := network{
		deployment:       dep,
		mode:             networkModeBridge,
		bridgeModeConfig: config,
	}
	return &n
}

func newContainerModeNetwork(dep *deployment, config *ContainerModeNetworkConfig) *network {
	n := network{
		deployment:          dep,
		mode:                networkModeContainer,
		containerModeConfig: config,
	}
	return &n
}

func (n *network) create(ctx context.Context, docker *dockerClient) error {
	if !docker.networkExists(ctx, n.name()) {
		err := docker.createNetwork(ctx, n)
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO: Remove this after this function is used.
// nolint (unused)
func (n *network) delete(ctx context.Context, docker *dockerClient) error {
	if docker.networkExists(ctx, n.name()) {
		err := docker.deleteNetwork(ctx, n.name())
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *network) connectContainer(ctx context.Context, docker *dockerClient, containerName, ip string) error {
	return docker.connectContainerToBridgeModeNetwork(ctx, containerName, n.name(), ip)
}

// TODO: Remove this after this function is used.
// nolint (unused)
func (n *network) disconnectContainer(ctx context.Context, docker *dockerClient, containerName string) error {
	return docker.disconnectContainerFromNetwork(ctx, containerName, n.name())
}

func (n *network) name() string {
	if n.mode == networkModeBridge {
		return n.bridgeModeConfig.Name
	} else if n.mode == networkModeContainer {
		return n.containerModeConfig.Name
	} else {
		log.Fatalf("unknown network, possibly indicating a bug in the code!")
		return "{Network Unknown}"
	}
}

func (n *network) String() string {
	if n.mode == networkModeBridge {
		return fmt.Sprintf("{Network (Bridge) Name: %s}", n.name())
	} else if n.mode == networkModeContainer {
		return fmt.Sprintf("{Network (Container) Name: %s}", n.name())
	} else {
		log.Fatalf("unknown network, possibly indicating a bug in the code!")
		return "{Network Unknown}"
	}
}

func (n networkMap) String() string {
	return stringifyMap(n)
}
