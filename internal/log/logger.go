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
	logger, ok := ctx.Value(loggerKey).(zzzlogi.Logger)
	if !ok {
		panic("Unable to retrieve logger from context")
	}
	return logger
}

func WithLogger(ctx context.Context, logger zzzlogi.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
