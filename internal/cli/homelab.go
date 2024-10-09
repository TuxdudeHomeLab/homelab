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

func buildHomelabCmd(ctx context.Context, opts *clicommon.GlobalCmdOptions) *cobra.Command {
	ver := versionInfo(ctx)
	cmd := &cobra.Command{
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
	clicommon.AddHomelabFlags(ctx, cmd, opts)
	cmd.AddGroup(
		&cobra.Group{
			ID:    clicommon.ConfigCmdGroupID,
			Title: "Configuration:",
		},
		&cobra.Group{
			ID:    clicommon.ContainersCmdGroupID,
			Title: "Containers:",
		},
	)

	return cmd
}

func initHomelabCmd(ctx context.Context) *cobra.Command {
	globalOpts := clicommon.GlobalCmdOptions{}
	homelabCmd := buildHomelabCmd(ctx, &globalOpts)
	homelabCmd.AddCommand(cmds.ConfigCmd(ctx, &globalOpts))
	homelabCmd.AddCommand(cmds.GroupCmd(ctx, &globalOpts))
	homelabCmd.AddCommand(cmds.ContainerCmd(ctx, &globalOpts))
	return homelabCmd
}

func Exec(ctx context.Context, outW, errW io.Writer, args ...string) error {
	homelab := initHomelabCmd(ctx)
	homelab.SetOut(outW)
	homelab.SetErr(errW)
	homelab.SetArgs(args)
	return homelab.Execute()
}
