package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/tuxdudehomelab/homelab/internal/cli/clicommon"
	"github.com/tuxdudehomelab/homelab/internal/cli/cmds"
	"github.com/tuxdudehomelab/homelab/internal/cli/version"
)

const (
	homelabCmdStr = "homelab"

	cliConfigFlagStr  = "cli-config"
	configsDirFlagStr = "configs-dir"

	defaultPkgVersion   = "unset"
	defaultPkgCommit    = "unset"
	defaultPkgTimestamp = "unset"
)

func versionInfo(ctx context.Context) *version.VersionInfo {
	ver, ok := version.VersionInfoFromContext(ctx)
	if ok {
		return ver
	}
	return version.NewVersionInfo(defaultPkgVersion, defaultPkgCommit, defaultPkgTimestamp)
}

func buildHomelabCmd(ctx context.Context, opt *clicommon.GlobalCmdOptions) *cobra.Command {
	ver := versionInfo(ctx)
	h := &cobra.Command{
		Use:           homelabCmdStr,
		Version:       fmt.Sprintf("%s [Revision: %s @ %s]", ver.PackageVersion, ver.PackageCommit, ver.PackageTimestamp),
		SilenceUsage:  false,
		SilenceErrors: false,
		Short:         "Homelab is a CLI for managing configuration and deployment of docket containers.",
		Long: `A CLI for managing both the configuration and deployment of groups of docker containers on a given host.

The configuration is managed using a yaml file. The configuration specifies the container groups and individual containers, their properties and how to deploy them.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: print ascii art.
			return fmt.Errorf("homelab sub-command is required")
		},
	}

	h.PersistentFlags().StringVar(
		&opt.CLIConfig, cliConfigFlagStr, "", "The path to the Homelab CLI config")
	if h.MarkPersistentFlagFilename(cliConfigFlagStr) != nil {
		log(ctx).Fatalf("failed to mark --%s flag as filename flag", cliConfigFlagStr)
	}
	h.PersistentFlags().StringVar(
		&opt.ConfigsDir, configsDirFlagStr, "", "The path to the directory containing the homelab configs")
	if h.MarkPersistentFlagDirname(configsDirFlagStr) != nil {
		log(ctx).Fatalf("failed to mark --%s flag as dirname flag", configsDirFlagStr)
	}
	h.MarkFlagsMutuallyExclusive(cliConfigFlagStr, configsDirFlagStr)

	h.AddGroup(
		&cobra.Group{
			ID:    clicommon.ConfigCmdGroupID,
			Title: "Configuration:",
		},
		&cobra.Group{
			ID:    clicommon.ContainersCmdGroupID,
			Title: "Containers:",
		},
	)

	return h
}

func initHomelabCmd(ctx context.Context) *cobra.Command {
	globalOpts := clicommon.GlobalCmdOptions{}
	homelabCmd := buildHomelabCmd(ctx, &globalOpts)
	homelabCmd.AddCommand(cmds.ShowConfigCmd(ctx, &globalOpts))
	homelabCmd.AddCommand(cmds.StartCmd(ctx, &globalOpts))
	homelabCmd.AddCommand(cmds.StopCmd(ctx, &globalOpts))
	return homelabCmd
}

func Exec(ctx context.Context, outW, errW io.Writer, args ...string) error {
	homelab := initHomelabCmd(ctx)
	homelab.SetOut(outW)
	homelab.SetErr(errW)
	homelab.SetArgs(args)
	return homelab.Execute()
}
