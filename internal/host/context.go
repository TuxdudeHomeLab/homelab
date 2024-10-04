package host

import "context"

var (
	hostInfoKey = ctxKeyHostInfo{}
)

type ctxKeyHostInfo struct{}

func HostInfoFromContext(ctx context.Context) (*HostInfo, bool) {
	host, ok := ctx.Value(hostInfoKey).(*HostInfo)
	return host, ok
}

func MustHostInfo(ctx context.Context) *HostInfo {
	if h, found := HostInfoFromContext(ctx); found {
		return h
	}
	log(ctx).Fatalf("Host info not found in context")
	return nil
}

func WithHostInfo(ctx context.Context, host *HostInfo) context.Context {
	return context.WithValue(ctx, hostInfoKey, host)
}
