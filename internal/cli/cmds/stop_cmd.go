package cmds

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicommon"
	"github.com/tuxdudehomelab/homelab/internal/cli/errors"
	"github.com/tuxdudehomelab/homelab/internal/deployment"
	"github.com/tuxdudehomelab/homelab/internal/docker"
)

const (
	stopCmdStr = "stop"
)

type stopCmdOptions struct {
	cgOptions clicommon.ContainerGroupOptions
}

func StopCmd(ctx context.Context, globalOptions *clicommon.GlobalCmdOptions) *cobra.Command {
	options := stopCmdOptions{}

	cmd := &cobra.Command{
		Use:     stopCmdStr,
		GroupID: clicommon.ContainersCmdGroupID,
		Short:   "Stops one or more containers",
		Long:    `Stops one or more containers as specified in the homelab configuration. Containers can be stopped individually, as a group or all groups.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return clicommon.ValidateContainerGroupFlags(cmd, &options.cgOptions)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			err := execStopCmd(homelabContext(ctx), &options, globalOptions)
			if err != nil {
				return errors.NewHomelabRuntimeError(err)
			}
			return nil
		},
	}
	clicommon.AddContainerGroupFlags(cmd, &options.cgOptions)
	return cmd
}

func execStopCmd(ctx context.Context, options *stopCmdOptions, globalOptions *clicommon.GlobalCmdOptions) error {
	dep, err := clicommon.BuildDeployment(ctx, "stop", globalOptions)
	if err != nil {
		return err
	}

	return clicommon.ExecContainerGroupCmd(
		ctx,
		"stop",
		"Stopping containers",
		&options.cgOptions,
		dep,
		func(c *deployment.Container, dc *docker.Client) error {
			stopped, err := c.Stop(ctx, dc)
			if err == nil && !stopped {
				log(ctx).Warnf("Container %s cannot be stopped since it was not found", c.Name())
				log(ctx).WarnEmpty()
			}
			return err
		},
	)
}
