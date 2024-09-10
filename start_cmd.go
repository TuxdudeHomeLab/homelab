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

	// Identify the containers that are in scope for this command invocation.
	// Run start() against each of those containers.

	// start() for a single container involves these steps:
	// 1. Validate the container is allowed to run on the current host.
	// 2. Create the network for the container if it doesn't exist already.
	// 3. Execute any pre-start commands.
	// 4. Pull the container image.
	// 5. Purge (i.e. stop and remove) any previously existing containers
	// under the same name.
	// 6. Create the container.
	// 7. Start the container.

	res := queryContainers(s.dep, options)

	log.Infof("Result containers =\n%s", res)
	for _, c := range res {
		log.Infof("%s IsAllowed on host: %t", c, c.isAllowedOnCurrentHost())
	}

	return nil
}
