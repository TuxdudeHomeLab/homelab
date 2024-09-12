package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type CLIConfig struct {
	HomelabCLIConfig struct {
		ConfigsPath string `yaml:"configsPath"`
	} `yaml:"homelab"`
}

func cliConfigPath(cliConfigFlag string) (string, error) {
	// Use the flag from the command line if present.
	if len(cliConfigFlag) > 0 {
		log.Debugf("Using Homelab CLI config path from command line flag: %s", cliConfigFlag)
		return cliConfigFlag, nil
	}
	// Fall back to the default path - "~/.homelab/config.yaml".
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to obtain the user's home directory for reading the homelab CLI config, reason: %w", err)
	}
	path, err := filepath.Abs(fmt.Sprintf(defaultCLIConfigPathFormat, homeDir))
	if err != nil {
		return "", fmt.Errorf("failed to determine absolute path of the homelab CLI config, reason: %w", err)
	}
	log.Debugf("Using default Homelab CLI config path: %s", path)
	return path, nil
}

func parseCLIConfig(cliConfigFlag string) (*CLIConfig, error) {
	path, err := cliConfigPath(cliConfigFlag)
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

	log.Tracef("Homelab CLI Config:\n%v\n", prettyPrintJSON(config))
	return &config, nil
}

func configsPath(cliConfigFlag string) (string, error) {
	config, err := parseCLIConfig(cliConfigFlag)
	if err != nil {
		return "", err
	}
	return config.HomelabCLIConfig.ConfigsPath, nil
}
