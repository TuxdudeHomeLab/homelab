package main

import (
	"context"
	"fmt"
)

var (
	configEnvHostIP                = "HOST_IP"
	configEnvHostName              = "HOST_NAME"
	configEnvHumanFriendlyHostName = "HUMAN_FRIENDLY_HOST_NAME"
)

type configEnv struct {
	env         envMap
	envKeyOrder []string
}

type envMap map[string]string

func newConfigEnv(ctx context.Context) *configEnv {
	host, found := hostInfoFromContext(ctx)
	if !found {
		log(ctx).Fatalf("Unable to find host info in context")
	}

	c := configEnv{}
	return c.override(
		ctx,
		envMap{
			configEnvHostIP:                host.ip.String(),
			configEnvHostName:              host.hostName,
			configEnvHumanFriendlyHostName: host.humanFriendlyHostName,
		},
		[]string{
			configEnvHostIP,
			configEnvHostName,
			configEnvHumanFriendlyHostName,
		})
}

func (c *configEnv) override(ctx context.Context, override envMap, order []string) *configEnv {
	if len(override) != len(order) {
		log(ctx).Fatalf("Override map (len:%d) and order slice (len:%d) are of unequal lengths", len(override), len(order))
	}
	res := configEnv{
		env:         envMap{},
		envKeyOrder: make([]string, 0),
	}
	for _, k := range c.envKeyOrder {
		res.env[k] = c.env[k]
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
	return &res
}

func configEnvSearchKey(env string) string {
	return fmt.Sprintf("$$%s$$", env)
}
