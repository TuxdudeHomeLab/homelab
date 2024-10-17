package config

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicommon"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicontext"
	"github.com/tuxdudehomelab/homelab/internal/cli/errors"
	"github.com/tuxdudehomelab/homelab/internal/utils"
)

func ShowConfigCmd(ctx context.Context, opts *clicommon.GlobalCmdOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Shows the homelab config",
		Long:  `Displays the homelab configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			err := execShowConfigCmd(clicontext.HomelabContext(ctx), opts)
			if err != nil {
				return errors.NewHomelabRuntimeError(err)
			}
			return nil
		},
	}
}

func execShowConfigCmd(ctx context.Context, opts *clicommon.GlobalCmdOptions) error {
	dep, err := clicommon.BuildDeployment(ctx, "config show", opts)
	if err != nil {
		return err
	}

	log(ctx).Infof("Homelab config:\n%s", utils.PrettyPrintJSON(dep.Config))
	return nil
}
