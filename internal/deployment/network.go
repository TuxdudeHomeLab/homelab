package deployment

import (
	"context"
	"fmt"
	"net/netip"

	dnetwork "github.com/docker/docker/api/types/network"
	"github.com/tuxdudehomelab/homelab/internal/config"
	"github.com/tuxdudehomelab/homelab/internal/docker"
	"github.com/tuxdudehomelab/homelab/internal/newutils"
)

type Network struct {
	networkName       string
	mode              networkMode
	bridgeModeInfo    *bridgeModeNetworkInfo
	containerModeInfo *containerModeNetworkInfo
}

type bridgeModeNetworkInfo struct {
	priority          int
	hostInterfaceName string
	cidr              netip.Prefix
	gateway           netip.Addr
}

type containerModeNetworkInfo struct {
	container config.ContainerReference
}

type NetworkMap map[string]*Network

const (
	networkModeUnknown networkMode = iota
	networkModeBridge
	networkModeContainer
)

type networkMode uint8

func newBridgeModeNetwork(name string, priority int, info *bridgeModeNetworkInfo) *Network {
	n := Network{
		networkName:    name,
		mode:           networkModeBridge,
		bridgeModeInfo: info,
	}
	return &n
}

func newContainerModeNetwork(name string, info *containerModeNetworkInfo) *Network {
	n := Network{
		networkName:       name,
		mode:              networkModeContainer,
		containerModeInfo: info,
	}
	return &n
}

func (n *Network) create(ctx context.Context, dc *docker.Client) error {
	if n.mode == networkModeContainer {
		log(ctx).Debugf("Nothing to do for creating container mode network %s", n.name())
		return nil
	}

	// TODO: Validate that the existing network and the new network have
	// exactly the same properties if we choose to reuse the existing
	// network, and display a warning when they differ.
	if !dc.NetworkExists(ctx, n.name()) {
		err := dc.CreateNetwork(ctx, n.name(), n.createOptions())
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

//nolint:nolintlint,unused // TODO: Remove this after this function is used.
func (n *Network) remove(ctx context.Context, dc *docker.Client) error {
	if dc.NetworkExists(ctx, n.name()) {
		err := dc.RemoveNetwork(ctx, n.name())
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *Network) connectContainer(ctx context.Context, dc *docker.Client, containerName, ip string) error {
	return dc.ConnectContainerToBridgeModeNetwork(ctx, containerName, n.name(), ip)
}

//nolint:nolintlint,unused // TODO: Remove this after this function is used.
func (n *Network) disconnectContainer(ctx context.Context, dc *docker.Client, containerName string) error {
	return dc.DisconnectContainerFromNetwork(ctx, containerName, n.name())
}

func (n *Network) createOptions() dnetwork.CreateOptions {
	if n.mode != networkModeBridge {
		panic("Only bridge mode network creation is possible")
	}

	return dnetwork.CreateOptions{
		Driver:     "bridge",
		Scope:      "local",
		EnableIPv6: newutils.NewBool(false),
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

func (n *Network) name() string {
	return n.networkName
}

func (n *Network) String() string {
	if n.mode == networkModeBridge {
		return fmt.Sprintf("{Network (Bridge) Name: %s}", n.name())
	} else if n.mode == networkModeContainer {
		return fmt.Sprintf("{Network (Container) Name: %s}", n.name())
	} else {
		panic("unknown network mode, possibly indicating a bug in the code!")
	}
}
