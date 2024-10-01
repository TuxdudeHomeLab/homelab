package log

import (
	"context"

	"github.com/tuxdude/zzzlogi"
)

var (
	loggerKey = ctxKeyLogger{}
)

type ctxKeyLogger struct{}

func Log(ctx context.Context) zzzlogi.Logger {
	if logger, found := ctx.Value(loggerKey).(zzzlogi.Logger); found {
		return logger
	}
	panic("Unable to retrieve logger from context")
}

func WithLogger(ctx context.Context, logger zzzlogi.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
