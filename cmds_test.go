package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/tuxdude/zzzlog"
)

var executeHomelabCmdTests = []struct {
	name    string
	args    []string
	ctxInfo *testContextInfo
	want    string
}{
	{
		name: "Homelab Command - Show Version",
		args: []string{
			"--version",
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `homelab version my-pkg-version \[Revision: my-pkg-commit @ my-pkg-timestamp\]`,
	},
	{
		name: "Homelab Command - Show Help",
		args: []string{
			"--help",
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `(?s)A CLI for managing both the configuration and deployment of groups of docker containers on a given host\.
The configuration is managed using a yaml file\. The configuration specifies the container groups and individual containers, their properties and how to deploy them\.
Usage:
.+
Use "homelab \[command\] --help" for more information about a command\.`,
	},
	{
		name: "Homelab Command - Show Config",
		args: []string{
			"show-config",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/show-config-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `Homelab config:
{
  "ipam": {
    "networks": {
      "bridgeModeNetworks": \[
        {
          "name": "net1",
          "hostInterfaceName": "docker-net1",
          "cidr": "172\.18\.100\.0/24",
          "priority": 1,
          "containers": \[
            {
              "ip": "172\.18\.100\.11",
              "container": {
                "group": "g1",
                "container": "c1"
              }
            },
            {
              "ip": "172\.18\.100\.12",
              "container": {
                "group": "g1",
                "container": "c2"
              }
            }
          \]
        },
        {
          "name": "net2",
          "hostInterfaceName": "docker-net2",
          "cidr": "172\.18\.101\.0/24",
          "priority": 1,
          "containers": \[
            {
              "ip": "172\.18\.101\.21",
              "container": {
                "group": "g2",
                "container": "c3"
              }
            }
          \]
        }
      \]
    }
  },
  "hosts": \[
    {
      "name": "fakehost",
      "allowedContainers": \[
        {
          "group": "g1",
          "container": "c1"
        }
      \]
    },
    {
      "name": "host2"
    }
  \],
  "groups": \[
    {
      "name": "g1",
      "order": 1
    },
    {
      "name": "g2",
      "order": 2
    }
  \],
  "containers": \[
    {
      "info": {
        "group": "g1",
        "container": "c1"
      },
      "image": {
        "image": "abc/xyz"
      },
      "lifecycle": {
        "order": 1
      }
    },
    {
      "info": {
        "group": "g1",
        "container": "c2"
      },
      "image": {
        "image": "abc/xyz2"
      },
      "lifecycle": {
        "order": 2
      }
    },
    {
      "info": {
        "group": "g2",
        "container": "c3"
      },
      "image": {
        "image": "abc/xyz3"
      },
      "lifecycle": {
        "order": 1
      }
    }
  \]
}`,
	},
	{
		name: "Homelab Command - Show Config - Custom CLI Config Path",
		args: []string{
			"show-config",
			"--cli-config",
			fmt.Sprintf("%s/testdata/cli-configs/show-config-cmd/config.yaml", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `Homelab config:
{
  "ipam": {
    "networks": {
      "bridgeModeNetworks": \[
        {
          "name": "net1",
          "hostInterfaceName": "docker-net1",
          "cidr": "172\.18\.100\.0/24",
          "priority": 1,
          "containers": \[
            {
              "ip": "172\.18\.100\.11",
              "container": {
                "group": "g1",
                "container": "c1"
              }
            }
          \]
        }
      \]
    }
  },
  "hosts": \[
    {
      "name": "fakehost",
      "allowedContainers": \[
        {
          "group": "g1",
          "container": "c1"
        }
      \]
    }
  \],
  "groups": \[
    {
      "name": "g1",
      "order": 1
    }
  \],
  "containers": \[
    {
      "info": {
        "group": "g1",
        "container": "c1"
      },
      "image": {
        "image": "abc/xyz"
      },
      "lifecycle": {
        "order": 10
      }
    }
  \]
}`,
	},
	{
		name: "Homelab Command - Start - All Groups With Real Host Info",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				validImagesForPull: stringSet{
					"abc/xyz": {},
				},
			}),
			useRealHostInfo: true,
		},
		want: `Container g1-c1 not allowed to run on host [^\s]+
Container g1-c2 not allowed to run on host [^\s]+
Container g2-c3 not allowed to run on host [^\s]+`,
	},
	{
		name: "Homelab Command - Start - All Groups",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				validImagesForPull: stringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
			}),
		},
		want: `Pulling image: abc/xyz
Created network net1
Started container g1-c1
Container g1-c2 not allowed to run on host FakeHost
Pulling image: abc/xyz3
Created network net2
Started container g2-c3`,
	},
	{
		name: "Homelab Command - Start - All Groups - Container Create Warning",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				validImagesForPull: stringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
				warnContainerCreate: stringSet{
					"g1-c1": {},
				},
			}),
		},
		want: `Pulling image: abc/xyz
Created network net1
Warnings encountered while creating the container g1-c1
1 - first warning generated during container create for g1-c1 on the fake docker host
2 - second warning generated during container create for g1-c1 on the fake docker host
3 - third warning generated during container create for g1-c1 on the fake docker host
Started container g1-c1
Container g1-c2 not allowed to run on host FakeHost
Pulling image: abc/xyz3
Created network net2
Started container g2-c3`,
	},
	{
		name: "Homelab Command - Start - All Groups - Network Create Warning",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				validImagesForPull: stringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
				warnNetworkCreate: stringSet{
					"net1": {},
				},
			}),
		},
		want: `Pulling image: abc/xyz
Warning encountered while creating the network net1
warning generated during network create for network net1 on the fake docker host
Created network net1
Started container g1-c1
Container g1-c2 not allowed to run on host FakeHost
Pulling image: abc/xyz3
Created network net2
Started container g2-c3`,
	},
	{
		name: "Homelab Command - Start - All Groups - One Existing Image",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				existingImages: stringSet{
					"abc/xyz3": {},
				},
				validImagesForPull: stringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
			}),
		},
		want: `Pulling image: abc/xyz
Created network net1
Started container g1-c1
Container g1-c2 not allowed to run on host FakeHost
Pulled newer version of image abc/xyz3: [a-z0-9]{64}
Created network net2
Started container g2-c3`,
	},
	{
		name: "Homelab Command - Start - All Groups With Multiple Same Order Containers",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd-with-multiple-same-order-containers", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				validImagesForPull: stringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
					"abc/xyz4": {},
				},
			}),
		},
		want: `Pulling image: abc/xyz
Created network net1
Started container g1-c1
Container g1-c2 not allowed to run on host FakeHost
Pulling image: abc/xyz3
Started container g1-c3
Pulling image: abc/xyz4
Created network net2
Started container g2-c4`,
	},
	{
		name: "Homelab Command - Start - All Groups With No Network Endpoints Containers",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd-with-no-network-endpoints-containers", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				validImagesForPull: stringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
					"abc/xyz4": {},
				},
			}),
		},
		want: `Pulling image: abc/xyz
Created network net1
Started container g1-c1
Container g1-c2 not allowed to run on host FakeHost
Pulling image: abc/xyz3
Container g1-c3 has no network endpoints configured, this is uncommon!
Started container g1-c3
Pulling image: abc/xyz4
Created network net2
Started container g2-c4`,
	},
	{
		name: "Homelab Command - Start - One Group",
		args: []string{
			"start",
			"--group",
			"g1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				validImagesForPull: stringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Pulling image: abc/xyz
Created network net1
Started container g1-c1
Container g1-c2 not allowed to run on host FakeHost`,
	},
	{
		name: "Homelab Command - Start - One Container",
		args: []string{
			"start",
			"--group",
			"g1",
			"--container",
			"c1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				validImagesForPull: stringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Pulling image: abc/xyz
Created network net1
Started container g1-c1`,
	},
}

func TestExecHomelabCmd(t *testing.T) {
	initPkgVersionInfoForTest()
	for _, test := range executeHomelabCmdTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			out, gotErr := execHomelabCmdTest(tc.ctxInfo, nil, tc.args...)
			if gotErr != nil {
				testLogErrorNotNilWithOutput(t, "execHomelabCmd()", tc.name, out, gotErr)
				return
			}

			if !testRegexMatchJoinNewLines(t, "execHomelabCmd()", tc.name, "command output", tc.want, out.String()) {
				return
			}
		})
	}
}

var executeHomelabCmdLogLevelTests = []struct {
	name     string
	args     []string
	ctxInfo  *testContextInfo
	logLevel *zzzlog.Level
}{
	{
		name: "Homelab Command - Start - All Groups - Trace Log Level",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				validImagesForPull: stringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
			}),
		},
		logLevel: newLogLevel(zzzlog.LvlTrace),
	},
	{
		name: "Homelab Command - Start - All Groups - Debug Log Level",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				validImagesForPull: stringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
			}),
		},
		logLevel: newLogLevel(zzzlog.LvlDebug),
	},
	{
		name: "Homelab Command - Start - All Groups - Info Log Level",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				validImagesForPull: stringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
			}),
		},
		logLevel: newLogLevel(zzzlog.LvlInfo),
	},
	{
		name: "Homelab Command - Start - All Groups - Warn Log Level",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				validImagesForPull: stringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
			}),
		},
		logLevel: newLogLevel(zzzlog.LvlWarn),
	},
	{
		name: "Homelab Command - Start - All Groups - Error Log Level",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				validImagesForPull: stringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
			}),
		},
		logLevel: newLogLevel(zzzlog.LvlError),
	},
	{
		name: "Homelab Command - Start - All Groups - Fatal Log Level",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				validImagesForPull: stringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
			}),
		},
		logLevel: newLogLevel(zzzlog.LvlFatal),
	},
}

func TestExecHomelabCmdLogLevel(t *testing.T) {
	for _, test := range executeHomelabCmdLogLevelTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			out, gotErr := execHomelabCmdTest(tc.ctxInfo, tc.logLevel, tc.args...)
			if gotErr != nil {
				testLogErrorNotNilWithOutput(t, "execHomelabCmd()", tc.name, out, gotErr)
				return
			}
		})
	}
}

var executeHomelabCmdErrorTests = []struct {
	name    string
	args    []string
	ctxInfo *testContextInfo
	want    string
}{
	{
		name: "Homelab Command - Missing Subcommand",
		args: []string{},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `homelab sub-command is required`,
	},
	{
		name: "Homelab Command - Show Config - Non Existing CLI Config Path",
		args: []string{
			"show-config",
			"--cli-config",
			fmt.Sprintf("%s/testdata/foobar.yaml", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `show-config failed while determining the configs path, reason: failed to open homelab CLI config file, reason: open .+/homelab/testdata/foobar\.yaml: no such file or directory`,
	},
	{
		name: "Homelab Command - Show Config - Invalid Empty CLI Config",
		args: []string{
			"show-config",
			"--cli-config",
			fmt.Sprintf("%s/testdata/cli-configs/invalid-empty-config/config.yaml", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `show-config failed while determining the configs path, reason: failed to parse homelab CLI config, reason: EOF`,
	},
	{
		name: "Homelab Command - Show Config - Invalid Garbage CLI Config",
		args: []string{
			"show-config",
			"--cli-config",
			fmt.Sprintf("%s/testdata/cli-configs/invalid-garbage-config/config.yaml", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `show-config failed while determining the configs path, reason: failed to parse homelab CLI config, reason: yaml: unmarshal errors:
  line 1: cannot unmarshal !!str ` + "`foo bar`" + ` into main.CLIConfig`,
	},
	{
		name: "Homelab Command - Show Config - Invalid CLI Config With Empty Configs Path",
		args: []string{
			"show-config",
			"--cli-config",
			fmt.Sprintf("%s/testdata/cli-configs/invalid-config-with-empty-configs-path/config.yaml", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `show-config failed while determining the configs path, reason: homelab configs path setting in homelab.configsPath is empty/unset in the homelab CLI config`,
	},
	{
		name: "Homelab Command - Show Config - Invalid CLI Config With Invalid Configs Path",
		args: []string{
			"show-config",
			"--cli-config",
			fmt.Sprintf("%s/testdata/cli-configs/invalid-config-with-invalid-configs-path/config.yaml", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `show-config failed while parsing the configs, reason: os\.Stat\(\) failed on homelab configs path, reason: stat /foo2/bar2: no such file or directory`,
	},
	{
		name: "Homelab Command - Show Config - Non Existing Configs Path",
		args: []string{
			"show-config",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/foobar", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `show-config failed while parsing the configs, reason: os\.Stat\(\) failed on homelab configs path, reason: stat .+/homelab/testdata/foobar: no such file or directory`,
	},
	{
		name: "Homelab Command - Start - No Group Flag",
		args: []string{
			"start",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `--group flag must be specified when --all-groups is false`,
	},
	{
		name: "Homelab Command - Start - Container Flag Without Group Flag",
		args: []string{
			"start",
			"--container",
			"c1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `when --all-groups is false, --group flag must be specified when specifying the --container flag`,
	},
	{
		name: "Homelab Command - Start - Group Flag With AllGroups Flag",
		args: []string{
			"start",
			"--all-groups",
			"--group",
			"g1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `--group flag cannot be specified when all-groups is true`,
	},
	{
		name: "Homelab Command - Start - Container Flag With AllGroups Flag",
		args: []string{
			"start",
			"--all-groups",
			"--container",
			"c1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `--container flag cannot be specified when all-groups is true`,
	},
	{
		name: "Homelab Command - Start - Non Existing CLI Config Path",
		args: []string{
			"start",
			"--all-groups",
			"--cli-config",
			fmt.Sprintf("%s/testdata/foobar.yaml", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `start failed while determining the configs path, reason: failed to open homelab CLI config file, reason: open .+/homelab/testdata/foobar\.yaml: no such file or directory`,
	},
	{
		name: "Homelab Command - Start - Non Existing Configs Path",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/foobar", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `start failed while parsing the configs, reason: os\.Stat\(\) failed on homelab configs path, reason: stat .+/homelab/testdata/foobar: no such file or directory`,
	},
	{
		name: "Homelab Command - Start - One Non Existing Group",
		args: []string{
			"start",
			"--group",
			"g3",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `start failed while querying containers, reason: group g3 not found`,
	},
	{
		name: "Homelab Command - Start - One Non Existing Container In Invalid Group",
		args: []string{
			"start",
			"--group",
			"g3",
			"--container",
			"c3",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `start failed while querying containers, reason: group g3 not found`,
	},
	{
		name: "Homelab Command - Start - One Non Existing Container In Valid Group",
		args: []string{
			"start",
			"--group",
			"g1",
			"--container",
			"c3",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `start failed while querying containers, reason: container {g1 c3} not found`,
	},
	{
		name: "Homelab Command - Start - Failure",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		want: `start failed for 2 containers, reason\(s\):
1 - Failed to start container g1-c1, reason:failed to pull the image abc/xyz, reason: image abc/xyz not found or invalid and cannot be pulled by the fake docker host
2 - Failed to start container g2-c3, reason:failed to pull the image abc/xyz3, reason: image abc/xyz3 not found or invalid and cannot be pulled by the fake docker host`,
	},
}

func TestExecHomelabCmdErrors(t *testing.T) {
	for _, test := range executeHomelabCmdErrorTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, gotErr := execHomelabCmdTest(tc.ctxInfo, nil, tc.args...)
			if gotErr == nil {
				testLogErrorNil(t, "execHomelabCmd()", tc.name, tc.want)
				return
			}

			if !testRegexMatch(t, "execHomelabCmd()", tc.name, "gotErr error string", tc.want, gotErr.Error()) {
				return
			}
		})
	}
}

var executeHomelabCmdEnvErrorTests = []struct {
	name    string
	args    []string
	ctxInfo *testContextInfo
	envs    testEnvMap
	want    string
}{
	{
		name: "Homelab Command - Show Config - Default CLI Config Path - Home Directory Doesn't Exist",
		args: []string{
			"show-config",
		},
		ctxInfo: &testContextInfo{},
		envs: testEnvMap{
			"HOME": "",
		},
		want: `show-config failed while determining the configs path, reason: failed to obtain the user's home directory for reading the homelab CLI config, reason: \$HOME is not defined`,
	},
	{
		name: "Homelab Command - Show Config - Default CLI Config Path Doesn't Exist",
		args: []string{
			"show-config",
		},
		ctxInfo: &testContextInfo{},
		envs: testEnvMap{
			"HOME": "/foo/bar",
		},
		want: `show-config failed while determining the configs path, reason: failed to open homelab CLI config file, reason: open /foo/bar/\.homelab/config\.yaml: no such file or directory`,
	},
	{
		name: "Homelab Command - Start - Docker Client Creation Failed",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", pwd()),
		},
		ctxInfo: &testContextInfo{},
		envs: testEnvMap{
			"DOCKER_HOST": "/var/run/foobar-docker.sock",
		},
		want: "failed to create a new docker API client, reason: unable to parse docker host `/var/run/foobar-docker\\.sock`",
	},
}

func TestExecHomelabCmdEnvErrors(t *testing.T) {
	for _, tc := range executeHomelabCmdEnvErrorTests {
		t.Run(tc.name, func(t *testing.T) {
			setTestEnv(t, tc.envs)

			_, gotErr := execHomelabCmdTest(tc.ctxInfo, nil, tc.args...)
			if gotErr == nil {
				testLogErrorNil(t, "execHomelabCmd()", tc.name, tc.want)
				return
			}

			if !testRegexMatch(t, "execHomelabCmd()", tc.name, "gotErr error string", tc.want, gotErr.Error()) {
				return
			}
		})
	}
}

func initPkgVersionInfoForTest() {
	pkgVersion = "my-pkg-version"
	pkgCommit = "my-pkg-commit"
	pkgTimestamp = "my-pkg-timestamp"
}

func execHomelabCmdTest(ctxInfo *testContextInfo, logLevel *zzzlog.Level, args ...string) (fmt.Stringer, error) {
	buf := new(bytes.Buffer)
	lvl := zzzlog.LvlInfo
	if logLevel != nil {
		lvl = *logLevel
	}
	ctxInfo.logger = newCapturingVanillaTestLogger(lvl, buf)
	ctx := newTestContext(ctxInfo)
	err := execHomelabCmd(ctx, buf, buf, args...)
	return buf, err
}
