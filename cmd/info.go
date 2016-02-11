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
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func info(cmd *cobra.Command, args []string) (err error) {
	res, err := dbx.GetInfo()
	if err != nil {
		return err
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 8, 1, ' ', 0)
	fmt.Fprintf(w, "Name:\t%s\n", res.Name)
	fmt.Fprintf(w, "Team Id:\t%s\n", res.TeamId)
	fmt.Fprintf(w, "Licensed Users:\t%d\n", res.NumLicensedUsers)
	fmt.Fprintf(w, "Provisioned Users:\t%d\n", res.NumProvisionedUsers)
	w.Flush()
	return
}

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get team information",
	RunE:  info,
}

func init() {
	teamCmd.AddCommand(infoCmd)
}
