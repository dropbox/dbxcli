package cmd

import (
	"context"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/users"
)

type mockUsersClient struct {
	getAccountFn        func(*users.GetAccountArg) (*users.BasicAccount, error)
	getCurrentAccountFn func() (*users.FullAccount, error)
	getSpaceUsageFn     func() (*users.SpaceUsage, error)
}

func (m *mockUsersClient) GetAccount(arg *users.GetAccountArg) (*users.BasicAccount, error) {
	if m.getAccountFn != nil {
		return m.getAccountFn(arg)
	}
	return nil, nil
}

func (m *mockUsersClient) GetAccountContext(ctx context.Context, arg *users.GetAccountArg) (*users.BasicAccount, error) {
	return m.GetAccount(arg)
}

func (m *mockUsersClient) GetCurrentAccount() (*users.FullAccount, error) {
	if m.getCurrentAccountFn != nil {
		return m.getCurrentAccountFn()
	}
	return nil, nil
}

func (m *mockUsersClient) GetCurrentAccountContext(ctx context.Context) (*users.FullAccount, error) {
	return m.GetCurrentAccount()
}

func (m *mockUsersClient) GetSpaceUsage() (*users.SpaceUsage, error) {
	if m.getSpaceUsageFn != nil {
		return m.getSpaceUsageFn()
	}
	return nil, nil
}

func (m *mockUsersClient) GetSpaceUsageContext(ctx context.Context) (*users.SpaceUsage, error) {
	return m.GetSpaceUsage()
}

func stubUsersClient(t *testing.T, client usersClient) {
	t.Helper()

	origNew := usersNewFunc
	usersNewFunc = func(_ dropbox.Config) usersClient { return client }
	t.Cleanup(func() { usersNewFunc = origNew })
}
