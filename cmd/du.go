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

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/users"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

func du(cmd *cobra.Command, args []string) (err error) {
	dbx := users.New(config)
	usage, err := dbx.GetSpaceUsage()
	if err != nil {
		return
	}

	fmt.Printf("Used: %s\n", humanize.IBytes(usage.Used))
	fmt.Printf("Type: %s\n", usage.Allocation.Tag)

	allocation := usage.Allocation

	switch allocation.Tag {
	case "individual":
		fmt.Printf("Allocated: %s\n", humanize.IBytes(allocation.Individual.Allocated))
	case "team":
		fmt.Printf("Allocated: %s (Used: %s)\n",
			humanize.IBytes(allocation.Team.Allocated),
			humanize.IBytes(allocation.Team.Used))
	}

	return
}

// duCmd represents the du command
var duCmd = &cobra.Command{
	Use:   "du",
	Short: "Display usage information",
	RunE:  du,
}

func init() {
	RootCmd.AddCommand(duCmd)
}
