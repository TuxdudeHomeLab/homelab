package networks

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicommon"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicontext"
	"github.com/tuxdudehomelab/homelab/internal/cli/errors"
)

const (
	createCmdStr = "create"
)

func CreateCmd(ctx context.Context, opts *clicommon.GlobalCmdOptions) *cobra.Command {
	return &cobra.Command{
		Use:   createCmdStr,
		Short: "Creates one or more networks in the deployment",
		Long:  `Creates one or more networks that are specified in the homelab configuration.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("Expected exactly one network name argument to be specified, but found %d instead", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			err := execNetworksCreateCmd(clicontext.HomelabContext(ctx), args[0], opts)
			if err != nil {
				return errors.NewHomelabRuntimeError(err)
			}
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return clicommon.AutoCompleteNetworks(ctx, args, "networks start autocomplete", opts)
		},
	}
}

func execNetworksCreateCmd(ctx context.Context, network string, opts *clicommon.GlobalCmdOptions) error {
	dep, err := clicommon.BuildDeployment(ctx, "networks create", opts)
	if err != nil {
		return err
	}

	return clicommon.ExecNetworksCmd(
		ctx,
		"networks create",
		"Creating networks",
		network,
		dep,
		clicommon.ExecCreateNetwork,
	)
}
