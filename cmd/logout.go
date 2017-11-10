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
	"os"
	"path"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/auth"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// Command logout revokes all saved API tokens and deletes auth.json.
func logout(cmd *cobra.Command, args []string) error {
	dir, err := homedir.Dir()
	if err != nil {
		return err
	}
	filePath := path.Join(dir, ".config", "dbxcli", configFileName)

	tokMap, err := readTokens(filePath)
	if err != nil {
		return err
	}

	for domain, tokens := range tokMap {
		for _, token := range tokens {
			config := dropbox.Config{token, dropbox.LogOff, nil, "", domain, nil, nil, nil}
			client := auth.New(config)
			client.TokenRevoke()
			if err != nil {
				return err
			}
		}
	}

	err = os.Remove(filePath)
	if err != nil {
		return err
	}

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
