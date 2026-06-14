package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/oauth2"
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

func restoreOAuthCredentials(t *testing.T) {
	t.Helper()

	origPersonalAppKey := personalAppKey
	origPersonalAppSecret := personalAppSecret
	origTeamAccessAppKey := teamAccessAppKey
	origTeamAccessAppSecret := teamAccessAppSecret
	origTeamManageAppKey := teamManageAppKey
	origTeamManageAppSecret := teamManageAppSecret
	origReadAppKey := readAppKey
	origReadAppSecret := readAppSecret
	origReadAppCredentials := readAppCredentials
	t.Cleanup(func() {
		personalAppKey = origPersonalAppKey
		personalAppSecret = origPersonalAppSecret
		teamAccessAppKey = origTeamAccessAppKey
		teamAccessAppSecret = origTeamAccessAppSecret
		teamManageAppKey = origTeamManageAppKey
		teamManageAppSecret = origTeamManageAppSecret
		readAppKey = origReadAppKey
		readAppSecret = origReadAppSecret
		readAppCredentials = origReadAppCredentials
	})
}

func mockOAuthAppCredentials(t *testing.T) {
	t.Helper()

	restoreOAuthCredentials(t)
	setOAuthCredentials(tokenPersonal, "personal-test-key", "personal-test-secret")
	setOAuthCredentials(tokenTeamAccess, "team-access-test-key", "team-access-test-secret")
	setOAuthCredentials(tokenTeamManage, "team-manage-test-key", "team-manage-test-secret")
	readAppCredentials = func(tokType string) (appCredentials, error) {
		t.Fatal("app credential prompt should not be used")
		return appCredentials{}, nil
	}
}

func mockAuthorization(t *testing.T, code string, accessToken string) {
	t.Helper()

	mockOAuthAppCredentials(t)

	origReadAuthorizationCode := readAuthorizationCode
	origExchangeAuthorizationCode := exchangeAuthorizationCode
	t.Cleanup(func() {
		readAuthorizationCode = origReadAuthorizationCode
		exchangeAuthorizationCode = origExchangeAuthorizationCode
	})

	readAuthorizationCode = func() (string, error) {
		return code, nil
	}
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, gotCode string) (*oauth2.Token, error) {
		if gotCode != code {
			t.Fatalf("expected authorization code %q, got %q", code, gotCode)
		}
		return &oauth2.Token{AccessToken: accessToken}, nil
	}
}

func TestGetAccessTokenUsesExistingToken(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(authFile, []byte(`{"":{"personal":"existing-token"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(envAuthFile, authFile)

	origReadAuthorizationCode := readAuthorizationCode
	origExchangeAuthorizationCode := exchangeAuthorizationCode
	t.Cleanup(func() {
		readAuthorizationCode = origReadAuthorizationCode
		exchangeAuthorizationCode = origExchangeAuthorizationCode
	})
	readAuthorizationCode = func() (string, error) {
		t.Fatal("authorization prompt should not be used for existing token")
		return "", nil
	}
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string) (*oauth2.Token, error) {
		t.Fatal("authorization exchange should not be used for existing token")
		return nil, nil
	}

	token, filePath, err := getAccessToken(tokenPersonal, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if token != "existing-token" {
		t.Fatalf("expected existing token, got %q", token)
	}
	if filePath != authFile {
		t.Fatalf("expected auth file %q, got %q", authFile, filePath)
	}
}

func TestGetAccessTokenForceRefreshesToken(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(authFile, []byte(`{"":{"personal":"old-token"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(envAuthFile, authFile)
	mockAuthorization(t, "auth-code", "new-token")

	token, filePath, err := getAccessToken(tokenPersonal, "", true)
	if err != nil {
		t.Fatal(err)
	}
	if token != "new-token" {
		t.Fatalf("expected new token, got %q", token)
	}
	if filePath != authFile {
		t.Fatalf("expected auth file %q, got %q", authFile, filePath)
	}

	tokens, err := readTokens(authFile)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[""][tokenPersonal] != "new-token" {
		t.Fatalf("expected saved token to be refreshed, got %q", tokens[""][tokenPersonal])
	}
}

func TestGetAccessTokenMissingTokenWithBundledCredentialsReturnsLoginError(t *testing.T) {
	restoreOAuthCredentials(t)
	setOAuthCredentials(tokenPersonal, defaultPersonalAppKey, defaultPersonalAppSecret)

	authFile := filepath.Join(t.TempDir(), "auth.json")
	t.Setenv(envAuthFile, authFile)

	origReadAuthorizationCode := readAuthorizationCode
	origExchangeAuthorizationCode := exchangeAuthorizationCode
	t.Cleanup(func() {
		readAuthorizationCode = origReadAuthorizationCode
		exchangeAuthorizationCode = origExchangeAuthorizationCode
	})

	readAppCredentials = func(tokType string) (appCredentials, error) {
		t.Fatal("app credential prompt should not run for command lazy auth")
		return appCredentials{}, nil
	}
	readAuthorizationCode = func() (string, error) {
		t.Fatal("authorization prompt should not run when app credentials are missing")
		return "", nil
	}
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string) (*oauth2.Token, error) {
		t.Fatal("authorization exchange should not run when app credentials are missing")
		return nil, nil
	}

	_, _, err := getAccessToken(tokenPersonal, "", false)
	if err == nil {
		t.Fatal("expected missing credentials error")
	}
	if !strings.Contains(err.Error(), "dbxcli login --app-key=<your-app-key>") {
		t.Fatalf("expected login hint, got %q", err)
	}
	if !strings.Contains(err.Error(), envAccessToken) {
		t.Fatalf("expected %s hint, got %q", envAccessToken, err)
	}
}

func TestLoginCommandForTokenType(t *testing.T) {
	tests := []struct {
		tokType string
		want    string
	}{
		{tokenPersonal, "dbxcli login --app-key=<your-app-key>"},
		{tokenTeamAccess, "dbxcli login team-access --app-key=<your-app-key>"},
		{tokenTeamManage, "dbxcli login team-manage --app-key=<your-app-key>"},
	}

	for _, tt := range tests {
		if got := loginCommand(tt.tokType); got != tt.want {
			t.Fatalf("loginCommand(%q) = %q, want %q", tt.tokType, got, tt.want)
		}
	}
}

func TestRequestAccessTokenRejectsEmptyToken(t *testing.T) {
	mockOAuthAppCredentials(t)

	origReadAuthorizationCode := readAuthorizationCode
	origExchangeAuthorizationCode := exchangeAuthorizationCode
	t.Cleanup(func() {
		readAuthorizationCode = origReadAuthorizationCode
		exchangeAuthorizationCode = origExchangeAuthorizationCode
	})

	readAuthorizationCode = func() (string, error) {
		return "auth-code", nil
	}
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string) (*oauth2.Token, error) {
		return &oauth2.Token{}, nil
	}

	if _, err := requestAccessToken(tokenPersonal, ""); err == nil {
		t.Fatal("expected empty access token to return an error")
	}
}

func TestRequestAccessTokenReturnsReadError(t *testing.T) {
	mockOAuthAppCredentials(t)

	origReadAuthorizationCode := readAuthorizationCode
	origExchangeAuthorizationCode := exchangeAuthorizationCode
	t.Cleanup(func() {
		readAuthorizationCode = origReadAuthorizationCode
		exchangeAuthorizationCode = origExchangeAuthorizationCode
	})

	readAuthorizationCode = func() (string, error) {
		return "", errors.New("read failed")
	}
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string) (*oauth2.Token, error) {
		t.Fatal("authorization exchange should not run when reading code fails")
		return nil, nil
	}

	if _, err := requestAccessToken(tokenPersonal, ""); err == nil {
		t.Fatal("expected authorization code read error")
	}
}

func TestRequestAccessTokenPromptsForAppCredentialsWhenUsingBundledDefaults(t *testing.T) {
	restoreOAuthCredentials(t)
	setOAuthCredentials(tokenPersonal, defaultPersonalAppKey, defaultPersonalAppSecret)

	origReadAuthorizationCode := readAuthorizationCode
	origExchangeAuthorizationCode := exchangeAuthorizationCode
	t.Cleanup(func() {
		readAuthorizationCode = origReadAuthorizationCode
		exchangeAuthorizationCode = origExchangeAuthorizationCode
	})

	readAppCredentials = func(tokType string) (appCredentials, error) {
		if tokType != tokenPersonal {
			t.Fatalf("expected personal app credentials prompt, got %q", tokType)
		}
		return appCredentials{Key: "prompt-key", Secret: "prompt-secret"}, nil
	}
	readAuthorizationCode = func() (string, error) {
		return "auth-code", nil
	}
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string) (*oauth2.Token, error) {
		if conf.ClientID != "prompt-key" {
			t.Fatalf("expected prompted app key, got %q", conf.ClientID)
		}
		if conf.ClientSecret != "prompt-secret" {
			t.Fatalf("expected prompted app secret, got %q", conf.ClientSecret)
		}
		return &oauth2.Token{AccessToken: "access-token"}, nil
	}

	token, err := requestAccessToken(tokenPersonal, "")
	if err != nil {
		t.Fatal(err)
	}
	if token != "access-token" {
		t.Fatalf("expected access token, got %q", token)
	}
	if personalAppKey != "prompt-key" {
		t.Fatalf("expected prompted app key to be saved for this process, got %q", personalAppKey)
	}
}

func TestReadAppCredentialsReadsVisibleKeyAndMaskedSecret(t *testing.T) {
	restoreOAuthCredentials(t)

	var keyPrompts []string
	readAppKey = func(prompt string) (string, error) {
		keyPrompts = append(keyPrompts, prompt)
		return "visible-key", nil
	}

	var secretPrompts []string
	readAppSecret = func(prompt string) (string, error) {
		secretPrompts = append(secretPrompts, prompt)
		return "masked-secret", nil
	}

	creds, err := readAppCredentials(tokenPersonal)
	if err != nil {
		t.Fatal(err)
	}
	if creds.Key != "visible-key" {
		t.Fatalf("expected app key, got %q", creds.Key)
	}
	if creds.Secret != "masked-secret" {
		t.Fatalf("expected app secret, got %q", creds.Secret)
	}
	if len(keyPrompts) != 1 {
		t.Fatalf("expected one app key prompt, got %d", len(keyPrompts))
	}
	if keyPrompts[0] != "Dropbox app key: " {
		t.Fatalf("unexpected app key prompt %q", keyPrompts[0])
	}
	if len(secretPrompts) != 1 {
		t.Fatalf("expected one app secret prompt, got %d", len(secretPrompts))
	}
	if secretPrompts[0] != "Dropbox app secret: " {
		t.Fatalf("unexpected app secret prompt %q", secretPrompts[0])
	}
}

func TestRequestAccessTokenUsesConfiguredAppCredentials(t *testing.T) {
	restoreOAuthCredentials(t)
	setOAuthCredentials(tokenPersonal, "configured-key", "configured-secret")

	origReadAuthorizationCode := readAuthorizationCode
	origExchangeAuthorizationCode := exchangeAuthorizationCode
	t.Cleanup(func() {
		readAuthorizationCode = origReadAuthorizationCode
		exchangeAuthorizationCode = origExchangeAuthorizationCode
	})

	readAppCredentials = func(tokType string) (appCredentials, error) {
		t.Fatal("app credential prompt should not be used")
		return appCredentials{}, nil
	}
	readAuthorizationCode = func() (string, error) {
		return "auth-code", nil
	}
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string) (*oauth2.Token, error) {
		if conf.ClientID != "configured-key" {
			t.Fatalf("expected configured app key, got %q", conf.ClientID)
		}
		if conf.ClientSecret != "configured-secret" {
			t.Fatalf("expected configured app secret, got %q", conf.ClientSecret)
		}
		return &oauth2.Token{AccessToken: "access-token"}, nil
	}

	if _, err := requestAccessToken(tokenPersonal, ""); err != nil {
		t.Fatal(err)
	}
}

func TestRequestAccessTokenRejectsEmptyAppCredentials(t *testing.T) {
	restoreOAuthCredentials(t)
	setOAuthCredentials(tokenPersonal, defaultPersonalAppKey, defaultPersonalAppSecret)

	origReadAuthorizationCode := readAuthorizationCode
	origExchangeAuthorizationCode := exchangeAuthorizationCode
	t.Cleanup(func() {
		readAuthorizationCode = origReadAuthorizationCode
		exchangeAuthorizationCode = origExchangeAuthorizationCode
	})

	readAppCredentials = func(tokType string) (appCredentials, error) {
		return appCredentials{Key: " ", Secret: "secret"}, nil
	}
	readAuthorizationCode = func() (string, error) {
		t.Fatal("authorization code prompt should not run when app credentials are invalid")
		return "", nil
	}
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string) (*oauth2.Token, error) {
		t.Fatal("authorization exchange should not run when app credentials are invalid")
		return nil, nil
	}

	if _, err := requestAccessToken(tokenPersonal, ""); err == nil {
		t.Fatal("expected empty app credentials to fail")
	}
}
