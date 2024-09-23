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
  "Global": {
    "Env": null,
    "MountDefs": null,
    "Container": {
      "StopSignal": "",
      "StopTimeout": 0,
      "RestartPolicy": {
        "Mode": "",
        "MaxRetryCount": 0
      },
      "DomainName": "",
      "DNSSearch": null,
      "Env": null,
      "Mounts": null,
      "Labels": null
    }
  },
  "IPAM": {
    "Networks": {
      "BridgeModeNetworks": \[
        {
          "Name": "net1",
          "HostInterfaceName": "docker-net1",
          "CIDR": "172\.18\.100\.0/24",
          "Priority": 1,
          "Containers": \[
            {
              "IP": "172\.18\.100\.11",
              "Container": {
                "Group": "g1",
                "Container": "c1"
              }
            },
            {
              "IP": "172\.18\.100\.12",
              "Container": {
                "Group": "g1",
                "Container": "c2"
              }
            }
          \]
        },
        {
          "Name": "net2",
          "HostInterfaceName": "docker-net2",
          "CIDR": "172\.18\.101\.0/24",
          "Priority": 1,
          "Containers": \[
            {
              "IP": "172\.18\.101\.21",
              "Container": {
                "Group": "g2",
                "Container": "c3"
              }
            }
          \]
        }
      \],
      "ContainerModeNetworks": null
    }
  },
  "Hosts": \[
    {
      "Name": "fakehost",
      "AllowedContainers": \[
        {
          "Group": "g1",
          "Container": "c1"
        }
      \]
    },
    {
      "Name": "host2",
      "AllowedContainers": null
    }
  \],
  "Groups": \[
    {
      "Name": "g1",
      "Order": 1
    },
    {
      "Name": "g2",
      "Order": 2
    }
  \],
  "Containers": \[
    {
      "Info": {
        "Group": "g1",
        "Container": "c1"
      },
      "Config": {
        "Env": null
      },
      "Image": {
        "Image": "abc/xyz",
        "SkipImagePull": false,
        "IgnoreImagePullFailures": false,
        "PullImageBeforeStop": false
      },
      "Metadata": {
        "Labels": null
      },
      "Lifecycle": {
        "Order": 1,
        "StartPreHook": "",
        "RestartPolicy": {
          "Mode": "",
          "MaxRetryCount": 0
        },
        "AutoRemove": false,
        "StopSignal": "",
        "StopTimeout": 0
      },
      "User": {
        "User": "",
        "PrimaryGroup": "",
        "AdditionalGroups": null
      },
      "Filesystem": {
        "ReadOnlyRootfs": false,
        "Mounts": null,
        "Devices": null
      },
      "Network": {
        "HostName": "",
        "DomainName": "",
        "DNSServers": null,
        "DNSOptions": null,
        "DNSSearch": null,
        "PublishedPorts": null
      },
      "Security": {
        "Privileged": false,
        "Sysctls": null,
        "CapAdd": null,
        "CapDrop": null
      },
      "Health": {
        "Cmd": null,
        "Retries": 0,
        "Interval": "",
        "Timeout": "",
        "StartPeriod": "",
        "StartInterval": ""
      },
      "Runtime": {
        "AttachToTty": false,
        "ShmSize": "",
        "Env": null,
        "Entrypoint": null,
        "Args": null
      }
    },
    {
      "Info": {
        "Group": "g1",
        "Container": "c2"
      },
      "Config": {
        "Env": null
      },
      "Image": {
        "Image": "abc/xyz2",
        "SkipImagePull": false,
        "IgnoreImagePullFailures": false,
        "PullImageBeforeStop": false
      },
      "Metadata": {
        "Labels": null
      },
      "Lifecycle": {
        "Order": 2,
        "StartPreHook": "",
        "RestartPolicy": {
          "Mode": "",
          "MaxRetryCount": 0
        },
        "AutoRemove": false,
        "StopSignal": "",
        "StopTimeout": 0
      },
      "User": {
        "User": "",
        "PrimaryGroup": "",
        "AdditionalGroups": null
      },
      "Filesystem": {
        "ReadOnlyRootfs": false,
        "Mounts": null,
        "Devices": null
      },
      "Network": {
        "HostName": "",
        "DomainName": "",
        "DNSServers": null,
        "DNSOptions": null,
        "DNSSearch": null,
        "PublishedPorts": null
      },
      "Security": {
        "Privileged": false,
        "Sysctls": null,
        "CapAdd": null,
        "CapDrop": null
      },
      "Health": {
        "Cmd": null,
        "Retries": 0,
        "Interval": "",
        "Timeout": "",
        "StartPeriod": "",
        "StartInterval": ""
      },
      "Runtime": {
        "AttachToTty": false,
        "ShmSize": "",
        "Env": null,
        "Entrypoint": null,
        "Args": null
      }
    },
    {
      "Info": {
        "Group": "g2",
        "Container": "c3"
      },
      "Config": {
        "Env": null
      },
      "Image": {
        "Image": "abc/xyz3",
        "SkipImagePull": false,
        "IgnoreImagePullFailures": false,
        "PullImageBeforeStop": false
      },
      "Metadata": {
        "Labels": null
      },
      "Lifecycle": {
        "Order": 1,
        "StartPreHook": "",
        "RestartPolicy": {
          "Mode": "",
          "MaxRetryCount": 0
        },
        "AutoRemove": false,
        "StopSignal": "",
        "StopTimeout": 0
      },
      "User": {
        "User": "",
        "PrimaryGroup": "",
        "AdditionalGroups": null
      },
      "Filesystem": {
        "ReadOnlyRootfs": false,
        "Mounts": null,
        "Devices": null
      },
      "Network": {
        "HostName": "",
        "DomainName": "",
        "DNSServers": null,
        "DNSOptions": null,
        "DNSSearch": null,
        "PublishedPorts": null
      },
      "Security": {
        "Privileged": false,
        "Sysctls": null,
        "CapAdd": null,
        "CapDrop": null
      },
      "Health": {
        "Cmd": null,
        "Retries": 0,
        "Interval": "",
        "Timeout": "",
        "StartPeriod": "",
        "StartInterval": ""
      },
      "Runtime": {
        "AttachToTty": false,
        "ShmSize": "",
        "Env": null,
        "Entrypoint": null,
        "Args": null
      }
    }
  \]
}`,
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
					"abc/xyz": {},
				},
			}),
		},
		want: `Pulling image: abc/xyz
Created network net1
Started container g1-c1
Container g1-c2 not allowed to run on host FakeHost
Container g2-c3 not allowed to run on host FakeHost`,
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
					"execHomelabCmd()\nTest Case: %q\nFailure: gotErr != nil\nReason: %v\nOutput: %v",
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
					"execHomelabCmd()\nTest Case: %q\nFailure: output did not match the want regex\nReason:\n\nout:\n%s\nwant:\n%s\n", tc.name, out, tc.want)
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

func initPkgVersionInfoForTest() {
	pkgVersion = "my-pkg-version"
	pkgCommit = "my-pkg-commit"
	pkgTimestamp = "my-pkg-timestamp"
}

func execHomelabCmdTest(ctxInfo *testContextInfo, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	ctxInfo.logger = newCapturingTestLogger(buf)
	ctx := newTestContext(ctxInfo)
	err := execHomelabCmd(ctx, buf, buf, args...)
	return buf.String(), err
}
