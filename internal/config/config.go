package config

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/tuxdudehomelab/homelab/internal/config/env"
	"github.com/tuxdudehomelab/homelab/internal/utils"
	"gopkg.in/yaml.v3"
)

// Homelab represents the entire homelab deployment configuration.
type Homelab struct {
	Global     Global           `yaml:"global,omitempty" json:"global,omitempty"`
	IPAM       IPAM             `yaml:"ipam,omitempty" json:"ipam,omitempty"`
	Hosts      []Host           `yaml:"hosts,omitempty" json:"hosts,omitempty"`
	Groups     []ContainerGroup `yaml:"groups,omitempty" json:"groups,omitempty"`
	Containers []Container      `yaml:"containers,omitempty" json:"containers,omitempty"`
	Ignore     []IgnoredConfig  `yaml:"ignore,omitempty" json:"ignore,omitempty"`
}

// HomelabGroupsOnly represents a minimal group name information only version
// of the homelab deployment configuration.
type HomelabGroupsOnly struct {
	Groups []ContainerGroupNameOnly `yaml:"groups,omitempty" json:"groups,omitempty"`
}

// HomelabGroupsOnly represents a minimal container name information only
// version of the homelab deployment configuration.
type HomelabContainersOnly struct {
	Containers []ContainerNameOnly `yaml:"containers,omitempty" json:"containers,omitempty"`
}

// Global represents the configuration that will be applied
// across the entire homelab deployment.
type Global struct {
	BaseDir   string          `yaml:"baseDir,omitempty" json:"baseDir,omitempty"`
	Env       []ConfigEnv     `yaml:"env,omitempty" json:"env,omitempty"`
	MountDefs []Mount         `yaml:"mountDefs,omitempty" json:"mountDefs,omitempty"`
	Container GlobalContainer `yaml:"container,omitempty" json:"container,omitempty"`
}

// GlobalContainer represents container related configuration that
// will be applied globally across all containers.
type GlobalContainer struct {
	StopSignal    string                 `yaml:"stopSignal,omitempty" json:"stopSignal,omitempty"`
	StopTimeout   int                    `yaml:"stopTimeout,omitempty" json:"stopTimeout,omitempty"`
	RestartPolicy ContainerRestartPolicy `yaml:"restartPolicy,omitempty" json:"restartPolicy,omitempty"`
	DomainName    string                 `yaml:"domainName,omitempty" json:"domainName,omitempty"`
	DNSSearch     []string               `yaml:"dnsSearch,omitempty" json:"dnsSearch,omitempty"`
	Env           []ContainerEnv         `yaml:"env,omitempty" json:"env,omitempty"`
	Mounts        []Mount                `yaml:"mounts,omitempty" json:"mounts,omitempty"`
	Labels        []Label                `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// ConfigEnv is a pair of environment variable name and value that will be
// substituted in all string field values read from the homelab
// configuration file.
type ConfigEnv struct {
	Var          string `yaml:"var,omitempty" json:"var,omitempty"`
	Value        string `yaml:"value,omitempty" json:"value,omitempty"`
	ValueCommand string `yaml:"valueCommand,omitempty" json:"valueCommand,omitempty"`
}

// IPAM represents the IP Addressing and management information for
// all containers in the homelab configuration.
type IPAM struct {
	Networks Networks `yaml:"networks,omitempty" json:"networks,omitempty"`
}

// Networks represents all networks in the homelab configuration.
type Networks struct {
	BridgeModeNetworks    []BridgeModeNetwork    `yaml:"bridgeModeNetworks,omitempty" json:"bridgeModeNetworks,omitempty"`
	ContainerModeNetworks []ContainerModeNetwork `yaml:"containerModeNetworks,omitempty" json:"containerModeNetworks,omitempty"`
}

// BridgeModeNetwork represents a docker bridge mode network that one
// or more containers attach to.
type BridgeModeNetwork struct {
	Name              string        `yaml:"name,omitempty" json:"name,omitempty"`
	HostInterfaceName string        `yaml:"hostInterfaceName,omitempty" json:"hostInterfaceName,omitempty"`
	CIDR              string        `yaml:"cidr,omitempty" json:"cidr,omitempty"`
	Priority          int           `yaml:"priority,omitempty" json:"priority,omitempty"`
	Containers        []ContainerIP `yaml:"containers,omitempty" json:"containers,omitempty"`
}

// ContainerModeNetwork represents a container network meant to attach a
// container to another container's network stack.
type ContainerModeNetwork struct {
	Name                string               `yaml:"name,omitempty" json:"name,omitempty"`
	Container           ContainerReference   `yaml:"container,omitempty" json:"container,omitempty"`
	AttachingContainers []ContainerReference `yaml:"attachingContainers,omitempty" json:"attachingContainers,omitempty"`
}

// ContainerIP represents the IP information for a container.
type ContainerIP struct {
	IP        string             `yaml:"ip,omitempty" json:"ip,omitempty"`
	Container ContainerReference `yaml:"container,omitempty" json:"container,omitempty"`
}

// Host represents the host specific information.
type Host struct {
	Name              string               `yaml:"name,omitempty" json:"name,omitempty"`
	AllowedContainers []ContainerReference `yaml:"allowedContainers,omitempty" json:"allowedContainers,omitempty"`
}

// ContainerReference identifies a specific container part of a group.
type ContainerReference struct {
	Group     string `yaml:"group,omitempty" json:"group,omitempty"`
	Container string `yaml:"container,omitempty" json:"container,omitempty"`
}

// ContainerGroup represents a single logical container group, which is
// basically a collection of containers within.
type ContainerGroup struct {
	Name  string `yaml:"name,omitempty" json:"name,omitempty"`
	Order int    `yaml:"order,omitempty" json:"order,omitempty"`
}

// ContainerGroupNameOnly represents a minimal single logical container group that
// includes just the name of the group.
type ContainerGroupNameOnly struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
}

// Container represents a single docker container.
type Container struct {
	Info       ContainerReference     `yaml:"info,omitempty" json:"info,omitempty"`
	Config     ContainerConfigOptions `yaml:"config,omitempty" json:"config,omitempty"`
	Image      ContainerImage         `yaml:"image,omitempty" json:"image,omitempty"`
	Metadata   ContainerMetadata      `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Lifecycle  ContainerLifecycle     `yaml:"lifecycle,omitempty" json:"lifecycle,omitempty"`
	User       ContainerUser          `yaml:"user,omitempty" json:"user,omitempty"`
	Filesystem ContainerFilesystem    `yaml:"fs,omitempty" json:"fs,omitempty"`
	Network    ContainerNetwork       `yaml:"network,omitempty" json:"network,omitempty"`
	Security   ContainerSecurity      `yaml:"security,omitempty" json:"security,omitempty"`
	Health     ContainerHealth        `yaml:"health,omitempty" json:"health,omitempty"`
	Runtime    ContainerRuntime       `yaml:"runtime,omitempty" json:"runtime,omitempty"`
}

// ContainerNameOnly represents a single docker container with just the
// group and the container name.
type ContainerNameOnly struct {
	Info ContainerReference `yaml:"info,omitempty" json:"info,omitempty"`
}

// ContainerConfigOptions represents options that are applied while
// evaluating the config for this docker container.
type ContainerConfigOptions struct {
	Env []ConfigEnv `yaml:"env,omitempty" json:"env,omitempty"`
}

// ContainerImage respresents the image configuration for the docker
// container.
type ContainerImage struct {
	Image                   string `yaml:"image,omitempty" json:"image,omitempty"`
	SkipImagePull           bool   `yaml:"skipImagePull,omitempty" json:"skipImage,omitempty"`
	IgnoreImagePullFailures bool   `yaml:"ignoreImagePullFailures,omitempty" json:"ignoreImagePullFailures,omitempty"`
	PullImageBeforeStop     bool   `yaml:"pullImageBeforeStop,omitempty" json:"pullImageBeforeStop,omitempty"`
}

// ContainerMetadata represents the metadata for the docker container.
type ContainerMetadata struct {
	Labels []Label `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// ContainerLifecycle represents the lifecycle information for the
// docker container.
type ContainerLifecycle struct {
	Order         int                    `yaml:"order,omitempty" json:"order,omitempty"`
	StartPreHook  string                 `yaml:"startPreHook,omitempty" json:"startPreHook,omitempty"`
	RestartPolicy ContainerRestartPolicy `yaml:"restartPolicy,omitempty" json:"restartPolicy,omitempty"`
	AutoRemove    bool                   `yaml:"autoRemove,omitempty" json:"autoRemove,omitempty"`
	StopSignal    string                 `yaml:"stopSignal,omitempty" json:"stopSignal,omitempty"`
	StopTimeout   int                    `yaml:"stopTimeout,omitempty" json:"stopTimeout,omitempty"`
}

// ContainerUser represents the user and group information for the
// docker container.
type ContainerUser struct {
	User             string   `yaml:"user,omitempty" json:"user,omitempty"`
	PrimaryGroup     string   `yaml:"primaryGroup,omitempty" json:"primaryGroup,omitempty"`
	AdditionalGroups []string `yaml:"additionalGroups,omitempty" json:"additionalGroups,omitempty"`
}

// ContainerFilesystem represents the fileystem information for the
// docker container.
type ContainerFilesystem struct {
	ReadOnlyRootfs bool            `yaml:"readOnlyRootfs,omitempty" json:"readOnlyRootfs,omitempty"`
	Mounts         []Mount         `yaml:"mounts,omitempty" json:"mounts,omitempty"`
	Devices        ContainerDevice `yaml:"devices,omitempty" json:"devices,omitempty"`
}

// ContainerNetwork represents the networking information for the
// docker container.
type ContainerNetwork struct {
	HostName       string          `yaml:"hostName,omitempty" json:"hostName,omitempty"`
	DomainName     string          `yaml:"domainName,omitempty" json:"domainName,omitempty"`
	DNSServers     []string        `yaml:"dnsServers,omitempty" json:"dnsServers,omitempty"`
	DNSOptions     []string        `yaml:"dnsOptions,omitempty" json:"dnsOptions,omitempty"`
	DNSSearch      []string        `yaml:"dnsSearch,omitempty" json:"dnsSearch,omitempty"`
	ExtraHosts     []string        `yaml:"extraHosts,omitempty" json:"extraHosts,omitempty"`
	PublishedPorts []PublishedPort `yaml:"publishedPorts,omitempty" json:"publishedPorts,omitempty"`
}

// ContainerSecurity represents the security information for the
// docker container.
type ContainerSecurity struct {
	Privileged bool     `yaml:"privileged,omitempty" json:"privileged,omitempty"`
	Sysctls    []Sysctl `yaml:"sysctls,omitempty" json:"sysctls,omitempty"`
	CapAdd     []string `yaml:"capAdd,omitempty" json:"capAdd,omitempty"`
	CapDrop    []string `yaml:"capDrop,omitempty" json:"capDrop,omitempty"`
}

// ContainerHealth represents the health check options for the
// docker container.
type ContainerHealth struct {
	Cmd           []string `yaml:"cmd,omitempty" json:"cmd,omitempty"`
	Retries       int      `yaml:"retries,omitempty" json:"retries,omitempty"`
	Interval      string   `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout       string   `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	StartPeriod   string   `yaml:"startPeriod,omitempty" json:"startPeriod,omitempty"`
	StartInterval string   `yaml:"startInterval,omitempty" json:"startInterval,omitempty"`
}

// ContainerRuntime represents the execution and runtime information
// for the docker container.
type ContainerRuntime struct {
	AttachToTty bool           `yaml:"tty,omitempty" json:"tty,omitempty"`
	ShmSize     string         `yaml:"shmSize,omitempty" json:"shmSize,omitempty"`
	Env         []ContainerEnv `yaml:"env,omitempty" json:"env,omitempty"`
	Entrypoint  []string       `yaml:"entrypoint,omitempty" json:"entrypoint,omitempty"`
	Args        []string       `yaml:"args,omitempty" json:"args,omitempty"`
}

// Mount represents a filesystem mount.
type Mount struct {
	Name     string `yaml:"name,omitempty" json:"name,omitempty"`
	Type     string `yaml:"type,omitempty" json:"type,omitempty"`
	Src      string `yaml:"src,omitempty" json:"src,omitempty"`
	Dst      string `yaml:"dst,omitempty" json:"dst,omitempty"`
	ReadOnly bool   `yaml:"readOnly,omitempty" json:"readOnly,omitempty"`
	Options  string `yaml:"options,omitempty" json:"options,omitempty"`
}

// ContainerDevice represents the set of devices exposed to a container.
type ContainerDevice struct {
	Static         []Device `yaml:"static,omitempty" json:"static,omitempty"`
	DynamicCommand string   `yaml:"dynamic,omitempty" json:"dynamic,omitempty"`
}

// Device represents a device node that will be exposed to a container.
type Device struct {
	Src           string `yaml:"src,omitempty" json:"src,omitempty"`
	Dst           string `yaml:"dst,omitempty" json:"dst,omitempty"`
	DisallowRead  bool   `yaml:"disallowRead,omitempty" json:"disallowRead,omitempty"`
	DisallowWrite bool   `yaml:"disallowWrite,omitempty" json:"disallowWrite,omitempty"`
	DisallowMknod bool   `yaml:"disallowMknod,omitempty" json:"disallowMknod,omitempty"`
}

// Sysctl represents a sysctl config to apply to a container.
type Sysctl struct {
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

// PublishedPort represents a port published from a container.
type PublishedPort struct {
	ContainerPort string `yaml:"containerPort,omitempty" json:"containerPort,omitempty"`
	Protocol      string `yaml:"proto,omitempty" json:"proto,omitempty"`
	HostIP        string `yaml:"hostIp,omitempty" json:"hostIp,omitempty"`
	HostPort      string `yaml:"hostPort,omitempty" json:"hostPort,omitempty"`
}

// Label represents a label set on a container.
type Label struct {
	Name  string `yaml:"name,omitempty" json:"name,omitempty"`
	Value string `yaml:"value,omitempty" json:"value,omitempty"`
}

// ContainerRestartPolicy represents the restart policy for the container.
type ContainerRestartPolicy struct {
	Mode          string `yaml:"mode,omitempty" json:"mode,omitempty"`
	MaxRetryCount int    `yaml:"maxRetryCount,omitempty" json:"maxRetryCount,omitempty"`
}

// IgnoredConfig represents arbitrary information that can be thrown into
// the configuration that will not be directly interpreted by homelab. The
// common use case for this is to define reusable blocks of configuration
// using anchors and aliases in yaml.
type IgnoredConfig interface{}

func (h *Homelab) Parse(ctx context.Context, r io.Reader) error {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	err := dec.Decode(h)
	if err != nil {
		return fmt.Errorf("failed to parse homelab config, reason: %w", err)
	}
	// Clear out any parsed data under Ignore.
	h.Ignore = nil

	log(ctx).Tracef("Homelab Config:\n%v\n", utils.PrettyPrintJSON(h))
	return nil
}

func (h *HomelabGroupsOnly) Parse(ctx context.Context, r io.Reader) error {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(false)
	err := dec.Decode(h)
	if err != nil {
		return fmt.Errorf("failed to parse homelab groups only config, reason: %w", err)
	}
	return nil
}

func (h *HomelabGroupsOnly) ListGroups() []string {
	var groups []string
	for _, g := range h.Groups {
		groups = append(groups, g.Name)
	}
	return groups
}

func (h *HomelabContainersOnly) Parse(ctx context.Context, r io.Reader) error {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(false)
	err := dec.Decode(h)
	if err != nil {
		return fmt.Errorf("failed to parse homelab containers only config, reason: %w", err)
	}
	return nil
}

func (h *HomelabContainersOnly) ListContainers() []string {
	var containers []string
	for _, ct := range h.Containers {
		containers = append(containers, fmt.Sprintf("%s/%s", ct.Info.Group, ct.Info.Container))
	}
	slices.Sort(containers)
	return containers
}

func (g *Global) ApplyConfigEnv(env *env.ConfigEnvManager) {
	for i, m := range g.MountDefs {
		g.MountDefs[i].Src = env.Apply(m.Src)
		g.MountDefs[i].Dst = env.Apply(m.Dst)
		g.MountDefs[i].Options = env.Apply(m.Options)
	}
	g.Container.DomainName = env.Apply(g.Container.DomainName)
	for i, d := range g.Container.DNSSearch {
		g.Container.DNSSearch[i] = env.Apply(d)
	}
	for i, e := range g.Container.Env {
		g.Container.Env[i].Var = env.Apply(e.Var)
		g.Container.Env[i].Value = env.Apply(e.Value)
		g.Container.Env[i].ValueCommand = env.Apply(e.ValueCommand)
	}
	for i, m := range g.Container.Mounts {
		g.Container.Mounts[i].Src = env.Apply(m.Src)
		g.Container.Mounts[i].Dst = env.Apply(m.Dst)
		g.Container.Mounts[i].Options = env.Apply(m.Options)
	}
}

func (c *Container) ApplyConfigEnv(env *env.ConfigEnvManager) {
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
	for i, d := range c.Filesystem.Devices.Static {
		c.Filesystem.Devices.Static[i].Src = env.Apply(d.Src)
		c.Filesystem.Devices.Static[i].Dst = env.Apply(d.Dst)
	}
	c.Filesystem.Devices.DynamicCommand = env.Apply(c.Filesystem.Devices.DynamicCommand)
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
	for i, e := range c.Network.ExtraHosts {
		c.Network.ExtraHosts[i] = env.Apply(e)
	}
	for i, p := range c.Network.PublishedPorts {
		c.Network.PublishedPorts[i].ContainerPort = env.Apply(p.ContainerPort)
		c.Network.PublishedPorts[i].Protocol = env.Apply(p.Protocol)
		c.Network.PublishedPorts[i].HostIP = env.Apply(p.HostIP)
		c.Network.PublishedPorts[i].HostPort = env.Apply(p.HostPort)
	}
	for i, e := range c.Runtime.Env {
		c.Runtime.Env[i].Var = env.Apply(e.Var)
		c.Runtime.Env[i].Value = env.Apply(e.Value)
		c.Runtime.Env[i].ValueCommand = env.Apply(e.ValueCommand)
	}
}
