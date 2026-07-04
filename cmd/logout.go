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
	"errors"
	"os"
	"sort"

	"github.com/dropbox/dbxcli/v3/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/auth"
	"github.com/spf13/cobra"
)

const (
	logoutStatusLoggedOut        = "logged_out"
	logoutStatusAlreadyLoggedOut = "already_logged_out"
	logoutKindAuth               = "auth"
)

type logoutResult struct {
	RemovedSavedCredentials bool `json:"removed_saved_credentials"`
	RemoteTokenRevoked      bool `json:"remote_token_revoked"`
}

var revokeAccessToken = func(domain string, token string) error {
	cfg := dropbox.Config{
		Token:           token,
		LogLevel:        dropbox.LogOff,
		Logger:          nil,
		AsMemberID:      "",
		Domain:          domain,
		Client:          nil,
		HeaderGenerator: nil,
		URLGenerator:    nil,
	}
	client := auth.NewContext(cfg)
	return client.TokenRevokeContext(currentContext())
}

// Command logout revokes all saved API tokens and deletes auth.json.
func logout(cmd *cobra.Command, args []string) error {
	if os.Getenv(envAccessToken) != "" {
		return newCodedError(jsonErrorCodeEnvTokenStillActive, errors.New("DBXCLI_ACCESS_TOKEN is set; unset it before running logout"))
	}

	filePath, err := authFilePath()
	if err != nil {
		return err
	}

	tokMap, err := readTokens(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return renderLogoutResult(cmd, logoutStatusAlreadyLoggedOut, false, false, nil)
		}
		return err
	}

	tokenCount := 0
	revokeFailed := false
	for _, domain := range sortedTokenDomains(tokMap) {
		tokens := tokMap[domain]
		for _, tokenType := range sortedTokenTypes(tokens) {
			token := tokens[tokenType]
			if token.AccessToken == "" {
				continue
			}
			tokenCount++
			if err = revokeAccessToken(domain, token.AccessToken); err != nil {
				revokeFailed = true
				if commandOutputFormat(cmd) == output.FormatText {
					commandOutput(cmd).Warn("could not revoke token (may be expired): %v", err)
				}
			}
		}
	}

	if err := os.Remove(filePath); err != nil {
		return err
	}

	warnings := []jsonWarning(nil)
	if revokeFailed {
		warnings = append(warnings, jsonWarning{
			Code:    jsonWarningCodeTokenRevokeFailed,
			Message: "Saved credentials were removed locally, but one or more Dropbox tokens could not be revoked remotely.",
		})
	}
	return renderLogoutResult(cmd, logoutStatusLoggedOut, true, tokenCount > 0 && !revokeFailed, warnings)
}

func renderLogoutResult(cmd *cobra.Command, status string, removedSavedCredentials bool, remoteTokenRevoked bool, warnings []jsonWarning) error {
	return renderJSONOperationOutputWithWarnings(cmd, nil, []jsonOperationResult{
		newJSONOperationResult(status, logoutKindAuth, nil, logoutResult{
			RemovedSavedCredentials: removedSavedCredentials,
			RemoteTokenRevoked:      remoteTokenRevoked,
		}),
	}, warnings)
}

func sortedTokenDomains(tokMap TokenMap) []string {
	domains := make([]string, 0, len(tokMap))
	for domain := range tokMap {
		domains = append(domains, domain)
	}
	sort.Strings(domains)
	return domains
}

func sortedTokenTypes(tokens map[string]storedCredential) []string {
	tokenTypes := make([]string, 0, len(tokens))
	for tokenType := range tokens {
		tokenTypes = append(tokenTypes, tokenType)
	}
	sort.Strings(tokenTypes)
	return tokenTypes
}

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout [flags]",
	Short: "Log out of the current session",
	Long: `Log out of the current session.

Logout revokes saved Dropbox access tokens by default and removes local saved
credentials. If DBXCLI_ACCESS_TOKEN is set, unset it before running logout;
environment-provided tokens are not saved locally and cannot be removed by
dbxcli.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := initCommandContext(cmd); err != nil {
			return err
		}
		return validateOutputFormat(cmd)
	},
	RunE: logout,
}

func init() {
	RootCmd.AddCommand(logoutCmd)
	enableStructuredOutput(logoutCmd)
}
