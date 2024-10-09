package clicommon

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/cli/cliconfig"
)

const (
	cliConfigFlagStr  = "cli-config"
	configsDirFlagStr = "configs-dir"
)

type GlobalCmdOptions struct {
	cliConfig  string
	configsDir string
}

func configsPath(ctx context.Context, cmd string, opts *GlobalCmdOptions) (string, error) {
	configsPath, err := cliconfig.ConfigsPath(ctx, opts.cliConfig, opts.configsDir)
	if err != nil {
		return "", fmt.Errorf("%s failed while determining the configs path, reason: %w", cmd, err)
	}
	return configsPath, nil
}

func AddHomelabFlags(ctx context.Context, cmd *cobra.Command, opts *GlobalCmdOptions) {
	cmd.PersistentFlags().StringVar(
		&opts.cliConfig, cliConfigFlagStr, "", "The path to the Homelab CLI config")
	if cmd.MarkPersistentFlagFilename(cliConfigFlagStr) != nil {
		log(ctx).Fatalf("failed to mark --%s flag as filename flag", cliConfigFlagStr)
	}
	cmd.PersistentFlags().StringVar(
		&opts.configsDir, configsDirFlagStr, "", "The path to the directory containing the homelab configs")
	if cmd.MarkPersistentFlagDirname(configsDirFlagStr) != nil {
		log(ctx).Fatalf("failed to mark --%s flag as dirname flag", configsDirFlagStr)
	}
	cmd.MarkFlagsMutuallyExclusive(cliConfigFlagStr, configsDirFlagStr)
}
