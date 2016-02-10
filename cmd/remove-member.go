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

	"github.com/dropbox/dropbox-sdk-go/team"
	"github.com/spf13/cobra"
)

func removeMember(cmd *cobra.Command, args []string) (err error) {
	email := args[0]
	arg := team.NewMembersRemoveArg()
	arg.User = &team.UserSelectorArg{Tag: "email", Email: email}
	res, err := dbx.MembersRemove(arg)
	if err != nil {
		return err
	}
	if res.Tag == "complete" {
		fmt.Printf("User successfully removed from team.\n")
	}
	return
}

// removeMemberCmd represents the remove-member command
var removeMemberCmd = &cobra.Command{
	Use:   "remove-member",
	Short: "Remove member from a team",
	RunE:  removeMember,
}

func init() {
	teamCmd.AddCommand(removeMemberCmd)
}
