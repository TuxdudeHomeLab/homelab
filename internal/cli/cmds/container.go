package cmds

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicommon"
	"github.com/tuxdudehomelab/homelab/internal/cli/cmds/container"
)

const (
	containerCmdStr = "container"
)

func ContainerCmd(ctx context.Context, opts *clicommon.GlobalCmdOptions) *cobra.Command {
	cmd := buildContainerCmd(ctx)
	cmd.AddCommand(container.StartCmd(ctx, opts))
	cmd.AddCommand(container.StopCmd(ctx, opts))
	cmd.AddCommand(container.PurgeCmd(ctx, opts))
	return cmd
}

func buildContainerCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:     containerCmdStr,
		GroupID: clicommon.ContainersCmdGroupID,
		Short:   "Homelab deployment container related commands",
		Long:    `Manipulate deployment of containers within one or more containers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("homelab container sub-command is required")
		},
	}
}
