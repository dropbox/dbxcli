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
	"strings"

	"github.com/spf13/cobra"
)

func loginTokenType(name string) (string, error) {
	switch name {
	case "", tokenPersonal:
		return tokenPersonal, nil
	case tokenTeamAccess, "team-access", "team_access":
		return tokenTeamAccess, nil
	case tokenTeamManage, "team-manage", "team_manage":
		return tokenTeamManage, nil
	default:
		return "", fmt.Errorf("unknown login token type %q; expected personal, team-access, or team-manage", name)
	}
}

func loginAppKeyFromFlag(cmd *cobra.Command, tokType string) error {
	appKey, _ := cmd.Flags().GetString("app-key")
	appKey = strings.TrimSpace(appKey)
	if appKey == "" {
		return nil
	}

	setOAuthCredentials(tokType, appKey)
	return nil
}

func login(cmd *cobra.Command, args []string) error {
	domain, _ := cmd.Flags().GetString("domain")

	tokenName := ""
	if len(args) > 0 {
		tokenName = args[0]
	}
	tokType, err := loginTokenType(tokenName)
	if err != nil {
		return err
	}
	if err := loginAppKeyFromFlag(cmd, tokType); err != nil {
		return err
	}

	_, filePath, err := getAccessToken(tokType, domain, true)
	if err != nil {
		return err
	}

	commandOutput(cmd).Info("Credentials saved to %s", filePath)
	return nil
}

var loginCmd = &cobra.Command{
	Use:   "login [personal|team-access|team-manage]",
	Short: "Log in and save Dropbox credentials",
	Long: `Log in and save Dropbox credentials.

By default, login stores credentials for regular Dropbox user commands.
Use "team-access" for --as-member commands or "team-manage" for team commands.`,
	Args: cobra.MaximumNArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := initCommandContext(cmd); err != nil {
			return err
		}
		return validateOutputFormat(cmd)
	},
	RunE: login,
}

func init() {
	loginCmd.Flags().String("app-key", "", "Dropbox app key to use for this login")
	RootCmd.AddCommand(loginCmd)
}
