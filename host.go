package main

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"os"
	"runtime"
	"strings"
)

type hostInfo struct {
	hostName              string
	humanFriendlyHostName string
	ip                    netip.Addr
	numCPUs               int
	os                    string
	arch                  string
	dockerPlatform        string
}

type stringSet map[string]bool

const (
	osLinux   = "linux"
	archAmd64 = "amd64"
	archArm64 = "arm64"
)

func newHostInfo(ctx context.Context) *hostInfo {
	res := hostInfo{
		humanFriendlyHostName: systemHostName(ctx),
		ip:                    interfaceIP(ctx),
		numCPUs:               runtime.NumCPU(),
		os:                    runtime.GOOS,
		arch:                  runtime.GOARCH,
		dockerPlatform:        archToDockerPlatform(runtime.GOARCH),
	}
	res.hostName = strings.ToLower(res.humanFriendlyHostName)

	log(ctx).Debugf("Host name: %s", res.hostName)
	log(ctx).Debugf("Human Friendly Host name: %s", res.humanFriendlyHostName)
	log(ctx).Debugf("Host IP: %s", res.ip)
	log(ctx).Debugf("Num CPUs = %d", res.numCPUs)
	log(ctx).Debugf("OS = %s", res.os)
	log(ctx).Debugf("Arch = %s", res.arch)
	log(ctx).Debugf("Docker Platform = %s", res.dockerPlatform)
	log(ctx).DebugEmpty()

	if res.os != osLinux {
		log(ctx).Fatalf("Only linux OS is supported, found OS: %s", res.os)
	}
	if res.arch != archAmd64 && res.arch != archArm64 {
		log(ctx).Fatalf("Only amd64 and arm64 platforms are supported, found Arch: %s", res.arch)
	}

	return &res
}

func systemHostName(ctx context.Context) string {
	res, err := os.Hostname()
	if err != nil {
		log(ctx).Fatalf("Unable to determine the current machine's host name, %v", err)
	}
	return res
}

func interfaceIP(ctx context.Context) netip.Addr {
	conn, err := net.Dial("udp", "10.1.1.1:1234")
	if err != nil {
		log(ctx).Fatalf("Unable to determine the current machine's IP, %v", err)
	}
	defer conn.Close()

	ip, ok := netip.AddrFromSlice(conn.LocalAddr().(*net.UDPAddr).IP)
	if !ok {
		log(ctx).Fatalf("Unable to convert host IP %s of net.IP type to netip.Addr type")
	}
	return ip
}

func archToDockerPlatform(arch string) string {
	switch arch {
	case archAmd64:
		return "linux/amd64"
	case archArm64:
		return "linux/arm64/v8"
	default:
		return fmt.Sprintf("unsupported-docker-arch-%s", arch)
	}
}
