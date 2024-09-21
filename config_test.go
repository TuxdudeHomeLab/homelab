package main

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	dcontainer "github.com/docker/docker/api/types/container"
	dnetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/google/go-cmp/cmp"
)

var (
	// TODO: Remove this after we start using fakeConfigEnv.
	_ = fakeConfigEnv
)

var parseAndValidateConfigUsingReaderTests = []struct {
	name              string
	config            string
	want              *HomelabConfig
	wantDockerConfigs containerDockerConfigMap
}{
	{
		name: "Valid extensive config",
		config: `
global:
  env:
    - var: MY_CONFIG_VAR_1
      value: MY_CONFIG_VAR_1_VALUE
    - var: MY_CONFIG_VAR_2
      value: MY_CONFIG_VAR_2_VALUE
    - var: MY_CONFIG_VAR_3
      valueCommand: /foo/bar/some-env-var-cmd
  mountDefs:
    - name: mount-def-1
      type: bind
      src: /abc/def/ghi
      dst: /pqr/stu/vwx
      readOnly: true
    - name: mount-def-2
      type: bind
      src: /abc1/def1
      dst: /pqr2/stu2/vwx2
    - name: homelab-self-signed-tls-cert
      type: bind
      src: /path/to/my/self/signed/cert/on/host
      dst: /path/to/my/self/signed/cert/on/container
  container:
    stopSignal: SIGTERM
    stopTimeout: 8
    restartPolicy:
      mode: unless-stopped
    domainName: example.tld
    dnsSearch:
      - dns-search-1
      - dns-search-2
    env:
      - var: MY_CONTAINER_ENV_VAR_1
        value: MY_CONTAINER_ENV_VAR_1_VALUE
      - var: MY_CONTAINER_ENV_VAR_2
        value: MY_CONTAINER_ENV_VAR_2_VALUE
      - var: MY_CONTAINER_ENV_VAR_3
        valueCommand: /foo2/bar2/some-other-env-var-cmd
    mounts:
      - name: mount-def-1
      - name: mount-def-2
      - name: mount-def-3
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
  - name: fakehost
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
      skipImagePull: false
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
      restartPolicy:
        mode: always
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
          disallowRead: false
          disallowWrite: true
          disallowMknod: true
        - src: /dev/dev1
          dst: /dev/dev2
        - src: /dev/foo2
          dst: /dev/bar2
          disallowRead: true
          disallowWrite: true
          disallowMknod: false
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
          hostIp: 127.0.0.1
          hostPort: 53
        - containerPort: 53
          proto: udp
          hostIp: 127.0.0.1
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
    health:
      cmd:
        - my-health-cmd
        - health-arg-1
        - health-arg-2
      retries: 3
      interval: 60s
      timeout: 10s
      startPeriod: 3m
      startInterval: 10s
    runtime:
      tty: true
      shmSize: 1g
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
		want: &HomelabConfig{
			Global: GlobalConfig{
				Env: []ConfigEnv{
					{
						Var:   "MY_CONFIG_VAR_1",
						Value: "MY_CONFIG_VAR_1_VALUE",
					},
					{
						Var:   "MY_CONFIG_VAR_2",
						Value: "MY_CONFIG_VAR_2_VALUE",
					},
					{
						Var:          "MY_CONFIG_VAR_3",
						ValueCommand: "/foo/bar/some-env-var-cmd",
					},
				},
				MountDefs: []MountConfig{
					{
						Name:     "mount-def-1",
						Type:     "bind",
						Src:      "/abc/def/ghi",
						Dst:      "/pqr/stu/vwx",
						ReadOnly: true,
					},
					{
						Name: "mount-def-2",
						Type: "bind",
						Src:  "/abc1/def1",
						Dst:  "/pqr2/stu2/vwx2",
					},
					{
						Name: "homelab-self-signed-tls-cert",
						Type: "bind",
						Src:  "/path/to/my/self/signed/cert/on/host",
						Dst:  "/path/to/my/self/signed/cert/on/container",
					},
				},
				Container: GlobalContainerConfig{
					StopSignal:  "SIGTERM",
					StopTimeout: 8,
					RestartPolicy: ContainerRestartPolicyConfig{
						Mode: "unless-stopped",
					},
					DomainName: "example.tld",
					DNSSearch: []string{
						"dns-search-1",
						"dns-search-2",
					},
					Env: []ContainerEnv{
						{
							Var:   "MY_CONTAINER_ENV_VAR_1",
							Value: "MY_CONTAINER_ENV_VAR_1_VALUE",
						},
						{
							Var:   "MY_CONTAINER_ENV_VAR_2",
							Value: "MY_CONTAINER_ENV_VAR_2_VALUE",
						},
						{
							Var:          "MY_CONTAINER_ENV_VAR_3",
							ValueCommand: "/foo2/bar2/some-other-env-var-cmd",
						},
					},
					Mounts: []MountConfig{
						{
							Name: "mount-def-1",
						},
						{
							Name: "mount-def-2",
						},
						{
							Name:     "mount-def-3",
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
					Name: fakeHostName,
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
						SkipImagePull:           false,
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
						Order:        1,
						StartPreHook: "$$SCRIPTS_DIR$$/my-start-prehook.sh",
						RestartPolicy: ContainerRestartPolicyConfig{
							Mode: "always",
						},
						AutoRemove:  true,
						StopSignal:  "SIGHUP",
						StopTimeout: 10,
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
								Src:           "/dev/foo",
								Dst:           "/dev/bar",
								DisallowWrite: true,
								DisallowMknod: true,
							},
							{
								Src: "/dev/dev1",
								Dst: "/dev/dev2",
							},
							{
								Src:           "/dev/foo2",
								Dst:           "/dev/bar2",
								DisallowRead:  true,
								DisallowWrite: true,
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
								Protocol:      "tcp",
								HostIP:        "127.0.0.1",
								HostPort:      53,
							},
							{
								ContainerPort: 53,
								Protocol:      "udp",
								HostIP:        "127.0.0.1",
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
					Health: ContainerHealthConfig{
						Cmd: []string{
							"my-health-cmd",
							"health-arg-1",
							"health-arg-2",
						},
						Retries:       3,
						Interval:      "60s",
						Timeout:       "10s",
						StartPeriod:   "3m",
						StartInterval: "10s",
					},
					Runtime: ContainerRuntimeConfig{
						AttachToTty: true,
						ShmSize:     "1g",
						Env: []ContainerEnv{
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
		wantDockerConfigs: containerDockerConfigMap{
			ContainerReference{
				Group:     "group1",
				Container: "ct1",
			}: &containerDockerConfigs{
				ContainerConfig: &dcontainer.Config{
					Hostname:   "foobar",
					Domainname: "somedomain",
					User:       "$$CURRENT_USER$$:$$CURRENT_GROUP$$",
					ExposedPorts: nat.PortSet{
						"53/tcp": struct{}{},
						"53/udp": struct{}{},
					},
					Tty: true,
					Env: []string{
						"MY_CONTAINER_ENV_VAR_1=MY_CONTAINER_ENV_VAR_1_VALUE",
						"MY_CONTAINER_ENV_VAR_2=MY_CONTAINER_ENV_VAR_2_VALUE",
						"MY_CONTAINER_ENV_VAR_3=",
						"MY_ENV=MY_ENV_VALUE",
						"MY_ENV_2=",
						"MY_ENV_3=SomeHostName.$$HUMAN_FRIENDLY_HOST_NAME$$.SomeDomainName",
					},
					Cmd: []string{
						"foo",
						"bar",
						"baz",
					},
					Image: "tuxdude/homelab-base:master",
					Entrypoint: []string{
						"my-custom-entrypoint",
						"ep-arg1",
						"ep-arg2",
					},
					Labels: map[string]string{
						"my.ct1.label.name.1": "my.ct1.label.value.1",
						"my.ct1.label.name.2": "my.ct1.label.value.2",
					},
					StopSignal:  "SIGHUP",
					StopTimeout: newInt(10),
				},
				HostConfig: &dcontainer.HostConfig{
					Binds: []string{
						"/abc/def/ghi:/pqr/stu/vwx:ro",
						"/abc1/def1:/pqr2/stu2/vwx2",
						"/path/to/my/self/signed/cert/on/host:/path/to/my/self/signed/cert/on/container",
						"/foo:/bar:ro",
						"$$CONFIG_DIR$$/generated/config.yml:/data/blocky/config/config.yml:ro",
						":/tmp/cache",
					},
					NetworkMode: "group1-bridge",
					PortBindings: nat.PortMap{
						"53/tcp": []nat.PortBinding{
							{
								HostIP:   "127.0.0.1",
								HostPort: "53",
							},
						},
						"53/udp": []nat.PortBinding{
							{
								HostIP:   "127.0.0.1",
								HostPort: "53",
							},
						},
					},
					RestartPolicy: dcontainer.RestartPolicy{
						Name: "always",
					},
					AutoRemove: true,
					CapAdd: []string{
						"SYS_RAWIO",
						"SYS_ADMIN",
					},
					CapDrop: []string{
						"NET_ADMIN",
						"SYS_MODULE",
					},
					DNS: []string{
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
					GroupAdd: []string{
						"dialout",
						"someRandomGroup",
					},
					Privileged:     true,
					ReadonlyRootfs: true,
					Sysctls: map[string]string{
						"net.ipv4.conf.all.src_valid_mark": "1",
						"net.ipv4.ip_forward":              "1",
					},
				},
				NetworkConfig: &dnetwork.NetworkingConfig{
					EndpointsConfig: map[string]*dnetwork.EndpointSettings{
						"group1-bridge": {
							IPAMConfig: &dnetwork.EndpointIPAMConfig{
								IPv4Address: "172.18.18.11",
							},
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
    order: 3
  - name: group3
    order: 2`,
		want: &HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "group1",
					Order: 1,
				},
				{
					Name:  "group2",
					Order: 3,
				},
				{
					Name:  "group3",
					Order: 2,
				},
			},
		},
		wantDockerConfigs: containerDockerConfigMap{},
	},
}

func TestParseConfigUsingReader(t *testing.T) {
	for _, test := range parseAndValidateConfigUsingReaderTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := strings.NewReader(tc.config)
			got, gotErr := buildDeploymentFromReader(input, fakeHostInfo)
			if gotErr != nil {
				t.Errorf(
					"buildDeploymentFromReader()\nTest Case: %q\nFailure: gotErr != nil\nReason: %v",
					tc.name, gotErr)
				return
			}

			if diff := cmp.Diff(tc.want, got.config); diff != "" {
				t.Errorf(
					"buildDeploymentFromReader()\nTest Case: %q\nFailure: got did not match the want config\nDiff(-want +got): %s", tc.name, diff)
			}

			if diff := cmp.Diff(tc.wantDockerConfigs, got.containerDockerConfigs); diff != "" {
				t.Errorf(
					"buildDeploymentFromReader()\nTest Case: %q\nFailure: docker configs got did not match the want config\nDiff(-want +got): %s", tc.name, diff)
			}
		})
	}
}

var validParseAndValidateConfigsFromPathTests = []struct {
	name              string
	configsPath       string
	want              *HomelabConfig
	wantDockerConfigs containerDockerConfigMap
}{
	{
		name:        "Valid multi file config",
		configsPath: "parse-configs-valid",
		want: &HomelabConfig{
			Global: GlobalConfig{
				Env: []ConfigEnv{
					{
						Var:   "MY_GLOBAL_FOO",
						Value: "MY_GLOBAL_BAR",
					},
				},
				Container: GlobalContainerConfig{
					StopSignal:  "SIGTERM",
					StopTimeout: 5,
					RestartPolicy: ContainerRestartPolicyConfig{
						Mode:          "on-failure",
						MaxRetryCount: 5,
					},
					DomainName: "somedomain",
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
					Name: fakeHostName,
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
		wantDockerConfigs: containerDockerConfigMap{
			ContainerReference{
				Group:     "g1",
				Container: "c1",
			}: &containerDockerConfigs{
				ContainerConfig: &dcontainer.Config{
					Domainname:  "somedomain",
					Image:       "abc/xyz",
					StopTimeout: newInt(5),
				},
				HostConfig: &dcontainer.HostConfig{
					NetworkMode: "net1",
					RestartPolicy: dcontainer.RestartPolicy{
						Name:              "on-failure",
						MaximumRetryCount: 5,
					},
				},
				NetworkConfig: &dnetwork.NetworkingConfig{
					EndpointsConfig: map[string]*dnetwork.EndpointSettings{
						"net1": {
							IPAMConfig: &dnetwork.EndpointIPAMConfig{
								IPv4Address: "172.18.100.11",
							},
						},
					},
				},
			},
			ContainerReference{
				Group:     "g1",
				Container: "c2",
			}: &containerDockerConfigs{
				ContainerConfig: &dcontainer.Config{
					Domainname:  "somedomain",
					Image:       "abc/xyz2",
					StopTimeout: newInt(5),
				},
				HostConfig: &dcontainer.HostConfig{
					NetworkMode: "net1",
					RestartPolicy: dcontainer.RestartPolicy{
						Name:              "on-failure",
						MaximumRetryCount: 5,
					},
				},
				NetworkConfig: &dnetwork.NetworkingConfig{
					EndpointsConfig: map[string]*dnetwork.EndpointSettings{
						"net1": {
							IPAMConfig: &dnetwork.EndpointIPAMConfig{
								IPv4Address: "172.18.100.12",
							},
						},
					},
				},
			},
			ContainerReference{
				Group:     "g2",
				Container: "c3",
			}: &containerDockerConfigs{
				ContainerConfig: &dcontainer.Config{
					Domainname:  "somedomain",
					Image:       "abc/xyz3",
					StopTimeout: newInt(5),
				},
				HostConfig: &dcontainer.HostConfig{
					NetworkMode: "net2",
					RestartPolicy: dcontainer.RestartPolicy{
						Name:              "on-failure",
						MaximumRetryCount: 5,
					},
				},
				NetworkConfig: &dnetwork.NetworkingConfig{
					EndpointsConfig: map[string]*dnetwork.EndpointSettings{
						"net2": {
							IPAMConfig: &dnetwork.EndpointIPAMConfig{
								IPv4Address: "172.18.101.21",
							},
						},
					},
				},
			},
			ContainerReference{
				Group:     "g3",
				Container: "c4",
			}: &containerDockerConfigs{
				ContainerConfig: &dcontainer.Config{
					Domainname:  "somedomain",
					Image:       "abc/xyz4",
					StopTimeout: newInt(5),
				},
				HostConfig: &dcontainer.HostConfig{
					NetworkMode: "none",
					RestartPolicy: dcontainer.RestartPolicy{
						Name:              "on-failure",
						MaximumRetryCount: 5,
					},
				},
			},
		},
	},
}

func TestParseAndValidateConfigsFromPath(t *testing.T) {
	for _, test := range validParseAndValidateConfigsFromPathTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := fmt.Sprintf("%s/testdata/%s", pwd(), tc.configsPath)
			got, gotErr := buildDeploymentFromConfigsPath(p, fakeHostInfo)
			if gotErr != nil {
				t.Errorf(
					"buildDeploymentFromConfigsPath()\nTest Case: %q\nFailure: gotErr != nil\nReason: %v",
					tc.name, gotErr)
				return
			}

			if diff := cmp.Diff(tc.want, got.config); diff != "" {
				t.Errorf(
					"buildDeploymentFromConfigsPath()\nTest Case: %q\nFailure: got did not match the want config\nDiff(-want +got): %s", tc.name, diff)
			}

			if diff := cmp.Diff(tc.wantDockerConfigs, got.containerDockerConfigs); diff != "" {
				t.Errorf(
					"buildDeploymentFromConfigsPath()\nTest Case: %q\nFailure: docker configs got did not match the want config\nDiff(-want +got): %s", tc.name, diff)
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

			match, err := regexp.MatchString(fmt.Sprintf("^%s$", tc.want), gotErr.Error())
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

			_, gotErr := buildDeploymentFromConfig(&tc.config, fakeHostInfo)
			if gotErr != nil {
				t.Errorf(
					"buildDeploymentFromConfig()\nTest Case: %q\nFailure: gotErr != nil\nReason: %v",
					tc.name, gotErr)
				return
			}

			// TODO: Validate the built deployment.
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
				Env: []ConfigEnv{
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
				Env: []ConfigEnv{
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
				Env: []ConfigEnv{
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
				Env: []ConfigEnv{
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
		want: `mount name cannot be empty in global config mount defs`,
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
		want: `mount name foo cannot have an empty value for src in global config mount defs`,
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
		want: `mount name foo cannot have an empty value for dst in global config mount defs`,
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
		want: `bind mount name foo cannot specify options in global config mount defs`,
	},
	{
		name: "Global Container Config Negative Stop Timeout",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					StopTimeout: -1,
				},
			},
		},
		want: `container stop timeout -1 cannot be negative in global container config`,
	},
	{
		name: "Global Container Config Restart Policy MaxRetryCount Set With Non-On-Failure Mode",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					RestartPolicy: ContainerRestartPolicyConfig{
						Mode:          "always",
						MaxRetryCount: 5,
					},
				},
			},
		},
		want: `restart policy max retry count can be set only when the mode is on-failure in global container config`,
	},
	{
		name: "Global Container Config Invalid Restart Policy Mode",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					RestartPolicy: ContainerRestartPolicyConfig{
						Mode: "garbage",
					},
				},
			},
		},
		want: `invalid restart policy mode garbage in global container config, valid values are \[ 'no', 'always', 'on-failure', 'unless-stopped' \]`,
	},
	{
		name: "Global Container Config Negative Restart Policy MaxRetryCount",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					RestartPolicy: ContainerRestartPolicyConfig{
						Mode:          "on-failure",
						MaxRetryCount: -1,
					},
				},
			},
		},
		want: `restart policy max retry count -1 cannot be negative in global container config`,
	},
	{
		name: "Empty Global Container Config Env Var",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					Env: []ContainerEnv{
						{
							Value: "foo-bar",
						},
					},
				},
			},
		},
		want: `empty env var in global container config`,
	},
	{
		name: "Duplicate Global Container Config Env Var",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					Env: []ContainerEnv{
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
		},
		want: `env var FOO specified more than once in global container config`,
	},
	{
		name: "Global Container Config Env Var Without Value And ValueCommand",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					Env: []ContainerEnv{
						{
							Var: "FOO",
						},
					},
				},
			},
		},
		want: `neither value nor valueCommand specified for env var FOO in global container config`,
	},
	{
		name: "Global Container Config Env Var With Both Value And ValueCommand",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					Env: []ContainerEnv{
						{
							Var:          "FOO",
							Value:        "my-foo-bar",
							ValueCommand: "/foo/bar/baz",
						},
					},
				},
			},
		},
		want: `exactly one of value or valueCommand must be specified for env var FOO in global container config`,
	},
	{
		name: "Global Container Config Empty Mount Name",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					Mounts: []MountConfig{
						{
							Type:     "bind",
							Src:      "/foo",
							Dst:      "/bar",
							ReadOnly: true,
						},
					},
				},
			},
		},
		want: `mount name cannot be empty in global container config mounts`,
	},
	{
		name: "Global Container Config Duplicate Mounts",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					Mounts: []MountConfig{
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
		},
		want: `mount name mount-foo1 defined more than once in global container config mounts`,
	},
	{
		name: "Global Container Config Mount With Invalid Mount Type",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					Mounts: []MountConfig{
						{
							Name: "foo",
							Type: "garbage",
							Src:  "/foo",
							Dst:  "/bar",
						},
					},
				},
			},
		},
		want: `unsupported mount type garbage for mount foo in global container config mounts`,
	},
	{
		name: "Global Container Config Mount With Empty Src",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					Mounts: []MountConfig{
						{
							Name: "foo",
							Type: "bind",
							Dst:  "/bar",
						},
					},
				},
			},
		},
		want: `mount name foo cannot have an empty value for src in global container config mounts`,
	},
	{
		name: "Global Container Config Mount With Empty Dst",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					Mounts: []MountConfig{
						{
							Name: "foo",
							Type: "bind",
							Src:  "/foo",
						},
					},
				},
			},
		},
		want: `mount name foo cannot have an empty value for dst in global container config mounts`,
	},
	{
		name: "Global Container Config Bind Mount With Options",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					Mounts: []MountConfig{
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
		},
		want: `bind mount name foo cannot specify options in global container config mounts`,
	},
	{
		name: "Global Container Config Mount Def Reference Not Found",
		config: HomelabConfig{
			Global: GlobalConfig{
				MountDefs: []MountConfig{
					{
						Name: "foo",
						Type: "bind",
						Src:  "/foo",
						Dst:  "/bar",
					},
				},
				Container: GlobalContainerConfig{
					Mounts: []MountConfig{
						{
							Name: "foo2",
						},
					},
				},
			},
		},
		want: `mount specified by just the name foo2 not found in defs in global container config mounts`,
	},
	{
		name: "Empty Global Container Config Label Name",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					Labels: []LabelConfig{
						{
							Value: "foo-bar",
						},
					},
				},
			},
		},
		want: `empty label name in global container config`,
	},
	{
		name: "Duplicate Global Container Config Label Name",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					Labels: []LabelConfig{
						{
							Name:  "FOO",
							Value: "foo-bar",
						},
						{
							Name:  "FOO2",
							Value: "foo-bar-2",
						},
						{
							Name:  "FOO",
							Value: "foo-bar-3",
						},
					},
				},
			},
		},
		want: `label name FOO specified more than once in global container config`,
	},
	{
		name: "Global Container Config Empty Label Value",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					Labels: []LabelConfig{
						{
							Name: "FOO",
						},
					},
				},
			},
		},
		want: `empty label value for label FOO in global container config`,
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
		want: `network net1 cannot have a non-positive priority 0`,
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
		want: `CIDR 172\.18\.100\.0/31 of network net1 \(prefix length: 31\) cannot have a prefix length more than 30 which makes the network unusable for container IP address allocations`,
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
		want: `CIDR 172\.18\.100\.0/32 of network net1 \(prefix length: 32\) cannot have a prefix length more than 30 which makes the network unusable for container IP address allocations`,
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
		want: `container {Group:group1 Container:ct1} endpoint in network net1 cannot have an IP 172\.18\.101\.2 that does not belong to the network CIDR 172\.18\.100\.0/24`,
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
		want: `container {Group:group1 Container:ct1} endpoint in network net1 cannot have an IP 172\.18\.100\.0 matching the network address 172\.18\.100\.0`,
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
		want: `container {Group:group1 Container:ct1} endpoint in network net1 cannot have an IP 172\.18\.100\.1 matching the gateway address 172\.18\.100\.1`,
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
		want: `container {Group:group1 Container:ct1} cannot have multiple endpoints in network net1`,
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
		want: `network net1 cannot have a non-positive priority 0`,
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
		want: `group g1 cannot have a non-positive order 0`,
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
		want: `group g1 cannot have a non-positive order 0`,
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
		want: `group g1 cannot have a non-positive order -1`,
	},
	{
		name: "Container Group Definition Missing",
		config: HomelabConfig{
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
				},
			},
		},
		want: `group definition missing in groups config for the container {Group:g1 Container:c1} in the containers config`,
	},
	{
		name: "Duplicate Container Definition",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
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
						Image: "foo/bar2:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
				},
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar3:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
				},
			},
		},
		want: `container {Group:g1 Container:c1} defined more than once in the containers config`,
	},
	{
		name: "Empty Container Config Env Var",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Config: ContainerConfigOptions{
						Env: []ConfigEnv{
							{
								Value: "foo-bar",
							},
						},
					},
				},
			},
		},
		want: `empty env var in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Duplicate Container Config Env Var",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Config: ContainerConfigOptions{
						Env: []ConfigEnv{
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
			},
		},
		want: `env var FOO specified more than once in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Env Var Without Value And ValueCommand",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Config: ContainerConfigOptions{
						Env: []ConfigEnv{
							{
								Var: "FOO",
							},
						},
					},
				},
			},
		},
		want: `neither value nor valueCommand specified for env var FOO in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Env Var With Both Value And ValueCommand",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Config: ContainerConfigOptions{
						Env: []ConfigEnv{
							{
								Var:          "FOO",
								Value:        "my-foo-bar",
								ValueCommand: "/foo/bar/baz",
							},
						},
					},
				},
			},
		},
		want: `exactly one of value or valueCommand must be specified for env var FOO in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Empty Container Config Image",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
				},
			},
		},
		want: `image cannot be empty in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config SkipImagePull And IgnoreImagePullFailures Both Set To True",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image:                   "foo/bar:123",
						SkipImagePull:           true,
						IgnoreImagePullFailures: true,
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
				},
			},
		},
		want: `ignoreImagePullFailures cannot be true when skipImagePull is true in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config SkipImagePull And PullImageBeforeStop Both Set To True",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image:               "foo/bar:123",
						SkipImagePull:       true,
						PullImageBeforeStop: true,
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
				},
			},
		},
		want: `pullImageBeforeStop cannot be true when skipImagePull is true in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Empty Container Config Label Name",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Metadata: ContainerMetadataConfig{
						Labels: []LabelConfig{
							{
								Value: "foo-bar",
							},
						},
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
				},
			},
		},
		want: `empty label name in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Duplicate Container Config Label Name",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Metadata: ContainerMetadataConfig{
						Labels: []LabelConfig{
							{
								Name:  "FOO",
								Value: "foo-bar",
							},
							{
								Name:  "FOO2",
								Value: "foo-bar-2",
							},
							{
								Name:  "FOO",
								Value: "foo-bar-3",
							},
						},
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
				},
			},
		},
		want: `label name FOO specified more than once in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Empty Label Value",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Metadata: ContainerMetadataConfig{
						Labels: []LabelConfig{
							{
								Name: "FOO",
							},
						},
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
				},
			},
		},
		want: `empty label value for label FOO in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Empty Order",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
				},
			},
		},
		want: `container order 0 cannot be non-positive in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Negative Order",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: -1,
					},
				},
			},
		},
		want: `container order -1 cannot be non-positive in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Restart Policy MaxRetryCount Set With Non-On-Failure Mode",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
						RestartPolicy: ContainerRestartPolicyConfig{
							Mode:          "always",
							MaxRetryCount: 5,
						},
					},
				},
			},
		},
		want: `restart policy max retry count can be set only when the mode is on-failure in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Invalid Restart Policy Mode",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
						RestartPolicy: ContainerRestartPolicyConfig{
							Mode: "garbage",
						},
					},
				},
			},
		},
		want: `invalid restart policy mode garbage in container {Group: g1 Container:c1} config, valid values are \[ 'no', 'always', 'on-failure', 'unless-stopped' \]`,
	},
	{
		name: "Container Config Negative Restart Policy MaxRetryCount",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
						RestartPolicy: ContainerRestartPolicyConfig{
							Mode:          "on-failure",
							MaxRetryCount: -1,
						},
					},
				},
			},
		},
		want: `restart policy max retry count -1 cannot be negative in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Negative StopTimeout",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order:       1,
						StopTimeout: -1,
					},
				},
			},
		},
		want: `container stop timeout -1 cannot be negative in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config PrimaryUserGroup Without User",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					User: ContainerUserConfig{
						PrimaryGroup: "my-user-group",
					},
				},
			},
		},
		want: `container user primary group cannot be set without setting the user in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Device Missing Src And Dst",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Filesystem: ContainerFilesystemConfig{
						Devices: []DeviceConfig{
							{},
						},
					},
				},
			},
		},
		want: `device src cannot be empty in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Device Missing Src",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Filesystem: ContainerFilesystemConfig{
						Devices: []DeviceConfig{
							{
								Dst: "/dev/my-target-dev",
							},
						},
					},
				},
			},
		},
		want: `device src cannot be empty in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Empty Mount Name",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Filesystem: ContainerFilesystemConfig{
						Mounts: []MountConfig{
							{
								Type:     "bind",
								Src:      "/foo",
								Dst:      "/bar",
								ReadOnly: true,
							},
						},
					},
				},
			},
		},
		want: `mount name cannot be empty in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Duplicate Mounts Within Container Config",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Filesystem: ContainerFilesystemConfig{
						Mounts: []MountConfig{
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
			},
		},
		want: `mount name mount-foo1 defined more than once in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Duplicate Mounts Within Container And Global Configs Combined",
		config: HomelabConfig{
			Global: GlobalConfig{
				Container: GlobalContainerConfig{
					Mounts: []MountConfig{
						{
							Name: "mount-foo2",
							Type: "bind",
							Src:  "/foo2-global",
							Dst:  "/bar2-global",
						},
						{
							Name: "mount-foo3",
							Type: "bind",
							Src:  "/foo3",
							Dst:  "/bar3",
						},
					},
				},
			},
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Filesystem: ContainerFilesystemConfig{
						Mounts: []MountConfig{
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
						},
					},
				},
			},
		},
		want: `mount name mount-foo2 defined more than once in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Mount With Invalid Mount Type",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Filesystem: ContainerFilesystemConfig{
						Mounts: []MountConfig{
							{
								Name: "foo",
								Type: "garbage",
								Src:  "/foo",
								Dst:  "/bar",
							},
						},
					},
				},
			},
		},
		want: `unsupported mount type garbage for mount foo in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Mount With Empty Src",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Filesystem: ContainerFilesystemConfig{
						Mounts: []MountConfig{
							{
								Name: "foo",
								Type: "bind",
								Dst:  "/bar",
							},
						},
					},
				},
			},
		},
		want: `mount name foo cannot have an empty value for src in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Mount With Empty Dst",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Filesystem: ContainerFilesystemConfig{
						Mounts: []MountConfig{
							{
								Name: "foo",
								Type: "bind",
								Src:  "/foo",
							},
						},
					},
				},
			},
		},
		want: `mount name foo cannot have an empty value for dst in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Bind Mount With Options",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Filesystem: ContainerFilesystemConfig{
						Mounts: []MountConfig{
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
			},
		},
		want: `bind mount name foo cannot specify options in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Mount Def Reference Not Found",
		config: HomelabConfig{
			Global: GlobalConfig{
				MountDefs: []MountConfig{
					{
						Name: "foo",
						Type: "bind",
						Src:  "/foo",
						Dst:  "/bar",
					},
				},
			},
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Filesystem: ContainerFilesystemConfig{
						Mounts: []MountConfig{
							{
								Name: "foo2",
							},
						},
					},
				},
			},
		},
		want: `mount specified by just the name foo2 not found in defs in container {Group: g1 Container:c1} config mounts`,
	},
	{
		name: "Container Config Published Port - Container Port Empty",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Network: ContainerNetworkConfig{
						PublishedPorts: []PublishedPortConfig{
							{
								Protocol: "tcp",
								HostIP:   "127.0.0.1",
								HostPort: 5001,
							},
						},
					},
				},
			},
		},
		want: `published container port 0 cannot be non-positive in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Published Port - Container Port Negative",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Network: ContainerNetworkConfig{
						PublishedPorts: []PublishedPortConfig{
							{
								ContainerPort: -1,
								Protocol:      "tcp",
								HostIP:        "127.0.0.1",
								HostPort:      5001,
							},
						},
					},
				},
			},
		},
		want: `published container port -1 cannot be non-positive in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Published Port - Protocol Empty",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Network: ContainerNetworkConfig{
						PublishedPorts: []PublishedPortConfig{
							{
								ContainerPort: 10001,
								HostIP:        "127.0.0.1",
								HostPort:      5001,
							},
						},
					},
				},
			},
		},
		want: `published container port 10001 specifies an invalid protocol  in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Published Port - Protocol Invalid",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Network: ContainerNetworkConfig{
						PublishedPorts: []PublishedPortConfig{
							{
								ContainerPort: 10001,
								Protocol:      "garbage",
								HostIP:        "127.0.0.1",
								HostPort:      5001,
							},
						},
					},
				},
			},
		},
		want: `published container port 10001 specifies an invalid protocol garbage in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Published Port - Host IP Empty",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Network: ContainerNetworkConfig{
						PublishedPorts: []PublishedPortConfig{
							{
								ContainerPort: 10001,
								Protocol:      "tcp",
								HostPort:      5001,
							},
						},
					},
				},
			},
		},
		want: `published host IP cannot be empty for container port 10001 in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Published Port - Host IP Invalid",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Network: ContainerNetworkConfig{
						PublishedPorts: []PublishedPortConfig{
							{
								ContainerPort: 10001,
								Protocol:      "tcp",
								HostIP:        "abc.def.ghi.jkl",
								HostPort:      5001,
							},
						},
					},
				},
			},
		},
		want: `published host IP abc\.def\.ghi\.jkl for container port 10001 is invalid in container {Group: g1 Container:c1} config, reason: ParseAddr\("abc\.def\.ghi\.jkl"\): unexpected character \(at "abc\.def\.ghi\.jkl"\)`,
	},
	{
		name: "Container Config Published Port - Host Port Empty",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Network: ContainerNetworkConfig{
						PublishedPorts: []PublishedPortConfig{
							{
								ContainerPort: 10001,
								Protocol:      "tcp",
								HostIP:        "127.0.0.1",
							},
						},
					},
				},
			},
		},
		want: `published host port 0 cannot be non-positive in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Published Port - Host Port Negative",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Network: ContainerNetworkConfig{
						PublishedPorts: []PublishedPortConfig{
							{
								ContainerPort: 10001,
								Protocol:      "tcp",
								HostIP:        "127.0.0.1",
								HostPort:      -1,
							},
						},
					},
				},
			},
		},
		want: `published host port -1 cannot be non-positive in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Empty Container Config Sysctl Key",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Security: ContainerSecurityConfig{
						Sysctls: []SysctlConfig{
							{
								Value: "foo-bar",
							},
						},
					},
				},
			},
		},
		want: `empty sysctl key in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Duplicate Container Config Sysctl Key",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Security: ContainerSecurityConfig{
						Sysctls: []SysctlConfig{
							{
								Key:   "FOO",
								Value: "foo-bar",
							},
							{
								Key:   "FOO2",
								Value: "foo-bar-2",
							},
							{
								Key:   "FOO",
								Value: "foo-bar-3",
							},
						},
					},
				},
			},
		},
		want: `sysctl key FOO specified more than once in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Config Empty Sysctl Value",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Security: ContainerSecurityConfig{
						Sysctls: []SysctlConfig{
							{
								Key: "FOO",
							},
						},
					},
				},
			},
		},
		want: `empty sysctl value for sysctl FOO in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Health Config - Negative Retries",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Health: ContainerHealthConfig{
						Retries: -1,
					},
				},
			},
		},
		want: `health check retries -1 cannot be negative in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Health Config - Invalid Interval",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Health: ContainerHealthConfig{
						Interval: "garbage",
					},
				},
			},
		},
		want: `health check interval garbage is invalid in container {Group: g1 Container:c1} config, reason: time: invalid duration "garbage"`,
	},
	{
		name: "Container Health Config - Invalid Timeout",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Health: ContainerHealthConfig{
						Timeout: "garbage",
					},
				},
			},
		},
		want: `health check timeout garbage is invalid in container {Group: g1 Container:c1} config, reason: time: invalid duration "garbage"`,
	},
	{
		name: "Container Health Config - Invalid StartPeriod",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Health: ContainerHealthConfig{
						StartPeriod: "garbage",
					},
				},
			},
		},
		want: `health check start period garbage is invalid in container {Group: g1 Container:c1} config, reason: time: invalid duration "garbage"`,
	},
	{
		name: "Container Health Config - Invalid StartInterval",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Health: ContainerHealthConfig{
						StartInterval: "garbage",
					},
				},
			},
		},
		want: `health check start interval garbage is invalid in container {Group: g1 Container:c1} config, reason: time: invalid duration "garbage"`,
	},
	{
		name: "Container ShmSize Invalid Unit",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Runtime: ContainerRuntimeConfig{
						ShmSize: "1foobar",
					},
				},
			},
		},
		want: `invalid shmSize 1foobar in container {Group: g1 Container:c1} config, reason: invalid suffix: 'foobar'`,
	},
	{
		name: "Container ShmSize Invalid Value",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Runtime: ContainerRuntimeConfig{
						ShmSize: "garbage",
					},
				},
			},
		},
		want: `invalid shmSize garbage in container {Group: g1 Container:c1} config, reason: invalid size: 'garbage'`,
	},
	{
		name: "Empty Container Env Var",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Runtime: ContainerRuntimeConfig{
						Env: []ContainerEnv{
							{
								Value: "foo-bar",
							},
						},
					},
				},
			},
		},
		want: `empty env var in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Duplicate Container Env Var",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Runtime: ContainerRuntimeConfig{
						Env: []ContainerEnv{
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
			},
		},
		want: `env var FOO specified more than once in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Env Var Without Value And ValueCommand",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Runtime: ContainerRuntimeConfig{
						Env: []ContainerEnv{
							{
								Var: "FOO",
							},
						},
					},
				},
			},
		},
		want: `neither value nor valueCommand specified for env var FOO in container {Group: g1 Container:c1} config`,
	},
	{
		name: "Container Env Var With Both Value And ValueCommand",
		config: HomelabConfig{
			Groups: []ContainerGroupConfig{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []ContainerConfig{
				{
					Info: ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: ContainerImageConfig{
						Image: "foo/bar:123",
					},
					Lifecycle: ContainerLifecycleConfig{
						Order: 1,
					},
					Runtime: ContainerRuntimeConfig{
						Env: []ContainerEnv{
							{
								Var:          "FOO",
								Value:        "my-foo-bar",
								ValueCommand: "/foo/bar/baz",
							},
						},
					},
				},
			},
		},
		want: `exactly one of value or valueCommand must be specified for env var FOO in container {Group: g1 Container:c1} config`,
	},
}

func TestValidateConfigErrors(t *testing.T) {
	for _, test := range validateConfigErrorTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, gotErr := buildDeploymentFromConfig(&tc.config, fakeHostInfo)
			if gotErr == nil {
				t.Errorf(
					"HomelabConfig.validate()\nTest Case: %q\nFailure: gotErr == nil\nReason: want = %q",
					tc.name, tc.want)
				return
			}

			match, err := regexp.MatchString(fmt.Sprintf("^%s$", tc.want), gotErr.Error())
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
