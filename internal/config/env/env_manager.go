package env

import (
	"context"

	"github.com/tuxdudehomelab/homelab/internal/host"
)

var (
	configEnvHostIP                = "HOST_IP"
	configEnvHostName              = "HOST_NAME"
	configEnvHumanFriendlyHostName = "HUMAN_FRIENDLY_HOST_NAME"
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
	h, found := host.HostInfoFromContext(ctx)
	if !found {
		log(ctx).Fatalf("Unable to find host info in context")
	}

	return EnvMap{
			configEnvHostIP:                h.IP.String(),
			configEnvHostName:              h.HostName,
			configEnvHumanFriendlyHostName: h.HumanFriendlyHostName,
		}, EnvOrder{
			configEnvHostIP,
			configEnvHostName,
			configEnvHumanFriendlyHostName,
		}
}
