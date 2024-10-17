package clicommon

import (
	"context"
	"fmt"
	"strings"

	"github.com/tuxdudehomelab/homelab/internal/deployment"
	"github.com/tuxdudehomelab/homelab/internal/docker"
)

const (
	AllNetworks = "all"
)

func ExecNetworksCmd(ctx context.Context, cmd, action, network string, dep *deployment.Deployment, fn func(context.Context, *deployment.Network, *docker.Client) error) error {
	res, err := dep.QueryNetwork(ctx, network)
	if err != nil {
		return fmt.Errorf("%s failed while querying networks, reason: %w", cmd, err)
	}

	dc := docker.NewClient(ctx)
	defer dc.Close()

	log(ctx).Debugf("%s command - %s: ", cmd, action)
	for _, n := range res {
		log(ctx).Debugf("%s", n.Name())
	}
	log(ctx).DebugEmpty()

	var errList []error
	for _, n := range res {
		// We ignore the errors to keep moving forward even if the action
		// fails on one or more networks.
		if err := fn(ctx, n, dc); err != nil {
			errList = append(errList, err)
		}
	}

	if len(errList) > 0 {
		var sb strings.Builder
		for i, e := range errList {
			sb.WriteString(fmt.Sprintf("\n%d - %s", i+1, e))
		}
		return fmt.Errorf("%s failed for %d networks, reason(s):%s", cmd, len(errList), sb.String())
	}
	return nil
}

func ExecCreateNetwork(ctx context.Context, n *deployment.Network, dc *docker.Client) error {
	created, err := n.Create(ctx, dc)
	if err == nil && !created {
		log(ctx).Warnf("Network %s not created since it already exists", n.Name())
		log(ctx).WarnEmpty()
	}
	return err
}

func ExecDeleteNetwork(ctx context.Context, n *deployment.Network, dc *docker.Client) error {
	deleted, err := n.Delete(ctx, dc)
	if err == nil && !deleted {
		log(ctx).Warnf("Network %s not deleted since it doesn't exist already", n.Name())
		log(ctx).WarnEmpty()
	}
	return err
}
