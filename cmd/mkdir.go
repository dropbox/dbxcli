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

	"github.com/dropbox/dropbox-sdk-go-unofficial/files"
	"github.com/spf13/cobra"
)

func mkdir(cmd *cobra.Command, args []string) (err error) {
	if len(args) != 1 {
		return errors.New("`mkdir` requires a `directory` argument")
	}

	dst, err := validatePath(args[0])
	if err != nil {
		return
	}

	arg := files.NewCreateFolderArg(dst)

	dbx := files.New(config)
	if _, err = dbx.CreateFolder(arg); err != nil {
		return
	}

	return
}

// mkdirCmd represents the mkdir command
var mkdirCmd = &cobra.Command{
	Use:   "mkdir [flags] <directory>",
	Short: "Create a new directory",
	RunE:  mkdir,
}

func init() {
	RootCmd.AddCommand(mkdirCmd)
}
