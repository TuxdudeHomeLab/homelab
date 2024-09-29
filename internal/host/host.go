package host

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"os"
	"runtime"
	"strings"
)

type HostInfo struct {
	HostName              string
	HumanFriendlyHostName string
	IP                    netip.Addr
	NumCPUs               int
	OS                    string
	Arch                  string
	DockerPlatform        string
}

const (
	osLinux   = "linux"
	archAmd64 = "amd64"
	archArm64 = "arm64"
)

func NewHostInfo(ctx context.Context) *HostInfo {
	res := HostInfo{
		HumanFriendlyHostName: systemHostName(ctx),
		IP:                    interfaceIP(ctx),
		NumCPUs:               runtime.NumCPU(),
		OS:                    runtime.GOOS,
		Arch:                  runtime.GOARCH,
		DockerPlatform:        archToDockerPlatform(runtime.GOARCH),
	}
	res.HostName = strings.ToLower(res.HumanFriendlyHostName)

	log(ctx).Debugf("Host name: %s", res.HostName)
	log(ctx).Debugf("Human Friendly Host name: %s", res.HumanFriendlyHostName)
	log(ctx).Debugf("Host IP: %s", res.IP)
	log(ctx).Debugf("Num CPUs = %d", res.NumCPUs)
	log(ctx).Debugf("OS = %s", res.OS)
	log(ctx).Debugf("Arch = %s", res.Arch)
	log(ctx).Debugf("Docker Platform = %s", res.DockerPlatform)
	log(ctx).DebugEmpty()

	if res.OS != osLinux {
		log(ctx).Fatalf("Only linux OS is supported, found OS: %s", res.OS)
	}
	if res.Arch != archAmd64 && res.Arch != archArm64 {
		log(ctx).Fatalf("Only amd64 and arm64 platforms are supported, found Arch: %s", res.Arch)
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
