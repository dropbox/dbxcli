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

	"github.com/dropbox/dropbox-sdk-go/files"
	"github.com/spf13/cobra"
)

func printFolderMetadata(e *files.FolderMetadata, longFormat bool) {
	if longFormat {
		fmt.Printf("-\t\t-\t-\t\t")
	}
	fmt.Printf("%s\n", e.Name)
}

func printFileMetadata(e *files.FileMetadata, longFormat bool) {
	if longFormat {
		fmt.Printf("%s\t%s\t%s\t", e.Rev, humanizeSize(e.Size), humanizeDate(e.ServerModified))
	}
	fmt.Printf("%s\n", e.Name)
}

func ls(cmd *cobra.Command, args []string) (err error) {
	path := ""
	if len(args) > 0 {
		if path, err = parseDropboxUri(args[0]); err != nil {
			return
		}
	}

	arg := files.NewListFolderArg()
	arg.Path = path

	res, err := dbx.ListFolder(arg)
	if err != nil {
		return
	}

	entries := res.Entries

	for res.HasMore {
		arg := files.NewListFolderContinueArg()
		arg.Cursor = res.Cursor

		res, err = dbx.ListFolderContinue(arg)
		if err != nil {
			return
		}

		entries = append(entries, res.Entries...)
	}

	long, _ := cmd.Flags().GetBool("long")
	if long {
		fmt.Printf("Revision\tSize\tLast modified\tPath\n")
	}

	for _, e := range entries {
		switch e.Tag {
		case "folder":
			printFolderMetadata(e.Folder, long)
		case "file":
			printFileMetadata(e.File, long)
		}
	}

	return
}

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List files",
	RunE:  ls,
}

func init() {
	RootCmd.AddCommand(lsCmd)

	lsCmd.Flags().BoolP("long", "l", false, "Long listing")
}
