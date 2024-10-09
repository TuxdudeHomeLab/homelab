package group

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicommon"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicontext"
	"github.com/tuxdudehomelab/homelab/internal/cli/errors"
)

const (
	startCmdStr = "start"
)

func StartCmd(ctx context.Context, globalOptions *clicommon.GlobalCmdOptions) *cobra.Command {
	return &cobra.Command{
		Use:   startCmdStr,
		Short: "Starts one or more containers in the group",
		Long:  `Starts one or more containers in the requested group as specified in the homelab configuration. Containers can be started individually, as a group or all groups (by using 'all' as the group name).`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("Expected exactly one group name argument to be specified, but found %d instead", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			err := execGroupStartCmd(clicontext.HomelabContext(ctx), args[0], globalOptions)
			if err != nil {
				return errors.NewHomelabRuntimeError(err)
			}
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return clicommon.AutoCompleteGroups(ctx, args, "group start autocomplete", globalOptions)
		},
	}
}

func execGroupStartCmd(ctx context.Context, group string, globalOptions *clicommon.GlobalCmdOptions) error {
	dep, err := clicommon.BuildDeployment(ctx, "group start", globalOptions)
	if err != nil {
		return err
	}

	var action string
	if group == "all" {
		action = "Starting containers in all groups"
	} else {
		action = fmt.Sprintf("Starting containers in group %s", group)
	}

	return clicommon.ExecContainerGroupCmd(
		ctx,
		"group start",
		action,
		group,
		"",
		dep,
		clicommon.ExecStartContainer,
	)
}
