package cmd

import (
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/users"
)

type usersClient interface {
	GetAccount(*users.GetAccountArg) (*users.BasicAccount, error)
	GetCurrentAccount() (*users.FullAccount, error)
	GetSpaceUsage() (*users.SpaceUsage, error)
}

var usersNewFunc = func(cfg dropbox.Config) usersClient {
	return users.New(cfg)
}
