package docker

import (
	"testing"

	"github.com/tuxdudehomelab/homelab/internal/testhelpers"
)

var mustRestartPolicyModeFromStringPanicTests = []struct {
	name   string
	policy string
}{
	{
		name:   "MustRestartPolicyModeFromString() - Panic - Garbage Policy",
		policy: "garbage",
	},
}

func TestMustRestartPolicyModeFromStringPanics(t *testing.T) {
	t.Parallel()

	want := `unable to convert restart policy mode garbage setting, reason: invalid restart policy mode string: garbage, possibly indicating a bug in the code`
	for _, tc := range mustRestartPolicyModeFromStringPanicTests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defer testhelpers.ExpectPanic(t, "MustRestartPolicyModeFromString()", tc.name, want)
			_ = MustRestartPolicyModeFromString(tc.policy)
		})
	}
}
