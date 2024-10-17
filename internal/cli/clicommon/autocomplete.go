package clicommon

import (
	"context"
	"slices"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/config"
)

func AutoCompleteGroups(ctx context.Context, args []string, cmd string, opts *GlobalCmdOptions) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}
	groups, err := groupsOnly(ctx, cmd, opts)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return groups, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
}

func AutoCompleteContainers(ctx context.Context, args []string, cmd string, opts *GlobalCmdOptions) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}
	containers, err := containersOnly(ctx, cmd, opts)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return containers, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
}

func AutoCompleteNetworks(ctx context.Context, args []string, cmd string, opts *GlobalCmdOptions) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}
	networks, err := networksOnly(ctx, cmd, opts)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return networks, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
}

func groupsOnly(ctx context.Context, cmd string, opts *GlobalCmdOptions) ([]string, error) {
	h, err := buildHomelabGroupsOnly(ctx, cmd, opts)
	if err != nil {
		return nil, err
	}
	groups := h.ListGroups()
	if slices.Index(groups, AllGroups) == -1 {
		groups = append(groups, AllGroups)
	}
	slices.Sort(groups)
	return groups, nil
}

func containersOnly(ctx context.Context, cmd string, opts *GlobalCmdOptions) ([]string, error) {
	h, err := buildHomelabContainersOnly(ctx, cmd, opts)
	if err != nil {
		return nil, err
	}
	return h.ListContainers(), nil
}

func networksOnly(ctx context.Context, cmd string, opts *GlobalCmdOptions) ([]string, error) {
	h, err := buildHomelabNetworksOnly(ctx, cmd, opts)
	if err != nil {
		return nil, err
	}
	networks := h.ListNetworks()
	if slices.Index(networks, AllNetworks) == -1 {
		networks = append(networks, AllNetworks)
	}
	slices.Sort(networks)
	return networks, nil
}

func buildHomelabGroupsOnly(ctx context.Context, cmd string, opts *GlobalCmdOptions) (*config.HomelabGroupsOnly, error) {
	path, err := configsPath(ctx, cmd, opts)
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

func buildHomelabContainersOnly(ctx context.Context, cmd string, opts *GlobalCmdOptions) (*config.HomelabContainersOnly, error) {
	path, err := configsPath(ctx, cmd, opts)
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

func buildHomelabNetworksOnly(ctx context.Context, cmd string, opts *GlobalCmdOptions) (*config.HomelabNetworksOnly, error) {
	path, err := configsPath(ctx, cmd, opts)
	if err != nil {
		return nil, err
	}

	r, err := config.MergedConfigsReader(ctx, path)
	if err != nil {
		return nil, err
	}

	conf := config.HomelabNetworksOnly{}
	err = conf.Parse(ctx, r)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}
