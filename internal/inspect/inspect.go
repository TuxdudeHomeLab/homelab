package inspect

import "context"

var (
	homelabInspectLevelKey = ctxKeyHomelabInspectLevel{}
)

type ctxKeyHomelabInspectLevel struct{}

const (
	HomelabInspectLevelNone = iota
	HomelabInspectLevelDebug
	HomelabInspectLevelTrace
)

type HomelabInspectLevel uint8

func HomelabInspectLevelFromContext(ctx context.Context) HomelabInspectLevel {
	lvl, ok := ctx.Value(homelabInspectLevelKey).(HomelabInspectLevel)
	if !ok {
		return HomelabInspectLevelNone
	}
	return lvl
}

func WithHomelabInspectLevel(ctx context.Context, lvl HomelabInspectLevel) context.Context {
	return context.WithValue(ctx, homelabInspectLevelKey, lvl)
}
