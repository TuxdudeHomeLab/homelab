package main

import "context"

func homelabConfigsPath(ctx context.Context, cliConfigFlag string, configsDirFlag string) (string, error) {
	// Use the flag from the command line if present.
	if len(configsDirFlag) > 0 {
		log(ctx).Debugf("Using Homelab configs path from command line flag: %s", configsDirFlag)
		return configsDirFlag, nil
	}
	path, err := configsPathFromCLIConfig(ctx, cliConfigFlag)
	if err != nil {
		return "", err
	}

	log(ctx).Debugf("Using Homelab configs path from CLI config: %s", path)
	return path, nil
}
