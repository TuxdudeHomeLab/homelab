package env

import (
	"context"

	"github.com/tuxdudehomelab/homelab/internal/host"
	"github.com/tuxdudehomelab/homelab/internal/user"
)

var (
	configEnvHostIP                = "HOST_IP"
	configEnvHostName              = "HOST_NAME"
	configEnvHumanFriendlyHostName = "HUMAN_FRIENDLY_HOST_NAME"
	configEnvUserName              = "USER_NAME"
	configEnvUserID                = "USER_ID"
	configEnvUserPrimaryGroupName  = "USER_PRIMARY_GROUP_NAME"
	configEnvUserPrimaryGroupID    = "USER_PRIMARY_GROUP_ID"
)

type ConfigEnvManager struct {
	env *configEnv
}

type EnvMap map[string]string
type EnvOrder []string

func NewSystemConfigEnvManager(ctx context.Context) *ConfigEnvManager {
	envMap, envOrder := defaultEnv(ctx)
	return &ConfigEnvManager{
		env: newConfigEnv(ctx, envMap, envOrder),
	}
}

func (c *ConfigEnvManager) NewGlobalConfigEnvManager(ctx context.Context, env EnvMap, order EnvOrder) *ConfigEnvManager {
	// TODO: Apply env variables specific to the config that are relevant
	// globally within the config.
	return &ConfigEnvManager{
		env: c.env.override(ctx, env, order),
	}
}

func (c *ConfigEnvManager) NewContainerConfigEnvManager(ctx context.Context, env EnvMap, order EnvOrder) *ConfigEnvManager {
	// TODO: Apply env variables specific to the config that are relevant
	// within the container config.
	return &ConfigEnvManager{
		env: c.env.override(ctx, env, order),
	}
}

func (c *ConfigEnvManager) Apply(input string) string {
	return c.env.apply(input)
}

func defaultEnv(ctx context.Context) (EnvMap, EnvOrder) {
	h := host.MustHostInfo(ctx)
	u := user.MustUserInfo(ctx)
	return EnvMap{
			configEnvHostIP:                h.IP.String(),
			configEnvHostName:              h.HostName,
			configEnvHumanFriendlyHostName: h.HumanFriendlyHostName,
			configEnvUserName:              u.User.Username,
			configEnvUserID:                u.User.Uid,
			configEnvUserPrimaryGroupName:  u.PrimaryGroup.Name,
			configEnvUserPrimaryGroupID:    u.PrimaryGroup.Gid,
		}, EnvOrder{
			configEnvHostIP,
			configEnvHostName,
			configEnvHumanFriendlyHostName,
			configEnvUserName,
			configEnvUserID,
			configEnvUserPrimaryGroupName,
			configEnvUserPrimaryGroupID,
		}
}
