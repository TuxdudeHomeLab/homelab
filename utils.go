package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	logLevelEnvVar   = "HOMELAB_LOG_LEVEL"
	logLevelEnvDebug = "debug"
	logLevelEnvTrace = "trace"
)

// Returns the JSON formatted string representation of the specified object.
func prettyPrintJSON(x interface{}) string {
	p, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		return fmt.Sprintf("%#v", x)
	}
	return string(p)
}

func stringifyMap[K comparable, V any](m map[K]V) string {
	var sb strings.Builder
	sb.WriteString("[")
	first := true
	for _, v := range m {
		if first {
			sb.WriteString(fmt.Sprintf("%v", v))
			first = false
		} else {
			sb.WriteString(fmt.Sprintf(", %v", v))
		}
	}
	if first {
		sb.WriteString("empty")
	}
	sb.WriteString("]")
	return sb.String()
}

func logToErrorAndReturn(format string, args ...interface{}) error {
	log.Errorf(format, args...)
	log.ErrorEmpty()
	return fmt.Errorf(format, args...)
}

// TODO: Remove this after this function is used.
// nolint (unused)
func logToWarnAndReturn(format string, args ...interface{}) error {
	log.Warnf(format, args...)
	log.WarnEmpty()
	return fmt.Errorf(format, args...)
}

func isLogLevelDebug() bool {
	return isEnvValue(logLevelEnvVar, logLevelEnvDebug)
}

func isLogLevelTrace() bool {
	return isEnvValue(logLevelEnvVar, logLevelEnvTrace)
}

func isEnvValue(envVar string, envValue string) bool {
	val, isVarSet := os.LookupEnv(envVar)
	return isVarSet && val == envValue
}
