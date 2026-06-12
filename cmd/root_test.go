package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/spf13/cobra"
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
	cmd := &cobra.Command{Use: "ls"}
	cmd.Flags().BoolP("verbose", "v", false, "")
	cmd.Flags().String("as-member", "", "")
	cmd.Flags().String("domain", "", "")
	root.AddCommand(cmd)
	return cmd
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
