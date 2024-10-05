package cmds

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicommon"
	"github.com/tuxdudehomelab/homelab/internal/cli/errors"
	"github.com/tuxdudehomelab/homelab/internal/deployment"
	"github.com/tuxdudehomelab/homelab/internal/docker"
	"github.com/tuxdudehomelab/homelab/internal/host"
)

const (
	startCmdStr = "start"
)

type startCmdOptions struct {
	cgOptions clicommon.ContainerGroupOptions
}

func StartCmd(ctx context.Context, globalOptions *clicommon.GlobalCmdOptions) *cobra.Command {
	options := startCmdOptions{}

	cmd := &cobra.Command{
		Use:     startCmdStr,
		GroupID: clicommon.ContainersCmdGroupID,
		Short:   "Starts one or more containers",
		Long:    `Starts one or more containers as specified in the homelab configuration. Containers can be started individually, as a group or all groups.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return clicommon.ValidateContainerGroupFlags(cmd, &options.cgOptions)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			err := execStartCmd(homelabContext(ctx), &options, globalOptions)
			if err != nil {
				return errors.NewHomelabRuntimeError(err)
			}
			return nil
		},
	}
	clicommon.AddContainerGroupFlags(cmd, &options.cgOptions)
	return cmd
}

func execStartCmd(ctx context.Context, options *startCmdOptions, globalOptions *clicommon.GlobalCmdOptions) error {
	dep, err := clicommon.BuildDeployment(ctx, "start", globalOptions)
	if err != nil {
		return err
	}

	h := host.MustHostInfo(ctx)
	return clicommon.ExecContainerGroupCmd(
		ctx,
		"start",
		"Starting containers",
		&options.cgOptions,
		dep,
		func(c *deployment.Container, dc *docker.DockerClient) error {
			started, err := c.Start(ctx, dc)
			if err == nil && !started {
				log(ctx).Warnf("Container %s not allowed to run on host %s", c.Name(), h.HumanFriendlyHostName)
				log(ctx).WarnEmpty()
			}
			return err
		},
	)
}
