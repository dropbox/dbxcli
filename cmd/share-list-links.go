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

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

type shareLinkListInput struct {
	Path       string `json:"path,omitempty"`
	DirectOnly bool   `json:"direct_only"`
}

func shareListLinks(cmd *cobra.Command, args []string) (err error) {
	return shareLinkListWithWarnings(cmd, args, []jsonWarning{{
		Code:    jsonWarningCodeDeprecatedCommand,
		Message: "use `dbxcli share-link list` instead",
	}})
}

func shareLinkList(cmd *cobra.Command, args []string) error {
	return shareLinkListWithWarnings(cmd, args, nil)
}

func shareLinkListWithWarnings(cmd *cobra.Command, args []string, warnings []jsonWarning) error {
	if len(args) > 1 {
		return invalidArgumentsErrorWithDetails("`share-link list` accepts at most one `path` argument", argumentErrorDetails("path"))
	}

	arg := sharing.NewListSharedLinksArg()
	if len(args) == 1 {
		path, err := validatePath(args[0])
		if err != nil {
			return err
		}
		arg.Path = path
		arg.DirectOnly = true
	}

	dbx := newSharedLinkClient(config)
	links, err := listSharedLinks(dbx, arg)
	if err != nil {
		return err
	}

	if arg.Path != "" {
		commandVerboseStatus(cmd, "Listed %d shared links for %s", len(links), arg.Path)
	} else {
		commandVerboseStatus(cmd, "Listed %d shared links", len(links))
	}

	entries, ok := shareLinkJSONMetadataListFromDropbox(links)
	if !ok {
		return errors.New("found unknown shared link type")
	}

	return commandOutput(cmd).Render(func(w io.Writer) error {
		return renderSharedLinks(w, links)
	}, newJSONCommandOperationOutput(
		cmd,
		shareLinkListInput{
			Path:       arg.Path,
			DirectOnly: arg.DirectOnly,
		},
		shareLinkJSONOperationResults(shareLinkJSONStatusListed, entries),
		warnings,
	))
}

func listSharedLinks(dbx sharedLinkClient, arg *sharing.ListSharedLinksArg) ([]sharing.IsSharedLinkMetadata, error) {
	var links []sharing.IsSharedLinkMetadata
	for {
		res, err := dbx.ListSharedLinks(arg)
		if err != nil {
			return nil, err
		}
		links = append(links, res.Links...)

		if !res.HasMore {
			break
		}
		if res.Cursor == "" {
			return nil, errors.New("shared link list has more results but no cursor")
		}
		arg = sharing.NewListSharedLinksArg()
		arg.Cursor = res.Cursor
	}

	return links, nil
}

func renderSharedLinks(out io.Writer, links []sharing.IsSharedLinkMetadata) error {
	for _, l := range links {
		name, url, ok := sharedLinkDisplay(l)
		if !ok {
			return errors.New("found unknown shared link type")
		}
		if _, err := fmt.Fprintf(out, "%s\t%s\n", name, url); err != nil {
			return err
		}
	}

	return nil
}

func sharedLinkDisplay(link sharing.IsSharedLinkMetadata) (name string, url string, ok bool) {
	switch sl := link.(type) {
	case *sharing.FileLinkMetadata:
		return sharedLinkMetadataDisplay(sl.SharedLinkMetadata)
	case *sharing.FolderLinkMetadata:
		return sharedLinkMetadataDisplay(sl.SharedLinkMetadata)
	case *sharing.SharedLinkMetadata:
		return sharedLinkMetadataDisplay(*sl)
	default:
		return "", "", false
	}
}

func sharedLinkMetadataDisplay(sl sharing.SharedLinkMetadata) (name string, url string, ok bool) {
	name = sl.Name
	if name == "" {
		name = sl.PathLower
	}
	return name, sl.Url, sl.Url != ""
}

var shareLinkListCmd = &cobra.Command{
	Use:   "list [path]",
	Short: "List shared links",
	Long: `List shared links.
When path is supplied, dbxcli lists direct shared links for that Dropbox path only.`,
	Example: `  dbxcli share-link list
  dbxcli share-link list /file.txt`,
	RunE: shareLinkList,
}

var shareListLinksCmd = &cobra.Command{
	Use:        "link [path]",
	Short:      "List shared links",
	Deprecated: "use `dbxcli share-link list` instead",
	RunE:       shareListLinks,
}

func init() {
	shareLinkCmd.AddCommand(shareLinkListCmd)
	shareListCmd.AddCommand(shareListLinksCmd)
	enableStructuredOutput(shareLinkListCmd)
	enableStructuredOutput(shareListLinksCmd)
}
