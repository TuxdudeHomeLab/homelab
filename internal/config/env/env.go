package env

import (
	"context"
	"fmt"
	"strings"

	"github.com/tuxdudehomelab/homelab/internal/host"
)

var (
	configEnvHostIP                = "HOST_IP"
	configEnvHostName              = "HOST_NAME"
	configEnvHumanFriendlyHostName = "HUMAN_FRIENDLY_HOST_NAME"
)

type ConfigEnv struct {
	env         EnvMap
	envKeyOrder []string
	replacer    *strings.Replacer
}

type EnvMap map[string]string

func NewConfigEnv(ctx context.Context) *ConfigEnv {
	h, found := host.HostInfoFromContext(ctx)
	if !found {
		log(ctx).Fatalf("Unable to find host info in context")
	}

	c := ConfigEnv{}
	return c.Override(
		ctx,
		EnvMap{
			configEnvHostIP:                h.IP.String(),
			configEnvHostName:              h.HostName,
			configEnvHumanFriendlyHostName: h.HumanFriendlyHostName,
		},
		[]string{
			configEnvHostIP,
			configEnvHostName,
			configEnvHumanFriendlyHostName,
		})
}

func (c *ConfigEnv) Override(ctx context.Context, override EnvMap, order []string) *ConfigEnv {
	if len(override) != len(order) {
		log(ctx).Fatalf("Override map (len:%d) and order slice (len:%d) are of unequal lengths", len(override), len(order))
	}
	res := ConfigEnv{
		env:         EnvMap{},
		envKeyOrder: make([]string, 0),
	}
	for _, k := range c.envKeyOrder {
		v := c.env[k]
		res.env[k] = v
		res.envKeyOrder = append(res.envKeyOrder, k)
	}
	for _, k := range order {
		newVal, found := override[k]
		if !found {
			log(ctx).Fatalf("Expected key %s not found in override map input", k)
		}
		sk := configEnvSearchKey(k)
		if _, found := res.env[sk]; !found {
			res.envKeyOrder = append(res.envKeyOrder, sk)
		}
		res.env[sk] = newVal
	}
	var repl []string
	for _, k := range res.envKeyOrder {
		repl = append(repl, k, res.env[k])
	}
	res.replacer = strings.NewReplacer(repl...)
	return &res
}

func (c *ConfigEnv) Apply(input string) string {
	return c.replacer.Replace(input)
}

func configEnvSearchKey(env string) string {
	return fmt.Sprintf("$$%s$$", env)
}
