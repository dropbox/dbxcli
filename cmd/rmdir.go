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

	"github.com/dropbox/dropbox-sdk-go-unofficial/files"
	"github.com/spf13/cobra"
)

func rmdir(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("`rmdir` requires a `directory` argument")
	}

	path, err := validatePath(args[0])
	if err != nil {
		return err
	}

	pathMetaData, err := getFileMetadata(path)
	if err != nil {
		return err
	}
	if pathMetaData.Tag != folder {
		return fmt.Errorf("rmdir: failed to remove ‘%s’: Not a directory", path)
	}

	arg := files.NewDeleteArg(path)

	if _, err = dbx.Delete(arg); err != nil {
		return err
	}

	return nil
}

// rmdirCmd represents the rmdir command
var rmdirCmd = &cobra.Command{
	Use:   "rmdir <directory>",
	Short: "Remove directory",
	RunE:  rmdir,
}

func init() {
	RootCmd.AddCommand(rmdirCmd)
}
