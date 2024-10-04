package user

import (
	"bytes"
	"context"
	"testing"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdude/zzzlogi"
	l "github.com/tuxdudehomelab/homelab/internal/log"
	"github.com/tuxdudehomelab/homelab/internal/testhelpers"
)

func TestRetrieveUserInfoFromEmptyContext(t *testing.T) {
	tc := "Retrieve User Info - Empty Context"
	want := `User info not found in context`

	t.Run(tc, func(t *testing.T) {
		ctx := context.Background()
		ctx = l.WithLogger(ctx, newTestLogger())

		defer testhelpers.ExpectPanic(t, "MustUserInfo()", tc, want)
		_ = MustUserInfo(ctx)
	})
}

func newTestLogger() zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	config.SkipCallerInfo = true
	config.PanicInFatal = true
	config.Dest = new(bytes.Buffer)
	return zzzlog.NewLogger(config)
}
