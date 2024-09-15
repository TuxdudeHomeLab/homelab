package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var validParseTests = []struct {
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
        - type: bind
          src: $$CONFIG_DIR$$/generated/config.yml
          dst: /data/blocky/config/config.yml
          readOnly: true
        - name: homelab-self-signed-tls-cert
        - type: tmpfs
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
        - baz
`,
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
								Type:     "bind",
								Src:      "$$CONFIG_DIR$$/generated/config.yml",
								Dst:      "/data/blocky/config/config.yml",
								ReadOnly: true,
							},
							{
								Name: "homelab-self-signed-tls-cert",
							},
							{
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
}

func TestParseConfig(t *testing.T) {
	for _, test := range validParseTests {
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
