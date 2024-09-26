package main

import "os"

type testOSEnvMap map[string]string

func setTestEnv(envs testOSEnvMap) {
	for k, v := range envs {
		err := os.Setenv(k, v)
		if err != nil {
			panic(err)
		}
	}
}

func clearTestEnv(envs testOSEnvMap) {
	for k := range envs {
		err := os.Unsetenv(k)
		if err != nil {
			panic(err)
		}
	}
}
