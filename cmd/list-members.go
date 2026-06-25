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
	"io"
	"text/tabwriter"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/team"
	"github.com/spf13/cobra"
)

func listMembers(cmd *cobra.Command, args []string) (err error) {
	dbx := teamNewFunc(config)
	arg := team.NewMembersListArg()
	members, err := listTeamMembers(dbx, arg)
	if err != nil {
		return err
	}

	commandVerboseStatus(cmd, "Listed %d team members", len(members))

	return commandOutput(cmd).Render(func(w io.Writer) error {
		return renderTeamMembers(w, members)
	}, newJSONCommandOperationOutput(cmd, teamInfoInput{}, teamMemberOperationResults(members), nil))
}

func listTeamMembers(dbx teamClient, arg *team.MembersListArg) ([]*team.TeamMemberInfo, error) {
	var members []*team.TeamMemberInfo
	res, err := dbx.MembersList(arg)
	if err != nil {
		return nil, err
	}
	members = append(members, res.Members...)

	for res.HasMore {
		if res.Cursor == "" {
			return nil, errors.New("team member list has more results but no cursor")
		}
		res, err = dbx.MembersListContinue(team.NewMembersListContinueArg(res.Cursor))
		if err != nil {
			return nil, err
		}
		members = append(members, res.Members...)
	}
	return members, nil
}

func renderTeamMembers(out io.Writer, members []*team.TeamMemberInfo) error {
	if len(members) == 0 {
		return nil
	}

	w := new(tabwriter.Writer)
	w.Init(out, 4, 8, 1, ' ', 0)
	fmtStr := "%s\t%s\t%s\t%s\t%s\n"
	fmt.Fprintf(w, fmtStr, "Name", "Id", "Status", "Email", "Role")
	for _, member := range members {
		fmt.Fprintf(w, fmtStr,
			member.Profile.Name.DisplayName,
			member.Profile.TeamMemberId,
			member.Profile.Status.Tag,
			member.Profile.Email,
			member.Role.Tag)
	}
	return w.Flush()
}

// listMembersCmd represents the list-members command
var listMembersCmd = &cobra.Command{
	Use:   "list-members",
	Short: "List team members",
	RunE:  listMembers,
}

func init() {
	teamCmd.AddCommand(listMembersCmd)
	enableStructuredOutput(listMembersCmd)
}
