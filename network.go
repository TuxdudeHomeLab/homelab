package main

import (
	"context"
	"fmt"
	"net/netip"

	dnetwork "github.com/docker/docker/api/types/network"
)

type network struct {
	networkName    string
	priority       int
	mode           networkMode
	bridgeModeInfo *bridgeModeNetworkInfo
}

type bridgeModeNetworkInfo struct {
	hostInterfaceName string
	cidr              netip.Prefix
	gateway           netip.Addr
}

type networkMap map[string]*network

const (
	networkModeUnknown networkMode = iota
	networkModeBridge
	networkModeContainer
)

type networkMode uint8

func newBridgeModeNetwork(name string, priority int, info *bridgeModeNetworkInfo) *network {
	n := network{
		networkName:    name,
		priority:       priority,
		mode:           networkModeBridge,
		bridgeModeInfo: info,
	}
	return &n
}

func newContainerModeNetwork(name string, priority int) *network {
	n := network{
		networkName: name,
		priority:    priority,
		mode:        networkModeContainer,
	}
	return &n
}

func (n *network) create(ctx context.Context, docker *dockerClient) error {
	if n.mode == networkModeContainer {
		log(ctx).Debugf("Nothing to do for creating container mode network %s", n.name())
		return nil
	}

	// TODO: Validate that the existing network and the new network have
	// exactly the same properties if we choose to reuse the existing
	// network, and display a warning when they differ.
	if !docker.networkExists(ctx, n.name()) {
		log(ctx).Debugf("Creating network %s ...", n.name())
		err := docker.createNetwork(ctx, n.name(), n.createOptions())
		if err != nil {
			return err
		}
		log(ctx).Infof("Created network %s", n.name())
		log(ctx).InfoEmpty()
	} else {
		log(ctx).Debugf("Not re-creating existing network %s", n.name())
	}
	return nil
}

// TODO: Remove this after this function is used.
// nolint (unused)
func (n *network) delete(ctx context.Context, docker *dockerClient) error {
	if docker.networkExists(ctx, n.name()) {
		err := docker.removeNetwork(ctx, n.name())
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

func (n *network) createOptions() dnetwork.CreateOptions {
	if n.mode != networkModeBridge {
		panic("Only bridge mode network creation is possible")
	}

	return dnetwork.CreateOptions{
		Driver:     "bridge",
		Scope:      "local",
		EnableIPv6: newBool(false),
		IPAM: &dnetwork.IPAM{
			Driver: "default",
			Config: []dnetwork.IPAMConfig{
				{
					Subnet:  n.bridgeModeInfo.cidr.String(),
					Gateway: n.bridgeModeInfo.gateway.String(),
				},
			},
		},
		Internal:   false,
		Attachable: false,
		Ingress:    false,
		ConfigOnly: false,
		Options: map[string]string{
			"com.docker.network.bridge.enable_icc":           "true",
			"com.docker.network.bridge.enable_ip_masquerade": "true",
			"com.docker.network.bridge.host_binding_ipv4":    n.bridgeModeInfo.gateway.String(),
			"com.docker.network.bridge.name":                 n.bridgeModeInfo.hostInterfaceName,
			"com.docker.network.bridge.mtu":                  "1500",
		},
	}
}

func (n *network) name() string {
	return n.networkName
}

func (n *network) String() string {
	if n.mode == networkModeBridge {
		return fmt.Sprintf("{Network (Bridge) Name: %s}", n.name())
	} else if n.mode == networkModeContainer {
		return fmt.Sprintf("{Network (Container) Name: %s}", n.name())
	} else {
		panic("unknown network mode, possibly indicating a bug in the code!")
	}
}

func (n networkMap) String() string {
	return stringifyMap(n)
}
