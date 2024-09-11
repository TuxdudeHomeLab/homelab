package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/TwiN/deepmerge"
	"gopkg.in/yaml.v3"
)

// HomelabConfig represents the entire homelab deployment configuration.
type HomelabConfig struct {
	Global     GlobalConfig           `yaml:"global"`
	IPAM       IPAMConfig             `yaml:"ipam"`
	Hosts      []HostConfig           `yaml:"hosts"`
	Groups     []ContainerGroupConfig `yaml:"groups"`
	Containers []ContainerConfig      `yaml:"containers"`
}

// GlobalConfig represents the configuration that will be applied
// globally across all containers.
type GlobalConfig struct {
	Env                  []EnvConfig    `yaml:"env"`
	Volumes              []VolumeConfig `yaml:"volumes"`
	ContainerStopTimeout int            `yaml:"containerStopTimeout"`
	RestartPolicy        string         `yaml:"restartPolicy"`
	TimeZone             string         `yaml:"timeZone"`
	DomainName           string         `yaml:"domainName"`
	DNSSearch            string         `yaml:"dnsSearch"`
}

// EnvConfig is a pair of environment variable name and value that will be
// substituted in all string field values read from the homelab
// configuration file.
type EnvConfig struct {
	Var          string `yaml:"var"`
	Value        string `yaml:"value"`
	ValueCommand string `yaml:"valueCommand"`
}

// IPAMConfig represents the IP Addressing and management information for
// all containers in the homelab configuration.
type IPAMConfig struct {
	Networks NetworksConfig `yaml:"networks"`
}

// NetworksConfig represents all networks in the homelab configuration.
type NetworksConfig struct {
	BridgeModeNetworks    []BridgeModeNetworkConfig    `yaml:"bridgeModeNetworks"`
	ContainerModeNetworks []ContainerModeNetworkConfig `yaml:"containerModeNetworks"`
}

// BridgeModeNetworkConfig represents a docker bridge mode network that one
// or more containers attach to.
type BridgeModeNetworkConfig struct {
	Name              string              `yaml:"name"`
	HostInterfaceName string              `yaml:"hostInterfaceName"`
	Cidr              string              `yaml:"cidr"`
	Priority          int                 `yaml:"priority"`
	Containers        []ContainerIPConfig `yaml:"containers"`
}

// ContainerModeNetworkConfig represents a container network meant to attach a
// container to another container's network stack.
type ContainerModeNetworkConfig struct {
	Name       string               `yaml:"name"`
	Priority   int                  `yaml:"priority"`
	Containers []ContainerReference `yaml:"containers"`
}

// ContainerIP represents the IP information for a container.
type ContainerIPConfig struct {
	IP        string             `yaml:"ip"`
	Container ContainerReference `yaml:"container"`
}

// HostConfig represents the host specific information.
type HostConfig struct {
	Name              string               `yaml:"name"`
	AllowedContainers []ContainerReference `yaml:"allowedContainers"`
}

// ContainerReference identifies a specific container part of a group.
type ContainerReference struct {
	Group     string `yaml:"group"`
	Container string `yaml:"container"`
}

// ContainerGroupConfig represents a single logical container group, which is
// basically a collection of containers within.
type ContainerGroupConfig struct {
	Name  string `yaml:"name"`
	Order int    `yaml:"order"`
}

// ContainerConfig represents a single docker container.
type ContainerConfig struct {
	Name                    string                `yaml:"name"`
	ParentGroup             string                `yaml:"parentGroup"`
	Image                   string                `yaml:"image"`
	Order                   int                   `yaml:"order"`
	StopTimeout             int                   `yaml:"stopTimeout"`
	SkipImagePull           bool                  `yaml:"skipImagePull"`
	IgnoreImagePullFailures bool                  `yaml:"ignoreImagePullFailures"`
	PullImageAfterStop      bool                  `yaml:"pullImageAfterStop"`
	StartPreHook            string                `yaml:"startPreHook"`
	User                    string                `yaml:"user"`
	UserGroup               string                `yaml:"group"`
	GroupAdd                []string              `yaml:"groupAdd"`
	HostName                string                `yaml:"hostName"`
	DomainName              string                `yaml:"domainName"`
	DynamicFlagsCommand     string                `yaml:"dynamicFlagsCommand"`
	Devices                 []DeviceConfig        `yaml:"devices"`
	ShmSize                 string                `yaml:"shmSize"`
	CapAdd                  []string              `yaml:"capAdd"`
	Mounts                  []string              `yaml:"mounts"`
	Volumes                 []VolumeConfig        `yaml:"volumes"`
	Env                     []ContainerEnvConfig  `yaml:"env"`
	PublishedPorts          []PublishedPortConfig `yaml:"publishedPorts"`
	Labels                  []LabelConfig         `yaml:"labels"`
	StopSignal              string                `yaml:"stopSignal"`
	HealthCmd               string                `yaml:"healthCmd"`
	Entrypoint              string                `yaml:"entrypoint"`
	Args                    []string              `yaml:"args"`
}

// VolumeConfig represents a bind mounted volume.
type VolumeConfig struct {
	Name     string `yaml:"name"`
	Src      string `yaml:"src"`
	Dst      string `yaml:"dst"`
	ReadOnly bool   `yaml:"readOnly,omitempty"`
}

// DeviceConfig represents a device node that will be exposed to a container.
type DeviceConfig struct {
	Src string `yaml:"src"`
	Dst string `yaml:"dst"`
}

// ContainerEnvConfig represents an environment variable and value pair that will be set
// on the specified container.
type ContainerEnvConfig struct {
	Var          string `yaml:"var"`
	Value        string `yaml:"value,omitempty"`
	ValueCommand string `yaml:"valueCommand,omitempty"`
}

// PublishedPortConfig represents a port published from a container.
type PublishedPortConfig struct {
	Src string `yaml:"src"`
	Dst string `yaml:"dst"`
}

// LabelConfig represents a label set on a container.
type LabelConfig struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

func homelabConfigsPath() (string, error) {
	// Use the flag from the command line if present.
	if isFlagPassed(homelabConfigsDirFlag) {
		log.Debugf("Using Homelab configs path from command line flag: %s", *homelabConfigsDir)
		return *homelabConfigsDir, nil
	}
	path, err := configsPath()
	if err != nil {
		return "", err
	}

	log.Debugf("Using Homelab configs path from CLI config: %s", path)
	return path, nil
}

func mergedConfigReader(path string) (io.Reader, error) {
	var result []byte
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d == nil || d.IsDir() {
			return nil
		}
		ext := filepath.Ext(p)
		if ext != ".yml" && ext != ".yaml" {
			return nil
		}

		log.Debugf("Picked up homelab config: %s", p)
		configFile, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("failed to read homelab config file %q, reason: %w", p, err)
		}
		result, err = deepmerge.YAML(result, configFile)
		if err != nil {
			return fmt.Errorf("failed to deep merge config file %q, reason: %w", p, err)
		}
		return nil
	})
	log.DebugEmpty()

	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no homelab configs found in %s", path)
	}

	return bytes.NewReader(result), nil
}

func (h *HomelabConfig) parseUsingReader(s io.Reader) error {
	dec := yaml.NewDecoder(s)
	dec.KnownFields(true)
	err := dec.Decode(h)
	if err != nil {
		return fmt.Errorf("failed to parse homelab config, reason: %w", err)
	}

	log.Tracef("Homelab Config:\n%v\n", prettyPrintJSON(h))
	return nil
}

func (h *HomelabConfig) parse() error {
	path, err := homelabConfigsPath()
	if err != nil {
		return err
	}

	pathStat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("os.Stat() failed on homelab configs path, reason: %w", err)
	}
	if !pathStat.IsDir() {
		return fmt.Errorf("homelab configs path %q must be a directory", path)
	}

	m, err := mergedConfigReader(path)
	if err != nil {
		return err
	}

	return h.parseUsingReader(m)
}

func (h *HomelabConfig) validate() error {
	// TODO: Perform the following (and more) validations:
	// 1. Validate global config:
	//     a. No duplicate global config env names.
	//     b. Validate mandatory properties of every global config env.
	//     c. Every global config env specifies exactly one of value or
	//        valueCommand, but not both.
	//     d. Validate mandatory properties of every global config volume.
	//     e. No duplicate global config volume names.
	// 2. Validate hosts config:
	//     a. No duplicate host names.
	//     b. No duplicate allowed containers (i.e. combination of group
	//        and container name).
	// 3. Validate IPAM config:
	//     a. No duplicate network names across bridge and container mode
	//        networks.
	//     b. No duplicate host interface names across bridge networks.
	//     c. No overlapping CIDR across networks.
	//     d. No duplicate container names within a bridge or container
	//        mode network.
	//     e. All IPs in a bridge network belong to the CIDR.
	//     f. No duplicate IPs within a bridge network.
	// 4. Groups config:
	//     a. No duplicate group names.
	//     b. Order defined for all the groups.
	// 5. Container configs:
	//     a. Parent group name is a valid group defined under group config.
	//     b. No duplicate container names within the same group.
	//     c. Order defined for all the containers.
	//     d. Image defined for all the containers.
	//     e. Validate mandatory properties of every device config.
	//     f. Validate manadatory properties of every container config volume.
	//     g. Volume pure name references are valid global config volume references.
	//     h. Validate manadatory properties of every container config env.
	//     i. Every container config env specifies exactly one of value or
	//        valueCommand, but not both.
	//     j. Validate mandatory properties of every published port config.
	//     k. Validate mandatory properties of every label config.
	return nil
}
