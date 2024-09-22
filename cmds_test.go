package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

var executeHomelabCmdTests = []struct {
	name    string
	args    []string
	preExec func()
	want    string
}{
	{
		name: "Homelab Command - Show Version",
		args: []string{
			"--version",
		},
		preExec: func() {
			pkgVersion = "my-pkg-version"
			pkgCommit = "my-pkg-commit"
			pkgTimestamp = "my-pkg-timestamp"
		},
		want: `homelab version my-pkg-version \[Revision: my-pkg-commit @ my-pkg-timestamp\]`,
	},
	{
		name: "Homelab Command - Show Help",
		args: []string{
			"--help",
		},
		want: `(?s)A CLI for managing both the configuration and deployment of groups of docker containers on a given host\.\n\nThe configuration is managed using a yaml file. The configuration specifies the container groups and individual containers, their properties and how to deploy them.\n\nUsage:\n.+Use "homelab \[command\] --help" for more information about a command\.`,
	},
	{
		name: "Homelab Command - Show Config",
		args: []string{
			"show-config",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/show-config-cmd", pwd()),
		},
		want: `(?s)Homelab config:
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
}

func TestExecHomelabCmd(t *testing.T) {
	for _, test := range executeHomelabCmdTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.preExec != nil {
				tc.preExec()
			}
			_, out, gotErr := execHomelabCmdTest(tc.args...)
			if gotErr != nil {
				t.Errorf(
					"execHomelabCmd()\nTest Case: %q\nFailure: gotErr != nil\nReason: %v\nOutput: %v",
					tc.name, gotErr, out)
				return
			}

			match, err := regexp.MatchString(fmt.Sprintf("^%s$", tc.want), strings.TrimSpace(out))
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

func execHomelabCmdTest(args ...string) (*cobra.Command, string, error) {
	buf := new(bytes.Buffer)
	cmd := initHomelabCmd(testContextWithLogger(capturingTestLogger(buf)))
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	c, err := cmd.ExecuteC()
	return c, buf.String(), err
}
