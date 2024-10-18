package utils

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/docker/go-units"
	"gopkg.in/yaml.v3"
)

type StringSet map[string]struct{}

func MustParseDuration(d string) time.Duration {
	res, err := time.ParseDuration(d)
	if err != nil {
		panic(err.Error())
	}
	return res
}

func MustParseRAMInBytes(size string) int64 {
	res, err := units.RAMInBytes(size)
	if err != nil {
		panic(err.Error())
	}
	return res
}

// Returns the YAML string representation of the specified object.
func PrettyPrintYAML(x interface{}) string {
	res := bytes.Buffer{}
	enc := yaml.NewEncoder(&res)
	enc.SetIndent(2)
	err := enc.Encode(x)
	if err != nil {
		return fmt.Sprintf("%#v", x)
	}
	return res.String()
}

func LogToErrorAndReturn(ctx context.Context, format string, args ...interface{}) error {
	log(ctx).Errorf(format, args...)
	log(ctx).ErrorEmpty()
	return fmt.Errorf(format, args...)
}
