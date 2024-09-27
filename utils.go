package main

import (
	"context"
	"fmt"

	"github.com/tuxdude/zzzlogi"

	"github.com/clarketm/json"
)

var (
	homelabInspectLevelKey = ctxKeyHomelabInspectLevel{}
	loggerKey              = ctxKeyLogger{}
	dockerAPIClientKey     = ctxKeyDockerAPIClient{}
	hostInfoKey            = ctxKeyHostInfo{}
)

const (
	homelabInspectLevelNone = iota
	homelabInspectLevelDebug
	homelabInspectLevelTrace
)

type homelabInspectLevel uint8

type stringSet map[string]struct{}

type ctxKeyHomelabInspectLevel struct{}
type ctxKeyLogger struct{}
type ctxKeyDockerAPIClient struct{}
type ctxKeyHostInfo struct{}

func homelabInspectLevelFromContext(ctx context.Context) homelabInspectLevel {
	lvl, ok := ctx.Value(homelabInspectLevelKey).(homelabInspectLevel)
	if !ok {
		return homelabInspectLevelNone
	}
	return lvl
}

func log(ctx context.Context) zzzlogi.Logger {
	logger, ok := ctx.Value(loggerKey).(zzzlogi.Logger)
	if !ok {
		panic("Unable to retriever logger from context")
	}
	return logger
}

func dockerAPIClientFromContext(ctx context.Context) (dockerAPIClient, bool) {
	client, ok := ctx.Value(dockerAPIClientKey).(dockerAPIClient)
	return client, ok
}

func hostInfoFromContext(ctx context.Context) (*hostInfo, bool) {
	host, ok := ctx.Value(hostInfoKey).(*hostInfo)
	return host, ok
}

func withHomelabInspectLevel(ctx context.Context, lvl homelabInspectLevel) context.Context {
	return context.WithValue(ctx, homelabInspectLevelKey, lvl)
}

func withLogger(ctx context.Context, logger zzzlogi.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// This is used purely by tests.
// nolint:unused
func withDockerAPIClient(ctx context.Context, client dockerAPIClient) context.Context {
	return context.WithValue(ctx, dockerAPIClientKey, client)
}

// This is used purely by tests.
// nolint:unused
func withHostInfo(ctx context.Context, host *hostInfo) context.Context {
	return context.WithValue(ctx, hostInfoKey, host)
}

// Returns the JSON formatted string representation of the specified object.
func prettyPrintJSON(x interface{}) string {
	p, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		return fmt.Sprintf("%#v", x)
	}
	return string(p)
}

func logToErrorAndReturn(ctx context.Context, format string, args ...interface{}) error {
	log(ctx).Errorf(format, args...)
	log(ctx).ErrorEmpty()
	return fmt.Errorf(format, args...)
}

func newBool(b bool) *bool {
	return &b
}
