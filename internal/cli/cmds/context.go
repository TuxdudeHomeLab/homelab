package cmds

import (
	"context"

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
	if _, found := docker.APIClientFromContext(ctx); !found {
		ctx = docker.WithAPIClient(ctx, docker.MustRealAPIClient(ctx))
	}
	return ctx
}
