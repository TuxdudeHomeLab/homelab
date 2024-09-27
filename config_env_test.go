package main

import (
	"context"
	"testing"
)

var configEnvTests = []struct {
	name    string
	ctxInfo *testContextInfo
	test    func(*testing.T, context.Context, string, *configEnv)
}{
	{
		name: "Config Env - New",
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		test: func(t *testing.T, ctx context.Context, tc string, env *configEnv) {
			wantEnvMap := envMap{
				"$$HOST_IP$$":                  "10.76.77.78",
				"$$HOST_NAME$$":                "fakehost",
				"$$HUMAN_FRIENDLY_HOST_NAME$$": "FakeHost",
			}
			wantKeyOrder := []string{
				"$$HOST_IP$$",
				"$$HOST_NAME$$",
				"$$HUMAN_FRIENDLY_HOST_NAME$$",
			}

			if !testCmpDiff(t, "newConfigEnv()", tc, "configEnv struct env map", wantEnvMap, env.env) {
				return
			}
			if !testCmpDiff(t, "newConfigEnv()", tc, "configEnv struct env key order", wantKeyOrder, env.envKeyOrder) {
				return
			}
		},
	},
	{
		name: "Config Env - Override - No overlap",
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		test: func(t *testing.T, ctx context.Context, tc string, env *configEnv) {
			override := envMap{
				"FOO1": "foo1",
				"BAR1": "bar1",
				"BAZ1": "baz1",
			}
			overrideOrder := []string{
				"FOO1",
				"BAR1",
				"BAZ1",
			}
			wantEnvMap := envMap{
				"$$HOST_IP$$":                  "10.76.77.78",
				"$$HOST_NAME$$":                "fakehost",
				"$$HUMAN_FRIENDLY_HOST_NAME$$": "FakeHost",
				"$$FOO1$$":                     "foo1",
				"$$BAR1$$":                     "bar1",
				"$$BAZ1$$":                     "baz1",
			}
			wantKeyOrder := []string{
				"$$HOST_IP$$",
				"$$HOST_NAME$$",
				"$$HUMAN_FRIENDLY_HOST_NAME$$",
				"$$FOO1$$",
				"$$BAR1$$",
				"$$BAZ1$$",
			}

			got := env.override(ctx, override, overrideOrder)
			if !testCmpDiff(t, "newConfigEnv()", tc, "configEnv struct", wantEnvMap, got.env) {
				return
			}
			if !testCmpDiff(t, "newConfigEnv()", tc, "configEnv struct env key order", wantKeyOrder, got.envKeyOrder) {
				return
			}
		},
	},
	{
		name: "Config Env - Override - With overlap",
		ctxInfo: &testContextInfo{
			dockerHost: newEmptyFakeDockerHost(),
		},
		test: func(t *testing.T, ctx context.Context, tc string, env *configEnv) {
			override := envMap{
				"FOO1":                     "foo1",
				"HOST_IP":                  "10.76.77.178",
				"BAR1":                     "bar1",
				"HUMAN_FRIENDLY_HOST_NAME": "FakeHost2",
			}
			overrideOrder := []string{
				"FOO1",
				"HOST_IP",
				"BAR1",
				"HUMAN_FRIENDLY_HOST_NAME",
			}
			wantEnvMap := envMap{
				"$$HOST_IP$$":                  "10.76.77.178",
				"$$HOST_NAME$$":                "fakehost",
				"$$HUMAN_FRIENDLY_HOST_NAME$$": "FakeHost2",
				"$$FOO1$$":                     "foo1",
				"$$BAR1$$":                     "bar1",
			}
			wantKeyOrder := []string{
				"$$HOST_IP$$",
				"$$HOST_NAME$$",
				"$$HUMAN_FRIENDLY_HOST_NAME$$",
				"$$FOO1$$",
				"$$BAR1$$",
			}

			got := env.override(ctx, override, overrideOrder)
			if !testCmpDiff(t, "newConfigEnv()", tc, "configEnv struct", wantEnvMap, got.env) {
				return
			}
			if !testCmpDiff(t, "newConfigEnv()", tc, "configEnv struct env key order", wantKeyOrder, got.envKeyOrder) {
				return
			}
		},
	},
}

func TestConfigEnv(t *testing.T) {
	for _, tc := range configEnvTests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newTestContext(tc.ctxInfo)
			env := newConfigEnv(ctx)
			tc.test(t, ctx, tc.name, env)
		})
	}
}
