package config

import (
	"context"
	"fmt"
	"io"

	"github.com/tuxdudehomelab/homelab/internal/config/env"
	"github.com/tuxdudehomelab/homelab/internal/utils"
	"gopkg.in/yaml.v3"
)

// HomelabConfig represents the entire homelab deployment configuration.
type HomelabConfig struct {
	Global     GlobalConfig           `yaml:"global,omitempty" json:"global,omitempty"`
	IPAM       IPAMConfig             `yaml:"ipam,omitempty" json:"ipam,omitempty"`
	Hosts      []HostConfig           `yaml:"hosts,omitempty" json:"hosts,omitempty"`
	Groups     []ContainerGroupConfig `yaml:"groups,omitempty" json:"groups,omitempty"`
	Containers []ContainerConfig      `yaml:"containers,omitempty" json:"containers,omitempty"`
}

// GlobalConfig represents the configuration that will be applied
// across the entire homelab deployment.
type GlobalConfig struct {
	BaseDir   string                `yaml:"baseDir,omitempty" json:"baseDir,omitempty"`
	Env       []ConfigEnv           `yaml:"env,omitempty" json:"env,omitempty"`
	MountDefs []MountConfig         `yaml:"mountDefs,omitempty" json:"mountDefs,omitempty"`
	Container GlobalContainerConfig `yaml:"container,omitempty" json:"container,omitempty"`
}

// GlobalContainerConfig represents container related configuration that
// will be applied globally across all containers.
type GlobalContainerConfig struct {
	StopSignal    string                       `yaml:"stopSignal,omitempty" json:"stopSignal,omitempty"`
	StopTimeout   int                          `yaml:"stopTimeout,omitempty" json:"stopTimeout,omitempty"`
	RestartPolicy ContainerRestartPolicyConfig `yaml:"restartPolicy,omitempty" json:"restartPolicy,omitempty"`
	DomainName    string                       `yaml:"domainName,omitempty" json:"domainName,omitempty"`
	DNSSearch     []string                     `yaml:"dnsSearch,omitempty" json:"dnsSearch,omitempty"`
	Env           []ContainerEnv               `yaml:"env,omitempty" json:"env,omitempty"`
	Mounts        []MountConfig                `yaml:"mounts,omitempty" json:"mounts,omitempty"`
	Labels        []LabelConfig                `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// ConfigEnv is a pair of environment variable name and value that will be
// substituted in all string field values read from the homelab
// configuration file.
type ConfigEnv struct {
	Var          string `yaml:"var,omitempty" json:"var,omitempty"`
	Value        string `yaml:"value,omitempty" json:"value,omitempty"`
	ValueCommand string `yaml:"valueCommand,omitempty" json:"valueCommand,omitempty"`
}

// IPAMConfig represents the IP Addressing and management information for
// all containers in the homelab configuration.
type IPAMConfig struct {
	Networks NetworksConfig `yaml:"networks,omitempty" json:"networks,omitempty"`
}

// NetworksConfig represents all networks in the homelab configuration.
type NetworksConfig struct {
	BridgeModeNetworks    []BridgeModeNetworkConfig    `yaml:"bridgeModeNetworks,omitempty" json:"bridgeModeNetworks,omitempty"`
	ContainerModeNetworks []ContainerModeNetworkConfig `yaml:"containerModeNetworks,omitempty" json:"containerModeNetworks,omitempty"`
}

// BridgeModeNetworkConfig represents a docker bridge mode network that one
// or more containers attach to.
type BridgeModeNetworkConfig struct {
	Name              string              `yaml:"name,omitempty" json:"name,omitempty"`
	HostInterfaceName string              `yaml:"hostInterfaceName,omitempty" json:"hostInterfaceName,omitempty"`
	CIDR              string              `yaml:"cidr,omitempty" json:"cidr,omitempty"`
	Priority          int                 `yaml:"priority,omitempty" json:"priority,omitempty"`
	Containers        []ContainerIPConfig `yaml:"containers,omitempty" json:"containers,omitempty"`
}

// ContainerModeNetworkConfig represents a container network meant to attach a
// container to another container's network stack.
type ContainerModeNetworkConfig struct {
	Name                string               `yaml:"name,omitempty" json:"name,omitempty"`
	Container           ContainerReference   `yaml:"container,omitempty" json:"container,omitempty"`
	AttachingContainers []ContainerReference `yaml:"attachingContainers,omitempty" json:"attachingContainers,omitempty"`
}

// ContainerIP represents the IP information for a container.
type ContainerIPConfig struct {
	IP        string             `yaml:"ip,omitempty" json:"ip,omitempty"`
	Container ContainerReference `yaml:"container,omitempty" json:"container,omitempty"`
}

// HostConfig represents the host specific information.
type HostConfig struct {
	Name              string               `yaml:"name,omitempty" json:"name,omitempty"`
	AllowedContainers []ContainerReference `yaml:"allowedContainers,omitempty" json:"allowedContainers,omitempty"`
}

// ContainerReference identifies a specific container part of a group.
type ContainerReference struct {
	Group     string `yaml:"group,omitempty" json:"group,omitempty"`
	Container string `yaml:"container,omitempty" json:"container,omitempty"`
}

// ContainerGroupConfig represents a single logical container group, which is
// basically a collection of containers within.
type ContainerGroupConfig struct {
	Name  string `yaml:"name,omitempty" json:"name,omitempty"`
	Order int    `yaml:"order,omitempty" json:"order,omitempty"`
}

// ContainerConfig represents a single docker container.
type ContainerConfig struct {
	Info       ContainerReference        `yaml:"info,omitempty" json:"info,omitempty"`
	Config     ContainerConfigOptions    `yaml:"config,omitempty" json:"config,omitempty"`
	Image      ContainerImageConfig      `yaml:"image,omitempty" json:"image,omitempty"`
	Metadata   ContainerMetadataConfig   `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Lifecycle  ContainerLifecycleConfig  `yaml:"lifecycle,omitempty" json:"lifecycle,omitempty"`
	User       ContainerUserConfig       `yaml:"user,omitempty" json:"user,omitempty"`
	Filesystem ContainerFilesystemConfig `yaml:"fs,omitempty" json:"fs,omitempty"`
	Network    ContainerNetworkConfig    `yaml:"network,omitempty" json:"network,omitempty"`
	Security   ContainerSecurityConfig   `yaml:"security,omitempty" json:"security,omitempty"`
	Health     ContainerHealthConfig     `yaml:"health,omitempty" json:"health,omitempty"`
	Runtime    ContainerRuntimeConfig    `yaml:"runtime,omitempty" json:"runtime,omitempty"`
}

// ContainerConfigOptions represents options that are applied while
// evaluating the config for this docker container.
type ContainerConfigOptions struct {
	Env []ConfigEnv `yaml:"env,omitempty" json:"env,omitempty"`
}

// ContainerImageConfig respresents the image configuration for the docker
// container.
type ContainerImageConfig struct {
	Image                   string `yaml:"image,omitempty" json:"image,omitempty"`
	SkipImagePull           bool   `yaml:"skipImagePull,omitempty" json:"skipImage,omitempty"`
	IgnoreImagePullFailures bool   `yaml:"ignoreImagePullFailures,omitempty" json:"ignoreImagePullFailures,omitempty"`
	PullImageBeforeStop     bool   `yaml:"pullImageBeforeStop,omitempty" json:"pullImageBeforeStop,omitempty"`
}

// ContainerMetadataConfig represents the metadata for the docker container.
type ContainerMetadataConfig struct {
	Labels []LabelConfig `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// ContainerLifecycleConfig represents the lifecycle information for the
// docker container.
type ContainerLifecycleConfig struct {
	Order         int                          `yaml:"order,omitempty" json:"order,omitempty"`
	StartPreHook  string                       `yaml:"startPreHook,omitempty" json:"startPreHook,omitempty"`
	RestartPolicy ContainerRestartPolicyConfig `yaml:"restartPolicy,omitempty" json:"restartPolicy,omitempty"`
	AutoRemove    bool                         `yaml:"autoRemove,omitempty" json:"autoRemove,omitempty"`
	StopSignal    string                       `yaml:"stopSignal,omitempty" json:"stopSignal,omitempty"`
	StopTimeout   int                          `yaml:"stopTimeout,omitempty" json:"stopTimeout,omitempty"`
}

// ContainerUserConfig represents the user and group information for the
// docker container.
type ContainerUserConfig struct {
	User             string   `yaml:"user,omitempty" json:"user,omitempty"`
	PrimaryGroup     string   `yaml:"primaryGroup,omitempty" json:"primaryGroup,omitempty"`
	AdditionalGroups []string `yaml:"additionalGroups,omitempty" json:"additionalGroups,omitempty"`
}

// ContainerFilesystemConfig represents the fileystem information for the
// docker container.
type ContainerFilesystemConfig struct {
	ReadOnlyRootfs bool           `yaml:"readOnlyRootfs,omitempty" json:"readOnlyRootfs,omitempty"`
	Mounts         []MountConfig  `yaml:"mounts,omitempty" json:"mounts,omitempty"`
	Devices        []DeviceConfig `yaml:"devices,omitempty" json:"devices,omitempty"`
}

// ContainerNetworkConfig represents the networking information for the
// docker container.
type ContainerNetworkConfig struct {
	HostName       string                `yaml:"hostName,omitempty" json:"hostName,omitempty"`
	DomainName     string                `yaml:"domainName,omitempty" json:"domainName,omitempty"`
	DNSServers     []string              `yaml:"dnsServers,omitempty" json:"dnsServers,omitempty"`
	DNSOptions     []string              `yaml:"dnsOptions,omitempty" json:"dnsOptions,omitempty"`
	DNSSearch      []string              `yaml:"dnsSearch,omitempty" json:"dnsSearch,omitempty"`
	PublishedPorts []PublishedPortConfig `yaml:"publishedPorts,omitempty" json:"publishedPorts,omitempty"`
}

// ContainerSecurityConfig represents the security information for the
// docker container.
type ContainerSecurityConfig struct {
	Privileged bool           `yaml:"privileged,omitempty" json:"privileged,omitempty"`
	Sysctls    []SysctlConfig `yaml:"sysctls,omitempty" json:"sysctls,omitempty"`
	CapAdd     []string       `yaml:"capAdd,omitempty" json:"capAdd,omitempty"`
	CapDrop    []string       `yaml:"capDrop,omitempty" json:"capDrop,omitempty"`
}

// ContainerHealthConfig represents the health check options for the
// docker container.
type ContainerHealthConfig struct {
	Cmd           []string `yaml:"cmd,omitempty" json:"cmd,omitempty"`
	Retries       int      `yaml:"retries,omitempty" json:"retries,omitempty"`
	Interval      string   `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout       string   `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	StartPeriod   string   `yaml:"startPeriod,omitempty" json:"startPeriod,omitempty"`
	StartInterval string   `yaml:"startInterval,omitempty" json:"startInterval,omitempty"`
}

// ContainerRuntimeConfig represents the execution and runtime information
// for the docker container.
type ContainerRuntimeConfig struct {
	AttachToTty bool           `yaml:"tty,omitempty" json:"tty,omitempty"`
	ShmSize     string         `yaml:"shmSize,omitempty" json:"shmSize,omitempty"`
	Env         []ContainerEnv `yaml:"env,omitempty" json:"env,omitempty"`
	Entrypoint  []string       `yaml:"entrypoint,omitempty" json:"entrypoint,omitempty"`
	Args        []string       `yaml:"args,omitempty" json:"args,omitempty"`
}

// MountConfig represents a filesystem mount.
type MountConfig struct {
	Name     string `yaml:"name,omitempty" json:"name,omitempty"`
	Type     string `yaml:"type,omitempty" json:"type,omitempty"`
	Src      string `yaml:"src,omitempty" json:"src,omitempty"`
	Dst      string `yaml:"dst,omitempty" json:"dst,omitempty"`
	ReadOnly bool   `yaml:"readOnly,omitempty" json:"readOnly,omitempty"`
	Options  string `yaml:"options,omitempty" json:"options,omitempty"`
}

// DeviceConfig represents a device node that will be exposed to a container.
type DeviceConfig struct {
	Src           string `yaml:"src,omitempty" json:"src,omitempty"`
	Dst           string `yaml:"dst,omitempty" json:"dst,omitempty"`
	DisallowRead  bool   `yaml:"disallowRead,omitempty" json:"disallowRead,omitempty"`
	DisallowWrite bool   `yaml:"disallowWrite,omitempty" json:"disallowWrite,omitempty"`
	DisallowMknod bool   `yaml:"disallowMknod,omitempty" json:"disallowMknod,omitempty"`
}

// SysctlConfig represents a sysctl config to apply to a container.
type SysctlConfig struct {
	Key   string `yaml:"key,omitempty" json:"key,omitempty"`
	Value string `yaml:"value,omitempty" json:"value,omitempty"`
}

// ContainerEnv represents an environment variable and value pair that will be set
// on the specified container.
type ContainerEnv struct {
	Var          string `yaml:"var,omitempty" json:"var,omitempty"`
	Value        string `yaml:"value,omitempty" json:"value,omitempty"`
	ValueCommand string `yaml:"valueCommand,omitempty" json:"valueCommand,omitempty"`
}

// PublishedPortConfig represents a port published from a container.
type PublishedPortConfig struct {
	ContainerPort int    `yaml:"containerPort,omitempty" json:"containerPort,omitempty"`
	Protocol      string `yaml:"proto,omitempty" json:"proto,omitempty"`
	HostIP        string `yaml:"hostIp,omitempty" json:"hostIp,omitempty"`
	HostPort      int    `yaml:"hostPort,omitempty" json:"hostPort,omitempty"`
}

// LabelConfig represents a label set on a container.
type LabelConfig struct {
	Name  string `yaml:"name,omitempty" json:"name,omitempty"`
	Value string `yaml:"value,omitempty" json:"value,omitempty"`
}

type ContainerRestartPolicyConfig struct {
	Mode          string `yaml:"mode,omitempty" json:"mode,omitempty"`
	MaxRetryCount int    `yaml:"maxRetryCount,omitempty" json:"maxRetryCount,omitempty"`
}

func (h *HomelabConfig) Parse(ctx context.Context, r io.Reader) error {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	err := dec.Decode(h)
	if err != nil {
		return fmt.Errorf("failed to parse homelab config, reason: %w", err)
	}

	log(ctx).Tracef("Homelab Config:\n%v\n", utils.PrettyPrintJSON(h))
	return nil
}

func (c *ContainerConfig) ApplyConfigEnv(env *env.ConfigEnv) {
	c.Lifecycle.StartPreHook = env.Apply(c.Lifecycle.StartPreHook)
	c.User.User = env.Apply(c.User.User)
	c.User.PrimaryGroup = env.Apply(c.User.PrimaryGroup)
	for i, g := range c.User.AdditionalGroups {
		c.User.AdditionalGroups[i] = env.Apply(g)
	}
	for i, m := range c.Filesystem.Mounts {
		c.Filesystem.Mounts[i].Src = env.Apply(m.Src)
		c.Filesystem.Mounts[i].Dst = env.Apply(m.Dst)
		c.Filesystem.Mounts[i].Options = env.Apply(m.Options)
	}
	for i, d := range c.Filesystem.Devices {
		c.Filesystem.Devices[i].Src = env.Apply(d.Src)
		c.Filesystem.Devices[i].Dst = env.Apply(d.Dst)
	}
	c.Network.HostName = env.Apply(c.Network.HostName)
	c.Network.DomainName = env.Apply(c.Network.DomainName)
	for i, d := range c.Network.DNSServers {
		c.Network.DNSServers[i] = env.Apply(d)
	}
	for i, d := range c.Network.DNSOptions {
		c.Network.DNSOptions[i] = env.Apply(d)
	}
	for i, d := range c.Network.DNSSearch {
		c.Network.DNSSearch[i] = env.Apply(d)
	}
	for i, p := range c.Network.PublishedPorts {
		c.Network.PublishedPorts[i].HostIP = env.Apply(p.HostIP)
	}
	for i, e := range c.Runtime.Env {
		c.Runtime.Env[i].Var = env.Apply(e.Var)
		c.Runtime.Env[i].Value = env.Apply(e.Value)
		c.Runtime.Env[i].ValueCommand = env.Apply(e.ValueCommand)
	}
}
