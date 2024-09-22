package main

import (
	"context"
	"io"
	"os"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdude/zzzlogi"
)

var (
	fakeConfigEnv = newConfigEnv(fakeHostInfo)
)

func testContext() context.Context {
	return testContextWithLogger(testLogger())
}

func testContextWithLogger(logger zzzlogi.Logger) context.Context {
	return withLogger(context.Background(), logger)
}

func testLogger() zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	config.SkipCallerInfo = true
	return zzzlog.NewLogger(config)
}

func capturingTestLogger(w io.Writer) zzzlogi.Logger {
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
