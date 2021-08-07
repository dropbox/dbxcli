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
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

const deletedItemFormatString = "<<%s>>"

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
	fmt.Fprintf(w, formatFolderMetadata(e, longFormat))
}

// Invoked by search.go and revs.go
func printFileMetadata(w io.Writer, e *files.FileMetadata, longFormat bool) {
	fmt.Fprintf(w, formatFileMetadata(e, longFormat))
}

func formatFolderMetadata(e *files.FolderMetadata, longFormat bool) string {
	text := fmt.Sprintf("%s\t", e.PathDisplay)
	if longFormat {
		text = fmt.Sprintf("-\t-\t-\t") + text
	}
	return text
}

func formatFileMetadata(e *files.FileMetadata, longFormat bool) string {
	text := fmt.Sprintf("%s\t", e.PathDisplay)
	if longFormat {
		text = fmt.Sprintf("%s\t%s\t%s\t", e.Rev, humanize.IBytes(e.Size), humanize.Time(e.ServerModified)) + text
	}
	return text
}

func formatDeletedMetadata(e *files.DeletedMetadata, longFormat bool) string {
	text := fmt.Sprintf("%s\t", e.PathDisplay)
	if longFormat {
		text = fmt.Sprintf("-\t-\t-\t") + text
	}
	return text
}

func SetPathDisplayAsDeleted(metadata files.IsMetadata) {
	switch item := metadata.(type) {
	case *files.FileMetadata:
		item.PathDisplay = fmt.Sprintf(deletedItemFormatString, item.PathDisplay)
	case *files.FolderMetadata:
		item.PathDisplay = fmt.Sprintf(deletedItemFormatString, item.PathDisplay)
	case *files.DeletedMetadata:
		item.PathDisplay = fmt.Sprintf(deletedItemFormatString, item.PathDisplay)
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
	long, _ := cmd.Flags().GetBool("long")

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 8, 1, ' ', 0)
	itemCounter := 0
	printItem := func(message string) {
		itemCounter = itemCounter + 1
		fmt.Fprint(w, message)
		if (itemCounter%4 == 0) || long {
			fmt.Fprintln(w)
		}
	}

	dbx := files.New(config)

	// check if given object exists
	var metaRes files.IsMetadata
	metaRes, err = getFileMetadata(dbx, path)
	if err != nil {
		return err
	}

	if long {
		fmt.Fprint(w, "Revision\tSize\tLast modified\tPath\n")
	}

	// handle case when path to file was given
	switch f := metaRes.(type) {
	case *files.FileMetadata:
		if !onlyDeleted {
			printItem(formatFileMetadata(f, long))
			err = w.Flush()
			return err
		}
	}

	res, err := dbx.ListFolder(arg)

	var entries []files.IsMetadata
	if err != nil {
		listRevisionError, ok := err.(files.ListRevisionsAPIError)
		if ok {
			// Don't treat a "not_folder" error as fatal; recover by sending a
			// get_metadata request for the same path and using that response instead.
			if listRevisionError.EndpointError.Path.Tag == files.LookupErrorNotFolder {
				var metaRes files.IsMetadata
				metaRes, err = getFileMetadata(dbx, path)
				entries = []files.IsMetadata{metaRes}
			} else {
				// Return if there's an error other than "not_folder" or if the follow-up
				// metadata request fails.
				return err
			}
		} else {
			return err
		}
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

	for _, entry := range entries {
		deletedItem, isDeleted := entry.(*files.DeletedMetadata)
		if isDeleted {
			revisionArg := files.NewListRevisionsArg(deletedItem.PathLower)
			res, err := dbx.ListRevisions(revisionArg)
			if err != nil {
				listRevisionError, ok := err.(files.ListRevisionsAPIError)
				if ok {
					// We have a ListRevisionsAPIERror
					if listRevisionError.EndpointError.Path.Tag == files.LookupErrorNotFile {
						// Don't treat a "not_file" error as fatal; recover by sending a
						// get_metadata request for the same path and using that response instead.
						revision, err := getFileMetadata(dbx, deletedItem.PathLower)
						if err != nil {
							return err
						}
						entry = revision
					}
				}
			} else if len(res.Entries) == 0 {
				// Occasionally revisions will be returned with an empty Revision entry list.
				// So we just use the original entry.
			} else {
				entry = res.Entries[0]
			}
			SetPathDisplayAsDeleted(entry)
		}
		switch f := entry.(type) {
		case *files.FileMetadata:
			if !onlyDeleted {
				printItem(formatFileMetadata(f, long))
			}
		case *files.FolderMetadata:
			if !onlyDeleted {
				printItem(formatFolderMetadata(f, long))
			}
		case *files.DeletedMetadata:
			printItem(formatDeletedMetadata(f, long))
		}
	}

	err = w.Flush()
	return err
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
}
