package main

import "context"

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

	dockerClient, err := newDockerClient(s.dep.host.dockerPlatform)
	if err != nil {
		return err
	}
	defer dockerClient.close()

	res := queryContainers(s.dep, options)
	log.Debugf("start command - Starting containers: ")
	for _, c := range res {
		log.Debugf("%s", c.name())
	}
	log.DebugEmpty()

	ctx := context.Background()
	for _, c := range res {
		// We ignore the errors to keep moving forward even if one or more
		// of the containers fail to start.
		_ = c.start(ctx, dockerClient)
	}

	return nil
}
