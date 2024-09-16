package main

func homelabConfigsPath(cliConfigFlag string, configsDirFlag string) (string, error) {
	// Use the flag from the command line if present.
	if len(configsDirFlag) > 0 {
		log.Debugf("Using Homelab configs path from command line flag: %s", configsDirFlag)
		return configsDirFlag, nil
	}
	path, err := configsPath(cliConfigFlag)
	if err != nil {
		return "", err
	}

	log.Debugf("Using Homelab configs path from CLI config: %s", path)
	return path, nil
}
