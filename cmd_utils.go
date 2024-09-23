package main

import "context"

func queryContainers(ctx context.Context, dep *deployment, allGroups bool, group, container string) (containerList, error) {
	if allGroups {
		return containerMapToList(dep.queryAllContainers()), nil
	}
	if group != "" && container == "" {
		ctMap, err := dep.queryAllContainersInGroup(group)
		if err != nil {
			return nil, err
		}
		return containerMapToList(ctMap), nil
	}
	if group != "" {
		ct, err := dep.queryContainer(ContainerReference{Group: group, Container: container})
		if err != nil {
			return nil, err
		}
		return containerList{ct}, nil
	}
	log(ctx).Fatalf("Invalid scenario, possibly indicating a bug in the code")
	return nil, nil
}
