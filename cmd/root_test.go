package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dbxcli/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/common"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/users"
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

func TestCompletionJSONUnsupportedOutputReturnsError(t *testing.T) {
	t.Setenv(envAccessToken, "")
	t.Setenv(envAuthFile, filepath.Join(t.TempDir(), "missing-auth.json"))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := &cobra.Command{
		Use:               "dbxcli",
		SilenceUsage:      true,
		SilenceErrors:     true,
		PersistentPreRunE: initDbx,
	}
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"completion", "--output=json"})
	root.PersistentFlags().BoolP("verbose", "v", false, "")
	root.PersistentFlags().String(outputFlag, "text", "")
	root.PersistentFlags().String("as-member", "", "")
	root.PersistentFlags().String("domain", "", "")
	root.AddCommand(newCompletionCmd())

	err := root.Execute()
	if !errors.Is(err, output.ErrStructuredOutputUnsupported) {
		t.Fatalf("error = %v, want structured output unsupported", err)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want no text help output", got)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestHelpOutputRemainsTextWithJSONFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "root help",
			args: []string{"--help", "--output=json"},
		},
		{
			name: "root no command",
			args: []string{"--output=json"},
		},
		{
			name: "command help",
			args: []string{"version", "--help", "--output=json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			root := &cobra.Command{
				Use:               "dbxcli",
				SilenceUsage:      true,
				SilenceErrors:     true,
				PersistentPreRunE: initDbx,
			}
			root.SetOut(&stdout)
			root.SetErr(&stderr)
			root.SetArgs(tt.args)
			root.PersistentFlags().BoolP("verbose", "v", false, "")
			root.PersistentFlags().String(outputFlag, "text", "")
			root.PersistentFlags().String("as-member", "", "")
			root.PersistentFlags().String("domain", "", "")
			root.AddCommand(NewVersionCommand("test-version"))

			if err := root.Execute(); err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			if got := stdout.String(); !strings.Contains(got, "Usage:") {
				t.Fatalf("stdout = %q, want text help", got)
			}
			if strings.Contains(stdout.String(), `"ok"`) {
				t.Fatalf("stdout = %q, want text help without JSON envelope", stdout.String())
			}
			if got := stderr.String(); got != "" {
				t.Fatalf("stderr = %q, want empty", got)
			}
		})
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
	stubUsersClient(t, &mockUsersClient{
		getCurrentAccountFn: func() (*users.FullAccount, error) {
			return testRootFullAccount(common.NewUserRootInfo("root-ns", "home-ns")), nil
		},
	})

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
	if config.PathRoot != `{".tag": "root", "root": "root-ns"}` {
		t.Fatalf("expected path root from env token account, got %q", config.PathRoot)
	}
}

func TestInitDbxAccessTokenEnvTakesPrecedenceOverAuthFile(t *testing.T) {
	origConfig := config
	defer func() { config = origConfig }()
	stubUsersClient(t, &mockUsersClient{
		getCurrentAccountFn: func() (*users.FullAccount, error) {
			return testRootFullAccount(common.NewUserRootInfo("env-root", "env-home")), nil
		},
	})

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
	if config.PathRoot != `{".tag": "root", "root": "env-root"}` {
		t.Fatalf("expected path root from env token account, got %q", config.PathRoot)
	}
}

func TestInitDbxAccessTokenEnvBypassesRefresh(t *testing.T) {
	origConfig := config
	defer func() { config = origConfig }()
	restoreOAuthCredentials(t)
	stubUsersClient(t, &mockUsersClient{
		getCurrentAccountFn: func() (*users.FullAccount, error) {
			return testRootFullAccount(common.NewUserRootInfo("env-root", "env-home")), nil
		},
	})

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
	if config.PathRoot != `{".tag": "root", "root": "env-root"}` {
		t.Fatalf("expected path root from env token account, got %q", config.PathRoot)
	}
}

func TestInitDbxUsesAuthFileEnv(t *testing.T) {
	origConfig := config
	defer func() { config = origConfig }()
	stubUsersClient(t, &mockUsersClient{
		getCurrentAccountFn: func() (*users.FullAccount, error) {
			return testRootFullAccount(common.NewTeamRootInfo("team-root", "home-ns", "/Member")), nil
		},
	})

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
	if config.PathRoot != `{".tag": "root", "root": "team-root"}` {
		t.Fatalf("expected path root from saved token account, got %q", config.PathRoot)
	}
}

func TestWithRootNamespaceSkipsTeamManage(t *testing.T) {
	cfg := makeDropboxConfig("team-token", false, "", "")
	stubUsersClient(t, &mockUsersClient{
		getCurrentAccountFn: func() (*users.FullAccount, error) {
			t.Fatal("team manage token should not fetch current account")
			return nil, nil
		},
	})

	got := withRootNamespace(cfg, tokenTeamManage)

	if got.PathRoot != "" {
		t.Fatalf("path root = %q, want empty", got.PathRoot)
	}
}

func TestWithRootNamespaceKeepsConfigOnAccountError(t *testing.T) {
	cfg := makeDropboxConfig("token", false, "dbmid:member", "api.example.com")
	stubUsersClient(t, &mockUsersClient{
		getCurrentAccountFn: func() (*users.FullAccount, error) {
			return nil, errors.New("account unavailable")
		},
	})

	got := withRootNamespace(cfg, tokenPersonal)

	if got.Token != cfg.Token || got.AsMemberID != cfg.AsMemberID || got.Domain != cfg.Domain {
		t.Fatalf("config = %#v, want original token/as-member/domain", got)
	}
	if got.PathRoot != "" {
		t.Fatalf("path root = %q, want empty on account error", got.PathRoot)
	}
}

func TestRootNamespaceID(t *testing.T) {
	tests := []struct {
		name    string
		account *users.FullAccount
		want    string
	}{
		{
			name:    "team root",
			account: testRootFullAccount(common.NewTeamRootInfo("team-root", "home-ns", "/Member")),
			want:    "team-root",
		},
		{
			name:    "user root",
			account: testRootFullAccount(common.NewUserRootInfo("user-root", "home-ns")),
			want:    "user-root",
		},
		{
			name:    "nil account",
			account: nil,
			want:    "",
		},
		{
			name:    "nil root info",
			account: testRootFullAccount(nil),
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rootNamespaceID(tt.account); got != tt.want {
				t.Fatalf("root namespace ID = %q, want %q", got, tt.want)
			}
		})
	}
}

func testRootFullAccount(rootInfo common.IsRootInfo) *users.FullAccount {
	return users.NewFullAccount(
		"dbid:root",
		users.NewName("Test", "User", "Test", "Test User", "TU"),
		"test@example.com",
		true,
		false,
		"en",
		"",
		false,
		nil,
		rootInfo,
	)
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
