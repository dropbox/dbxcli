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
	"os"
	"text/tabwriter"

	"github.com/dropbox/dropbox-sdk-go-unofficial/users"
	"github.com/spf13/cobra"
)

// printFullAccount prints the account details returned by GetCurrentAccount
func printFullAccount(w *tabwriter.Writer, fa *users.FullAccount) {
	fmt.Fprintf(w, "Logged in as %s <%s>\n\n", fa.Name.DisplayName, fa.Email)
	fmt.Fprintf(w, "Account Id:\t%s\n", fa.AccountId)
	fmt.Fprintf(w, "Account Type:\t%s\n", fa.AccountType.Tag)
	fmt.Fprintf(w, "Locale:\t%s\n", fa.Locale)
	fmt.Fprintf(w, "Referral Link:\t%s\n", fa.ReferralLink)
	fmt.Fprintf(w, "Profile Photo Url:\t%s\n", fa.ProfilePhotoUrl)
	fmt.Fprintf(w, "Paired Account:\t%t\n", fa.IsPaired)
	if fa.Team != nil {
		fmt.Fprintf(w, "Team:\n  Name:\t%s\n  Id:\t%s\n  Member Id:\t%s\n", fa.Team.Name, fa.Team.Id, fa.TeamMemberId)
	}
}

// printBasicAccount prints the account details returned by GetAccount
func printBasicAccount(w *tabwriter.Writer, ba *users.BasicAccount) {
	fmt.Fprintf(w, "Name:\t%s\n", ba.Name.DisplayName)
	if !ba.EmailVerified {
		ba.Email += " (unverified)"
	}
	fmt.Fprintf(w, "Email:\t%s\n", ba.Email)
	fmt.Fprintf(w, "Is Teammate:\t%t\n", ba.IsTeammate)
	if ba.TeamMemberId != "" {
		fmt.Fprintf(w, "Team Member Id:\t%s\n", ba.TeamMemberId)
	}
	fmt.Fprintf(w, "Profile Photo URL:\t%s\n", ba.ProfilePhotoUrl)
}

func account(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return errors.New("`account` accepts an optional `id` argument")
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 8, 1, ' ', 0)

	if len(args) == 0 {
		// If no arguments are provided get the current user's account
		res, err := dbx.GetCurrentAccount()
		if err != nil {
			return err
		}
		printFullAccount(w, res)
	} else {
		// Otherwise look up an account with the provided ID
		arg := users.NewGetAccountArg(args[0])
		res, err := dbx.GetAccount(arg)
		if err != nil {
			return err
		}
		printBasicAccount(w, res)
	}

	w.Flush()

	return nil
}

var accountCmd = &cobra.Command{
	Use:     "account [flags] [<account-id>]",
	Short:   "Display account information",
	Example: "  dbxcli account\n  dbxcli account dbid:xxxx",
	RunE:    account,
}

func init() {
	RootCmd.AddCommand(accountCmd)
}
