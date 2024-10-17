package deployment

import (
	"testing"

	"github.com/tuxdudehomelab/homelab/internal/config"
	"github.com/tuxdudehomelab/homelab/internal/testhelpers"
)

func TestNetworkCreateOptionsPanics(t *testing.T) {
	t.Parallel()

	tc := "network CreateOptions With Container Mode Network Panics"
	want := `Only bridge mode network creation is possible`

	t.Run(tc, func(t *testing.T) {
		t.Parallel()

		net := newContainerModeNetwork("net-foo", &containerModeNetworkInfo{
			container: config.ContainerReference{
				Group:     "some-group",
				Container: "some-ct",
			},
		})

		defer testhelpers.ExpectPanic(t, "network.createOptions()", tc, want)
		_ = net.createOptions()
	})
}

func TestNetworkStringerPanics(t *testing.T) {
	t.Parallel()

	tc := "network Stringer Invalid Mode Panics"
	want := `unknown network mode, possibly indicating a bug in the code!`

	t.Run(tc, func(t *testing.T) {
		t.Parallel()

		net := &Network{
			mode: NetworkModeUnknown,
		}

		defer testhelpers.ExpectPanic(t, "network.String()", tc, want)
		_ = net.String()
	})
}
