package cmdexec

import (
	"errors"
	"fmt"
	"os/exec"
)

type executor struct{}

func (e *executor) Run(bin string, args ...string) (string, error) {
	cmd := exec.Command(bin, args...)
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return "", fmt.Errorf("command failed %s %q, reason: %w, stderr: %s", bin, args, err, ee.Stderr)
		}
		return "", fmt.Errorf("command failed %s %q, reason: %w", bin, args, err)
	}
	return string(out), nil
}
