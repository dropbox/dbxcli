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
	"os"

	"github.com/dropbox/dropbox-sdk-go-unofficial/files"
	"github.com/spf13/cobra"
)

func revs(cmd *cobra.Command, args []string) (err error) {
	path, err := parseDropboxUri(args[0])
	if err != nil {
		return
	}

	arg := files.NewListRevisionsArg()
	arg.Path = path

	res, err := dbx.ListRevisions(arg)
	if err != nil {
		return
	}

	long, _ := cmd.Flags().GetBool("long")

	if long {
		fmt.Printf("Revision\tSize\tLast modified\tPath\n")
	}

	for _, e := range res.Entries {
		if long {
			printFileMetadata(os.Stdout, e, long)
		} else {
			fmt.Printf("%s\n", e.Rev)
		}
	}

	return
}

// revsCmd represents the revs command
var revsCmd = &cobra.Command{
	Use:   "revs",
	Short: "List file revisions",
	RunE:  revs,
}

func init() {
	RootCmd.AddCommand(revsCmd)

	revsCmd.Flags().BoolP("long", "l", false, "Long listing")
}
