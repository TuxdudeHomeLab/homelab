package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	defaultCLIConfigPathFormat = "%s/.homelab/config.yaml"
)

type CLIConfig struct {
	HomelabCLIConfig struct {
		ConfigsPath string `yaml:"configsPath,omitempty"`
	} `yaml:"homelab,omitempty"`
}

func defaultCLIConfigPath(ctx context.Context) (string, error) {
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

func cliConfigPath(ctx context.Context, cliConfigFlag string) (string, error) {
	// Use the flag from the command line if present.
	if len(cliConfigFlag) > 0 {
		log(ctx).Debugf("Using Homelab CLI config path from command line flag: %s", cliConfigFlag)
		return cliConfigFlag, nil
	}
	// Fall back to the default path.
	return defaultCLIConfigPath(ctx)
}

func parseCLIConfig(ctx context.Context, cliConfigFlag string) (*CLIConfig, error) {
	path, err := cliConfigPath(ctx, cliConfigFlag)
	if err != nil {
		return nil, err
	}

	configFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open homelab CLI config file, reason: %w", err)
	}
	defer configFile.Close()

	var config CLIConfig
	dec := yaml.NewDecoder(configFile)
	dec.KnownFields(true)
	err = dec.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse homelab CLI config, reason: %w", err)
	}

	log(ctx).Tracef("Homelab CLI Config:\n%v\n", prettyPrintJSON(config))
	return &config, nil
}

func configsPathFromCLIConfig(ctx context.Context, cliConfigFlag string) (string, error) {
	config, err := parseCLIConfig(ctx, cliConfigFlag)
	if err != nil {
		return "", err
	}
	p := config.HomelabCLIConfig.ConfigsPath
	if len(p) == 0 {
		return "", fmt.Errorf("homelab configs path setting in homelab.configsPath is empty/unset in the homelab CLI config")
	}
	return p, nil
}
