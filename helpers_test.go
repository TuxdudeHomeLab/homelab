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

type testOSEnvMap map[string]string

type wrappedReader func(p []byte) (int, error)

func (actual wrappedReader) Read(p []byte) (int, error) {
	return actual(p)
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
	config.PanicInFatal = true
	return zzzlog.NewLogger(config)
}

func newCapturingTestLogger(w io.Writer) zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	config.SkipCallerInfo = true
	config.PanicInFatal = true
	config.Dest = w
	return zzzlog.NewLogger(config)
}

func newCapturingVanillaTestLogger(w io.Writer) zzzlogi.Logger {
	config := zzzlog.NewVanillaLoggerConfig()
	config.Dest = w
	config.PanicInFatal = true
	return zzzlog.NewLogger(config)
}

func setTestEnv(envs testOSEnvMap) {
	for k, v := range envs {
		err := os.Setenv(k, v)
		if err != nil {
			panic(err)
		}
	}
}

func clearTestEnv(envs testOSEnvMap) {
	for k := range envs {
		err := os.Unsetenv(k)
		if err != nil {
			panic(err)
		}
	}
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
