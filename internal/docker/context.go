package docker

import (
	"context"
	"time"
)

var (
	dockerAPIClientKey                 = ctxKeyAPIClient{}
	containerStopAndRemoveKillDelayKey = ctxKeyContainerStopAndRemoveKillDelay{}
)

type ctxKeyAPIClient struct{}
type ctxKeyContainerStopAndRemoveKillDelay struct{}

func APIClientFromContext(ctx context.Context) (APIClient, bool) {
	client, ok := ctx.Value(dockerAPIClientKey).(APIClient)
	return client, ok
}

func MustAPIClient(ctx context.Context) APIClient {
	if d, found := APIClientFromContext(ctx); found {
		return d
	}
	log(ctx).Fatalf("Docker API Client not found in context")
	return nil
}

func WithAPIClient(ctx context.Context, client APIClient) context.Context {
	return context.WithValue(ctx, dockerAPIClientKey, client)
}

func getContainerStopAndRemoveKillDelay(ctx context.Context) (time.Duration, bool) {
	delay, ok := ctx.Value(containerStopAndRemoveKillDelayKey).(time.Duration)
	return delay, ok
}

func WithContainerStopAndRemoveKillDelay(ctx context.Context, delay time.Duration) context.Context {
	return context.WithValue(ctx, containerStopAndRemoveKillDelayKey, delay)
}
