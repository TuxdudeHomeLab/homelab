package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/netip"
	"os"
	"path/filepath"

	"github.com/TwiN/deepmerge"
	"gopkg.in/yaml.v3"
)

// HomelabConfig represents the entire homelab deployment configuration.
type HomelabConfig struct {
	Global     GlobalConfig           `yaml:"global,omitempty"`
	IPAM       IPAMConfig             `yaml:"ipam,omitempty"`
	Hosts      []HostConfig           `yaml:"hosts,omitempty"`
	Groups     []ContainerGroupConfig `yaml:"groups,omitempty"`
	Containers []ContainerConfig      `yaml:"containers,omitempty"`
}

// GlobalConfig represents the configuration that will be applied
// across the entire homelab deployment.
type GlobalConfig struct {
	Env       []ConfigEnv           `yaml:"env,omitempty"`
	MountDefs []MountConfig         `yaml:"mountDefs,omitempty"`
	Container GlobalContainerConfig `yaml:"container,omitempty"`
}

// GlobalContainerConfig represents container related configuration that
// will be applied globally across all containers.
type GlobalContainerConfig struct {
	StopSignal    string                       `yaml:"stopSignal,omitempty"`
	StopTimeout   int                          `yaml:"stopTimeout,omitempty"`
	RestartPolicy ContainerRestartPolicyConfig `yaml:"restartPolicy,omitempty"`
	DomainName    string                       `yaml:"domainName,omitempty"`
	DNSSearch     []string                     `yaml:"dnsSearch,omitempty"`
	Env           []ContainerEnv               `yaml:"env,omitempty"`
	Mounts        []MountConfig                `yaml:"mounts,omitempty"`
	Labels        []LabelConfig                `yaml:"labels,omitempty"`
}

// ConfigEnv is a pair of environment variable name and value that will be
// substituted in all string field values read from the homelab
// configuration file.
type ConfigEnv struct {
	Var          string `yaml:"var,omitempty"`
	Value        string `yaml:"value,omitempty"`
	ValueCommand string `yaml:"valueCommand,omitempty"`
}

// IPAMConfig represents the IP Addressing and management information for
// all containers in the homelab configuration.
type IPAMConfig struct {
	Networks NetworksConfig `yaml:"networks,omitempty"`
}

// NetworksConfig represents all networks in the homelab configuration.
type NetworksConfig struct {
	BridgeModeNetworks    []BridgeModeNetworkConfig    `yaml:"bridgeModeNetworks,omitempty"`
	ContainerModeNetworks []ContainerModeNetworkConfig `yaml:"containerModeNetworks,omitempty"`
}

// BridgeModeNetworkConfig represents a docker bridge mode network that one
// or more containers attach to.
type BridgeModeNetworkConfig struct {
	Name              string              `yaml:"name,omitempty"`
	HostInterfaceName string              `yaml:"hostInterfaceName,omitempty"`
	CIDR              string              `yaml:"cidr,omitempty"`
	Priority          int                 `yaml:"priority,omitempty"`
	Containers        []ContainerIPConfig `yaml:"containers,omitempty"`
}

// ContainerModeNetworkConfig represents a container network meant to attach a
// container to another container's network stack.
type ContainerModeNetworkConfig struct {
	Name       string               `yaml:"name,omitempty"`
	Priority   int                  `yaml:"priority,omitempty"`
	Containers []ContainerReference `yaml:"containers,omitempty"`
}

// ContainerIP represents the IP information for a container.
type ContainerIPConfig struct {
	IP        string             `yaml:"ip,omitempty"`
	Container ContainerReference `yaml:"container,omitempty"`
}

// HostConfig represents the host specific information.
type HostConfig struct {
	Name              string               `yaml:"name,omitempty"`
	AllowedContainers []ContainerReference `yaml:"allowedContainers,omitempty"`
}

// ContainerReference identifies a specific container part of a group.
type ContainerReference struct {
	Group     string `yaml:"group,omitempty"`
	Container string `yaml:"container,omitempty"`
}

// ContainerGroupConfig represents a single logical container group, which is
// basically a collection of containers within.
type ContainerGroupConfig struct {
	Name  string `yaml:"name,omitempty"`
	Order int    `yaml:"order,omitempty"`
}

// ContainerConfig represents a single docker container.
type ContainerConfig struct {
	Info       ContainerReference        `yaml:"info,omitempty"`
	Config     ContainerConfigOptions    `yaml:"config,omitempty"`
	Image      ContainerImageConfig      `yaml:"image,omitempty"`
	Metadata   ContainerMetadataConfig   `yaml:"metadata,omitempty"`
	Lifecycle  ContainerLifecycleConfig  `yaml:"lifecycle,omitempty"`
	User       ContainerUserConfig       `yaml:"user,omitempty"`
	Filesystem ContainerFilesystemConfig `yaml:"fs,omitempty"`
	Network    ContainerNetworkConfig    `yaml:"network,omitempty"`
	Security   ContainerSecurityConfig   `yaml:"security,omitempty"`
	Health     ContainerHealthConfig     `yaml:"health,omitempty"`
	Runtime    ContainerRuntimeConfig    `yaml:"runtime,omitempty"`
}

// ContainerImageConfig respresents the image configuration for the docker
// container.
type ContainerImageConfig struct {
	Image                   string `yaml:"image,omitempty"`
	SkipImagePull           bool   `yaml:"skipImagePull,omitempty"`
	IgnoreImagePullFailures bool   `yaml:"ignoreImagePullFailures,omitempty"`
	PullImageBeforeStop     bool   `yaml:"pullImageBeforeStop,omitempty"`
}

// ContainerConfigOptions represents options that are applied while
// evaluating the config for this docker container.
type ContainerConfigOptions struct {
	Env []ConfigEnv `yaml:"env,omitempty"`
}

// ContainerMetadataConfig represents the metadata for the docker container.
type ContainerMetadataConfig struct {
	Labels []LabelConfig `yaml:"labels,omitempty"`
}

// ContainerLifecycleConfig represents the lifecycle information for the
// docker container.
type ContainerLifecycleConfig struct {
	Order         int                          `yaml:"order,omitempty"`
	StartPreHook  string                       `yaml:"startPreHook,omitempty"`
	RestartPolicy ContainerRestartPolicyConfig `yaml:"restartPolicy,omitempty"`
	AutoRemove    bool                         `yaml:"autoRemove,omitempty"`
	StopSignal    string                       `yaml:"stopSignal,omitempty"`
	StopTimeout   int                          `yaml:"stopTimeout,omitempty"`
}

// ContainerUserConfig represents the user and group information for the
// docker container.
type ContainerUserConfig struct {
	User             string   `yaml:"user,omitempty"`
	PrimaryGroup     string   `yaml:"primaryGroup,omitempty"`
	AdditionalGroups []string `yaml:"additionalGroups,omitempty"`
}

// ContainerFilesystemConfig represents the fileystem information for the
// docker container.
type ContainerFilesystemConfig struct {
	ReadOnlyRootfs bool           `yaml:"readOnlyRootfs,omitempty"`
	Mounts         []MountConfig  `yaml:"mounts,omitempty"`
	Devices        []DeviceConfig `yaml:"devices,omitempty"`
}

// ContainerNetworkConfig represents the networking information for the
// docker container.
type ContainerNetworkConfig struct {
	HostName       string                `yaml:"hostName,omitempty"`
	DomainName     string                `yaml:"domainName,omitempty"`
	DNSServers     []string              `yaml:"dnsServers,omitempty"`
	DNSOptions     []string              `yaml:"dnsOptions,omitempty"`
	DNSSearch      []string              `yaml:"dnsSearch,omitempty"`
	PublishedPorts []PublishedPortConfig `yaml:"publishedPorts,omitempty"`
}

// ContainerSecurityConfig represents the security information for the
// docker container.
type ContainerSecurityConfig struct {
	Privileged bool           `yaml:"privileged,omitempty"`
	Sysctls    []SysctlConfig `yaml:"sysctls,omitempty"`
	CapAdd     []string       `yaml:"capAdd,omitempty"`
	CapDrop    []string       `yaml:"capDrop,omitempty"`
}

// ContainerHealthConfig represents the health check options for the
// docker container.
type ContainerHealthConfig struct {
	Cmd           string `yaml:"cmd,omitempty"`
	Interval      string `yaml:"interval,omitempty"`
	Retries       int    `yaml:"retries,omitempty"`
	StartInterval string `yaml:"startInterval,omitempty"`
	StartPeriod   string `yaml:"startPeriod,omitempty"`
	Timeout       string `yaml:"timeout,omitempty"`
}

// ContainerRuntimeConfig represents the execution and runtime information
// for the docker container.
type ContainerRuntimeConfig struct {
	AttachToTty bool           `yaml:"tty,omitempty"`
	ShmSize     string         `yaml:"shmSize,omitempty"`
	Env         []ContainerEnv `yaml:"env,omitempty"`
	Entrypoint  []string       `yaml:"entrypoint,omitempty"`
	Args        []string       `yaml:"args,omitempty"`
}

// MountConfig represents a filesystem mount.
type MountConfig struct {
	Name     string `yaml:"name,omitempty"`
	Type     string `yaml:"type,omitempty"`
	Src      string `yaml:"src,omitempty"`
	Dst      string `yaml:"dst,omitempty"`
	ReadOnly bool   `yaml:"readOnly,omitempty"`
	Options  string `yaml:"options,omitempty"`
}

// DeviceConfig represents a device node that will be exposed to a container.
type DeviceConfig struct {
	Src string `yaml:"src,omitempty"`
	Dst string `yaml:"dst,omitempty"`
}

// SysctlConfig represents a sysctl config to apply to a container.
type SysctlConfig struct {
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`
}

// ContainerEnv represents an environment variable and value pair that will be set
// on the specified container.
type ContainerEnv struct {
	Var          string `yaml:"var,omitempty"`
	Value        string `yaml:"value,omitempty"`
	ValueCommand string `yaml:"valueCommand,omitempty"`
}

// PublishedPortConfig represents a port published from a container.
type PublishedPortConfig struct {
	ContainerPort int    `yaml:"containerPort,omitempty"`
	Proto         string `yaml:"proto,omitempty"`
	HostIP        string `yaml:"hostIp,omitempty"`
	HostPort      int    `yaml:"hostPort,omitempty"`
}

// LabelConfig represents a label set on a container.
type LabelConfig struct {
	Name  string `yaml:"name,omitempty"`
	Value string `yaml:"value,omitempty"`
}

type ContainerRestartPolicyConfig struct {
	Mode          string `yaml:"mode,omitempty"`
	MaxRetryCount int    `yaml:"maxRetryCount,omitempty"`
}

func mergedConfigReader(path string) (io.Reader, error) {
	var result []byte
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to read contents of directory %s, reason: %w", path, err)
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
			return fmt.Errorf("failed to read homelab config file %s, reason: %w", p, err)
		}
		result, err = deepmerge.YAML(result, configFile)
		if err != nil {
			return fmt.Errorf("failed to deep merge config file %s, reason: %w", p, err)
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

func (h *HomelabConfig) parse(r io.Reader) error {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	err := dec.Decode(h)
	if err != nil {
		return fmt.Errorf("failed to parse homelab config, reason: %w", err)
	}

	log.Tracef("Homelab Config:\n%v\n", prettyPrintJSON(h))
	return nil
}

func (h *HomelabConfig) parseConfigs(configsPath string) error {
	pathStat, err := os.Stat(configsPath)
	if err != nil {
		return fmt.Errorf("os.Stat() failed on homelab configs path, reason: %w", err)
	}
	if !pathStat.IsDir() {
		return fmt.Errorf("homelab configs path %s must be a directory", configsPath)
	}

	m, err := mergedConfigReader(configsPath)
	if err != nil {
		return err
	}

	return h.parse(m)
}

func (h *HomelabConfig) validate() error {
	err := validateGlobalConfig(&h.Global)
	if err != nil {
		return err
	}

	err = validateHostsConfig(h.Hosts)
	if err != nil {
		return err
	}

	err = validateIPAMConfig(&h.IPAM)
	if err != nil {
		return err
	}

	err = validateGroupsConfig(h.Groups)
	if err != nil {
		return err
	}

	err = validateContainersConfig(h.Containers)
	if err != nil {
		return err
	}

	return nil
}

func validateGlobalConfig(config *GlobalConfig) error {
	err := validateConfigEnv(config.Env, "global config")
	if err != nil {
		return err
	}

	err = validateMountsConfig(config.MountDefs, nil, nil, "global config mount defs")
	if err != nil {
		return err
	}

	err = validateGlobalContainerConfig(&config.Container, config.MountDefs)
	if err != nil {
		return err
	}

	return nil
}

func validateConfigEnv(config []ConfigEnv, location string) error {
	envs := make(map[string]bool)
	for _, e := range config {
		if len(e.Var) == 0 {
			return fmt.Errorf("empty env var in %s", location)
		}
		if envs[e.Var] {
			return fmt.Errorf("env var %s specified more than once in %s", e.Var, location)
		}
		envs[e.Var] = true

		if len(e.Value) == 0 && len(e.ValueCommand) == 0 {
			return fmt.Errorf("neither value nor valueCommand specified for env var %s in %s", e.Var, location)
		}
		if len(e.Value) > 0 && len(e.ValueCommand) > 0 {
			return fmt.Errorf("exactly one of value or valueCommand must be specified for env var %s in %s", e.Var, location)
		}
	}
	return nil
}

func validateContainerEnv(config []ContainerEnv, location string) error {
	envs := make(map[string]bool)
	for _, e := range config {
		if len(e.Var) == 0 {
			return fmt.Errorf("empty env var in %s", location)
		}
		if envs[e.Var] {
			return fmt.Errorf("env var %s specified more than once in %s", e.Var, location)
		}
		envs[e.Var] = true

		if len(e.Value) == 0 && len(e.ValueCommand) == 0 {
			return fmt.Errorf("neither value nor valueCommand specified for env var %s in %s", e.Var, location)
		}
		if len(e.Value) > 0 && len(e.ValueCommand) > 0 {
			return fmt.Errorf("exactly one of value or valueCommand must be specified for env var %s in %s", e.Var, location)
		}
	}
	return nil
}

func validateLabelsConfig(config []LabelConfig, location string) error {
	labels := make(map[string]bool)
	for _, l := range config {
		if len(l.Name) == 0 {
			return fmt.Errorf("empty label name in %s", location)
		}
		if labels[l.Name] {
			return fmt.Errorf("label name %s specified more than once in %s", l.Name, location)
		}
		labels[l.Name] = true

		if len(l.Value) == 0 {
			return fmt.Errorf("empty label value for label %s in %s", l.Name, location)
		}
	}
	return nil
}

func validateMountsConfig(config, commonConfig, globalDefs []MountConfig, location string) error {
	// First build a map of the mounts from the globalDefs (which should
	// already have been validated).
	globalMountDefs := make(map[string]bool)
	for _, m := range globalDefs {
		globalMountDefs[m.Name] = true
	}

	// Build a map of the mounts from the commonConfig next which acts
	// as the first set of mounts to apply. These should also have been
	// validated prior and hence we don't validate them here again.
	mounts := make(map[string]bool)
	for _, m := range commonConfig {
		mounts[m.Name] = true
	}

	// Finally iterate and validate the mounts in the current level config.
	for _, m := range config {
		if len(m.Name) == 0 {
			return fmt.Errorf("mount name is empty in %s", location)
		}
		if mounts[m.Name] {
			return fmt.Errorf("mount name %s defined more than once in %s", m.Name, location)
		}
		mounts[m.Name] = true

		if len(m.Type) == 0 && len(m.Src) == 0 && len(m.Dst) == 0 && !m.ReadOnly {
			// This is a mount with just the name. Match this against the
			// global mount defs.
			if !globalMountDefs[m.Name] {
				return fmt.Errorf("mount specified by just the name %s not found in defs", m.Name)
			}
			// No further validation needed for a mount referencing a def.
			return nil
		}

		if m.Type != "bind" {
			return fmt.Errorf("unsupported mount type %s for mount %s in %s", m.Type, m.Name, location)
		}
		if len(m.Src) == 0 {
			return fmt.Errorf("mount name %s has empty value for src in %s", m.Name, location)
		}
		if len(m.Dst) == 0 {
			return fmt.Errorf("mount name %s has empty value for dst in %s", m.Name, location)
		}
		if len(m.Options) > 0 {
			return fmt.Errorf("mount name %s specifies options in %s, that are not supported when mount type is bind", m.Name, location)
		}
	}
	return nil
}

func validateGlobalContainerConfig(config *GlobalContainerConfig, globalMountDefs []MountConfig) error {
	if config.StopTimeout < 0 {
		return fmt.Errorf("container stop timeout cannot be negative (%d) in global container config", config.StopTimeout)
	}
	err := validateContainerRestartPolicy(&config.RestartPolicy, "global container config")
	if err != nil {
		return err
	}
	err = validateContainerEnv(config.Env, "global container config")
	if err != nil {
		return err
	}
	err = validateMountsConfig(config.Mounts, nil, globalMountDefs, "global container config mounts")
	if err != nil {
		return err
	}
	err = validateLabelsConfig(config.Labels, "global container config")
	if err != nil {
		return err
	}
	return nil
}

func validateContainerRestartPolicy(config *ContainerRestartPolicyConfig, location string) error {
	if config.Mode != "on-failure" && config.MaxRetryCount != 0 {
		return fmt.Errorf("restart policy max retry count can be set in %s only when the mode is on-failure", location)
	}
	if len(config.Mode) == 0 {
		return nil
	}
	if _, err := restartPolicyModeFromString(config.Mode); err != nil {
		return fmt.Errorf("invalid restart policy mode %s in %s, valid values are %s", config.Mode, location, restartPolicyModeValidValues())
	}
	if config.MaxRetryCount < 0 {
		return fmt.Errorf("restart policy max retry count (%d) in %s cannot be negative", config.MaxRetryCount, location)
	}
	return nil
}

func validateIPAMConfig(config *IPAMConfig) error {
	networks := make(map[string]bool)
	hostInterfaces := make(map[string]bool)
	bridgeModeNetworks := config.Networks.BridgeModeNetworks
	prefixes := make(map[netip.Prefix]string)
	for _, n := range bridgeModeNetworks {
		if len(n.Name) == 0 {
			return fmt.Errorf("network name cannot be empty")
		}
		if networks[n.Name] {
			return fmt.Errorf("network %s defined more than once in the IPAM config", n.Name)
		}

		if len(n.HostInterfaceName) == 0 {
			return fmt.Errorf("host interface name of network %s cannot be empty", n.Name)
		}
		if hostInterfaces[n.HostInterfaceName] {
			return fmt.Errorf("host interface name %s of network %s is already used by another network in the IPAM config", n.HostInterfaceName, n.Name)
		}
		if n.Priority <= 0 {
			return fmt.Errorf("network %s has a non-positive priority %d", n.Name, n.Priority)
		}

		networks[n.Name] = true
		hostInterfaces[n.HostInterfaceName] = true
		prefix, err := netip.ParsePrefix(n.CIDR)
		if err != nil {
			return fmt.Errorf("CIDR %s of network %s is invalid, reason: %w", n.CIDR, n.Name, err)
		}
		netAddr := prefix.Addr()
		if !netAddr.Is4() {
			return fmt.Errorf("CIDR %s of network %s is not an IPv4 subnet CIDR", n.CIDR, n.Name)
		}
		masked := prefix.Masked()
		if masked.Addr() != netAddr {
			return fmt.Errorf("CIDR %s of network %s is not the same as the network address %s", n.CIDR, n.Name, masked)
		}
		prefixLen := prefix.Bits()
		if prefixLen > 30 {
			return fmt.Errorf("CIDR %s of network %s (prefix length: %d) has a prefix length more than 30 which makes the network unusable for container IP address allocations", n.CIDR, n.Name, prefixLen)
		}
		if !netAddr.IsPrivate() {
			return fmt.Errorf("CIDR %s of network %s is not within the RFC1918 private address space", n.CIDR, n.Name)
		}
		for pre, preNet := range prefixes {
			if prefix.Overlaps(pre) {
				return fmt.Errorf("CIDR %s of network %s overlaps with CIDR %s of network %s", n.CIDR, n.Name, pre, preNet)
			}
		}
		prefixes[prefix] = n.Name

		gatewayAddr := netAddr.Next()
		containers := make(map[ContainerReference]bool)
		containerIPs := make(map[netip.Addr]bool)
		for _, cip := range n.Containers {
			ip := cip.IP
			ct := cip.Container
			err := validateContainerReference(&ct)
			if err != nil {
				return fmt.Errorf("container IP config within network %s has invalid container reference, reason: %w", n.Name, err)
			}

			caddr, err := netip.ParseAddr(ip)
			if err != nil {
				return fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s has invalid IP %s, reason: %w", ct.Group, ct.Container, n.Name, ip, err)
			}
			if !prefix.Contains(caddr) {
				return fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s has IP %s that does not belong to the network CIDR %s", ct.Group, ct.Container, n.Name, ip, prefix)
			}
			if caddr.Compare(netAddr) == 0 {
				return fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s has IP %s matching the network address %s", ct.Group, ct.Container, n.Name, ip, netAddr)
			}
			if caddr.Compare(gatewayAddr) == 0 {
				return fmt.Errorf("container {Group:%s Container:%s} endpoint in network %s has IP %s matching the gateway address %s", ct.Group, ct.Container, n.Name, ip, gatewayAddr)
			}
			if containers[ct] {
				return fmt.Errorf("container {Group:%s Container:%s} has multiple endpoints in network %s", ct.Group, ct.Container, n.Name)
			}
			if containerIPs[caddr] {
				return fmt.Errorf("IP %s of container {Group:%s Container:%s} is already in use by another container in network %s", ip, ct.Group, ct.Container, n.Name)
			}

			containers[ct] = true
			containerIPs[caddr] = true
		}
	}
	containerModeNetworks := config.Networks.ContainerModeNetworks
	for _, n := range containerModeNetworks {
		if len(n.Name) == 0 {
			return fmt.Errorf("network name cannot be empty")
		}
		if networks[n.Name] {
			return fmt.Errorf("network %s defined more than once in the IPAM config", n.Name)
		}
		if n.Priority <= 0 {
			return fmt.Errorf("network %s has a non-positive priority %d", n.Name, n.Priority)
		}
		networks[n.Name] = true

		containers := make(map[ContainerReference]bool)
		for _, ct := range n.Containers {
			err := validateContainerReference(&ct)
			if err != nil {
				return fmt.Errorf("container IP config within network %s has invalid container reference, reason: %w", n.Name, err)
			}
			if containers[ct] {
				return fmt.Errorf("container {Group:%s Container:%s} is connected to multiple container mode network stacks", ct.Group, ct.Container)
			}
			containers[ct] = true
		}
	}
	return nil
}

func validateHostsConfig(hosts []HostConfig) error {
	hostNames := make(map[string]bool)
	for _, h := range hosts {
		if len(h.Name) == 0 {
			return fmt.Errorf("host name cannot be empty in the hosts config")
		}
		if hostNames[h.Name] {
			return fmt.Errorf("host %s defined more than once in the hosts config", h.Name)
		}
		hostNames[h.Name] = true

		containers := make(map[ContainerReference]bool)
		for _, ct := range h.AllowedContainers {
			err := validateContainerReference(&ct)
			if err != nil {
				return fmt.Errorf("allowed container config within host %s has invalid container reference, reason: %w", h.Name, err)
			}
			if containers[ct] {
				return fmt.Errorf("container {Group:%s Container:%s} defined more than once in the hosts config for host %s", ct.Group, ct.Container, h.Name)
			}
			containers[ct] = true
		}
	}
	return nil
}

func validateGroupsConfig(groups []ContainerGroupConfig) error {
	groupNames := make(map[string]bool)
	for _, g := range groups {
		if len(g.Name) == 0 {
			return fmt.Errorf("group name cannot be empty in the groups config")
		}
		if groupNames[g.Name] {
			return fmt.Errorf("group %s defined more than once in the groups config", g.Name)
		}
		if g.Order < 1 {
			return fmt.Errorf("group %s has a non-positive order %d", g.Name, g.Order)
		}

		groupNames[g.Name] = true
	}
	return nil
}

func validateContainersConfig(containers []ContainerConfig) error {
	// TODO: Perform the following (and more) validations:
	// Container configs:
	//     a. Parent group name is a valid group defined under group config.
	//     b. No duplicate container names within the same group.
	//     c. Order defined for all the containers.
	//     d. Image defined for all the containers.
	//     e. Validate mandatory properties of every device config.
	//     f. Validate manadatory properties of every container config mount.
	//     g. Mount pure name references are valid global config mount references.
	//     h. Validate manadatory properties of every container config env.
	//     i. Every container config env specifies exactly one of value or
	//        valueCommand, but not both.
	//     j. Validate mandatory properties of every published port config.
	//     k. Validate mandatory properties of every label config.
	return nil
}

func validateContainerReference(ref *ContainerReference) error {
	if len(ref.Group) == 0 {
		return fmt.Errorf("container reference cannot have an empty group name")
	}
	if len(ref.Container) == 0 {
		return fmt.Errorf("container reference cannot have an empty container name")
	}
	return nil
}
