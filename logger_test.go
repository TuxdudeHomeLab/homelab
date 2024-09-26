package main

import (
	"io"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdude/zzzlogi"
)

func newTestLogger() zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	config.SkipCallerInfo = true
	config.PanicInFatal = true
	return zzzlog.NewLogger(config)
}

func newCapturingTestLogger(lvl zzzlog.Level, w io.Writer) zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	config.MaxLevel = lvl
	config.SkipCallerInfo = true
	config.PanicInFatal = true
	config.Dest = w
	return zzzlog.NewLogger(config)
}

func newCapturingVanillaTestLogger(lvl zzzlog.Level, w io.Writer) zzzlogi.Logger {
	config := zzzlog.NewVanillaLoggerConfig()
	config.MaxLevel = lvl
	config.Dest = w
	config.PanicInFatal = true
	return zzzlog.NewLogger(config)
}

func newLogLevel(lvl zzzlog.Level) *zzzlog.Level {
	return &lvl
}
