package cmds

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicommon"
	"github.com/tuxdudehomelab/homelab/internal/cli/cmds/group"
)

const (
	groupCmdStr = "group"
)

func GroupCmd(ctx context.Context, opts *clicommon.GlobalCmdOptions) *cobra.Command {
	cmd := buildGroupCmd(ctx)
	cmd.AddCommand(group.StartCmd(ctx, opts))
	cmd.AddCommand(group.StopCmd(ctx, opts))
	cmd.AddCommand(group.PurgeCmd(ctx, opts))
	return cmd
}

func buildGroupCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:     groupCmdStr,
		GroupID: clicommon.ContainersCmdGroupID,
		Short:   "Homelab deployment group related commands",
		Long:    `Manipulate deployment of containers within one or more groups.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("homelab group sub-command is required")
		},
	}
}
