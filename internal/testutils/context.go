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
)

type TestContextInfo struct {
	InspectLevel                    inspect.HomelabInspectLevel
	Logger                          zzzlogi.Logger
	Version                         *version.VersionInfo
	DockerHost                      docker.DockerAPIClient
	ContainerStopAndRemoveKillDelay time.Duration
	UseRealHostInfo                 bool
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
	if info.DockerHost != nil {
		ctx = docker.WithDockerAPIClient(ctx, info.DockerHost)
	}
	if info.ContainerStopAndRemoveKillDelay != 0 {
		ctx = docker.WithContainerStopAndRemoveKillDelay(ctx, info.ContainerStopAndRemoveKillDelay)
	}
	if !info.UseRealHostInfo {
		ctx = host.WithHostInfo(ctx, fakehost.NewFakeHostInfo())
	}
	return ctx
}
