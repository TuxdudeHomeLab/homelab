package container

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
		Short: "Purges the container",
		Long:  `Purges the requested container as specified in the homelab configuration. The name is specified in the group/container format.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("Expected exactly one container name argument to be specified, but found %d instead", len(args))
			}
			_, _, err := validateContainerName(args[0])
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			err := execContainerPurgeCmd(clicontext.HomelabContext(ctx), args[0], globalOptions)
			if err != nil {
				return errors.NewHomelabRuntimeError(err)
			}
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return clicommon.AutoCompleteContainers(ctx, args, "container purge autocomplete", globalOptions)
		},
	}
}

func execContainerPurgeCmd(ctx context.Context, containerArg string, globalOptions *clicommon.GlobalCmdOptions) error {
	group, container := mustContainerName(containerArg)
	dep, err := clicommon.BuildDeployment(ctx, "container purge", globalOptions)
	if err != nil {
		return err
	}

	return clicommon.ExecContainerGroupCmd(
		ctx,
		"container purge",
		fmt.Sprintf("Purging container %s in group %s", container, group),
		group,
		container,
		dep,
		clicommon.ExecPurgeContainer,
	)
}
