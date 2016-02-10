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

import "github.com/spf13/cobra"

func cp(cmd *cobra.Command, args []string) (err error) {
	arg, err := makeRelocationArg(args[0], args[1])
	if err != nil {
		return
	}

	if _, err = dbx.Copy(arg); err != nil {
		return
	}

	return
}

// cpCmd represents the cp command
var cpCmd = &cobra.Command{
	Use:   "cp",
	Short: "Copy files",
	RunE:  cp,
}

func init() {
	RootCmd.AddCommand(cpCmd)
}
