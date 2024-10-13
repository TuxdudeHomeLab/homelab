package cmdexec

import "context"

var (
	executorKey = ctxKeyExecutor{}
)

type ctxKeyExecutor struct{}

func ExecutorFromContext(ctx context.Context) (Executor, bool) {
	client, ok := ctx.Value(executorKey).(Executor)
	return client, ok
}

func MustExecutor(ctx context.Context) Executor {
	if d, found := ExecutorFromContext(ctx); found {
		return d
	}
	log(ctx).Fatalf("Executor not found in context")
	return nil
}

func WithExecutor(ctx context.Context, exec Executor) context.Context {
	return context.WithValue(ctx, executorKey, exec)
}
