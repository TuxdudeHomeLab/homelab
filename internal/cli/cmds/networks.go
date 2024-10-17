package cmds

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicommon"
	"github.com/tuxdudehomelab/homelab/internal/cli/cmds/networks"
)

func NetworksCmd(ctx context.Context, opts *clicommon.GlobalCmdOptions) *cobra.Command {
	cmd := buildNetworksCmd(ctx)
	cmd.AddCommand(networks.CreateCmd(ctx, opts))
	cmd.AddCommand(networks.DeleteCmd(ctx, opts))
	return cmd
}

func buildNetworksCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:     "networks",
		GroupID: clicommon.NetworksCmdGroupID,
		Short:   "Homelab network related commands",
		Long:    `Manipulate networks within the deployment.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("homelab networks sub-command is required")
		},
	}
}
