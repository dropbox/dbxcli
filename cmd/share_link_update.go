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
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

type shareLinkUpdateOptions struct {
	expires          *time.Time
	removeExpiration bool
	allowDownload    bool
}

func shareLinkUpdate(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("`share link update` requires a `url` argument")
	}

	url := args[0]
	if url == "" {
		return errors.New("`share link update` requires a non-empty URL")
	}

	opts, err := parseShareLinkUpdateOptions(cmd)
	if err != nil {
		return err
	}

	settings := sharing.NewSharedLinkSettings()
	if opts.expires != nil {
		settings.Expires = opts.expires
	}
	if opts.allowDownload {
		settings.AllowDownload = true
	}

	arg := sharing.NewModifySharedLinkSettingsArgs(url, settings)
	arg.RemoveExpiration = opts.removeExpiration

	dbx := newSharedLinkClient(config)
	if _, err := dbx.ModifySharedLinkSettings(arg); err != nil {
		return err
	}

	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		commandOutput(cmd).Status("Updated shared link %s", url)
	}

	return nil
}

func parseShareLinkUpdateOptions(cmd *cobra.Command) (shareLinkUpdateOptions, error) {
	expiresChanged := cmd.Flags().Changed("expires")
	removeExpiration, err := cmd.Flags().GetBool("remove-expiration")
	if err != nil {
		return shareLinkUpdateOptions{}, err
	}
	allowDownload, err := cmd.Flags().GetBool("allow-download")
	if err != nil {
		return shareLinkUpdateOptions{}, err
	}

	if expiresChanged && removeExpiration {
		return shareLinkUpdateOptions{}, errors.New("`--expires` and `--remove-expiration` cannot be used together")
	}
	if !expiresChanged && !removeExpiration && !allowDownload {
		return shareLinkUpdateOptions{}, errors.New("at least one shared link setting flag is required")
	}

	var expires *time.Time
	if expiresChanged {
		value, err := cmd.Flags().GetString("expires")
		if err != nil {
			return shareLinkUpdateOptions{}, err
		}
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return shareLinkUpdateOptions{}, fmt.Errorf("invalid --expires %q: use RFC3339 timestamp", value)
		}
		expires = &parsed
	}

	return shareLinkUpdateOptions{
		expires:          expires,
		removeExpiration: removeExpiration,
		allowDownload:    allowDownload,
	}, nil
}

var shareLinkUpdateCmd = &cobra.Command{
	Use:   "update <url>",
	Short: "Update shared link settings",
	RunE:  shareLinkUpdate,
}

func init() {
	shareLinkUpdateCmd.Flags().String("expires", "", "Set shared link expiration time as an RFC3339 timestamp")
	shareLinkUpdateCmd.Flags().Bool("remove-expiration", false, "Remove the shared link expiration time")
	shareLinkUpdateCmd.Flags().Bool("allow-download", false, "Allow downloads from the shared link")
	shareLinkCmd.AddCommand(shareLinkUpdateCmd)
}
