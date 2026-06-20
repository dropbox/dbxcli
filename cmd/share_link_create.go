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

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

func shareLinkCreate(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("`share link create` requires a `path` argument")
	}

	path, err := validatePath(args[0])
	if err != nil {
		return err
	}
	if path == "" {
		return errors.New("cannot create a shared link for Dropbox root")
	}

	dbx := newSharedLinkClient(config)
	arg := sharing.NewCreateSharedLinkWithSettingsArg(path)
	link, err := dbx.CreateSharedLinkWithSettings(arg)
	if err != nil {
		link, err = existingSharedLink(dbx, path, err)
		if err != nil {
			return err
		}
	}

	url, ok := sharedLinkURL(link)
	if !ok {
		return errors.New("shared link response did not include a URL")
	}

	return commandOutput(cmd).RenderText(func(w io.Writer) error {
		_, err := fmt.Fprintln(w, url)
		return err
	})
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

var shareLinkCreateCmd = &cobra.Command{
	Use:   "create <path>",
	Short: "Create a shared link",
	RunE:  shareLinkCreate,
}

func init() {
	shareLinkCmd.AddCommand(shareLinkCreateCmd)
}
