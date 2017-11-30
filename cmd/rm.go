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

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"

	"github.com/spf13/cobra"
)

func rm(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("rm: missing operand")
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	var deletePaths []string
	dbx := files.New(config)

	// Validate remove paths before executing removal
	for i := range args {
		path, err := validatePath(args[i])
		if err != nil {
			return err
		}

		pathMetaData, err := getFileMetadata(dbx, path)
		if err != nil {
			return err
		}

		if _, ok := pathMetaData.(*files.FileMetadata); !ok {
			folderArg := files.NewListFolderArg(path)
			res, err := dbx.ListFolder(folderArg)
			if err != nil {
				return err
			}
			if len(res.Entries) != 0 && !force {
				return fmt.Errorf("rm: cannot remove ‘%s’: Directory not empty, use `--force` or `-f` to proceed", path)
			}
		}
		deletePaths = append(deletePaths, path)
	}

	// Execute removals
	for _, path := range deletePaths {
		arg := files.NewDeleteArg(path)

		if _, err = dbx.DeleteV2(arg); err != nil {
			return err
		}
	}

	return nil
}

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm [flags] <file>",
	Short: "Remove files",
	RunE:  rm,
}

func init() {
	RootCmd.AddCommand(rmCmd)
	rmCmd.Flags().BoolP("force", "f", false, "Force removal")
}
