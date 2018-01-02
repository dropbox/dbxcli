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
	"path/filepath"
	"reflect"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/sharing"
	"github.com/spf13/cobra"
)

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

	// TODO: Remove the /Users/Dropbox part from the path
	path, err := filepath.Abs(args[0])
	if err != nil {
		return
	}

	arg := sharing.ListSharedLinksArg{Path: path}
	// This method can be called with a path and just get that share link.
	res, err := dbx.ListSharedLinks(&arg)
	if err != nil || len(res.Links) == 0 {
		print("File / folder does not yet have a sharelink, creating one...\n")
	} else {
		fmt.Printf("%+v\n", res)
		printLinks(res.Links)
		return
	}

	// The file had no share link, let's get it.
	arg2 := sharing.NewCreateSharedLinkWithSettingsArg(path)
	res2, err2 := dbx.CreateSharedLinkWithSettings(arg2)
	if err2 != nil {
		return
	}

	//fmt.Printf("%+v\n", res2)
	print(reflect.TypeOf(&res2).String())

	/*
		printLinks(res.Links)

		for res.HasMore {
			arg = sharing.NewListSharedLinksArg()
			arg.Cursor = res.Cursor

			res, err = dbx.ListSharedLinks(arg)
			if err != nil {
				return
			}

			printLinks(res.Links)
		}
	*/

	return
}

func printShareLinkUsage() {
	fmt.Printf("Usage: %s share createlink [file / folder path]\n", os.Args[0])
}
