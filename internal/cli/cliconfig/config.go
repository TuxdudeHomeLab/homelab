package cliconfig

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/tuxdudehomelab/homelab/internal/utils"
)

type CLIConfig struct {
	HomelabCLIConfig struct {
		ConfigsPath string `yaml:"configsPath,omitempty"`
	} `yaml:"homelab,omitempty"`
}

func (c *CLIConfig) parse(ctx context.Context, path string) error {
	configFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open homelab CLI config file, reason: %w", err)
	}
	defer configFile.Close()

	dec := yaml.NewDecoder(configFile)
	dec.KnownFields(true)
	err = dec.Decode(c)
	if err != nil {
		return fmt.Errorf("failed to parse homelab CLI config, reason: %w", err)
	}

	log(ctx).Tracef("Homelab CLI Config:\n%s\n", utils.PrettyPrintYAML(c))
	return nil
}
