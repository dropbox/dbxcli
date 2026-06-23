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
	"strings"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

type shareLinkCreateOptions struct {
	expires          *time.Time
	removeExpiration bool
	allowDownload    bool
	disallowDownload bool
	access           *sharing.RequestedLinkAccessLevel
	audience         *sharing.LinkAudience
	password         sharedLinkPasswordOptions
}

type shareLinkCreateInput struct {
	Path             string `json:"path"`
	Access           string `json:"access,omitempty"`
	Audience         string `json:"audience,omitempty"`
	Expires          string `json:"expires,omitempty"`
	RemoveExpiration bool   `json:"remove_expiration,omitempty"`
	AllowDownload    bool   `json:"allow_download,omitempty"`
	DisallowDownload bool   `json:"disallow_download,omitempty"`
	Password         bool   `json:"password,omitempty"`
}

type shareLinkCreateOutput struct {
	Input    shareLinkCreateInput  `json:"input"`
	Result   shareLinkJSONMetadata `json:"result"`
	Existing bool                  `json:"existing"`
}

func shareLinkCreate(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("`share-link create` requires a `path` argument")
	}

	path, err := validatePath(args[0])
	if err != nil {
		return err
	}
	if path == "" {
		return errors.New("cannot create a shared link for Dropbox root")
	}

	opts, err := parseShareLinkCreateOptions(cmd)
	if err != nil {
		return err
	}

	dbx := newSharedLinkClient(config)
	link, err := createSharedLink(dbx, path, opts)
	usedExisting := false
	if err != nil {
		link, err = existingSharedLink(dbx, path, err)
		if err != nil {
			return err
		}
		link, err = applyExistingSharedLinkCreateOptions(dbx, link, opts)
		if err != nil {
			return err
		}
		usedExisting = true
	}

	url, ok := sharedLinkURL(link)
	if !ok {
		return errors.New("shared link response did not include a URL")
	}

	out := commandOutput(cmd)
	if usedExisting {
		commandVerboseStatus(cmd, "Using existing shared link for %s", path)
	} else {
		commandVerboseStatus(cmd, "Created shared link for %s", path)
	}

	result, ok := shareLinkJSONMetadataFromDropbox(link)
	if !ok {
		return errors.New("found unknown shared link type")
	}

	return out.Render(func(w io.Writer) error {
		_, err := fmt.Fprintln(w, url)
		return err
	}, shareLinkCreateOutput{
		Input:    newShareLinkCreateInput(path, opts),
		Result:   result,
		Existing: usedExisting,
	})
}

func newShareLinkCreateInput(path string, opts shareLinkCreateOptions) shareLinkCreateInput {
	input := shareLinkCreateInput{
		Path:             path,
		RemoveExpiration: opts.removeExpiration,
		AllowDownload:    opts.allowDownload,
		DisallowDownload: opts.disallowDownload,
		Password:         opts.password.set,
	}
	if opts.access != nil {
		input.Access = opts.access.Tag
	}
	if opts.audience != nil {
		input.Audience = opts.audience.Tag
	}
	if opts.expires != nil {
		input.Expires = opts.expires.UTC().Format(time.RFC3339)
	}
	return input
}

func createSharedLink(dbx sharedLinkClient, path string, opts shareLinkCreateOptions) (sharing.IsSharedLinkMetadata, error) {
	if opts.disallowDownload {
		return dbx.CreateSharedLinkWithRawSettings(path, rawSharedLinkSettingsFromCreateOptions(opts))
	}

	arg := sharing.NewCreateSharedLinkWithSettingsArg(path)
	if opts.hasCreateSettings() {
		arg.Settings = sharing.NewSharedLinkSettings()
		applySharedLinkCreateSettings(arg.Settings, opts)
	}
	return dbx.CreateSharedLinkWithSettings(arg)
}

func parseShareLinkCreateOptions(cmd *cobra.Command) (shareLinkCreateOptions, error) {
	var opts shareLinkCreateOptions

	if cmd.Flags().Changed("expires") {
		expires, err := shareLinkExpiresFlag(cmd)
		if err != nil {
			return opts, err
		}
		opts.expires = expires
	}

	if cmd.Flags().Changed("remove-expiration") {
		removeExpiration, err := cmd.Flags().GetBool("remove-expiration")
		if err != nil {
			return opts, err
		}
		opts.removeExpiration = removeExpiration
	}

	if cmd.Flags().Changed("allow-download") {
		allowDownload, err := cmd.Flags().GetBool("allow-download")
		if err != nil {
			return opts, err
		}
		opts.allowDownload = allowDownload
	}

	if cmd.Flags().Changed("disallow-download") {
		disallowDownload, err := cmd.Flags().GetBool("disallow-download")
		if err != nil {
			return opts, err
		}
		opts.disallowDownload = disallowDownload
	}

	if cmd.Flags().Changed("access") {
		access, err := shareLinkAccessFlag(cmd)
		if err != nil {
			return opts, err
		}
		opts.access = access
	}

	if cmd.Flags().Changed("audience") {
		audience, err := shareLinkAudienceFlag(cmd)
		if err != nil {
			return opts, err
		}
		opts.audience = audience
	}

	password, err := sharedLinkPasswordFromFlags(cmd)
	if err != nil {
		return opts, err
	}
	opts.password = password

	if opts.expires != nil && opts.removeExpiration {
		return opts, errors.New("`--expires` and `--remove-expiration` cannot be used together")
	}
	if opts.allowDownload && opts.disallowDownload {
		return opts, errors.New("`--allow-download` and `--disallow-download` cannot be used together")
	}

	return opts, nil
}

func applyExistingSharedLinkCreateOptions(dbx sharedLinkClient, link sharing.IsSharedLinkMetadata, opts shareLinkCreateOptions) (sharing.IsSharedLinkMetadata, error) {
	if opts.access != nil {
		return nil, errors.New("cannot apply `--access` because the shared link already exists")
	}
	if opts.expires == nil && !opts.removeExpiration && !opts.allowDownload && !opts.disallowDownload && opts.audience == nil && !opts.password.set {
		return link, nil
	}

	url, ok := sharedLinkURL(link)
	if !ok {
		return nil, errors.New("existing shared link response did not include a URL")
	}

	if opts.disallowDownload {
		if err := dbx.ModifySharedLinkSettingsRaw(url, rawSharedLinkSettingsFromCreateOptions(opts), opts.removeExpiration); err != nil {
			return nil, err
		}
		return link, nil
	}

	if opts.expires != nil || opts.removeExpiration || opts.allowDownload || opts.audience != nil || opts.password.set {
		settings := sharing.NewSharedLinkSettings()
		applySharedLinkCreateSettings(settings, opts)

		arg := sharing.NewModifySharedLinkSettingsArgs(url, settings)
		arg.RemoveExpiration = opts.removeExpiration

		updated, err := dbx.ModifySharedLinkSettings(arg)
		if err != nil {
			return nil, err
		}
		link = updated
	}

	return link, nil
}

func (opts shareLinkCreateOptions) hasCreateSettings() bool {
	return opts.expires != nil || opts.allowDownload || opts.access != nil || opts.audience != nil || opts.password.set
}

func applySharedLinkCreateSettings(settings *sharing.SharedLinkSettings, opts shareLinkCreateOptions) {
	if opts.expires != nil {
		settings.Expires = opts.expires
	}
	if opts.allowDownload {
		settings.AllowDownload = true
	}
	if opts.access != nil {
		settings.Access = opts.access
	}
	if opts.audience != nil {
		settings.Audience = opts.audience
	}
	if opts.password.set {
		settings.RequirePassword = true
		settings.LinkPassword = opts.password.password
	}
}

func rawSharedLinkSettingsFromCreateOptions(opts shareLinkCreateOptions) *rawSharedLinkSettings {
	settings := &rawSharedLinkSettings{
		Expires:  opts.expires,
		Audience: opts.audience,
		Access:   opts.access,
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
	return settings
}

func existingSharedLink(dbx sharedLinkClient, path string, err error) (sharing.IsSharedLinkMetadata, error) {
	apiErr, ok := createSharedLinkWithSettingsAPIError(err)
	if !ok || apiErr.EndpointError == nil ||
		apiErr.EndpointError.Tag != sharing.CreateSharedLinkWithSettingsErrorSharedLinkAlreadyExists {
		return nil, err
	}

	if link, ok := linkFromAlreadyExists(apiErr.EndpointError.SharedLinkAlreadyExists); ok {
		return link, nil
	}

	return findExistingSharedLink(dbx, path)
}

func createSharedLinkWithSettingsAPIError(err error) (*sharing.CreateSharedLinkWithSettingsAPIError, bool) {
	var apiErr sharing.CreateSharedLinkWithSettingsAPIError
	if errors.As(err, &apiErr) {
		return &apiErr, true
	}

	var apiErrPtr *sharing.CreateSharedLinkWithSettingsAPIError
	if errors.As(err, &apiErrPtr) {
		return apiErrPtr, true
	}

	return nil, false
}

func linkFromAlreadyExists(meta *sharing.SharedLinkAlreadyExistsMetadata) (sharing.IsSharedLinkMetadata, bool) {
	if meta == nil || meta.Tag != sharing.SharedLinkAlreadyExistsMetadataMetadata || meta.Metadata == nil {
		return nil, false
	}
	return meta.Metadata, true
}

func findExistingSharedLink(dbx sharedLinkClient, requestedPath string) (sharing.IsSharedLinkMetadata, error) {
	arg := sharing.NewListSharedLinksArg()
	arg.Path = requestedPath
	arg.DirectOnly = true

	var firstDirect sharing.IsSharedLinkMetadata

	for {
		res, err := dbx.ListSharedLinks(arg)
		if err != nil {
			return nil, err
		}

		for _, link := range res.Links {
			linkPath, ok := sharedLinkPathLower(link)
			if ok {
				if sameDropboxPath(linkPath, requestedPath) {
					return link, nil
				}
				continue
			}

			if firstDirect == nil {
				firstDirect = link
			}
		}

		if !res.HasMore {
			break
		}
		if res.Cursor == "" {
			return nil, errors.New("shared link lookup has more results but no cursor")
		}
		arg.Cursor = res.Cursor
	}

	if firstDirect != nil {
		return firstDirect, nil
	}

	return nil, fmt.Errorf("shared link already exists but no direct link was found for %q", requestedPath)
}

func sharedLinkURL(link sharing.IsSharedLinkMetadata) (string, bool) {
	switch link := link.(type) {
	case *sharing.FileLinkMetadata:
		return nonEmptyString(link.Url)
	case *sharing.FolderLinkMetadata:
		return nonEmptyString(link.Url)
	case *sharing.SharedLinkMetadata:
		return nonEmptyString(link.Url)
	default:
		return "", false
	}
}

func sharedLinkPathLower(link sharing.IsSharedLinkMetadata) (string, bool) {
	switch link := link.(type) {
	case *sharing.FileLinkMetadata:
		return nonEmptyString(link.PathLower)
	case *sharing.FolderLinkMetadata:
		return nonEmptyString(link.PathLower)
	case *sharing.SharedLinkMetadata:
		return nonEmptyString(link.PathLower)
	default:
		return "", false
	}
}

func nonEmptyString(value string) (string, bool) {
	return value, value != ""
}

func sameDropboxPath(a string, b string) bool {
	return strings.EqualFold(cleanDropboxPath(a), cleanDropboxPath(b))
}

func shareLinkExpiresFlag(cmd *cobra.Command) (*time.Time, error) {
	value, err := cmd.Flags().GetString("expires")
	if err != nil {
		return nil, err
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("invalid --expires %q: use RFC3339 timestamp", value)
	}
	return &parsed, nil
}

func shareLinkAccessFlag(cmd *cobra.Command) (*sharing.RequestedLinkAccessLevel, error) {
	value, err := cmd.Flags().GetString("access")
	if err != nil {
		return nil, err
	}
	switch value {
	case sharing.RequestedLinkAccessLevelViewer:
		return requestedLinkAccessLevel(sharing.RequestedLinkAccessLevelViewer), nil
	case sharing.RequestedLinkAccessLevelEditor:
		return requestedLinkAccessLevel(sharing.RequestedLinkAccessLevelEditor), nil
	case sharing.RequestedLinkAccessLevelMax:
		return requestedLinkAccessLevel(sharing.RequestedLinkAccessLevelMax), nil
	default:
		return nil, fmt.Errorf("invalid --access %q: use viewer, editor, or max", value)
	}
}

func requestedLinkAccessLevel(tag string) *sharing.RequestedLinkAccessLevel {
	return &sharing.RequestedLinkAccessLevel{Tagged: dropbox.Tagged{Tag: tag}}
}

func shareLinkAudienceFlag(cmd *cobra.Command) (*sharing.LinkAudience, error) {
	value, err := cmd.Flags().GetString("audience")
	if err != nil {
		return nil, err
	}
	switch value {
	case sharing.LinkAudiencePublic:
		return linkAudience(sharing.LinkAudiencePublic), nil
	case sharing.LinkAudienceTeam:
		return linkAudience(sharing.LinkAudienceTeam), nil
	case sharing.LinkAudienceMembers:
		return linkAudience(sharing.LinkAudienceMembers), nil
	case "no-one":
		return linkAudience(sharing.LinkAudienceNoOne), nil
	default:
		return nil, fmt.Errorf("invalid --audience %q: use public, team, members, or no-one", value)
	}
}

func linkAudience(tag string) *sharing.LinkAudience {
	return &sharing.LinkAudience{Tagged: dropbox.Tagged{Tag: tag}}
}

var shareLinkCreateCmd = &cobra.Command{
	Use:   "create <path>",
	Short: "Create a shared link",
	Long: `Create a shared link for a Dropbox file or folder.
If a direct shared link already exists, dbxcli returns that existing URL.
Settings flags request Dropbox shared-link settings; account, team, and folder policies may still restrict the result.`,
	Example: `  dbxcli share-link create /file.txt
  dbxcli share-link create /folder
  dbxcli share-link create /file.txt --audience team
  dbxcli share-link create /file.txt --expires 2026-07-01T00:00:00Z
  dbxcli share-link create /file.txt --password-prompt`,
	RunE: shareLinkCreate,
}

func init() {
	shareLinkCreateCmd.Flags().String("access", "", "Set shared link access level: viewer, editor, or max")
	shareLinkCreateCmd.Flags().String("audience", "", "Set shared link audience: public, team, members, or no-one")
	shareLinkCreateCmd.Flags().Bool("allow-download", false, "Allow downloads from the shared link")
	shareLinkCreateCmd.Flags().Bool("disallow-download", false, "Disallow downloads from the shared link")
	shareLinkCreateCmd.Flags().String("expires", "", "Set shared link expiration time as an RFC3339 timestamp")
	shareLinkCreateCmd.Flags().Bool("remove-expiration", false, "Remove expiration when returning an existing shared link")
	addSharedLinkPasswordFlags(shareLinkCreateCmd)
	shareLinkCmd.AddCommand(shareLinkCreateCmd)
	enableStructuredOutput(shareLinkCreateCmd)
}
