package docker

import (
	"fmt"

	dcontainer "github.com/docker/docker/api/types/container"
)

func RestartPolicyModeFromString(pol string) (dcontainer.RestartPolicyMode, error) {
	switch pol {
	case "", "no":
		return dcontainer.RestartPolicyDisabled, nil
	case "always":
		return dcontainer.RestartPolicyAlways, nil
	case "on-failure":
		return dcontainer.RestartPolicyOnFailure, nil
	case "unless-stopped":
		return dcontainer.RestartPolicyUnlessStopped, nil
	default:
		return "", fmt.Errorf("invalid restart policy mode string: %s", pol)
	}
}

func RestartPolicyModeValidValues() string {
	return "[ 'no', 'always', 'on-failure', 'unless-stopped' ]"
}
