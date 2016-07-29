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

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/sharing"
	"github.com/spf13/cobra"
)

func shareListFolders(cmd *cobra.Command, args []string) (err error) {
	arg := sharing.NewListFoldersArgs()

	dbx := sharing.New(config)
	res, err := dbx.ListFolders(arg)
	if err != nil {
		return
	}

	printFolders(res.Entries)

	for len(res.Cursor) > 0 {
		continueArg := sharing.NewListFoldersContinueArg(res.Cursor)

		res, err = dbx.ListFoldersContinue(continueArg)
		if err != nil {
			return
		}

		printFolders(res.Entries)
	}

	return
}

func printFolders(entries []*sharing.SharedFolderMetadata) {
	for _, f := range entries {
		fmt.Printf("%v\t%v\n", f.PathLower, f.PreviewUrl)
	}
}

var shareListFoldersCmd = &cobra.Command{
	Use:   "list-folders",
	Short: "List shared folders",
	RunE:  shareListFolders,
}

func init() {
	shareCmd.AddCommand(shareListFoldersCmd)
}
