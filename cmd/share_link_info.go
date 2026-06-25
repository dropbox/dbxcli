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
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

type shareLinkInfoOptions struct {
	path     string
	password sharedLinkPasswordOptions
}

type shareLinkInfoInput struct {
	URL      string `json:"url"`
	Path     string `json:"path,omitempty"`
	Password bool   `json:"password,omitempty"`
}

func shareLinkInfo(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return invalidArgumentsError("`share-link info` requires a `url` argument")
	}

	url := args[0]
	if url == "" {
		return invalidArgumentsError("`share-link info` requires a non-empty URL")
	}

	opts, err := parseShareLinkInfoOptions(cmd)
	if err != nil {
		return err
	}

	dbx := newSharedLinkClient(config)
	arg := sharing.NewGetSharedLinkMetadataArg(url)
	if opts.path != "" {
		arg.Path = sharedLinkAPIPath(opts.path)
	}
	if opts.password.set {
		arg.LinkPassword = opts.password.password
	}

	link, err := dbx.GetSharedLinkMetadata(arg)
	if err != nil {
		return err
	}

	result, ok := shareLinkJSONMetadataFromDropbox(link)
	if !ok {
		return errors.New("found unknown shared link type")
	}

	return commandOutput(cmd).Render(func(w io.Writer) error {
		return renderSharedLinkInfo(w, link)
	}, newJSONOperationOutput(
		shareLinkInfoInput{
			URL:      url,
			Path:     opts.path,
			Password: opts.password.set,
		},
		[]jsonOperationResult{shareLinkJSONOperationResult(shareLinkJSONStatusFound, result)},
		nil,
	))
}

func parseShareLinkInfoOptions(cmd *cobra.Command) (shareLinkInfoOptions, error) {
	var opts shareLinkInfoOptions

	if localFlagChanged(cmd, "path") {
		path, err := localStringFlag(cmd, "path")
		if err != nil {
			return opts, err
		}
		if path == "" {
			return opts, invalidArgumentsError("`--path` requires a non-empty path")
		}
		opts.path = path
	}

	password, err := sharedLinkPasswordFromFlags(cmd)
	if err != nil {
		return opts, err
	}
	opts.password = password

	return opts, nil
}

func renderSharedLinkInfo(out io.Writer, link sharing.IsSharedLinkMetadata) error {
	metadata, linkType, ok := sharedLinkBaseMetadata(link)
	if !ok {
		return errors.New("found unknown shared link type")
	}

	w := new(tabwriter.Writer)
	w.Init(out, 4, 8, 1, ' ', 0)

	_, _ = fmt.Fprintf(w, "Type:\t%s\n", linkType)
	_, _ = fmt.Fprintf(w, "Name:\t%s\n", metadata.Name)
	_, _ = fmt.Fprintf(w, "URL:\t%s\n", metadata.Url)
	if metadata.PathLower != "" {
		_, _ = fmt.Fprintf(w, "Path:\t%s\n", metadata.PathLower)
	}
	if metadata.Id != "" {
		_, _ = fmt.Fprintf(w, "ID:\t%s\n", metadata.Id)
	}
	if metadata.Expires != nil {
		_, _ = fmt.Fprintf(w, "Expires:\t%s\n", metadata.Expires.Format(time.RFC3339))
	}
	if metadata.LinkPermissions != nil {
		renderSharedLinkPermissions(w, metadata.LinkPermissions)
	}

	if file, ok := link.(*sharing.FileLinkMetadata); ok {
		_, _ = fmt.Fprintf(w, "Revision:\t%s\n", file.Rev)
		_, _ = fmt.Fprintf(w, "Size:\t%s\n", humanize.IBytes(file.Size))
		_, _ = fmt.Fprintf(w, "Server Modified:\t%s\n", file.ServerModified.Format(time.RFC3339))
	}

	return w.Flush()
}

func renderSharedLinkPermissions(w io.Writer, permissions *sharing.LinkPermissions) {
	if permissions.ResolvedVisibility != nil {
		_, _ = fmt.Fprintf(w, "Resolved Visibility:\t%s\n", permissions.ResolvedVisibility.Tag)
	}
	if permissions.RequestedVisibility != nil {
		_, _ = fmt.Fprintf(w, "Requested Visibility:\t%s\n", permissions.RequestedVisibility.Tag)
	}
	if permissions.EffectiveAudience != nil {
		_, _ = fmt.Fprintf(w, "Effective Audience:\t%s\n", permissions.EffectiveAudience.Tag)
	}
	if permissions.LinkAccessLevel != nil {
		_, _ = fmt.Fprintf(w, "Access Level:\t%s\n", permissions.LinkAccessLevel.Tag)
	}
	_, _ = fmt.Fprintf(w, "Can Revoke:\t%t\n", permissions.CanRevoke)
	_, _ = fmt.Fprintf(w, "Allow Download:\t%t\n", permissions.AllowDownload)
}

func sharedLinkBaseMetadata(link sharing.IsSharedLinkMetadata) (*sharing.SharedLinkMetadata, string, bool) {
	switch link := link.(type) {
	case *sharing.FileLinkMetadata:
		return &link.SharedLinkMetadata, "file", true
	case *sharing.FolderLinkMetadata:
		return &link.SharedLinkMetadata, "folder", true
	case *sharing.SharedLinkMetadata:
		return link, "link", true
	default:
		return nil, "", false
	}
}

var shareLinkInfoCmd = &cobra.Command{
	Use:   "info <url>",
	Short: "Display shared link information",
	Long: `Display metadata and permissions for a shared link.
Use --path to inspect a file or folder inside a folder shared link.`,
	Example: `  dbxcli share-link info https://www.dropbox.com/s/example/file.txt
  dbxcli share-link info https://www.dropbox.com/s/example/folder --path /nested/file.txt`,
	RunE: shareLinkInfo,
}

func init() {
	shareLinkInfoCmd.Flags().String("path", "", "Display metadata for a path inside the shared link")
	addSharedLinkPasswordFlags(shareLinkInfoCmd)
	shareLinkCmd.AddCommand(shareLinkInfoCmd)
	enableStructuredOutput(shareLinkInfoCmd)
}
