package main

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
)

type hostInfo struct {
	hostName              string
	humanFriendlyHostName string
	ip                    net.IP
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

func newHostInfo() *hostInfo {
	res := hostInfo{
		humanFriendlyHostName: systemHostName(),
		ip:                    interfaceIP(),
		numCPUs:               runtime.NumCPU(),
		os:                    runtime.GOOS,
		arch:                  runtime.GOARCH,
		dockerPlatform:        archToDockerPlatform(runtime.GOARCH),
	}
	res.hostName = strings.ToLower(res.humanFriendlyHostName)

	log.Debugf("Host name: %s", res.hostName)
	log.Debugf("Human Friendly Host name: %s", res.humanFriendlyHostName)
	log.Debugf("Host IP: %s", res.ip)
	log.Debugf("Num CPUs = %d", res.numCPUs)
	log.Debugf("OS = %s", res.os)
	log.Debugf("Arch = %s", res.arch)
	log.Debugf("Docker Platform = %s", res.dockerPlatform)
	log.DebugEmpty()

	if res.os != osLinux {
		log.Fatalf("Only linux OS is supported, found OS: %s", res.os)
	}
	if res.arch != archAmd64 && res.arch != archArm64 {
		log.Fatalf("Only amd64 and arm64 platforms are supported, found Arch: %s", res.arch)
	}

	return &res
}

func systemHostName() string {
	res, err := os.Hostname()
	if err != nil {
		log.Fatalf("Unable to determine the current machine's host name, %v", err)
	}
	return res
}

func interfaceIP() net.IP {
	conn, err := net.Dial("udp", "10.1.1.1:1234")
	if err != nil {
		log.Fatalf("Unable to determine the current machine's IP, %v", err)
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP
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
