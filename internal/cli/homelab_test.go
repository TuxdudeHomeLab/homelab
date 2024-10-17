package cli

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdudehomelab/homelab/internal/cli/version"
	"github.com/tuxdudehomelab/homelab/internal/cmdexec/fakecmdexec"
	"github.com/tuxdudehomelab/homelab/internal/docker"
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
			"config",
			"show",
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
			"config",
			"show",
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
		name: "Homelab Command - Groups Start - All Groups With Real Host Info",
		args: []string{
			"groups",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
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
		name: "Homelab Command - Groups Start - All Groups With Real User Info",
		args: []string{
			"groups",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
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
Creating container g1-c1
Starting container g1-c1
Container g1-c2 not allowed to run on host FakeHost
Pulling image: abc/xyz3
Created network net2
Creating container g2-c3
Starting container g2-c3`,
	},
	{
		name: "Homelab Command - Groups Start - All Groups",
		args: []string{
			"groups",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
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
Creating container g1-c1
Starting container g1-c1
Container g1-c2 not allowed to run on host FakeHost
Pulling image: abc/xyz3
Created network net2
Creating container g2-c3
Starting container g2-c3`,
	},
	{
		name: "Homelab Command - Groups Start - All Groups - Container Create Warning",
		args: []string{
			"groups",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
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
Creating container g1-c1
Warnings encountered while creating the container g1-c1
1 - first warning generated during container create for g1-c1 on the fake docker host
2 - second warning generated during container create for g1-c1 on the fake docker host
3 - third warning generated during container create for g1-c1 on the fake docker host
Starting container g1-c1
Container g1-c2 not allowed to run on host FakeHost
Pulling image: abc/xyz3
Created network net2
Creating container g2-c3
Starting container g2-c3`,
	},
	{
		name: "Homelab Command - Groups Start - All Groups - Network Create Warning",
		args: []string{
			"groups",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
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
Creating container g1-c1
Starting container g1-c1
Container g1-c2 not allowed to run on host FakeHost
Pulling image: abc/xyz3
Created network net2
Creating container g2-c3
Starting container g2-c3`,
	},
	{
		name: "Homelab Command - Groups Start - All Groups - One Existing Image",
		args: []string{
			"groups",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
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
Creating container g1-c1
Starting container g1-c1
Container g1-c2 not allowed to run on host FakeHost
Pulled newer version of image abc/xyz3: [a-z0-9]{64}
Created network net2
Creating container g2-c3
Starting container g2-c3`,
	},
	{
		name: "Homelab Command - Groups Start - All Groups With Multiple Same Order Containers",
		args: []string{
			"groups",
			"start",
			"all",
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
Creating container g1-c1
Starting container g1-c1
Container g1-c2 not allowed to run on host FakeHost
Pulling image: abc/xyz3
Creating container g1-c3
Starting container g1-c3
Pulling image: abc/xyz4
Created network net2
Creating container g2-c4
Starting container g2-c4`,
	},
	{
		name: "Homelab Command - Groups Start - All Groups With No Network Endpoints Containers",
		args: []string{
			"groups",
			"start",
			"all",
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
Creating container g1-c1
Starting container g1-c1
Container g1-c2 not allowed to run on host FakeHost
Pulling image: abc/xyz3
Container g1-c3 has no network endpoints configured, this is uncommon!
Creating container g1-c3
Starting container g1-c3
Pulling image: abc/xyz4
Created network net2
Creating container g2-c4
Starting container g2-c4`,
	},
	{
		name: "Homelab Command - Groups Start - All Groups - One Container With Start Pre-Hook",
		args: []string{
			"groups",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-cmd-with-start-pre-hook", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			Executor: fakecmdexec.NewFakeExecutor(&fakecmdexec.FakeExecutorInitInfo{
				ValidCmds: []fakecmdexec.FakeValidCmdInfo{
					{
						Cmd: []string{
							"custom-start-prehook",
							"arg1",
							"arg2",
						},
						Output: "Output from a custom start prehook",
					},
				},
			}),
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
			}),
		},
		want: `Pulling image: abc/xyz
Created network net1
Creating container g1-c1
Starting container g1-c1
Container g1-c2 not allowed to run on host FakeHost
Output from start pre-hook for container g2-c3 >>>
Output from a custom start prehook
Pulling image: abc/xyz3
Created network net2
Creating container g2-c3
Starting container g2-c3`,
	},
	{
		name: "Homelab Command - Groups Start - All Groups - One Container With Ignore Image Pull Failures",
		args: []string{
			"groups",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-stop-cmds-with-ignore-image-pull-failures", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
				FailImagePull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Pulling image: abc/xyz
Ignoring - Image pull for container g1-c1 failed, reason: failed while pulling the image abc/xyz, reason: failed to pull image abc/xyz on the fake docker host
Created network net1
Creating container g1-c1
Starting container g1-c1
Container g1-c2 not allowed to run on host FakeHost
Pulling image: abc/xyz3
Created network net2
Creating container g2-c3
Starting container g2-c3`,
	},
	{
		name: "Homelab Command - Groups Start - One Group",
		args: []string{
			"groups",
			"start",
			"g1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
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
Creating container g1-c1
Starting container g1-c1
Container g1-c2 not allowed to run on host FakeHost`,
	},
	{
		name: "Homelab Command - Containers Start - One Container",
		args: []string{
			"containers",
			"start",
			"g1/c1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
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
Creating container g1-c1
Starting container g1-c1`,
	},
	{
		name: "Homelab Command - Groups Stop - All Groups",
		args: []string{
			"groups",
			"stop",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateCreated,
					},
					{
						Name:  "g2-c3",
						Image: "abc/xyz3",
						State: docker.ContainerStateRemoving,
					},
				},
			}),
		},
		want: `Container g1-c1 cannot be stopped since it is in state Created
Container g1-c2 cannot be stopped since it was not found
Container g2-c3 cannot be stopped since it is in state Removing`,
	},
	{
		name: "Homelab Command - Groups Stop - All Groups - One Container With Ignore Image Pull Failures",
		args: []string{
			"groups",
			"stop",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/start-stop-cmds-with-ignore-image-pull-failures", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRunning,
					},
					{
						Name:  "g2-c3",
						Image: "abc/xyz3",
						State: docker.ContainerStateRunning,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
				FailImagePull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Pulling image: abc/xyz
Ignoring - Image pull for container g1-c1 failed, reason: failed while pulling the image abc/xyz, reason: failed to pull image abc/xyz on the fake docker host
Stopping container g1-c1
Container g1-c2 cannot be stopped since it was not found
Stopping container g2-c3`,
	},
	{
		name: "Homelab Command - Groups Stop - One Group",
		args: []string{
			"groups",
			"stop",
			"g1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRunning,
					},
				},
			}),
		},
		want: `Stopping container g1-c1
Container g1-c2 cannot be stopped since it was not found`,
	},
	{
		name: "Homelab Command - Containers Stop - One Container - Not Found",
		args: []string{
			"containers",
			"stop",
			"g1/c1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Container g1-c1 cannot be stopped since it was not found`,
	},
	{
		name: "Homelab Command - Groups Purge - All Groups",
		args: []string{
			"groups",
			"purge",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateCreated,
					},
					{
						Name:  "g1-c2",
						Image: "abc/xyz2",
						State: docker.ContainerStatePaused,
					},
				},
			}),
		},
		want: `Removing container g1-c1
Stopping container g1-c2
Removing container g1-c2
Container g2-c3 cannot be purged since it was not found`,
	},
	{
		name: "Homelab Command - Groups Purge - One Group",
		args: []string{
			"groups",
			"purge",
			"g1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRunning,
					},
				},
			}),
		},
		want: `Stopping container g1-c1
Removing container g1-c1
Container g1-c2 cannot be purged since it was not found`,
	},
	{
		name: "Homelab Command - Containers Purge - One Container - Not Found",
		args: []string{
			"containers",
			"purge",
			"g1/c1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Container g1-c1 cannot be purged since it was not found`,
	},
	{
		name: "Homelab Command - Networks Create - One Network",
		args: []string{
			"networks",
			"create",
			"net1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/networks-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Created network net1`,
	},
	{
		name: "Homelab Command - Networks Create - One Network - Warning",
		args: []string{
			"networks",
			"create",
			"net1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/networks-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				WarnNetworkCreate: utils.StringSet{
					"net1": {},
				},
			}),
		},
		want: `Warning encountered while creating the network net1
warning generated during network create for network net1 on the fake docker host
Created network net1`,
	},
	{
		name: "Homelab Command - Networks Create - One Network - Exists Already",
		args: []string{
			"networks",
			"create",
			"net1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/networks-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Networks: []*fakedocker.FakeNetworkInitInfo{
					{
						Name: "net1",
					},
				},
			}),
		},
		want: `Network net1 not created since it already exists`,
	},
	{
		name: "Homelab Command - Networks Delete - One Network - Network Doesn't Exist",
		args: []string{
			"networks",
			"delete",
			"net1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/networks-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Network net1 not deleted since it doesn't exist already`,
	},
	{
		name: "Homelab Command - Networks Delete - One Network - Network Exists",
		args: []string{
			"networks",
			"delete",
			"net1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/networks-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Networks: []*fakedocker.FakeNetworkInitInfo{
					{
						Name: "net1",
					},
				},
			}),
		},
		want: `Deleted network net1`,
	},
}

func TestExecHomelabCmd(t *testing.T) {
	t.Parallel()

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

var executeHomelabCmdRealEverythingTests = []struct {
	name string
	args []string
	want string
}{
	{
		name: "Homelab Command - Show Config - Real Everything",
		args: []string{
			"config",
			"show",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/show-config-cmd", testhelpers.Pwd()),
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
		name: "Homelab Command - Groups Start - All Groups - Real Everything",
		args: []string{
			"groups",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		want: `Container g1-c1 not allowed to run on host [^\s]+
Container g1-c2 not allowed to run on host [^\s]+
Container g2-c3 not allowed to run on host [^\s]+`,
	},
	{
		name: "Homelab Command - Groups Stop - All Groups - Real Everything",
		args: []string{
			"groups",
			"stop",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		want: `Container g1-c1 cannot be stopped since it was not found
Container g1-c2 cannot be stopped since it was not found
Container g2-c3 cannot be stopped since it was not found`,
	},
	{
		name: "Homelab Command - Groups Purge - All Groups - Real Everything",
		args: []string{
			"groups",
			"purge",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		want: `Container g1-c1 cannot be purged since it was not found
Container g1-c2 cannot be purged since it was not found
Container g2-c3 cannot be purged since it was not found`,
	},
}

func TestExecHomelabCmdRealEverything(t *testing.T) {
	t.Parallel()

	for _, test := range executeHomelabCmdRealEverythingTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			out, gotErr := execHomelabCmdTest(
				&testutils.TestContextInfo{
					UseRealHostInfo: true,
					UseRealUserInfo: true,
					UseRealExecutor: true,
				},
				nil,
				tc.args...,
			)
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

var executeHomelabCmdLogLevels = []zzzlog.Level{
	zzzlog.LvlTrace,
	zzzlog.LvlDebug,
	zzzlog.LvlInfo,
	zzzlog.LvlWarn,
	zzzlog.LvlError,
	zzzlog.LvlFatal,
}

var executeHomelabCmdLogLevelTests = []struct {
	name    string
	args    []string
	ctxInfo func() *testutils.TestContextInfo
}{
	{
		name: "Homelab Command - Groups Start - All Groups",
		args: []string{
			"groups",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
					ValidImagesForPull: utils.StringSet{
						"abc/xyz":  {},
						"abc/xyz3": {},
					},
				}),
			}
		},
	},
	{
		name: "Homelab Command - Groups Stop - All Groups",
		args: []string{
			"groups",
			"stop",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
					Containers: []*fakedocker.FakeContainerInitInfo{
						{
							Name:  "g1-c1",
							Image: "abc/xyz",
							State: docker.ContainerStateCreated,
						},
						{
							Name:  "g1-c2",
							Image: "abc/xyz2",
							State: docker.ContainerStateRunning,
						},
						{
							Name:  "g2-c3",
							Image: "abc/xyz3",
							State: docker.ContainerStateRestarting,
						},
					},
				}),
			}
		},
	},
	{
		name: "Homelab Command - Groups Purge - All Groups",
		args: []string{
			"groups",
			"purge",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
					Containers: []*fakedocker.FakeContainerInitInfo{
						{
							Name:  "g1-c1",
							Image: "abc/xyz",
							State: docker.ContainerStateCreated,
						},
						{
							Name:  "g1-c2",
							Image: "abc/xyz2",
							State: docker.ContainerStateRunning,
						},
						{
							Name:  "g2-c3",
							Image: "abc/xyz3",
							State: docker.ContainerStateDead,
						},
					},
				}),
			}
		},
	},
	{
		name: "Homelab Command - Networks Create",
		args: []string{
			"networks",
			"create",
			"net1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/networks-cmd", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
	},
	{
		name: "Homelab Command - Networks Delete",
		args: []string{
			"networks",
			"delete",
			"net1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/networks-cmd", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
					Networks: []*fakedocker.FakeNetworkInitInfo{
						{
							Name: "net1",
						},
					},
				}),
			}
		},
	},
}

func TestExecHomelabCmdLogLevel(t *testing.T) {
	t.Parallel()

	for _, test := range executeHomelabCmdLogLevelTests {
		tc := test
		for _, l := range executeHomelabCmdLogLevels {
			lvl := l
			tcName := fmt.Sprintf("%s - %v Log Level", tc.name, lvl)
			t.Run(tcName, func(t *testing.T) {
				t.Parallel()

				out, gotErr := execHomelabCmdTest(tc.ctxInfo(), &lvl, tc.args...)
				if gotErr != nil {
					testhelpers.LogErrorNotNilWithOutput(t, "Exec()", tcName, out, gotErr)
					return
				}
			})
		}
	}
}

var executeHomelabCmdErrorTests = []struct {
	name    string
	args    []string
	ctxInfo *testutils.TestContextInfo
	want    string
}{
	{
		name: "Homelab Base Command - Missing Subcommand",
		args: nil,
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `homelab sub-command is required`,
	},
	{
		name: "Homelab Config Command - Missing Subcommand",
		args: []string{
			"config",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `homelab config sub-command is required`,
	},
	{
		name: "Homelab Groups Command - Missing Subcommand",
		args: []string{
			"groups",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `homelab group sub-command is required`,
	},
	{
		name: "Homelab Containers Command - Missing Subcommand",
		args: []string{
			"containers",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `homelab container sub-command is required`,
	},
	{
		name: "Homelab Networks Command - Missing Subcommand",
		args: []string{
			"networks",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `homelab networks sub-command is required`,
	},
	{
		name: "Homelab Command - Groups Start - Zero Group Name Args",
		args: []string{
			"groups",
			"start",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one group name argument to be specified, but found 0 instead`,
	},
	{
		name: "Homelab Command - Groups Stop - Zero Group Name Args",
		args: []string{
			"groups",
			"stop",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one group name argument to be specified, but found 0 instead`,
	},
	{
		name: "Homelab Command - Groups Purge - Zero Group Name Args",
		args: []string{
			"groups",
			"purge",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one group name argument to be specified, but found 0 instead`,
	},
	{
		name: "Homelab Command - Groups Start - Multiple Group Name Args",
		args: []string{
			"groups",
			"start",
			"g1",
			"g2",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one group name argument to be specified, but found 2 instead`,
	},
	{
		name: "Homelab Command - Groups Stop - Multiple Group Name Args",
		args: []string{
			"groups",
			"stop",
			"g1",
			"g2",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one group name argument to be specified, but found 2 instead`,
	},
	{
		name: "Homelab Command - Groups Purge - Multiple Group Name Args",
		args: []string{
			"groups",
			"purge",
			"g1",
			"g2",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one group name argument to be specified, but found 2 instead`,
	},
	{
		name: "Homelab Command - Containers Start - Zero Container Name Args",
		args: []string{
			"containers",
			"start",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one container name argument to be specified, but found 0 instead`,
	},
	{
		name: "Homelab Command - Containers Stop - Zero Container Name Args",
		args: []string{
			"containers",
			"stop",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one container name argument to be specified, but found 0 instead`,
	},
	{
		name: "Homelab Command - Containers Purge - Zero Container Name Args",
		args: []string{
			"containers",
			"purge",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one container name argument to be specified, but found 0 instead`,
	},
	{
		name: "Homelab Command - Containers Start - Multiple Container Name Args",
		args: []string{
			"containers",
			"start",
			"g1/c1",
			"g2/c2",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one container name argument to be specified, but found 2 instead`,
	},
	{
		name: "Homelab Command - Containers Stop - Multiple Container Name Args",
		args: []string{
			"containers",
			"stop",
			"g1/c1",
			"g2/c2",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one container name argument to be specified, but found 2 instead`,
	},
	{
		name: "Homelab Command - Containers Purge - Multiple Container Name Args",
		args: []string{
			"containers",
			"purge",
			"g1/c1",
			"g2/c2",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one container name argument to be specified, but found 2 instead`,
	},
	{
		name: "Homelab Command - Containers Start - Invalid Container Name",
		args: []string{
			"containers",
			"start",
			"foobar",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Container name must be specified in the form 'group/container'`,
	},
	{
		name: "Homelab Command - Containers Stop - Invalid Container Name",
		args: []string{
			"containers",
			"stop",
			"foobar",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Container name must be specified in the form 'group/container'`,
	},
	{
		name: "Homelab Command - Containers Purge - Invalid Container Name",
		args: []string{
			"containers",
			"purge",
			"foobar",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Container name must be specified in the form 'group/container'`,
	},
	{
		name: "Homelab Command - Groups Start - Failure",
		args: []string{
			"groups",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `groups start failed for 2 containers, reason\(s\):
1 - Failed to start container g1-c1, reason:failed to pull the image abc/xyz, reason: image abc/xyz not found or invalid and cannot be pulled by the fake docker host
2 - Failed to start container g2-c3, reason:failed to pull the image abc/xyz3, reason: image abc/xyz3 not found or invalid and cannot be pulled by the fake docker host`,
	},
	{
		name: "Homelab Command - Groups Stop - Failure",
		args: []string{
			"groups",
			"stop",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRunning,
					},
					{
						Name:  "g2-c3",
						Image: "abc/xyz3",
						State: docker.ContainerStateRestarting,
					},
				},
				FailContainerStop: utils.StringSet{
					"g1-c1": {},
					"g2-c3": {},
				},
			}),
		},
		want: `groups stop failed for 2 containers, reason\(s\):
1 - Failed to stop container g1-c1, reason:failed to stop the container, reason: failed to stop container g1-c1 on the fake docker host
2 - Failed to stop container g2-c3, reason:failed to stop the container, reason: failed to stop container g2-c3 on the fake docker host`,
	},
	{
		name: "Homelab Command - Groups Purge - Failure",
		args: []string{
			"groups",
			"purge",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRunning,
					},
					{
						Name:  "g2-c3",
						Image: "abc/xyz3",
						State: docker.ContainerStateRemoving,
					},
				},
				FailContainerStop: utils.StringSet{
					"g1-c1": {},
				},
			}),
		},
		want: `groups purge failed for 2 containers, reason\(s\):
1 - Failed to purge container g1-c1, reason:failed to stop the container, reason: failed to stop container g1-c1 on the fake docker host
2 - Failed to purge container g2-c3, reason:failed to purge container g2-c3 after 6 attempts`,
	},
	{
		name: "Homelab Command - Networks Create - Zero Network Name Args",
		args: []string{
			"networks",
			"create",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one network name argument to be specified, but found 0 instead`,
	},
	{
		name: "Homelab Command - Networks Create - Multiple Network Name Args",
		args: []string{
			"networks",
			"create",
			"net1",
			"net2",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one network name argument to be specified, but found 2 instead`,
	},
	{
		name: "Homelab Command - Networks Create - Invalid Network Name",
		args: []string{
			"networks",
			"create",
			"net11",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/networks-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `networks create failed while querying networks, reason: network net11 not found`,
	},
	{
		name: "Homelab Command - Networks Create - Container Mode Network",
		args: []string{
			"networks",
			"create",
			"net3",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/networks-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `networks create failed for 1 networks, reason\(s\):
1 - container mode network net3 cannot be created`,
	},
	{
		name: "Homelab Command - Networks Create - Failure",
		args: []string{
			"networks",
			"create",
			"net1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/networks-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				FailNetworkCreate: utils.StringSet{
					"net1": {},
				},
			}),
		},
		want: `networks create failed for 1 networks, reason\(s\):
1 - failed to create the network, reason: failed to create network net1 on the fake docker host`,
	},
	{
		name: "Homelab Command - Networks Delete - Zero Network Name Args",
		args: []string{
			"networks",
			"delete",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one network name argument to be specified, but found 0 instead`,
	},
	{
		name: "Homelab Command - Networks Delete - Multiple Network Name Args",
		args: []string{
			"networks",
			"delete",
			"net1",
			"net2",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one network name argument to be specified, but found 2 instead`,
	},
	{
		name: "Homelab Command - Networks Delete - Invalid Network Name",
		args: []string{
			"networks",
			"delete",
			"net11",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/networks-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `networks delete failed while querying networks, reason: network net11 not found`,
	},
	{
		name: "Homelab Command - Networks Delete - Container Mode Network",
		args: []string{
			"networks",
			"delete",
			"net3",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/networks-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `networks delete failed for 1 networks, reason\(s\):
1 - container mode network net3 cannot be deleted`,
	},
	{
		name: "Homelab Command - Networks Delete - Failure",
		args: []string{
			"networks",
			"delete",
			"net1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/networks-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Networks: []*fakedocker.FakeNetworkInitInfo{
					{
						Name: "net1",
					},
				},
				FailNetworkRemove: utils.StringSet{
					"net1": {},
				},
			}),
		},
		want: `networks delete failed for 1 networks, reason\(s\):
1 - failed to remove the network, reason: failed to remove network net1 on the fake docker host`,
	},
}

func TestExecHomelabCmdErrors(t *testing.T) {
	t.Parallel()

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
			"config",
			"show",
		},
		ctxInfo: &testutils.TestContextInfo{},
		envs: testhelpers.TestEnvMap{
			"HOME": "",
		},
		want: `config show failed while determining the configs path, reason: failed to obtain the user's home directory for reading the homelab CLI config, reason: \$HOME is not defined`,
	},
	{
		name: "Homelab Command - Show Config - Default CLI Config Path Doesn't Exist",
		args: []string{
			"config",
			"show",
		},
		ctxInfo: &testutils.TestContextInfo{},
		envs: testhelpers.TestEnvMap{
			"HOME": "/foo/bar",
		},
		want: `config show failed while determining the configs path, reason: failed to open homelab CLI config file, reason: open /foo/bar/\.homelab/config\.yaml: no such file or directory`,
	},
}

//nolint:paralleltest // Test sets environment variables.
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
		name: "Homelab Command - Groups Start - Docker Client Creation Failed",
		args: []string{
			"groups",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{},
		envs: testhelpers.TestEnvMap{
			"DOCKER_HOST": "/var/run/foobar-docker.sock",
		},
		want: "Failed to create a new docker API client, reason: unable to parse docker host `/var/run/foobar-docker\\.sock`",
	},
}

//nolint:paralleltest // Test sets environment variables.
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

var executeHomelabConfigCmds = []struct {
	cmdArgs        []string
	cmdNameInError string
	cmdDesc        string
}{
	{
		cmdArgs: []string{
			"config",
			"show",
		},
		cmdNameInError: "config show",
		cmdDesc:        "Show Config",
	},
	{
		cmdArgs: []string{
			"groups",
			"start",
			"all",
		},
		cmdNameInError: "groups start",
		cmdDesc:        "Groups Start",
	},
	{
		cmdArgs: []string{
			"groups",
			"stop",
			"all",
		},
		cmdNameInError: "groups stop",
		cmdDesc:        "Groups Stop",
	},
	{
		cmdArgs: []string{
			"groups",
			"purge",
			"all",
		},
		cmdNameInError: "groups purge",
		cmdDesc:        "Groups Purge",
	},
	{
		cmdArgs: []string{
			"containers",
			"start",
			"g1/c1",
		},
		cmdNameInError: "containers start",
		cmdDesc:        "Containers Start",
	},
	{
		cmdArgs: []string{
			"containers",
			"stop",
			"g1/c1",
		},
		cmdNameInError: "containers stop",
		cmdDesc:        "Containers Stop",
	},
	{
		cmdArgs: []string{
			"containers",
			"purge",
			"g1/c1",
		},
		cmdNameInError: "containers purge",
		cmdDesc:        "Containers Purge",
	},
	{
		cmdArgs: []string{
			"networks",
			"create",
			"net1",
		},
		cmdNameInError: "networks create",
		cmdDesc:        "Networks Create",
	},
	{
		cmdArgs: []string{
			"networks",
			"delete",
			"net1",
		},
		cmdNameInError: "networks delete",
		cmdDesc:        "Networks Delete",
	},
}

var executeHomelabConfigCmdErrorTests = []struct {
	name    string
	args    []string
	ctxInfo func() *testutils.TestContextInfo
	want    string
}{
	{
		name: "Homelab Command - %s - Non Existing CLI Config Path",
		args: []string{
			"--cli-config",
			fmt.Sprintf("%s/testdata/foobar.yaml", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `%s failed while determining the configs path, reason: failed to open homelab CLI config file, reason: open .+/testdata/foobar\.yaml: no such file or directory`,
	},
	{
		name: "Homelab Command - %s - Non Existing Configs Path",
		args: []string{
			"--configs-dir",
			fmt.Sprintf("%s/testdata/foobar", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `%s failed while parsing the configs, reason: os\.Stat\(\) failed on homelab configs path, reason: stat .+/testdata/foobar: no such file or directory`,
	},
	{
		name: "Homelab Command - %s - Invalid Empty CLI Config",
		args: []string{
			"--cli-config",
			fmt.Sprintf("%s/testdata/cli-configs/invalid-empty-config/config.yaml", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `%s failed while determining the configs path, reason: failed to parse homelab CLI config, reason: EOF`,
	},
	{
		name: "Homelab Command - %s - Invalid Garbage CLI Config",
		args: []string{
			"--cli-config",
			fmt.Sprintf("%s/testdata/cli-configs/invalid-garbage-config/config.yaml", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `%s failed while determining the configs path, reason: failed to parse homelab CLI config, reason: yaml: unmarshal errors:
  line 1: cannot unmarshal !!str ` + "`foo bar`" + ` into cliconfig.CLIConfig`,
	},
	{
		name: "Homelab Command - %s  - Invalid CLI Config With Empty Configs Path",
		args: []string{
			"--cli-config",
			fmt.Sprintf("%s/testdata/cli-configs/invalid-config-with-empty-configs-path/config.yaml", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `%s failed while determining the configs path, reason: homelab configs path setting in homelab.configsPath is empty/unset in the homelab CLI config`,
	},
	{
		name: "Homelab Command - %s - Invalid CLI Config With Invalid Configs Path",
		args: []string{
			"--cli-config",
			fmt.Sprintf("%s/testdata/cli-configs/invalid-config-with-invalid-configs-path/config.yaml", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `%s failed while parsing the configs, reason: os\.Stat\(\) failed on homelab configs path, reason: stat /foo2/bar2: no such file or directory`,
	},
}

func TestExecHomelabConfigCmdErrors(t *testing.T) {
	t.Parallel()

	for _, test := range executeHomelabConfigCmdErrorTests {
		tc := test
		for _, c := range executeHomelabConfigCmds {
			cmd := c
			tcName := fmt.Sprintf(tc.name, cmd.cmdDesc)
			t.Run(tcName, func(t *testing.T) {
				t.Parallel()

				args := append(cmd.cmdArgs, tc.args...)
				want := fmt.Sprintf(tc.want, cmd.cmdNameInError)

				_, gotErr := execHomelabCmdTest(tc.ctxInfo(), nil, args...)
				if gotErr == nil {
					testhelpers.LogErrorNil(t, "Exec()", tcName, want)
					return
				}

				if !testhelpers.RegexMatch(t, "Exec()", tcName, "gotErr error string", want, gotErr.Error()) {
					return
				}
			})
		}
	}
}

var executeHomelabGroupsCmds = []struct {
	cmdArgs        []string
	cmdNameInError string
	cmdDesc        string
}{
	{
		cmdArgs: []string{
			"groups",
			"start",
		},
		cmdNameInError: "groups start",
		cmdDesc:        "Groups Start",
	},
	{
		cmdArgs: []string{
			"groups",
			"stop",
		},
		cmdNameInError: "groups stop",
		cmdDesc:        "Groups Stop",
	},
	{
		cmdArgs: []string{
			"groups",
			"purge",
		},
		cmdNameInError: "groups purge",
		cmdDesc:        "Groups Purge",
	},
}

var executeHomelabGroupsCmdTests = []struct {
	name    string
	args    []string
	ctxInfo func() *testutils.TestContextInfo
	want    string
}{
	{
		name: "Homelab Command - %s - One Empty Group",
		args: []string{
			"g3",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `%s is a no-op since no containers were found matching the specified criteria`,
	},
}

func TestExecHomelabGroupsCmd(t *testing.T) {
	t.Parallel()

	for _, test := range executeHomelabGroupsCmdTests {
		tc := test
		for _, c := range executeHomelabGroupsCmds {
			cmd := c
			tcName := fmt.Sprintf(tc.name, cmd.cmdDesc)
			t.Run(tcName, func(t *testing.T) {
				t.Parallel()

				args := append(cmd.cmdArgs, tc.args...)
				want := fmt.Sprintf(tc.want, cmd.cmdNameInError)

				out, gotErr := execHomelabCmdTest(tc.ctxInfo(), nil, args...)
				if gotErr != nil {
					testhelpers.LogErrorNotNilWithOutput(t, "Exec()", tcName, out, gotErr)
					return
				}

				if !testhelpers.RegexMatchJoinNewLines(t, "Exec()", tcName, "command output", want, out.String()) {
					return
				}
			})
		}
	}
}

var executeHomelabGroupsCmdErrorTests = []struct {
	name    string
	args    []string
	ctxInfo func() *testutils.TestContextInfo
	want    string
}{
	{
		name: "Homelab Command - %s - One Non Existing Group",
		args: []string{
			"g4",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `%s failed while querying containers, reason: group g4 not found`,
	},
}

func TestExecHomelabGroupsCmdErrors(t *testing.T) {
	t.Parallel()

	for _, test := range executeHomelabGroupsCmdErrorTests {
		tc := test
		for _, c := range executeHomelabGroupsCmds {
			cmd := c
			tcName := fmt.Sprintf(tc.name, cmd.cmdDesc)
			t.Run(tcName, func(t *testing.T) {
				t.Parallel()

				args := append(cmd.cmdArgs, tc.args...)
				want := fmt.Sprintf(tc.want, cmd.cmdNameInError)

				_, gotErr := execHomelabCmdTest(tc.ctxInfo(), nil, args...)
				if gotErr == nil {
					testhelpers.LogErrorNil(t, "Exec()", tcName, want)
					return
				}

				if !testhelpers.RegexMatch(t, "Exec()", tcName, "gotErr error string", want, gotErr.Error()) {
					return
				}
			})
		}
	}
}

var executeHomelabContainerCmds = []struct {
	cmdArgs        []string
	cmdNameInError string
	cmdDesc        string
}{
	{
		cmdArgs: []string{
			"containers",
			"start",
		},
		cmdNameInError: "containers start",
		cmdDesc:        "Containers Start",
	},
	{
		cmdArgs: []string{
			"containers",
			"stop",
		},
		cmdNameInError: "containers stop",
		cmdDesc:        "Containers Stop",
	},
	{
		cmdArgs: []string{
			"containers",
			"purge",
		},
		cmdNameInError: "containers purge",
		cmdDesc:        "Containers Purge",
	},
}

var executeHomelabContainerCmdErrorTests = []struct {
	name    string
	args    []string
	ctxInfo func() *testutils.TestContextInfo
	want    string
}{
	{
		name: "Homelab Command - %s - One Non Existing Container In Invalid Group",
		args: []string{
			"g4/c3",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `%s failed while querying containers, reason: group g4 not found`,
	},
	{
		name: "Homelab Command - %s - One Non Existing Container In Valid Group",
		args: []string{
			"g1/c3",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `%s failed while querying containers, reason: container {g1 c3} not found`,
	},
}

func TestExecHomelabContainerCmdErrors(t *testing.T) {
	t.Parallel()

	for _, test := range executeHomelabContainerCmdErrorTests {
		tc := test
		for _, c := range executeHomelabContainerCmds {
			cmd := c
			tcName := fmt.Sprintf(tc.name, cmd.cmdDesc)
			t.Run(tcName, func(t *testing.T) {
				t.Parallel()

				args := append(cmd.cmdArgs, tc.args...)
				want := fmt.Sprintf(tc.want, cmd.cmdNameInError)

				_, gotErr := execHomelabCmdTest(tc.ctxInfo(), nil, args...)
				if gotErr == nil {
					testhelpers.LogErrorNil(t, "Exec()", tcName, want)
					return
				}

				if !testhelpers.RegexMatch(t, "Exec()", tcName, "gotErr error string", want, gotErr.Error()) {
					return
				}
			})
		}
	}
}

var executeHomelabNetworksCmdErrorTests = []struct {
	name    string
	args    []string
	ctxInfo func() *testutils.TestContextInfo
	want    string
}{
	{
		name: "Homelab Command - %s - One Non Existing Network",
		args: []string{
			"net6",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/networks-cmd", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `%s failed while querying networks, reason: network net6 not found`,
	},
}

func TestExecHomelabNetworksCmdErrors(t *testing.T) {
	t.Parallel()

	for _, test := range executeHomelabNetworksCmdErrorTests {
		tc := test
		for _, c := range executeHomelabNetworksCmds {
			cmd := c
			tcName := fmt.Sprintf(tc.name, cmd.cmdDesc)
			t.Run(tcName, func(t *testing.T) {
				t.Parallel()

				args := append(cmd.cmdArgs, tc.args...)
				want := fmt.Sprintf(tc.want, cmd.cmdNameInError)

				_, gotErr := execHomelabCmdTest(tc.ctxInfo(), nil, args...)
				if gotErr == nil {
					testhelpers.LogErrorNil(t, "Exec()", tcName, want)
					return
				}

				if !testhelpers.RegexMatch(t, "Exec()", tcName, "gotErr error string", want, gotErr.Error()) {
					return
				}
			})
		}
	}
}

var executeHomelabGroupsCmdCompletionTests = []struct {
	name        string
	preCmdArgs  []string
	postCmdArgs []string
	ctxInfo     func() *testutils.TestContextInfo
	want        string
}{
	{
		name: "Homelab Command - %s - Completion - All Group Names",
		preCmdArgs: []string{
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
			"__complete",
		},
		postCmdArgs: []string{
			"",
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `all
g1
g2
g3
:36
Completion ended with directive: ShellCompDirectiveNoFileComp, ShellCompDirectiveKeepOrder`,
	},
	{
		name: "Homelab Command - %s - Completion - No Group Names",
		preCmdArgs: []string{
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
			"__complete",
		},
		postCmdArgs: []string{
			"g1",
			"",
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `:36
Completion ended with directive: ShellCompDirectiveNoFileComp, ShellCompDirectiveKeepOrder`,
	},
	{
		name: "Homelab Command - %s - Completion - Invalid CLI Config",
		preCmdArgs: []string{
			"--cli-config",
			fmt.Sprintf("%s/testdata/cli-configs/invalid-empty-config/config.yaml", testhelpers.Pwd()),
			"__complete",
		},
		postCmdArgs: []string{
			"",
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `:1
Completion ended with directive: ShellCompDirectiveError`,
	},
	{
		name: "Homelab Command - %s - Completion - Invalid Homelab Config - Merge Fail",
		preCmdArgs: []string{
			"--configs-dir",
			fmt.Sprintf("%s/testdata/parse-configs-invalid-deepmerge-fail", testhelpers.Pwd()),
			"__complete",
		},
		postCmdArgs: []string{
			"",
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `:1
Completion ended with directive: ShellCompDirectiveError`,
	},
	{
		name: "Homelab Command - %s - Completion - Invalid Homelab Config - Parse Fail",
		preCmdArgs: []string{
			"--configs-dir",
			fmt.Sprintf("%s/testdata/parse-group-only-configs-invalid-config", testhelpers.Pwd()),
			"__complete",
		},
		postCmdArgs: []string{
			"",
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `:1
Completion ended with directive: ShellCompDirectiveError`,
	},
}

func TestExecHomelabGroupsCmdCompletions(t *testing.T) {
	t.Parallel()

	for _, test := range executeHomelabGroupsCmdCompletionTests {
		tc := test
		for _, c := range executeHomelabGroupsCmds {
			cmd := c
			tcName := fmt.Sprintf(tc.name, cmd.cmdDesc)
			t.Run(tcName, func(t *testing.T) {
				t.Parallel()

				args := append(tc.preCmdArgs, cmd.cmdArgs...)
				args = append(args, tc.postCmdArgs...)

				out, gotErr := execHomelabCmdTest(tc.ctxInfo(), nil, args...)
				if gotErr != nil {
					testhelpers.LogErrorNotNilWithOutput(t, "Exec()", tcName, out, gotErr)
					return
				}

				if !testhelpers.RegexMatchJoinNewLines(t, "Exec()", tcName, "command output", tc.want, out.String()) {
					return
				}
			})
		}
	}
}

var executeHomelabContainerCmdCompletionTests = []struct {
	name        string
	preCmdArgs  []string
	postCmdArgs []string
	ctxInfo     func() *testutils.TestContextInfo
	want        string
}{
	{
		name: "Homelab Command - %s - Completion - All Container Names",
		preCmdArgs: []string{
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
			"__complete",
		},
		postCmdArgs: []string{
			"",
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `g1/c1
g1/c2
g2/c3
:36
Completion ended with directive: ShellCompDirectiveNoFileComp, ShellCompDirectiveKeepOrder`,
	},
	{
		name: "Homelab Command - %s - Completion - No Container Names",
		preCmdArgs: []string{
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
			"__complete",
		},
		postCmdArgs: []string{
			"g1/c1",
			"",
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `:36
Completion ended with directive: ShellCompDirectiveNoFileComp, ShellCompDirectiveKeepOrder`,
	},
	{
		name: "Homelab Command - %s - Completion - Invalid CLI Config",
		preCmdArgs: []string{
			"--cli-config",
			fmt.Sprintf("%s/testdata/cli-configs/invalid-empty-config/config.yaml", testhelpers.Pwd()),
			"__complete",
		},
		postCmdArgs: []string{
			"",
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `:1
Completion ended with directive: ShellCompDirectiveError`,
	},
	{
		name: "Homelab Command - %s - Completion - Invalid Homelab Config",
		preCmdArgs: []string{
			"--configs-dir",
			fmt.Sprintf("%s/testdata/parse-configs-invalid-deepmerge-fail", testhelpers.Pwd()),
			"__complete",
		},
		postCmdArgs: []string{
			"",
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `:1
Completion ended with directive: ShellCompDirectiveError`,
	},
	{
		name: "Homelab Command - %s - Completion - Invalid Homelab Config - Parse Fail",
		preCmdArgs: []string{
			"--configs-dir",
			fmt.Sprintf("%s/testdata/parse-container-only-configs-invalid-config", testhelpers.Pwd()),
			"__complete",
		},
		postCmdArgs: []string{
			"",
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `:1
Completion ended with directive: ShellCompDirectiveError`,
	},
}

func TestExecHomelabContainerCmdCompletions(t *testing.T) {
	t.Parallel()

	for _, test := range executeHomelabContainerCmdCompletionTests {
		tc := test
		for _, c := range executeHomelabContainerCmds {
			cmd := c
			tcName := fmt.Sprintf(tc.name, cmd.cmdDesc)
			t.Run(tcName, func(t *testing.T) {
				t.Parallel()

				args := append(tc.preCmdArgs, cmd.cmdArgs...)
				args = append(args, tc.postCmdArgs...)

				out, gotErr := execHomelabCmdTest(tc.ctxInfo(), nil, args...)
				if gotErr != nil {
					testhelpers.LogErrorNotNilWithOutput(t, "Exec()", tcName, out, gotErr)
					return
				}

				if !testhelpers.RegexMatchJoinNewLines(t, "Exec()", tcName, "command output", tc.want, out.String()) {
					return
				}
			})
		}
	}
}

var executeHomelabNetworksCmds = []struct {
	cmdArgs        []string
	cmdNameInError string
	cmdDesc        string
}{
	{
		cmdArgs: []string{
			"networks",
			"create",
		},
		cmdNameInError: "networks create",
		cmdDesc:        "Networks Create",
	},
	{
		cmdArgs: []string{
			"networks",
			"delete",
		},
		cmdNameInError: "networks delete",
		cmdDesc:        "Networks Delete",
	},
}

var executeHomelabNetworksCmdCompletionTests = []struct {
	name        string
	preCmdArgs  []string
	postCmdArgs []string
	ctxInfo     func() *testutils.TestContextInfo
	want        string
}{
	{
		name: "Homelab Command - %s - Completion - All Network Names",
		preCmdArgs: []string{
			"--configs-dir",
			fmt.Sprintf("%s/testdata/networks-cmd", testhelpers.Pwd()),
			"__complete",
		},
		postCmdArgs: []string{
			"",
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `all
net1
net2
net3
net4
net5
:36
Completion ended with directive: ShellCompDirectiveNoFileComp, ShellCompDirectiveKeepOrder`,
	},
	{
		name: "Homelab Command - %s - Completion - No Network Names",
		preCmdArgs: []string{
			"--configs-dir",
			fmt.Sprintf("%s/testdata/containers-and-groups-cmds", testhelpers.Pwd()),
			"__complete",
		},
		postCmdArgs: []string{
			"net1",
			"",
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `:36
Completion ended with directive: ShellCompDirectiveNoFileComp, ShellCompDirectiveKeepOrder`,
	},
	{
		name: "Homelab Command - %s - Completion - Invalid CLI Config",
		preCmdArgs: []string{
			"--cli-config",
			fmt.Sprintf("%s/testdata/cli-configs/invalid-empty-config/config.yaml", testhelpers.Pwd()),
			"__complete",
		},
		postCmdArgs: []string{
			"",
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `:1
Completion ended with directive: ShellCompDirectiveError`,
	},
	{
		name: "Homelab Command - %s - Completion - Invalid Homelab Config - Merge Fail",
		preCmdArgs: []string{
			"--configs-dir",
			fmt.Sprintf("%s/testdata/parse-configs-invalid-deepmerge-fail", testhelpers.Pwd()),
			"__complete",
		},
		postCmdArgs: []string{
			"",
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `:1
Completion ended with directive: ShellCompDirectiveError`,
	},
	{
		name: "Homelab Command - %s - Completion - Invalid Homelab Config - Parse Fail",
		preCmdArgs: []string{
			"--configs-dir",
			fmt.Sprintf("%s/testdata/parse-networks-only-configs-invalid-config", testhelpers.Pwd()),
			"__complete",
		},
		postCmdArgs: []string{
			"",
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `:1
Completion ended with directive: ShellCompDirectiveError`,
	},
}

func TestExecHomelabNetworksCmdCompletions(t *testing.T) {
	t.Parallel()

	for _, test := range executeHomelabNetworksCmdCompletionTests {
		tc := test
		for _, c := range executeHomelabNetworksCmds {
			cmd := c
			tcName := fmt.Sprintf(tc.name, cmd.cmdDesc)
			t.Run(tcName, func(t *testing.T) {
				t.Parallel()

				args := append(tc.preCmdArgs, cmd.cmdArgs...)
				args = append(args, tc.postCmdArgs...)

				out, gotErr := execHomelabCmdTest(tc.ctxInfo(), nil, args...)
				if gotErr != nil {
					testhelpers.LogErrorNotNilWithOutput(t, "Exec()", tcName, out, gotErr)
					return
				}

				if !testhelpers.RegexMatchJoinNewLines(t, "Exec()", tcName, "command output", tc.want, out.String()) {
					return
				}
			})
		}
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
	if ctxInfo.ContainerPurgeKillAttempts == 0 {
		// Reduce the number of attempts to keep the tests executing quickly.
		ctxInfo.ContainerPurgeKillAttempts = 5
	}

	ctx := testutils.NewTestContext(ctxInfo)
	err := Exec(ctx, buf, buf, args...)
	return buf, err
}
