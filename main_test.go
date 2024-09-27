package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

var mainRunTests = []struct {
	name       string
	args       []string
	ctxInfo    *testContextInfo
	wantStatus int
	wantOutput string
}{
	{
		name: "Main - run() - Missing Subcommand",
		args: []string{},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
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
		ctxInfo: &testContextInfo{
			inspectLevel: homelabInspectLevelDebug,
			dockerHost:   newEmptyFakeDockerHost(),
		},
		wantStatus: 1,
		wantOutput: `(?s)Error: homelab sub-command is required
Usage:
.+
Use "homelab \[command\] --help" for more information about a command\.`,
	},
	{
		name: "Main - run() - Missing Subcommand - Trace Inspect Level",
		args: []string{},
		ctxInfo: &testContextInfo{
			inspectLevel: homelabInspectLevelTrace,
			dockerHost:   newEmptyFakeDockerHost(),
		},
		wantStatus: 1,
		wantOutput: `(?s)Error: homelab sub-command is required
Usage:
.+
Use "homelab \[command\] --help" for more information about a command\.`,
	},
	{
		name: "Main - run() - Start - All Groups",
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
		wantStatus: 0,
		wantOutput: `(?s).+INFO.+Pulling image: abc/xyz
.+INFO.+Created network net1
.+INFO.+Started container g1-c1
.+WARN.+Container g1-c2 not allowed to run on host FakeHost
.+INFO.+Pulling image: abc/xyz3
.+INFO.+Created network net2
.+INFO.+Started container g2-c3`,
	},
	{
		name: "Main - Start - Non Existing Configs Path",
		args: []string{
			"start",
			"--all-groups",
			"--configs-dir",
			fmt.Sprintf("%s/testdata/foobar", pwd()),
		},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		wantStatus: 1,
		wantOutput: `.+ERROR.+start failed while parsing the configs, reason: os\.Stat\(\) failed on homelab configs path, reason: stat .+/homelab/testdata/foobar: no such file or directory`,
	},
}

func TestMainRun(t *testing.T) {
	for _, tc := range mainRunTests {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			ctx := newTestContext(tc.ctxInfo)

			gotStatus := run(ctx, buf, buf, tc.args...)
			if gotStatus != tc.wantStatus {
				testLogCustomWithOutput(t, "run()", tc.name, buf, fmt.Sprintf("gotStatus (%d) != wantStatus (%d)", gotStatus, tc.wantStatus))
				return
			}

			if !testRegexMatch(t, "run()", tc.name, "output", tc.wantOutput, strings.TrimSpace(buf.String())) {
				return
			}
		})
	}
}

var mainRunEnvTests = []struct {
	name       string
	args       []string
	ctxInfo    *testContextInfo
	envs       testEnvMap
	wantStatus int
	wantOutput string
}{
	{
		name: "Main - run() - Missing Subcommand - Debug Inspect Level Using Env",
		args: []string{},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		envs: testEnvMap{
			"HOMELAB_INSPECT_LEVEL": "debug",
		},
		wantStatus: 1,
		wantOutput: `(?s)Error: homelab sub-command is required
Usage:
.+
Use "homelab \[command\] --help" for more information about a command\.`,
	},
	{
		name: "Main - run() - Missing Subcommand - Trace Inspect Level Using Env",
		args: []string{},
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		envs: testEnvMap{
			"HOMELAB_INSPECT_LEVEL": "trace",
		},
		wantStatus: 1,
		wantOutput: `(?s)Error: homelab sub-command is required
Usage:
.+
Use "homelab \[command\] --help" for more information about a command\.`,
	},
}

func TestMainRunEnv(t *testing.T) {
	for _, tc := range mainRunEnvTests {
		t.Run(tc.name, func(t *testing.T) {
			setTestEnv(t, tc.envs)
			buf := new(bytes.Buffer)
			ctx := newTestContext(tc.ctxInfo)

			gotStatus := run(ctx, buf, buf, tc.args...)
			if gotStatus != tc.wantStatus {
				testLogCustomWithOutput(t, "run()", tc.name, buf, fmt.Sprintf("gotStatus (%d) != wantStatus (%d)", gotStatus, tc.wantStatus))
				return
			}

			if !testRegexMatch(t, "run()", tc.name, "output", tc.wantOutput, strings.TrimSpace(buf.String())) {
				return
			}
		})
	}
}
