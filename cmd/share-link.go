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

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/sharing"
	"github.com/spf13/cobra"
)

func shareLink(cmd *cobra.Command, args []string) (err error) {
	if len(args) != 1 {
		return errors.New("`link` requires `path`")
	}

	path := args[0]

	// TODO(bonafidehan): support other visibility.
	arg := sharing.NewCreateSharedLinkWithSettingsArg(path)

	dbx := sharing.New(config)
	res, err := dbx.CreateSharedLinkWithSettings(arg)
	if err != nil {
		return
	}

	switch sl := res.(type) {
	case *sharing.FileLinkMetadata:
		printLinkMetadata(sl.SharedLinkMetadata)
	case *sharing.FolderLinkMetadata:
		printLinkMetadata(sl.SharedLinkMetadata)
	}

	return
}

func printLinkMetadata(sl sharing.SharedLinkMetadata) {
	fmt.Printf("%v\t%v\n", sl.Name, sl.Url)
}

// searchCmd represents the search command
var shareLinkCmd = &cobra.Command{
	Use:   "link <path>",
	Short: "Create a shared link with public visibility",
	RunE:  shareLink,
}

func init() {
	shareCmd.AddCommand(shareLinkCmd)
}
