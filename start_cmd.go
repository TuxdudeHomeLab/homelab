package main

type startCmdHandler struct {
	dep *deployment
}

func newStartCmdHandler() *startCmdHandler {
	return &startCmdHandler{}
}

func (s *startCmdHandler) containerAndGroupFlags() bool {
	return true
}

func (s *startCmdHandler) run(options *cmdOptions) error {
	err := validateContainerAndGroupFlags(&options.containerAndGroup)
	if err != nil {
		return err
	}

	s.dep, err = buildDeployment()
	if err != nil {
		return err
	}

	res := queryContainers(s.dep, options)
	log.Debugf("start command - result containers =\n%s", res)
	for _, c := range res {
		// We ignore the errors to keep moving forward even if one or more
		// of the containers fail to start.
		_ = c.start()
	}

	return nil
}
