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
	"strings"

	"github.com/dropbox/dbxcli/v3/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
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
		return invalidArgumentsErrorWithDetails("cp requires a source and a destination", argumentsErrorDetails("source", "destination"))
	}

	opts, err := parseRelocationOptions(cmd)
	if err != nil {
		return err
	}

	var cpErrors []error
	var cpErrorDetails []map[string]any
	var relocationArgs []*files.RelocationArg
	var results []jsonOperationResult
	collectResults := commandOutputFormat(cmd) == output.FormatJSON

	dbx := filesNewFunc(config)
	destIsFolder := len(argsToCopy) > 1 || strings.HasSuffix(destination, "/") || isRemoteFolder(dbx, destination)

	for _, argument := range argsToCopy {
		dst := relocationDestination(argument, destination, destIsFolder)
		arg, err := makeRelocationArg(argument, dst)
		if err != nil {
			relocationError := fmt.Errorf("Error validating copy for %s to %s: %v", argument, dst, err)
			cpErrors = append(cpErrors, relocationError)
			cpErrorDetails = append(cpErrorDetails, relocationFailureDetails(argument, dst))
		} else {
			result, skipped, err := relocationSkipIfDestinationExists(dbx, arg, opts)
			if err != nil {
				cpErrors = append(cpErrors, fmt.Errorf("copy %q to %q: %v", arg.FromPath, arg.ToPath, err))
				cpErrorDetails = append(cpErrorDetails, relocationFailureDetails(arg.FromPath, arg.ToPath))
				continue
			}
			if skipped {
				if collectResults {
					results = append(results, relocationOperationResult(relocationJSONStatusSkipped, result))
				}
				continue
			}
			relocationArgs = append(relocationArgs, arg)
		}
	}

	for _, arg := range relocationArgs {
		res, err := dbx.CopyV2Context(currentContext(), arg)
		if err != nil {
			if result, skipped := relocationSkipAfterDestinationConflict(dbx, arg, err, opts); skipped {
				if collectResults {
					results = append(results, relocationOperationResult(relocationJSONStatusSkipped, result))
				}
				continue
			}
			copyError := fmt.Errorf("copy %q to %q: %v", arg.FromPath, arg.ToPath, err)
			cpErrors = append(cpErrors, copyError)
			cpErrorDetails = append(cpErrorDetails, relocationFailureDetails(arg.FromPath, arg.ToPath))
			continue
		}
		if collectResults {
			result, err := newRelocationResult(arg, res)
			if err != nil {
				copyError := fmt.Errorf("copy %q to %q: %v", arg.FromPath, arg.ToPath, err)
				cpErrors = append(cpErrors, copyError)
				cpErrorDetails = append(cpErrorDetails, relocationFailureDetails(arg.FromPath, arg.ToPath))
				continue
			}
			results = append(results, relocationOperationResult(relocationJSONStatusCopied, result))
		}
	}

	if len(cpErrors) > 0 {
		for _, cpError := range cpErrors {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", cpError)
		}
		return relocationAggregateError("cp", "copy", len(cpErrors), cpErrorDetails)
	}

	if !collectResults {
		return nil
	}
	return renderJSONOperationOutput(cmd, nil, results)
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
	enableStructuredOutput(cpCmd)
	cpCmd.Flags().String("if-exists", relocationIfExistsFail, "What to do when the destination exists: fail or skip")
}
