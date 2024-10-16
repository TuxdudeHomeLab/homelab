package utils

import (
	"testing"

	"github.com/tuxdudehomelab/homelab/internal/testhelpers"
)

func TestMustParseDurationPanics(t *testing.T) {
	t.Parallel()

	tc := "MustParseDuration Panics On Error"
	input := "abc-garbage"
	want := `time: invalid duration "abc-garbage"`

	t.Run(tc, func(t *testing.T) {
		t.Parallel()

		defer testhelpers.ExpectPanic(t, "MustParseDuration()", tc, want)
		_ = MustParseDuration(input)
	})
}

func TestMustParseRAMInBytes(t *testing.T) {
	t.Parallel()

	tc := "MustParseRAMInBytes Panics On Error"
	input := "abc-garbage"
	want := `invalid size: 'abc-garbage'`

	t.Run(tc, func(t *testing.T) {
		t.Parallel()

		defer testhelpers.ExpectPanic(t, "MustParseRAMInBytes()", tc, want)
		_ = MustParseRAMInBytes(input)
	})
}
