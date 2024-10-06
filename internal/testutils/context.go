package testutils

import (
	"context"
	"time"

	"github.com/tuxdude/zzzlogi"
	"github.com/tuxdudehomelab/homelab/internal/cli/version"
	"github.com/tuxdudehomelab/homelab/internal/docker"
	"github.com/tuxdudehomelab/homelab/internal/docker/fakedocker"
	"github.com/tuxdudehomelab/homelab/internal/host"
	"github.com/tuxdudehomelab/homelab/internal/host/fakehost"
	"github.com/tuxdudehomelab/homelab/internal/inspect"
	"github.com/tuxdudehomelab/homelab/internal/log"
	"github.com/tuxdudehomelab/homelab/internal/user"
	"github.com/tuxdudehomelab/homelab/internal/user/fakeuser"
)

type TestContextInfo struct {
	InspectLevel            inspect.HomelabInspectLevel
	Logger                  zzzlogi.Logger
	Version                 *version.VersionInfo
	DockerHost              docker.APIClient
	ContainerPurgeKillDelay time.Duration
	UseRealUserInfo         bool
	UseRealHostInfo         bool
}

func NewVanillaTestContext() context.Context {
	return NewTestContext(
		&TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		})
}

func NewTestContext(info *TestContextInfo) context.Context {
	ctx := context.Background()
	if info.InspectLevel != inspect.HomelabInspectLevelNone {
		ctx = inspect.WithHomelabInspectLevel(ctx, info.InspectLevel)
	}
	if info.Logger != nil {
		ctx = log.WithLogger(ctx, info.Logger)
	} else {
		ctx = log.WithLogger(ctx, NewTestLogger())
	}
	if info.Version != nil {
		ctx = version.WithVersionInfo(ctx, info.Version)
	}
	if !info.UseRealUserInfo {
		ctx = user.WithUserInfo(ctx, fakeuser.NewFakeUserInfo())
	}
	if !info.UseRealHostInfo {
		ctx = host.WithHostInfo(ctx, fakehost.NewFakeHostInfo())
	}
	if info.DockerHost != nil {
		ctx = docker.WithAPIClient(ctx, info.DockerHost)
	}
	if info.ContainerPurgeKillDelay != 0 {
		ctx = docker.WithContainerPurgeKillDelay(ctx, info.ContainerPurgeKillDelay)
	}
	return ctx
}
