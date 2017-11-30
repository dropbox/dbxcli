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
	"os"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/spf13/cobra"
)

func cp(cmd *cobra.Command, args []string) error {
	var destination string
	var argsToCopy []string

	if len(args) > 2 {
		destination = args[len(args)-1]
		argsToCopy = args[0 : len(args)-1]
	} else if len(args) == 2 {
		destination = args[1]
		argsToCopy = append(argsToCopy, args[0])
	} else {
		return errors.New("cp requires a source and a destination")
	}

	var cpErrors []error
	var relocationArgs []*files.RelocationArg

	for _, argument := range argsToCopy {
		arg, err := makeRelocationArg(argument, destination+"/"+argument)
		if err != nil {
			relocationError := fmt.Errorf("Error validating copy for %s to %s: %v", argument, destination, err)
			cpErrors = append(cpErrors, relocationError)
		} else {
			relocationArgs = append(relocationArgs, arg)
		}
	}

	dbx := files.New(config)
	for _, arg := range relocationArgs {
		if _, err := dbx.CopyV2(arg); err != nil {
			copyError := fmt.Errorf("Copy error: %v", arg)
			cpErrors = append(cpErrors, copyError)
		}
	}

	for _, cpError := range cpErrors {
		fmt.Fprintf(os.Stderr, "%v\n", cpError)
	}

	return nil
}

// cpCmd represents the cp command
var cpCmd = &cobra.Command{
	Use: "cp [flags] <source> [more sources] <target>",
	Short: "Copy a file or folder to a different location in the user's Dropbox. " +
		"If the source path is a folder all its contents will be copied.",
	RunE: cp,
}

func init() {
	RootCmd.AddCommand(cpCmd)
}
