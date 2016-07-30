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

func shareListLinks(cmd *cobra.Command, args []string) (err error) {
	arg := sharing.NewListSharedLinksArg()

	dbx := sharing.New(config)
	res, err := dbx.ListSharedLinks(arg)
	if err != nil {
		return
	}

	printLinks(res.Links)

	for res.HasMore {
		arg = sharing.NewListSharedLinksArg()
		arg.Cursor = res.Cursor

		res, err = dbx.ListSharedLinks(arg)
		if err != nil {
			return
		}

		printLinks(res.Links)
	}

	return
}

func printLinks(links []sharing.IsSharedLinkMetadata) {
	for _, l := range links {
		switch sl := l.(type) {
		case *sharing.FileLinkMetadata:
			printLink(sl.SharedLinkMetadata)
		case *sharing.FolderLinkMetadata:
			printLink(sl.SharedLinkMetadata)
		default:
			fmt.Printf("found unknown shared link type")
		}
	}
}

func printLink(sl sharing.SharedLinkMetadata) {
	fmt.Printf("%v\t%v\n", sl.Name, sl.Url)
}

var shareListLinksCmd = &cobra.Command{
	Use:   "link",
	Short: "List shared links",
	RunE:  shareListLinks,
}

func init() {
	shareListCmd.AddCommand(shareListLinksCmd)
}
