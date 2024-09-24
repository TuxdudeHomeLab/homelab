package main

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
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
		name: "Container Start - Doesn't Exist Already - No Network Endpoints",
		config: buildSingleContainerNoNetworkConfig(ContainerReference{
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
			go func() {
				time.Sleep(2 * time.Second)
				docker := fakeDockerHostFromContext(ctx)
				err := docker.forceRemoveContainer("g1-c1")
				if err != nil {
					panic(err)
				}
			}()
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
	{
		name: "Container Start - Exists Already In Running State Requiring Multiple Stops",
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
						name:               "g1-c1",
						image:              "abc/xyz",
						state:              containerStateRunning,
						requiredExtraStops: 5,
					},
				},
				validImagesForPull: stringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Running State Requiring Multiple Kills",
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
						name:               "g1-c1",
						image:              "abc/xyz",
						state:              containerStateRunning,
						requiredExtraStops: 1000,
						requiredExtraKills: 4,
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
			buf := new(bytes.Buffer)
			tc.ctxInfo.logger = newCapturingTestLogger(buf)
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
				tc.preExec(ctx)
			}

			gotStarted, gotErr := ct.start(ctx, dockerClient)
			if gotErr != nil {
				t.Errorf(
					"container.start()\nTest Case: %q\nFailure: gotErr != nil\n\nOut:\n%s\nReason: %v",
					tc.name, buf.String(), gotErr)
				return
			}
			wantStarted := !tc.wantNotStarted
			if gotStarted != wantStarted {
				t.Errorf(
					"container.start()\nTest Case: %q\n\nOut:\n%s\nReason: gotStarted (%t) != wantStarted (%t)",
					tc.name, buf.String(), gotStarted, wantStarted)
			}

			docker := fakeDockerHostFromContext(ctx)
			gotState := docker.getContainerState(fmt.Sprintf("%s-%s", tc.cRef.Group, tc.cRef.Container))
			if gotState != containerStateRunning {
				t.Errorf(
					"Container State after container.start()\nTest Case: %q\n\nOut:\n%s\nReason: gotState (%s) != containerStateRunning",
					tc.name, buf.String(), gotState)
			}
		})
	}
}

var containerStartErrorTests = []struct {
	name    string
	config  HomelabConfig
	cRef    ContainerReference
	ctxInfo *testContextInfo
	want    string
}{
	{
		name: "Container Start - Unkillable Existing Container",
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
						name:               "g1-c1",
						image:              "abc/xyz",
						state:              containerStateRunning,
						requiredExtraStops: 1000,
						requiredExtraKills: 5,
					},
				},
				validImagesForPull: stringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to stop and remove container g1-c1 after 6 attempts`,
	},
	{
		name: "Container Start - Inspect Existing Container Failure",
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
						name:        "g1-c1",
						image:       "abc/xyz",
						state:       containerStateRunning,
						failInspect: true,
					},
				},
				validImagesForPull: stringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to retrieve the container state, reason: failed to inspect container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Start - Stop Existing Container Failure",
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
						name:     "g1-c1",
						image:    "abc/xyz",
						state:    containerStateRunning,
						failStop: true,
					},
				},
				validImagesForPull: stringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to stop the container, reason: failed to stop container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Start - Remove Existing Container Failure",
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
						name:       "g1-c1",
						image:      "abc/xyz",
						state:      containerStateRestarting,
						failRemove: true,
					},
				},
				validImagesForPull: stringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to remove the container, reason: failed to remove container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Start - Create Container Failure",
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
				failContainerCreate: stringSet{
					"g1-c1": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to create the container, reason: failed to create container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Start - Start Container Failure",
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
				failContainerStart: stringSet{
					"g1-c1": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to start the container, reason: failed to start container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Start - Primary Network Create Failure",
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
				failNetworkCreate: stringSet{
					"g1-bridge": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to create the network, reason: failed to create network g1-bridge on the fake docker host`,
	},
	{
		name: "Container Start - Secondary Network Create Failure",
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
				failNetworkCreate: stringSet{
					"proxy-bridge": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to create the network, reason: failed to create network proxy-bridge on the fake docker host`,
	},
	{
		name: "Container Start - Secondary Network Connect Failure",
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
				failNetworkConnect: stringSet{
					"proxy-bridge": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to connect container g1-c1 to network proxy-bridge, reason: failed to connect container g1-c1 to network proxy-bridge on the fake docker host`,
	},
}

func TestContainerStartErrors(t *testing.T) {
	for _, test := range containerStartErrorTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			buf := new(bytes.Buffer)
			tc.ctxInfo.logger = newCapturingTestLogger(buf)
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

			gotStarted, gotErr := ct.start(ctx, dockerClient)
			if gotErr == nil {
				t.Errorf(
					"container.start()\nTest Case: %q\nFailure: gotErr == nil\n\nOut:\n%s\nReason: want = %q",
					tc.name, buf.String(), tc.want)
				return
			}
			if gotStarted {
				t.Errorf(
					"container.start()\nTest Case: %q\n\nOut:\n%s\nReason: gotStarted (%t) != wantStarted (false)",
					tc.name, buf.String(), gotStarted)
			}

			match, err := regexp.MatchString(fmt.Sprintf("^%s$", tc.want), gotErr.Error())
			if err != nil {
				t.Errorf(
					"container.start()\nTest Case: %q\nFailure: unexpected exception while matching against gotErr error string\n\nOut:\n%s\nReason: error = %v", tc.name, buf.String(), err)
				return
			}

			if !match {
				t.Errorf(
					"container.start()\nTest Case: %q\nFailure: gotErr did not match the want regex\n\nOut:\n%s\nReason:\n\ngotErr = %q\n\twant = %q", tc.name, buf.String(), gotErr, tc.want)
			}
		})
	}
}

func buildSingleContainerConfig(ct ContainerReference, image string) HomelabConfig {
	config := buildSingleContainerNoNetworkConfig(ct, image)
	config.IPAM = IPAMConfig{
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
				{
					Name:              "proxy-bridge",
					HostInterfaceName: "docker-prx",
					CIDR:              "172.18.201.0/24",
					Priority:          2,
					Containers: []ContainerIPConfig{
						{
							IP:        "172.18.201.11",
							Container: ct,
						},
					},
				},
			},
		},
	}
	return config
}

func buildSingleContainerNoNetworkConfig(ct ContainerReference, image string) HomelabConfig {
	return HomelabConfig{
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
