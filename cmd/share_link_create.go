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
	access           *sharing.RequestedLinkAccessLevel
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
	arg := sharing.NewCreateSharedLinkWithSettingsArg(path)
	if opts.hasCreateSettings() {
		arg.Settings = sharing.NewSharedLinkSettings()
		applySharedLinkCreateSettings(arg.Settings, opts)
	}
	link, err := dbx.CreateSharedLinkWithSettings(arg)
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

	return out.RenderText(func(w io.Writer) error {
		_, err := fmt.Fprintln(w, url)
		return err
	})
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

	if cmd.Flags().Changed("access") {
		access, err := shareLinkAccessFlag(cmd)
		if err != nil {
			return opts, err
		}
		opts.access = access
	}

	if opts.expires != nil && opts.removeExpiration {
		return opts, errors.New("`--expires` and `--remove-expiration` cannot be used together")
	}

	return opts, nil
}

func applyExistingSharedLinkCreateOptions(dbx sharedLinkClient, link sharing.IsSharedLinkMetadata, opts shareLinkCreateOptions) (sharing.IsSharedLinkMetadata, error) {
	if opts.access != nil {
		return nil, errors.New("cannot apply `--access` because the shared link already exists")
	}
	if opts.expires == nil && !opts.removeExpiration && !opts.allowDownload {
		return link, nil
	}

	url, ok := sharedLinkURL(link)
	if !ok {
		return nil, errors.New("existing shared link response did not include a URL")
	}

	settings := sharing.NewSharedLinkSettings()
	applySharedLinkCreateSettings(settings, opts)

	arg := sharing.NewModifySharedLinkSettingsArgs(url, settings)
	arg.RemoveExpiration = opts.removeExpiration

	return dbx.ModifySharedLinkSettings(arg)
}

func (opts shareLinkCreateOptions) hasCreateSettings() bool {
	return opts.expires != nil || opts.allowDownload || opts.access != nil
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

var shareLinkCreateCmd = &cobra.Command{
	Use:   "create <path>",
	Short: "Create a shared link",
	RunE:  shareLinkCreate,
}

func init() {
	shareLinkCreateCmd.Flags().String("access", "", "Set shared link access level: viewer, editor, or max")
	shareLinkCreateCmd.Flags().Bool("allow-download", false, "Allow downloads from the shared link")
	shareLinkCreateCmd.Flags().String("expires", "", "Set shared link expiration time as an RFC3339 timestamp")
	shareLinkCreateCmd.Flags().Bool("remove-expiration", false, "Remove expiration when returning an existing shared link")
	shareLinkCmd.AddCommand(shareLinkCreateCmd)
}
