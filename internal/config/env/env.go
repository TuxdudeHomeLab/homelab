package env

import (
	"context"
	"fmt"
	"strings"
)

type configEnv struct {
	env         EnvMap
	envKeyOrder EnvOrder
	replacer    *strings.Replacer
}

func newConfigEnv(ctx context.Context, env EnvMap, order EnvOrder) *configEnv {
	c := configEnv{}
	return c.override(ctx, env, order)
}

func (c *configEnv) override(ctx context.Context, override EnvMap, order EnvOrder) *configEnv {
	if len(override) != len(order) {
		log(ctx).Fatalf("Override map (len:%d) and order slice (len:%d) are of unequal lengths", len(override), len(order))
	}
	res := configEnv{
		env:         EnvMap{},
		envKeyOrder: EnvOrder{},
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

func (c *configEnv) apply(input string) string {
	return c.replacer.Replace(input)
}

func configEnvSearchKey(env string) string {
	return fmt.Sprintf("$$%s$$", env)
}
