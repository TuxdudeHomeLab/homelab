package main

import (
	"context"
	"fmt"
)

type network struct {
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

func newBridgeModeNetwork(config *BridgeModeNetworkConfig) *network {
	n := network{
		mode:             networkModeBridge,
		bridgeModeConfig: config,
	}
	return &n
}

func newContainerModeNetwork(config *ContainerModeNetworkConfig) *network {
	n := network{
		mode:                networkModeContainer,
		containerModeConfig: config,
	}
	return &n
}

func (n *network) create(ctx context.Context, docker *dockerClient) error {
	// TODO: Validate that the existing network and the new network have
	// exactly the same properties if we choose to reuse the existing
	// network, and display a warning when they differ.
	if !docker.networkExists(ctx, n.name()) {
		log.Debugf("Creating network %s ...", n.name())
		err := docker.createNetwork(ctx, n)
		if err != nil {
			return err
		}
		log.Infof("Created network %s", n.name())
		log.InfoEmpty()
	} else {
		log.Debugf("Not re-creating existing network %s", n.name())
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
