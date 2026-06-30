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
		return invalidArgumentsErrorWithDetails("mv command requires a source and a destination", argumentsErrorDetails("source", "destination"))
	}

	opts, err := parseRelocationOptions(cmd)
	if err != nil {
		return err
	}

	var mvErrors []error
	var mvErrorDetails []map[string]any
	var relocationArgs []*files.RelocationArg
	var results []jsonOperationResult
	collectResults := commandOutputFormat(cmd) == output.FormatJSON

	dbx := filesNewFunc(config)
	destIsFolder := len(argsToMove) > 1 || strings.HasSuffix(destination, "/") || isRemoteFolder(dbx, destination)

	for _, argument := range argsToMove {
		dst := relocationDestination(argument, destination, destIsFolder)
		arg, err := makeRelocationArg(argument, dst)
		if err != nil {
			mvErrors = append(mvErrors, fmt.Errorf("Error validating move for %s to %s: %v", argument, dst, err))
			mvErrorDetails = append(mvErrorDetails, relocationFailureDetails(argument, dst))
		} else {
			result, skipped, err := relocationSkipIfDestinationExists(dbx, arg, opts)
			if err != nil {
				mvErrors = append(mvErrors, fmt.Errorf("move %q to %q: %v", arg.FromPath, arg.ToPath, err))
				mvErrorDetails = append(mvErrorDetails, relocationFailureDetails(arg.FromPath, arg.ToPath))
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
		res, err := dbx.MoveV2(arg)
		if err != nil {
			if result, skipped := relocationSkipAfterDestinationConflict(dbx, arg, err, opts); skipped {
				if collectResults {
					results = append(results, relocationOperationResult(relocationJSONStatusSkipped, result))
				}
				continue
			}
			moveError := fmt.Errorf("move %q to %q: %v", arg.FromPath, arg.ToPath, err)
			mvErrors = append(mvErrors, moveError)
			mvErrorDetails = append(mvErrorDetails, relocationFailureDetails(arg.FromPath, arg.ToPath))
			continue
		}
		if collectResults {
			result, err := newRelocationResult(arg, res)
			if err != nil {
				moveError := fmt.Errorf("move %q to %q: %v", arg.FromPath, arg.ToPath, err)
				mvErrors = append(mvErrors, moveError)
				mvErrorDetails = append(mvErrorDetails, relocationFailureDetails(arg.FromPath, arg.ToPath))
				continue
			}
			results = append(results, relocationOperationResult(relocationJSONStatusMoved, result))
		}
	}

	if len(mvErrors) > 0 {
		for _, mvError := range mvErrors {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", mvError)
		}
		return relocationAggregateError("mv", "move", len(mvErrors), mvErrorDetails)
	}

	if !collectResults {
		return nil
	}
	return renderJSONOperationOutput(cmd, nil, results)
}

// mvCmd represents the mv command
var mvCmd = &cobra.Command{
	Use:   "mv [flags] <source> [more sources] <target>",
	Short: "Move files",
	RunE:  mv,
}

func init() {
	RootCmd.AddCommand(mvCmd)
	enableStructuredOutput(mvCmd)
	mvCmd.Flags().String("if-exists", relocationIfExistsFail, "What to do when the destination exists: fail or skip")
}
