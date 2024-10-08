package cmds

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicommon"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicontext"
	"github.com/tuxdudehomelab/homelab/internal/cli/errors"
	"github.com/tuxdudehomelab/homelab/internal/deployment"
	"github.com/tuxdudehomelab/homelab/internal/docker"
)

const (
	purgeCmdStr = "purge"
)

type purgeCmdOptions struct {
	cgOptions clicommon.ContainerGroupOptions
}

func PurgeCmd(ctx context.Context, globalOptions *clicommon.GlobalCmdOptions) *cobra.Command {
	options := purgeCmdOptions{}

	cmd := &cobra.Command{
		Use:     purgeCmdStr,
		GroupID: clicommon.ContainersCmdGroupID,
		Short:   "Purges one or more containers",
		Long:    `Purges one or more containers as specified in the homelab configuration. Containers can be purged individually, as a group or all groups.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return clicommon.ValidateContainerGroupFlags(cmd, &options.cgOptions)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			err := execPurgeCmd(clicontext.HomelabContext(ctx), &options, globalOptions)
			if err != nil {
				return errors.NewHomelabRuntimeError(err)
			}
			return nil
		},
	}
	clicommon.AddContainerGroupFlags(cmd, &options.cgOptions)
	return cmd
}

func execPurgeCmd(ctx context.Context, options *purgeCmdOptions, globalOptions *clicommon.GlobalCmdOptions) error {
	dep, err := clicommon.BuildDeployment(ctx, "purge", globalOptions)
	if err != nil {
		return err
	}

	return clicommon.ExecContainerGroupCmd(
		ctx,
		"purge",
		"Purging containers",
		&options.cgOptions,
		dep,
		func(c *deployment.Container, dc *docker.Client) error {
			purged, err := c.Purge(ctx, dc)
			if err == nil && !purged {
				log(ctx).Warnf("Container %s cannot be purged since it was not found", c.Name())
				log(ctx).WarnEmpty()
			}
			return err
		},
	)
}
