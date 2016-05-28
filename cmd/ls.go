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
	"strings"
	"text/tabwriter"

	"github.com/dropbox/dropbox-sdk-go-unofficial/files"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

func getFileMetadata(path string) (*files.Metadata, error) {
	arg := files.NewGetMetadataArg(path)

	res, err := dbx.GetMetadata(arg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func printFolderMetadata(w io.Writer, e *files.FolderMetadata, longFormat bool) {
	if longFormat {
		fmt.Fprintf(w, "-\t-\t-\t")
	}
	fmt.Fprintf(w, "%s\n", e.Name)
}

func printFileMetadata(w io.Writer, e *files.FileMetadata, longFormat bool) {
	if longFormat {
		fmt.Fprintf(w, "%s\t%s\t%s\t", e.Rev, humanize.Bytes(e.Size), humanize.Time(e.ServerModified))
	}
	fmt.Fprintf(w, "%s\n", e.Name)
}

func ls(cmd *cobra.Command, args []string) (err error) {
	path := ""
	if len(args) > 0 {
		if path, err = validatePath(args[0]); err != nil {
			return
		}
	}

	arg := files.NewListFolderArg(path)

	res, err := dbx.ListFolder(arg)
	var entries []*files.Metadata
	if err != nil {
		if strings.Contains(err.Error(), "path/not_folder/") {
			metaRes, _ := getFileMetadata(path)
			entries = []*files.Metadata{metaRes}
			err = nil
		} else {
			return err
		}
	} else {
		entries = res.Entries

		for res.HasMore {
			arg := files.NewListFolderContinueArg(res.Cursor)

			res, err = dbx.ListFolderContinue(arg)
			if err != nil {
				return
			}

			entries = append(entries, res.Entries...)
		}
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 8, 1, ' ', 0)
	long, _ := cmd.Flags().GetBool("long")
	if long {
		fmt.Fprintf(w, "Revision\tSize\tLast modified\tPath\n")
	}

	for _, e := range entries {
		switch e.Tag {
		case "folder":
			printFolderMetadata(w, e.Folder, long)
		case "file":
			printFileMetadata(w, e.File, long)
		}
	}
	w.Flush()

	return
}

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls [flags] [<path>]",
	Short: "List folders",
	Long: `List Folders.
	Attempting ls on files will fail with 'Error: path/not_folder/.'

	Examples:
	$ dbxcli ls / # Or, dbxcli ls
	$ dbxcli ls some-folder
	$ dbxcli ls /some-folder # Or dbxcli ls some-folder/
	$ dbxcli ls -l # Or, dbxcli ls --long
	`,
	RunE: ls,
}

func init() {
	RootCmd.AddCommand(lsCmd)

	lsCmd.Flags().BoolP("long", "l", false, "Long listing")
}
