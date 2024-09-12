package main

import (
	"github.com/spf13/cobra"
)

var (
	homelabCmdStr    = "homelab"
	showConfigCmdStr = "show-config"
	startCmdStr      = "start"

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

func buildHomelabCmd(options *globalCmdOptions) *cobra.Command {
	h := &cobra.Command{
		Use:   homelabCmdStr,
		Short: "Homelab is a CLI for managing configuration and deployment of docket containers.",
		Long: `A CLI for managing both the configuration and deployment of groups of docker containers on a given host.

The configuration is managed using a yaml file. The configuration specifies the container groups and individual containers, their properties and how to deploy them.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Just display a help message that a sub-command needs to
			// be invoked.
			// TODO: print ascii art.
			_ = cmd.Help()
			return nil
		},
	}

	h.PersistentFlags().StringVar(
		&options.cliConfig, cliConfigFlagStr, "", "The path to the Homelab CLI config")
	if h.MarkPersistentFlagFilename(cliConfigFlagStr) != nil {
		log.Fatalf("failed to mark --%s flag as filename flag", cliConfigFlagStr)
	}

	h.PersistentFlags().StringVar(
		&options.configsDir, configsDirFlagStr, "", "The path to the directory containing the homelab configs")
	if h.MarkPersistentFlagDirname(configsDirFlagStr) != nil {
		log.Fatalf("failed to mark --%s flag as dirname flag", configsDirFlagStr)
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

func initHomelabCmd() *cobra.Command {
	globalOptions := globalCmdOptions{}
	homelabCmd := buildHomelabCmd(&globalOptions)
	homelabCmd.AddCommand(buildShowConfigCmd(&globalOptions))
	homelabCmd.AddCommand(buildStartCmd(&globalOptions))
	return homelabCmd
}

func execHomelabCmd() error {
	homelab := initHomelabCmd()
	return homelab.Execute()
}
