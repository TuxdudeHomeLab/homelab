package main

import (
	"context"
	"io"
	"net/netip"
	"os"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdude/zzzlogi"
)

const (
	fakeHostName              = "fakehost"
	fakeHumanFriendlyHostName = "FakeHost"
	fakeHostIP                = "10.76.77.78"
	fakeHostNumCPUs           = 32
	fakeHostOS                = "linux"
	fakeHostArch              = "amd64"
	fakeHostDockerPlatform    = "linux/amd64"
)

var (
	fakeHostInfo = &hostInfo{
		hostName:              fakeHostName,
		humanFriendlyHostName: fakeHumanFriendlyHostName,
		ip:                    netip.MustParseAddr(fakeHostIP),
		numCPUs:               fakeHostNumCPUs,
		os:                    fakeHostOS,
		arch:                  fakeHostArch,
		dockerPlatform:        fakeHostDockerPlatform,
	}
	fakeConfigEnv = newConfigEnv(fakeHostInfo)
)

func testContext() context.Context {
	return testContextWithLogger(testLogger())
}

func testContextWithLogger(logger zzzlogi.Logger) context.Context {
	return withLogger(context.Background(), logger)
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

func testLogger() zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	config.SkipCallerInfo = true
	return zzzlog.NewLogger(config)
}
