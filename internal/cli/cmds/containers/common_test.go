package containers

import (
	"testing"

	"github.com/tuxdudehomelab/homelab/internal/testhelpers"
)

var mustContainerNamePanicTests = []struct {
	name          string
	containerName string
}{
	{
		name:          "MustContainerName() - Panic - Empty Container Name",
		containerName: "",
	},
	{
		name:          "MustContainerName() - Panic - Invalid Container Name",
		containerName: "foo",
	},
}

func TestMustContainerNamePanics(t *testing.T) {
	t.Parallel()

	want := `Container name must be specified in the form 'group/container'`
	for _, tc := range mustContainerNamePanicTests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defer testhelpers.ExpectPanic(t, "mustContainerName()", tc.name, want)
			_, _ = mustContainerName(tc.containerName)
		})
	}
}
