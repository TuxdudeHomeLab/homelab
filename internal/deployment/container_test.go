package deployment

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	dcontainer "github.com/docker/docker/api/types/container"
	"github.com/tuxdude/zzzlog"
	"github.com/tuxdudehomelab/homelab/internal/cmdexec/fakecmdexec"
	"github.com/tuxdudehomelab/homelab/internal/config"
	"github.com/tuxdudehomelab/homelab/internal/docker"
	"github.com/tuxdudehomelab/homelab/internal/docker/fakedocker"
	"github.com/tuxdudehomelab/homelab/internal/inspect"
	"github.com/tuxdudehomelab/homelab/internal/testhelpers"
	"github.com/tuxdudehomelab/homelab/internal/testutils"
	"github.com/tuxdudehomelab/homelab/internal/utils"
)

var containerStartTests = []struct {
	name           string
	config         config.Homelab
	cRef           config.ContainerReference
	ctxInfo        *testutils.TestContextInfo
	preExec        func(context.Context)
	wantNotStarted bool
}{
	{
		name: "Container Start - Doesn't Exist Already",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Doesn't Exist Already - With Start Pre-Hook",
		config: buildCustomSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz",
			func(ct *config.Container) {
				ct.Lifecycle.StartPreHook = []string{
					"custom-start-prehook",
					"arg1",
					"arg2",
				}
			},
		),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
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
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Doesn't Exist Already - Skip Image Pull",
		config: buildCustomSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz",
			func(ct *config.Container) {
				ct.Image.SkipImagePull = true
			},
		),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
	},
	{
		name: "Container Start - Doesn't Exist Already - Ignore Image Pull Failures",
		config: buildCustomSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz",
			func(ct *config.Container) {
				ct.Image.IgnoreImagePullFailures = true
			},
		),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
				FailImagePull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Doesn't Exist Already - Existing Image",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ExistingImages: utils.StringSet{
					"abc/xyz": {},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Doesn't Exist Already - No Network Endpoints",
		config: buildSingleContainerNoNetworkConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Created State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateCreated,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Running State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
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
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Paused State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStatePaused,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Restarting State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRestarting,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Removing State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRemoving,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		preExec: func(ctx context.Context) {
			go func() {
				time.Sleep(200 * time.Millisecond)
				d := fakedocker.FakeDockerHostFromContext(ctx)
				err := d.ForceRemoveContainer("g1-c1")
				if err != nil {
					panic(err)
				}
			}()
		},
	},
	{
		name: "Container Start - Exists Already In Exited State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateExited,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Dead State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateDead,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Running State Requiring Multiple Stops",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:               "g1-c1",
						Image:              "abc/xyz",
						State:              docker.ContainerStateRunning,
						RequiredExtraStops: 5,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Exists Already In Running State Requiring Multiple Kills",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:               "g1-c1",
						Image:              "abc/xyz",
						State:              docker.ContainerStateRunning,
						RequiredExtraStops: 1000,
						RequiredExtraKills: 4,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
	{
		name: "Container Start - Doesn't Exist Already - Container Mode Network",
		config: buildSingleContainerWithContainerModeNetworkConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz",
			config.ContainerReference{
				Group:     "g1",
				Container: "c2",
			}),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
	},
}

func TestContainerStart(t *testing.T) {
	t.Parallel()

	for _, test := range containerStartTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			buf := new(bytes.Buffer)
			tc.ctxInfo.Logger = testutils.NewCapturingTestLogger(zzzlog.LvlDebug, buf)
			// Enable debug inspect level while running the container start tests
			// for extra code coverage.
			if tc.ctxInfo.InspectLevel == inspect.HomelabInspectLevelNone {
				tc.ctxInfo.InspectLevel = inspect.HomelabInspectLevelDebug
			}
			if tc.ctxInfo.ContainerPurgeKillDelay == 0 {
				// Reduce this delay to keep the tests executing quickly.
				tc.ctxInfo.ContainerPurgeKillDelay = 100 * time.Millisecond
			}
			ctx := testutils.NewTestContext(tc.ctxInfo)

			dep, gotErr := FromConfig(ctx, &tc.config)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "FromConfig()", tc.name, gotErr)
				return
			}

			dc := docker.NewClient(ctx)
			defer dc.Close()

			ct, gotErr := dep.queryContainer(tc.cRef)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "deployment.queryContainer()", tc.name, gotErr)
				return
			}

			if tc.preExec != nil {
				tc.preExec(ctx)
			}

			gotStarted, gotErr := ct.Start(ctx, dc)
			if gotErr != nil {
				testhelpers.LogErrorNotNilWithOutput(t, "container.start()", tc.name, buf, gotErr)
				return
			}
			wantStarted := !tc.wantNotStarted
			if gotStarted != wantStarted {
				testhelpers.LogCustomWithOutput(t, "container.start()", tc.name, buf, fmt.Sprintf("gotStarted (%t) != wantStarted (%t)", gotStarted, wantStarted))
			}

			d := fakedocker.FakeDockerHostFromContext(ctx)
			gotState := d.GetContainerState(fmt.Sprintf("%s-%s", tc.cRef.Group, tc.cRef.Container))
			if gotState != docker.ContainerStateRunning {
				testhelpers.LogCustomWithOutput(t, "Container state after container.start()", tc.name, buf, fmt.Sprintf("gotState (%s) != ContainerStateRunning", gotState))
			}
		})
	}
}

var containerStartErrorTests = []struct {
	name      string
	config    config.Homelab
	cRef      config.ContainerReference
	ctxInfo   *testutils.TestContextInfo
	wantPanic bool
	want      string
}{
	{
		name: "Container Start - Doesn't Exist Already - With Start Pre-Hook",
		config: buildCustomSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz",
			func(ct *config.Container) {
				ct.Lifecycle.StartPreHook = []string{
					"custom-start-prehook",
					"arg1",
					"arg2",
				}
			},
		),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			Executor: fakecmdexec.NewFakeExecutor(&fakecmdexec.FakeExecutorInitInfo{
				ErrorCmds: []fakecmdexec.FakeErrorCmdInfo{
					{
						Cmd: []string{
							"custom-start-prehook",
							"arg1",
							"arg2",
						},
						Err: fmt.Errorf("custom-start-prehook command not found"),
					},
				},
			}),
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:encountered error while running the start pre-hook for container g1-c1, reason: custom-start-prehook command not found`,
	},
	{
		name: "Container Start - Image Not Available",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		want: `Failed to start container g1-c1, reason:failed to pull the image abc/xyz, reason: image abc/xyz not found or invalid and cannot be pulled by the fake docker host`,
	},
	{
		name: "Container Start - Image Pull Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
				FailImagePull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed while pulling the image abc/xyz, reason: failed to pull image abc/xyz on the fake docker host`,
	},
	{
		name: "Container Start - No Local Image After Pull",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
				NoImageAfterPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:image abc/xyz not available locally after a successful pull, possibly indicating a bug or a system failure!`,
	},
	{
		name: "Container Start - Kill Existing Container Fails",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:               "g1-c1",
						Image:              "abc/xyz",
						State:              docker.ContainerStateRunning,
						RequiredExtraStops: 1000,
					},
				},
				FailContainerKill: utils.StringSet{
					"g1-c1": {},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to purge container g1-c1 after 6 attempts`,
	},
	{
		name: "Container Start - Unkillable Existing Container",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:               "g1-c1",
						Image:              "abc/xyz",
						State:              docker.ContainerStateRunning,
						RequiredExtraStops: 1000,
						RequiredExtraKills: 5,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to purge container g1-c1 after 6 attempts`,
	},
	{
		name: "Container Start - Container State Unknown",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateUnknown,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		wantPanic: true,
		want:      `container g1-c1 is in an unsupported state Unknown, possibly indicating a bug in the code`,
	},
	{
		name: "Container Start - Inspect Existing Container Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
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
				FailContainerInspect: utils.StringSet{
					"g1-c1": {},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to retrieve the container state, reason: failed to inspect container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Start - Stop Existing Container Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
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
				FailContainerStop: utils.StringSet{
					"g1-c1": {},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to stop the container, reason: failed to stop container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Start - Remove Existing Container Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRestarting,
					},
				},
				FailContainerRemove: utils.StringSet{
					"g1-c1": {},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to remove the container, reason: failed to remove container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Start - Create Container Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateDead,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
				FailContainerCreate: utils.StringSet{
					"g1-c1": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to create the container, reason: failed to create container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Start - Start Container Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStatePaused,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
				FailContainerStart: utils.StringSet{
					"g1-c1": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to start the container, reason: failed to start container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Start - Primary Network Create Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStatePaused,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
				FailNetworkCreate: utils.StringSet{
					"g1-bridge": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to create the network, reason: failed to create network g1-bridge on the fake docker host`,
	},
	{
		name: "Container Start - Secondary Network Create Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStatePaused,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
				FailNetworkCreate: utils.StringSet{
					"proxy-bridge": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to create the network, reason: failed to create network proxy-bridge on the fake docker host`,
	},
	{
		name: "Container Start - Secondary Network Connect Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStatePaused,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
				FailNetworkConnect: utils.StringSet{
					"proxy-bridge": {},
				},
			}),
		},
		want: `Failed to start container g1-c1, reason:failed to connect container g1-c1 to network proxy-bridge, reason: failed to connect container g1-c1 to network proxy-bridge on the fake docker host`,
	},
}

func TestContainerStartErrors(t *testing.T) {
	t.Parallel()

	for _, test := range containerStartErrorTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			buf := new(bytes.Buffer)
			tc.ctxInfo.Logger = testutils.NewCapturingTestLogger(zzzlog.LvlDebug, buf)
			if tc.ctxInfo.ContainerPurgeKillDelay == 0 {
				// Reduce this delay to keep the tests executing quickly.
				tc.ctxInfo.ContainerPurgeKillDelay = 100 * time.Millisecond
			}
			ctx := testutils.NewTestContext(tc.ctxInfo)

			if tc.wantPanic {
				defer testhelpers.ExpectPanicWithOutput(t, "container.start()", tc.name, buf, tc.want)
			}

			dep, gotErr := FromConfig(ctx, &tc.config)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "FromConfig()", tc.name, gotErr)
				return
			}

			dc := docker.NewClient(ctx)
			defer dc.Close()

			ct, gotErr := dep.queryContainer(tc.cRef)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "deployment.queryContainer()", tc.name, gotErr)
				return
			}

			gotStarted, gotErr := ct.Start(ctx, dc)
			if gotErr == nil {
				testhelpers.LogErrorNilWithOutput(t, "container.start()", tc.name, buf, tc.want)
				return
			}
			if gotStarted {
				testhelpers.LogCustomWithOutput(t, "container.start()", tc.name, buf, "gotStarted (true) != wantStarted (false)")
				return
			}

			if !testhelpers.RegexMatchWithOutput(t, "container.start()", tc.name, buf, "gotErr error string", tc.want, gotErr.Error()) {
				return
			}
		})
	}
}

var containerStopTests = []struct {
	name                    string
	config                  config.Homelab
	cRef                    config.ContainerReference
	ctxInfo                 *testutils.TestContextInfo
	preExec                 func(context.Context)
	wantContainerStopIssued bool
	wantStoppedReturnVal    bool
	wantState               docker.ContainerState
}{
	{
		name: "Container Stop - Doesn't Exist Already",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		wantContainerStopIssued: false,
		wantStoppedReturnVal:    false,
		wantState:               docker.ContainerStateNotFound,
	},
	{
		name: "Container Stop - Exists Already In Created State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateCreated,
					},
				},
			}),
		},
		wantContainerStopIssued: false,
		wantStoppedReturnVal:    true,
		wantState:               docker.ContainerStateCreated,
	},
	{
		name: "Container Stop - Exists Already In Running State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
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
		wantContainerStopIssued: true,
		wantStoppedReturnVal:    true,
		wantState:               docker.ContainerStateExited,
	},
	{
		name: "Container Stop - Exists Already In Running State - Pull Image Before Stop",
		config: buildCustomSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz",
			func(ct *config.Container) {
				ct.Image.PullImageBeforeStop = true
			},
		),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
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
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		wantContainerStopIssued: true,
		wantStoppedReturnVal:    true,
		wantState:               docker.ContainerStateExited,
	},
	{
		name: "Container Stop - Exists Already In Running State - Pull Image Before Stop - Ignore Image Pull Failures",
		config: buildCustomSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz",
			func(ct *config.Container) {
				ct.Image.PullImageBeforeStop = true
				ct.Image.IgnoreImagePullFailures = true
			},
		),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
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
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
				FailImagePull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		wantContainerStopIssued: true,
		wantStoppedReturnVal:    true,
		wantState:               docker.ContainerStateExited,
	},
	{
		name: "Container Stop - Exists Already In Paused State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStatePaused,
					},
				},
			}),
		},
		wantContainerStopIssued: true,
		wantStoppedReturnVal:    true,
		wantState:               docker.ContainerStateExited,
	},
	{
		name: "Container Stop - Exists Already In Restarting State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRestarting,
					},
				},
			}),
		},
		wantContainerStopIssued: true,
		wantStoppedReturnVal:    true,
		wantState:               docker.ContainerStateExited,
	},
	{
		name: "Container Stop - Exists Already In Removing State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRemoving,
					},
				},
			}),
		},
		wantContainerStopIssued: false,
		wantStoppedReturnVal:    true,
		wantState:               docker.ContainerStateRemoving,
	},
	{
		name: "Container Stop - Exists Already In Exited State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateExited,
					},
				},
			}),
		},
		wantContainerStopIssued: false,
		wantStoppedReturnVal:    true,
		wantState:               docker.ContainerStateExited,
	},
	{
		name: "Container Stop - Exists Already In Dead State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateDead,
					},
				},
			}),
		},
		wantContainerStopIssued: false,
		wantStoppedReturnVal:    true,
		wantState:               docker.ContainerStateDead,
	},
}

func TestContainerStop(t *testing.T) {
	t.Parallel()

	for _, test := range containerStopTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			buf := new(bytes.Buffer)
			tc.ctxInfo.Logger = testutils.NewCapturingTestLogger(zzzlog.LvlDebug, buf)
			// Enable debug inspect level while running the container start tests
			// for extra code coverage.
			if tc.ctxInfo.InspectLevel == inspect.HomelabInspectLevelNone {
				tc.ctxInfo.InspectLevel = inspect.HomelabInspectLevelDebug
			}
			if tc.ctxInfo.ContainerPurgeKillDelay == 0 {
				// Reduce this delay to keep the tests executing quickly.
				tc.ctxInfo.ContainerPurgeKillDelay = 100 * time.Millisecond
			}
			ctx := testutils.NewTestContext(tc.ctxInfo)

			dep, gotErr := FromConfig(ctx, &tc.config)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "FromConfig()", tc.name, gotErr)
				return
			}

			dc := docker.NewClient(ctx)
			defer dc.Close()

			ct, gotErr := dep.queryContainer(tc.cRef)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "deployment.queryContainer()", tc.name, gotErr)
				return
			}

			if tc.preExec != nil {
				tc.preExec(ctx)
			}

			gotStoppedReturnVal, gotErr := ct.Stop(ctx, dc)
			if gotErr != nil {
				testhelpers.LogErrorNotNilWithOutput(t, "container.stop()", tc.name, buf, gotErr)
				return
			}
			if gotStoppedReturnVal != tc.wantStoppedReturnVal {
				testhelpers.LogCustomWithOutput(t, "container.stop() return value", tc.name, buf, fmt.Sprintf("gotStopped (%t) != wantStopped (%t)", gotStoppedReturnVal, tc.wantStoppedReturnVal))
			}

			cName := fmt.Sprintf("%s-%s", tc.cRef.Group, tc.cRef.Container)
			d := fakedocker.FakeDockerHostFromContext(ctx)
			gotStopIssued := d.ContainerStopIssued(cName)
			if gotStopIssued != tc.wantContainerStopIssued {
				testhelpers.LogCustomWithOutput(t, "ContainerStop issued", tc.name, buf, fmt.Sprintf("got (%t) != want (%t)", gotStopIssued, tc.wantContainerStopIssued))
			}

			gotState := d.GetContainerState(cName)
			if gotState != tc.wantState {
				testhelpers.LogCustomWithOutput(t, "Container state after container.stop()", tc.name, buf, fmt.Sprintf("got (%s) != want (%s)", gotState, tc.wantState))
			}
		})
	}
}

var containerStopErrorTests = []struct {
	name      string
	config    config.Homelab
	cRef      config.ContainerReference
	ctxInfo   *testutils.TestContextInfo
	wantPanic bool
	want      string
}{
	{
		name: "Container Stop - Stop Existing Container Fails",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
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
				FailContainerStop: utils.StringSet{
					"g1-c1": {},
				},
			}),
		},
		want: `Failed to stop container g1-c1, reason:failed to stop the container, reason: failed to stop container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Stop - Container State Unknown",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateUnknown,
					},
				},
			}),
		},
		wantPanic: true,
		want:      `container g1-c1 is in an unsupported state Unknown, possibly indicating a bug in the code`,
	},
	{
		name: "Container Stop - Inspect Existing Container Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
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
				FailContainerInspect: utils.StringSet{
					"g1-c1": {},
				},
			}),
		},
		want: `Failed to stop container g1-c1, reason:failed to retrieve the container state, reason: failed to inspect container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Stop - Exists Already In Running State - Pull Image Before Stop - Image Pull Failure",
		config: buildCustomSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz",
			func(ct *config.Container) {
				ct.Image.PullImageBeforeStop = true
			},
		),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
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
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
				FailImagePull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to stop container g1-c1, reason:failed while pulling the image abc/xyz, reason: failed to pull image abc/xyz on the fake docker host`,
	},
}

func TestContainerStopErrors(t *testing.T) {
	t.Parallel()

	for _, test := range containerStopErrorTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			buf := new(bytes.Buffer)
			tc.ctxInfo.Logger = testutils.NewCapturingTestLogger(zzzlog.LvlDebug, buf)
			if tc.ctxInfo.ContainerPurgeKillDelay == 0 {
				// Reduce this delay to keep the tests executing quickly.
				tc.ctxInfo.ContainerPurgeKillDelay = 100 * time.Millisecond
			}
			ctx := testutils.NewTestContext(tc.ctxInfo)

			if tc.wantPanic {
				defer testhelpers.ExpectPanicWithOutput(t, "container.start()", tc.name, buf, tc.want)
			}

			dep, gotErr := FromConfig(ctx, &tc.config)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "FromConfig()", tc.name, gotErr)
				return
			}

			dc := docker.NewClient(ctx)
			defer dc.Close()

			ct, gotErr := dep.queryContainer(tc.cRef)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "deployment.queryContainer()", tc.name, gotErr)
				return
			}

			gotStoppedReturnVal, gotErr := ct.Stop(ctx, dc)
			if gotErr == nil {
				testhelpers.LogErrorNilWithOutput(t, "container.stop()", tc.name, buf, tc.want)
				return
			}
			if gotStoppedReturnVal {
				testhelpers.LogCustomWithOutput(t, "container.stop() return value", tc.name, buf, "gotStopped (true) != wantStopped (false)")
			}
			if !testhelpers.RegexMatchWithOutput(t, "container.stop()", tc.name, buf, "gotErr error string", tc.want, gotErr.Error()) {
				return
			}
		})
	}
}

var containerPurgeTests = []struct {
	name       string
	config     config.Homelab
	cRef       config.ContainerReference
	ctxInfo    *testutils.TestContextInfo
	preExec    func(context.Context)
	wantPurged bool
}{
	{
		name: "Container Purge - Doesn't Exist Already",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{}),
		},
		wantPurged: false,
	},
	{
		name: "Container Purge - In Created State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateCreated,
					},
				},
			}),
		},
		wantPurged: true,
	},
	{
		name: "Container Purge - In Running State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
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
		wantPurged: true,
	},
	{
		name: "Container Purge - In Paused State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStatePaused,
					},
				},
			}),
		},
		wantPurged: true,
	},
	{
		name: "Container Purge - In Restarting State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRestarting,
					},
				},
			}),
		},
		wantPurged: true,
	},
	{
		name: "Container Purge - In Removing State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRemoving,
					},
				},
			}),
		},
		preExec: func(ctx context.Context) {
			go func() {
				time.Sleep(200 * time.Millisecond)
				d := fakedocker.FakeDockerHostFromContext(ctx)
				err := d.ForceRemoveContainer("g1-c1")
				if err != nil {
					panic(err)
				}
			}()
		},
		wantPurged: true,
	},
	{
		name: "Container Purge - In Exited State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateExited,
					},
				},
			}),
		},
		wantPurged: true,
	},
	{
		name: "Container Purge - In Dead State",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateDead,
					},
				},
			}),
		},
		wantPurged: true,
	},
	{
		name: "Container Purge - In Running State Requiring Multiple Stops",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:               "g1-c1",
						Image:              "abc/xyz",
						State:              docker.ContainerStateRunning,
						RequiredExtraStops: 5,
					},
				},
			}),
		},
		wantPurged: true,
	},
	{
		name: "Container Purge - In Running State Requiring Multiple Kills",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:               "g1-c1",
						Image:              "abc/xyz",
						State:              docker.ContainerStateRunning,
						RequiredExtraStops: 1000,
						RequiredExtraKills: 4,
					},
				},
			}),
		},
		wantPurged: true,
	},
}

func TestContainerPurge(t *testing.T) {
	t.Parallel()

	for _, test := range containerPurgeTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			buf := new(bytes.Buffer)
			tc.ctxInfo.Logger = testutils.NewCapturingTestLogger(zzzlog.LvlDebug, buf)
			// Enable debug inspect level while running the container start tests
			// for extra code coverage.
			if tc.ctxInfo.InspectLevel == inspect.HomelabInspectLevelNone {
				tc.ctxInfo.InspectLevel = inspect.HomelabInspectLevelDebug
			}
			if tc.ctxInfo.ContainerPurgeKillDelay == 0 {
				// Reduce this delay to keep the tests executing quickly.
				tc.ctxInfo.ContainerPurgeKillDelay = 100 * time.Millisecond
			}
			ctx := testutils.NewTestContext(tc.ctxInfo)

			dep, gotErr := FromConfig(ctx, &tc.config)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "FromConfig()", tc.name, gotErr)
				return
			}

			dc := docker.NewClient(ctx)
			defer dc.Close()

			ct, gotErr := dep.queryContainer(tc.cRef)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "deployment.queryContainer()", tc.name, gotErr)
				return
			}

			if tc.preExec != nil {
				tc.preExec(ctx)
			}

			gotPurged, gotErr := ct.Purge(ctx, dc)
			if gotErr != nil {
				testhelpers.LogErrorNotNilWithOutput(t, "container.Purge()", tc.name, buf, gotErr)
				return
			}
			if gotPurged != tc.wantPurged {
				testhelpers.LogCustomWithOutput(t, "container.Purge()", tc.name, buf, fmt.Sprintf("gotPurged (%t) != wantPurged (%t)", gotPurged, tc.wantPurged))
			}

			d := fakedocker.FakeDockerHostFromContext(ctx)
			gotState := d.GetContainerState(fmt.Sprintf("%s-%s", tc.cRef.Group, tc.cRef.Container))
			if gotState != docker.ContainerStateNotFound {
				testhelpers.LogCustomWithOutput(t, "Container state after container.Purge()", tc.name, buf, fmt.Sprintf("gotState (%s) when container is expected to be purged", gotState))
			}
		})
	}
}

var containerPurgeErrorTests = []struct {
	name      string
	config    config.Homelab
	cRef      config.ContainerReference
	ctxInfo   *testutils.TestContextInfo
	wantPanic bool
	want      string
}{
	{
		name: "Container Purge - Kill Existing Container Fails",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:               "g1-c1",
						Image:              "abc/xyz",
						State:              docker.ContainerStateRunning,
						RequiredExtraStops: 1000,
					},
				},
				FailContainerKill: utils.StringSet{
					"g1-c1": {},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to purge container g1-c1, reason:failed to purge container g1-c1 after 6 attempts`,
	},
	{
		name: "Container Purge - Unkillable Existing Container",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:               "g1-c1",
						Image:              "abc/xyz",
						State:              docker.ContainerStateRunning,
						RequiredExtraStops: 1000,
						RequiredExtraKills: 5,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to purge container g1-c1, reason:failed to purge container g1-c1 after 6 attempts`,
	},
	{
		name: "Container Purge - Container State Unknown",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateUnknown,
					},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		wantPanic: true,
		want:      `container g1-c1 is in an unsupported state Unknown, possibly indicating a bug in the code`,
	},
	{
		name: "Container Purge - Inspect Existing Container Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
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
				FailContainerInspect: utils.StringSet{
					"g1-c1": {},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to purge container g1-c1, reason:failed to retrieve the container state, reason: failed to inspect container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Purge - Stop Existing Container Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
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
				FailContainerStop: utils.StringSet{
					"g1-c1": {},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to purge container g1-c1, reason:failed to stop the container, reason: failed to stop container g1-c1 on the fake docker host`,
	},
	{
		name: "Container Purge - Remove Existing Container Failure",
		config: buildSingleContainerConfig(
			config.ContainerReference{
				Group:     "g1",
				Container: "c1",
			},
			"abc/xyz"),
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				Containers: []*fakedocker.FakeContainerInitInfo{
					{
						Name:  "g1-c1",
						Image: "abc/xyz",
						State: docker.ContainerStateRestarting,
					},
				},
				FailContainerRemove: utils.StringSet{
					"g1-c1": {},
				},
				ValidImagesForPull: utils.StringSet{
					"abc/xyz": {},
				},
			}),
		},
		want: `Failed to purge container g1-c1, reason:failed to remove the container, reason: failed to remove container g1-c1 on the fake docker host`,
	},
}

func TestContainerPurgeErrors(t *testing.T) {
	t.Parallel()

	for _, test := range containerPurgeErrorTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			buf := new(bytes.Buffer)
			tc.ctxInfo.Logger = testutils.NewCapturingTestLogger(zzzlog.LvlDebug, buf)
			if tc.ctxInfo.ContainerPurgeKillDelay == 0 {
				// Reduce this delay to keep the tests executing quickly.
				tc.ctxInfo.ContainerPurgeKillDelay = 100 * time.Millisecond
			}
			ctx := testutils.NewTestContext(tc.ctxInfo)

			if tc.wantPanic {
				defer testhelpers.ExpectPanicWithOutput(t, "container.Purge()", tc.name, buf, tc.want)
			}

			dep, gotErr := FromConfig(ctx, &tc.config)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "FromConfig()", tc.name, gotErr)
				return
			}

			dc := docker.NewClient(ctx)
			defer dc.Close()

			ct, gotErr := dep.queryContainer(tc.cRef)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "deployment.queryContainer()", tc.name, gotErr)
				return
			}

			gotPurged, gotErr := ct.Purge(ctx, dc)
			if gotErr == nil {
				testhelpers.LogErrorNilWithOutput(t, "container.Purge()", tc.name, buf, tc.want)
				return
			}
			if gotPurged {
				testhelpers.LogCustomWithOutput(t, "container.Purge()", tc.name, buf, "gotPurged (true) != wantPurged (false)")
				return
			}

			if !testhelpers.RegexMatchWithOutput(t, "container.Purge()", tc.name, buf, "gotErr error string", tc.want, gotErr.Error()) {
				return
			}
		})
	}
}

var containerDockerConfigTests = []struct {
	name              string
	config            config.Homelab
	cRef              config.ContainerReference
	ctxInfo           *testutils.TestContextInfo
	wantDockerConfigs *containerDockerConfigs
}{
	{
		name: "Container Docker Configs - Mounts",
		config: config.Homelab{
			Global: config.Global{
				BaseDir: testhelpers.HomelabBaseDir(),
				MountDefs: []config.Mount{
					{
						Name:     "mount-def-1",
						Type:     "bind",
						Src:      "/abc/def/ghi",
						Dst:      "/pqr/stu/vwx",
						ReadOnly: true,
					},
					{
						Name: "mount-def-2",
						Type: "bind",
						Src:  "/abc1/def1",
						Dst:  "/pqr2/stu2/vwx2",
					},
					{
						Name: "homelab-self-signed-tls-cert",
						Type: "bind",
						Src:  "/path/to/my/self/signed/cert/on/host",
						Dst:  "/path/to/my/self/signed/cert/on/container",
					},
				},
				Container: config.GlobalContainer{
					Mounts: []config.Mount{
						{
							Name: "mount-def-1",
						},
						{
							Name: "mount-def-2",
						},
						{
							Name:     "mount-def-3",
							Src:      "/foo",
							Dst:      "/bar",
							ReadOnly: true,
						},
					},
				},
			},
			Groups: []config.ContainerGroup{
				{
					Name:  "g1",
					Order: 1,
				},
			},
			Containers: []config.Container{
				{
					Info: config.ContainerReference{
						Group:     "g1",
						Container: "c1",
					},
					Image: config.ContainerImage{
						Image: "abc/xyz:latest",
					},
					Lifecycle: config.ContainerLifecycle{
						Order: 1,
					},
					Filesystem: config.ContainerFilesystem{
						Mounts: []config.Mount{
							{
								Name: "some-other-mount-1",
								Type: "bind",
								Src:  "testdata/dummy-base-dir/abc",
								Dst:  "/abc",
							},
							{
								Name: "some-other-mount-2",
								Type: "bind",
								Src:  "testdata/dummy-base-dir/g1/c1/some/random/dir",
								Dst:  "/xyz",
							},
							{
								Name:     "blocky-config-mount",
								Type:     "bind",
								Src:      "testdata/dummy-base-dir/g1/c1/configs/generated/config.yml",
								Dst:      "/data/blocky/config/config.yml",
								ReadOnly: true,
							},
							{
								Name: "homelab-self-signed-tls-cert",
							},
							{
								Name: "some-other-mount-3",
								Type: "bind",
								Src:  "testdata/dummy-base-dir/g1/c1/data/my-data",
								Dst:  "/foo123/bar123/my-data",
							},
						},
					},
				},
			},
		},
		cRef: config.ContainerReference{
			Group:     "g1",
			Container: "c1",
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		wantDockerConfigs: &containerDockerConfigs{
			ContainerConfig: &dcontainer.Config{
				Image: "abc/xyz:latest",
			},
			HostConfig: &dcontainer.HostConfig{
				Binds: []string{
					"/abc/def/ghi:/pqr/stu/vwx:ro",
					"/abc1/def1:/pqr2/stu2/vwx2",
					"/foo:/bar:ro",
					"testdata/dummy-base-dir/abc:/abc",
					"testdata/dummy-base-dir/g1/c1/some/random/dir:/xyz",
					"testdata/dummy-base-dir/g1/c1/configs/generated/config.yml:/data/blocky/config/config.yml:ro",
					"/path/to/my/self/signed/cert/on/host:/path/to/my/self/signed/cert/on/container",
					"testdata/dummy-base-dir/g1/c1/data/my-data:/foo123/bar123/my-data",
				},
				NetworkMode: "none",
			},
		},
	},
}

func TestContainerDockerConfigs(t *testing.T) {
	t.Parallel()

	for _, test := range containerDockerConfigTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			buf := new(bytes.Buffer)
			tc.ctxInfo.Logger = testutils.NewCapturingTestLogger(zzzlog.LvlDebug, buf)
			// Enable debug inspect level while running the container start tests
			// for extra code coverage.
			if tc.ctxInfo.InspectLevel == inspect.HomelabInspectLevelNone {
				tc.ctxInfo.InspectLevel = inspect.HomelabInspectLevelDebug
			}
			if tc.ctxInfo.ContainerPurgeKillDelay == 0 {
				// Reduce this delay to keep the tests executing quickly.
				tc.ctxInfo.ContainerPurgeKillDelay = 100 * time.Millisecond
			}
			ctx := testutils.NewTestContext(tc.ctxInfo)

			dep, gotErr := FromConfig(ctx, &tc.config)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "FromConfig()", tc.name, gotErr)
				return
			}

			ct, gotErr := dep.queryContainer(tc.cRef)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "deployment.queryContainer()", tc.name, gotErr)
				return
			}

			got := ct.generateDockerConfigs()
			if !testhelpers.CmpDiff(t, "container.generateDockerConfigs()", tc.name, "docker configs", tc.wantDockerConfigs, got) {
				return
			}
		})
	}
}

func buildCustomSingleContainerConfig(ct config.ContainerReference, image string, fn func(*config.Container)) config.Homelab {
	h := buildSingleContainerConfig(ct, image)
	fn(&h.Containers[0])
	return h
}

func buildSingleContainerConfig(ct config.ContainerReference, image string) config.Homelab {
	conf := buildSingleContainerNoNetworkConfig(ct, image)
	conf.IPAM = config.IPAM{
		Networks: config.Networks{
			BridgeModeNetworks: []config.BridgeModeNetwork{
				{
					Name:              fmt.Sprintf("%s-bridge", ct.Group),
					HostInterfaceName: fmt.Sprintf("docker-%s", ct.Group),
					CIDR:              "172.18.101.0/24",
					Priority:          1,
					Containers: []config.ContainerIP{
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
					Containers: []config.ContainerIP{
						{
							IP:        "172.18.201.11",
							Container: ct,
						},
					},
				},
			},
		},
	}
	return conf
}

func buildSingleContainerWithContainerModeNetworkConfig(ct config.ContainerReference, image string, connectTo config.ContainerReference) config.Homelab {
	conf := buildSingleContainerNoNetworkConfig(ct, image)
	conf.IPAM = config.IPAM{
		Networks: config.Networks{
			ContainerModeNetworks: []config.ContainerModeNetwork{
				{
					Name:      "net1",
					Container: connectTo,
					AttachingContainers: []config.ContainerReference{
						ct,
					},
				},
			},
		},
	}
	return conf
}

func buildSingleContainerNoNetworkConfig(ct config.ContainerReference, image string) config.Homelab {
	return config.Homelab{
		Global: config.Global{
			BaseDir: testhelpers.HomelabBaseDir(),
		},
		Hosts: []config.Host{
			{
				Name: "fakehost",
				AllowedContainers: []config.ContainerReference{
					ct,
				},
			},
		},
		Groups: []config.ContainerGroup{
			{
				Name:  ct.Group,
				Order: 1,
			},
		},
		Containers: []config.Container{
			{
				Info: ct,
				Image: config.ContainerImage{
					Image: image,
				},
				Lifecycle: config.ContainerLifecycle{
					Order: 1,
				},
			},
		},
	}
}
