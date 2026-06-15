package cmd

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"golang.org/x/oauth2"

	"github.com/spf13/cobra"
)

func newLoginTestCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "login"}
	cmd.Flags().String("domain", "", "")
	cmd.Flags().String("app-key", "", "")
	return cmd
}

func TestLoginTokenType(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"", tokenPersonal},
		{"personal", tokenPersonal},
		{"team-access", tokenTeamAccess},
		{"team_access", tokenTeamAccess},
		{"teamAccess", tokenTeamAccess},
		{"team-manage", tokenTeamManage},
		{"team_manage", tokenTeamManage},
		{"teamManage", tokenTeamManage},
	}

	for _, tt := range tests {
		got, err := loginTokenType(tt.name)
		if err != nil {
			t.Fatalf("loginTokenType(%q): %v", tt.name, err)
		}
		if got != tt.want {
			t.Fatalf("loginTokenType(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestLoginTokenTypeRejectsUnknown(t *testing.T) {
	if _, err := loginTokenType("unknown"); err == nil {
		t.Fatal("expected unknown token type to fail")
	}
}

func TestLoginWritesCredentials(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "auth.json")
	t.Setenv(envAuthFile, authFile)
	mockAuthorization(t, "auth-code", "login-token")

	cmd := newLoginTestCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := login(cmd, nil); err != nil {
		t.Fatal(err)
	}

	tokens, err := readTokens(authFile)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[""][tokenPersonal].AccessToken != "login-token" {
		t.Fatalf("expected login token to be saved, got %q", tokens[""][tokenPersonal].AccessToken)
	}
	if tokens[""][tokenPersonal].RefreshToken == "" {
		t.Fatal("expected login credentials to include a refresh token")
	}
	if out.String() != "Credentials saved to "+authFile+"\n" {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestLoginWritesSelectedTokenType(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "auth.json")
	t.Setenv(envAuthFile, authFile)
	mockAuthorization(t, "auth-code", "team-token")

	cmd := newLoginTestCommand()

	if err := login(cmd, []string{"team-manage"}); err != nil {
		t.Fatal(err)
	}

	tokens, err := readTokens(authFile)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[""][tokenTeamManage].AccessToken != "team-token" {
		t.Fatalf("expected team manage token to be saved, got %q", tokens[""][tokenTeamManage].AccessToken)
	}
}

func TestLoginUsesAppKeyFlag(t *testing.T) {
	restoreOAuthCredentials(t)
	setOAuthCredentials(tokenPersonal, defaultPersonalAppKey)

	authFile := filepath.Join(t.TempDir(), "auth.json")
	t.Setenv(envAuthFile, authFile)

	origReadAuthorizationCode := readAuthorizationCode
	origExchangeAuthorizationCode := exchangeAuthorizationCode
	t.Cleanup(func() {
		readAuthorizationCode = origReadAuthorizationCode
		exchangeAuthorizationCode = origExchangeAuthorizationCode
	})

	readAppCredentials = func(tokType string) (appCredentials, error) {
		t.Fatal("app credential prompt should not be used when app key flag is set")
		return appCredentials{}, nil
	}
	readAuthorizationCode = func() (string, error) {
		return "auth-code", nil
	}
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string, verifier string) (*oauth2.Token, error) {
		if conf.ClientID != "flag-key" {
			t.Fatalf("expected flag app key, got %q", conf.ClientID)
		}
		if conf.ClientSecret != "" {
			t.Fatalf("expected no client secret for PKCE, got %q", conf.ClientSecret)
		}
		return &oauth2.Token{AccessToken: "flag-token", RefreshToken: "refresh-token"}, nil
	}

	cmd := newLoginTestCommand()
	if err := cmd.Flags().Set("app-key", "flag-key"); err != nil {
		t.Fatal(err)
	}

	if err := login(cmd, nil); err != nil {
		t.Fatal(err)
	}

	tokens, err := readTokens(authFile)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[""][tokenPersonal].AccessToken != "flag-token" {
		t.Fatalf("expected login token to be saved, got %q", tokens[""][tokenPersonal].AccessToken)
	}
	if tokens[""][tokenPersonal].AppKey != "flag-key" {
		t.Fatalf("expected saved app key, got %q", tokens[""][tokenPersonal].AppKey)
	}
}

func TestLoginAppKeyFlagUsesBundledDefaultKey(t *testing.T) {
	restoreOAuthCredentials(t)
	setOAuthCredentials(tokenPersonal, defaultPersonalAppKey)

	authFile := filepath.Join(t.TempDir(), "auth.json")
	t.Setenv(envAuthFile, authFile)

	origReadAuthorizationCode := readAuthorizationCode
	origExchangeAuthorizationCode := exchangeAuthorizationCode
	t.Cleanup(func() {
		readAuthorizationCode = origReadAuthorizationCode
		exchangeAuthorizationCode = origExchangeAuthorizationCode
	})

	readAppCredentials = func(tokType string) (appCredentials, error) {
		t.Fatal("full app credential prompt should not be used when app key flag is set")
		return appCredentials{}, nil
	}
	readAuthorizationCode = func() (string, error) {
		return "auth-code", nil
	}
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string, verifier string) (*oauth2.Token, error) {
		if conf.ClientID != "flag-key" {
			t.Fatalf("expected flag app key, got %q", conf.ClientID)
		}
		if conf.ClientSecret != "" {
			t.Fatalf("expected no client secret for PKCE, got %q", conf.ClientSecret)
		}
		return &oauth2.Token{AccessToken: "flag-token", RefreshToken: "refresh-token"}, nil
	}

	cmd := newLoginTestCommand()
	if err := cmd.Flags().Set("app-key", "flag-key"); err != nil {
		t.Fatal(err)
	}

	if err := login(cmd, nil); err != nil {
		t.Fatal(err)
	}
}

func TestLoginAppKeyFlagSetsKey(t *testing.T) {
	restoreOAuthCredentials(t)
	setOAuthCredentials(tokenPersonal, defaultPersonalAppKey)

	cmd := newLoginTestCommand()
	if err := cmd.Flags().Set("app-key", "flag-key"); err != nil {
		t.Fatal(err)
	}

	if err := loginAppKeyFromFlag(cmd, tokenPersonal); err != nil {
		t.Fatal(err)
	}
	appKey := oauthCredentials(tokenPersonal)
	if appKey != "flag-key" {
		t.Fatalf("expected app key from flag, got %q", appKey)
	}
}

func TestLoginAndLogoutSkipRootAuthPreRun(t *testing.T) {
	if loginCmd.PersistentPreRunE == nil {
		t.Fatal("login command should override root auth pre-run")
	}
	if logoutCmd.PersistentPreRunE == nil {
		t.Fatal("logout command should override root auth pre-run")
	}
	if err := loginCmd.PersistentPreRunE(loginCmd, nil); err != nil {
		t.Fatalf("login pre-run returned error: %v", err)
	}
	if err := logoutCmd.PersistentPreRunE(logoutCmd, nil); err != nil {
		t.Fatalf("logout pre-run returned error: %v", err)
	}
}

func TestLoginCommandDefinesAppKeyFlagOnly(t *testing.T) {
	if loginCmd.Flags().Lookup("app-key") == nil {
		t.Fatal("login command should define --app-key")
	}
	if loginCmd.Flags().Lookup("app-secret") != nil {
		t.Fatal("login command should not define --app-secret")
	}
}
