package version

import "context"

var (
	versionInfoKey = ctxKeyVersionInfo{}
)

type ctxKeyVersionInfo struct{}

func VersionInfoFromContext(ctx context.Context) (*VersionInfo, bool) {
	ver, ok := ctx.Value(versionInfoKey).(*VersionInfo)
	return ver, ok
}

func WithVersionInfo(ctx context.Context, ver *VersionInfo) context.Context {
	return context.WithValue(ctx, versionInfoKey, ver)
}
