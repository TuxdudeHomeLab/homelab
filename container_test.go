package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

var containerStartTests = []struct {
	name           string
	config         HomelabConfig
	cRef           ContainerReference
	ctxInfo        *testContextInfo
	preExec        func(context.Context)
	wantNotStarted bool
}{
	{
		name: "Container Start - Doesn't Exist Already",
		config: buildSingleContainerConfig(ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
			"abc/xyz"),
		cRef: ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				validImagesForPull: stringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Created State",
		config: buildSingleContainerConfig(ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
			"abc/xyz"),
		cRef: ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				containers: []*fakeContainerInitInfo{
					{
						name:  "g1-c1",
						image: "abc/xyz",
						state: containerStateCreated,
					},
				},
				validImagesForPull: stringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Running State",
		config: buildSingleContainerConfig(ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
			"abc/xyz"),
		cRef: ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				containers: []*fakeContainerInitInfo{
					{
						name:  "g1-c1",
						image: "abc/xyz",
						state: containerStateRunning,
					},
				},
				validImagesForPull: stringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Paused State",
		config: buildSingleContainerConfig(ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
			"abc/xyz"),
		cRef: ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				containers: []*fakeContainerInitInfo{
					{
						name:  "g1-c1",
						image: "abc/xyz",
						state: containerStatePaused,
					},
				},
				validImagesForPull: stringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Restarting State",
		config: buildSingleContainerConfig(ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
			"abc/xyz"),
		cRef: ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				containers: []*fakeContainerInitInfo{
					{
						name:  "g1-c1",
						image: "abc/xyz",
						state: containerStateRestarting,
					},
				},
				validImagesForPull: stringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Removing State",
		config: buildSingleContainerConfig(ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
			"abc/xyz"),
		cRef: ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				containers: []*fakeContainerInitInfo{
					{
						name:  "g1-c1",
						image: "abc/xyz",
						state: containerStateRemoving,
					},
				},
				validImagesForPull: stringSet{
					"abc/xyz": {},
				},
			}),
		},
		preExec: func(ctx context.Context) {
			time.Sleep(2 * time.Second)
			docker := fakeDockerHostFromContext(ctx)
			err := docker.forceRemoveContainer("g1-c1")
			if err != nil {
				panic(err)
			}
		},
	},
	{
		name: "Container Start - Exists Already In Exited State",
		config: buildSingleContainerConfig(ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
			"abc/xyz"),
		cRef: ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				containers: []*fakeContainerInitInfo{
					{
						name:  "g1-c1",
						image: "abc/xyz",
						state: containerStateExited,
					},
				},
				validImagesForPull: stringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Dead State",
		config: buildSingleContainerConfig(ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
			"abc/xyz"),
		cRef: ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testContextInfo{
			dockerHost: newFakeDockerHost(&fakeDockerHostInitInfo{
				containers: []*fakeContainerInitInfo{
					{
						name:  "g1-c1",
						image: "abc/xyz",
						state: containerStateDead,
					},
				},
				validImagesForPull: stringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
}

func TestContainerStart(t *testing.T) {
	for _, test := range containerStartTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := newTestContext(tc.ctxInfo)

			dep, gotErr := buildDeploymentFromConfig(ctx, &tc.config)
			if gotErr != nil {
				t.Errorf(
					"buildDeploymentFromConfig()\nTest Case: %q\nFailure: gotErr != nil\nReason: %v",
					tc.name, gotErr)
				return
			}

			dockerClient, gotErr := newDockerClient(ctx, dep.host.dockerPlatform, dep.host.arch)
			if gotErr != nil {
				t.Errorf(
					"newDockerClient()\nTest Case: %q\nFailure: gotErr != nil\nReason: %v",
					tc.name, gotErr)
				return
			}
			defer dockerClient.close()

			ct, gotErr := dep.queryContainer(tc.cRef)
			if gotErr != nil {
				t.Errorf(
					"deployment.queryContainer()\nTest Case: %q\nFailure: gotErr != nil\nReason: %v",
					tc.name, gotErr)
				return
			}

			if tc.preExec != nil {
				go tc.preExec(ctx)
			}

			gotStarted, gotErr := ct.start(ctx, dockerClient)
			if gotErr != nil {
				t.Errorf(
					"container.start()\nTest Case: %q\nFailure: gotErr != nil\nReason: %v",
					tc.name, gotErr)
				return
			}
			wantStarted := !tc.wantNotStarted
			if gotStarted != wantStarted {
				t.Errorf(
					"container.start()\nTest Case: %q\nReason: gotStarted (%t) != wantStarted (%t)",
					tc.name, gotStarted, wantStarted)
			}

			docker := fakeDockerHostFromContext(ctx)
			gotState := docker.getContainerState(fmt.Sprintf("%s-%s", tc.cRef.Group, tc.cRef.Container))
			if gotState != containerStateRunning {
				t.Errorf(
					"Container State after container.start()\nTest Case: %q\nReason: gotState (%s) != containerStateRunning",
					tc.name, gotState)
			}
		})
	}
}

func buildSingleContainerConfig(ct ContainerReference, image string) HomelabConfig {
	return HomelabConfig{
		IPAM: IPAMConfig{
			Networks: NetworksConfig{
				BridgeModeNetworks: []BridgeModeNetworkConfig{
					{
						Name:              fmt.Sprintf("%s-bridge", ct.Group),
						HostInterfaceName: fmt.Sprintf("docker-%s", ct.Group),
						CIDR:              "172.18.101.0/24",
						Priority:          1,
						Containers: []ContainerIPConfig{
							{
								IP:        "172.18.101.11",
								Container: ct,
							},
						},
					},
				},
			},
		},
		Hosts: []HostConfig{
			{
				Name: fakeHostName,
				AllowedContainers: []ContainerReference{
					ct,
				},
			},
		},
		Groups: []ContainerGroupConfig{
			{
				Name:  ct.Group,
				Order: 1,
			},
		},
		Containers: []ContainerConfig{
			{
				Info: ct,
				Image: ContainerImageConfig{
					Image: image,
				},
				Lifecycle: ContainerLifecycleConfig{
					Order: 1,
				},
			},
		},
	}
}
