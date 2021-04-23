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
	"fmt"
	"os"
	"path"
	"strings"

	"golang.org/x/oauth2"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

const (
	configFileName  = "dbxcli"
	tokenPersonal   = "personal"
	tokenTeamAccess = "teamAccess"
	tokenTeamManage = "teamManage"
)

var (
	personalAppKey      = "mvhz183vwqibe7q"
	personalAppSecret   = "q0kquhzgetjwcz1"
	teamAccessAppKey    = "zud1va492pnehkc"
	teamAccessAppSecret = "p3ginm1gy0kmj54"
	teamManageAppKey    = "xxe04eai4wmlitv"
	teamManageAppSecret = "t8ms714yun7nu5s"
)

// TokenMap maps profiles to domain map to a map of commands to tokens.
// For each profile and domain, we want to save different tokens depending on the
// command type: personal, team access and team manage
type TokenMap map[string]map[string]map[string]string

var config dropbox.Config

func oauthConfig(tokenType string, domain string) *oauth2.Config {
	var appKey, appSecret string
	switch tokenType {
	case "personal":
		appKey, appSecret = viper.GetString("app_key_personal"), viper.GetString("app_secret_personal")
	case "teamAccess":
		appKey, appSecret = viper.GetString("app_key_team_access"), viper.GetString("app_secret_team_access")
	case "teamManage":
		appKey, appSecret = viper.GetString("app_key_team_manage"), viper.GetString("app_secret_team_manage")
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

func readTokens() (TokenMap, error) {
	var tokens TokenMap
	if err := viper.UnmarshalKey("tokens", &tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

func writeTokens(tokens TokenMap) {
	viper.Set("tokens", tokens)
	if err := viper.WriteConfig(); err != nil {
		return
	}
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

func initDbx(cmd *cobra.Command, args []string) (err error) {
	verbose, _ := cmd.Flags().GetBool("verbose")
	asMember, _ := cmd.Flags().GetString("as-member")

	tokType := tokenType(cmd)
	conf := oauthConfig(tokType, viper.GetString("domain"))

	tokenMap, err := readTokens()
	if tokenMap == nil {
		tokenMap = make(TokenMap)
	}
	if tokenMap[viper.GetString("profile")] == nil {
		tokenMap[viper.GetString("profile")] = make(map[string]map[string]string)
	}
	profileTokens := tokenMap[viper.GetString("profile")]
	if profileTokens[viper.GetString("domain")] == nil {
		profileTokens[viper.GetString("domain")] = make(map[string]string)
	}
	tokens := profileTokens[viper.GetString("domain")]

	if err != nil || tokens[tokType] == "" {
		fmt.Printf("1. Go to %v\n", conf.AuthCodeURL("state"))
		fmt.Printf("2. Click \"Allow\" (you might have to log in first).\n")
		fmt.Printf("3. Copy the authorization code.\n")
		fmt.Printf("Enter the authorization code here: ")

		var code string
		if _, err = fmt.Scan(&code); err != nil {
			return
		}
		var token *oauth2.Token
		ctx := context.Background()
		token, err = conf.Exchange(ctx, code)
		if err != nil {
			return
		}
		tokens[tokType] = token.AccessToken
		writeTokens(tokenMap)
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
		Domain:          viper.GetString("domain"),
		Client:          nil,
		HeaderGenerator: nil,
		URLGenerator:    nil,
	}

	return
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

	dir, err := homedir.Dir()
	if err != nil {
		return
	}
	viper.SetConfigName(configFileName)
	viper.SetConfigType("json")

	// default configuration path
	viper.AddConfigPath(path.Join(dir, ".config"))
	// short configuration path (useful for docker and testing)
	viper.AddConfigPath("/config/")
	// super useful for testing
	viper.AddConfigPath(".")

	viper.SetDefault("app_key_personal", personalAppKey)
	viper.SetDefault("app_secret_personal", personalAppSecret)
	viper.SetDefault("app_key_team_access", teamAccessAppKey)
	viper.SetDefault("app_secret_team_access", teamAccessAppSecret)
	viper.SetDefault("app_key_team_manage", teamManageAppKey)
	viper.SetDefault("app_secret_team_manage", teamManageAppSecret)
	viper.SetDefault("domain", "")

	viper.BindEnv("app_key_personal", "DROPBOX_PERSONAL_APP_KEY")
	viper.BindEnv("app_secret_personal", "DROPBOX_PERSONAL_APP_SECRET")
	viper.BindEnv("app_key_team_access", "DROPBOX_TEAM_APP_KEY")
	viper.BindEnv("app_secret_team_access", "DROPBOX_TEAM_APP_SECRET")
	viper.BindEnv("app_key_team_manage", "DROPBOX_MANAGE_APP_KEY")
	viper.BindEnv("app_secret_team_manage", "DROPBOX_MANAGE_APP_SECRET")
	viper.BindEnv("domain", "DROPBOX_DOMAIN")

	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
	RootCmd.PersistentFlags().String("as-member", "", "Member ID to perform action as")
	RootCmd.PersistentFlags().String("profile", "default", "Set the configuration profile [for tokens]")
	// This flag should only be used for testing. Marked hidden so it doesn't clutter usage etc.
	RootCmd.PersistentFlags().String("domain", "", "Override default Dropbox domain, useful for testing")
	RootCmd.PersistentFlags().MarkHidden("domain")

	viper.BindPFlag("domain", RootCmd.PersistentFlags().Lookup("domain"))
	viper.BindPFlag("profile", RootCmd.PersistentFlags().Lookup("profile"))

	viper.ReadInConfig()

	if err := viper.SafeWriteConfig(); err != nil {
		return
	}

}
