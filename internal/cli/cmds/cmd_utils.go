package cmds

import (
	"context"
	"fmt"

	"github.com/tuxdudehomelab/homelab/internal/cli/cliconfig"
	"github.com/tuxdudehomelab/homelab/internal/deployment"
	"github.com/tuxdudehomelab/homelab/internal/docker"
	"github.com/tuxdudehomelab/homelab/internal/host"
	"github.com/tuxdudehomelab/homelab/internal/user"
)

func homelabContext(ctx context.Context) context.Context {
	if _, found := user.UserInfoFromContext(ctx); !found {
		ctx = user.WithUserInfo(ctx, user.NewUserInfo(ctx))
	}
	if _, found := host.HostInfoFromContext(ctx); !found {
		ctx = host.WithHostInfo(ctx, host.NewHostInfo(ctx))
	}
	if _, found := docker.DockerAPIClientFromContext(ctx); !found {
		ctx = docker.WithDockerAPIClient(ctx, docker.MustRealDockerAPIClient(ctx))
	}
	return ctx
}

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
