package cmds

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicommon"
	"github.com/tuxdudehomelab/homelab/internal/cli/errors"
	"github.com/tuxdudehomelab/homelab/internal/utils"
)

const (
	showConfigCmdStr = "show-config"
)

func ShowConfigCmd(ctx context.Context, globalOptions *clicommon.GlobalCmdOptions) *cobra.Command {
	return &cobra.Command{
		Use:     showConfigCmdStr,
		GroupID: clicommon.ConfigCmdGroupID,
		Short:   "Shows the homelab config",
		Long:    `Displays the homelab configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			err := execShowConfigCmd(ctx, cmd, args, globalOptions)
			if err != nil {
				return errors.NewHomelabRuntimeError(err)
			}
			return nil
		},
	}
}

func execShowConfigCmd(ctx context.Context, cmd *cobra.Command, args []string, globalOptions *clicommon.GlobalCmdOptions) error {
	dep, err := deploymentFromCommand(ctx, "show-config", globalOptions.CLIConfig, globalOptions.ConfigsDir)
	if err != nil {
		return err
	}

	log(ctx).Infof("Homelab config:\n%s", utils.PrettyPrintJSON(dep.Config))
	return nil
}
