package main

import "github.com/spf13/cobra"

func buildShowConfigCmd(globalOptions *globalCmdOptions) *cobra.Command {
	return &cobra.Command{
		Use:     showConfigCmdStr,
		GroupID: configCmdGroupID,
		Short:   "Shows the homelab config",
		Long:    `Displays the homelab configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return execShowConfigCmd(cmd, args, globalOptions)
		},
	}
}

func execShowConfigCmd(cmd *cobra.Command, args []string, globalOptions *globalCmdOptions) error {
	configsPath, err := homelabConfigsPath(globalOptions.cliConfig, globalOptions.configsDir)
	if err != nil {
		return err
	}

	config := HomelabConfig{}
	err = config.parse(configsPath)
	if err != nil {
		return err
	}

	log.Infof("Homelab config:\n%s", prettyPrintJSON(config))
	return nil
}
