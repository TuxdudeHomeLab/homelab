package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"
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
	multiNewLineRegex, err := regexp.Compile(`\n+`)
	if err != nil {
		panic(err)
	}
	for _, test := range executeHomelabCmdTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			out, gotErr := execHomelabCmdTest(tc.ctxInfo, tc.args...)
			if gotErr != nil {
				t.Errorf(
					"execHomelabCmd()\nTest Case: %q\nFailure: gotErr != nil\nReason: %v\n\nOut:\n%v",
					tc.name, gotErr, out)
				return
			}

			match, err := regexp.MatchString(fmt.Sprintf("^%s$", tc.want), strings.TrimSpace(multiNewLineRegex.ReplaceAllString(out, "\n")))
			if err != nil {
				t.Errorf(
					"execHomelabCmd()\nTest Case: %q\nFailure: unexpected exception while matching against gotErr error string\nReason: error = %v", tc.name, err)
				return
			}

			if !match {
				t.Errorf(
					"execHomelabCmd()\nTest Case: %q\nFailure: output did not match the want regex\nReason:\n\nOut:\n%s\nwant:\n%s\n", tc.name, out, tc.want)
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

			_, gotErr := execHomelabCmdTest(tc.ctxInfo, tc.args...)
			if gotErr == nil {
				t.Errorf(
					"execHomelabCmd()\nTest Case: %q\nFailure: gotErr == nil\nReason: want = %q",
					tc.name, tc.want)
				return
			}

			match, err := regexp.MatchString(fmt.Sprintf("^%s$", tc.want), gotErr.Error())
			if err != nil {
				t.Errorf(
					"execHomelabCmd()\nTest Case: %q\nFailure: unexpected exception while matching against gotErr error string\nReason: error = %v", tc.name, err)
				return
			}

			if !match {
				t.Errorf(
					"execHomelabCmd()\nTest Case: %q\nFailure: gotErr did not match the want regex\nReason:\n\ngotErr = %q\n\twant = %q", tc.name, gotErr, tc.want)
			}
		})
	}
}

var executeHomelabCmdOSEnvErrorTests = []struct {
	name    string
	args    []string
	ctxInfo *testContextInfo
	envs    testOSEnvMap
	want    string
}{
	{
		name: "Homelab Command - Show Config - Default CLI Config Path - Home Directory Doesn't Exist",
		args: []string{
			"show-config",
		},
		ctxInfo: &testContextInfo{},
		envs: testOSEnvMap{
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
		envs: testOSEnvMap{
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
		envs: testOSEnvMap{
			"DOCKER_HOST": "/var/run/foobar-docker.sock",
		},
		want: "failed to create a new docker API client, reason: unable to parse docker host `/var/run/foobar-docker\\.sock`",
	},
}

func TestExecHomelabCmdOSEnvErrors(t *testing.T) {
	for _, tc := range executeHomelabCmdOSEnvErrorTests {
		t.Run(tc.name, func(t *testing.T) {
			setTestEnv(tc.envs)

			_, gotErr := execHomelabCmdTest(tc.ctxInfo, tc.args...)
			if gotErr == nil {
				t.Errorf(
					"execHomelabCmd()\nTest Case: %q\nFailure: gotErr == nil\nReason: want = %q",
					tc.name, tc.want)
				return
			}

			match, err := regexp.MatchString(fmt.Sprintf("^%s$", tc.want), gotErr.Error())
			if err != nil {
				t.Errorf(
					"execHomelabCmd()\nTest Case: %q\nFailure: unexpected exception while matching against gotErr error string\nReason: error = %v", tc.name, err)
				return
			}

			if !match {
				t.Errorf(
					"execHomelabCmd()\nTest Case: %q\nFailure: gotErr did not match the want regex\nReason:\n\ngotErr = %q\n\twant = %q", tc.name, gotErr, tc.want)
			}

			t.Cleanup(func() {
				clearTestEnv(tc.envs)
			})
		})
	}
}

func initPkgVersionInfoForTest() {
	pkgVersion = "my-pkg-version"
	pkgCommit = "my-pkg-commit"
	pkgTimestamp = "my-pkg-timestamp"
}

func execHomelabCmdTest(ctxInfo *testContextInfo, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	ctxInfo.logger = newCapturingVanillaTestLogger(buf)
	ctx := newTestContext(ctxInfo)
	err := execHomelabCmd(ctx, buf, buf, args...)
	return buf.String(), err
}
