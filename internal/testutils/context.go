package testutils

import (
	"context"

	"github.com/tuxdude/zzzlogi"
	"github.com/tuxdudehomelab/homelab/internal/docker"
	"github.com/tuxdudehomelab/homelab/internal/docker/fakedocker"
	"github.com/tuxdudehomelab/homelab/internal/host"
	"github.com/tuxdudehomelab/homelab/internal/host/fakehost"
	"github.com/tuxdudehomelab/homelab/internal/inspect"
	"github.com/tuxdudehomelab/homelab/internal/log"
)

type TestContextInfo struct {
	InspectLevel    inspect.HomelabInspectLevel
	Logger          zzzlogi.Logger
	DockerHost      docker.DockerAPIClient
	UseRealHostInfo bool
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
	if info.DockerHost != nil {
		ctx = docker.WithDockerAPIClient(ctx, info.DockerHost)
	}
	if !info.UseRealHostInfo {
		ctx = host.WithHostInfo(ctx, fakehost.NewFakeHostInfo())
	}
	return ctx
}
