package main

import (
	"context"
	"testing"
)

func TestRetrieveLoggerFromEmptyContext(t *testing.T) {
	tc := "Retrieve Logger - Empty Context"
	want := `Unable to retrieve logger from context`

	t.Run(tc, func(t *testing.T) {
		ctx := context.Background()

		defer testExpectPanic(t, "log()", tc, want)
		_ = log(ctx)
	})
}
