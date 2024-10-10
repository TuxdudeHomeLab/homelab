package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/clarketm/json"
)

type StringSet map[string]struct{}

func MustParseDuration(d string) time.Duration {
	res, err := time.ParseDuration(d)
	if err != nil {
		panic(err)
	}
	return res
}

// Returns the JSON formatted string representation of the specified object.
func PrettyPrintJSON(x interface{}) string {
	p, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		return fmt.Sprintf("%#v", x)
	}
	return string(p)
}

func LogToErrorAndReturn(ctx context.Context, format string, args ...interface{}) error {
	log(ctx).Errorf(format, args...)
	log(ctx).ErrorEmpty()
	return fmt.Errorf(format, args...)
}
