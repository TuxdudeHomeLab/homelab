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
	purgeCmdStr = "purge"
)

func PurgeCmd(ctx context.Context, globalOptions *clicommon.GlobalCmdOptions) *cobra.Command {
	return &cobra.Command{
		Use:   purgeCmdStr,
		Short: "Purges one or more containers in the group",
		Long:  `Purges one or more containers in the requested group as specified in the homelab configuration. Containers can be purged individually, as a group or all groups (by using 'all' as the group name).`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("Expected exactly one group name argument to be specified, but found %d instead", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			err := execGroupPurgeCmd(clicontext.HomelabContext(ctx), args[0], globalOptions)
			if err != nil {
				return errors.NewHomelabRuntimeError(err)
			}
			return nil
		},
	}
}

func execGroupPurgeCmd(ctx context.Context, group string, globalOptions *clicommon.GlobalCmdOptions) error {
	dep, err := clicommon.BuildDeployment(ctx, "group purge", globalOptions)
	if err != nil {
		return err
	}

	var action string
	if group == "all" {
		action = "Purging containers in all groups"
	} else {
		action = fmt.Sprintf("Purging containers in group %s", group)
	}
	return clicommon.ExecContainerGroupCmd(
		ctx,
		"group purge",
		action,
		group,
		"",
		dep,
		clicommon.ExecPurgeContainer,
	)
}
