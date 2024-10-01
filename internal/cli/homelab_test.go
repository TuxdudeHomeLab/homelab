package cli

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdudehomelab/homelab/internal/cli/version"
	"github.com/tuxdudehomelab/homelab/internal/docker/fakedocker"
	"github.com/tuxdudehomelab/homelab/internal/testhelpers"
	"github.com/tuxdudehomelab/homelab/internal/testutils"
	"github.com/tuxdudehomelab/homelab/internal/utils"
)

var executeHomelabCmdTests = []struct {
	name    string
	args    []string
	ctxInfo *testutils.TestContextInfo
	want    string
}{
	{
		name: "Homelab Command - Show Version",
		args: []string{
			"--version",
		},
		ctxInfo: &testutils.TestContextInfo{
			Version:    version.NewVersionInfo("my-pkg-version", "my-pkg-commit", "my-pkg-timestamp"),
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `homelab version my-pkg-version \[Revision: my-pkg-commit @ my-pkg-timestamp\]`,
	},
	{
		name: "Homelab Command - Show Help",
		args: []string{
			"--help",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
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
			fmt.Sprintf("%s/testdata/show-config-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Homelab config:
{
  "global": {
    "baseDir": "testdata/dummy-base-dir"
  },
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
			fmt.Sprintf("%s/testdata/cli-configs/show-config-cmd/config.yaml", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Homelab config:
{
  "global": {
    "baseDir": "testdata/dummy-base-dir"
  },
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
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{},
			}),
			UseRealHostInfo: true,
		},
		want: `Container g1-c1 not allowed to run on host [^\s]+
Container g1-c2 not allowed to run on host [^\s]+
Container g2-c3 not allowed to run on host [^\s]+`,
	},
	{
		name: "Homelab Command - Start - All Groups With Real User Info",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
			}),
			UseRealUserInfo: true,
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
		name: "Homelab Command - Start - All Groups",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
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
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
				WarnContainerCreate: utils.StringSet{
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
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
				WarnNetworkCreate: utils.StringSet{
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
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ExistingImages: utils.StringSet{
					"abc/xyz3": {},
				},
				ValidImagesForPull: utils.StringSet{
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
			fmt.Sprintf("%s/testdata/start-cmd-with-multiple-same-order-containers", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
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
			fmt.Sprintf("%s/testdata/start-cmd-with-no-network-endpoints-containers", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
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
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
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
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
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
	for _, test := range executeHomelabCmdTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			out, gotErr := execHomelabCmdTest(tc.ctxInfo, nil, tc.args...)
			if gotErr != nil {
				testhelpers.LogErrorNotNilWithOutput(t, "Exec()", tc.name, out, gotErr)
				return
			}

			if !testhelpers.RegexMatchJoinNewLines(t, "Exec()", tc.name, "command output", tc.want, out.String()) {
				return
			}
		})
	}
}

var executeHomelabCmdLogLevelTests = []struct {
	name     string
	args     []string
	ctxInfo  *testutils.TestContextInfo
	logLevel *zzzlog.Level
}{
	{
		name: "Homelab Command - Start - All Groups - Trace Log Level",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
			}),
		},
		logLevel: testhelpers.NewLogLevel(zzzlog.LvlTrace),
	},
	{
		name: "Homelab Command - Start - All Groups - Debug Log Level",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
			}),
		},
		logLevel: testhelpers.NewLogLevel(zzzlog.LvlDebug),
	},
	{
		name: "Homelab Command - Start - All Groups - Info Log Level",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
			}),
		},
		logLevel: testhelpers.NewLogLevel(zzzlog.LvlInfo),
	},
	{
		name: "Homelab Command - Start - All Groups - Warn Log Level",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
			}),
		},
		logLevel: testhelpers.NewLogLevel(zzzlog.LvlWarn),
	},
	{
		name: "Homelab Command - Start - All Groups - Error Log Level",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
			}),
		},
		logLevel: testhelpers.NewLogLevel(zzzlog.LvlError),
	},
	{
		name: "Homelab Command - Start - All Groups - Fatal Log Level",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
			}),
		},
		logLevel: testhelpers.NewLogLevel(zzzlog.LvlFatal),
	},
}

func TestExecHomelabCmdLogLevel(t *testing.T) {
	for _, test := range executeHomelabCmdLogLevelTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			out, gotErr := execHomelabCmdTest(tc.ctxInfo, tc.logLevel, tc.args...)
			if gotErr != nil {
				testhelpers.LogErrorNotNilWithOutput(t, "Exec()", tc.name, out, gotErr)
				return
			}
		})
	}
}

var executeHomelabCmdErrorTests = []struct {
	name    string
	args    []string
	ctxInfo *testutils.TestContextInfo
	want    string
}{
	{
		name: "Homelab Command - Missing Subcommand",
		args: []string{},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `homelab sub-command is required`,
	},
	{
		name: "Homelab Command - Show Config - Non Existing CLI Config Path",
		args: []string{
			"show-config",
			"--cli-config",
			fmt.Sprintf("%s/testdata/foobar.yaml", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `show-config failed while determining the configs path, reason: failed to open homelab CLI config file, reason: open .+/testdata/foobar\.yaml: no such file or directory`,
	},
	{
		name: "Homelab Command - Show Config - Invalid Empty CLI Config",
		args: []string{
			"show-config",
			"--cli-config",
			fmt.Sprintf("%s/testdata/cli-configs/invalid-empty-config/config.yaml", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `show-config failed while determining the configs path, reason: failed to parse homelab CLI config, reason: EOF`,
	},
	{
		name: "Homelab Command - Show Config - Invalid Garbage CLI Config",
		args: []string{
			"show-config",
			"--cli-config",
			fmt.Sprintf("%s/testdata/cli-configs/invalid-garbage-config/config.yaml", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `show-config failed while determining the configs path, reason: failed to parse homelab CLI config, reason: yaml: unmarshal errors:
  line 1: cannot unmarshal !!str ` + "`foo bar`" + ` into cliconfig.CLIConfig`,
	},
	{
		name: "Homelab Command - Show Config - Invalid CLI Config With Empty Configs Path",
		args: []string{
			"show-config",
			"--cli-config",
			fmt.Sprintf("%s/testdata/cli-configs/invalid-config-with-empty-configs-path/config.yaml", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `show-config failed while determining the configs path, reason: homelab configs path setting in homelab.configsPath is empty/unset in the homelab CLI config`,
	},
	{
		name: "Homelab Command - Show Config - Invalid CLI Config With Invalid Configs Path",
		args: []string{
			"show-config",
			"--cli-config",
			fmt.Sprintf("%s/testdata/cli-configs/invalid-config-with-invalid-configs-path/config.yaml", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `show-config failed while parsing the configs, reason: os\.Stat\(\) failed on homelab configs path, reason: stat /foo2/bar2: no such file or directory`,
	},
	{
		name: "Homelab Command - Show Config - Non Existing Configs Path",
		args: []string{
			"show-config",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/foobar", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `show-config failed while parsing the configs, reason: os\.Stat\(\) failed on homelab configs path, reason: stat .+/testdata/foobar: no such file or directory`,
	},
	{
		name: "Homelab Command - Start - No Group Flag",
		args: []string{
			"start",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
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
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
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
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
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
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `--container flag cannot be specified when all-groups is true`,
	},
	{
		name: "Homelab Command - Start - Non Existing CLI Config Path",
		args: []string{
			"start",
			"--all-groups",
			"--cli-config",
			fmt.Sprintf("%s/testdata/foobar.yaml", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `start failed while determining the configs path, reason: failed to open homelab CLI config file, reason: open .+/testdata/foobar\.yaml: no such file or directory`,
	},
	{
		name: "Homelab Command - Start - Non Existing Configs Path",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/foobar", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `start failed while parsing the configs, reason: os\.Stat\(\) failed on homelab configs path, reason: stat .+/testdata/foobar: no such file or directory`,
	},
	{
		name: "Homelab Command - Start - One Non Existing Group",
		args: []string{
			"start",
			"--group",
			"g3",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
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
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
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
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `start failed while querying containers, reason: container {g1 c3} not found`,
	},
	{
		name: "Homelab Command - Start - Failure",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
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
				testhelpers.LogErrorNil(t, "Exec()", tc.name, tc.want)
				return
			}

			if !testhelpers.RegexMatch(t, "Exec()", tc.name, "gotErr error string", tc.want, gotErr.Error()) {
				return
			}
		})
	}
}

var executeHomelabCmdEnvErrorTests = []struct {
	name    string
	args    []string
	ctxInfo *testutils.TestContextInfo
	envs    testhelpers.TestEnvMap
	want    string
}{
	{
		name: "Homelab Command - Show Config - Default CLI Config Path - Home Directory Doesn't Exist",
		args: []string{
			"show-config",
		},
		ctxInfo: &testutils.TestContextInfo{},
		envs: testhelpers.TestEnvMap{
			"HOME": "",
		},
		want: `show-config failed while determining the configs path, reason: failed to obtain the user's home directory for reading the homelab CLI config, reason: \$HOME is not defined`,
	},
	{
		name: "Homelab Command - Show Config - Default CLI Config Path Doesn't Exist",
		args: []string{
			"show-config",
		},
		ctxInfo: &testutils.TestContextInfo{},
		envs: testhelpers.TestEnvMap{
			"HOME": "/foo/bar",
		},
		want: `show-config failed while determining the configs path, reason: failed to open homelab CLI config file, reason: open /foo/bar/\.homelab/config\.yaml: no such file or directory`,
	},
}

func TestExecHomelabCmdEnvErrors(t *testing.T) {
	for _, tc := range executeHomelabCmdEnvErrorTests {
		t.Run(tc.name, func(t *testing.T) {
			testhelpers.SetTestEnv(t, tc.envs)

			_, gotErr := execHomelabCmdTest(tc.ctxInfo, nil, tc.args...)
			if gotErr == nil {
				testhelpers.LogErrorNil(t, "Exec()", tc.name, tc.want)
				return
			}

			if !testhelpers.RegexMatch(t, "Exec()", tc.name, "gotErr error string", tc.want, gotErr.Error()) {
				return
			}
		})
	}
}

var executeHomelabCmdEnvPanicTests = []struct {
	name    string
	args    []string
	ctxInfo *testutils.TestContextInfo
	envs    testhelpers.TestEnvMap
	want    string
}{
	{
		name: "Homelab Command - Start - Docker Client Creation Failed",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{},
		envs: testhelpers.TestEnvMap{
			"DOCKER_HOST": "/var/run/foobar-docker.sock",
		},
		want: "Failed to create a new docker API client, reason: unable to parse docker host `/var/run/foobar-docker\\.sock`",
	},
}

func TestExecHomelabCmdEnvPanics(t *testing.T) {
	for _, tc := range executeHomelabCmdEnvPanicTests {
		t.Run(tc.name, func(t *testing.T) {
			testhelpers.SetTestEnv(t, tc.envs)

			buf := new(bytes.Buffer)
			defer testhelpers.ExpectPanicWithOutput(t, "Exec()", tc.name, buf, tc.want)
			_, _ = execHomelabCmdTestWithBuf(tc.ctxInfo, nil, buf, tc.args...)
		})
	}
}

func execHomelabCmdTest(ctxInfo *testutils.TestContextInfo, logLevel *zzzlog.Level, args ...string) (fmt.Stringer, error) {
	buf := new(bytes.Buffer)
	return execHomelabCmdTestWithBuf(ctxInfo, logLevel, buf, args...)
}

func execHomelabCmdTestWithBuf(ctxInfo *testutils.TestContextInfo, logLevel *zzzlog.Level, buf *bytes.Buffer, args ...string) (fmt.Stringer, error) {
	lvl := zzzlog.LvlInfo
	if logLevel != nil {
		lvl = *logLevel
	}
	ctxInfo.Logger = testutils.NewCapturingVanillaTestLogger(lvl, buf)
	ctx := testutils.NewTestContext(ctxInfo)
	err := Exec(ctx, buf, buf, args...)
	return buf, err
}
