package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAuthFilePathUsesEnv(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "custom-auth.json")
	t.Setenv(envAuthFile, authFile)

	path, err := authFilePath()
	if err != nil {
		t.Fatal(err)
	}
	if path != authFile {
		t.Fatalf("expected auth file %q, got %q", authFile, path)
	}
}

func TestReadTokensReadsTokenMap(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(authFile, []byte(`{"":{"personal":"personal-token"},"api.example.com":{"teamManage":"team-token"}}`), 0600); err != nil {
		t.Fatal(err)
	}

	tokens, err := readTokens(authFile)
	if err != nil {
		t.Fatal(err)
	}

	if tokens[""][tokenPersonal] != "personal-token" {
		t.Fatalf("expected personal token, got %q", tokens[""][tokenPersonal])
	}
	if tokens["api.example.com"][tokenTeamManage] != "team-token" {
		t.Fatalf("expected team token, got %q", tokens["api.example.com"][tokenTeamManage])
	}
}

func TestReadTokensReturnsUnmarshalError(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(authFile, []byte(`not-json`), 0600); err != nil {
		t.Fatal(err)
	}

	if _, err := readTokens(authFile); err == nil {
		t.Fatal("expected invalid JSON to return an error")
	}
}

func TestWriteTokensCreatesParentDirectory(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "nested", "auth.json")
	want := TokenMap{
		"": {
			tokenPersonal: "personal-token",
		},
		"api.example.com": {
			tokenTeamAccess: "team-access-token",
		},
	}

	if err := writeTokens(authFile, want); err != nil {
		t.Fatal(err)
	}

	got, err := readTokens(authFile)
	if err != nil {
		t.Fatal(err)
	}

	if got[""][tokenPersonal] != want[""][tokenPersonal] {
		t.Fatalf("expected personal token %q, got %q", want[""][tokenPersonal], got[""][tokenPersonal])
	}
	if got["api.example.com"][tokenTeamAccess] != want["api.example.com"][tokenTeamAccess] {
		t.Fatalf("expected team access token %q, got %q", want["api.example.com"][tokenTeamAccess], got["api.example.com"][tokenTeamAccess])
	}
}
