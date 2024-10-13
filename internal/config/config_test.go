package config

import (
	"bytes"
	"testing"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdudehomelab/homelab/internal/config/env"
	"github.com/tuxdudehomelab/homelab/internal/deepcopy"
	logger "github.com/tuxdudehomelab/homelab/internal/log"
	"github.com/tuxdudehomelab/homelab/internal/testhelpers"
	"github.com/tuxdudehomelab/homelab/internal/testutils"
)

var applyConfigEnvToContainerTests = []struct {
	name              string
	container         Container
	globalEnvMap      env.EnvMap
	globalEnvOrder    env.EnvOrder
	containerEnvMap   env.EnvMap
	containerEnvOrder env.EnvOrder
	want              Container
}{
	{
		name: "Container Config - ApplyConfigEnv - Exhaustive",
		container: Container{
			Info: ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			Image: ContainerImage{
				Image: "foo/bar:abc",
			},
			Lifecycle: ContainerLifecycle{
				Order:        1,
				StartPreHook: "$$CONTAINER_SCRIPTS_DIR$$/my-start-prehook.sh",
			},
			User: ContainerUser{
				User:         "$$USER_ID$$",
				PrimaryGroup: "$$USER_PRIMARY_GROUP_ID$$",
				AdditionalGroups: []string{
					"$$ENV_GROUP_FOO$$",
					"$$ENV_GROUP_BAR$$",
				},
			},
			Filesystem: ContainerFilesystem{
				ReadOnlyRootfs: true,
				Mounts: []Mount{
					{
						Name: "mount-1",
						Type: "bind",
						Src:  "$$CONTAINER_CONFIGS_DIR$$/config.yml",
						Dst:  "/data/$$USER_NAME$$/config.yml",
					},
					{
						Name: "some-other-mount",
						Type: "bind",
						Src:  "$$CONTAINER_DATA_DIR$$/my-data",
						Dst:  "$$ENV_MOUNT_DEST_DIR$$/my-data",
					},
					{
						Name: "some-other-mount-2",
						Type: "bind",
						Src:  "$$CONTAINER_BASE_DIR$$/abc/xyz",
						Dst:  "/abc/xyz",
					},
					{
						Name: "some-other-mount-3",
						Type: "bind",
						Src:  "$$CONTAINER_GROUP_BASE_DIR$$/tuv/wxy",
						Dst:  "/foo123/bar123/baz123",
					},
					{
						Name:    "tmpfs-mount",
						Type:    "tmpfs",
						Dst:     "/tmp/cache-$$USER_PRIMARY_GROUP_NAME$$",
						Options: "tmpfs-size=$$ENV_TMPFS_SIZE$$",
					},
				},
				Devices: ContainerDevice{
					Static: []Device{
						{
							Src: "$$ENV_SRC_DEV$$",
							Dst: "$$ENV_DST_DEV$$",
						},
					},
				},
			},
			Network: ContainerNetwork{
				HostName:   "MyHost-$$HOST_NAME$$",
				DomainName: "$$ENV_DOMAIN$$",
				DNSServers: []string{
					"1.1.1.1",
					"$$ENV_DNS_SERVER$$",
				},
				DNSOptions: []string{
					"dns-option-1",
					"$$ENV_DNS_OPTION$$",
				},
				DNSSearch: []string{
					"dns-ct-search-1",
					"$$ENV_DNS_SEARCH$$",
				},
				ExtraHosts: []string{
					"my-extra-host-1",
					"$$HOST_NAME$$-extra",
				},
				PublishedPorts: []PublishedPort{
					{
						ContainerPort: "$$MY_CONTAINER_PORT$$",
						Protocol:      "tcp",
						HostIP:        "$$HOST_IP$$",
						HostPort:      "$$MY_HOST_PORT$$",
					},
				},
			},
			Runtime: ContainerRuntime{
				Env: []ContainerEnv{
					{
						Var:   "$$ENV_VAR_1$$",
						Value: "$$ENV_VAR_1_VAL$$",
					},
					{
						Var: "MY_ENV_2",
						ValueCommand: []string{
							"$$ENV_VAR_2_VAL_CMD$$",
						},
					},
					{
						Var:   "MY_ENV_3",
						Value: "SomeHostName.$$HUMAN_FRIENDLY_HOST_NAME$$.SomeDomainName",
					},
				},
				Args: []string{
					"foo-$$HOST_NAME$$",
					"bar-$$ENV_VAR_1$$",
				},
			},
		},
		globalEnvMap: env.EnvMap{
			"ENV_GROUP_FOO":      "groupfoo",
			"ENV_GROUP_BAR":      "groupbar",
			"ENV_MOUNT_DEST_DIR": "/mnt/dst1",
			"ENV_TMPFS_SIZE":     "100000000",
			"ENV_SRC_DEV":        "/dev/src",
			"ENV_DST_DEV":        "/dev/dst",
			"ENV_DOMAIN":         "my-domain",
			"ENV_DNS_SERVER":     "10.11.11.11",
			"ENV_DNS_OPTION":     "dns-option-2",
			"ENV_DNS_SEARCH":     "dns-ct-search-2",
			"MY_HOST_PORT":       "678",
		},
		globalEnvOrder: env.EnvOrder{
			"ENV_GROUP_FOO",
			"ENV_GROUP_BAR",
			"ENV_MOUNT_DEST_DIR",
			"ENV_TMPFS_SIZE",
			"ENV_SRC_DEV",
			"ENV_DST_DEV",
			"ENV_DOMAIN",
			"ENV_DNS_SERVER",
			"ENV_DNS_OPTION",
			"ENV_DNS_SEARCH",
			"MY_HOST_PORT",
		},
		containerEnvMap: env.EnvMap{
			"ENV_VAR_1":         "MY_ENV_1",
			"ENV_VAR_1_VAL":     "my-env-1-val",
			"ENV_VAR_2_VAL_CMD": "cat /foo/bar.txt",
			"MY_CONTAINER_PORT": "12345",
		},
		containerEnvOrder: env.EnvOrder{
			"ENV_VAR_1",
			"ENV_VAR_1_VAL",
			"ENV_VAR_2_VAL_CMD",
			"MY_CONTAINER_PORT",
		},
		want: Container{
			Info: ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			Image: ContainerImage{
				Image: "foo/bar:abc",
			},
			Lifecycle: ContainerLifecycle{
				Order:        1,
				StartPreHook: "/tmp/base-dir/g1/c1/scripts/my-start-prehook.sh",
			},
			User: ContainerUser{
				User:         "55555",
				PrimaryGroup: "44444",
				AdditionalGroups: []string{
					"groupfoo",
					"groupbar",
				},
			},
			Filesystem: ContainerFilesystem{
				ReadOnlyRootfs: true,
				Mounts: []Mount{
					{
						Name: "mount-1",
						Type: "bind",
						Src:  "/tmp/base-dir/g1/c1/configs/config.yml",
						Dst:  "/data/fakeuser/config.yml",
					},
					{
						Name: "some-other-mount",
						Type: "bind",
						Src:  "/tmp/base-dir/g1/c1/data/my-data",
						Dst:  "/mnt/dst1/my-data",
					},
					{
						Name: "some-other-mount-2",
						Type: "bind",
						Src:  "/tmp/base-dir/g1/c1/abc/xyz",
						Dst:  "/abc/xyz",
					},
					{
						Name: "some-other-mount-3",
						Type: "bind",
						Src:  "/tmp/base-dir/g1/tuv/wxy",
						Dst:  "/foo123/bar123/baz123",
					},
					{
						Name:    "tmpfs-mount",
						Type:    "tmpfs",
						Dst:     "/tmp/cache-fakegroup1",
						Options: "tmpfs-size=100000000",
					},
				},
				Devices: ContainerDevice{
					Static: []Device{
						{
							Src: "/dev/src",
							Dst: "/dev/dst",
						},
					},
				},
			},
			Network: ContainerNetwork{
				HostName:   "MyHost-fakehost",
				DomainName: "my-domain",
				DNSServers: []string{
					"1.1.1.1",
					"10.11.11.11",
				},
				DNSOptions: []string{
					"dns-option-1",
					"dns-option-2",
				},
				DNSSearch: []string{
					"dns-ct-search-1",
					"dns-ct-search-2",
				},
				ExtraHosts: []string{
					"my-extra-host-1",
					"fakehost-extra",
				},
				PublishedPorts: []PublishedPort{
					{
						ContainerPort: "12345",
						Protocol:      "tcp",
						HostIP:        "10.76.77.78",
						HostPort:      "678",
					},
				},
			},
			Runtime: ContainerRuntime{
				Env: []ContainerEnv{
					{
						Var:   "MY_ENV_1",
						Value: "my-env-1-val",
					},
					{
						Var: "MY_ENV_2",
						ValueCommand: []string{
							"cat /foo/bar.txt",
						},
					},
					{
						Var:   "MY_ENV_3",
						Value: "SomeHostName.FakeHost.SomeDomainName",
					},
				},
				Args: []string{
					"foo-fakehost",
					"bar-MY_ENV_1",
				},
			},
		},
	},
}

func TestApplyConfigEnvToContainer(t *testing.T) {
	t.Parallel()

	for _, test := range applyConfigEnvToContainerTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := testutils.NewCapturingTestLogger(zzzlog.LvlInfo, new(bytes.Buffer))
			ctx := testutils.NewTestContext(&testutils.TestContextInfo{})
			ctx = logger.WithLogger(ctx, l)

			e := env.NewSystemConfigEnvManager(ctx)
			e = e.NewGlobalConfigEnvManager(ctx, "/tmp/base-dir", tc.globalEnvMap, tc.globalEnvOrder)
			e = e.NewContainerConfigEnvManager(ctx, "/tmp/base-dir/g1", "/tmp/base-dir/g1/c1", tc.containerEnvMap, tc.containerEnvOrder)

			got := deepcopy.MustCopy(tc.container)
			got.ApplyConfigEnv(e)
			if !testhelpers.CmpDiff(t, "Container.ApplyConfigEnv()", tc.name, "apply result", tc.want, got) {
				return
			}
		})
	}
}

var applyConfigEnvToGlobalTests = []struct {
	name           string
	global         Global
	globalEnvMap   env.EnvMap
	globalEnvOrder env.EnvOrder
	want           Global
}{
	{
		name: "Global Container Config - ApplyConfigEnv - Exhaustive",
		global: Global{
			MountDefs: []Mount{
				{
					Name: "mount-1",
					Type: "bind",
					Src:  "/$$HOST_NAME$$/config.yml",
					Dst:  "/data/$$USER_NAME$$/config.yml",
				},
				{
					Name: "some-other-mount",
					Type: "bind",
					Src:  "/$$USER_PRIMARY_GROUP_ID$$/my-data",
					Dst:  "$$ENV_MOUNT_DEST_DIR_1$$/my-data",
				},
				{
					Name: "some-other-mount-2",
					Type: "bind",
					Src:  "/$$USER_ID$$/abc/xyz",
					Dst:  "/abc/xyz",
				},
				{
					Name:    "tmpfs-mount",
					Type:    "tmpfs",
					Dst:     "/tmp/cache-$$USER_PRIMARY_GROUP_NAME$$",
					Options: "tmpfs-size=$$ENV_TMPFS_SIZE_1$$",
				},
			},
			Container: GlobalContainer{
				DomainName: "$$ENV_DOMAIN$$",
				DNSSearch: []string{
					"dns-ct-search-1",
					"$$ENV_DNS_SEARCH$$",
				},
				Env: []ContainerEnv{
					{
						Var:   "$$ENV_VAR_1$$",
						Value: "$$ENV_VAR_1_VAL$$",
					},
					{
						Var: "MY_ENV_2",
						ValueCommand: []string{
							"$$ENV_VAR_2_VAL_CMD$$",
						},
					},
					{
						Var:   "MY_ENV_3",
						Value: "SomeHostName.$$HUMAN_FRIENDLY_HOST_NAME$$.SomeDomainName",
					},
				},
				Mounts: []Mount{
					{
						Name: "mount-11",
						Type: "bind",
						Src:  "/$$HUMAN_FRIENDLY_HOST_NAME$$/config.yml",
						Dst:  "/data/$$USER_NAME$$/config.yml",
					},
					{
						Name: "some-other-mount-12",
						Type: "bind",
						Src:  "$$HOMELAB_BASE_DIR$$/my-data",
						Dst:  "$$ENV_MOUNT_DEST_DIR_2$$/my-data",
					},
					{
						Name: "some-other-mount-13",
						Type: "bind",
						Src:  "/$$HOST_IP$$/abc/xyz",
						Dst:  "/abc/xyz",
					},
					{
						Name:    "tmpfs-mount-14",
						Type:    "tmpfs",
						Dst:     "/tmp/cache-$$USER_PRIMARY_GROUP_NAME$$",
						Options: "tmpfs-size=$$ENV_TMPFS_SIZE_2$$",
					},
				},
			},
		},
		globalEnvMap: env.EnvMap{
			"ENV_GROUP_FOO":        "groupfoo",
			"ENV_GROUP_BAR":        "groupbar",
			"ENV_MOUNT_DEST_DIR_1": "/mnt/dst1",
			"ENV_MOUNT_DEST_DIR_2": "/mnt/dst2",
			"ENV_TMPFS_SIZE_1":     "100000000",
			"ENV_TMPFS_SIZE_2":     "200000000",
			"ENV_DOMAIN":           "my-domain",
			"ENV_DNS_SERVER":       "10.11.11.11",
			"ENV_DNS_SEARCH":       "dns-ct-search-2",
			"ENV_VAR_1":            "MY_ENV_1",
			"ENV_VAR_1_VAL":        "my-env-1-val",
			"ENV_VAR_2_VAL_CMD":    "cat /foo/bar.txt",
		},
		globalEnvOrder: env.EnvOrder{
			"ENV_GROUP_FOO",
			"ENV_GROUP_BAR",
			"ENV_MOUNT_DEST_DIR_1",
			"ENV_MOUNT_DEST_DIR_2",
			"ENV_TMPFS_SIZE_1",
			"ENV_TMPFS_SIZE_2",
			"ENV_DOMAIN",
			"ENV_DNS_SERVER",
			"ENV_DNS_SEARCH",
			"ENV_VAR_1",
			"ENV_VAR_1_VAL",
			"ENV_VAR_2_VAL_CMD",
		},
		want: Global{
			MountDefs: []Mount{
				{
					Name: "mount-1",
					Type: "bind",
					Src:  "/fakehost/config.yml",
					Dst:  "/data/fakeuser/config.yml",
				},
				{
					Name: "some-other-mount",
					Type: "bind",
					Src:  "/44444/my-data",
					Dst:  "/mnt/dst1/my-data",
				},
				{
					Name: "some-other-mount-2",
					Type: "bind",
					Src:  "/55555/abc/xyz",
					Dst:  "/abc/xyz",
				},
				{
					Name:    "tmpfs-mount",
					Type:    "tmpfs",
					Dst:     "/tmp/cache-fakegroup1",
					Options: "tmpfs-size=100000000",
				},
			},
			Container: GlobalContainer{
				DomainName: "my-domain",
				DNSSearch: []string{
					"dns-ct-search-1",
					"dns-ct-search-2",
				},
				Env: []ContainerEnv{
					{
						Var:   "MY_ENV_1",
						Value: "my-env-1-val",
					},
					{
						Var: "MY_ENV_2",
						ValueCommand: []string{
							"cat /foo/bar.txt",
						},
					},
					{
						Var:   "MY_ENV_3",
						Value: "SomeHostName.FakeHost.SomeDomainName",
					},
				},
				Mounts: []Mount{
					{
						Name: "mount-11",
						Type: "bind",
						Src:  "/FakeHost/config.yml",
						Dst:  "/data/fakeuser/config.yml",
					},
					{
						Name: "some-other-mount-12",
						Type: "bind",
						Src:  "/tmp/base-dir/my-data",
						Dst:  "/mnt/dst2/my-data",
					},
					{
						Name: "some-other-mount-13",
						Type: "bind",
						Src:  "/10.76.77.78/abc/xyz",
						Dst:  "/abc/xyz",
					},
					{
						Name:    "tmpfs-mount-14",
						Type:    "tmpfs",
						Dst:     "/tmp/cache-fakegroup1",
						Options: "tmpfs-size=200000000",
					},
				},
			},
		},
	},
}

func TestApplyConfigEnvToGlobal(t *testing.T) {
	t.Parallel()

	for _, test := range applyConfigEnvToGlobalTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := testutils.NewCapturingTestLogger(zzzlog.LvlInfo, new(bytes.Buffer))
			ctx := testutils.NewTestContext(&testutils.TestContextInfo{})
			ctx = logger.WithLogger(ctx, l)

			e := env.NewSystemConfigEnvManager(ctx)
			e = e.NewGlobalConfigEnvManager(ctx, "/tmp/base-dir", tc.globalEnvMap, tc.globalEnvOrder)

			got := deepcopy.MustCopy(tc.global)
			got.ApplyConfigEnv(e)
			if !testhelpers.CmpDiff(t, "Global.ApplyConfigEnv()", tc.name, "apply result", tc.want, got) {
				return
			}
		})
	}
}
