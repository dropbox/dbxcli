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
	"github.com/spf13/cobra"
)

var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Sharing commands",
}

var shareListCmd = &cobra.Command{
	Use:   "list",
	Short: "List shared things",
}

var shareGetLinkCmd = &cobra.Command{
	Use:   "getlink",
	Short: "Get share link for file / folder (create if it doesn't exist)",
	RunE:  getShareLink,
}

func init() {
	RootCmd.AddCommand(shareCmd)
	shareCmd.AddCommand(shareListCmd)
	shareCmd.AddCommand(shareGetLinkCmd)
}
