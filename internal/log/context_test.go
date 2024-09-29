package log

import (
	"context"
	"testing"

	"github.com/tuxdudehomelab/homelab/internal/testhelpers"
)

func TestRetrieveLoggerFromEmptyContext(t *testing.T) {
	tc := "Retrieve Logger - Empty Context"
	want := `Unable to retrieve logger from context`

	t.Run(tc, func(t *testing.T) {
		ctx := context.Background()

		defer testhelpers.ExpectPanic(t, "log()", tc, want)
		_ = Log(ctx)
	})
}
