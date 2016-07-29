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

	// TODO(bonafidehan): handle paging. Currently uses default limit of 1000.

	for _, f := range res.Entries {
		fmt.Printf("%v\n", f.PathLower)
	}

	return
}

var shareListCmd = &cobra.Command{
	Use:   "list-folders",
	Short: "List shared folders",
	RunE:  shareListFolders,
}

func init() {
	shareCmd.AddCommand(shareListCmd)
}
