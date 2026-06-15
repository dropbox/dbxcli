package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestLogoutRevokesLegacyAndRefreshableCredentials(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(authFile, []byte(`{"":{"personal":"legacy-token","teamManage":{"access_token":"refreshable-token","refresh_token":"refresh-token","app_key":"app-key"}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(envAuthFile, authFile)

	origRevokeAccessToken := revokeAccessToken
	t.Cleanup(func() {
		revokeAccessToken = origRevokeAccessToken
	})

	revoked := make(map[string]bool)
	revokeAccessToken = func(domain string, token string) error {
		revoked[domain+":"+token] = true
		return nil
	}

	if err := logout(&cobra.Command{Use: "logout"}, nil); err != nil {
		t.Fatal(err)
	}

	if !revoked[":legacy-token"] {
		t.Fatalf("expected legacy token to be revoked, got %#v", revoked)
	}
	if !revoked[":refreshable-token"] {
		t.Fatalf("expected refreshable token to be revoked, got %#v", revoked)
	}
	if _, err := os.Stat(authFile); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected auth file to be removed, got %v", err)
	}
}
