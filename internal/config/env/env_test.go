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

//nolint:thelper // The struct includes test (and not helper) code for each test case.
var configEnvTests = []struct {
	name string
	test func(*testing.T, context.Context, *configEnv, string)
}{
	{
		name: "Config Env - New",
		test: func(t *testing.T, ctx context.Context, env *configEnv, tc string) {
			wantEnvMap := EnvMap{
				"$$ENV1$$": "my-env-1",
				"$$ENV2$$": "my-env-2",
				"$$ENV3$$": "my-env-3",
			}
			wantKeyOrder := EnvOrder{
				"$$ENV1$$",
				"$$ENV2$$",
				"$$ENV3$$",
			}

			if !testhelpers.CmpDiff(t, "newConfigEnv()", tc, "configEnv struct env map", wantEnvMap, env.env) {
				return
			}
			if !testhelpers.CmpDiff(t, "newConfigEnv()", tc, "configEnv struct env key order", wantKeyOrder, env.envKeyOrder) {
				return
			}
		},
	},
	{
		name: "Config Env - override - No overlap",
		test: func(t *testing.T, ctx context.Context, env *configEnv, tc string) {
			override := EnvMap{
				"FOO1": "foo1",
				"BAR1": "bar1",
				"BAZ1": "baz1",
			}
			overrideOrder := EnvOrder{
				"FOO1",
				"BAR1",
				"BAZ1",
			}
			wantEnvMap := EnvMap{
				"$$ENV1$$": "my-env-1",
				"$$ENV2$$": "my-env-2",
				"$$ENV3$$": "my-env-3",
				"$$FOO1$$": "foo1",
				"$$BAR1$$": "bar1",
				"$$BAZ1$$": "baz1",
			}
			wantKeyOrder := EnvOrder{
				"$$ENV1$$",
				"$$ENV2$$",
				"$$ENV3$$",
				"$$FOO1$$",
				"$$BAR1$$",
				"$$BAZ1$$",
			}

			got := env.override(ctx, override, overrideOrder)
			if !testhelpers.CmpDiff(t, "newConfigEnv()", tc, "configEnv struct", wantEnvMap, got.env) {
				return
			}
			if !testhelpers.CmpDiff(t, "newConfigEnv()", tc, "configEnv struct env key order", wantKeyOrder, got.envKeyOrder) {
				return
			}
		},
	},
	{
		name: "Config Env - override - With overlap",
		test: func(t *testing.T, ctx context.Context, env *configEnv, tc string) {
			override := EnvMap{
				"FOO1": "foo1",
				"ENV1": "env1",
				"BAR1": "bar1",
				"ENV3": "env3",
			}
			overrideOrder := EnvOrder{
				"FOO1",
				"ENV1",
				"BAR1",
				"ENV3",
			}
			wantEnvMap := EnvMap{
				"$$ENV1$$": "env1",
				"$$ENV2$$": "my-env-2",
				"$$ENV3$$": "env3",
				"$$FOO1$$": "foo1",
				"$$BAR1$$": "bar1",
			}
			wantKeyOrder := EnvOrder{
				"$$ENV1$$",
				"$$ENV2$$",
				"$$ENV3$$",
				"$$FOO1$$",
				"$$BAR1$$",
			}

			got := env.override(ctx, override, overrideOrder)
			if !testhelpers.CmpDiff(t, "newConfigEnv()", tc, "configEnv struct", wantEnvMap, got.env) {
				return
			}
			if !testhelpers.CmpDiff(t, "newConfigEnv()", tc, "configEnv struct env key order", wantKeyOrder, got.envKeyOrder) {
				return
			}
		},
	},
	{
		name: "Config Env - apply - No Match",
		test: func(t *testing.T, ctx context.Context, env *configEnv, tc string) {
			input := "my_env1=$ENV1$$;my_env2=$$ENV2$;my_env3=ENV3"
			want := "my_env1=$ENV1$$;my_env2=$$ENV2$;my_env3=ENV3"

			got := env.apply(input)
			if !testhelpers.CmpDiff(t, "configEnv.apply()", tc, "apply result", want, got) {
				return
			}
		},
	},
	{
		name: "Config Env - apply - Single Match",
		test: func(t *testing.T, ctx context.Context, env *configEnv, tc string) {
			input := "my_env1=$ENV1$$;my_env2=$$ENV2$$;my_env3=ENV3"
			want := "my_env1=$ENV1$$;my_env2=my-env-2;my_env3=ENV3"

			got := env.apply(input)
			if !testhelpers.CmpDiff(t, "configEnv.apply()", tc, "apply result", want, got) {
				return
			}
		},
	},
	{
		name: "Config Env - apply - Multi Match",
		test: func(t *testing.T, ctx context.Context, env *configEnv, tc string) {
			input := "my_env1=$$ENV1$$;my_env2=$$ENV2$$;my_env3=$$ENV3$$"
			want := "my_env1=my-env-1;my_env2=my-env-2;my_env3=my-env-3"

			got := env.apply(input)
			if !testhelpers.CmpDiff(t, "configEnv.apply()", tc, "apply result", want, got) {
				return
			}
		},
	},
	{
		name: "Config Env - apply - Recursive Match",
		test: func(t *testing.T, ctx context.Context, env *configEnv, tc string) {
			input := "foo=$$FOO$$"
			want := "foo=$$BAR$$"

			env = env.override(
				ctx,
				EnvMap{
					"FOO": "$$BAR$$",
					"BAR": "$$BAZ$$",
					"BAZ": "baz1",
				},
				EnvOrder{
					"FOO",
					"BAR",
					"BAZ",
				},
			)
			got := env.apply(input)
			if !testhelpers.CmpDiff(t, "configEnv.apply()", tc, "apply result", want, got) {
				return
			}
		},
	},

	{
		name: "Config Env - override - Unequal Lengths Between Override Map And Order",
		test: func(t *testing.T, ctx context.Context, env *configEnv, tc string) {
			want := `Override map \(len:2\) and order slice \(len:1\) are of unequal lengths`
			override := EnvMap{
				"FOO1": "foo1",
				"BAR1": "bar1",
			}
			overrideOrder := EnvOrder{
				"FOO1",
			}

			defer testhelpers.ExpectPanic(t, "configEnv.override()", tc, want)
			_ = env.override(ctx, override, overrideOrder)
		},
	},
	{
		name: "Config Env - override - Invalid Override Map Key In Order",
		test: func(t *testing.T, ctx context.Context, env *configEnv, tc string) {
			want := `Expected key BAZ1 not found in override map input`
			override := EnvMap{
				"FOO1": "foo1",
				"BAR1": "bar1",
			}
			overrideOrder := EnvOrder{
				"FOO1",
				"BAZ1",
			}

			defer testhelpers.ExpectPanic(t, "configEnv.override()", tc, want)
			_ = env.override(ctx, override, overrideOrder)
		},
	},
}

func TestConfigEnv(t *testing.T) {
	t.Parallel()

	initEnvMap := EnvMap{
		"ENV1": "my-env-1",
		"ENV2": "my-env-2",
		"ENV3": "my-env-3",
	}
	initEnvOrder := EnvOrder{
		"ENV1",
		"ENV2",
		"ENV3",
	}
	for _, tc := range configEnvTests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := testutils.NewCapturingTestLogger(zzzlog.LvlInfo, new(bytes.Buffer))
			ctx := testutils.NewTestContext(&testutils.TestContextInfo{})
			ctx = logger.WithLogger(ctx, l)

			env := newConfigEnv(ctx, initEnvMap, initEnvOrder)
			tc.test(t, ctx, env, tc.name)
		})
	}
}

var configEnvInitPanicTests = []struct {
	name         string
	initEnvMap   EnvMap
	initEnvOrder EnvOrder
	want         string
	test         func(*testing.T, context.Context, *configEnv, string)
}{
	{
		name: "Config Env - New - Unequal Lengths Between Override Map And Order",
		initEnvMap: EnvMap{
			"ENV1": "my-env-1",
			"ENV2": "my-env-2",
			"ENV3": "my-env-3",
		},
		initEnvOrder: EnvOrder{
			"ENV1",
			"ENV2",
		},
		want: `Override map \(len:3\) and order slice \(len:2\) are of unequal lengths`,
	},
	{
		name: "Config Env - new - Invalid Map Key In Order",
		initEnvMap: EnvMap{
			"ENV1": "my-env-1",
			"ENV2": "my-env-2",
			"ENV3": "my-env-3",
		},
		initEnvOrder: EnvOrder{
			"ENV1",
			"ENV4",
			"ENV3",
		},
		want: `Expected key ENV4 not found in override map input`,
	},
}

func TestConfigEnvInitPanic(t *testing.T) {
	t.Parallel()

	for _, tc := range configEnvInitPanicTests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := testutils.NewCapturingTestLogger(zzzlog.LvlInfo, new(bytes.Buffer))
			ctx := logger.WithLogger(context.Background(), l)

			defer testhelpers.ExpectPanic(t, "newConfigEnv", tc.name, tc.want)
			_ = newConfigEnv(ctx, tc.initEnvMap, tc.initEnvOrder)
		})
	}
}
