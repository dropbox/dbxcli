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

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/users"
	"github.com/spf13/cobra"
)

type accountInput struct {
	AccountID string `json:"account_id,omitempty"`
}

type accountOutput struct {
	Input   accountInput `json:"input"`
	Account jsonAccount  `json:"account"`
}

type jsonAccount struct {
	Type            string           `json:"type"`
	AccountID       string           `json:"account_id"`
	Name            *jsonAccountName `json:"name,omitempty"`
	Email           string           `json:"email"`
	EmailVerified   bool             `json:"email_verified"`
	Disabled        bool             `json:"disabled"`
	ProfilePhotoURL string           `json:"profile_photo_url,omitempty"`
	Locale          string           `json:"locale,omitempty"`
	ReferralLink    string           `json:"referral_link,omitempty"`
	IsPaired        *bool            `json:"is_paired,omitempty"`
	AccountType     string           `json:"account_type,omitempty"`
	IsTeammate      *bool            `json:"is_teammate,omitempty"`
	TeamMemberID    string           `json:"team_member_id,omitempty"`
	Team            *jsonAccountTeam `json:"team,omitempty"`
}

type jsonAccountName struct {
	GivenName       string `json:"given_name,omitempty"`
	Surname         string `json:"surname,omitempty"`
	FamiliarName    string `json:"familiar_name,omitempty"`
	DisplayName     string `json:"display_name,omitempty"`
	AbbreviatedName string `json:"abbreviated_name,omitempty"`
}

type jsonAccountTeam struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	MemberID string `json:"member_id,omitempty"`
}

// renderFullAccount prints the account details returned by GetCurrentAccount.
func renderFullAccount(out io.Writer, fa *users.FullAccount) error {
	w := new(tabwriter.Writer)
	w.Init(out, 4, 8, 1, ' ', 0)

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

	return w.Flush()
}

// renderBasicAccount prints the account details returned by GetAccount.
func renderBasicAccount(out io.Writer, ba *users.BasicAccount) error {
	w := new(tabwriter.Writer)
	w.Init(out, 4, 8, 1, ' ', 0)

	fmt.Fprintf(w, "Name:\t%s\n", ba.Name.DisplayName)
	email := ba.Email
	if !ba.EmailVerified {
		email += " (unverified)"
	}
	fmt.Fprintf(w, "Email:\t%s\n", email)
	fmt.Fprintf(w, "Is Teammate:\t%t\n", ba.IsTeammate)
	if ba.TeamMemberId != "" {
		fmt.Fprintf(w, "Team Member Id:\t%s\n", ba.TeamMemberId)
	}
	fmt.Fprintf(w, "Profile Photo URL:\t%s\n", ba.ProfilePhotoUrl)

	return w.Flush()
}

func account(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return errors.New("`account` accepts an optional `id` argument")
	}

	dbx := usersNewFunc(config)
	out := commandOutput(cmd)

	if len(args) == 0 {
		// If no arguments are provided get the current user's account
		res, err := dbx.GetCurrentAccount()
		if err != nil {
			return err
		}
		return out.Render(func(w io.Writer) error {
			return renderFullAccount(w, res)
		}, accountOutput{
			Input:   accountInput{},
			Account: jsonFullAccount(res),
		})
	}

	// Otherwise look up an account with the provided ID
	arg := users.NewGetAccountArg(args[0])
	res, err := dbx.GetAccount(arg)
	if err != nil {
		return err
	}
	return out.Render(func(w io.Writer) error {
		return renderBasicAccount(w, res)
	}, accountOutput{
		Input: accountInput{
			AccountID: args[0],
		},
		Account: jsonBasicAccount(res),
	})
}

func jsonFullAccount(fa *users.FullAccount) jsonAccount {
	account := jsonAccountFromBase("full", fa.Account)
	account.Locale = fa.Locale
	account.ReferralLink = fa.ReferralLink
	account.IsPaired = boolPtr(fa.IsPaired)
	if fa.AccountType != nil {
		account.AccountType = fa.AccountType.Tag
	}
	account.TeamMemberID = fa.TeamMemberId
	if fa.Team != nil {
		account.Team = &jsonAccountTeam{
			ID:       fa.Team.Id,
			Name:     fa.Team.Name,
			MemberID: fa.TeamMemberId,
		}
	}
	return account
}

func jsonBasicAccount(ba *users.BasicAccount) jsonAccount {
	account := jsonAccountFromBase("basic", ba.Account)
	account.IsTeammate = boolPtr(ba.IsTeammate)
	account.TeamMemberID = ba.TeamMemberId
	return account
}

func jsonAccountFromBase(accountType string, account users.Account) jsonAccount {
	return jsonAccount{
		Type:            accountType,
		AccountID:       account.AccountId,
		Name:            jsonAccountNameFromDropbox(account.Name),
		Email:           account.Email,
		EmailVerified:   account.EmailVerified,
		Disabled:        account.Disabled,
		ProfilePhotoURL: account.ProfilePhotoUrl,
	}
}

func jsonAccountNameFromDropbox(name *users.Name) *jsonAccountName {
	if name == nil {
		return nil
	}
	return &jsonAccountName{
		GivenName:       name.GivenName,
		Surname:         name.Surname,
		FamiliarName:    name.FamiliarName,
		DisplayName:     name.DisplayName,
		AbbreviatedName: name.AbbreviatedName,
	}
}

func boolPtr(value bool) *bool {
	return &value
}

var accountCmd = &cobra.Command{
	Use:     "account [flags] [<account-id>]",
	Short:   "Display account information",
	Example: "  dbxcli account\n  dbxcli account dbid:xxxx",
	RunE:    account,
}

func init() {
	RootCmd.AddCommand(accountCmd)
	enableStructuredOutput(accountCmd)
}
