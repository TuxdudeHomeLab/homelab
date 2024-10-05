package clicommon

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/cli/cliconfig"
)

const (
	allGroupsFlagStr = "all-groups"
	groupFlagStr     = "group"
	containerFlagStr = "container"
)

type GlobalCmdOptions struct {
	CLIConfig  string
	ConfigsDir string
}

type ContainerGroupOptions struct {
	allGroups bool
	group     string
	container string
}

func AddContainerGroupFlags(cmd *cobra.Command, options *ContainerGroupOptions) {
	cmd.Flags().BoolVar(
		&options.allGroups, allGroupsFlagStr, false, "Stop containers in all groups.")
	cmd.Flags().StringVar(
		&options.group, groupFlagStr, "", "Stop one or all containers in the specified group.")
	cmd.Flags().StringVar(
		&options.container, containerFlagStr, "", "Stop the specified container.")
}

func ValidateContainerGroupFlags(cmd *cobra.Command, options *ContainerGroupOptions) error {
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
}

func configsPath(ctx context.Context, cmd string, options *GlobalCmdOptions) (string, error) {
	configsPath, err := cliconfig.ConfigsPath(ctx, options.CLIConfig, options.ConfigsDir)
	if err != nil {
		return "", fmt.Errorf("%s failed while determining the configs path, reason: %w", cmd, err)
	}
	return configsPath, nil
}
