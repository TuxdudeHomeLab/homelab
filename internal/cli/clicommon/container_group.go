package clicommon

import (
	"context"
	"fmt"
	"strings"

	"github.com/tuxdudehomelab/homelab/internal/deployment"
	"github.com/tuxdudehomelab/homelab/internal/docker"
)

func ExecContainerGroupCmd(ctx context.Context, cmd, action string, options *ContainerGroupOptions, dep *deployment.Deployment, fn func(*deployment.Container, *docker.Client) error) error {
	res, err := dep.QueryContainers(ctx, options.allGroups, options.group, options.container)
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

	var errList []error
	for _, c := range res {
		// We ignore the errors to keep moving forward even if the action
		// fails on one or more containers.
		if err := fn(c, dc); err != nil {
			errList = append(errList, err)
		}
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
