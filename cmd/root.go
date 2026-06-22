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
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

const (
	tokenPersonal   = "personal"
	tokenTeamAccess = "teamAccess"
	tokenTeamManage = "teamManage"

	defaultPersonalAppKey   = "07o23gulcj8qi69"
	defaultTeamAccessAppKey = "qyy1w4mbkj2wpiv"
	defaultTeamManageAppKey = "sa9pv32eixm1i3p"
)

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

var (
	personalAppKey   = defaultPersonalAppKey
	teamAccessAppKey = defaultTeamAccessAppKey
	teamManageAppKey = defaultTeamManageAppKey
)

var config dropbox.Config

func commandSkipsAuth(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		switch c.Name() {
		case "__complete", "__completeNoDesc", "completion", "help", "version":
			return true
		}
	}
	return false
}

func oauthCredentials(tokenType string) string {
	switch tokenType {
	case tokenPersonal:
		return personalAppKey
	case tokenTeamAccess:
		return teamAccessAppKey
	case tokenTeamManage:
		return teamManageAppKey
	default:
		return ""
	}
}

func setOAuthCredentials(tokenType string, appKey string) {
	switch tokenType {
	case tokenPersonal:
		personalAppKey = appKey
	case tokenTeamAccess:
		teamAccessAppKey = appKey
	case tokenTeamManage:
		teamManageAppKey = appKey
	}
}

func needsOAuthCredentialsOverride(tokenType string) bool {
	return oauthCredentials(tokenType) == ""
}

func validatePath(p string) (path string, err error) {
	path = p
	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}

	path = cleanDropboxPath(path)

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

func tokenType(cmd *cobra.Command) string {
	if cmd.Parent().Name() == "team" {
		return tokenTeamManage
	}
	if asMember, _ := cmd.Flags().GetString("as-member"); asMember != "" {
		return tokenTeamAccess
	}
	return tokenPersonal
}

func makeDropboxConfig(token string, verbose bool, asMember string, domain string) dropbox.Config {
	logLevel := dropbox.LogOff
	if verbose {
		logLevel = dropbox.LogInfo
	}

	return dropbox.Config{
		Token:           token,
		LogLevel:        logLevel,
		Logger:          nil,
		AsMemberID:      asMember,
		Domain:          domain,
		Client:          nil,
		HeaderGenerator: nil,
		URLGenerator:    nil,
	}
}

func initDbx(cmd *cobra.Command, args []string) (err error) {
	if err := validateOutputFormat(cmd); err != nil {
		return err
	}

	if commandSkipsAuth(cmd) {
		return nil
	}

	verbose, _ := cmd.Flags().GetBool("verbose")
	asMember, _ := cmd.Flags().GetString("as-member")
	domain, _ := cmd.Flags().GetString("domain")

	if accessToken := os.Getenv(envAccessToken); accessToken != "" {
		config = makeDropboxConfig(accessToken, verbose, asMember, domain)
		return nil
	}

	tokType := tokenType(cmd)
	accessToken, _, err := getAccessToken(tokType, domain, false)
	if err != nil {
		return err
	}

	config = makeDropboxConfig(accessToken, verbose, asMember, domain)

	return
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dbxcli",
	Short: "A command line tool for Dropbox users and team admins",
	Long: `Use dbxcli to quickly interact with your Dropbox, upload/download files,
manage your team and more. It is easy, scriptable and works on all platforms!`,
	SilenceUsage:      true,
	SilenceErrors:     true,
	PersistentPreRunE: initDbx,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	jsonErrorOutput := outputJSONRequested(os.Args[1:])
	cmd, err := RootCmd.ExecuteC()
	if err != nil {
		renderCommandErrorWithJSON(cmd, err, jsonErrorOutput)
		os.Exit(1)
	}
}

func loadOAuthCredentialsFromEnv() {
	personalAppKey = getEnv("DROPBOX_PERSONAL_APP_KEY", personalAppKey)
	teamAccessAppKey = getEnv("DROPBOX_TEAM_APP_KEY", teamAccessAppKey)
	teamManageAppKey = getEnv("DROPBOX_MANAGE_APP_KEY", teamManageAppKey)
}

func init() {
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
	RootCmd.PersistentFlags().String(outputFlag, "text", "Output format: text, json")
	RootCmd.PersistentFlags().String("as-member", "", "Member ID to perform action as")
	// This flag should only be used for testing. Marked hidden so it doesn't clutter usage etc.
	RootCmd.PersistentFlags().String("domain", "", "Override default Dropbox domain, useful for testing")
	_ = RootCmd.PersistentFlags().MarkHidden("domain")

	loadOAuthCredentialsFromEnv()
}
