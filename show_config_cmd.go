package main

import (
	"flag"
)

type showConfigCmdHandler struct{}

func newShowConfigCmdHandler() *showConfigCmdHandler {
	return &showConfigCmdHandler{}
}

func (s *showConfigCmdHandler) updateFlagSet(fs *flag.FlagSet) {
}

func (s *showConfigCmdHandler) run() error {
	config, err := parseHomelabConfig()
	if err != nil {
		return err
	}

	log.Infof("Homelab config:\n%s", prettyPrintJSON(config))
	return nil
}
