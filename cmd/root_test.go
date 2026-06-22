package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

func TestRootCmdUnknownCommandReturnsError(t *testing.T) {
	RootCmd.SetArgs([]string{"nonexistent-command"})
	err := RootCmd.Execute()
	if err == nil {
		t.Error("expected error for unknown command")
	}
}

func TestRootCmdInvalidFlagReturnsError(t *testing.T) {
	RootCmd.SetArgs([]string{"ls", "--invalidflag"})
	err := RootCmd.Execute()
	if err == nil {
		t.Error("expected error for invalid flag")
	}
}

func newAuthTestCommand() *cobra.Command {
	root := &cobra.Command{Use: "dbxcli"}
	root.PersistentFlags().String(outputFlag, "text", "")
	cmd := &cobra.Command{Use: "ls"}
	cmd.Flags().BoolP("verbose", "v", false, "")
	cmd.Flags().String("as-member", "", "")
	cmd.Flags().String("domain", "", "")
	root.AddCommand(cmd)
	return cmd
}

func TestInitDbxSkipsAuthForLocalCommands(t *testing.T) {
	t.Setenv(envAccessToken, "")
	t.Setenv(envAuthFile, filepath.Join(t.TempDir(), "missing-auth.json"))

	tests := []struct {
		name string
		cmd  *cobra.Command
	}{
		{
			name: "version",
			cmd:  &cobra.Command{Use: "version"},
		},
		{
			name: "help",
			cmd:  &cobra.Command{Use: "help"},
		},
		{
			name: "completion",
			cmd: func() *cobra.Command {
				root := &cobra.Command{Use: "dbxcli"}
				completion := &cobra.Command{Use: "completion"}
				bash := &cobra.Command{Use: "bash"}
				completion.AddCommand(bash)
				root.AddCommand(completion)
				return bash
			}(),
		},
		{
			name: "complete",
			cmd:  &cobra.Command{Use: "__complete"},
		},
		{
			name: "complete-no-desc",
			cmd:  &cobra.Command{Use: "__completeNoDesc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := initDbx(tt.cmd, nil); err != nil {
				t.Fatalf("expected auth to be skipped, got %v", err)
			}
		})
	}
}

func TestInitDbxValidatesOutputBeforeAuth(t *testing.T) {
	t.Setenv(envAccessToken, "")
	t.Setenv(envAuthFile, filepath.Join(t.TempDir(), "missing-auth.json"))

	cmd := newAuthTestCommand()
	if err := cmd.Root().PersistentFlags().Set(outputFlag, "yaml"); err != nil {
		t.Fatal(err)
	}

	err := initDbx(cmd, nil)
	if err == nil {
		t.Fatal("expected invalid output format to fail")
	}
	if !strings.Contains(err.Error(), `unsupported output format "yaml"`) {
		t.Fatalf("error = %q, want output format error", err.Error())
	}
}

func TestInitDbxRejectsUnsupportedStructuredOutputBeforeAuth(t *testing.T) {
	t.Setenv(envAccessToken, "")
	t.Setenv(envAuthFile, filepath.Join(t.TempDir(), "missing-auth.json"))

	cmd := newAuthTestCommand()
	if err := cmd.Root().PersistentFlags().Set(outputFlag, "json"); err != nil {
		t.Fatal(err)
	}

	err := initDbx(cmd, nil)
	if err == nil {
		t.Fatal("expected unsupported structured output to fail")
	}
	if !strings.Contains(err.Error(), "structured output is not supported") {
		t.Fatalf("error = %q, want structured output error", err.Error())
	}
}

func TestInitDbxStillRequiresAuthForDropboxCommands(t *testing.T) {
	t.Setenv(envAccessToken, "")
	t.Setenv(envAuthFile, filepath.Join(t.TempDir(), "missing-auth.json"))

	cmd := newAuthTestCommand()
	if err := initDbx(cmd, nil); err == nil {
		t.Fatal("expected Dropbox command to require auth")
	}
}

func TestInitDbxUsesAccessTokenEnv(t *testing.T) {
	origConfig := config
	defer func() { config = origConfig }()

	t.Setenv(envAccessToken, "env-token")
	t.Setenv(envAuthFile, filepath.Join(t.TempDir(), "missing-auth.json"))

	cmd := newAuthTestCommand()
	if err := cmd.Flags().Set("verbose", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("as-member", "dbmid:member"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("domain", "api.example.com"); err != nil {
		t.Fatal(err)
	}

	if err := initDbx(cmd, nil); err != nil {
		t.Fatal(err)
	}

	if config.Token != "env-token" {
		t.Fatalf("expected token from %s, got %q", envAccessToken, config.Token)
	}
	if config.AsMemberID != "dbmid:member" {
		t.Fatalf("expected as-member to be preserved, got %q", config.AsMemberID)
	}
	if config.Domain != "api.example.com" {
		t.Fatalf("expected domain to be preserved, got %q", config.Domain)
	}
	if config.LogLevel != dropbox.LogInfo {
		t.Fatalf("expected verbose log level, got %v", config.LogLevel)
	}
}

func TestInitDbxAccessTokenEnvTakesPrecedenceOverAuthFile(t *testing.T) {
	origConfig := config
	defer func() { config = origConfig }()

	authFile := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(authFile, []byte(`{"":{"personal":"file-token"}}`), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv(envAccessToken, "env-token")
	t.Setenv(envAuthFile, authFile)

	cmd := newAuthTestCommand()
	if err := initDbx(cmd, nil); err != nil {
		t.Fatal(err)
	}

	if config.Token != "env-token" {
		t.Fatalf("expected %s to take precedence, got %q", envAccessToken, config.Token)
	}
}

func TestInitDbxAccessTokenEnvBypassesRefresh(t *testing.T) {
	origConfig := config
	defer func() { config = origConfig }()
	restoreOAuthCredentials(t)

	expired := time.Now().Add(-time.Hour)
	authFile := filepath.Join(t.TempDir(), "auth.json")
	if err := writeTokens(authFile, TokenMap{
		"": {
			tokenPersonal: {
				AccessToken:  "file-token",
				RefreshToken: "refresh-token",
				Expiry:       &expired,
				AppKey:       "app-key",
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	refreshOAuthToken = func(ctx context.Context, conf *oauth2.Config, token *oauth2.Token) (*oauth2.Token, error) {
		t.Fatal("refresh should not run when DBXCLI_ACCESS_TOKEN is set")
		return nil, nil
	}

	t.Setenv(envAccessToken, "env-token")
	t.Setenv(envAuthFile, authFile)

	cmd := newAuthTestCommand()
	if err := initDbx(cmd, nil); err != nil {
		t.Fatal(err)
	}

	if config.Token != "env-token" {
		t.Fatalf("expected %s to take precedence, got %q", envAccessToken, config.Token)
	}
}

func TestInitDbxUsesAuthFileEnv(t *testing.T) {
	origConfig := config
	defer func() { config = origConfig }()

	authFile := filepath.Join(t.TempDir(), "custom-auth.json")
	if err := os.WriteFile(authFile, []byte(`{"api.example.com":{"personal":"file-token"}}`), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv(envAccessToken, "")
	t.Setenv(envAuthFile, authFile)

	cmd := newAuthTestCommand()
	if err := cmd.Flags().Set("domain", "api.example.com"); err != nil {
		t.Fatal(err)
	}

	if err := initDbx(cmd, nil); err != nil {
		t.Fatal(err)
	}

	if config.Token != "file-token" {
		t.Fatalf("expected token from %s, got %q", envAuthFile, config.Token)
	}
	if config.Domain != "api.example.com" {
		t.Fatalf("expected domain from flag, got %q", config.Domain)
	}
}

func unsetEnvForTest(t *testing.T, key string) {
	t.Helper()

	old, exists := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if exists {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

func TestLoadOAuthCredentialsFromEnvKeepsAppKeyFallbacks(t *testing.T) {
	restoreOAuthCredentials(t)

	for _, key := range []string{
		"DROPBOX_PERSONAL_APP_KEY",
		"DROPBOX_TEAM_APP_KEY",
		"DROPBOX_MANAGE_APP_KEY",
	} {
		unsetEnvForTest(t, key)
	}

	teamAccessAppKey = "team-access-key"
	teamManageAppKey = "team-manage-key"

	loadOAuthCredentialsFromEnv()

	if teamManageAppKey != "team-manage-key" {
		t.Fatalf("expected team manage app key fallback, got %q", teamManageAppKey)
	}
	if teamAccessAppKey != "team-access-key" {
		t.Fatalf("expected team access app key fallback, got %q", teamAccessAppKey)
	}
}
