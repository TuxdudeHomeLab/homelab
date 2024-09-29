package testinit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	repoBaseDirAsWorkDir()
}

func repoBaseDirAsWorkDir() {
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("Unable to set the repo base dir as the current directory in tests, reason: %w", err))
	}

	for len(wd) > 0 && wd != "/" && !strings.HasSuffix(wd, "/homelab") {
		wd = filepath.Dir(wd)
	}

	if wd == "/" {
		panic(fmt.Errorf("Unable to find the homelab repo base directory to set as the current directory in tests"))
	}

	if err := os.Chdir(wd); err != nil {
		panic(fmt.Errorf("Unable to change the current directory to %s in tests, reason: %w", wd, err))
	}
}
