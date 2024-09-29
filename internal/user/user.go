package user

import (
	"context"
	"os/user"
)

type UserInfo struct {
	User         user.User
	PrimaryGroup user.Group
	AllGroups    []user.Group
}

var (
	_ = NewUserInfo
)

func NewUserInfo(ctx context.Context) *UserInfo {
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

	return &UserInfo{
		User:         *u,
		PrimaryGroup: *pg,
		AllGroups:    groups,
	}

}
