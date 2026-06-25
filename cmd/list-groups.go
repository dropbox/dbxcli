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
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/team_common"
	"github.com/spf13/cobra"
)

func listGroups(cmd *cobra.Command, args []string) (err error) {
	dbx := teamNewFunc(config)
	arg := team.NewGroupsListArg()
	groups, err := listTeamGroups(dbx, arg)
	if err != nil {
		return err
	}

	commandVerboseStatus(cmd, "Listed %d team groups", len(groups))

	return commandOutput(cmd).Render(func(w io.Writer) error {
		return renderTeamGroups(w, groups)
	}, newJSONOperationOutput(teamInfoInput{}, teamGroupOperationResults(groups), nil))
}

func listTeamGroups(dbx teamClient, arg *team.GroupsListArg) ([]*team_common.GroupSummary, error) {
	var groups []*team_common.GroupSummary
	res, err := dbx.GroupsList(arg)
	if err != nil {
		return nil, err
	}
	groups = append(groups, res.Groups...)

	for res.HasMore {
		if res.Cursor == "" {
			return nil, errors.New("team group list has more results but no cursor")
		}
		res, err = dbx.GroupsListContinue(team.NewGroupsListContinueArg(res.Cursor))
		if err != nil {
			return nil, err
		}
		groups = append(groups, res.Groups...)
	}
	return groups, nil
}

func renderTeamGroups(out io.Writer, groups []*team_common.GroupSummary) error {
	if len(groups) == 0 {
		return nil
	}

	w := new(tabwriter.Writer)
	w.Init(out, 4, 8, 1, ' ', 0)
	fmt.Fprintf(w, "Name\tId\t# Members\tExternal Id\n")
	for _, group := range groups {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", group.GroupName, group.GroupId, group.MemberCount, group.GroupExternalId)
	}
	return w.Flush()
}

// listGroupsCmd represents the list-groups command
var listGroupsCmd = &cobra.Command{
	Use:   "list-groups",
	Short: "List groups",
	RunE:  listGroups,
}

func init() {
	teamCmd.AddCommand(listGroupsCmd)
	enableStructuredOutput(listGroupsCmd)
}
