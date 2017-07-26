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
	"sort"
	"text/tabwriter"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

// Sends a get_metadata request for a given path and returns the response
func getFileMetadata(c files.Client, path string) (files.IsMetadata, error) {
	arg := files.NewGetMetadataArg(path)

	res, err := c.GetMetadata(arg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func printFolderMetadata(w io.Writer, e *files.FolderMetadata, longFormat bool) {
	if longFormat {
		fmt.Fprintf(w, "-\t-\t-\t")
	}
	fmt.Fprintf(w, "%s\t", e.PathDisplay)
}

func printFileMetadata(w io.Writer, e *files.FileMetadata, longFormat bool) {
	if longFormat {
		fmt.Fprintf(w, "%s\t%s\t%s\t", e.Rev, humanize.IBytes(e.Size), humanize.Time(e.ServerModified))
	}
	fmt.Fprintf(w, "%s\t", e.PathDisplay)
}

func ls(cmd *cobra.Command, args []string) (err error) {
	path := ""
	if len(args) > 0 {
		if path, err = validatePath(args[0]); err != nil {
			return err
		}
	}
	dbx := files.New(config)

	arg := files.NewListFolderArg(path)

	res, err := dbx.ListFolder(arg)
	var entries []files.IsMetadata
	if err != nil {
		switch e := err.(type) {
		case files.ListFolderAPIError:
			// Don't treat a "not_folder" error as fatal; recover by sending a
			// get_metadata request for the same path and using that response instead.
			if e.EndpointError.Path.Tag == files.LookupErrorNotFolder {
				var metaRes files.IsMetadata
				metaRes, err = getFileMetadata(dbx, path)
				entries = []files.IsMetadata{metaRes}
			} else {
				return err
			}
		default:
			return err
		}

		// Return if there's an error other than "not_folder" or if the follow-up
		// metadata request fails.
		if err != nil {
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

	long, _ := cmd.Flags().GetBool("long")
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 8, 1, ' ', 0)
	defer w.Flush()
	if long {
		fmt.Fprintf(w, "Revision\tSize\tLast modified\tPath\n")
	}
	for i, entry := range entries {
		switch f := entry.(type) {
		case *files.FileMetadata:
			printFileMetadata(w, f, long)
		case *files.FolderMetadata:
			printFolderMetadata(w, f, long)
		}
		if i%4 == 0 || long {
			fmt.Fprintln(w)
		}
	}

	return err
}

func listOfEntryNames(entries []files.IsMetadata) []string {
	listOfEntryNames := []string{}

	for _, entry := range entries {
		switch entry.(type) {
		case *files.FolderMetadata:
			listOfEntryNames = append(listOfEntryNames, entry.(*files.FolderMetadata).Name)
		case *files.FileMetadata:
			listOfEntryNames = append(listOfEntryNames, entry.(*files.FileMetadata).Name)
		}
	}

	sort.Strings(listOfEntryNames)
	return listOfEntryNames
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
}
