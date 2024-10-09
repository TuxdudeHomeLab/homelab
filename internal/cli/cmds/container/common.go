package container

import (
	"fmt"
	"strings"
)

func validateContainerName(name string) (string, string, error) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Container name must be specified in the form 'group/container'")
	}
	return parts[0], parts[1], nil
}

func mustContainerName(name string) (string, string) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 {
		panic("Container name must be specified in the form 'group/container'")
	}
	return parts[0], parts[1]
}
