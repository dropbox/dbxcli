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
	"text/tabwriter"

	"github.com/dropbox/dbxcli/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

const deletedItemFormatString = "<<%s>>"

type lsInput struct {
	Path           string `json:"path"`
	Recursive      bool   `json:"recursive"`
	IncludeDeleted bool   `json:"include_deleted"`
	OnlyDeleted    bool   `json:"only_deleted"`
	Long           bool   `json:"long"`
	Sort           string `json:"sort,omitempty"`
	Reverse        bool   `json:"reverse"`
	Time           string `json:"time,omitempty"`
	TimeFormat     string `json:"time_format,omitempty"`
}

type lsOutput struct {
	Input   lsInput        `json:"input"`
	Entries []jsonMetadata `json:"entries"`
}

// Sends a get_metadata request for a given path and returns the response
func getFileMetadata(c files.Client, path string) (files.IsMetadata, error) {
	arg := files.NewGetMetadataArg(path)

	arg.IncludeDeleted = true

	res, err := c.GetMetadata(arg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Invoked by search.go
func printFolderMetadata(w io.Writer, e *files.FolderMetadata, longFormat bool) {
	_, _ = fmt.Fprint(w, formatFolderMetadata(e, longFormat))
}

func formatFolderMetadata(e *files.FolderMetadata, longFormat bool) string {
	text := fmt.Sprintf("%s\t", e.PathDisplay)
	if longFormat {
		text = "-\t-\t-\t" + text
	}
	return text
}

func formatFileMetadata(e *files.FileMetadata, longFormat bool) string {
	return formatFileMetadataWithOpts(e, listOptions{long: longFormat})
}

func formatFileMetadataWithOpts(e *files.FileMetadata, opts listOptions) string {
	text := fmt.Sprintf("%s\t", e.PathDisplay)
	if opts.long {
		t := getTime(e, opts)
		text = fmt.Sprintf("%s\t%s\t%s\t", e.Rev, humanize.IBytes(e.Size), formatTime(t, opts)) + text
	}
	return text
}

func formatDeletedMetadata(e *files.DeletedMetadata, longFormat bool) string {
	text := fmt.Sprintf("%s\t", e.PathDisplay)
	if longFormat {
		text = "-\t-\t-\t" + text
	}
	return text
}

func setPathDisplayAsDeleted(metadata files.IsMetadata) {
	switch item := metadata.(type) {
	case *files.FileMetadata:
		item.PathDisplay = fmt.Sprintf(deletedItemFormatString, item.PathDisplay)
	case *files.FolderMetadata:
		item.PathDisplay = fmt.Sprintf(deletedItemFormatString, item.PathDisplay)
	case *files.DeletedMetadata:
		item.PathDisplay = fmt.Sprintf(deletedItemFormatString, item.PathDisplay)
	}
}

func parseLsOptions(cmd *cobra.Command) listOptions {
	long, _ := cmd.Flags().GetBool("long")
	timeField, _ := cmd.Flags().GetString("time")
	timeFormat, _ := cmd.Flags().GetString("time-format")
	sortBy, _ := cmd.Flags().GetString("sort")
	reverse, _ := cmd.Flags().GetBool("reverse")
	return listOptions{
		long:       long,
		timeField:  timeField,
		timeFormat: timeFormat,
		sortBy:     sortBy,
		reverse:    reverse,
	}
}

func ls(cmd *cobra.Command, args []string) (err error) {

	path := ""
	if len(args) > 0 {
		if path, err = validatePath(args[0]); err != nil {
			return err
		}
	}

	arg := files.NewListFolderArg(path)
	arg.Recursive, _ = cmd.Flags().GetBool("recurse")
	arg.IncludeDeleted, _ = cmd.Flags().GetBool("include-deleted")
	onlyDeleted, _ := cmd.Flags().GetBool("only-deleted")
	arg.IncludeDeleted = arg.IncludeDeleted || onlyDeleted
	opts := parseLsOptions(cmd)

	dbx := filesNewFunc(config)

	var entries []files.IsMetadata
	if path != "" {
		var metaRes files.IsMetadata
		metaRes, err = getFileMetadata(dbx, path)
		if err != nil {
			return err
		}

		switch f := metaRes.(type) {
		case *files.FileMetadata:
			if !onlyDeleted {
				entries = []files.IsMetadata{f}
				return renderLsOutput(cmd, path, arg, onlyDeleted, opts, entries)
			}
		}
	}

	res, err := dbx.ListFolder(arg)

	if err != nil {
		if !isListFolderNotFolderError(err) {
			return err
		}
		// Don't treat a "not_folder" error as fatal; recover by sending a
		// get_metadata request for the same path and using that response instead.
		var metaRes files.IsMetadata
		metaRes, _ = getFileMetadata(dbx, path)
		entries = []files.IsMetadata{metaRes}
	} else {
		entries = res.Entries

		for res.HasMore {
			arg := files.NewListFolderContinueArg(res.Cursor)

			res, err = dbx.ListFolderContinue(arg)
			if err != nil {
				return err
			}

			entries = append(entries, res.Entries...)
		}
	}

	sortEntries(entries, opts)
	entries, err = prepareLsEntries(dbx, entries, onlyDeleted)
	if err != nil {
		return err
	}

	return renderLsOutput(cmd, path, arg, onlyDeleted, opts, entries)
}

func prepareLsEntries(dbx files.Client, entries []files.IsMetadata, onlyDeleted bool) ([]files.IsMetadata, error) {
	var filtered []files.IsMetadata
	for _, entry := range entries {
		deletedItem, isDeleted := entry.(*files.DeletedMetadata)
		if isDeleted {
			revisionArg := files.NewListRevisionsArg(deletedItem.PathLower)
			res, err := dbx.ListRevisions(revisionArg)
			if err != nil {
				if isListRevisionsNotFileError(err) {
					// Don't treat a "not_file" error as fatal; recover by sending a
					// get_metadata request for the same path and using that response instead.
					revision, err := getFileMetadata(dbx, deletedItem.PathLower)
					if err != nil {
						return nil, err
					}
					entry = revision
				}
			} else if len(res.Entries) == 0 {
				// Occasionally revisions will be returned with an empty Revision entry list.
				// So we just use the original entry.
			} else {
				entry = res.Entries[0]
			}
			setPathDisplayAsDeleted(entry)
		}
		switch entry.(type) {
		case *files.FileMetadata, *files.FolderMetadata:
			if !onlyDeleted {
				filtered = append(filtered, entry)
			}
		case *files.DeletedMetadata:
			filtered = append(filtered, entry)
		}
	}
	return filtered, nil
}

func renderLsOutput(cmd *cobra.Command, path string, arg *files.ListFolderArg, onlyDeleted bool, opts listOptions, entries []files.IsMetadata) error {
	out := commandOutput(cmd)
	if commandOutputFormat(cmd) != output.FormatJSON {
		return out.RenderText(func(w io.Writer) error {
			return renderLsResults(w, entries, opts)
		})
	}

	return out.Render(nil, lsOutput{
		Input:   newLsInput(path, arg, onlyDeleted, opts),
		Entries: jsonMetadataListFromLsEntries(entries),
	})
}

func newLsInput(path string, arg *files.ListFolderArg, onlyDeleted bool, opts listOptions) lsInput {
	displayPath := path
	if displayPath == "" {
		displayPath = "/"
	}
	return lsInput{
		Path:           displayPath,
		Recursive:      arg.Recursive,
		IncludeDeleted: arg.IncludeDeleted,
		OnlyDeleted:    onlyDeleted,
		Long:           opts.long,
		Sort:           opts.sortBy,
		Reverse:        opts.reverse,
		Time:           opts.timeField,
		TimeFormat:     opts.timeFormat,
	}
}

func jsonMetadataListFromLsEntries(entries []files.IsMetadata) []jsonMetadata {
	result := make([]jsonMetadata, 0, len(entries))
	for _, entry := range entries {
		result = append(result, jsonMetadataFromLsEntry(entry))
	}
	return result
}

func jsonMetadataFromLsEntry(entry files.IsMetadata) jsonMetadata {
	result := jsonMetadataFromDropbox(entry)
	if path, ok := undecoratedDeletedPath(result.PathDisplay); ok {
		result.PathDisplay = path
		result.Deleted = true
	}
	return result
}

func undecoratedDeletedPath(path string) (string, bool) {
	if strings.HasPrefix(path, "<<") && strings.HasSuffix(path, ">>") {
		return strings.TrimSuffix(strings.TrimPrefix(path, "<<"), ">>"), true
	}
	return path, false
}

func renderLsResults(out io.Writer, entries []files.IsMetadata, opts listOptions) error {
	w := new(tabwriter.Writer)
	w.Init(out, 4, 8, 1, ' ', 0)
	itemCounter := 0
	printItem := func(message string) {
		itemCounter = itemCounter + 1
		_, _ = fmt.Fprint(w, message)
		if (itemCounter%4 == 0) || opts.long {
			_, _ = fmt.Fprintln(w)
		}
	}

	if opts.long {
		_, _ = fmt.Fprint(w, "Revision\tSize\tLast modified\tPath\n")
	}

	for _, entry := range entries {
		switch f := entry.(type) {
		case *files.FileMetadata:
			printItem(formatFileMetadataWithOpts(f, opts))
		case *files.FolderMetadata:
			printItem(formatFolderMetadata(f, opts.long))
		case *files.DeletedMetadata:
			printItem(formatDeletedMetadata(f, opts.long))
		}
	}

	return finishListOutput(w, itemCounter, opts)
}

func isListFolderNotFolderError(err error) bool {
	var apiErr files.ListFolderAPIError
	return errors.As(err, &apiErr) &&
		apiErr.EndpointError != nil &&
		apiErr.EndpointError.Path != nil &&
		apiErr.EndpointError.Path.Tag == files.LookupErrorNotFolder
}

func isListRevisionsNotFileError(err error) bool {
	var apiErr files.ListRevisionsAPIError
	return errors.As(err, &apiErr) &&
		apiErr.EndpointError != nil &&
		apiErr.EndpointError.Path != nil &&
		apiErr.EndpointError.Path.Tag == files.LookupErrorNotFile
}

func finishListOutput(w *tabwriter.Writer, itemCounter int, opts listOptions) error {
	if itemCounter > 0 && !opts.long && itemCounter%4 != 0 {
		_, _ = fmt.Fprintln(w)
	}
	return w.Flush()
}

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls [flags] [<path>]",
	Short: "List files and folders",
	Example: `  dbxcli ls / # Or just 'ls'
  dbxcli ls /some-folder # Or 'ls some-folder'
  dbxcli ls /some-folder/some-file.pdf
  dbxcli ls -l`,
	RunE: ls,
}

func init() {
	RootCmd.AddCommand(lsCmd)

	lsCmd.Flags().BoolP("long", "l", false, "Long listing")
	lsCmd.Flags().BoolP("recurse", "R", false, "Recursively list all subfolders")
	lsCmd.Flags().BoolP("include-deleted", "d", false, "Include deleted files")
	lsCmd.Flags().BoolP("only-deleted", "D", false, "Only show deleted files")
	lsCmd.Flags().String("sort", "", "Sort by: name, size, time, type")
	lsCmd.Flags().BoolP("reverse", "r", false, "Reverse sort order")
	lsCmd.Flags().String("time", "server", "Time field: server, client")
	lsCmd.Flags().String("time-format", "", "Time format: short (2006-01-02 15:04), rfc3339")
	enableStructuredOutput(lsCmd)
}
