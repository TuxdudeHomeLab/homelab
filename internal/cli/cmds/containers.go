package cmds

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicommon"
	"github.com/tuxdudehomelab/homelab/internal/cli/cmds/containers"
)

const (
	containersCmdStr = "containers"
)

func ContainersCmd(ctx context.Context, opts *clicommon.GlobalCmdOptions) *cobra.Command {
	cmd := buildContainersCmd(ctx)
	cmd.AddCommand(containers.StartCmd(ctx, opts))
	cmd.AddCommand(containers.StopCmd(ctx, opts))
	cmd.AddCommand(containers.PurgeCmd(ctx, opts))
	return cmd
}

func buildContainersCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:     containersCmdStr,
		GroupID: clicommon.ContainersCmdGroupID,
		Short:   "Homelab deployment container related commands",
		Long:    `Manipulate deployment of containers within one or more containers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("homelab container sub-command is required")
		},
	}
}
