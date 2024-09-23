package main

import (
	"context"
	"io"
	"os"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdude/zzzlogi"
)

var (
	// TODO: Remove this after configEnv gets actually used in the code.
	_ = newConfigEnv(newFakeHostInfo())
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

func newTestLogger() zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	config.SkipCallerInfo = true
	return zzzlog.NewLogger(config)
}

func newCapturingTestLogger(w io.Writer) zzzlogi.Logger {
	config := zzzlog.NewVanillaLoggerConfig()
	config.Dest = w
	return zzzlog.NewLogger(config)
}

func pwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}

func newInt(i int) *int {
	return &i
}
