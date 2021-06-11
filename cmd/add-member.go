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
	"fmt"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/team"
	"github.com/spf13/cobra"
)

func addMember(cmd *cobra.Command, args []string) (err error) {
	if len(args) != 3 {
		return errors.New("`add-member` requires `email`, `first`, and `last` arguments")
	}
	dbx := team.New(config)

	email := args[0]
	firstName := args[1]
	lastName := args[2]
	member := team.NewMemberAddArg(email)
	member.MemberGivenName = firstName
	member.MemberSurname = lastName
	arg := team.NewMembersAddArg([]*team.MemberAddArg{member})
	res, err := dbx.MembersAdd(arg)
	if err != nil {
		return err
	}
	if res.Tag == "complete" {
		fmt.Printf("User successfully added to the team.\n")
	}
	return
}

// addMemberCmd represents the add-member command
var addMemberCmd = &cobra.Command{
	Use:   "add-member [flags] <email> <first-name> <last-name>",
	Short: "Add a new member to a team",
	RunE:  addMember,
}

func init() {
	teamCmd.AddCommand(addMemberCmd)
}
