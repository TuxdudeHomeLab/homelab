package main

import (
	"context"

	"github.com/tuxdude/zzzlogi"
)

type testContextInfo struct {
	logger          zzzlogi.Logger
	dockerHost      dockerAPIClient
	useRealHostInfo bool
}

func newVanillaTestContext() context.Context {
	return newTestContext(
		&testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		})
}

func newTestContext(info *testContextInfo) context.Context {
	ctx := context.Background()
	if info.logger != nil {
		ctx = withLogger(ctx, info.logger)
	} else {
		ctx = withLogger(ctx, newTestLogger())
	}
	if info.dockerHost != nil {
		ctx = withDockerAPIClient(ctx, info.dockerHost)
	}
	if !info.useRealHostInfo {
		ctx = withHostInfo(ctx, newFakeHostInfo())
	}
	return ctx
}
