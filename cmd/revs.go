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

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/spf13/cobra"
)

func revs(cmd *cobra.Command, args []string) (err error) {
	if len(args) != 1 {
		return errors.New("`revs` requires a `file` argument")
	}

	path, err := validatePath(args[0])
	if err != nil {
		return
	}

	arg := files.NewListRevisionsArg(path)

	dbx := files.New(config)
	res, err := dbx.ListRevisions(arg)
	if err != nil {
		return
	}

	machineReadable, _ := cmd.Flags().GetBool("machine")
	long, _ := cmd.Flags().GetBool("long")

	//If machine is set imply long
	long = long || machineReadable

	if long {
		fmt.Printf("Revision\tSize\tLast modified\tPath\n")
	}

	for _, e := range res.Entries {
		if long {
			printFileMetadata(os.Stdout, e, long, machineReadable)
		} else {
			fmt.Printf("%s\n", e.Rev)
		}
	}

	return
}

// revsCmd represents the revs command
var revsCmd = &cobra.Command{
	Use:   "revs [flags] <file>",
	Short: "List file revisions",
	RunE:  revs,
}

func init() {
	RootCmd.AddCommand(revsCmd)

	revsCmd.Flags().BoolP("long", "l", false, "Long listing")
	revsCmd.Flags().BoolP("machine", "m", false, "Machine readable file size and time")
}
