package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

const (
	homelabCmdStr = "homelab"

	cliConfigFlagStr  = "cli-config"
	configsDirFlagStr = "configs-dir"

	allGroupsFlagStr = "all-groups"
	groupFlagStr     = "group"
	containerFlagStr = "container"

	configCmdGroupID     = "config"
	containersCmdGroupID = "containers"
)

type globalCmdOptions struct {
	cliConfig  string
	configsDir string
}

func buildHomelabCmd(ctx context.Context, options *globalCmdOptions) *cobra.Command {
	h := &cobra.Command{
		Use:           homelabCmdStr,
		Version:       fmt.Sprintf("%s [Revision: %s @ %s]", pkgVersion, pkgCommit, pkgTimestamp),
		SilenceUsage:  false,
		SilenceErrors: false,
		Short:         "Homelab is a CLI for managing configuration and deployment of docket containers.",
		Long: `A CLI for managing both the configuration and deployment of groups of docker containers on a given host.

The configuration is managed using a yaml file. The configuration specifies the container groups and individual containers, their properties and how to deploy them.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: print ascii art.
			return fmt.Errorf("homelab sub-command is required")
		},
	}

	h.PersistentFlags().StringVar(
		&options.cliConfig, cliConfigFlagStr, "", "The path to the Homelab CLI config")
	if h.MarkPersistentFlagFilename(cliConfigFlagStr) != nil {
		log(ctx).Fatalf("failed to mark --%s flag as filename flag", cliConfigFlagStr)
	}
	h.PersistentFlags().StringVar(
		&options.configsDir, configsDirFlagStr, "", "The path to the directory containing the homelab configs")
	if h.MarkPersistentFlagDirname(configsDirFlagStr) != nil {
		log(ctx).Fatalf("failed to mark --%s flag as dirname flag", configsDirFlagStr)
	}
	h.MarkFlagsMutuallyExclusive(cliConfigFlagStr, configsDirFlagStr)

	h.AddGroup(
		&cobra.Group{
			ID:    configCmdGroupID,
			Title: "Configuration:",
		},
		&cobra.Group{
			ID:    containersCmdGroupID,
			Title: "Containers:",
		},
	)

	return h
}

func initHomelabCmd(ctx context.Context) *cobra.Command {
	globalOptions := globalCmdOptions{}
	homelabCmd := buildHomelabCmd(ctx, &globalOptions)
	homelabCmd.AddCommand(buildShowConfigCmd(ctx, &globalOptions))
	homelabCmd.AddCommand(buildStartCmd(ctx, &globalOptions))
	return homelabCmd
}

func execHomelabCmd(ctx context.Context) error {
	homelab := initHomelabCmd(ctx)
	return homelab.Execute()
}
