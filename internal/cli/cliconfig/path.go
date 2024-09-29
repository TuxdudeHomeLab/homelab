package cliconfig

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultCLIConfigPathFormat = "%s/.homelab/config.yaml"
)

func ConfigsPath(ctx context.Context, cliConfigFlag string, configsDirFlag string) (string, error) {
	// Use the flag from the command line if present.
	if len(configsDirFlag) > 0 {
		log(ctx).Debugf("Using Homelab configs path from command line flag: %s", configsDirFlag)
		return configsDirFlag, nil
	}
	path, err := configsPathFromConfig(ctx, cliConfigFlag)
	if err != nil {
		return "", err
	}

	log(ctx).Debugf("Using Homelab configs path from CLI config: %s", path)
	return path, nil
}

func configsPathFromConfig(ctx context.Context, cliConfigFlag string) (string, error) {
	path, err := configPath(ctx, cliConfigFlag)
	if err != nil {
		return "", err
	}

	config := CLIConfig{}
	if err := config.parse(ctx, path); err != nil {
		return "", err
	}
	p := config.HomelabCLIConfig.ConfigsPath
	if len(p) == 0 {
		return "", fmt.Errorf("homelab configs path setting in homelab.configsPath is empty/unset in the homelab CLI config")
	}
	return p, nil
}

func defaultPath(ctx context.Context) (string, error) {
	// The default CLI config path - "~/.homelab/config.yaml".
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to obtain the user's home directory for reading the homelab CLI config, reason: %w", err)
	}

	path, err := filepath.Abs(fmt.Sprintf(defaultCLIConfigPathFormat, homeDir))
	if err != nil {
		return "", fmt.Errorf("failed to determine absolute path of the homelab CLI config, reason: %w", err)
	}

	log(ctx).Debugf("Using default Homelab CLI config path: %s", path)
	return path, nil
}

func configPath(ctx context.Context, cliConfigFlag string) (string, error) {
	// Use the flag from the command line if present.
	if len(cliConfigFlag) > 0 {
		log(ctx).Debugf("Using Homelab CLI config path from command line flag: %s", cliConfigFlag)
		return cliConfigFlag, nil
	}
	// Fall back to the default path.
	return defaultPath(ctx)
}
