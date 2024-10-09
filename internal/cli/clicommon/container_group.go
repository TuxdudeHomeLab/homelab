package clicommon

import (
	"context"
	"fmt"
	"strings"

	"github.com/tuxdudehomelab/homelab/internal/deployment"
	"github.com/tuxdudehomelab/homelab/internal/docker"
	"github.com/tuxdudehomelab/homelab/internal/host"
)

func ExecContainerGroupCmd(ctx context.Context, cmd, action, group, container string, dep *deployment.Deployment, fn func(context.Context, *deployment.Container, *host.HostInfo, *docker.Client) error) error {
	res, err := queryContainers(ctx, dep, group, container)
	if err != nil {
		return fmt.Errorf("%s failed while querying containers, reason: %w", cmd, err)
	}

	dc := docker.NewClient(ctx)
	defer dc.Close()

	log(ctx).Debugf("%s command - %s: ", cmd, action)
	for _, c := range res {
		log(ctx).Debugf("%s", c.Name())
	}
	log(ctx).DebugEmpty()

	h := host.MustHostInfo(ctx)
	var errList []error
	for _, ct := range res {
		// We ignore the errors to keep moving forward even if the action
		// fails on one or more containers.
		if err := fn(ctx, ct, h, dc); err != nil {
			errList = append(errList, err)
		}
	}

	if len(res) == 0 {
		log(ctx).Warnf("%s is a no-op since no containers were found matching the specified criteria", cmd)
	}

	if len(errList) > 0 {
		var sb strings.Builder
		for i, e := range errList {
			sb.WriteString(fmt.Sprintf("\n%d - %s", i+1, e))
		}
		return fmt.Errorf("%s failed for %d containers, reason(s):%s", cmd, len(errList), sb.String())
	}
	return nil
}

func ExecStartContainer(ctx context.Context, c *deployment.Container, h *host.HostInfo, dc *docker.Client) error {
	started, err := c.Start(ctx, dc)
	if err == nil && !started {
		log(ctx).Warnf("Container %s not allowed to run on host %s", c.Name(), h.HumanFriendlyHostName)
		log(ctx).WarnEmpty()
	}
	return err
}

func ExecStopContainer(ctx context.Context, c *deployment.Container, h *host.HostInfo, dc *docker.Client) error {
	_, err := c.Stop(ctx, dc)
	return err
}

func ExecPurgeContainer(ctx context.Context, c *deployment.Container, h *host.HostInfo, dc *docker.Client) error {
	purged, err := c.Purge(ctx, dc)
	if err == nil && !purged {
		log(ctx).Warnf("Container %s cannot be purged since it was not found", c.Name())
		log(ctx).WarnEmpty()
	}
	return err
}

func queryContainers(ctx context.Context, dep *deployment.Deployment, group, container string) (deployment.ContainerList, error) {
	if group == "all" {
		return dep.QueryAllContainersInAllGroups(ctx)
	}
	if group != "" && container == "" {
		return dep.QueryAllContainersInGroup(ctx, group)
	}
	return dep.QueryContainer(ctx, group, container)
}
