package main

import (
	"context"
	"fmt"
)

func deploymentFromCommand(ctx context.Context, command, cliConfig, configsDir string) (*deployment, error) {
	configsPath, err := homelabConfigsPath(ctx, cliConfig, configsDir)
	if err != nil {
		return nil, fmt.Errorf("%s failed while determining the configs path, reason: %w", command, err)
	}

	dep, err := buildDeploymentFromConfigsPath(ctx, configsPath)
	if err != nil {
		return nil, fmt.Errorf("%s failed while parsing the configs, reason: %w", command, err)
	}

	return dep, nil
}
