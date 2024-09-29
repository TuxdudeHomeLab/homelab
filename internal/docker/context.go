package docker

import "context"

var (
	dockerAPIClientKey = ctxKeyDockerAPIClient{}
)

type ctxKeyDockerAPIClient struct{}

func DockerAPIClientFromContext(ctx context.Context) (DockerAPIClient, bool) {
	client, ok := ctx.Value(dockerAPIClientKey).(DockerAPIClient)
	return client, ok
}

func WithDockerAPIClient(ctx context.Context, client DockerAPIClient) context.Context {
	return context.WithValue(ctx, dockerAPIClientKey, client)
}
