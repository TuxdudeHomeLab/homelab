package main

import (
	"context"
	"os/user"
)

type userInfo struct {
	user         user.User
	primaryGroup user.Group
	allGroups    []user.Group
}

var (
	_ = newUserInfo
)

func newUserInfo(ctx context.Context) *userInfo {
	u, err := user.Current()
	if err != nil {
		log(ctx).Fatalf("Unable to retrieve the current user info, reason: %v", err)
	}

	gids, err := u.GroupIds()
	if err != nil {
		log(ctx).Fatalf("Unable to retrieve the Group IDs for the user %s, reason: %v", u.Username, err)
	}

	var pg *user.Group
	var groups []user.Group
	for _, gid := range gids {
		g, err := user.LookupGroupId(gid)
		if err != nil {
			log(ctx).Fatalf("Unable to retrieve the group info for group ID %s, reason: %v", gid, err)
		}
		if gid == u.Gid {
			pg = g
		}
		groups = append(groups, *g)
	}

	return &userInfo{
		user:         *u,
		primaryGroup: *pg,
		allGroups:    groups,
	}

}
