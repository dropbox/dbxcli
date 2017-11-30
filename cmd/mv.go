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

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/spf13/cobra"
)

func mv(cmd *cobra.Command, args []string) error {
	var destination string
	var argsToMove []string

	if len(args) > 2 {
		destination = args[len(args)-1]
		argsToMove = args[0 : len(args)-1]
	} else if len(args) == 2 {
		destination = args[1]
		argsToMove = append(argsToMove, args[0])
	} else {
		return fmt.Errorf("mv command requires a source and a destination")
	}

	var mvErrors []error
	var relocationArgs []*files.RelocationArg

	for _, argument := range argsToMove {
		arg, err := makeRelocationArg(argument, destination+"/"+argument)
		if err != nil {
			relocationError := fmt.Errorf("Error validating move for %s to %s: %v", argument, destination, err)
			mvErrors = append(mvErrors, relocationError)
		} else {
			relocationArgs = append(relocationArgs, arg)
		}
	}

	dbx := files.New(config)
	for _, arg := range relocationArgs {
		if _, err := dbx.MoveV2(arg); err != nil {
			moveError := fmt.Errorf("Move error: %v", arg)
			mvErrors = append(mvErrors, moveError)
		}
	}

	for _, mvError := range mvErrors {
		fmt.Fprintf(os.Stderr, "%v\n", mvError)
	}

	return nil
}

// mvCmd represents the mv command
var mvCmd = &cobra.Command{
	Use:   "mv [flags] <source> <target>",
	Short: "Move files",
	RunE:  mv,
}

func init() {
	RootCmd.AddCommand(mvCmd)
}
