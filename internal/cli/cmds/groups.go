package cmds

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicommon"
	"github.com/tuxdudehomelab/homelab/internal/cli/cmds/groups"
)

const (
	groupsCmdStr = "groups"
)

func GroupsCmd(ctx context.Context, opts *clicommon.GlobalCmdOptions) *cobra.Command {
	cmd := buildGroupsCmd(ctx)
	cmd.AddCommand(groups.StartCmd(ctx, opts))
	cmd.AddCommand(groups.StopCmd(ctx, opts))
	cmd.AddCommand(groups.PurgeCmd(ctx, opts))
	return cmd
}

func buildGroupsCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:     groupsCmdStr,
		GroupID: clicommon.ContainersCmdGroupID,
		Short:   "Homelab deployment group related commands",
		Long:    `Manipulate deployment of containers within one or more groups.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("homelab group sub-command is required")
		},
	}
}
