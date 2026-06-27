package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

type logoutOutputForTest struct {
	OK            bool   `json:"ok"`
	SchemaVersion string `json:"schema_version"`
	Command       string `json:"command"`
	Input         struct {
	} `json:"input"`
	Results []struct {
		Status string       `json:"status"`
		Kind   string       `json:"kind"`
		Input  struct{}     `json:"input"`
		Result logoutResult `json:"result"`
	} `json:"results"`
	Warnings []jsonWarning `json:"warnings"`
}

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

func TestLogoutJSONReturnsLoggedOut(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(authFile, []byte(`{"":{"personal":"legacy-token","teamManage":{"access_token":"refreshable-token","refresh_token":"refresh-token","app_key":"app-key"}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(envAuthFile, authFile)

	restoreRevokeAccessToken := stubRevokeAccessToken(t, func(domain string, token string) error {
		return nil
	})
	defer restoreRevokeAccessToken()

	cmd, stdout, stderr := logoutTestCommand("json")
	if err := logout(cmd, nil); err != nil {
		t.Fatal(err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	got := decodeLogoutOutput(t, stdout.String())
	assertLogoutResult(t, got, logoutStatusLoggedOut, true, true)
	if len(got.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want empty", got.Warnings)
	}
	if _, err := os.Stat(authFile); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected auth file to be removed, got %v", err)
	}
}

func TestLogoutJSONReturnsAlreadyLoggedOut(t *testing.T) {
	t.Setenv(envAuthFile, filepath.Join(t.TempDir(), "missing-auth.json"))

	restoreRevokeAccessToken := stubRevokeAccessToken(t, func(domain string, token string) error {
		t.Fatalf("revokeAccessToken(%q, %q) should not be called", domain, token)
		return nil
	})
	defer restoreRevokeAccessToken()

	cmd, stdout, stderr := logoutTestCommand("json")
	if err := logout(cmd, nil); err != nil {
		t.Fatal(err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	got := decodeLogoutOutput(t, stdout.String())
	assertLogoutResult(t, got, logoutStatusAlreadyLoggedOut, false, false)
	if len(got.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want empty", got.Warnings)
	}
}

func TestLogoutJSONWarnsOnRemoteRevokeFailureAfterRemovingCredentials(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(authFile, []byte(`{"":{"personal":"legacy-token"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(envAuthFile, authFile)

	restoreRevokeAccessToken := stubRevokeAccessToken(t, func(domain string, token string) error {
		return fmt.Errorf("revoke failed")
	})
	defer restoreRevokeAccessToken()

	cmd, stdout, stderr := logoutTestCommand("json")
	if err := logout(cmd, nil); err != nil {
		t.Fatal(err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	got := decodeLogoutOutput(t, stdout.String())
	assertLogoutResult(t, got, logoutStatusLoggedOut, true, false)
	if len(got.Warnings) != 1 {
		t.Fatalf("warnings = %+v, want one warning", got.Warnings)
	}
	if got.Warnings[0].Code != jsonWarningCodeTokenRevokeFailed {
		t.Fatalf("warning code = %q, want %q", got.Warnings[0].Code, jsonWarningCodeTokenRevokeFailed)
	}
	if _, err := os.Stat(authFile); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected auth file to be removed, got %v", err)
	}
}

func TestLogoutInvalidAuthFileReturnsErrorAndLeavesCredentials(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "auth.json")
	content := []byte(`not-json`)
	if err := os.WriteFile(authFile, content, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(envAuthFile, authFile)

	restoreRevokeAccessToken := stubRevokeAccessToken(t, func(domain string, token string) error {
		t.Fatalf("revokeAccessToken(%q, %q) should not be called", domain, token)
		return nil
	})
	defer restoreRevokeAccessToken()

	cmd, stdout, stderr := logoutTestCommand("json")
	if err := logout(cmd, nil); err == nil {
		t.Fatal("expected invalid auth file error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	got, err := os.ReadFile(authFile)
	if err != nil {
		t.Fatalf("read auth file: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("auth file = %q, want %q", got, content)
	}
}

func TestLogoutEnvTokenStillActiveReturnsJSONErrorAndLeavesCredentials(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(authFile, []byte(`{"":{"personal":"legacy-token"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(envAuthFile, authFile)
	t.Setenv(envAccessToken, "env-token")

	cmd, stdout, stderr := logoutTestCommand("json")
	err := logout(cmd, nil)
	if err == nil {
		t.Fatal("expected env token error")
	}
	if got, want := jsonErrorCode(err), jsonErrorCodeEnvTokenStillActive; got != want {
		t.Fatalf("jsonErrorCode = %q, want %q", got, want)
	}
	renderCommandError(cmd, err)
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	got := decodeJSONErrorResponse(t, stdout.String())
	if got.Error.Code != jsonErrorCodeEnvTokenStillActive {
		t.Fatalf("code = %q, want %q", got.Error.Code, jsonErrorCodeEnvTokenStillActive)
	}
	if _, err := os.Stat(authFile); err != nil {
		t.Fatalf("expected auth file to remain, got %v", err)
	}
}

func TestLogoutTextEnvTokenStillActiveReturnsErrorAndLeavesCredentials(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(authFile, []byte(`{"":{"personal":"legacy-token"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(envAuthFile, authFile)
	t.Setenv(envAccessToken, "env-token")

	cmd, stdout, stderr := logoutTestCommand("text")
	err := logout(cmd, nil)
	if err == nil {
		t.Fatal("expected env token error")
	}
	renderCommandError(cmd, err)
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), envAccessToken) {
		t.Fatalf("stderr = %q, want %s", stderr.String(), envAccessToken)
	}
	if _, err := os.Stat(authFile); err != nil {
		t.Fatalf("expected auth file to remain, got %v", err)
	}
}

func TestLogoutSupportsStructuredOutput(t *testing.T) {
	if !commandSupportsStructuredOutput(logoutCmd) {
		t.Fatal("logout command should support structured output")
	}
}

func logoutTestCommand(format string) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{Use: "logout"}
	cmd.Flags().String(outputFlag, format, "")
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	return cmd, &stdout, &stderr
}

func decodeLogoutOutput(t *testing.T, value string) logoutOutputForTest {
	t.Helper()

	var got logoutOutputForTest
	if err := json.Unmarshal([]byte(value), &got); err != nil {
		t.Fatalf("decode logout JSON output %q: %v", value, err)
	}
	return got
}

func assertLogoutResult(t *testing.T, got logoutOutputForTest, status string, removedSavedCredentials bool, remoteTokenRevoked bool) {
	t.Helper()

	if !got.OK {
		t.Fatal("ok = false, want true")
	}
	if got.SchemaVersion != jsonSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", got.SchemaVersion, jsonSchemaVersion)
	}
	if got.Command != "logout" {
		t.Fatalf("command = %q, want logout", got.Command)
	}
	if len(got.Results) != 1 {
		t.Fatalf("results = %+v, want one result", got.Results)
	}
	result := got.Results[0]
	if result.Status != status {
		t.Fatalf("status = %q, want %q", result.Status, status)
	}
	if result.Kind != logoutKindAuth {
		t.Fatalf("kind = %q, want %q", result.Kind, logoutKindAuth)
	}
	if result.Result.RemovedSavedCredentials != removedSavedCredentials {
		t.Fatalf("removed_saved_credentials = %v, want %v", result.Result.RemovedSavedCredentials, removedSavedCredentials)
	}
	if result.Result.RemoteTokenRevoked != remoteTokenRevoked {
		t.Fatalf("remote_token_revoked = %v, want %v", result.Result.RemoteTokenRevoked, remoteTokenRevoked)
	}
}

func stubRevokeAccessToken(t *testing.T, fn func(domain string, token string) error) func() {
	t.Helper()

	orig := revokeAccessToken
	revokeAccessToken = fn
	return func() {
		revokeAccessToken = orig
	}
}
