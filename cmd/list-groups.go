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

	"github.com/dropbox/dropbox-sdk-go/team"
	"github.com/spf13/cobra"
)

func listGroups(cmd *cobra.Command, args []string) (err error) {
	arg := team.NewGroupsListArg()
	res, err := dbx.GroupsList(arg)
	if err != nil {
		return err
	}

	if len(res.Groups) == 0 {
		return
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 8, 1, ' ', 0)
	fmt.Fprintf(w, "Name\tId\t# Members\tExternal Id\n")
	for _, group := range res.Groups {
		fmt.Fprintf(w, group.GroupName, group.GroupId, group.MemberCount, group.GroupExternalId)
	}
	w.Flush()
	return
}

// listGroupsCmd represents the list-groups command
var listGroupsCmd = &cobra.Command{
	Use:   "list-groups",
	Short: "List groups",
	RunE:  listGroups,
}

func init() {
	teamCmd.AddCommand(listGroupsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listGroupsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listGroupsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
