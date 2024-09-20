package main

import "fmt"

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

func newConfigEnv(host *hostInfo) *configEnv {
	c := configEnv{
		env: envMap{
			configEnvSearchKey(configEnvHostIP):                host.ip.String(),
			configEnvSearchKey(configEnvHostName):              host.hostName,
			configEnvSearchKey(configEnvHumanFriendlyHostName): host.humanFriendlyHostName,
		},
		envKeyOrder: []string{
			configEnvSearchKey(configEnvHostIP),
			configEnvSearchKey(configEnvHostName),
			configEnvSearchKey(configEnvHumanFriendlyHostName),
		},
	}
	return &c
}

// nolint: unused
func (c *configEnv) override(override envMap, order []string) *configEnv {
	res := configEnv{
		env:         envMap{},
		envKeyOrder: make([]string, 0),
	}
	for _, k := range c.envKeyOrder {
		res.env[k] = c.env[k]
		res.envKeyOrder = append(res.envKeyOrder, k)
	}
	for _, k := range order {
		newVal, ok := override[k]
		if !ok {
			log.Fatalf("Expected key %s not found in override map input", k)
		}
		sk := configEnvSearchKey(k)
		if _, ok := res.env[sk]; !ok {
			res.envKeyOrder = append(res.envKeyOrder, k)
		}
		res.env[sk] = newVal
	}
	return &res
}

func configEnvSearchKey(env string) string {
	return fmt.Sprintf("$$%s$$", env)
}
