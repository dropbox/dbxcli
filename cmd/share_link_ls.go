// Copyright © 2026 Dropbox, Inc.
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
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

type shareLinkLSInput struct {
	URL       string `json:"url"`
	Path      string `json:"path,omitempty"`
	Recursive bool   `json:"recursive"`
	Password  bool   `json:"password"`
}

type shareLinkLSEntry struct {
	Name string  `json:"name"`
	Path string  `json:"path"`
	Size *uint64 `json:"size,omitempty"`
}

func shareLinkLS(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return invalidArgumentsErrorWithDetails("`share-link ls` requires a `url` argument", argumentErrorDetails("url"))
	}
	return withJSONErrorDetails(listSharedLink(cmd, args[0]), urlErrorDetails(args[0]), operationErrorDetails("share_link_ls"))
}

func listSharedLink(cmd *cobra.Command, linkURL string) error {
	parsed, err := url.Parse(linkURL)
	if err != nil || parsed.Hostname() == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.User != nil {
		return invalidArgumentsErrorWithDetails("`share-link ls` requires a valid HTTP or HTTPS URL", argumentErrorDetails("url"))
	}
	input := shareLinkLSInput{URL: linkURL}
	relFolder := ""
	if localFlagChanged(cmd, "path") {
		value, err := localStringFlag(cmd, "path")
		if err != nil {
			return err
		}
		if value == "" {
			return invalidArgumentsErrorWithDetails("`--path` requires a non-empty path", flagErrorDetails("path"))
		}
		// Validate each component before normalization, using the same name rules as downloads.
		if value != "/" {
			for _, name := range strings.Split(strings.TrimPrefix(value, "/"), "/") {
				relFolder, err = sharedLinkEntryRelativePath(relFolder, name)
				if err != nil {
					return invalidArgumentsErrorWithDetails(err.Error(), mergeJSONErrorDetails(flagErrorDetails("path"), pathErrorDetails(value)))
				}
			}
		}
		input.Path = sharedLinkAPIPath(relFolder)
		if input.Path == "" {
			input.Path = "/"
		}
	}
	input.Recursive, err = localBoolFlag(cmd, "recursive")
	if err != nil {
		return err
	}
	password, err := sharedLinkPasswordFromFlags(cmd)
	if err != nil {
		return err
	}
	input.Password = password.set
	arg := sharing.NewGetSharedLinkMetadataArg(linkURL)
	arg.Path = sharedLinkAPIPath(relFolder)
	arg.LinkPassword = password.password
	if err := currentContext().Err(); err != nil {
		return err
	}
	link, err := newSharedLinkClient(config).GetSharedLinkMetadataContext(currentContext(), arg)
	if err != nil {
		return withJSONErrorDetails(err, pathErrorDetails(arg.Path))
	}
	if folder, ok := link.(*sharing.FolderLinkMetadata); !ok || folder == nil {
		return invalidArgumentsErrorWithDetails("shared link path is not a folder", pathErrorDetails(arg.Path))
	}
	results, err := listSharedLinkContents(filesNewFunc(config), arg, relFolder, input.Recursive)
	if err != nil {
		return err
	}
	return commandOutput(cmd).Render(func(out io.Writer) error {
		w := tabwriter.NewWriter(out, 4, 8, 1, ' ', 0)
		if _, err := fmt.Fprintln(w, "TYPE\tSIZE\tPATH"); err != nil {
			return err
		}
		for _, result := range results {
			entry := result.Result.(shareLinkLSEntry)
			size := "-"
			if entry.Size != nil {
				size = humanize.IBytes(*entry.Size)
			}
			if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", result.Kind, size, entry.Path); err != nil {
				return err
			}
		}
		return w.Flush()
	}, newJSONCommandOperationOutput(cmd, input, results, nil))
}

func listSharedLinkContents(dbx filesClient, arg *sharing.GetSharedLinkMetadataArg, start string, recursive bool) ([]jsonOperationResult, error) {
	results := []jsonOperationResult{}
	queue := []string{start}
	for len(queue) > 0 {
		if err := currentContext().Err(); err != nil {
			return nil, err
		}
		parent := queue[0]
		queue[0] = ""
		queue = queue[1:]
		entries, err := listSharedLinkFolderEntries(dbx, arg, parent)
		if err != nil {
			return nil, withJSONErrorDetails(err, pathErrorDetails(sharedLinkAPIPath(parent)))
		}
		children := []string{}
		for _, metadata := range entries {
			if err := currentContext().Err(); err != nil {
				return nil, err
			}
			var entry shareLinkLSEntry
			kind := ""
			switch item := metadata.(type) {
			case *files.FileMetadata:
				if item != nil {
					size := item.Size
					entry.Name, entry.Size, kind = item.Name, &size, "file"
				}
			case *files.FolderMetadata:
				if item != nil {
					entry.Name, kind = item.Name, "folder"
				}
			}
			if kind == "" {
				return nil, fmt.Errorf("unknown shared link entry type in %q", sharedLinkAPIPath(parent))
			}
			rel, err := sharedLinkEntryRelativePath(parent, entry.Name)
			if err != nil {
				return nil, withJSONErrorDetails(err, pathErrorDetails(sharedLinkAPIPath(parent)))
			}
			entry.Path = sharedLinkAPIPath(rel)
			if recursive && kind == "folder" {
				children = append(children, rel)
			}
			results = append(results, newJSONOperationResult(shareLinkJSONStatusListed, kind, nil, entry))
		}
		sort.Strings(children)
		queue = append(queue, children...)
	}
	// Keep only compact output records across folders; SDK pages are released after each folder.
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Result.(shareLinkLSEntry).Path < results[j].Result.(shareLinkLSEntry).Path
	})
	if err := currentContext().Err(); err != nil {
		return nil, err
	}
	return results, nil
}

var shareLinkLSCmd = &cobra.Command{
	Use:   "ls <url>",
	Short: "List the contents of a folder shared link",
	Long: `List files and folders inside a Dropbox folder shared link, including folders
not mounted in your Dropbox. Paths are relative to the shared-link root and can
be passed to share-link info --path or share-link download --path.
Use --path to start at a subfolder and --recursive (-r) to include its descendants.
Use share-link list to list shared links belonging to your Dropbox.`,
	Example: `  dbxcli share-link ls https://www.dropbox.com/s/example/folder
  dbxcli share-link ls https://www.dropbox.com/s/example/folder --path /documents --recursive
  dbxcli share-link ls https://www.dropbox.com/s/example/folder --output json`,
	RunE: shareLinkLS,
}

func init() {
	addSharedLinkPasswordFlags(shareLinkLSCmd)
	shareLinkLSCmd.Flags().String("path", "", "List a subfolder inside the shared link")
	shareLinkLSCmd.Flags().BoolP("recursive", "r", false, "Recursively list folder contents")
	shareLinkCmd.AddCommand(shareLinkLSCmd)
	enableStructuredOutput(shareLinkLSCmd)
}
