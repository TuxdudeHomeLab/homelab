package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/tuxdudehomelab/homelab/internal/docker/fakedocker"
	"github.com/tuxdudehomelab/homelab/internal/inspect"
	"github.com/tuxdudehomelab/homelab/internal/testhelpers"
	"github.com/tuxdudehomelab/homelab/internal/testutils"
	"github.com/tuxdudehomelab/homelab/internal/utils"
)

var mainRunWithContextTests = []struct {
	name       string
	args       []string
	ctxInfo    *testutils.TestContextInfo
	wantStatus int
	wantOutput string
}{
	{
		name: "Main - runWithContext() - Missing Subcommand",
		args: []string{},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		wantStatus: 1,
		wantOutput: `(?s)Error: homelab sub-command is required
Usage:
.+
Use "homelab \[command\] --help" for more information about a command\.`,
	},
	{
		name: "Main - Missing Subcommand - Debug Inspect Level",
		args: []string{},
		ctxInfo: &testutils.TestContextInfo{
			InspectLevel: inspect.HomelabInspectLevelDebug,
			DockerHost:   fakedocker.NewEmptyFakeDockerHost(),
		},
		wantStatus: 1,
		wantOutput: `(?s)Error: homelab sub-command is required
Usage:
.+
Use "homelab \[command\] --help" for more information about a command\.`,
	},
	{
		name: "Main - runWithContext() - Missing Subcommand - Trace Inspect Level",
		args: []string{},
		ctxInfo: &testutils.TestContextInfo{
			InspectLevel: inspect.HomelabInspectLevelTrace,
			DockerHost:   fakedocker.NewEmptyFakeDockerHost(),
		},
		wantStatus: 1,
		wantOutput: `(?s)Error: homelab sub-command is required
Usage:
.+
Use "homelab \[command\] --help" for more information about a command\.`,
	},
	{
		name: "Main - runWithContext() - Start - All Groups",
		args: []string{
			"group",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/main-start-all-groups", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewFakeDockerHost(&fakedocker.FakeDockerHostInitInfo{
				ValidImagesForPull: utils.StringSet{
					"abc/xyz":  {},
					"abc/xyz3": {},
				},
			}),
		},
		wantStatus: 0,
		wantOutput: `(?s).+INFO.+Pulling image: abc/xyz
.+INFO.+Created network net1
.+INFO.+Creating container g1-c1
.+INFO.+Starting container g1-c1
.+WARN.+Container g1-c2 not allowed to run on host FakeHost
.+INFO.+Pulling image: abc/xyz3
.+INFO.+Created network net2
.+INFO.+Creating container g2-c3
.+INFO.+Starting container g2-c3`,
	},
	{
		name: "Main - Start - Non Existing Configs Path",
		args: []string{
			"group",
			"start",
			"all",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/foobar", testhelpers.Pwd()),
		},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		wantStatus: 1,
		wantOutput: `.+ERROR.+start failed while parsing the configs, reason: os\.Stat\(\) failed on homelab configs path, reason: stat .+/homelab/testdata/foobar: no such file or directory`,
	},
}

func TestMainRunWithContext(t *testing.T) {
	t.Parallel()

	for _, tc := range mainRunWithContextTests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			buf := new(bytes.Buffer)
			ctx := testutils.NewTestContext(tc.ctxInfo)

			gotStatus := runWithContext(ctx, buf, buf, tc.args...)
			if gotStatus != tc.wantStatus {
				testhelpers.LogCustomWithOutput(t, "runWithContext()", tc.name, buf, fmt.Sprintf("gotStatus (%d) != wantStatus (%d)", gotStatus, tc.wantStatus))
				return
			}

			if !testhelpers.RegexMatch(t, "runWithContext()", tc.name, "output", tc.wantOutput, strings.TrimSpace(buf.String())) {
				return
			}
		})
	}
}

var mainRunEnvTests = []struct {
	name       string
	args       []string
	ctxInfo    *testutils.TestContextInfo
	envs       testhelpers.TestEnvMap
	wantStatus int
	wantOutput string
}{
	{
		name: "Main - runWithContext() - Missing Subcommand - Debug Inspect Level Using Env",
		args: []string{},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		envs: testhelpers.TestEnvMap{
			"HOMELAB_INSPECT_LEVEL": "debug",
		},
		wantStatus: 1,
		wantOutput: `(?s)Error: homelab sub-command is required
Usage:
.+
Use "homelab \[command\] --help" for more information about a command\.`,
	},
	{
		name: "Main - runWithContext() - Missing Subcommand - Trace Inspect Level Using Env",
		args: []string{},
		ctxInfo: &testutils.TestContextInfo{
			DockerHost: fakedocker.NewEmptyFakeDockerHost(),
		},
		envs: testhelpers.TestEnvMap{
			"HOMELAB_INSPECT_LEVEL": "trace",
		},
		wantStatus: 1,
		wantOutput: `(?s)Error: homelab sub-command is required
Usage:
.+
Use "homelab \[command\] --help" for more information about a command\.`,
	},
}

//nolint:paralleltest // Test sets environment variables.
func TestMainRunEnv(t *testing.T) {
	for _, tc := range mainRunEnvTests {
		t.Run(tc.name, func(t *testing.T) {
			testhelpers.SetTestEnv(t, tc.envs)
			buf := new(bytes.Buffer)
			ctx := testutils.NewTestContext(tc.ctxInfo)

			gotStatus := runWithContext(ctx, buf, buf, tc.args...)
			if gotStatus != tc.wantStatus {
				testhelpers.LogCustomWithOutput(t, "runWithContext()", tc.name, buf, fmt.Sprintf("gotStatus (%d) != wantStatus (%d)", gotStatus, tc.wantStatus))
				return
			}

			if !testhelpers.RegexMatch(t, "runWithContext()", tc.name, "output", tc.wantOutput, strings.TrimSpace(buf.String())) {
				return
			}
		})
	}
}

func TestMainRun(t *testing.T) {
	t.Parallel()

	tc := "Main - run() - Missing Subcommand"
	want := 1
	t.Run(tc, func(t *testing.T) {
		t.Parallel()

		got := run()
		if got != want {
			testhelpers.LogCustom(t, "run()", tc, fmt.Sprintf("gotStatus (%d) != wantStatus (%d)", got, want))
			return
		}
	})
}
