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
	"os"
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/files"
	"github.com/spf13/cobra"
)

func search(cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 {
		return errors.New("`search` requires a `query` argument")
	}

	// Parse path scope, if provided.
	scope := ""
	if len(args) == 2 {
		scope = args[1]
		if len(scope) > 0 {
			if !strings.HasPrefix(scope, "/") {
				return errors.New("`search` `path-scope` must begin with \"/\"")
			}
		}
	}

	arg := files.NewSearchArg(scope, args[0])

	res, err := dbx.Search(arg)
	if err != nil {
		return
	}

	long, _ := cmd.Flags().GetBool("long")
	if long {
		fmt.Printf("Revision\tSize\tLast modified\tPath\n")
	}

	for _, m := range res.Matches {
		e := m.Metadata
		switch e.Tag {
		case folder:
			printFolderMetadata(os.Stdout, e.Folder, long)
		case file:
			printFileMetadata(os.Stdout, e.File, long)
		}
	}

	return
}

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [flags] <query> [path-scope]",
	Short: "Search",
	RunE:  search,
}

func init() {
	RootCmd.AddCommand(searchCmd)
	searchCmd.Flags().BoolP("long", "l", false, "Long listing")
}
