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
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/xattr"

	"github.com/spf13/cobra"

	gitignore "github.com/sabhiram/go-gitignore"
)

const (
	dropboxXattr        string = "com.dropbox.attrs"
	dropboxIgnoredXattr string = "com.dropbox.ignored"
	ignoreFilesToShow   int    = 7
)

func getGitIgnorePaths(root, ignoreFilePath string) ([]string, error) {
	gi, err := gitignore.CompileIgnoreFile(ignoreFilePath)
	if err != nil {
		return nil, err
	}

	var targetedFiles []string
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if gi.MatchesPath(path) && hasDropboxAttr(path) {
			targetedFiles = append(targetedFiles, path)
		}
		return nil
	})

	return targetedFiles, nil
}

func hasDropboxAttr(path string) bool {
	if _, err := os.Stat(path); err == nil {
		if _, err := xattr.Get(path, dropboxXattr); err == nil {
			return true
		}
	}
	return false
}

func ignoreToggle(cmd *cobra.Command, args []string) (err error) {

	isGitignore, _ := cmd.Flags().GetBool("gitignore")
	root := "."

	var targetedFiles []string

	if isGitignore {
		if f, err := os.Stat(args[0]); err == nil && !f.IsDir() {
			targetedFiles, err = getGitIgnorePaths(root, args[0])
			if err != nil {
				return errors.New("the given file must be a .gitignore style file")
			}
		} else {
			return errors.New("the given file must be an existing .gitignore style file")
		}
	} else {
		if _, err := os.Stat(args[0]); err == nil && hasDropboxAttr(args[0]) {
			targetedFiles = append(targetedFiles, args[0])
		}
	}

	if len(targetedFiles) == 0 {
		fmt.Println("No files found...")
		return
	}

	fmt.Println("Toggling ignore state on the following file(s):")
	for idx, path := range targetedFiles {
		if idx >= ignoreFilesToShow {
			break
		}
		fmt.Println("\t- ", path)
	}
	if len(targetedFiles) >= ignoreFilesToShow {
		fmt.Printf("And %d more...\n", len(targetedFiles)-ignoreFilesToShow)
	}

	toggled := 0
	for _, path := range targetedFiles {
		if _, err := xattr.Get(path, dropboxIgnoredXattr); err == nil {
			if err := xattr.Remove(path, dropboxIgnoredXattr); err == nil {
				toggled++
			} else {
				panic(err)
			}
		} else {
			if err := xattr.Set(path, dropboxIgnoredXattr, []byte{1}); err == nil {
				toggled++
			} else {
				panic(err)
			}
		}
	}

	return
}

// ignoreToggleCmd represents the ignoreToggle command
var ignoreToggleCmd = &cobra.Command{
	Use:   "toggle-ignore [flags] <file_path/gitignore_path>",
	Short: "Ignore a file from Dropbox Sync",
	Long:  "Fully ignore a local file from syncing with Dropbox or a set of files defined in a gitignore style file",
	Args:  cobra.ExactArgs(1),
	RunE:  ignoreToggle,
}

func init() {
	RootCmd.AddCommand(ignoreToggleCmd)
	ignoreToggleCmd.Flags().BoolP("gitignore", "g", false, "Toggle ignored files based on the contents of a .gitignore style file")
}
