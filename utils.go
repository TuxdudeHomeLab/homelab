package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Returns the JSON formatted string representation of the specified object.
func prettyPrintJSON(x interface{}) string {
	p, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		return fmt.Sprintf("%#v", x)
	}
	return string(p)
}

func stringifyMap[V any](m map[string]V) string {
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

func logToWarnAndReturn(format string, args ...interface{}) error {
	log.Warnf(format, args...)
	log.WarnEmpty()
	return fmt.Errorf(format, args...)
}
