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
		name: "Homelab Command - Start - All Groups With Real Host Info",
		args: []string{
			"group",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
			"group",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		name: "Homelab Command - Start - All Groups",
		args: []string{
			"group",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		name: "Homelab Command - Start - All Groups - Container Create Warning",
		args: []string{
			"group",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		name: "Homelab Command - Start - All Groups - Network Create Warning",
		args: []string{
			"group",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		name: "Homelab Command - Start - All Groups - One Existing Image",
		args: []string{
			"group",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		name: "Homelab Command - Start - All Groups With Multiple Same Order Containers",
		args: []string{
			"group",
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
		name: "Homelab Command - Start - All Groups With No Network Endpoints Containers",
		args: []string{
			"group",
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
		name: "Homelab Command - Start - All Groups - One Container With Start Pre-Hook",
		args: []string{
			"group",
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
		name: "Homelab Command - Start - All Groups - One Container With Ignore Image Pull Failures",
		args: []string{
			"group",
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
		name: "Homelab Command - Start - One Group",
		args: []string{
			"group",
			"start",
			"g1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		name: "Homelab Command - Start - One Container",
		args: []string{
			"container",
			"start",
			"g1/c1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		name: "Homelab Command - Stop - All Groups",
		args: []string{
			"group",
			"stop",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		name: "Homelab Command - Stop - All Groups - One Container With Ignore Image Pull Failures",
		args: []string{
			"group",
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
		name: "Homelab Command - Stop - One Group",
		args: []string{
			"group",
			"stop",
			"g1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		name: "Homelab Command - Stop - One Container - Not Found",
		args: []string{
			"container",
			"stop",
			"g1/c1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		name: "Homelab Command - Purge - All Groups",
		args: []string{
			"group",
			"purge",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		name: "Homelab Command - Purge - One Group",
		args: []string{
			"group",
			"purge",
			"g1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		name: "Homelab Command - Purge - One Container - Not Found",
		args: []string{
			"container",
			"purge",
			"g1/c1",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		name: "Homelab Command - Start - All Groups - Real Everything",
		args: []string{
			"group",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
		},
		want: `Container g1-c1 not allowed to run on host [^\s]+
Container g1-c2 not allowed to run on host [^\s]+
Container g2-c3 not allowed to run on host [^\s]+`,
	},
	{
		name: "Homelab Command - Stop - All Groups - Real Everything",
		args: []string{
			"group",
			"stop",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
		},
		want: `Container g1-c1 cannot be stopped since it was not found
Container g1-c2 cannot be stopped since it was not found
Container g2-c3 cannot be stopped since it was not found`,
	},
	{
		name: "Homelab Command - Purge - All Groups - Real Everything",
		args: []string{
			"group",
			"purge",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		name: "Homelab Command - Start - All Groups",
		args: []string{
			"group",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		name: "Homelab Command - Stop - All Groups",
		args: []string{
			"group",
			"stop",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		name: "Homelab Command - Purge - All Groups",
		args: []string{
			"group",
			"purge",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		name: "Homelab Group Command - Missing Subcommand",
		args: []string{
			"group",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `homelab group sub-command is required`,
	},
	{
		name: "Homelab Container Command - Missing Subcommand",
		args: []string{
			"container",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `homelab container sub-command is required`,
	},
	{
		name: "Homelab Command - Group Start - Zero Group Name Args",
		args: []string{
			"group",
			"start",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one group name argument to be specified, but found 0 instead`,
	},
	{
		name: "Homelab Command - Group Stop - Zero Group Name Args",
		args: []string{
			"group",
			"stop",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one group name argument to be specified, but found 0 instead`,
	},
	{
		name: "Homelab Command - Group Purge - Zero Group Name Args",
		args: []string{
			"group",
			"purge",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one group name argument to be specified, but found 0 instead`,
	},
	{
		name: "Homelab Command - Group Start - Multiple Group Name Args",
		args: []string{
			"group",
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
		name: "Homelab Command - Group Stop - Multiple Group Name Args",
		args: []string{
			"group",
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
		name: "Homelab Command - Group Purge - Multiple Group Name Args",
		args: []string{
			"group",
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
		name: "Homelab Command - Container Start - Zero Container Name Args",
		args: []string{
			"container",
			"start",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one container name argument to be specified, but found 0 instead`,
	},
	{
		name: "Homelab Command - Container Stop - Zero Container Name Args",
		args: []string{
			"container",
			"stop",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one container name argument to be specified, but found 0 instead`,
	},
	{
		name: "Homelab Command - Container Purge - Zero Container Name Args",
		args: []string{
			"container",
			"purge",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Expected exactly one container name argument to be specified, but found 0 instead`,
	},
	{
		name: "Homelab Command - Container Start - Multiple Container Name Args",
		args: []string{
			"container",
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
		name: "Homelab Command - Container Stop - Multiple Container Name Args",
		args: []string{
			"container",
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
		name: "Homelab Command - Container Purge - Multiple Container Name Args",
		args: []string{
			"container",
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
		name: "Homelab Command - Container Start - Invalid Container Name",
		args: []string{
			"container",
			"start",
			"foobar",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Container name must be specified in the form 'group/container'`,
	},
	{
		name: "Homelab Command - Container Stop - Invalid Container Name",
		args: []string{
			"container",
			"stop",
			"foobar",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Container name must be specified in the form 'group/container'`,
	},
	{
		name: "Homelab Command - Container Purge - Invalid Container Name",
		args: []string{
			"container",
			"purge",
			"foobar",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Container name must be specified in the form 'group/container'`,
	},
	{
		name: "Homelab Command - Start - Failure",
		args: []string{
			"group",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `group start failed for 2 containers, reason\(s\):
1 - Failed to start container g1-c1, reason:failed to pull the image abc/xyz, reason: image abc/xyz not found or invalid and cannot be pulled by the fake docker host
2 - Failed to start container g2-c3, reason:failed to pull the image abc/xyz3, reason: image abc/xyz3 not found or invalid and cannot be pulled by the fake docker host`,
	},
	{
		name: "Homelab Command - Stop - Failure",
		args: []string{
			"group",
			"stop",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		want: `group stop failed for 2 containers, reason\(s\):
1 - Failed to stop container g1-c1, reason:failed to stop the container, reason: failed to stop container g1-c1 on the fake docker host
2 - Failed to stop container g2-c3, reason:failed to stop the container, reason: failed to stop container g2-c3 on the fake docker host`,
	},
	{
		name: "Homelab Command - Purge - Failure",
		args: []string{
			"group",
			"purge",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
		want: `group purge failed for 2 containers, reason\(s\):
1 - Failed to purge container g1-c1, reason:failed to stop the container, reason: failed to stop container g1-c1 on the fake docker host
2 - Failed to purge container g2-c3, reason:failed to purge container g2-c3 after 6 attempts`,
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
		name: "Homelab Command - Start - Docker Client Creation Failed",
		args: []string{
			"group",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
			"group",
			"start",
			"all",
		},
		cmdNameInError: "group start",
		cmdDesc:        "Group Start",
	},
	{
		cmdArgs: []string{
			"group",
			"stop",
			"all",
		},
		cmdNameInError: "group stop",
		cmdDesc:        "Group Stop",
	},
	{
		cmdArgs: []string{
			"group",
			"purge",
			"all",
		},
		cmdNameInError: "group purge",
		cmdDesc:        "Group Purge",
	},
	{
		cmdArgs: []string{
			"container",
			"start",
			"g1/c1",
		},
		cmdNameInError: "container start",
		cmdDesc:        "Container Start",
	},
	{
		cmdArgs: []string{
			"container",
			"stop",
			"g1/c1",
		},
		cmdNameInError: "container stop",
		cmdDesc:        "Container Stop",
	},
	{
		cmdArgs: []string{
			"container",
			"purge",
			"g1/c1",
		},
		cmdNameInError: "container purge",
		cmdDesc:        "Container Purge",
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

var executeHomelabGroupCmds = []struct {
	cmdArgs        []string
	cmdNameInError string
	cmdDesc        string
}{
	{
		cmdArgs: []string{
			"group",
			"start",
		},
		cmdNameInError: "group start",
		cmdDesc:        "Group Start",
	},
	{
		cmdArgs: []string{
			"group",
			"stop",
		},
		cmdNameInError: "group stop",
		cmdDesc:        "Group Stop",
	},
	{
		cmdArgs: []string{
			"group",
			"purge",
		},
		cmdNameInError: "group purge",
		cmdDesc:        "Group Purge",
	},
}

var executeHomelabGroupCmdTests = []struct {
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
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `%s is a no-op since no containers were found matching the specified criteria`,
	},
}

func TestExecHomelabGroupCmd(t *testing.T) {
	t.Parallel()

	for _, test := range executeHomelabGroupCmdTests {
		tc := test
		for _, c := range executeHomelabGroupCmds {
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

var executeHomelabGroupCmdErrorTests = []struct {
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
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `%s failed while querying containers, reason: group g4 not found`,
	},
}

func TestExecHomelabGroupCmdErrors(t *testing.T) {
	t.Parallel()

	for _, test := range executeHomelabGroupCmdErrorTests {
		tc := test
		for _, c := range executeHomelabGroupCmds {
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
			"container",
			"start",
		},
		cmdNameInError: "container start",
		cmdDesc:        "Container Start",
	},
	{
		cmdArgs: []string{
			"container",
			"stop",
		},
		cmdNameInError: "container stop",
		cmdDesc:        "Container Stop",
	},
	{
		cmdArgs: []string{
			"container",
			"purge",
		},
		cmdNameInError: "container purge",
		cmdDesc:        "Container Purge",
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
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
		},
		ctxInfo: func() *testutils.TestContextInfo {
			return &testutils.TestContextInfo{
				DockerHost: fakedocker.NewEmptyFakeDockerHost(),
			}
		},
		want: `%s failed while querying containers, reason: container {g1 c3} not found`,
	},
}

func TestExecHomelabContainerGroupCmdErrors(t *testing.T) {
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

var executeHomelabGroupCmdCompletionTests = []struct {
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
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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

func TestExecHomelabGroupCmdCompletions(t *testing.T) {
	t.Parallel()

	for _, test := range executeHomelabGroupCmdCompletionTests {
		tc := test
		for _, c := range executeHomelabGroupCmds {
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
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
			fmt.Sprintf("%s/testdata/container-group-cmd", testhelpers.Pwd()),
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
