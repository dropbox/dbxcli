// Copyright © 2016 Dropbox, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/oauth2"
)

const (
	configFileName         = "auth.json"
	envAccessToken         = "DBXCLI_ACCESS_TOKEN"
	envAuthFile            = "DBXCLI_AUTH_FILE"
	tokenAccessTypeParam   = "token_access_type"
	tokenAccessTypeOffline = "offline"
	tokenRefreshWindow     = 5 * time.Minute
)

// TokenMap maps domains to a map of commands to saved credentials.
// For each domain, we want to save different tokens depending on the
// command type: personal, team access and team manage.
type TokenMap map[string]map[string]storedCredential

type storedCredential struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token,omitempty"`
	TokenType    string     `json:"token_type,omitempty"`
	Expiry       *time.Time `json:"expiry,omitempty"`
	AppKey       string     `json:"app_key,omitempty"`
}

type appCredentials struct {
	Key string
}

var readAppKey = func(prompt string) (string, error) {
	fmt.Print(prompt)
	var value string
	if _, err := fmt.Scan(&value); err != nil {
		return "", err
	}
	return value, nil
}

var readAppCredentials = func(tokType string) (appCredentials, error) {
	fmt.Printf("Enter Dropbox %s app key.\n", appCredentialsName(tokType))
	appKey, err := readAppKey("Dropbox app key: ")
	if err != nil {
		return appCredentials{}, err
	}
	return appCredentials{Key: appKey}, nil
}

var readAuthorizationCode = func() (string, error) {
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return "", err
	}
	return code, nil
}

var generateOAuthVerifier = oauth2.GenerateVerifier

var exchangeAuthorizationCode = func(ctx context.Context, conf *oauth2.Config, code string, verifier string) (*oauth2.Token, error) {
	return conf.Exchange(ctx, code, oauth2.VerifierOption(verifier))
}

var refreshOAuthToken = func(ctx context.Context, conf *oauth2.Config, token *oauth2.Token) (*oauth2.Token, error) {
	expired := *token
	expired.Expiry = time.Now().Add(-time.Second)
	return conf.TokenSource(ctx, &expired).Token()
}

func oauthConfig(tokenType string, domain string) *oauth2.Config {
	appKey := oauthCredentials(tokenType)
	return oauthConfigWithAppKey(appKey, domain)
}

func oauthConfigWithAppKey(appKey string, domain string) *oauth2.Config {
	endpoint := dropbox.OAuthEndpoint(domain)
	endpoint.AuthStyle = oauth2.AuthStyleInParams

	return &oauth2.Config{
		ClientID: appKey,
		Endpoint: endpoint,
	}
}

func (c *storedCredential) UnmarshalJSON(b []byte) error {
	var accessToken string
	if err := json.Unmarshal(b, &accessToken); err == nil {
		*c = storedCredential{AccessToken: accessToken}
		return nil
	}

	type storedCredentialAlias storedCredential
	var credential storedCredentialAlias
	if err := json.Unmarshal(b, &credential); err != nil {
		return err
	}
	*c = storedCredential(credential)
	return nil
}

func storedCredentialFromOAuthToken(token *oauth2.Token, appKey string) storedCredential {
	credential := storedCredential{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.Type(),
		AppKey:       appKey,
	}
	if !token.Expiry.IsZero() {
		expiry := token.Expiry
		credential.Expiry = &expiry
	}
	return credential
}

func (c *storedCredential) oauthToken() *oauth2.Token {
	token := &oauth2.Token{
		AccessToken:  c.AccessToken,
		TokenType:    c.TokenType,
		RefreshToken: c.RefreshToken,
	}
	if c.Expiry != nil {
		token.Expiry = *c.Expiry
	}
	return token
}

func (c *storedCredential) shouldRefresh(now time.Time) bool {
	if c.RefreshToken == "" {
		return false
	}
	if c.AccessToken == "" {
		return true
	}
	if c.Expiry == nil {
		return false
	}
	return !c.Expiry.After(now.Add(tokenRefreshWindow))
}

func authFilePath() (string, error) {
	if filePath := os.Getenv(envAuthFile); filePath != "" {
		return homedir.Expand(filePath)
	}

	dir, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".config", "dbxcli", configFileName), nil
}

func readTokens(filePath string) (TokenMap, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var tokens TokenMap
	if err := json.Unmarshal(b, &tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

func writeTokens(filePath string, tokens TokenMap) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(filePath), 0700); err != nil {
			return err
		}
	}

	b, err := json.Marshal(tokens)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, b, 0600)
}

func getAccessToken(tokType string, domain string, force bool) (string, string, error) {
	filePath, err := authFilePath()
	if err != nil {
		return "", "", err
	}

	tokenMap, err := readTokens(filePath)
	if err != nil && !os.IsNotExist(err) {
		return "", "", fmt.Errorf("read auth file %q: %w", filePath, err)
	}
	if tokenMap == nil {
		tokenMap = make(TokenMap)
	}
	if tokenMap[domain] == nil {
		tokenMap[domain] = make(map[string]storedCredential)
	}
	tokens := tokenMap[domain]
	credential := tokens[tokType]

	if force || (credential.AccessToken == "" && credential.RefreshToken == "") {
		if !force {
			return "", "", missingAccessTokenError(tokType)
		}
		credential, err = requestAccessCredential(tokType, domain)
		if err != nil {
			return "", "", err
		}
		tokens[tokType] = credential
		if err = writeTokens(filePath, tokenMap); err != nil {
			return "", "", err
		}
	}

	if !force && credential.shouldRefresh(time.Now()) {
		credential, err = refreshStoredCredential(tokType, domain, credential)
		if err != nil {
			details := authTokenDetails(tokType)
			if jsonErrorCode(err) == jsonErrorCodeAppKeyRequired {
				return "", "", appKeyRequiredErrorfWithDetails("refresh saved Dropbox credentials: %w; run %q again", details, err, loginCommand(tokType))
			}
			return "", "", authRefreshFailedErrorfWithDetails("refresh saved Dropbox credentials: %w; run %q again", details, err, loginCommand(tokType))
		}
		tokens[tokType] = credential
		if err = writeTokens(filePath, tokenMap); err != nil {
			return "", "", err
		}
	}

	return credential.AccessToken, filePath, nil
}

func loginCommand(tokType string) string {
	switch tokType {
	case tokenTeamAccess:
		return "dbxcli login team-access"
	case tokenTeamManage:
		return "dbxcli login team-manage"
	default:
		return "dbxcli login"
	}
}

func missingAccessTokenError(tokType string) error {
	return authRequiredErrorfWithDetails("no saved Dropbox credentials; run %q first or set %s", authTokenDetails(tokType), loginCommand(tokType), envAccessToken)
}

func authTokenDetails(tokType string) map[string]any {
	return map[string]any{
		"token_type":    authTokenTypeName(tokType),
		"login_command": loginCommand(tokType),
		"env_var":       envAccessToken,
	}
}

func authTokenTypeName(tokType string) string {
	switch tokType {
	case tokenTeamAccess:
		return "team-access"
	case tokenTeamManage:
		return "team-manage"
	default:
		return "personal"
	}
}

func appCredentialsName(tokType string) string {
	switch tokType {
	case tokenTeamAccess:
		return "team access"
	case tokenTeamManage:
		return "team manage"
	default:
		return "personal"
	}
}

func ensureOAuthAppCredentials(tokType string) error {
	if !needsOAuthCredentialsOverride(tokType) {
		return nil
	}

	creds, err := readAppCredentials(tokType)
	if err != nil {
		return err
	}
	creds.Key = strings.TrimSpace(creds.Key)
	if creds.Key == "" {
		return appKeyRequiredErrorWithDetails("Dropbox app key is required", map[string]any{
			"token_type": authTokenTypeName(tokType),
		})
	}

	setOAuthCredentials(tokType, creds.Key)
	return nil
}

func requestAccessToken(tokType string, domain string) (string, error) {
	credential, err := requestAccessCredential(tokType, domain)
	if err != nil {
		return "", err
	}
	return credential.AccessToken, nil
}

func requestAccessCredential(tokType string, domain string) (storedCredential, error) {
	if err := ensureOAuthAppCredentials(tokType); err != nil {
		return storedCredential{}, err
	}

	conf := oauthConfig(tokType, domain)
	verifier := generateOAuthVerifier()
	authCodeURL := conf.AuthCodeURL("state",
		oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam(tokenAccessTypeParam, tokenAccessTypeOffline),
	)

	fmt.Printf("1. Go to %v\n", authCodeURL)
	fmt.Printf("2. Click \"Allow\" (you might have to log in first).\n")
	fmt.Printf("3. Copy the authorization code.\n")
	fmt.Printf("Enter the authorization code here: ")

	code, err := readAuthorizationCode()
	if err != nil {
		return storedCredential{}, err
	}
	token, err := exchangeAuthorizationCode(context.Background(), conf, code, verifier)
	if err != nil {
		return storedCredential{}, authExchangeFailedErrorfWithDetails("exchange authorization code: %w", map[string]any{
			"token_type": authTokenTypeName(tokType),
		}, err)
	}
	if token == nil || token.AccessToken == "" {
		return storedCredential{}, authExchangeFailedErrorWithDetails("authorization did not return an access token", map[string]any{
			"token_type": authTokenTypeName(tokType),
		})
	}
	if token.RefreshToken == "" {
		return storedCredential{}, authExchangeFailedErrorWithDetails("authorization did not return a refresh token", map[string]any{
			"token_type": authTokenTypeName(tokType),
		})
	}
	return storedCredentialFromOAuthToken(token, conf.ClientID), nil
}

func refreshStoredCredential(tokType string, domain string, credential storedCredential) (storedCredential, error) {
	appKey := credential.AppKey
	if appKey == "" {
		appKey = oauthCredentials(tokType)
	}
	if strings.TrimSpace(appKey) == "" {
		return storedCredential{}, appKeyRequiredErrorWithDetails("saved credentials cannot be refreshed without a Dropbox app key", map[string]any{
			"token_type": authTokenTypeName(tokType),
		})
	}

	token, err := refreshOAuthToken(context.Background(), oauthConfigWithAppKey(appKey, domain), credential.oauthToken())
	if err != nil {
		return storedCredential{}, err
	}
	if token == nil || token.AccessToken == "" {
		return storedCredential{}, authRefreshFailedErrorfWithDetails("token refresh did not return an access token", map[string]any{
			"token_type": authTokenTypeName(tokType),
		})
	}

	refreshed := storedCredentialFromOAuthToken(token, appKey)
	if refreshed.RefreshToken == "" {
		refreshed.RefreshToken = credential.RefreshToken
	}
	if refreshed.TokenType == "" {
		refreshed.TokenType = credential.TokenType
	}
	if refreshed.TokenType == "" {
		refreshed.TokenType = "Bearer"
	}
	return refreshed, nil
}
