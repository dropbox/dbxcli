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
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/sharing"
	"github.com/spf13/cobra"
)

const DefaultDropboxName = "/Users/daniel/Dropbox"

/**
Try to get the share link for a file if it already exists.
If it doesn't make a new share link for it.
*/
func shareLink(cmd *cobra.Command, args []string) (err error) {
	if len(args) != 1 {
		printShareLinkUsage()
		return
	}
	dbx := sharing.New(config)
	path, err := filepath.Abs(args[0])
	if err != nil {
		return
	}
	// Remove the Dropbox folder from the start.
	path = strings.Replace(path, getDropboxFolder(), "", 1)

	// Try to get a link if it already exists.
	if getExistingLink(dbx, path) {
		return
	}

	// The file had no share link, let's get it.
	getNewLink(dbx, path)

	return
}

func printShareLinkUsage() {
	fmt.Printf("Usage: %s share createlink [file / folder path]\n", os.Args[0])
}

func getExistingLink(dbx sharing.Client, path string) bool {
	arg := sharing.ListSharedLinksArg{Path: path}
	// This method can be called with a path and just get that share link.
	res, err := dbx.ListSharedLinks(&arg)
	if err != nil || len(res.Links) == 0 {
		print("File / folder does not yet have a sharelink, creating one...\n")
	} else {
		printLinks(res.Links)
		return true
	}
	return false
}

func getNewLink(dbx sharing.Client, path string) bool {
	// CreateSharedLinkWithSettings is cooked, I won't use it.
	arg := sharing.NewCreateSharedLinkArg(path)
	res, err := dbx.CreateSharedLink(arg)
	if err != nil {
		return false
	}
	fmt.Printf("%s %s\n", res.Path[1:], res.Url)
	return true
}

func getDropboxFolder() string {
	// I should be using a JSON parser here but it's a pain in Go.
	usr, _ := user.Current()
	homedir := usr.HomeDir
	infoFilePath := path.Join(homedir, ".dropbox/info.json")
	raw, err := ioutil.ReadFile(infoFilePath)
	if err != nil {
		print("Couldn't find Dropbox folder")
		return DefaultDropboxName
	}
	// This is obviously dirty.
	return strings.Split(strings.Split(string(raw), "\"path\": \"")[1], "\"")[0]
}
