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
	pathmod "path"

	"github.com/dropbox/dropbox-sdk-go-unofficial/files"
	"github.com/spf13/cobra"
)

func removePath(rawpath string, recursive bool) (err error) {
	path, err := validatePath(rawpath)
	if err != nil {
		return
	}

	pathMetaData, err := getFileMetadata(path)
	if err != nil {
		return
	}
	if pathMetaData.Tag == folder {
		if !recursive {
			return fmt.Errorf("rm: cannot remove ‘%s’: Is a directory", path)
		}
		entries, err := ListFiles(path)
		if err != nil {
			return err
		}
		for _, e := range entries {
			switch e.Tag {
			case "folder":
				removePath(pathmod.Join(path, e.Folder.Name), recursive)
			case "file":
				removePath(pathmod.Join(path, e.File.Name), recursive)
			}
		}
	}

	arg := files.NewDeleteArg(path)

	if _, err = dbx.Delete(arg); err != nil {
		return
	}
	return nil
}

func rm(cmd *cobra.Command, args []string) (err error) {
	if len(args) != 1 {
		return errors.New("`rm` requires a `file` or `folder` argument")
	}
	recursive, _ := cmd.Flags().GetBool("recursive")
	return removePath(args[0], recursive)
}

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm [flags] <file>",
	Short: "Remove files",
	Example: `dbxcli rm <file>
	dbxcli rm -r <folder>`,
	RunE: rm,
}

func init() {
	RootCmd.AddCommand(rmCmd)
	rmCmd.Flags().BoolP("recursive", "r", false, "Recursive removal")
}
