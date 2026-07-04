package cmd

import (
	"context"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/users"
)

type usersClient interface {
	GetAccountContext(context.Context, *users.GetAccountArg) (*users.BasicAccount, error)
	GetCurrentAccountContext(context.Context) (*users.FullAccount, error)
	GetSpaceUsageContext(context.Context) (*users.SpaceUsage, error)
}

var usersNewFunc = func(cfg dropbox.Config) usersClient {
	return users.NewContext(cfg)
}
