package clicommon

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/config"
)

func AutoCompleteGroups(ctx context.Context, args []string, cmd string, options *GlobalCmdOptions) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}
	groups, err := groupsOnly(ctx, cmd, options)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return groups, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
}

func AutoCompleteContainers(ctx context.Context, args []string, cmd string, options *GlobalCmdOptions) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}
	containers, err := containersOnly(ctx, cmd, options)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return containers, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
}

func groupsOnly(ctx context.Context, cmd string, options *GlobalCmdOptions) ([]string, error) {
	h, err := buildHomelabGroupsOnly(ctx, cmd, options)
	if err != nil {
		return nil, err
	}
	return h.ListGroups(), nil
}

func containersOnly(ctx context.Context, cmd string, options *GlobalCmdOptions) ([]string, error) {
	h, err := buildHomelabContainersOnly(ctx, cmd, options)
	if err != nil {
		return nil, err
	}
	return h.ListContainers(), nil
}

func buildHomelabGroupsOnly(ctx context.Context, cmd string, options *GlobalCmdOptions) (*config.HomelabGroupsOnly, error) {
	path, err := configsPath(ctx, cmd, options)
	if err != nil {
		return nil, err
	}

	r, err := config.MergedConfigsReader(ctx, path)
	if err != nil {
		return nil, err
	}

	conf := config.HomelabGroupsOnly{}
	err = conf.Parse(ctx, r)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func buildHomelabContainersOnly(ctx context.Context, cmd string, options *GlobalCmdOptions) (*config.HomelabContainersOnly, error) {
	path, err := configsPath(ctx, cmd, options)
	if err != nil {
		return nil, err
	}

	r, err := config.MergedConfigsReader(ctx, path)
	if err != nil {
		return nil, err
	}

	conf := config.HomelabContainersOnly{}
	err = conf.Parse(ctx, r)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}
