package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const (
	startCmdStr = "start"
)

type startCmdOptions struct {
	allGroups bool
	group     string
	container string
}

func buildStartCmd(ctx context.Context, globalOptions *globalCmdOptions) *cobra.Command {
	options := startCmdOptions{}

	s := &cobra.Command{
		Use:     startCmdStr,
		GroupID: containersCmdGroupID,
		Short:   "Starts one or more containers",
		Long:    `Starts one or more containers as specified in the homelab configuration. Containers can be started individually, as a group or all groups.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			gFlag := cmd.Flag(groupFlagStr)
			cFlag := cmd.Flag(containerFlagStr)
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
				return newHomelabRuntimeError(err)
			}
			return nil
		},
	}
	s.Flags().BoolVar(
		&options.allGroups, allGroupsFlagStr, false, "Start containers in all groups.")
	s.Flags().StringVar(
		&options.group, groupFlagStr, "", "Start one or all containers in the specified group.")
	s.Flags().StringVar(
		&options.container, containerFlagStr, "", "Start the specified container.")
	return s
}

func execStartCmd(ctx context.Context, cmd *cobra.Command, args []string, options *startCmdOptions, globalOptions *globalCmdOptions) error {
	dep, err := deploymentFromCommand(ctx, "start", globalOptions.cliConfig, globalOptions.configsDir)
	if err != nil {
		return err
	}

	dockerClient, err := newDockerClient(ctx, dep.host.dockerPlatform, dep.host.arch)
	if err != nil {
		return err
	}
	defer dockerClient.close()

	res, err := dep.queryContainers(ctx, options.allGroups, options.group, options.container)
	if err != nil {
		return fmt.Errorf("start failed while querying containers, reason: %w", err)
	}

	log(ctx).Debugf("start command - Starting containers: ")
	for _, c := range res {
		log(ctx).Debugf("%s", c.name())
	}
	log(ctx).DebugEmpty()

	var errList []error
	for _, c := range res {
		// We ignore the errors to keep moving forward even if one or more
		// of the containers fail to start.
		started, err := c.start(ctx, dockerClient)
		if err != nil {
			errList = append(errList, err)
		}
		if !started {
			log(ctx).Warnf("Container %s not allowed to run on host %s", c.name(), dep.host.humanFriendlyHostName)
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
