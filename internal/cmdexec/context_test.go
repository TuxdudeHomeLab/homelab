package cmdexec

import (
	"bytes"
	"context"
	"testing"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdude/zzzlogi"
	l "github.com/tuxdudehomelab/homelab/internal/log"
	"github.com/tuxdudehomelab/homelab/internal/testhelpers"
)

func TestRetrieveExecutorFromEmptyContext(t *testing.T) {
	t.Parallel()

	tc := "Retrieve Executor - Empty Context"
	want := `Executor not found in context`

	t.Run(tc, func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctx = l.WithLogger(ctx, newTestLogger())

		defer testhelpers.ExpectPanic(t, "MustExecutor()", tc, want)
		_ = MustExecutor(ctx)
	})
}

func newTestLogger() zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	config.SkipCallerInfo = true
	config.PanicInFatal = true
	config.Dest = new(bytes.Buffer)
	return zzzlog.NewLogger(config)
}
