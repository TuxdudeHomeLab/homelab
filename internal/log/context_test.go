package log

import (
	"context"
	"testing"

	"github.com/tuxdudehomelab/homelab/internal/testhelpers"
)

func TestRetrieveLoggerFromEmptyContext(t *testing.T) {
	t.Parallel()

	tc := "Retrieve Logger - Empty Context"
	want := `Unable to retrieve logger from context`

	t.Run(tc, func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		defer testhelpers.ExpectPanic(t, "log()", tc, want)
		_ = Log(ctx)
	})
}
