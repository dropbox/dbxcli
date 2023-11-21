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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

const (
	tokenPersonal   = "personal"
	tokenTeamAccess = "teamAccess"
	tokenTeamManage = "teamManage"
)

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

var (
	personalAppKey      = "mvhz183vwqibe7q"
	personalAppSecret   = "q0kquhzgetjwcz1"
	teamAccessAppKey    = "zud1va492pnehkc"
	teamAccessAppSecret = "p3ginm1gy0kmj54"
	teamManageAppKey    = "xxe04eai4wmlitv"
	teamManageAppSecret = "t8ms714yun7nu5s"
)

// TokenMap maps domains to a map of commands to tokens.
// For each domain, we want to save different tokens depending on the
// command type: personal, team access and team manage
type TokenMap map[string]map[string]string

var config dropbox.Config

func oauthConfig(tokenType string, domain string) *oauth2.Config {
	var appKey, appSecret string
	switch tokenType {
	case "personal":
		appKey, appSecret = personalAppKey, personalAppSecret
	case "teamAccess":
		appKey, appSecret = teamAccessAppKey, teamAccessAppSecret
	case "teamManage":
		appKey, appSecret = teamManageAppKey, teamManageAppSecret
	}
	return &oauth2.Config{
		ClientID:     appKey,
		ClientSecret: appSecret,
		Endpoint:     dropbox.OAuthEndpoint(domain),
	}
}

func validatePath(p string) (path string, err error) {
	path = p

	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}

	path = strings.TrimSuffix(path, "/")

	return
}

func makeRelocationArg(s string, d string) (arg *files.RelocationArg, err error) {
	src, err := validatePath(s)
	if err != nil {
		return
	}
	dst, err := validatePath(d)
	if err != nil {
		return
	}

	arg = files.NewRelocationArg(src, dst)

	return
}

func writeTokens(filePath string, tokens TokenMap) error {
	// Ensure config directory exists.
	if err := os.MkdirAll(filepath.Dir(filePath), 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	b, err := json.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("encode tokens: %w", err)
	}

	if err = os.WriteFile(filePath, b, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func tokenType(cmd *cobra.Command) string {
	if cmd.Parent().Name() == "team" {
		return tokenTeamManage
	}
	if asMember, _ := cmd.Flags().GetString("as-member"); asMember != "" {
		return tokenTeamAccess
	}
	return tokenPersonal
}

func initDbx(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	asMember, _ := cmd.Flags().GetString("as-member")
	domain, _ := cmd.Flags().GetString("domain")

	tokType := tokenType(cmd)
	conf := oauthConfig(tokType, domain)

	filePath, err := configFile()
	if err != nil {
		return fmt.Errorf("config file: %w", err)
	}

	tokenMap, err := readTokens(filePath)
	if err != nil {
		return fmt.Errorf("read tokens: %w", err)
	}

	if tokenMap[domain] == nil {
		tokenMap[domain] = make(map[string]string)
	}

	tokens := tokenMap[domain]
	if tokens[tokType] == "" {
		fmt.Printf("1. Go to %v\n", conf.AuthCodeURL("state"))
		fmt.Printf("2. Click \"Allow\" (you might have to log in first).\n")
		fmt.Printf("3. Copy the authorization code.\n")
		fmt.Printf("Enter the authorization code here: ")

		var code string
		if _, err = fmt.Scan(&code); err != nil {
			return fmt.Errorf("read code: %w", err)
		}

		var token *oauth2.Token
		ctx := context.Background()
		token, err = conf.Exchange(ctx, code)
		if err != nil {
			return fmt.Errorf("token exchange: %w", err)
		}

		tokens[tokType] = token.AccessToken
		if err := writeTokens(filePath, tokenMap); err != nil {
			return err
		}
	}

	logLevel := dropbox.LogOff
	if verbose {
		logLevel = dropbox.LogInfo
	}
	config = dropbox.Config{
		Token:           tokens[tokType],
		LogLevel:        logLevel,
		Logger:          nil,
		AsMemberID:      asMember,
		Domain:          domain,
		Client:          nil,
		HeaderGenerator: nil,
		URLGenerator:    nil,
	}

	return nil
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dbxcli",
	Short: "A command line tool for Dropbox users and team admins",
	Long: `Use dbxcli to quickly interact with your Dropbox, upload/download files,
manage your team and more. It is easy, scriptable and works on all platforms!`,
	SilenceUsage:      true,
	PersistentPreRunE: initDbx,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
	RootCmd.PersistentFlags().String("as-member", "", "Member ID to perform action as")
	// This flag should only be used for testing. Marked hidden so it doesn't clutter usage etc.
	RootCmd.PersistentFlags().String("domain", "", "Override default Dropbox domain, useful for testing")
	RootCmd.PersistentFlags().MarkHidden("domain")

	personalAppKey = getEnv(personalAppKeyEnv, personalAppKey)
	personalAppSecret = getEnv(personalAppSecretEnv, personalAppSecret)
	teamAccessAppKey = getEnv(teamAccessAppKeyEnv, teamAccessAppKey)
	teamAccessAppSecret = getEnv(teamAccessAppSecretEnv, teamAccessAppSecret)
	teamManageAppKey = getEnv(teamManageAppKeyEnv, teamManageAppKey)
	teamManageAppSecret = getEnv(teamManageAppSecretEnv, teamAccessAppSecret)
}
