package cmds

import (
	"context"
	"fmt"

	"github.com/tuxdudehomelab/homelab/internal/cli/cliconfig"
	"github.com/tuxdudehomelab/homelab/internal/deployment"
)

func deploymentFromCommand(ctx context.Context, command, cliConfig, configsDir string) (*deployment.Deployment, error) {
	configsPath, err := cliconfig.ConfigsPath(ctx, cliConfig, configsDir)
	if err != nil {
		return nil, fmt.Errorf("%s failed while determining the configs path, reason: %w", command, err)
	}

	dep, err := deployment.FromConfigsPath(ctx, configsPath)
	if err != nil {
		return nil, fmt.Errorf("%s failed while parsing the configs, reason: %w", command, err)
	}

	return dep, nil
}
