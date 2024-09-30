package docker

import (
	"context"
	"time"
)

var (
	dockerAPIClientKey                 = ctxKeyDockerAPIClient{}
	containerStopAndRemoveKillDelayKey = ctxKeyContainerStopAndRemoveKillDelay{}
)

type ctxKeyDockerAPIClient struct{}
type ctxKeyContainerStopAndRemoveKillDelay struct{}

func DockerAPIClientFromContext(ctx context.Context) (DockerAPIClient, bool) {
	client, ok := ctx.Value(dockerAPIClientKey).(DockerAPIClient)
	return client, ok
}

func WithDockerAPIClient(ctx context.Context, client DockerAPIClient) context.Context {
	return context.WithValue(ctx, dockerAPIClientKey, client)
}

func getContainerStopAndRemoveKillDelay(ctx context.Context) (time.Duration, bool) {
	delay, ok := ctx.Value(containerStopAndRemoveKillDelayKey).(time.Duration)
	return delay, ok
}

func WithContainerStopAndRemoveKillDelay(ctx context.Context, delay time.Duration) context.Context {
	return context.WithValue(ctx, containerStopAndRemoveKillDelayKey, delay)
}
