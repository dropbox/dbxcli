// Copyright © 2017 Dropbox, Inc.
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
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/sharing"
	"github.com/spf13/cobra"
)

// Try to get the share link for a file if it already exists.
// If it doesn't make a new share link for it.
func getShareLink(cmd *cobra.Command, args []string) (err error) {
	if len(args) != 1 {
		printShareLinkUsage()
		return
	}

	dbx := sharing.New(config)
	path, err := filepath.Abs(args[0])
	if err != nil {
		return
	}

	// Confirm that the file exists.
	exists, err := exists(path)
	if !exists || err != nil {
		fmt.Printf("The file / folder specified (\"%s\") does not exist.\n", path)
		return
	}

	// Try to get a link if it already exists.
	if getExistingLink(dbx, path) != nil {
		return
	}

	// The file had no share link, let's get it.
	getNewLink(dbx, path)

	return
}

func printShareLinkUsage() {
	fmt.Printf("Usage: %s share getlink [file / folder path]\n", os.Args[0])
}

// Try to get an existing share link for a file / folder.
// It returns true if the file / folder had a link. Otherwise it returns false.
func getExistingLink(dbx sharing.Client, path string) (err error) {
	// Remove the Dropbox folder from the start.
	path = strings.Replace(path, getDropboxFolder(), "", 1)

	arg := sharing.ListSharedLinksArg{Path: path}
	// This method can be called with a path and just get that share link.
	res, err := dbx.ListSharedLinks(&arg)
	if err != nil || len(res.Links) == 0 {
		return err
	}
	printLinks(res.Links)
	return nil
}

// Create and print a link for file / folder that doesn't yet have one.
// CreateSharedLinkWithSettings doesn't allow pending uploads,
// so we use the partially deprecated CreateSharedLink.
func getNewLink(dbx sharing.Client, path string) (err error) {
	// Determine whether the target is a file or folder.
	fi, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	arg := sharing.NewCreateSharedLinkArg(strings.Replace(path, getDropboxFolder(), "", 1))
	// Get the sharelink even if the file isn't fully uploaded yet.
	arg.PendingUpload = new(sharing.PendingUploadMode)
	switch mode := fi.Mode(); {
	case mode.IsDir():
		arg.PendingUpload.Tag = sharing.PendingUploadModeFolder
	case mode.IsRegular():
		arg.PendingUpload.Tag = sharing.PendingUploadModeFile
	}
	res, err := dbx.CreateSharedLink(arg)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return err
	}
	fmt.Printf("%s\t%s\n", res.Path[1:], res.Url)
	return nil
}

// Return the path of the Dropbox folder.
func getDropboxFolder() string {
	// I should be using a JSON parser here but it's a pain in Go.
	usr, _ := user.Current()
	homedir := usr.HomeDir
	infoFilePath := path.Join(homedir, ".dropbox/info.json")
	raw, err := ioutil.ReadFile(infoFilePath)
	if err != nil {
		print("Couldn't find Dropbox folder")
		return ""
	}
	// This is obviously dirty.
	return strings.Split(strings.Split(string(raw), "\"path\": \"")[1], "\"")[0]
}

// Check whether a file / folder exists.
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
