// Command homelab is used to manage the configuration, deployment, and
// orchestration of a group of docker containers on a given host.
package main

import (
	"context"
	"errors"
	"os"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdude/zzzlogi"
)

var (
	// The package information will be populated by the build system.
	pkgVersion   = "unset"
	pkgCommit    = "unset"
	pkgTimestamp = "unset"
)

func buildLogger() zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	if isLogLevelTrace() {
		config.MaxLevel = zzzlog.LvlTrace
		config.SkipCallerInfo = false
	} else if isLogLevelDebug() {
		config.MaxLevel = zzzlog.LvlDebug
		config.SkipCallerInfo = false
	} else {
		config.MaxLevel = zzzlog.LvlInfo
		config.SkipCallerInfo = true
	}
	return zzzlog.NewLogger(config)
}

func run() int {
	logger := buildLogger()
	ctx := withLogger(context.Background(), logger)
	err := execHomelabCmd(ctx, os.Stdout, os.Stderr, os.Args...)
	if err == nil {
		return 0
	}

	// Only log homelab runtime errors. Other errors are from cobra flag
	// and command parsing. These errors are displayed already by cobra
	// along with the usage.
	hre := &homelabRuntimeError{}
	if errors.As(err, &hre) {
		logger.Errorf("%s", hre)
	}
	return 1
}

func main() {
	os.Exit(run())
}
