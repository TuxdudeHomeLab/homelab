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

func WithHostInfo(ctx context.Context, host *HostInfo) context.Context {
	return context.WithValue(ctx, hostInfoKey, host)
}
