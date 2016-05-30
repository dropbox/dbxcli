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

	"github.com/spf13/cobra"
)

func account(cmd *cobra.Command, args []string) error {
	res, err := dbx.GetCurrentAccount()
	if err != nil {
		return err
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 8, 1, ' ', 0)

	fmt.Fprintf(w, "Logged in as %s <%s>\n\n", res.Name.DisplayName, res.Email)
	fmt.Fprintf(w, "Account Id:\t%s\n", res.AccountId)
	fmt.Fprintf(w, "Account Type:\t%s\n", res.AccountType.Tag)
	fmt.Fprintf(w, "Locale:\t%s\n", res.Locale)
	fmt.Fprintf(w, "Referral Link:\t%s\n", res.ReferralLink)
	fmt.Fprintf(w, "Paired Account:\t%t\n", res.IsPaired)
	if res.Team != nil {
		fmt.Fprintf(w, "Team:\n  Name:\t%s\n  Id:\t%s\n  Member Id:\t%s\n", res.Team.Name, res.Team.Id, res.TeamMemberId)
	}

	w.Flush()

	return nil
}

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Display information about the current account",
	RunE:  account,
}

func init() {
	RootCmd.AddCommand(accountCmd)
}
