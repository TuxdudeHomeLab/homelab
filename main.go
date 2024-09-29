// Command homelab is used to manage the configuration, deployment, and
// orchestration of a group of docker containers on a given host.
package main

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdude/zzzlogi"
	"github.com/tuxdudehomelab/homelab/internal/cli"
	clierrors "github.com/tuxdudehomelab/homelab/internal/cli/errors"
	"github.com/tuxdudehomelab/homelab/internal/cli/version"
	"github.com/tuxdudehomelab/homelab/internal/inspect"
	"github.com/tuxdudehomelab/homelab/internal/log"
)

const (
	homelabInspectLevelEnvVar   = "HOMELAB_INSPECT_LEVEL"
	homelabInspectLevelEnvDebug = "debug"
	homelabInspectLevelEnvTrace = "trace"
)

var (
	// The package information will be populated by the build system.
	pkgVersion   = "unset"
	pkgCommit    = "unset"
	pkgTimestamp = "unset"
)

func buildLogger(ctx context.Context, dest io.Writer) zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	config.Dest = dest
	lvl := inspect.HomelabInspectLevelFromContext(ctx)
	switch lvl {
	case inspect.HomelabInspectLevelTrace:
		config.MaxLevel = zzzlog.LvlTrace
		config.SkipCallerInfo = false
	case inspect.HomelabInspectLevelDebug:
		config.MaxLevel = zzzlog.LvlDebug
		config.SkipCallerInfo = false
	default:
		config.MaxLevel = zzzlog.LvlInfo
		config.SkipCallerInfo = true
	}
	return zzzlog.NewLogger(config)
}

func updateHomelabInspectLevel(ctx context.Context) context.Context {
	val, isVarSet := os.LookupEnv(homelabInspectLevelEnvVar)
	if isVarSet {
		if val == homelabInspectLevelEnvTrace {
			return inspect.WithHomelabInspectLevel(ctx, inspect.HomelabInspectLevelTrace)
		} else if val == homelabInspectLevelEnvDebug {
			return inspect.WithHomelabInspectLevel(ctx, inspect.HomelabInspectLevelDebug)
		}
	}
	return ctx
}

func runWithContext(ctx context.Context, outW io.Writer, errW io.Writer, args ...string) int {
	ctx = updateHomelabInspectLevel(ctx)
	ctx = log.WithLogger(ctx, buildLogger(ctx, outW))
	err := cli.Exec(ctx, outW, errW, args...)
	if err == nil {
		return 0
	}

	// Only log homelab runtime errors. Other errors are from cobra flag
	// and command parsing. These errors are displayed already by cobra
	// along with the usage.
	hre := &clierrors.HomelabRuntimeError{}
	if errors.As(err, &hre) {
		log.Log(ctx).Errorf("%s", hre)
	}
	return 1
}

func run() int {
	ctx := context.Background()
	ctx = version.WithVersionInfo(ctx, version.NewVersionInfo(pkgVersion, pkgCommit, pkgTimestamp))
	return runWithContext(ctx, os.Stdout, os.Stderr, os.Args[1:]...)
}

func main() {
	os.Exit(run())
}
