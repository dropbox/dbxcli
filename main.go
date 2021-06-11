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

package main

import (
	"fmt"
	"log"

	"github.com/dropbox/dbxcli/cmd"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("dbxcli version:", version)
		sdkVersion, specVersion := dropbox.Version()
		fmt.Println("SDK version:", sdkVersion)
		fmt.Println("Spec version:", specVersion)
	},
}

func init() {
	// Log date, time and file information by default
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	cmd.RootCmd.AddCommand(versionCmd)
}

func main() {
	cmd.Execute()
}
