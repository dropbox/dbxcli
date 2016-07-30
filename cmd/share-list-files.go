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

func shareListFiles(cmd *cobra.Command, args []string) (err error) {
	arg := sharing.NewListFilesArg()

	dbx := sharing.New(config)
	res, err := dbx.ListReceivedFiles(arg)
	if err != nil {
		return
	}

	printFiles(res.Entries)

	for len(res.Cursor) > 0 {
		continueArg := sharing.NewListFilesContinueArg(res.Cursor)

		res, err = dbx.ListReceivedFilesContinue(continueArg)
		if err != nil {
			return
		}

		printFiles(res.Entries)
	}

	return
}

func printFiles(entries []*sharing.SharedFileMetadata) {
	for _, f := range entries {
		fmt.Printf("%v\t%v\n", f.Name, f.PreviewUrl)
	}
}

var shareListFilesCmd = &cobra.Command{
	Use:   "received-file",
	Short: "List received files (not including those in shared folders)",
	RunE:  shareListFiles,
}

func init() {
	shareListCmd.AddCommand(shareListFilesCmd)
}
