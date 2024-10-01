package fakeuser

import (
	osuser "os/user"

	"github.com/tuxdudehomelab/homelab/internal/user"
)

const (
	FakeUserName             = "fakeuser"
	FakeUserID               = "55555"
	FakeUserDisplayName      = "FakeUser"
	FakeUserHomeDir          = "/home/fakeuser"
	FakeUserPrimaryGroupName = "fakegroup1"
	FakeUserPrimaryGroupID   = "44444"
	FakeUserOtherGroupsName1 = "fakegroup2"
	FakeUserOtherGroupsID1   = "44445"
	FakeUserOtherGroupsName2 = "fakegroup3"
	FakeUserOtherGroupsID2   = "44446"
)

func NewFakeUserInfo() *user.UserInfo {
	return &user.UserInfo{
		User: osuser.User{
			Uid:      FakeUserID,
			Gid:      FakeUserPrimaryGroupID,
			Username: FakeUserName,
			Name:     FakeUserDisplayName,
			HomeDir:  FakeUserHomeDir,
		},
		PrimaryGroup: osuser.Group{
			Gid:  FakeUserPrimaryGroupID,
			Name: FakeUserPrimaryGroupName,
		},
		AllGroups: []osuser.Group{
			{
				Gid:  FakeUserPrimaryGroupID,
				Name: FakeUserPrimaryGroupName,
			},
			{
				Gid:  FakeUserOtherGroupsID1,
				Name: FakeUserOtherGroupsName1,
			},
			{
				Gid:  FakeUserOtherGroupsID2,
				Name: FakeUserOtherGroupsName2,
			},
		},
	}
}
