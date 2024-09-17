package main

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var parseConfigUsingReaderTests = []struct {
	name   string
	config string
	want   HomelabConfig
}{
	{
		name: "Valid extensive config",
		config: `
global:
  env:
    - var: MY_VAR_1
      value: MY_VAR_1_VALUE
    - var: MY_VAR_2
      value: MY_VAR_2_VALUE
  mountDefs:
    - name: vol-def-1
      src: /abc/def/ghi
      dst: /pqr/stu/vwx
      readOnly: true
    - name: vol-def-2
      src: /abc1/def1
      dst: /pqr2/stu2/vwx2
  container:
    stopSignal: SIGTERM
    stopTimeout: 8
    restartPolicy: unless-stopped
    domainName: example.tld
    dnsSearch:
      - dns-search-1
      - dns-search-2
    env:
      - var: MY_CONTAINER_ENV_VAR_1
        value: MY_CONTAINER_ENV_VAR_1_VALUE
      - var: MY_CONTAINER_ENV_VAR_2
        value: MY_CONTAINER_ENV_VAR_2_VALUE
    mounts:
      - name: vol-def-1
      - name: vol-def-2
      - name: vol-def-3
        src: /foo
        dst: /bar
        readOnly: true
    labels:
      - name: my-label-1
        value: my-label-1-value
      - name: my-label-2
        value: my-label-2-value
ipam:
  networks:
    bridgeModeNetworks:
      - name: group1-bridge
        hostInterfaceName: docker-grp1
        cidr: 172.18.18.0/24
        priority: 1
        containers:
          - ip: 172.18.18.11
            container:
              group: group1
              container: ct1
          - ip: 172.18.18.12
            container:
              group: group1
              container: ct2
      - name: group2-bridge
        hostInterfaceName: docker-grp2
        cidr: 172.18.19.0/24
        priority: 1
        containers:
          - ip: 172.18.19.11
            container:
              group: group2
              container: ct3
      - name: common-bridge
        hostInterfaceName: docker-cmn
        cidr: 172.18.20.0/24
        priority: 2
        containers:
          - ip: 172.18.20.11
            container:
              group: group1
              container: ct1
          - ip: 172.18.20.12
            container:
              group: group1
              container: ct2
          - ip: 172.18.20.13
            container:
              group: group2
              container: ct3
    containerModeNetworks:
      - name: group3-ct4
        priority: 1
        containers:
          - group: group3
            container: ct5
          - group: group3
            container: ct6
          - group: group3
            container: ct7
hosts:
  - name: host1
    allowedContainers:
      - group: group1
        container: ct1
      - group: group2
        container: ct3
  - name: host2
    allowedContainers:
      - group: group1
        container: ct2
  - name: host3
groups:
  - name: group1
    order: 1
  - name: group2
    order: 2
  - name: group3
    order: 3
containers:
  - info:
      group: group1
      container: ct1
    image:
      image: tuxdude/homelab-base:master
      skipImagePull: true
      ignoreImagePullFailures: true
      pullImageBeforeStop: true
    metadata:
      labels:
        - name: my.ct1.label.name.1
          value: my.ct1.label.value.1
        - name: my.ct1.label.name.2
          value: my.ct1.label.value.2
    lifecycle:
      order: 1
      startPreHook: $$SCRIPTS_DIR$$/my-start-prehook.sh
      restartPolicy: always
      autoRemove: true
      stopSignal: SIGHUP
      stopTimeout: 10
    user:
      user: $$CURRENT_USER$$
      primaryGroup: $$CURRENT_GROUP$$
      additionalGroups:
        - dialout
        - someRandomGroup
    fs:
      readOnlyRootfs: true
      mounts:
        - name: blocky-config-mount
          type: bind
          src: $$CONFIG_DIR$$/generated/config.yml
          dst: /data/blocky/config/config.yml
          readOnly: true
        - name: homelab-self-signed-tls-cert
        - name: tmpfs-mount
          type: tmpfs
          dst: /tmp/cache
          options: tmpfs-size=100000000
      devices:
        - src: /dev/foo
          dst: /dev/bar
        - src: /dev/dev1
          dst: /dev/dev2
    network:
      hostName: foobar
      domainName: somedomain
      dnsServers:
        - 1.1.1.1
        - 1.0.0.1
      dnsOptions:
        - dns-option-1
        - dns-option-2
      dnsSearch:
        - dns-ct-search-1
        - dns-ct-search-2
      publishedPorts:
        - containerPort: 53
          proto: tcp
          hostIp: $$HOST_IP$$
          hostPort: 53
        - containerPort: 53
          proto: udp
          hostIp: $$HOST_IP$$
          hostPort: 53
    security:
      privileged: true
      sysctls:
        - key: net.ipv4.ip_forward
          value: 1
        - key: net.ipv4.conf.all.src_valid_mark
          value: 1
      capAdd:
        - SYS_RAWIO
        - SYS_ADMIN
      capDrop:
        - NET_ADMIN
        - SYS_MODULE
    runtime:
      tty: true
      shmSize: 1g
      healthCmd: my-health-cmd
      env:
        - var: MY_ENV
          value: MY_ENV_VALUE
        - var: MY_ENV_2
          valueCommand: cat /foo/bar/baz.txt
        - var: MY_ENV_3
          value: SomeHostName.$$HUMAN_FRIENDLY_HOST_NAME$$.SomeDomainName
      entrypoint:
        - my-custom-entrypoint
        - ep-arg1
        - ep-arg2
      args:
        - foo
        - bar
        - baz`,
		want: HomelabConfig{
			Global: GlobalConfig{
				Env: []GlobalEnvConfig{
					{
						Var:   "MY_VAR_1",
						Value: "MY_VAR_1_VALUE",
					},
					{
						Var:   "MY_VAR_2",
						Value: "MY_VAR_2_VALUE",
					},
				},
				MountDefs: []MountConfig{
					{
						Name:     "vol-def-1",
						Src:      "/abc/def/ghi",
						Dst:      "/pqr/stu/vwx",
						ReadOnly: true,
					},
					{
						Name: "vol-def-2",
						Src:  "/abc1/def1",
						Dst:  "/pqr2/stu2/vwx2",
					},
				},
				Container: GlobalContainerConfig{
					StopSignal:    "SIGTERM",
					StopTimeout:   8,
					RestartPolicy: "unless-stopped",
					DomainName:    "example.tld",
					DNSSearch: []string{
						"dns-search-1",
						"dns-search-2",
					},
					Env: []ContainerEnvConfig{
						{
							Var:   "MY_CONTAINER_ENV_VAR_1",
							Value: "MY_CONTAINER_ENV_VAR_1_VALUE",
						},
						{
							Var:   "MY_CONTAINER_ENV_VAR_2",
							Value: "MY_CONTAINER_ENV_VAR_2_VALUE",
						},
					},
					Mounts: []MountConfig{
						{
							Name: "vol-def-1",
						},
						{
							Name: "vol-def-2",
						},
						{
							Name:     "vol-def-3",
							Src:      "/foo",
							Dst:      "/bar",
							ReadOnly: true,
						},
					},
					Labels: []LabelConfig{
						{
							Name:  "my-label-1",
							Value: "my-label-1-value",
						},
						{
							Name:  "my-label-2",
							Value: "my-label-2-value",
						},
					},
				},
			},
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "group1-bridge",
							HostInterfaceName: "docker-grp1",
							CIDR:              "172.18.18.0/24",
							Priority:          1,
							Containers: []ContainerIPConfig{
								{
									IP: "172.18.18.11",
									Container: ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
								{
									IP: "172.18.18.12",
									Container: ContainerReference{
										Group:     "group1",
										Container: "ct2",
									},
								},
							},
						},
						{
							Name:              "group2-bridge",
							HostInterfaceName: "docker-grp2",
							CIDR:              "172.18.19.0/24",
							Priority:          1,
							Containers: []ContainerIPConfig{
								{
									IP: "172.18.19.11",
									Container: ContainerReference{
										Group:     "group2",
										Container: "ct3",
									},
								},
							},
						},
						{
							Name:              "common-bridge",
							HostInterfaceName: "docker-cmn",
							CIDR:              "172.18.20.0/24",
							Priority:          2,
							Containers: []ContainerIPConfig{
								{
									IP: "172.18.20.11",
									Container: ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
								{
									IP: "172.18.20.12",
									Container: ContainerReference{
										Group:     "group1",
										Container: "ct2",
									},
								},
								{
									IP: "172.18.20.13",
									Container: ContainerReference{
										Group:     "group2",
										Container: "ct3",
									},
								},
							},
						},
					},
					ContainerModeNetworks: []ContainerModeNetworkConfig{
						{
							Name:     "group3-ct4",
							Priority: 1,
							Containers: []ContainerReference{
								{
									Group:     "group3",
									Container: "ct5",
								},
								{
									Group:     "group3",
									Container: "ct6",
								},
								{
									Group:     "group3",
									Container: "ct7",
								},
							},
						},
					},
				},
			},
			Hosts: []HostConfig{
				{
					Name: "host1",
					AllowedContainers: []ContainerReference{
						{
							Group:     "group1",
							Container: "ct1",
						},
						{
							Group:     "group2",
							Container: "ct3",
						},
					},
				},
				{
					Name: "host2",
					AllowedContainers: []ContainerReference{
						{
							Group:     "group1",
							Container: "ct2",
						},
					},
				},
				{
					Name: "host3",
				},
			},
			Groups: []ContainerGroupConfig{
				{
					Name:  "group1",
					Order: 1,
				},
				{
					Name:  "group2",
					Order: 2,
				},
				{
					Name:  "group3",
					Order: 3,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "group1",
						Container: "ct1",
					},
					Image: ContainerImageConfig{
						Image:                   "tuxdude/homelab-base:master",
						SkipImagePull:           true,
						IgnoreImagePullFailures: true,
						PullImageBeforeStop:     true,
					},
					Metadata: ContainerMetadataConfig{
						Labels: []LabelConfig{
							{
								Name:  "my.ct1.label.name.1",
								Value: "my.ct1.label.value.1",
							},
							{
								Name:  "my.ct1.label.name.2",
								Value: "my.ct1.label.value.2",
							},
						},
					},
					Lifecycle: ContainerLifecycleConfig{
						Order:         1,
						StartPreHook:  "$$SCRIPTS_DIR$$/my-start-prehook.sh",
						RestartPolicy: "always",
						AutoRemove:    true,
						StopSignal:    "SIGHUP",
						StopTimeout:   10,
					},
					User: ContainerUserConfig{
						User:         "$$CURRENT_USER$$",
						PrimaryGroup: "$$CURRENT_GROUP$$",
						AdditionalGroups: []string{
							"dialout",
							"someRandomGroup",
						},
					},
					Filesystem: ContainerFilesystemConfig{
						ReadOnlyRootfs: true,
						Mounts: []MountConfig{
							{
								Name:     "blocky-config-mount",
								Type:     "bind",
								Src:      "$$CONFIG_DIR$$/generated/config.yml",
								Dst:      "/data/blocky/config/config.yml",
								ReadOnly: true,
							},
							{
								Name: "homelab-self-signed-tls-cert",
							},
							{
								Name:    "tmpfs-mount",
								Type:    "tmpfs",
								Dst:     "/tmp/cache",
								Options: "tmpfs-size=100000000",
							},
						},
						Devices: []DeviceConfig{
							{
								Src: "/dev/foo",
								Dst: "/dev/bar",
							},
							{
								Src: "/dev/dev1",
								Dst: "/dev/dev2",
							},
						},
					},
					Network: ContainerNetworkConfig{
						HostName:   "foobar",
						DomainName: "somedomain",
						DNSServers: []string{
							"1.1.1.1",
							"1.0.0.1",
						},
						DNSOptions: []string{
							"dns-option-1",
							"dns-option-2",
						},
						DNSSearch: []string{
							"dns-ct-search-1",
							"dns-ct-search-2",
						},
						PublishedPorts: []PublishedPortConfig{
							{
								ContainerPort: 53,
								Proto:         "tcp",
								HostIP:        "$$HOST_IP$$",
								HostPort:      53,
							},
							{
								ContainerPort: 53,
								Proto:         "udp",
								HostIP:        "$$HOST_IP$$",
								HostPort:      53,
							},
						},
					},
					Security: ContainerSecurityConfig{
						Privileged: true,
						Sysctls: []SysctlConfig{
							{
								Key:   "net.ipv4.ip_forward",
								Value: "1",
							},
							{
								Key:   "net.ipv4.conf.all.src_valid_mark",
								Value: "1",
							},
						},
						CapAdd: []string{
							"SYS_RAWIO",
							"SYS_ADMIN",
						},
						CapDrop: []string{
							"NET_ADMIN",
							"SYS_MODULE",
						},
					},
					Runtime: ContainerRuntimeConfig{
						AttachToTty: true,
						ShmSize:     "1g",
						HealthCmd:   "my-health-cmd",
						Env: []ContainerEnvConfig{
							{
								Var:   "MY_ENV",
								Value: "MY_ENV_VALUE",
							},
							{
								Var:          "MY_ENV_2",
								ValueCommand: "cat /foo/bar/baz.txt",
							},
							{
								Var:   "MY_ENV_3",
								Value: "SomeHostName.$$HUMAN_FRIENDLY_HOST_NAME$$.SomeDomainName",
							},
						},
						Entrypoint: []string{
							"my-custom-entrypoint",
							"ep-arg1",
							"ep-arg2",
						},
						Args: []string{
							"foo",
							"bar",
							"baz",
						},
					},
				},
			},
		},
	},
	{
		name: "Valid Groups Only config",
		config: `
groups:
  - name: group1
    order: 1
  - name: group2
  - name: group3
    order: 2`,
		want: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "group1",
					Order: 1,
				},
				{
					Name: "group2",
				},
				{
					Name:  "group3",
					Order: 2,
				},
			},
		},
	},
}

func TestParseConfigUsingReader(t *testing.T) {
	for _, test := range parseConfigUsingReaderTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := strings.NewReader(tc.config)
			got := HomelabConfig{}
			if gotErr := got.parse(input); nil != gotErr {
				t.Errorf(
					"HomelabConfig.parse()\nTest Case: %q\nFailure: gotErr != nil\nReason: %v",
					tc.name, gotErr)
				return
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf(
					"HomelabConfig.parse()\nTest Case: %q\nFailure: got did not match the want config\nDiff(-want +got): %s", tc.name, diff)
			}
		})
	}
}

var validParseConfigsFromPathTests = []struct {
	name        string
	configsPath string
	want        HomelabConfig
}{
	{
		name:        "Valid multi file config",
		configsPath: "parse-configs-valid",
		want: HomelabConfig{
			Global: GlobalConfig{
				Env: []GlobalEnvConfig{
					{
						Var:   "MY_GLOBAL_FOO",
						Value: "MY_GLOBAL_BAR",
					},
				},
				Container: GlobalContainerConfig{
					StopSignal:    "SIGTERM",
					StopTimeout:   5,
					RestartPolicy: "unless-stopped",
					DomainName:    "somedomain",
				},
			},
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
							Priority:          1,
							Containers: []ContainerIPConfig{
								{
									IP: "172.18.100.11",
									Container: ContainerReference{
										Group:     "g1",
										Container: "c1",
									},
								},
								{
									IP: "172.18.100.12",
									Container: ContainerReference{
										Group:     "g1",
										Container: "c2",
									},
								},
							},
						},
						{
							Name:              "net2",
							HostInterfaceName: "docker-net2",
							CIDR:              "172.18.101.0/24",
							Priority:          1,
							Containers: []ContainerIPConfig{
								{
									IP: "172.18.101.21",
									Container: ContainerReference{
										Group:     "g2",
										Container: "c3",
									},
								},
							},
						},
						{
							Name:              "net-common",
							HostInterfaceName: "docker-cmn",
							CIDR:              "172.19.200.0/24",
							Priority:          2,
							Containers: []ContainerIPConfig{
								{
									IP: "172.19.200.201",
									Container: ContainerReference{
										Group:     "g1",
										Container: "c1",
									},
								},
								{
									IP: "172.19.200.202",
									Container: ContainerReference{
										Group:     "g1",
										Container: "c2",
									},
								},
								{
									IP: "172.19.200.203",
									Container: ContainerReference{
										Group:     "g2",
										Container: "c3",
									},
								},
							},
						},
					},
					ContainerModeNetworks: []ContainerModeNetworkConfig{
						{
							Name:     "g3-c4",
							Priority: 1,
							Containers: []ContainerReference{
								{
									Group:     "g3",
									Container: "c5",
								},
								{
									Group:     "g3",
									Container: "c6",
								},
								{
									Group:     "g3",
									Container: "c7",
								},
							},
						},
					},
				},
			},
			Hosts: []HostConfig{
				{
					Name: "host1",
					AllowedContainers: []ContainerReference{
						{
							Group:     "g1",
							Container: "c1",
						},
						{
							Group:     "g3",
							Container: "c4",
						},
					},
				},
				{
					Name: "host2",
				},
				{
					Name: "host3",
					AllowedContainers: []ContainerReference{
						{
							Group:     "g2",
							Container: "c3",
						},
					},
				},
			},
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
				{
					Name:  "g2",
					Order: 2,
				},
				{
					Name:  "g3",
					Order: 3,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "abc/xyz",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
				},
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c2",
					},
					Image: ContainerImageConfig{
						Image: "abc/xyz2",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 2,
					},
				},
				{
					Info: ContainerReference{
						Group:     "g2",
						Container: "c3",
					},
					Image: ContainerImageConfig{
						Image: "abc/xyz3",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
				},
				{
					Info: ContainerReference{
						Group:     "g3",
						Container: "c4",
					},
					Image: ContainerImageConfig{
						Image: "abc/xyz4",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
				},
			},
		},
	},
}

func TestParseConfigsFromPath(t *testing.T) {
	for _, test := range validParseConfigsFromPathTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := fmt.Sprintf("%s/testdata/%s", pwd(), tc.configsPath)
			got := HomelabConfig{}
			if gotErr := got.parseConfigs(p); nil != gotErr {
				t.Errorf(
					"HomelabConfig.parseConfigs()\nTest Case: %q\nFailure: gotErr != nil\nReason: %v",
					tc.name, gotErr)
				return
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf(
					"HomelabConfig.parseConfigs()\nTest Case: %q\nFailure: got did not match the want config\nDiff(-want +got): %s", tc.name, diff)
			}
		})
	}
}

var parseConfigsErrorTests = []struct {
	name        string
	configsPath string
	want        string
}{
	{
		name:        "Non-existing configs dir path",
		configsPath: "foo-bar",
		want:        `os.Stat\(\) failed on homelab configs path, reason: stat [^ ]+foo-bar: no such file or directory`,
	},
	{
		name:        "No configs",
		configsPath: "parse-configs-invalid-empty-dir",
		want:        `no homelab configs found in [^ ]+/testdata/parse-configs-invalid-empty-dir`,
	},
	{
		name:        "File configs dir path",
		configsPath: "parse-configs-invalid-with-empty-file/.empty",
		want:        `homelab configs path [^ ]+/testdata/parse-configs-invalid-with-empty-file/.empty must be a directory`,
	},
	{
		name:        "Unreadable config dir",
		configsPath: "/root",
		want:        `failed to read contents of directory /root, reason: open /root: permission denied`,
	},
	{
		name:        "Unreadable config file",
		configsPath: "parse-configs-invalid-unreadable-config",
		want:        `failed to read homelab config file [^ ]+/testdata/parse-configs-invalid-unreadable-config/invalid-symlink.yaml, reason: open [^ ]+/testdata/parse-configs-invalid-unreadable-config/invalid-symlink.yaml: no such file or directory`,
	},
	{
		name:        "Deep merge configs fail",
		configsPath: "parse-configs-invalid-deepmerge-fail",
		want:        `failed to deep merge config file [^ ]+parse-configs-invalid-deepmerge-fail/config2.yaml, reason: error due to parameter with value of primitive type: only maps and slices/arrays can be merged, which means you cannot have define the same key twice for parameters that are not maps or slices/arrays`,
	},
	{
		name:        "Invalid config key",
		configsPath: "parse-configs-invalid-config-key",
		want:        `(?s)failed to parse homelab config, reason: yaml: unmarshal errors:.+: field someInvalidKey not found in type main\.GlobalContainerConfig`,
	},
}

func TestParseConfigsFromPathErrors(t *testing.T) {
	for _, test := range parseConfigsErrorTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := tc.configsPath
			if !strings.HasPrefix(tc.configsPath, "/") {
				p = fmt.Sprintf("%s/testdata/%s", pwd(), tc.configsPath)
			}
			c := HomelabConfig{}
			gotErr := c.parseConfigs(p)
			if gotErr == nil {
				t.Errorf(
					"HomelabConfig.parseConfigs()\nTest Case: %q\nFailure: gotErr == nil\nReason: want = %q",
					tc.name, tc.want)
				return
			}

			match, err := regexp.MatchString(tc.want, gotErr.Error())
			if err != nil {
				t.Errorf(
					"HomelabConfig.parseConfigs()\nTest Case: %q\nFailure: unexpected exception while matching against gotErr error string\nReason: error = %v", tc.name, err)
				return
			}

			if !match {
				t.Errorf(
					"HomelabConfig.parseConfigs()\nTest Case: %q\nFailure: gotErr did not match the want regex\nReason:\n\ngotErr = %q\n\twant = %q", tc.name, gotErr, tc.want)
			}
		})
	}
}

var validateConfigTests = []struct {
	name   string
	config HomelabConfig
}{
	{
		name:   "Valid Empty config",
		config: HomelabConfig{},
	},
	{
		name: "Valid IPAM config",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
							Priority:          1,
						},
						{
							Name:              "net2",
							HostInterfaceName: "docker-net2",
							CIDR:              "172.18.101.0/24",
							Priority:          1,
						},
					},
					ContainerModeNetworks: []ContainerModeNetworkConfig{
						{
							Name:     "net3",
							Priority: 1,
						},
						{
							Name:     "net4",
							Priority: 1,
						},
					},
				},
			},
		},
	},
}

func TestValidateConfig(t *testing.T) {
	for _, test := range validateConfigTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if gotErr := tc.config.validate(); nil != gotErr {
				t.Errorf(
					"HomelabConfig.validate()\nTest Case: %q\nFailure: gotErr != nil\nReason: %v",
					tc.name, gotErr)
				return
			}
		})
	}
}

var validateConfigErrorTests = []struct {
	name   string
	config HomelabConfig
	want   string
}{
	{
		name: "Empty Global Config Env Var",
		config: HomelabConfig{
			Global: GlobalConfig{
				Env: []GlobalEnvConfig{
					{
						Value: "foo-bar",
					},
				},
			},
		},
		want: `empty env var in global config`,
	},
	{
		name: "Duplicate Global Config Env Var",
		config: HomelabConfig{
			Global: GlobalConfig{
				Env: []GlobalEnvConfig{
					{
						Var:   "FOO",
						Value: "foo-bar",
					},
					{
						Var:   "FOO2",
						Value: "foo-bar-2",
					},
					{
						Var:   "FOO",
						Value: "foo-bar-3",
					},
				},
			},
		},
		want: `env var FOO specified more than once in global config`,
	},
	{
		name: "Global Config Env Var Without Value And ValueCommand",
		config: HomelabConfig{
			Global: GlobalConfig{
				Env: []GlobalEnvConfig{
					{
						Var: "FOO",
					},
				},
			},
		},
		want: `neither value nor valueCommand specified for env var FOO in global config`,
	},
	{
		name: "Global Config Env Var With Both Value And ValueCommand",
		config: HomelabConfig{
			Global: GlobalConfig{
				Env: []GlobalEnvConfig{
					{
						Var:          "FOO",
						Value:        "my-foo-bar",
						ValueCommand: "/foo/bar/baz",
					},
				},
			},
		},
		want: `exactly one of value or valueCommand must be specified for env var FOO in global config`,
	},
	{
		name: "Global Config Empty Mount Def Name",
		config: HomelabConfig{
			Global: GlobalConfig{
				MountDefs: []MountConfig{
					{
						Type:     "bind",
						Src:      "/foo",
						Dst:      "/bar",
						ReadOnly: true,
					},
				},
			},
		},
		want: `mount name is empty in global config mount defs`,
	},
	{
		name: "Global Config Duplicate Mount Defs",
		config: HomelabConfig{
			Global: GlobalConfig{
				MountDefs: []MountConfig{
					{
						Name: "mount-foo1",
						Type: "bind",
						Src:  "/foo1",
						Dst:  "/bar1",
					},
					{
						Name: "mount-foo2",
						Type: "bind",
						Src:  "/foo2",
						Dst:  "/bar2",
					},
					{
						Name: "mount-foo1",
						Type: "bind",
						Src:  "/foo3",
						Dst:  "/bar3",
					},
				},
			},
		},
		want: `mount name mount-foo1 defined more than once in global config mount defs`,
	},
	{
		name: "Global Config Mount Def With Invalid Mount Type",
		config: HomelabConfig{
			Global: GlobalConfig{
				MountDefs: []MountConfig{
					{
						Name: "foo",
						Type: "garbage",
						Src:  "/foo",
						Dst:  "/bar",
					},
				},
			},
		},
		want: `unsupported mount type garbage for mount foo in global config mount defs`,
	},
	{
		name: "Global Config Mount Def With Empty Src",
		config: HomelabConfig{
			Global: GlobalConfig{
				MountDefs: []MountConfig{
					{
						Name: "foo",
						Type: "bind",
						Dst:  "/bar",
					},
				},
			},
		},
		want: `mount name foo has empty value for src in global config mount defs`,
	},
	{
		name: "Global Config Mount Def With Empty Dst",
		config: HomelabConfig{
			Global: GlobalConfig{
				MountDefs: []MountConfig{
					{
						Name: "foo",
						Type: "bind",
						Src:  "/foo",
					},
				},
			},
		},
		want: `mount name foo has empty value for dst in global config mount defs`,
	},
	{
		name: "Global Config Bind Mount Def With Options",
		config: HomelabConfig{
			Global: GlobalConfig{
				MountDefs: []MountConfig{
					{
						Name:    "foo",
						Type:    "bind",
						Src:     "/foo",
						Dst:     "/bar",
						Options: "dummy-option1=val1,dummy-option2=val2",
					},
				},
			},
		},
		want: `mount name foo specifies options in global config mount defs, that are not supported when mount type is bind`,
	},
	{
		name: "Empty Bridge Mode Network Name",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
							Priority:          1,
						},
					},
				},
			},
		},
		want: `network name cannot be empty`,
	},
	{
		name: "Duplicate Bridge Mode Network",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
							Priority:          1,
						},
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1-2",
							CIDR:              "172.18.101.0/24",
							Priority:          1,
						},
					},
				},
			},
		},
		want: `network net1 defined more than once in the IPAM config`,
	},
	{
		name: "Empty Host Interface Name",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:     "net1",
							CIDR:     "172.18.100.0/24",
							Priority: 1,
						},
					},
				},
			},
		},
		want: `host interface name of network net1 cannot be empty`,
	},
	{
		name: "Duplicate network host interface names",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
							Priority:          1,
						},
						{
							Name:              "net2",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.101.0/24",
							Priority:          1,
						},
					},
				},
			},
		},
		want: `host interface name docker-net1 of network net2 is already used by another network in the IPAM config`,
	},
	{
		name: "Empty Bridge Mode Network Priority",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
						},
					},
				},
			},
		},
		want: `network net1 has a non-positive priority 0`,
	},
	{
		name: "Invalid CIDR - Empty",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							Priority:          1,
						},
					},
				},
			},
		},
		want: `CIDR  of network net1 is invalid, reason: netip\.ParsePrefix\(""\): no '/'`,
	},
	{
		name: "Invalid CIDR - Unparsable",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "garbage-cidr",
							Priority:          1,
						},
					},
				},
			},
		},
		want: `CIDR garbage-cidr of network net1 is invalid, reason: netip\.ParsePrefix\("garbage-cidr"\): no '/'`,
	},
	{
		name: "Invalid CIDR - Missing /",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.16",
							Priority:          1,
						},
					},
				},
			},
		},
		want: `CIDR 172\.18\.100\.16 of network net1 is invalid, reason: netip\.ParsePrefix\("172\.18\.100\.16"\): no '/'`,
	},
	{
		name: "Invalid CIDR - Wrong Prefix Length",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/33",
							Priority:          1,
						},
					},
				},
			},
		},
		want: `CIDR 172\.18\.100\.0/33 of network net1 is invalid, reason: netip\.ParsePrefix\("172\.18\.100\.0/33"\): prefix length out of range`,
	},
	{
		name: "Invalid CIDR - IPv6",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "2002::1234:abcd:ffff:c0a8:101/64",
							Priority:          1,
						},
					},
				},
			},
		},
		want: `CIDR 2002::1234:abcd:ffff:c0a8:101/64 of network net1 is not an IPv4 subnet CIDR`,
	},
	{
		name: "Invalid CIDR - Octets Out Of Range",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.260.0/24",
							Priority:          1,
						},
					},
				},
			},
		},
		want: `CIDR 172\.18\.260\.0/24 of network net1 is invalid, reason: netip\.ParsePrefix\("172\.18\.260\.0/24"\): ParseAddr\("172\.18\.260\.0"\): IPv4 field has value >255`,
	},
	{
		name: "Invalid CIDR - Not A Network Address",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.1/25",
							Priority:          1,
						},
					},
				},
			},
		},
		want: `CIDR 172\.18\.100\.1/25 of network net1 is not the same as the network address 172\.18\.100\.0/25`,
	},
	{
		name: "Invalid CIDR - Long Prefix 31",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/31",
							Priority:          1,
						},
					},
				},
			},
		},
		want: `CIDR 172\.18\.100\.0/31 of network net1 \(prefix length: 31\) has a prefix length more than 30 which makes the network unusable for container IP address allocations`,
	},
	{
		name: "Invalid CIDR - Long Prefix 32",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/32",
							Priority:          1,
						},
					},
				},
			},
		},
		want: `CIDR 172\.18\.100\.0/32 of network net1 \(prefix length: 32\) has a prefix length more than 30 which makes the network unusable for container IP address allocations`,
	},
	{
		name: "Non-RFC1918 CIDR - Public IPv4",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "11.12.13.0/24",
							Priority:          1,
						},
					},
				},
			},
		},
		want: `CIDR 11\.12\.13\.0/24 of network net1 is not within the RFC1918 private address space`,
	},
	{
		name: "Non-RFC1918 CIDR - Link Local",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "169.254.10.0/24",
							Priority:          1,
						},
					},
				},
			},
		},
		want: `CIDR 169\.254\.10\.0/24 of network net1 is not within the RFC1918 private address space`,
	},
	{
		name: "Non-RFC1918 CIDR - Multicast",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "224.0.0.0/26",
							Priority:          1,
						},
					},
				},
			},
		},
		want: `CIDR 224\.0\.0\.0/26 of network net1 is not within the RFC1918 private address space`,
	},
	{
		name: "Overlapping CIDR",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
							Priority:          1,
						},
						{
							Name:              "net2",
							HostInterfaceName: "docker-net2",
							CIDR:              "172.18.0.0/16",
							Priority:          1,
						},
					},
				},
			},
		},
		want: `CIDR 172\.18\.0\.0/16 of network net2 overlaps with CIDR 172\.18\.100\.0/24 of network net1`,
	},
	{
		name: "Bridge Mode Network Invalid Container Reference - Empty Group",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
							Priority:          1,
							Containers: []ContainerIPConfig{
								{
									IP: "172.18.100.11",
									Container: ContainerReference{
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container IP config within network net1 has invalid container reference, reason: container reference cannot have an empty group name`,
	},
	{
		name: "Bridge Mode Network Invalid Container Reference - Empty Container",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
							Priority:          1,
							Containers: []ContainerIPConfig{
								{
									IP: "172.18.100.11",
									Container: ContainerReference{
										Group: "g1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container IP config within network net1 has invalid container reference, reason: container reference cannot have an empty container name`,
	},
	{
		name: "Invalid Container IP - Unparsable",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
							Priority:          1,
							Containers: []ContainerIPConfig{
								{
									IP: "garbage-ip",
									Container: ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 has invalid IP garbage-ip, reason: ParseAddr\("garbage-ip"\): unable to parse IP`,
	},
	{
		name: "Invalid Container IP - Too Short",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
							Priority:          1,
							Containers: []ContainerIPConfig{
								{
									IP: "172.18.100",
									Container: ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 has invalid IP 172\.18\.100, reason: ParseAddr\("172\.18\.100"\): IPv4 address too short`,
	},
	{
		name: "Invalid Container IP - Too Long",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
							Priority:          1,
							Containers: []ContainerIPConfig{
								{
									IP: "172.18.100.1.2.3.4",
									Container: ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 has invalid IP 172\.18\.100\.1\.2\.3\.4, reason: ParseAddr\("172\.18\.100\.1\.2\.3\.4"\): IPv4 address too long`,
	},
	{
		name: "Container IP Not Within Network CIDR",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
							Priority:          1,
							Containers: []ContainerIPConfig{
								{
									IP: "172.18.101.2",
									Container: ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 has IP 172\.18\.101\.2 that does not belong to the network CIDR 172\.18\.100\.0/24`,
	},
	{
		name: "Container IP same as Network Address",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
							Priority:          1,
							Containers: []ContainerIPConfig{
								{
									IP: "172.18.100.0",
									Container: ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 has IP 172\.18\.100\.0 matching the network address 172\.18\.100\.0`,
	},
	{
		name: "Container IP same as Gateway Address",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
							Priority:          1,
							Containers: []ContainerIPConfig{
								{
									IP: "172.18.100.1",
									Container: ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} endpoint in network net1 has IP 172\.18\.100\.1 matching the gateway address 172\.18\.100\.1`,
	},
	{
		name: "Multiple Endpoints For Same Container Within A Bridge Mode Network",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
							Priority:          1,
							Containers: []ContainerIPConfig{
								{
									IP: "172.18.100.2",
									Container: ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
								{
									IP: "172.18.100.3",
									Container: ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct1} has multiple endpoints in network net1`,
	},
	{
		name: "Duplicate Container IPs",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
							Priority:          1,
							Containers: []ContainerIPConfig{
								{
									IP: "172.18.100.2",
									Container: ContainerReference{
										Group:     "group1",
										Container: "ct1",
									},
								},
								{
									IP: "172.18.100.3",
									Container: ContainerReference{
										Group:     "group1",
										Container: "ct2",
									},
								},
								{
									IP: "172.18.100.4",
									Container: ContainerReference{
										Group:     "group1",
										Container: "ct3",
									},
								},
								{
									IP: "172.18.100.2",
									Container: ContainerReference{
										Group:     "group1",
										Container: "ct4",
									},
								},
							},
						},
					},
				},
			},
		},
		want: `IP 172\.18\.100\.2 of container {Group:group1 Container:ct4} is already in use by another container in network net1`,
	},
	{
		name: "Empty Container Mode Network Name",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					ContainerModeNetworks: []ContainerModeNetworkConfig{
						{
							Priority: 1,
						},
					},
				},
			},
		},
		want: `network name cannot be empty`,
	},
	{
		name: "Duplicate container mode network",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					ContainerModeNetworks: []ContainerModeNetworkConfig{
						{
							Name:     "net1",
							Priority: 1,
						},
						{
							Name:     "net1",
							Priority: 2,
						},
					},
				},
			},
		},
		want: `network net1 defined more than once in the IPAM config`,
	},
	{
		name: "Duplicate bridge/container mode networks",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					BridgeModeNetworks: []BridgeModeNetworkConfig{
						{
							Name:              "net1",
							HostInterfaceName: "docker-net1",
							CIDR:              "172.18.100.0/24",
							Priority:          1,
						},
					},
					ContainerModeNetworks: []ContainerModeNetworkConfig{
						{
							Name:     "net1",
							Priority: 2,
						},
					},
				},
			},
		},
		want: `network net1 defined more than once in the IPAM config`,
	},
	{
		name: "Empty Container Mode Network Priority",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					ContainerModeNetworks: []ContainerModeNetworkConfig{
						{
							Name: "net1",
						},
					},
				},
			},
		},
		want: `network net1 has a non-positive priority 0`,
	},
	{
		name: "Container Mode Network Invalid Container Reference - Empty Group",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					ContainerModeNetworks: []ContainerModeNetworkConfig{
						{
							Name:     "net1",
							Priority: 1,
							Containers: []ContainerReference{
								{
									Container: "ct1",
								},
							},
						},
					},
				},
			},
		},
		want: `container IP config within network net1 has invalid container reference, reason: container reference cannot have an empty group name`,
	},
	{
		name: "Container Mode Network Invalid Container Reference - Empty Container",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					ContainerModeNetworks: []ContainerModeNetworkConfig{
						{
							Name:     "net1",
							Priority: 1,
							Containers: []ContainerReference{
								{
									Group: "g1",
								},
							},
						},
					},
				},
			},
		},
		want: `container IP config within network net1 has invalid container reference, reason: container reference cannot have an empty container name`,
	},

	{
		name: "Multiple Container Mode Network Stacks For Same Container",
		config: HomelabConfig{
			IPAM: IPAMConfig{
				Networks: NetworksConfig{
					ContainerModeNetworks: []ContainerModeNetworkConfig{
						{
							Name:     "net1",
							Priority: 1,
							Containers: []ContainerReference{
								{
									Group:     "group1",
									Container: "ct1",
								},
								{
									Group:     "group1",
									Container: "ct2",
								},
								{
									Group:     "group2",
									Container: "ct3",
								},
								{
									Group:     "group1",
									Container: "ct2",
								},
							},
						},
					},
				},
			},
		},
		want: `container {Group:group1 Container:ct2} is connected to multiple container mode network stacks`,
	},
	{
		name: "Empty Host Name In Hosts Config",
		config: HomelabConfig{
			Hosts: []HostConfig{
				{
					AllowedContainers: []ContainerReference{
						{
							Group:     "g1",
							Container: "ct1",
						},
					},
				},
			},
		},
		want: `host name cannot be empty in the hosts config`,
	},
	{
		name: "Duplicate Host Name In Hosts Config",
		config: HomelabConfig{
			Hosts: []HostConfig{
				{
					Name: "h1",
					AllowedContainers: []ContainerReference{
						{
							Group:     "g1",
							Container: "ct1",
						},
						{
							Group:     "g2",
							Container: "ct2",
						},
					},
				},
				{
					Name: "h1",
					AllowedContainers: []ContainerReference{
						{
							Group:     "g3",
							Container: "ct3",
						},
						{
							Group:     "g4",
							Container: "ct4",
						},
					},
				},
			},
		},
		want: `host h1 defined more than once in the hosts config`,
	},
	{
		name: "Invalid Container Reference Within Host Config - Empty Group",
		config: HomelabConfig{
			Hosts: []HostConfig{
				{
					Name: "h1",
					AllowedContainers: []ContainerReference{
						{
							Container: "ct1",
						},
					},
				},
			},
		},
		want: `allowed container config within host h1 has invalid container reference, reason: container reference cannot have an empty group name`,
	},
	{
		name: "Invalid Container Reference Within Host Config - Empty Container",
		config: HomelabConfig{
			Hosts: []HostConfig{
				{
					Name: "h1",
					AllowedContainers: []ContainerReference{
						{
							Group: "g1",
						},
					},
				},
			},
		},
		want: `allowed container config within host h1 has invalid container reference, reason: container reference cannot have an empty container name`,
	},
	{
		name: "Duplicate Container Within Host Config",
		config: HomelabConfig{
			Hosts: []HostConfig{
				{
					Name: "h1",
					AllowedContainers: []ContainerReference{
						{
							Group:     "g1",
							Container: "ct1",
						},
						{
							Group:     "g2",
							Container: "ct2",
						},
						{
							Group:     "g3",
							Container: "ct3",
						},
						{
							Group:     "g2",
							Container: "ct2",
						},
					},
				},
			},
		},
		want: `container {Group:g2 Container:ct2} defined more than once in the hosts config for host h1`,
	},
	{
		name: "Empty Group Name In Groups Config",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Order: 1,
				},
			},
		},
		want: `group name cannot be empty in the groups config`,
	},
	{
		name: "Duplicate Container Group Name",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
				{
					Name:  "g2",
					Order: 1,
				},
				{
					Name:  "g3",
					Order: 1,
				},
				{
					Name:  "g1",
					Order: 2,
				},
			},
		},
		want: `group g1 defined more than once in the groups config`,
	},
	{
		name: "Container Group Without Order",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name: "g1",
				},
			},
		},
		want: `group g1 has a non-positive order 0`,
	},
	{
		name: "Container Group With Zero Order",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 0,
				},
			},
		},
		want: `group g1 has a non-positive order 0`,
	},
	{
		name: "Container Group With Negative Order",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: -1,
				},
			},
		},
		want: `group g1 has a non-positive order -1`,
	},
}

func TestValidateConfigErrors(t *testing.T) {
	for _, test := range validateConfigErrorTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotErr := tc.config.validate()
			if gotErr == nil {
				t.Errorf(
					"HomelabConfig.validate()\nTest Case: %q\nFailure: gotErr == nil\nReason: want = %q",
					tc.name, tc.want)
				return
			}

			match, err := regexp.MatchString(tc.want, gotErr.Error())
			if err != nil {
				t.Errorf(
					"HomelabConfig.validate()\nTest Case: %q\nFailure: unexpected exception while matching against gotErr error string\nReason: error = %v", tc.name, err)
				return
			}

			if !match {
				t.Errorf(
					"HomelabConfig.validate()\nTest Case: %q\nFailure: gotErr did not match the want regex\nReason:\n\ngotErr = %q\n\twant = %q", tc.name, gotErr, tc.want)
			}
		})
	}
}
