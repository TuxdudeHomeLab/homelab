package testutils

import (
	"io"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdude/zzzlogi"
)

func NewTestLogger() zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	config.SkipCallerInfo = true
	config.PanicInFatal = true
	return zzzlog.NewLogger(config)
}

func NewCapturingTestLogger(lvl zzzlog.Level, w io.Writer) zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	config.MaxLevel = lvl
	config.SkipCallerInfo = true
	config.PanicInFatal = true
	config.Dest = w
	return zzzlog.NewLogger(config)
}

func NewCapturingVanillaTestLogger(lvl zzzlog.Level, w io.Writer) zzzlogi.Logger {
	config := zzzlog.NewVanillaLoggerConfig()
	config.MaxLevel = lvl
	config.Dest = w
	config.PanicInFatal = true
	return zzzlog.NewLogger(config)
}
