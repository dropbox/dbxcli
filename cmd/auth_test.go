package cmd

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

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

	if tokens[""][tokenPersonal].AccessToken != "personal-token" {
		t.Fatalf("expected personal token, got %q", tokens[""][tokenPersonal].AccessToken)
	}
	if tokens["api.example.com"][tokenTeamManage].AccessToken != "team-token" {
		t.Fatalf("expected team token, got %q", tokens["api.example.com"][tokenTeamManage].AccessToken)
	}
	if tokens[""][tokenPersonal].RefreshToken != "" {
		t.Fatalf("expected legacy personal token to have no refresh token, got %q", tokens[""][tokenPersonal].RefreshToken)
	}
}

func TestReadTokensReadsRefreshableCredentials(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(authFile, []byte(`{"":{"personal":{"access_token":"access-token","refresh_token":"refresh-token","token_type":"Bearer","expiry":"2030-01-02T03:04:05Z","app_key":"app-key"}}}`), 0600); err != nil {
		t.Fatal(err)
	}

	tokens, err := readTokens(authFile)
	if err != nil {
		t.Fatal(err)
	}

	credential := tokens[""][tokenPersonal]
	if credential.AccessToken != "access-token" {
		t.Fatalf("expected access token, got %q", credential.AccessToken)
	}
	if credential.RefreshToken != "refresh-token" {
		t.Fatalf("expected refresh token, got %q", credential.RefreshToken)
	}
	if credential.TokenType != "Bearer" {
		t.Fatalf("expected token type, got %q", credential.TokenType)
	}
	if credential.AppKey != "app-key" {
		t.Fatalf("expected app key, got %q", credential.AppKey)
	}
	if credential.Expiry == nil || credential.Expiry.Format(time.RFC3339) != "2030-01-02T03:04:05Z" {
		t.Fatalf("unexpected expiry: %v", credential.Expiry)
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
	expiry := time.Now().Add(time.Hour).UTC()
	want := TokenMap{
		"": {
			tokenPersonal: {
				AccessToken:  "personal-token",
				RefreshToken: "refresh-token",
				TokenType:    "Bearer",
				Expiry:       &expiry,
				AppKey:       "app-key",
			},
		},
		"api.example.com": {
			tokenTeamAccess: {
				AccessToken: "team-access-token",
			},
		},
	}

	if err := writeTokens(authFile, want); err != nil {
		t.Fatal(err)
	}

	got, err := readTokens(authFile)
	if err != nil {
		t.Fatal(err)
	}

	if got[""][tokenPersonal].AccessToken != want[""][tokenPersonal].AccessToken {
		t.Fatalf("expected personal token %q, got %q", want[""][tokenPersonal].AccessToken, got[""][tokenPersonal].AccessToken)
	}
	if got["api.example.com"][tokenTeamAccess].AccessToken != want["api.example.com"][tokenTeamAccess].AccessToken {
		t.Fatalf("expected team access token %q, got %q", want["api.example.com"][tokenTeamAccess].AccessToken, got["api.example.com"][tokenTeamAccess].AccessToken)
	}
}

func restoreOAuthCredentials(t *testing.T) {
	t.Helper()

	origPersonalAppKey := personalAppKey
	origTeamAccessAppKey := teamAccessAppKey
	origTeamManageAppKey := teamManageAppKey
	origReadAppKey := readAppKey
	origReadAppCredentials := readAppCredentials
	origGenerateOAuthVerifier := generateOAuthVerifier
	origRefreshOAuthToken := refreshOAuthToken
	t.Cleanup(func() {
		personalAppKey = origPersonalAppKey
		teamAccessAppKey = origTeamAccessAppKey
		teamManageAppKey = origTeamManageAppKey
		readAppKey = origReadAppKey
		readAppCredentials = origReadAppCredentials
		generateOAuthVerifier = origGenerateOAuthVerifier
		refreshOAuthToken = origRefreshOAuthToken
	})
}

func mockOAuthAppCredentials(t *testing.T) {
	t.Helper()

	restoreOAuthCredentials(t)
	setOAuthCredentials(tokenPersonal, "personal-test-key")
	setOAuthCredentials(tokenTeamAccess, "team-access-test-key")
	setOAuthCredentials(tokenTeamManage, "team-manage-test-key")
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
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, gotCode string, verifier string) (*oauth2.Token, error) {
		if gotCode != code {
			t.Fatalf("expected authorization code %q, got %q", code, gotCode)
		}
		if verifier == "" {
			t.Fatal("expected PKCE verifier")
		}
		return &oauth2.Token{
			AccessToken:  accessToken,
			RefreshToken: "refresh-token",
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(time.Hour),
		}, nil
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
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string, verifier string) (*oauth2.Token, error) {
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
	if tokens[""][tokenPersonal].AccessToken != "new-token" {
		t.Fatalf("expected saved token to be refreshed, got %q", tokens[""][tokenPersonal].AccessToken)
	}
	if tokens[""][tokenPersonal].RefreshToken == "" {
		t.Fatal("expected saved token to include a refresh token")
	}
}

func TestGetAccessTokenRefreshesExpiredCredential(t *testing.T) {
	expired := time.Now().Add(-time.Hour).UTC()
	authFile := filepath.Join(t.TempDir(), "auth.json")
	tokens := TokenMap{
		"": {
			tokenPersonal: {
				AccessToken:  "old-access",
				RefreshToken: "old-refresh",
				TokenType:    "Bearer",
				Expiry:       &expired,
				AppKey:       "stored-app-key",
			},
		},
	}
	if err := writeTokens(authFile, tokens); err != nil {
		t.Fatal(err)
	}
	t.Setenv(envAuthFile, authFile)

	restoreOAuthCredentials(t)
	refreshExpiry := time.Now().Add(time.Hour).UTC()
	refreshOAuthToken = func(ctx context.Context, conf *oauth2.Config, token *oauth2.Token) (*oauth2.Token, error) {
		if conf.ClientID != "stored-app-key" {
			t.Fatalf("expected stored app key for refresh, got %q", conf.ClientID)
		}
		if token.RefreshToken != "old-refresh" {
			t.Fatalf("expected old refresh token, got %q", token.RefreshToken)
		}
		return &oauth2.Token{
			AccessToken: "new-access",
			TokenType:   "Bearer",
			Expiry:      refreshExpiry,
		}, nil
	}

	token, _, err := getAccessToken(tokenPersonal, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if token != "new-access" {
		t.Fatalf("expected refreshed access token, got %q", token)
	}

	got, err := readTokens(authFile)
	if err != nil {
		t.Fatal(err)
	}
	credential := got[""][tokenPersonal]
	if credential.AccessToken != "new-access" {
		t.Fatalf("expected persisted access token, got %q", credential.AccessToken)
	}
	if credential.RefreshToken != "old-refresh" {
		t.Fatalf("expected refresh token to be preserved, got %q", credential.RefreshToken)
	}
	if credential.Expiry == nil || !credential.Expiry.Equal(refreshExpiry) {
		t.Fatalf("expected persisted expiry %v, got %v", refreshExpiry, credential.Expiry)
	}
}

func TestGetAccessTokenRefreshFailureLeavesAuthFileUnchanged(t *testing.T) {
	expired := time.Now().Add(-time.Hour).UTC()
	authFile := filepath.Join(t.TempDir(), "auth.json")
	tokens := TokenMap{
		"": {
			tokenPersonal: {
				AccessToken:  "old-access",
				RefreshToken: "old-refresh",
				TokenType:    "Bearer",
				Expiry:       &expired,
				AppKey:       "stored-app-key",
			},
		},
	}
	if err := writeTokens(authFile, tokens); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(authFile)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(envAuthFile, authFile)

	restoreOAuthCredentials(t)
	refreshOAuthToken = func(ctx context.Context, conf *oauth2.Config, token *oauth2.Token) (*oauth2.Token, error) {
		return nil, errors.New("refresh failed")
	}

	_, _, err = getAccessToken(tokenPersonal, "", false)
	if err == nil {
		t.Fatal("expected refresh failure")
	}
	if !strings.Contains(err.Error(), "dbxcli login") {
		t.Fatalf("expected login hint, got %q", err)
	}

	after, err := os.ReadFile(authFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("expected auth file to remain unchanged\nbefore: %s\nafter:  %s", before, after)
	}
}

func TestGetAccessTokenMissingTokenWithDefaultPersonalCredentialsReturnsLoginError(t *testing.T) {
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
		t.Fatal("app credential prompt should not run for command lazy auth")
		return appCredentials{}, nil
	}
	readAuthorizationCode = func() (string, error) {
		t.Fatal("authorization prompt should not run when app credentials are missing")
		return "", nil
	}
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string, verifier string) (*oauth2.Token, error) {
		t.Fatal("authorization exchange should not run when app credentials are missing")
		return nil, nil
	}

	_, _, err := getAccessToken(tokenPersonal, "", false)
	if err == nil {
		t.Fatal("expected missing credentials error")
	}
	if !strings.Contains(err.Error(), "dbxcli login") {
		t.Fatalf("expected login hint, got %q", err)
	}
	if !strings.Contains(err.Error(), envAccessToken) {
		t.Fatalf("expected %s hint, got %q", envAccessToken, err)
	}
}

func TestGetAccessTokenMissingTokenWithConfiguredAppKeyReturnsLoginError(t *testing.T) {
	restoreOAuthCredentials(t)
	setOAuthCredentials(tokenPersonal, "configured-key")

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
		t.Fatal("authorization prompt should not run for command lazy auth")
		return "", nil
	}
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string, verifier string) (*oauth2.Token, error) {
		t.Fatal("authorization exchange should not run for command lazy auth")
		return nil, nil
	}

	_, _, err := getAccessToken(tokenPersonal, "", false)
	if err == nil {
		t.Fatal("expected missing credentials error")
	}
	if !strings.Contains(err.Error(), "dbxcli login") {
		t.Fatalf("expected login hint, got %q", err)
	}
}

func TestLoginCommandForTokenType(t *testing.T) {
	tests := []struct {
		tokType string
		want    string
	}{
		{tokenPersonal, "dbxcli login"},
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
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string, verifier string) (*oauth2.Token, error) {
		return &oauth2.Token{}, nil
	}

	if _, err := requestAccessToken(tokenPersonal, ""); err == nil {
		t.Fatal("expected empty access token to return an error")
	}
}

func TestRequestAccessTokenRejectsMissingRefreshToken(t *testing.T) {
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
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string, verifier string) (*oauth2.Token, error) {
		return &oauth2.Token{AccessToken: "access-token"}, nil
	}

	if _, err := requestAccessToken(tokenPersonal, ""); err == nil {
		t.Fatal("expected missing refresh token to return an error")
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
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string, verifier string) (*oauth2.Token, error) {
		t.Fatal("authorization exchange should not run when reading code fails")
		return nil, nil
	}

	if _, err := requestAccessToken(tokenPersonal, ""); err == nil {
		t.Fatal("expected authorization code read error")
	}
}

func TestRequestAccessTokenPromptsForAppKeyWhenUsingBundledTeamDefaults(t *testing.T) {
	restoreOAuthCredentials(t)
	setOAuthCredentials(tokenTeamManage, defaultTeamManageAppKey)

	origReadAuthorizationCode := readAuthorizationCode
	origExchangeAuthorizationCode := exchangeAuthorizationCode
	t.Cleanup(func() {
		readAuthorizationCode = origReadAuthorizationCode
		exchangeAuthorizationCode = origExchangeAuthorizationCode
	})

	readAppCredentials = func(tokType string) (appCredentials, error) {
		if tokType != tokenTeamManage {
			t.Fatalf("expected team manage app credentials prompt, got %q", tokType)
		}
		return appCredentials{Key: "prompt-key"}, nil
	}
	readAuthorizationCode = func() (string, error) {
		return "auth-code", nil
	}
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string, verifier string) (*oauth2.Token, error) {
		if conf.ClientID != "prompt-key" {
			t.Fatalf("expected prompted app key, got %q", conf.ClientID)
		}
		if conf.ClientSecret != "" {
			t.Fatalf("expected no client secret for PKCE, got %q", conf.ClientSecret)
		}
		return &oauth2.Token{AccessToken: "access-token", RefreshToken: "refresh-token", TokenType: "Bearer", Expiry: time.Now().Add(time.Hour)}, nil
	}

	token, err := requestAccessToken(tokenTeamManage, "")
	if err != nil {
		t.Fatal(err)
	}
	if token != "access-token" {
		t.Fatalf("expected access token, got %q", token)
	}
	if teamManageAppKey != "prompt-key" {
		t.Fatalf("expected prompted app key to be saved for this process, got %q", teamManageAppKey)
	}
}

func TestRequestAccessTokenUsesDefaultPersonalAppKey(t *testing.T) {
	restoreOAuthCredentials(t)
	setOAuthCredentials(tokenPersonal, defaultPersonalAppKey)

	origReadAuthorizationCode := readAuthorizationCode
	origExchangeAuthorizationCode := exchangeAuthorizationCode
	t.Cleanup(func() {
		readAuthorizationCode = origReadAuthorizationCode
		exchangeAuthorizationCode = origExchangeAuthorizationCode
	})

	readAppCredentials = func(tokType string) (appCredentials, error) {
		t.Fatal("app credential prompt should not be used for the default personal app key")
		return appCredentials{}, nil
	}
	readAuthorizationCode = func() (string, error) {
		return "auth-code", nil
	}
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string, verifier string) (*oauth2.Token, error) {
		if conf.ClientID != defaultPersonalAppKey {
			t.Fatalf("expected default personal app key, got %q", conf.ClientID)
		}
		if conf.ClientSecret != "" {
			t.Fatalf("expected no client secret for PKCE, got %q", conf.ClientSecret)
		}
		return &oauth2.Token{AccessToken: "access-token", RefreshToken: "refresh-token", TokenType: "Bearer", Expiry: time.Now().Add(time.Hour)}, nil
	}

	if _, err := requestAccessToken(tokenPersonal, ""); err != nil {
		t.Fatal(err)
	}
}

func TestRequestAccessTokenUsesPKCEOfflineAuthURL(t *testing.T) {
	mockOAuthAppCredentials(t)

	origReadAuthorizationCode := readAuthorizationCode
	origExchangeAuthorizationCode := exchangeAuthorizationCode
	t.Cleanup(func() {
		readAuthorizationCode = origReadAuthorizationCode
		exchangeAuthorizationCode = origExchangeAuthorizationCode
	})

	const verifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	generateOAuthVerifier = func() string {
		return verifier
	}
	readAuthorizationCode = func() (string, error) {
		return "auth-code", nil
	}
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string, gotVerifier string) (*oauth2.Token, error) {
		if gotVerifier != verifier {
			t.Fatalf("expected verifier %q, got %q", verifier, gotVerifier)
		}
		return &oauth2.Token{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(time.Hour),
		}, nil
	}

	var err error
	out := captureStdout(t, func() {
		_, err = requestAccessToken(tokenPersonal, "")
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "token_access_type=offline") {
		t.Fatalf("expected offline token access type in auth URL, got %q", out)
	}
	if !strings.Contains(out, "code_challenge=E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM") {
		t.Fatalf("expected PKCE code challenge in auth URL, got %q", out)
	}
	if !strings.Contains(out, "code_challenge_method=S256") {
		t.Fatalf("expected PKCE S256 method in auth URL, got %q", out)
	}
}

func TestReadAppCredentialsReadsVisibleKey(t *testing.T) {
	restoreOAuthCredentials(t)

	var keyPrompts []string
	readAppKey = func(prompt string) (string, error) {
		keyPrompts = append(keyPrompts, prompt)
		return "visible-key", nil
	}

	creds, err := readAppCredentials(tokenPersonal)
	if err != nil {
		t.Fatal(err)
	}
	if creds.Key != "visible-key" {
		t.Fatalf("expected app key, got %q", creds.Key)
	}
	if len(keyPrompts) != 1 {
		t.Fatalf("expected one app key prompt, got %d", len(keyPrompts))
	}
	if keyPrompts[0] != "Dropbox app key: " {
		t.Fatalf("unexpected app key prompt %q", keyPrompts[0])
	}
}

func TestRequestAccessTokenUsesConfiguredAppCredentials(t *testing.T) {
	restoreOAuthCredentials(t)
	setOAuthCredentials(tokenPersonal, "configured-key")

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
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string, verifier string) (*oauth2.Token, error) {
		if conf.ClientID != "configured-key" {
			t.Fatalf("expected configured app key, got %q", conf.ClientID)
		}
		if conf.ClientSecret != "" {
			t.Fatalf("expected no client secret for PKCE, got %q", conf.ClientSecret)
		}
		return &oauth2.Token{AccessToken: "access-token", RefreshToken: "refresh-token", TokenType: "Bearer", Expiry: time.Now().Add(time.Hour)}, nil
	}

	if _, err := requestAccessToken(tokenPersonal, ""); err != nil {
		t.Fatal(err)
	}
}

func TestRequestAccessTokenRejectsEmptyAppCredentials(t *testing.T) {
	restoreOAuthCredentials(t)
	setOAuthCredentials(tokenTeamManage, defaultTeamManageAppKey)

	origReadAuthorizationCode := readAuthorizationCode
	origExchangeAuthorizationCode := exchangeAuthorizationCode
	t.Cleanup(func() {
		readAuthorizationCode = origReadAuthorizationCode
		exchangeAuthorizationCode = origExchangeAuthorizationCode
	})

	readAppCredentials = func(tokType string) (appCredentials, error) {
		return appCredentials{Key: " "}, nil
	}
	readAuthorizationCode = func() (string, error) {
		t.Fatal("authorization code prompt should not run when app credentials are invalid")
		return "", nil
	}
	exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string, verifier string) (*oauth2.Token, error) {
		t.Fatal("authorization exchange should not run when app credentials are invalid")
		return nil, nil
	}

	if _, err := requestAccessToken(tokenTeamManage, ""); err == nil {
		t.Fatal("expected empty app credentials to fail")
	}
}
