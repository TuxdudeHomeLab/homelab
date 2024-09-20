package main

import (
	"net/netip"
	"os"
)

const (
	fakeHostName              = "fakehost"
	fakeHumanFriendlyHostName = "FakeHost"
	fakeHostIP                = "10.76.77.78"
	fakeHostNumCPUs           = 32
	fakeHostOS                = "linux"
	fakeHostArch              = "amd64"
	fakeHostDockerPlatform    = "linux/amd64"
)

var (
	fakeHostInfo = &hostInfo{
		hostName:              fakeHostName,
		humanFriendlyHostName: fakeHumanFriendlyHostName,
		ip:                    netip.MustParseAddr(fakeHostIP),
		numCPUs:               fakeHostNumCPUs,
		os:                    fakeHostOS,
		arch:                  fakeHostArch,
		dockerPlatform:        fakeHostDockerPlatform,
	}
	fakeConfigEnv = newConfigEnv(fakeHostInfo)
)

func pwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}

// nolint: unused
func newInt(i int) *int {
	return &i
}
