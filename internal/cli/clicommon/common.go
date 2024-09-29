package clicommon

type GlobalCmdOptions struct {
	CLIConfig  string
	ConfigsDir string
}

const (
	AllGroupsFlagStr = "all-groups"
	GroupFlagStr     = "group"
	ContainerFlagStr = "container"

	ConfigCmdGroupID     = "config"
	ContainersCmdGroupID = "containers"
)
