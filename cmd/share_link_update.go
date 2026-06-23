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

	"github.com/dropbox/dbxcli/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

type shareLinkUpdateOptions struct {
	expires          *time.Time
	removeExpiration bool
	allowDownload    bool
	disallowDownload bool
	audience         *sharing.LinkAudience
	password         sharedLinkPasswordOptions
	removePassword   bool
}

type shareLinkUpdateInput struct {
	URL              string `json:"url"`
	Audience         string `json:"audience,omitempty"`
	Expires          string `json:"expires,omitempty"`
	RemoveExpiration bool   `json:"remove_expiration,omitempty"`
	AllowDownload    bool   `json:"allow_download,omitempty"`
	DisallowDownload bool   `json:"disallow_download,omitempty"`
	Password         bool   `json:"password,omitempty"`
	RemovePassword   bool   `json:"remove_password,omitempty"`
}

type shareLinkUpdateOutput struct {
	Input  shareLinkUpdateInput  `json:"input"`
	Result shareLinkJSONMetadata `json:"result"`
}

func shareLinkUpdate(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("`share-link update` requires a `url` argument")
	}

	url := args[0]
	if url == "" {
		return errors.New("`share-link update` requires a non-empty URL")
	}

	opts, err := parseShareLinkUpdateOptions(cmd)
	if err != nil {
		return err
	}

	dbx := newSharedLinkClient(config)
	if opts.usesRawSettings() {
		if err := dbx.ModifySharedLinkSettingsRaw(url, rawSharedLinkSettingsFromUpdateOptions(opts), opts.removeExpiration); err != nil {
			return err
		}
		return renderShareLinkUpdateOutput(cmd, dbx, url, opts, nil)
	}

	settings := sharing.NewSharedLinkSettings()
	if opts.expires != nil {
		settings.Expires = opts.expires
	}
	if opts.allowDownload {
		settings.AllowDownload = true
	}
	if opts.audience != nil {
		settings.Audience = opts.audience
	}
	if opts.password.set {
		settings.RequirePassword = true
		settings.LinkPassword = opts.password.password
	}

	arg := sharing.NewModifySharedLinkSettingsArgs(url, settings)
	arg.RemoveExpiration = opts.removeExpiration

	if opts.hasSDKSettings() {
		link, err := dbx.ModifySharedLinkSettings(arg)
		if err != nil {
			return err
		}
		return renderShareLinkUpdateOutput(cmd, dbx, url, opts, link)
	}

	return renderShareLinkUpdateOutput(cmd, dbx, url, opts, nil)
}

func renderShareLinkUpdateOutput(cmd *cobra.Command, dbx sharedLinkClient, url string, opts shareLinkUpdateOptions, link sharing.IsSharedLinkMetadata) error {
	commandVerboseStatus(cmd, "Updated shared link %s", url)

	if commandOutputFormat(cmd) != output.FormatJSON {
		return nil
	}

	if link == nil {
		arg := sharing.NewGetSharedLinkMetadataArg(url)
		if opts.password.set {
			arg.LinkPassword = opts.password.password
		}
		var err error
		link, err = dbx.GetSharedLinkMetadata(arg)
		if err != nil {
			return err
		}
	}

	result, ok := shareLinkJSONMetadataFromDropbox(link)
	if !ok {
		return errors.New("found unknown shared link type")
	}

	return commandOutput(cmd).Render(nil, shareLinkUpdateOutput{
		Input:  newShareLinkUpdateInput(url, opts),
		Result: result,
	})
}

func newShareLinkUpdateInput(url string, opts shareLinkUpdateOptions) shareLinkUpdateInput {
	input := shareLinkUpdateInput{
		URL:              url,
		RemoveExpiration: opts.removeExpiration,
		AllowDownload:    opts.allowDownload,
		DisallowDownload: opts.disallowDownload,
		Password:         opts.password.set,
		RemovePassword:   opts.removePassword,
	}
	if opts.audience != nil {
		input.Audience = opts.audience.Tag
	}
	if opts.expires != nil {
		input.Expires = opts.expires.UTC().Format(time.RFC3339)
	}
	return input
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
	disallowDownload, err := cmd.Flags().GetBool("disallow-download")
	if err != nil {
		return shareLinkUpdateOptions{}, err
	}
	audienceChanged := cmd.Flags().Changed("audience")
	password, err := sharedLinkPasswordFromFlags(cmd)
	if err != nil {
		return shareLinkUpdateOptions{}, err
	}
	removePassword, err := localBoolFlag(cmd, "remove-password")
	if err != nil {
		return shareLinkUpdateOptions{}, err
	}

	if expiresChanged && removeExpiration {
		return shareLinkUpdateOptions{}, errors.New("`--expires` and `--remove-expiration` cannot be used together")
	}
	if allowDownload && disallowDownload {
		return shareLinkUpdateOptions{}, errors.New("`--allow-download` and `--disallow-download` cannot be used together")
	}
	if password.set && removePassword {
		return shareLinkUpdateOptions{}, errors.New("password-setting flags and `--remove-password` cannot be used together")
	}
	if !expiresChanged && !removeExpiration && !allowDownload && !disallowDownload && !audienceChanged && !password.set && !removePassword {
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

	var audience *sharing.LinkAudience
	if audienceChanged {
		parsed, err := shareLinkAudienceFlag(cmd)
		if err != nil {
			return shareLinkUpdateOptions{}, err
		}
		audience = parsed
	}

	return shareLinkUpdateOptions{
		expires:          expires,
		removeExpiration: removeExpiration,
		allowDownload:    allowDownload,
		disallowDownload: disallowDownload,
		audience:         audience,
		password:         password,
		removePassword:   removePassword,
	}, nil
}

func (opts shareLinkUpdateOptions) hasSDKSettings() bool {
	return opts.expires != nil || opts.removeExpiration || opts.allowDownload || opts.audience != nil || opts.password.set
}

func (opts shareLinkUpdateOptions) usesRawSettings() bool {
	return opts.disallowDownload || opts.removePassword
}

func rawSharedLinkSettingsFromUpdateOptions(opts shareLinkUpdateOptions) *rawSharedLinkSettings {
	settings := &rawSharedLinkSettings{
		Expires:  opts.expires,
		Audience: opts.audience,
	}
	if opts.allowDownload || opts.disallowDownload {
		allowDownload := opts.allowDownload
		settings.AllowDownload = &allowDownload
	}
	if opts.password.set {
		requirePassword := true
		settings.RequirePassword = &requirePassword
		settings.LinkPassword = opts.password.password
	}
	if opts.removePassword {
		requirePassword := false
		settings.RequirePassword = &requirePassword
	}
	return settings
}

var shareLinkUpdateCmd = &cobra.Command{
	Use:   "update <url>",
	Short: "Update shared link settings",
	Long: `Update settings for an existing shared link.
At least one setting flag is required. Dropbox account, team, and folder policies may reject or constrain requested settings.`,
	Example: `  dbxcli share-link update https://www.dropbox.com/s/example/file.txt --audience team
  dbxcli share-link update https://www.dropbox.com/s/example/file.txt --expires 2026-07-01T00:00:00Z
  dbxcli share-link update https://www.dropbox.com/s/example/file.txt --remove-expiration
  dbxcli share-link update https://www.dropbox.com/s/example/file.txt --password-prompt
  dbxcli share-link update https://www.dropbox.com/s/example/file.txt --remove-password`,
	RunE: shareLinkUpdate,
}

func init() {
	shareLinkUpdateCmd.Flags().String("audience", "", "Set shared link audience: public, team, members, or no-one")
	shareLinkUpdateCmd.Flags().String("expires", "", "Set shared link expiration time as an RFC3339 timestamp")
	shareLinkUpdateCmd.Flags().Bool("remove-expiration", false, "Remove the shared link expiration time")
	shareLinkUpdateCmd.Flags().Bool("allow-download", false, "Allow downloads from the shared link")
	shareLinkUpdateCmd.Flags().Bool("disallow-download", false, "Disallow downloads from the shared link")
	addSharedLinkPasswordFlags(shareLinkUpdateCmd)
	shareLinkUpdateCmd.Flags().Bool("remove-password", false, "Remove the shared link password")
	shareLinkCmd.AddCommand(shareLinkUpdateCmd)
	enableStructuredOutput(shareLinkUpdateCmd)
}
