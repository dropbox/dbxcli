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
	"text/tabwriter"

	"github.com/dropbox/dropbox-sdk-go-unofficial/team"
	"github.com/spf13/cobra"
)

func listMembers(cmd *cobra.Command, args []string) (err error) {
	arg := team.NewMembersListArg()
	res, err := dbx.MembersList(arg)
	if err != nil {
		return err
	}

	if len(res.Members) == 0 {
		return
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 8, 1, ' ', 0)
	fmtStr := "%s\t%s\t%s\t%s\t%s\n"
	fmt.Fprintf(w, fmtStr, "Name", "Id", "Status", "Email", "Role")
	for _, member := range res.Members {
		fmt.Fprintf(w, fmtStr,
			member.Profile.Name.DisplayName,
			member.Profile.TeamMemberId,
			member.Profile.Status.Tag,
			member.Profile.Email,
			member.Role.Tag)
	}
	w.Flush()
	return
}

// listMembersCmd represents the list-members command
var listMembersCmd = &cobra.Command{
	Use:   "list-members",
	Short: "List team members",
	RunE:  listMembers,
}

func init() {
	teamCmd.AddCommand(listMembersCmd)
}
