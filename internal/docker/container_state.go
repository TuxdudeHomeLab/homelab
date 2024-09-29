package docker

const (
	ContainerStateUnknown ContainerState = iota
	ContainerStateNotFound
	ContainerStateCreated
	ContainerStateRunning
	ContainerStatePaused
	ContainerStateRestarting
	ContainerStateRemoving
	ContainerStateExited
	ContainerStateDead
)

type ContainerState uint8

func (c ContainerState) String() string {
	switch c {
	case ContainerStateUnknown:
		return "Unknown"
	case ContainerStateNotFound:
		return "NotFound"
	case ContainerStateCreated:
		return "Created"
	case ContainerStateRunning:
		return "Running"
	case ContainerStatePaused:
		return "Paused"
	case ContainerStateRestarting:
		return "Restarting"
	case ContainerStateRemoving:
		return "Removing"
	case ContainerStateExited:
		return "Exited"
	case ContainerStateDead:
		return "Dead"
	default:
		panic("Invalid scenario in ContainerState stringer, possibly indicating a bug in the code")
	}
}

func containerStateFromString(state string) ContainerState {
	switch state {
	case "created":
		return ContainerStateCreated
	case "running":
		return ContainerStateRunning
	case "paused":
		return ContainerStatePaused
	case "restarting":
		return ContainerStateRestarting
	case "removing":
		return ContainerStateRemoving
	case "exited":
		return ContainerStateExited
	case "dead":
		return ContainerStateDead
	default:
		return ContainerStateUnknown
	}
}
