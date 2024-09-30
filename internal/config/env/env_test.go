package env

import (
	"bytes"
	"context"
	"testing"

	"github.com/tuxdude/zzzlog"
	logger "github.com/tuxdudehomelab/homelab/internal/log"
	"github.com/tuxdudehomelab/homelab/internal/testhelpers"
	"github.com/tuxdudehomelab/homelab/internal/testutils"
)

var configEnvTests = []struct {
	name    string
	ctxInfo *testutils.TestContextInfo
	test    func(*testing.T, context.Context, string)
}{
	{
		name:    "Config Env - New",
		ctxInfo: &testutils.TestContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			wantEnvMap := EnvMap{
				"$$HOST_IP$$":                  "10.76.77.78",
				"$$HOST_NAME$$":                "fakehost",
				"$$HUMAN_FRIENDLY_HOST_NAME$$": "FakeHost",
			}
			wantKeyOrder := []string{
				"$$HOST_IP$$",
				"$$HOST_NAME$$",
				"$$HUMAN_FRIENDLY_HOST_NAME$$",
			}

			env := NewConfigEnv(ctx)
			if !testhelpers.CmpDiff(t, "NewConfigEnv()", tc, "configEnv struct env map", wantEnvMap, env.env) {
				return
			}
			if !testhelpers.CmpDiff(t, "NewConfigEnv()", tc, "configEnv struct env key order", wantKeyOrder, env.envKeyOrder) {
				return
			}
		},
	},
	{
		name:    "Config Env - Override - No overlap",
		ctxInfo: &testutils.TestContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			override := EnvMap{
				"FOO1": "foo1",
				"BAR1": "bar1",
				"BAZ1": "baz1",
			}
			overrideOrder := []string{
				"FOO1",
				"BAR1",
				"BAZ1",
			}
			wantEnvMap := EnvMap{
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

			env := NewConfigEnv(ctx)
			got := env.Override(ctx, override, overrideOrder)
			if !testhelpers.CmpDiff(t, "NewConfigEnv()", tc, "configEnv struct", wantEnvMap, got.env) {
				return
			}
			if !testhelpers.CmpDiff(t, "NewConfigEnv()", tc, "configEnv struct env key order", wantKeyOrder, got.envKeyOrder) {
				return
			}
		},
	},
	{
		name:    "Config Env - Override - With overlap",
		ctxInfo: &testutils.TestContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			override := EnvMap{
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
			wantEnvMap := EnvMap{
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

			env := NewConfigEnv(ctx)
			got := env.Override(ctx, override, overrideOrder)
			if !testhelpers.CmpDiff(t, "NewConfigEnv()", tc, "configEnv struct", wantEnvMap, got.env) {
				return
			}
			if !testhelpers.CmpDiff(t, "NewConfigEnv()", tc, "configEnv struct env key order", wantKeyOrder, got.envKeyOrder) {
				return
			}
		},
	},
	{
		name:    "Config Env - Apply - No Match",
		ctxInfo: &testutils.TestContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			input := "host_ip=$HOST_IP$$;host_name=$$HOST_NAME$;friendly=HUMAN_FRIENDLY_HOST_NAME"
			want := "host_ip=$HOST_IP$$;host_name=$$HOST_NAME$;friendly=HUMAN_FRIENDLY_HOST_NAME"

			env := NewConfigEnv(ctx)
			got := env.Apply(input)
			if !testhelpers.CmpDiff(t, "configEnv.apply()", tc, "apply result", want, got) {
				return
			}
		},
	},
	{
		name:    "Config Env - Apply - Single Match",
		ctxInfo: &testutils.TestContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			input := "host_ip=$HOST_IP$$;host_name=$$HOST_NAME$$;friendly=HUMAN_FRIENDLY_HOST_NAME"
			want := "host_ip=$HOST_IP$$;host_name=fakehost;friendly=HUMAN_FRIENDLY_HOST_NAME"

			env := NewConfigEnv(ctx)
			got := env.Apply(input)
			if !testhelpers.CmpDiff(t, "configEnv.apply()", tc, "apply result", want, got) {
				return
			}
		},
	},
	{
		name:    "Config Env - Apply - Multi Match",
		ctxInfo: &testutils.TestContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			input := "host_ip=$$HOST_IP$$;host_name=$$HOST_NAME$$;friendly=$$HUMAN_FRIENDLY_HOST_NAME$$"
			want := "host_ip=10.76.77.78;host_name=fakehost;friendly=FakeHost"

			env := NewConfigEnv(ctx)
			got := env.Apply(input)
			if !testhelpers.CmpDiff(t, "configEnv.apply()", tc, "apply result", want, got) {
				return
			}
		},
	},
	{
		name:    "Config Env - Apply - Recursive Match",
		ctxInfo: &testutils.TestContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			input := "foo=$$FOO$$"
			want := "foo=$$BAR$$"

			env := NewConfigEnv(ctx).Override(
				ctx,
				EnvMap{
					"FOO": "$$BAR$$",
					"BAR": "$$BAZ$$",
					"BAZ": "baz1",
				},
				[]string{
					"FOO",
					"BAR",
					"BAZ",
				},
			)
			got := env.Apply(input)
			if !testhelpers.CmpDiff(t, "configEnv.apply()", tc, "apply result", want, got) {
				return
			}
		},
	},
	{
		name:    "Config Env - New - Missing Host Info",
		ctxInfo: &testutils.TestContextInfo{},
		test: func(t *testing.T, _ context.Context, tc string) {
			want := `Unable to find host info in context`
			l := testutils.NewCapturingTestLogger(zzzlog.LvlInfo, new(bytes.Buffer))
			ctx := logger.WithLogger(context.Background(), l)

			defer testhelpers.ExpectPanic(t, "NewConfigEnv", tc, want)
			_ = NewConfigEnv(ctx)
		},
	},
	{
		name:    "Config Env - Override - Unequal Lengths Between Override Map And Order",
		ctxInfo: &testutils.TestContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			want := `Override map \(len:2\) and order slice \(len:1\) are of unequal lengths`
			override := EnvMap{
				"FOO1": "foo1",
				"BAR1": "bar1",
			}
			overrideOrder := []string{
				"FOO1",
			}
			l := testutils.NewCapturingTestLogger(zzzlog.LvlInfo, new(bytes.Buffer))
			ctx = logger.WithLogger(ctx, l)
			env := NewConfigEnv(ctx)

			defer testhelpers.ExpectPanic(t, "configEnv.override()", tc, want)
			_ = env.Override(ctx, override, overrideOrder)
		},
	},
	{
		name:    "Config Env - Override - Invalid Override Map Key In Order",
		ctxInfo: &testutils.TestContextInfo{},
		test: func(t *testing.T, ctx context.Context, tc string) {
			want := `Expected key BAZ1 not found in override map input`
			override := EnvMap{
				"FOO1": "foo1",
				"BAR1": "bar1",
			}
			overrideOrder := []string{
				"FOO1",
				"BAZ1",
			}
			l := testutils.NewCapturingTestLogger(zzzlog.LvlInfo, new(bytes.Buffer))
			ctx = logger.WithLogger(ctx, l)
			env := NewConfigEnv(ctx)

			defer testhelpers.ExpectPanic(t, "configEnv.override()", tc, want)
			_ = env.Override(ctx, override, overrideOrder)
		},
	},
}

func TestConfigEnv(t *testing.T) {
	for _, tc := range configEnvTests {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t, testutils.NewTestContext(tc.ctxInfo), tc.name)
		})
	}
}
