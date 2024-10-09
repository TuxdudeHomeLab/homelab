package env

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/tuxdude/zzzlog"
	logger "github.com/tuxdudehomelab/homelab/internal/log"
	"github.com/tuxdudehomelab/homelab/internal/testhelpers"
	"github.com/tuxdudehomelab/homelab/internal/testutils"
)

var systemConfigEnvManagerApplyTests = []struct {
	name  string
	input string
	want  string
}{
	{
		name:  "System Config Env Manager - Apply - HOST_IP",
		input: "foo-$$HOST_IP$$-bar",
		want:  "foo-10.76.77.78-bar",
	},
	{
		name:  "System Config Env Manager - Apply - HOST_NAME",
		input: "$$HOST_NAME$$",
		want:  "fakehost",
	},
	{
		name:  "System Config Env Manager - Apply - HUMAN_FRIENDLY_HOST_NAME",
		input: "$$HUMAN_FRIENDLY_HOST_NAME$$-foobar",
		want:  "FakeHost-foobar",
	},
	{
		name:  "System Config Env Manager - Apply - USER_NAME",
		input: "foobar-$$USER_NAME$$",
		want:  "foobar-fakeuser",
	},
	{
		name:  "System Config Env Manager - Apply - USER_ID",
		input: "foo-$$USER_ID$$-bar-$$USER_ID$$",
		want:  "foo-55555-bar-55555",
	},
	{
		name:  "System Config Env Manager - Apply - USER_PRIMARY_GROUP_NAME",
		input: "foobar-$$USER_PRIMARY_GROUP_NAME$$-foobar",
		want:  "foobar-fakegroup1-foobar",
	},
	{
		name:  "System Config Env Manager - Apply - USER_PRIMARY_GROUP_ID",
		input: "foo123-$$USER_PRIMARY_GROUP_ID$$-bar123",
		want:  "foo123-44444-bar123",
	},
	{
		name:  "System Config Env Manager - Apply - Multiple",
		input: "foo-$$HOST_NAME$$-$$HOST_IP$$-$$USER_NAME$$-$$USER_PRIMARY_GROUP_NAME$$",
		want:  "foo-fakehost-10.76.77.78-fakeuser-fakegroup1",
	},
}

func TestSystemConfigEnvManagerApply(t *testing.T) {
	t.Parallel()

	for _, test := range systemConfigEnvManagerApplyTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := testutils.NewCapturingTestLogger(zzzlog.LvlInfo, new(bytes.Buffer))
			ctx := testutils.NewTestContext(&testutils.TestContextInfo{})
			ctx = logger.WithLogger(ctx, l)

			env := NewSystemConfigEnvManager(ctx)
			got := env.Apply(tc.input)
			if got != tc.want {
				testhelpers.LogCustom(t, "SystemConfigEnvManager.Apply()", tc.name, fmt.Sprintf("got '%s' != want '%s'", got, tc.want))
			}
		})
	}
}

var globalConfigEnvManagerApplyTests = []struct {
	name           string
	globalEnvMap   EnvMap
	globalEnvOrder EnvOrder
	input          string
	want           string
}{
	{
		name: "Global Config Env Manager - Apply - HOMELAB_BASE_DIR",
		globalEnvMap: EnvMap{
			"MY_ENV_1": "my-env-1",
			"MY_ENV_2": "my-env-2",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"MY_ENV_2",
		},
		input: "$$HOMELAB_BASE_DIR$$/foo/bar/baz",
		want:  "/home/foobar/dummy-base-dir/foo/bar/baz",
	},
	{
		name: "Global Config Env Manager - Apply - HOST_IP",
		globalEnvMap: EnvMap{
			"MY_ENV_1": "my-env-1",
			"MY_ENV_2": "my-env-2",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"MY_ENV_2",
		},
		input: "foo-$$HOST_IP$$-bar",
		want:  "foo-10.76.77.78-bar",
	},
	{
		name: "Global Config Env Manager - Apply - HOST_NAME",
		globalEnvMap: EnvMap{
			"MY_ENV_1":  "my-env-1",
			"HOST_NAME": "my-host-name",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"HOST_NAME",
		},
		input: "$$HOST_NAME$$",
		want:  "my-host-name",
	},
	{
		name: "Global Config Env Manager - Apply - HUMAN_FRIENDLY_HOST_NAME",
		globalEnvMap: EnvMap{
			"MY_ENV_1":  "my-env-1",
			"HOST_NAME": "my-host-name",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"HOST_NAME",
		},
		input: "$$HUMAN_FRIENDLY_HOST_NAME$$-foobar",
		want:  "FakeHost-foobar",
	},
	{
		name:  "Global Config Env Manager - Apply - USER_NAME",
		input: "foobar-$$USER_NAME$$",
		want:  "foobar-fakeuser",
	},
	{
		name: "Global Config Env Manager - Apply - USER_ID",
		globalEnvMap: EnvMap{
			"USER_ID": "my-user-id",
		},
		globalEnvOrder: EnvOrder{
			"USER_ID",
		},
		input: "foo-$$USER_ID$$-bar-$$USER_ID$$",
		want:  "foo-my-user-id-bar-my-user-id",
	},
	{
		name: "Global Config Env Manager - Apply - USER_PRIMARY_GROUP_NAME",
		globalEnvMap: EnvMap{
			"MY_ENV_1": "my-env-1",
			"MY_ENV_2": "my-env-2",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"MY_ENV_2",
		},
		input: "foobar-$$USER_PRIMARY_GROUP_NAME$$-foobar",
		want:  "foobar-fakegroup1-foobar",
	},
	{
		name: "Global Config Env Manager - Apply - USER_PRIMARY_GROUP_ID",
		globalEnvMap: EnvMap{
			"MY_ENV_1": "my-env-1",
			"MY_ENV_2": "my-env-2",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"MY_ENV_2",
		},
		input: "foo123-$$USER_PRIMARY_GROUP_ID$$-bar123",
		want:  "foo123-44444-bar123",
	},
	{
		name: "Global Config Env Manager - Apply - Multiple",
		globalEnvMap: EnvMap{
			"MY_ENV_1": "my-env-1",
			"MY_ENV_2": "my-env-2",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"MY_ENV_2",
		},
		input: "$$HOMELAB_BASE_DIR$$/foo/bar/baz - $$HUMAN_FRIENDLY_HOST_NAME$$ - $$USER_ID$$ - $$USER_PRIMARY_GROUP_ID$$",
		want:  "/home/foobar/dummy-base-dir/foo/bar/baz - FakeHost - 55555 - 44444",
	},
}

func TestGlobalConfigEnvManagerApply(t *testing.T) {
	t.Parallel()

	for _, test := range globalConfigEnvManagerApplyTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := testutils.NewCapturingTestLogger(zzzlog.LvlInfo, new(bytes.Buffer))
			ctx := testutils.NewTestContext(&testutils.TestContextInfo{})
			ctx = logger.WithLogger(ctx, l)

			env := NewSystemConfigEnvManager(ctx).NewGlobalConfigEnvManager(ctx, "/home/foobar/dummy-base-dir", tc.globalEnvMap, tc.globalEnvOrder)
			got := env.Apply(tc.input)
			if got != tc.want {
				testhelpers.LogCustom(t, "GlobalConfigEnvManager.Apply()", tc.name, fmt.Sprintf("got '%s' != want '%s'", got, tc.want))
			}
		})
	}
}

var containerConfigEnvManagerApplyTests = []struct {
	name              string
	globalEnvMap      EnvMap
	globalEnvOrder    EnvOrder
	containerEnvMap   EnvMap
	containerEnvOrder EnvOrder
	input             string
	want              string
}{
	{
		name: "Container Config Env Manager - Apply - CONTAINER_GROUP_BASE_DIR",
		globalEnvMap: EnvMap{
			"MY_ENV_1": "my-env-1",
			"MY_ENV_2": "my-env-2",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"MY_ENV_2",
		},
		containerEnvMap: EnvMap{
			"MY_CT_ENV_1": "my-ct-env-1",
			"MY_CT_ENV_2": "my-ct-env-2",
		},
		containerEnvOrder: EnvOrder{
			"MY_CT_ENV_1",
			"MY_CT_ENV_2",
		},
		input: "$$CONTAINER_GROUP_BASE_DIR$$/foo/bar/baz",
		want:  "/home/foobar/dummy-base-dir/g1/foo/bar/baz",
	},
	{
		name: "Container Config Env Manager - Apply - CONTAINER_BASE_DIR",
		globalEnvMap: EnvMap{
			"MY_ENV_1": "my-env-1",
			"MY_ENV_2": "my-env-2",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"MY_ENV_2",
		},
		containerEnvMap: EnvMap{
			"MY_CT_ENV_1": "my-ct-env-1",
			"MY_CT_ENV_2": "my-ct-env-2",
		},
		containerEnvOrder: EnvOrder{
			"MY_CT_ENV_1",
			"MY_CT_ENV_2",
		},
		input: "$$CONTAINER_BASE_DIR$$/foo/bar/baz",
		want:  "/home/foobar/dummy-base-dir/g1/c1/foo/bar/baz",
	},
	{
		name: "Container Config Env Manager - Apply - CONTAINER_CONFIGS_DIR",
		globalEnvMap: EnvMap{
			"MY_ENV_1": "my-env-1",
			"MY_ENV_2": "my-env-2",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"MY_ENV_2",
		},
		containerEnvMap: EnvMap{
			"MY_CT_ENV_1": "my-ct-env-1",
			"MY_CT_ENV_2": "my-ct-env-2",
		},
		containerEnvOrder: EnvOrder{
			"MY_CT_ENV_1",
			"MY_CT_ENV_2",
		},
		input: "$$CONTAINER_CONFIGS_DIR$$/foo/bar/baz",
		want:  "/home/foobar/dummy-base-dir/g1/c1/configs/foo/bar/baz",
	},
	{
		name: "Container Config Env Manager - Apply - CONTAINER_DATA_DIR",
		globalEnvMap: EnvMap{
			"MY_ENV_1": "my-env-1",
			"MY_ENV_2": "my-env-2",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"MY_ENV_2",
		},
		containerEnvMap: EnvMap{
			"MY_CT_ENV_1": "my-ct-env-1",
			"MY_CT_ENV_2": "my-ct-env-2",
		},
		containerEnvOrder: EnvOrder{
			"MY_CT_ENV_1",
			"MY_CT_ENV_2",
		},
		input: "$$CONTAINER_DATA_DIR$$/foo/bar/baz",
		want:  "/home/foobar/dummy-base-dir/g1/c1/data/foo/bar/baz",
	},
	{
		name: "Container Config Env Manager - Apply - CONTAINER_SCRIPTS_DIR",
		globalEnvMap: EnvMap{
			"MY_ENV_1": "my-env-1",
			"MY_ENV_2": "my-env-2",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"MY_ENV_2",
		},
		containerEnvMap: EnvMap{
			"MY_CT_ENV_1": "my-ct-env-1",
			"MY_CT_ENV_2": "my-ct-env-2",
		},
		containerEnvOrder: EnvOrder{
			"MY_CT_ENV_1",
			"MY_CT_ENV_2",
		},
		input: "$$CONTAINER_SCRIPTS_DIR$$/foo/bar/baz",
		want:  "/home/foobar/dummy-base-dir/g1/c1/scripts/foo/bar/baz",
	},
	{
		name: "Container Config Env Manager - Apply - HOMELAB_BASE_DIR",
		globalEnvMap: EnvMap{
			"MY_ENV_1": "my-env-1",
			"MY_ENV_2": "my-env-2",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"MY_ENV_2",
		},
		containerEnvMap: EnvMap{
			"MY_CT_ENV_1": "my-ct-env-1",
			"MY_CT_ENV_2": "my-ct-env-2",
		},
		containerEnvOrder: EnvOrder{
			"MY_CT_ENV_1",
			"MY_CT_ENV_2",
		},
		input: "$$HOMELAB_BASE_DIR$$/foo/bar/baz",
		want:  "/home/foobar/dummy-base-dir/foo/bar/baz",
	},
	{
		name: "Container Config Env Manager - Apply - HOST_IP",
		globalEnvMap: EnvMap{
			"MY_ENV_1": "my-env-1",
			"MY_ENV_2": "my-env-2",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"MY_ENV_2",
		},
		containerEnvMap: EnvMap{
			"MY_CT_ENV_1": "my-ct-env-1",
			"MY_CT_ENV_2": "my-ct-env-2",
		},
		containerEnvOrder: EnvOrder{
			"MY_CT_ENV_1",
			"MY_CT_ENV_2",
		},
		input: "foo-$$HOST_IP$$-bar",
		want:  "foo-10.76.77.78-bar",
	},
	{
		name: "Container Config Env Manager - Apply - HOST_NAME",
		globalEnvMap: EnvMap{
			"MY_ENV_1":  "my-env-1",
			"HOST_NAME": "my-host-name",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"HOST_NAME",
		},
		containerEnvMap: EnvMap{
			"MY_CT_ENV_1": "my-ct-env-1",
			"MY_CT_ENV_2": "my-ct-env-2",
		},
		containerEnvOrder: EnvOrder{
			"MY_CT_ENV_1",
			"MY_CT_ENV_2",
		},
		input: "$$HOST_NAME$$",
		want:  "my-host-name",
	},
	{
		name: "Container Config Env Manager - Apply - HUMAN_FRIENDLY_HOST_NAME",
		globalEnvMap: EnvMap{
			"MY_ENV_1":  "my-env-1",
			"HOST_NAME": "my-host-name",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"HOST_NAME",
		},
		containerEnvMap: EnvMap{
			"MY_CT_ENV_1": "my-ct-env-1",
			"MY_CT_ENV_2": "my-ct-env-2",
		},
		containerEnvOrder: EnvOrder{
			"MY_CT_ENV_1",
			"MY_CT_ENV_2",
		},
		input: "$$HUMAN_FRIENDLY_HOST_NAME$$-foobar",
		want:  "FakeHost-foobar",
	},
	{
		name:  "Container Config Env Manager - Apply - USER_NAME",
		input: "foobar-$$USER_NAME$$",
		want:  "foobar-fakeuser",
	},
	{
		name: "Container Config Env Manager - Apply - USER_ID",
		globalEnvMap: EnvMap{
			"USER_ID": "my-user-id",
		},
		globalEnvOrder: EnvOrder{
			"USER_ID",
		},
		containerEnvMap: EnvMap{
			"MY_CT_ENV_1": "my-ct-env-1",
			"MY_CT_ENV_2": "my-ct-env-2",
		},
		containerEnvOrder: EnvOrder{
			"MY_CT_ENV_1",
			"MY_CT_ENV_2",
		},
		input: "foo-$$USER_ID$$-bar-$$USER_ID$$",
		want:  "foo-my-user-id-bar-my-user-id",
	},
	{
		name: "Container Config Env Manager - Apply - USER_PRIMARY_GROUP_NAME",
		globalEnvMap: EnvMap{
			"MY_ENV_1": "my-env-1",
			"MY_ENV_2": "my-env-2",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"MY_ENV_2",
		},
		containerEnvMap: EnvMap{
			"MY_CT_ENV_1": "my-ct-env-1",
			"MY_CT_ENV_2": "my-ct-env-2",
		},
		containerEnvOrder: EnvOrder{
			"MY_CT_ENV_1",
			"MY_CT_ENV_2",
		},
		input: "foobar-$$USER_PRIMARY_GROUP_NAME$$-foobar",
		want:  "foobar-fakegroup1-foobar",
	},
	{
		name: "Container Config Env Manager - Apply - USER_PRIMARY_GROUP_ID",
		globalEnvMap: EnvMap{
			"MY_ENV_1": "my-env-1",
			"MY_ENV_2": "my-env-2",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"MY_ENV_2",
		},
		containerEnvMap: EnvMap{
			"MY_CT_ENV_1": "my-ct-env-1",
			"MY_CT_ENV_2": "my-ct-env-2",
		},
		containerEnvOrder: EnvOrder{
			"MY_CT_ENV_1",
			"MY_CT_ENV_2",
		},
		input: "foo123-$$USER_PRIMARY_GROUP_ID$$-bar123",
		want:  "foo123-44444-bar123",
	},
	{
		name: "Container Config Env Manager - Apply - Multiple",
		globalEnvMap: EnvMap{
			"MY_ENV_1": "my-env-1",
			"MY_ENV_2": "my-env-2",
		},
		globalEnvOrder: EnvOrder{
			"MY_ENV_1",
			"MY_ENV_2",
		},
		containerEnvMap: EnvMap{
			"MY_CT_ENV_1": "my-ct-env-1",
			"MY_CT_ENV_2": "my-ct-env-2",
		},
		containerEnvOrder: EnvOrder{
			"MY_CT_ENV_1",
			"MY_CT_ENV_2",
		},
		input: "$$HOMELAB_BASE_DIR$$/foo/bar/baz - $$HUMAN_FRIENDLY_HOST_NAME$$ - $$USER_ID$$ - $$USER_PRIMARY_GROUP_ID$$",
		want:  "/home/foobar/dummy-base-dir/foo/bar/baz - FakeHost - 55555 - 44444",
	},
}

func TestContainerConfigEnvManagerApply(t *testing.T) {
	t.Parallel()

	for _, test := range containerConfigEnvManagerApplyTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := testutils.NewCapturingTestLogger(zzzlog.LvlInfo, new(bytes.Buffer))
			ctx := testutils.NewTestContext(&testutils.TestContextInfo{})
			ctx = logger.WithLogger(ctx, l)

			env := NewSystemConfigEnvManager(ctx)
			env = env.NewGlobalConfigEnvManager(ctx, "/home/foobar/dummy-base-dir", tc.globalEnvMap, tc.globalEnvOrder)
			env = env.NewContainerConfigEnvManager(ctx, "/home/foobar/dummy-base-dir/g1", "/home/foobar/dummy-base-dir/g1/c1", tc.containerEnvMap, tc.containerEnvOrder)
			got := env.Apply(tc.input)
			if got != tc.want {
				testhelpers.LogCustom(t, "ContainerConfigEnvManager.Apply()", tc.name, fmt.Sprintf("got '%s' != want '%s'", got, tc.want))
			}
		})
	}
}
