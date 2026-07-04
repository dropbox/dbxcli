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
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/common"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/users"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
var commandContext context.Context = context.Background()
var commandContextCancel context.CancelFunc

func currentContext() context.Context {
	if commandContext == nil {
		return context.Background()
	}
	return commandContext
}

func initCommandContext(cmd *cobra.Command) error {
	finishCommandContext()

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	timeout, err := commandTimeout(cmd)
	if err != nil {
		return err
	}
	if timeout < 0 {
		return invalidArgumentsErrorWithDetails("`--timeout` must be greater than or equal to 0", flagErrorDetails("timeout"))
	}
	if timeout > 0 {
		ctx, commandContextCancel = context.WithTimeout(ctx, timeout)
		cmd.SetContext(ctx)
	}
	commandContext = ctx
	return nil
}

func commandTimeout(cmd *cobra.Command) (time.Duration, error) {
	for _, flags := range []interface {
		Lookup(string) *pflag.Flag
		GetDuration(string) (time.Duration, error)
	}{
		cmd.Flags(),
		cmd.InheritedFlags(),
		cmd.PersistentFlags(),
	} {
		if flags.Lookup("timeout") != nil {
			return flags.GetDuration("timeout")
		}
	}
	return 0, nil
}

func finishCommandContext() {
	if commandContextCancel != nil {
		commandContextCancel()
		commandContextCancel = nil
	}
	commandContext = context.Background()
}

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
	currentAuthContext = nil

	if err := initCommandContext(cmd); err != nil {
		return err
	}

	if commandIsJSONHelp(cmd) {
		return nil
	}
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
		currentAuthContext = &authContext{
			Source:      authSourceEnv,
			Refreshable: false,
			AuthFile:    authFileNone,
		}
		config = withRootNamespace(config, tokenType(cmd))
		return nil
	}

	tokType := tokenType(cmd)
	credential, _, err := getAccessCredential(tokType, domain, false)
	if err != nil {
		return err
	}

	config = makeDropboxConfig(credential.AccessToken, verbose, asMember, domain)
	currentAuthContext = &authContext{
		Source:      authSourceSaved,
		Refreshable: credential.RefreshToken != "",
		AuthFile:    authFileKind(),
	}
	config = withRootNamespace(config, tokType)

	return
}

func withRootNamespace(cfg dropbox.Config, tokType string) dropbox.Config {
	// Team manage tokens are for administrative operations and don't need a path root.
	if tokType == tokenTeamManage {
		return cfg
	}

	account, err := usersNewFunc(cfg).GetCurrentAccountContext(currentContext())
	if err != nil {
		cfg.LogInfo("Warning: could not auto-detect root namespace (%v); team folders may not be accessible", err)
		return cfg
	}

	rootNamespaceID := rootNamespaceID(account)
	if rootNamespaceID == "" {
		return cfg
	}
	return cfg.WithRoot(rootNamespaceID)
}

func rootNamespaceID(account *users.FullAccount) string {
	if account == nil {
		return ""
	}

	switch ri := account.RootInfo.(type) {
	case *common.TeamRootInfo:
		return ri.RootNamespaceId
	case *common.UserRootInfo:
		return ri.RootNamespaceId
	default:
		return ""
	}
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dbxcli",
	Short: "Scriptable Dropbox CLI for files, shared links, teams, and automation",
	Long: `Use dbxcli to work with Dropbox files, folders, shared links, and team
workflows from a terminal. It supports text output for humans, structured JSON
output for automation, pipe-friendly transfers, refreshable OAuth login, and
direct-token automation for scripts, CI jobs, and agent-style workflows.`,
	SilenceUsage:      true,
	SilenceErrors:     true,
	PersistentPreRunE: initDbx,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	args := os.Args[1:]
	jsonErrorOutput := outputJSONRequested(args)
	if err := rawArgsHelpOutputFormatError(args); err != nil {
		renderCommandErrorWithJSON(RootCmd, err, jsonErrorOutput)
		os.Exit(exitCodeForError(err))
	}

	restoreDeprecatedCommands := func() {}
	if rawArgsRequestJSONHelp(args) {
		restoreDeprecatedCommands = temporarilyClearDeprecatedCommands(RootCmd)
	}
	defer restoreDeprecatedCommands()
	defer finishCommandContext()

	cmd, err := RootCmd.ExecuteC()
	if err != nil {
		renderCommandErrorWithJSON(cmd, err, jsonErrorOutput)
		os.Exit(exitCodeForError(err))
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
	RootCmd.PersistentFlags().Duration("timeout", 0, "Timeout for Dropbox network operations (0 disables; examples: 30s, 2m, 1h)")
	RootCmd.PersistentFlags().String("as-member", "", "Member ID to perform action as")
	// This flag should only be used for testing. Marked hidden so it doesn't clutter usage etc.
	RootCmd.PersistentFlags().String("domain", "", "Override default Dropbox domain, useful for testing")
	_ = RootCmd.PersistentFlags().MarkHidden("domain")

	loadOAuthCredentialsFromEnv()
	installJSONHelp(RootCmd)
}
