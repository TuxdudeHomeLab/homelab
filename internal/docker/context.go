package docker

import (
	"context"
	"time"
)

var (
	dockerAPIClientKey         = ctxKeyAPIClient{}
	containerPurgeKillDelayKey = ctxKeyContainerPurgeKillDelay{}
)

type ctxKeyAPIClient struct{}
type ctxKeyContainerPurgeKillDelay struct{}

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

func getContainerPurgeKillDelay(ctx context.Context) (time.Duration, bool) {
	delay, ok := ctx.Value(containerPurgeKillDelayKey).(time.Duration)
	return delay, ok
}

func WithContainerPurgeKillDelay(ctx context.Context, delay time.Duration) context.Context {
	return context.WithValue(ctx, containerPurgeKillDelayKey, delay)
}
