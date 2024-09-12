package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

type startCmdOptions struct {
	allGroups bool
	group     string
	container string
}

func buildStartCmd(globalOptions *globalCmdOptions) *cobra.Command {
	options := startCmdOptions{}

	s := &cobra.Command{
		Use:   startCmdStr,
		Short: "Starts one or more containers",
		Long:  `Starts one or more containers as specified in the homelab configuration. Containers can be started individually, as a group or all groups.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			gFlag := cmd.Flag(groupFlagStr)
			cFlag := cmd.Flag(containerFlagStr)
			if !options.allGroups && !gFlag.Changed {
				return fmt.Errorf("--group flag must be specified when --all-groups is either not specified or set to false.")
			} else if !gFlag.Changed && !cFlag.Changed {
				return fmt.Errorf("when --all-groups is false, either --group flag must be specified or both --group and --container flags must be specified.")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return execStartCmd(cmd, args, &options, globalOptions)
		},
	}
	s.Flags().BoolVar(
		&options.allGroups, allGroupsFlagStr, false, "Start containers in all groups.")
	s.Flags().StringVar(
		&options.group, groupFlagStr, "", "Start one or all containers in the specified group.")
	s.Flags().StringVar(
		&options.container, containerFlagStr, "", "Start the specified container.")
	s.MarkFlagsMutuallyExclusive(allGroupsFlagStr, groupFlagStr)
	s.MarkFlagsMutuallyExclusive(allGroupsFlagStr, containerFlagStr)
	return s
}

func execStartCmd(cmd *cobra.Command, args []string, options *startCmdOptions, globalOptions *globalCmdOptions) error {
	configsPath, err := homelabConfigsPath(globalOptions.cliConfig, globalOptions.configsDir)
	if err != nil {
		return err
	}

	dep, err := buildDeployment(configsPath)
	if err != nil {
		return err
	}

	dockerClient, err := newDockerClient(dep.host.dockerPlatform, dep.host.arch)
	if err != nil {
		return err
	}
	defer dockerClient.close()

	res := queryContainers(dep, options.allGroups, options.group, options.container)
	log.Debugf("start command - Starting containers: ")
	for _, c := range res {
		log.Debugf("%s", c.name())
	}
	log.DebugEmpty()

	ctx := context.Background()
	for _, c := range res {
		// We ignore the errors to keep moving forward even if one or more
		// of the containers fail to start.
		_ = c.start(ctx, dockerClient)
	}

	return nil
}
