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

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

func shareLinkRevoke(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("`share-link revoke` requires a `url` argument")
	}

	url := args[0]
	if url == "" {
		return errors.New("`share-link revoke` requires a non-empty URL")
	}

	dbx := newSharedLinkClient(config)
	arg := sharing.NewRevokeSharedLinkArg(url)
	if err := dbx.RevokeSharedLink(arg); err != nil {
		return err
	}

	commandVerboseStatus(cmd, "Revoked shared link %s", url)
	return nil
}

var shareLinkRevokeCmd = &cobra.Command{
	Use:   "revoke <url>",
	Short: "Revoke a shared link",
	RunE:  shareLinkRevoke,
}

func init() {
	shareLinkCmd.AddCommand(shareLinkRevokeCmd)
}
