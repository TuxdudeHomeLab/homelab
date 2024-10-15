package docker

import (
	"context"
)

var (
	dockerAPIClientKey            = ctxKeyAPIClient{}
	containerPurgeKillAttemptsKey = ctxKeyContainerPurgeKillAttempts{}
)

type ctxKeyAPIClient struct{}
type ctxKeyContainerPurgeKillAttempts struct{}

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

func getContainerPurgeKillAttempts(ctx context.Context) (uint32, bool) {
	delay, ok := ctx.Value(containerPurgeKillAttemptsKey).(uint32)
	return delay, ok
}

func WithContainerPurgeKillAttempts(ctx context.Context, attempts uint32) context.Context {
	return context.WithValue(ctx, containerPurgeKillAttemptsKey, attempts)
}
