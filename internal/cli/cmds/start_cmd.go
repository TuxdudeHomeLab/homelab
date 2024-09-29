package cmds

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicommon"
	"github.com/tuxdudehomelab/homelab/internal/cli/errors"
	"github.com/tuxdudehomelab/homelab/internal/docker"
)

const (
	startCmdStr = "start"
)

type startCmdOptions struct {
	allGroups bool
	group     string
	container string
}

func StartCmd(ctx context.Context, globalOptions *clicommon.GlobalCmdOptions) *cobra.Command {
	options := startCmdOptions{}

	s := &cobra.Command{
		Use:     startCmdStr,
		GroupID: clicommon.ContainersCmdGroupID,
		Short:   "Starts one or more containers",
		Long:    `Starts one or more containers as specified in the homelab configuration. Containers can be started individually, as a group or all groups.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			gFlag := cmd.Flag(clicommon.GroupFlagStr)
			cFlag := cmd.Flag(clicommon.ContainerFlagStr)
			if options.allGroups && gFlag.Changed {
				return fmt.Errorf("--group flag cannot be specified when all-groups is true")
			}
			if options.allGroups && cFlag.Changed {
				return fmt.Errorf("--container flag cannot be specified when all-groups is true")
			}
			if !options.allGroups && !gFlag.Changed && cFlag.Changed {
				return fmt.Errorf("when --all-groups is false, --group flag must be specified when specifying the --container flag")
			}
			if !options.allGroups && !gFlag.Changed {
				return fmt.Errorf("--group flag must be specified when --all-groups is false")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			err := execStartCmd(ctx, cmd, args, &options, globalOptions)
			if err != nil {
				return errors.NewHomelabRuntimeError(err)
			}
			return nil
		},
	}
	s.Flags().BoolVar(
		&options.allGroups, clicommon.AllGroupsFlagStr, false, "Start containers in all groups.")
	s.Flags().StringVar(
		&options.group, clicommon.GroupFlagStr, "", "Start one or all containers in the specified group.")
	s.Flags().StringVar(
		&options.container, clicommon.ContainerFlagStr, "", "Start the specified container.")
	return s
}

func execStartCmd(ctx context.Context, cmd *cobra.Command, args []string, options *startCmdOptions, globalOptions *clicommon.GlobalCmdOptions) error {
	dep, err := deploymentFromCommand(ctx, "start", globalOptions.CLIConfig, globalOptions.ConfigsDir)
	if err != nil {
		return err
	}

	dockerClient, err := docker.NewDockerClient(ctx, dep.Host.DockerPlatform, dep.Host.Arch)
	if err != nil {
		return err
	}
	defer dockerClient.Close()

	res, err := dep.QueryContainers(ctx, options.allGroups, options.group, options.container)
	if err != nil {
		return fmt.Errorf("start failed while querying containers, reason: %w", err)
	}

	log(ctx).Debugf("start command - Starting containers: ")
	for _, c := range res {
		log(ctx).Debugf("%s", c.Name())
	}
	log(ctx).DebugEmpty()

	var errList []error
	for _, c := range res {
		// We ignore the errors to keep moving forward even if one or more
		// of the containers fail to start.
		started, err := c.Start(ctx, dockerClient)
		if err != nil {
			errList = append(errList, err)
		}
		if !started {
			log(ctx).Warnf("Container %s not allowed to run on host %s", c.Name(), dep.Host.HumanFriendlyHostName)
		}
	}

	if len(errList) > 0 {
		var sb strings.Builder
		for i, e := range errList {
			sb.WriteString(fmt.Sprintf("\n%d - %s", i+1, e))
		}
		return fmt.Errorf("start failed for %d containers, reason(s):%s", len(errList), sb.String())
	}
	return nil
}
