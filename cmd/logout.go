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

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/auth"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Command logout revokes all saved API tokens and deletes auth.json.
func logout(cmd *cobra.Command, args []string) error {

	tokMap, err := readTokens()
	if err != nil {
		return err
	}
	profileTokMap := tokMap[viper.GetString("profile")]
	if profileTokMap == nil {
		return fmt.Errorf("cannot find profile")
	}

	for domain, tokens := range profileTokMap {
		for _, token := range tokens {
			config := dropbox.Config{
				Token:           token,
				LogLevel:        dropbox.LogOff,
				Logger:          nil,
				AsMemberID:      "",
				Domain:          domain,
				Client:          nil,
				HeaderGenerator: nil,
				URLGenerator:    nil,
			}
			client := auth.New(config)
			err = client.TokenRevoke()
			if err != nil {
				return err
			}
		}
	}

	delete(tokMap, viper.GetString("profile"))
	writeTokens(tokMap)

	return nil
}

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout [flags]",
	Short: "Log out of the current session",
	RunE:  logout,
}

func init() {
	RootCmd.AddCommand(logoutCmd)
}
