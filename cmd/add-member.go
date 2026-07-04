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
	"io"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/team"
	"github.com/spf13/cobra"
)

func addMember(cmd *cobra.Command, args []string) (err error) {
	if len(args) != 3 {
		return invalidArgumentsErrorWithDetails("`add-member` requires `email`, `first`, and `last` arguments", argumentsErrorDetails("email", "first", "last"))
	}
	dbx := teamNewFunc(config)

	email := args[0]
	firstName := args[1]
	lastName := args[2]
	member := team.NewMemberAddArg(email)
	member.MemberGivenName = firstName
	member.MemberSurname = lastName
	arg := team.NewMembersAddArg([]*team.MemberAddArg{member})
	res, err := dbx.MembersAddContext(currentContext(), arg)
	if err != nil {
		return withJSONErrorDetails(err, operationErrorDetails("team_add_member"), emailErrorDetails(email))
	}
	input := teamMemberAddInput{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}
	return commandOutput(cmd).Render(func(w io.Writer) error {
		return renderTeamMemberAdd(w, res)
	}, withJSONCommand(cmd, teamMemberAddOperationOutput(input, res)))
}

func renderTeamMemberAdd(out io.Writer, res *team.MembersAddLaunch) error {
	if res != nil && res.Tag == "complete" {
		_, err := fmt.Fprintln(out, "User successfully added to the team.")
		return err
	}
	return nil
}

// addMemberCmd represents the add-member command
var addMemberCmd = &cobra.Command{
	Use:   "add-member [flags] <email> <first-name> <last-name>",
	Short: "Add a new member to a team",
	RunE:  addMember,
}

func init() {
	teamCmd.AddCommand(addMemberCmd)
	enableStructuredOutput(addMemberCmd)
	setCommandDestructiveLevel(addMemberCmd, destructiveLevelAdmin)
}
