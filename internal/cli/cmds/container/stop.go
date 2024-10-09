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
	stopCmdStr = "stop"
)

func StopCmd(ctx context.Context, globalOptions *clicommon.GlobalCmdOptions) *cobra.Command {
	return &cobra.Command{
		Use:   stopCmdStr,
		Short: "Stops the container",
		Long:  `Stops the requested container as specified in the homelab configuration. The name is specified in the group/container format.`,
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
			err := execContainerStopCmd(clicontext.HomelabContext(ctx), args[0], globalOptions)
			if err != nil {
				return errors.NewHomelabRuntimeError(err)
			}
			return nil
		},
	}
}

func execContainerStopCmd(ctx context.Context, containerArg string, globalOptions *clicommon.GlobalCmdOptions) error {
	group, container := mustContainerName(containerArg)
	dep, err := clicommon.BuildDeployment(ctx, "container stop", globalOptions)
	if err != nil {
		return err
	}

	return clicommon.ExecContainerGroupCmd(
		ctx,
		"container stop",
		fmt.Sprintf("Stopping container %s in group %s", container, group),
		group,
		container,
		dep,
		clicommon.ExecStopContainer,
	)
}
