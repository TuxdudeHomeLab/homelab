package fakehost

import (
	"net/netip"

	"github.com/tuxdudehomelab/homelab/internal/host"
)

const (
	FakeHostName              = "fakehost"
	FakeHumanFriendlyHostName = "FakeHost"
	FakeHostIP                = "10.76.77.78"
	FakeHostNumCPUs           = 32
	FakeHostOS                = "linux"
	FakeHostArch              = "amd64"
	FakeHostDockerPlatform    = "linux/amd64"
)

func NewFakeHostInfo() *host.HostInfo {
	return &host.HostInfo{
		HostName:              FakeHostName,
		HumanFriendlyHostName: FakeHumanFriendlyHostName,
		IP:                    netip.MustParseAddr(FakeHostIP),
		NumCPUs:               FakeHostNumCPUs,
		OS:                    FakeHostOS,
		Arch:                  FakeHostArch,
		DockerPlatform:        FakeHostDockerPlatform,
	}
}
