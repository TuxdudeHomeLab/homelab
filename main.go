// Command homelab is used to manage the configuration, deployment, and
// orchestration of a group of docker containers on a given host.
package main

import (
	"flag"
	"os"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdude/zzzlogi"
)

const (
	defaultCLIConfigPathFormat = "%s/.homelab/config.yaml"
)

var (
	log = buildLogger()
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
	if !validateFlags() {
		return 1
	}
	err := handleSubCommand()
	if err != nil {
		log.Errorf("%s", err)
		return 1
	}
	return 0
}

func main() {
	flag.Parse()
	os.Exit(run())
}
