package config

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdudehomelab/homelab/internal/cmdexec"
	"github.com/tuxdudehomelab/homelab/internal/cmdexec/fakecmdexec"
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
				Order: 1,
				StartPreHook: []string{
					"$$CONTAINER_SCRIPTS_DIR$$/my-start-prehook.sh",
				},
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
					DynamicCommand: []string{
						"$$CONTAINER_BASE_DIR$$/my-devices-lister.sh",
						"$$HOST_NAME$$",
						"$$HUMAN_FRIENDLY_HOST_NAME$$",
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
						Var:   "MY_ENV_2",
						Value: "$$ENV_VAR_2_VAL_CMD$$",
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
				Order: 1,
				StartPreHook: []string{
					"/tmp/base-dir/g1/c1/scripts/my-start-prehook.sh",
				},
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
					DynamicCommand: []string{
						"/tmp/base-dir/g1/c1/my-devices-lister.sh",
						"fakehost",
						"FakeHost",
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
						Var:   "MY_ENV_2",
						Value: "cat /foo/bar.txt",
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

var applyCmdExecutorToContainerTests = []struct {
	name      string
	container Container
	exec      cmdexec.Executor
	want      Container
}{
	{
		name: "Container Config - ApplyCmdExecutor - Exhaustive",
		container: Container{
			Info: ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			Image: ContainerImage{
				Image: "foo/bar:abc",
			},
			Lifecycle: ContainerLifecycle{
				Order: 1,
			},
			Filesystem: ContainerFilesystem{
				Devices: ContainerDevice{
					Static: []Device{
						{
							Src: "/dev/stat1",
							Dst: "/dev/stat2",
						},
					},
					DynamicCommand: []string{
						"list-devices.sh",
						"dummy-arg1",
						"dummy-arg2",
					},
				},
			},
		},
		exec: fakecmdexec.NewFakeExecutor(&fakecmdexec.FakeExecutorInitInfo{
			ValidCmds: []fakecmdexec.FakeValidCmdInfo{
				{
					Cmd: []string{
						"list-devices.sh",
						"dummy-arg1",
						"dummy-arg2",
					},
					Output: "/dev/s1:/dev/d1:r,/dev/s2:/dev/d2:w,/dev/s3:/dev/d3:m,/dev/s4:/dev/d4:rw,/dev/s5:/dev/d5:rm,/dev/s6:/dev/d6:wm,/dev/s7:/dev/d7:rwm,/dev/s8:/dev/d8:",
				},
			},
		}),
		want: Container{
			Info: ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			Image: ContainerImage{
				Image: "foo/bar:abc",
			},
			Lifecycle: ContainerLifecycle{
				Order: 1,
			},
			Filesystem: ContainerFilesystem{
				Devices: ContainerDevice{
					Static: []Device{
						{
							Src: "/dev/stat1",
							Dst: "/dev/stat2",
						},
					},
					DynamicCommand: []string{
						"list-devices.sh",
						"dummy-arg1",
						"dummy-arg2",
					},
					Dynamic: []Device{
						{
							Src:           "/dev/s1",
							Dst:           "/dev/d1",
							DisallowWrite: true,
							DisallowMknod: true,
						},
						{
							Src:           "/dev/s2",
							Dst:           "/dev/d2",
							DisallowRead:  true,
							DisallowMknod: true,
						},
						{
							Src:           "/dev/s3",
							Dst:           "/dev/d3",
							DisallowRead:  true,
							DisallowWrite: true,
						},
						{
							Src:           "/dev/s4",
							Dst:           "/dev/d4",
							DisallowMknod: true,
						},
						{
							Src:           "/dev/s5",
							Dst:           "/dev/d5",
							DisallowWrite: true,
						},
						{
							Src:          "/dev/s6",
							Dst:          "/dev/d6",
							DisallowRead: true,
						},
						{
							Src: "/dev/s7",
							Dst: "/dev/d7",
						},
						{
							Src:           "/dev/s8",
							Dst:           "/dev/d8",
							DisallowRead:  true,
							DisallowWrite: true,
							DisallowMknod: true,
						},
					},
				},
			},
		},
	},
	{
		name: "Container Config - ApplyCmdExecutor - Real Executor",
		container: Container{
			Info: ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			Image: ContainerImage{
				Image: "foo/bar:abc",
			},
			Lifecycle: ContainerLifecycle{
				Order: 1,
			},
			Filesystem: ContainerFilesystem{
				Devices: ContainerDevice{
					Static: []Device{
						{
							Src: "/dev/stat1",
							Dst: "/dev/stat2",
						},
					},
					DynamicCommand: []string{
						"echo",
						"/dev/dyns1:/dev/dynd1:rw",
					},
				},
			},
		},
		exec: cmdexec.NewExecutor(),
		want: Container{
			Info: ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			Image: ContainerImage{
				Image: "foo/bar:abc",
			},
			Lifecycle: ContainerLifecycle{
				Order: 1,
			},
			Filesystem: ContainerFilesystem{
				Devices: ContainerDevice{
					Static: []Device{
						{
							Src: "/dev/stat1",
							Dst: "/dev/stat2",
						},
					},
					DynamicCommand: []string{
						"echo",
						"/dev/dyns1:/dev/dynd1:rw",
					},
					Dynamic: []Device{
						{
							Src:           "/dev/dyns1",
							Dst:           "/dev/dynd1",
							DisallowMknod: true,
						},
					},
				},
			},
		},
	},
}

func TestApplyCmdExecutorToContainer(t *testing.T) {
	t.Parallel()

	for _, test := range applyCmdExecutorToContainerTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := deepcopy.MustCopy(tc.container)
			gotErr := got.ApplyCmdExecutor(tc.exec)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "Container.ApplyCmdExecutor()", tc.name, gotErr)
				return
			}

			if !testhelpers.CmpDiff(t, "Container.ApplyCmdExecutor()", tc.name, "apply result", tc.want, got) {
				return
			}
		})
	}
}

var applyCmdExecutorToContainerErrorTests = []struct {
	name      string
	container Container
	exec      cmdexec.Executor
	want      string
}{
	{
		name: "Container Config - ApplyCmdExecutor - Command Failed - Real Executor",
		container: Container{
			Filesystem: ContainerFilesystem{
				Devices: ContainerDevice{
					DynamicCommand: []string{
						"ls",
						"-e",
					},
				},
			},
		},
		exec: cmdexec.NewExecutor(),
		want: `(?s)command failed ls \["-e"\], reason: exit status 2, stderr: ls: invalid option -- 'e'\s.+`,
	},
	{
		name: "Container Config - ApplyCmdExecutor - Executable Not Found - Real Executor",
		container: Container{
			Filesystem: ContainerFilesystem{
				Devices: ContainerDevice{
					DynamicCommand: []string{
						"invalid-command",
						"invalid-arg-1",
						"invalid-arg-2",
					},
				},
			},
		},
		exec: cmdexec.NewExecutor(),
		want: `command failed invalid-command \["invalid-arg-1" "invalid-arg-2"\], reason: exec: "invalid-command": executable file not found in \$PATH`,
	},
	{
		name: "Container Config - ApplyCmdExecutor - Command Execution Failed - 1",
		container: Container{
			Filesystem: ContainerFilesystem{
				Devices: ContainerDevice{
					DynamicCommand: []string{
						"invalid-command",
						"invalid-arg-1",
						"invalid-arg-2",
					},
				},
			},
		},
		exec: fakecmdexec.NewFakeExecutor(&fakecmdexec.FakeExecutorInitInfo{
			ValidCmds: []fakecmdexec.FakeValidCmdInfo{
				{
					Cmd: []string{
						"list-devices.sh",
						"dummy-arg1",
						"dummy-arg2",
					},
					Output: "/dev/s1:/dev/d1:rw",
				},
			},
		}),
		want: `invalid fake executor command invalid-command \["invalid-arg-1" "invalid-arg-2"\]`,
	},
	{
		name: "Container Config - ApplyCmdExecutor - Command Execution Failed - 2",
		container: Container{
			Filesystem: ContainerFilesystem{
				Devices: ContainerDevice{
					DynamicCommand: []string{
						"invalid-command",
						"invalid-arg-1",
						"invalid-arg-2",
					},
				},
			},
		},
		exec: fakecmdexec.NewFakeExecutor(&fakecmdexec.FakeExecutorInitInfo{
			ErrorCmds: []fakecmdexec.FakeErrorCmdInfo{
				{
					Cmd: []string{
						"invalid-command",
						"invalid-arg-1",
						"invalid-arg-2",
					},
					Err: fmt.Errorf("invalid command, reason: command not found"),
				},
			},
		}),
		want: `invalid command, reason: command not found`,
	},
	{
		name: "Container Config - ApplyCmdExecutor - Less Than Three Parts In Device Spec",
		container: Container{
			Filesystem: ContainerFilesystem{
				Devices: ContainerDevice{
					DynamicCommand: []string{
						"list-devices.sh",
					},
				},
			},
		},
		exec: fakecmdexec.NewFakeExecutor(&fakecmdexec.FakeExecutorInitInfo{
			ValidCmds: []fakecmdexec.FakeValidCmdInfo{
				{
					Cmd: []string{
						"list-devices.sh",
					},
					Output: "/dev/s1:/dev/d1",
				},
			},
		}),
		want: `expected three parts separated by ':' for each dynamic device spec, found 2`,
	},
	{
		name: "Container Config - ApplyCmdExecutor - More Than Three Parts In Device Spec",
		container: Container{
			Filesystem: ContainerFilesystem{
				Devices: ContainerDevice{
					DynamicCommand: []string{
						"list-devices.sh",
					},
				},
			},
		},
		exec: fakecmdexec.NewFakeExecutor(&fakecmdexec.FakeExecutorInitInfo{
			ValidCmds: []fakecmdexec.FakeValidCmdInfo{
				{
					Cmd: []string{
						"list-devices.sh",
					},
					Output: "/dev/s1:/dev/d1:r:w",
				},
			},
		}),
		want: `expected three parts separated by ':' for each dynamic device spec, found 4`,
	},
	{
		name: "Container Config - ApplyCmdExecutor - More Than Three Permissions",
		container: Container{
			Filesystem: ContainerFilesystem{
				Devices: ContainerDevice{
					DynamicCommand: []string{
						"list-devices.sh",
					},
				},
			},
		},
		exec: fakecmdexec.NewFakeExecutor(&fakecmdexec.FakeExecutorInitInfo{
			ValidCmds: []fakecmdexec.FakeValidCmdInfo{
				{
					Cmd: []string{
						"list-devices.sh",
					},
					Output: "/dev/s1:/dev/d1:rwmg",
				},
			},
		}),
		want: `mode part of dynamic device spec rwmg is invalid as it can be at most specify three permissions`,
	},
	{
		name: "Container Config - ApplyCmdExecutor - Read Permission More Than Once",
		container: Container{
			Filesystem: ContainerFilesystem{
				Devices: ContainerDevice{
					DynamicCommand: []string{
						"list-devices.sh",
					},
				},
			},
		},
		exec: fakecmdexec.NewFakeExecutor(&fakecmdexec.FakeExecutorInitInfo{
			ValidCmds: []fakecmdexec.FakeValidCmdInfo{
				{
					Cmd: []string{
						"list-devices.sh",
					},
					Output: "/dev/s1:/dev/d1:rrw",
				},
			},
		}),
		want: `mode part of dynamic device spec rrw specifies read more than once`,
	},
	{
		name: "Container Config - ApplyCmdExecutor - Write Permission More Than Once",
		container: Container{
			Filesystem: ContainerFilesystem{
				Devices: ContainerDevice{
					DynamicCommand: []string{
						"list-devices.sh",
					},
				},
			},
		},
		exec: fakecmdexec.NewFakeExecutor(&fakecmdexec.FakeExecutorInitInfo{
			ValidCmds: []fakecmdexec.FakeValidCmdInfo{
				{
					Cmd: []string{
						"list-devices.sh",
					},
					Output: "/dev/s1:/dev/d1:wmw",
				},
			},
		}),
		want: `mode part of dynamic device spec wmw specifies write more than once`,
	},
	{
		name: "Container Config - ApplyCmdExecutor - Mknod Permission More Than Once",
		container: Container{
			Filesystem: ContainerFilesystem{
				Devices: ContainerDevice{
					DynamicCommand: []string{
						"list-devices.sh",
					},
				},
			},
		},
		exec: fakecmdexec.NewFakeExecutor(&fakecmdexec.FakeExecutorInitInfo{
			ValidCmds: []fakecmdexec.FakeValidCmdInfo{
				{
					Cmd: []string{
						"list-devices.sh",
					},
					Output: "/dev/s1:/dev/d1:rmm",
				},
			},
		}),
		want: `mode part of dynamic device spec rmm specifies mknod more than once`,
	},
}

func TestApplyCmdExecutorToContainerErrors(t *testing.T) {
	t.Parallel()

	for _, test := range applyCmdExecutorToContainerErrorTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotErr := tc.container.ApplyCmdExecutor(tc.exec)
			if gotErr == nil {
				testhelpers.LogErrorNil(t, "Container.ApplyCmdExecutor()", tc.name, tc.want)
				return
			}

			if !testhelpers.RegexMatch(t, "Container.ApplyCmdExecutor()", tc.name, "gotErr error string", tc.want, gotErr.Error()) {
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
						Var:   "MY_ENV_2",
						Value: "$$ENV_VAR_2_VAL_CMD$$",
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
						Var:   "MY_ENV_2",
						Value: "cat /foo/bar.txt",
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
