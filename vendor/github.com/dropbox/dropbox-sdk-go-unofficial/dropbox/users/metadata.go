// Copyright (c) Dropbox, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package users

import (
	"encoding/json"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/common"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/users_common"
)

type fullAccount struct {
	Account
	Country      string                    `json:"country,omitempty"`
	Locale       string                    `json:"locale"`
	ReferralLink string                    `json:"referral_link"`
	Team         *FullTeam                 `json:"team,omitempty"`
	TeamMemberId string                    `json:"team_member_id,omitempty"`
	IsPaired     bool                      `json:"is_paired"`
	AccountType  *users_common.AccountType `json:"account_type"`
	RootInfo     json.RawMessage           `json:"root_info"`
}

func (f *FullAccount) UnmarshalJSON(b []byte) error {
	var fa fullAccount
	if err := json.Unmarshal(b, &fa); err != nil {
		return err
	}
	f.Account = fa.Account
	f.Country = fa.Country
	f.Locale = fa.Locale
	f.ReferralLink = fa.ReferralLink
	f.Team = fa.Team
	f.TeamMemberId = fa.TeamMemberId
	f.IsPaired = fa.IsPaired
	f.AccountType = fa.AccountType
	rootInfo, err := common.IsRootInfoFromJSON(fa.RootInfo)
	if err != nil {
		return err
	}
	f.RootInfo = rootInfo
	return nil
}
