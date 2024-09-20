package main

func queryContainers(dep *deployment, allGroups bool, group, container string) containerList {
	if allGroups {
		return containerMapToList(dep.queryAllContainers())
	} else if group != "" && container == "" {
		return containerMapToList(dep.queryAllContainersInGroup(group))
	} else if group != "" {
		c := dep.queryContainer(&ContainerReference{Group: group, Container: container})
		if c != nil {
			return containerList{c}
		}
		return nil
	}
	log.Fatalf("Invalid scenario, possibly indicating a bug in the code")
	return nil
}
