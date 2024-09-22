package main

import (
	"context"

	"github.com/spf13/cobra"
)

const (
	showConfigCmdStr = "show-config"
)

func buildShowConfigCmd(ctx context.Context, globalOptions *globalCmdOptions) *cobra.Command {
	return &cobra.Command{
		Use:     showConfigCmdStr,
		GroupID: configCmdGroupID,
		Short:   "Shows the homelab config",
		Long:    `Displays the homelab configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			err := execShowConfigCmd(ctx, cmd, args, globalOptions)
			if err != nil {
				return newHomelabRuntimeError(err)
			}
			return nil
		},
	}
}

func execShowConfigCmd(ctx context.Context, cmd *cobra.Command, args []string, globalOptions *globalCmdOptions) error {
	configsPath, err := homelabConfigsPath(ctx, globalOptions.cliConfig, globalOptions.configsDir)
	if err != nil {
		return err
	}

	config := HomelabConfig{}
	err = config.parseConfigs(ctx, configsPath)
	if err != nil {
		return err
	}

	log(ctx).Infof("Homelab config:\n%s", prettyPrintJSON(config))
	return nil
}
