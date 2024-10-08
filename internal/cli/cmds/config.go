package cmds

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicommon"
	"github.com/tuxdudehomelab/homelab/internal/cli/cmds/config"
)

const (
	configCmdStr = "config"
)

func ShowConfigCmd(ctx context.Context, globalOptions *clicommon.GlobalCmdOptions) *cobra.Command {
	cmd := buildShowConfigCmd(ctx)
	cmd.AddCommand(config.ShowConfigCmd(ctx, globalOptions))
	return cmd
}

func buildShowConfigCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:     configCmdStr,
		GroupID: clicommon.ConfigCmdGroupID,
		Short:   "Homelab config related commands",
		Long:    `Homelab configuration related commands.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("homelab config sub-command is required")
		},
	}
}
