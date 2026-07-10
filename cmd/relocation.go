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
	"io"
	"strings"

	"github.com/dropbox/dbxcli/v3/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

// relocationSpec captures the handful of values that differ between the mv and
// cp commands so they can share the single runRelocation driver below. The rest
// of the two commands (two-phase validate/execute loop, destination resolution,
// skip and conflict handling, dry-run preview, JSON accumulation) is identical.
type relocationSpec struct {
	command            string // "mv" / "cp"; used in the aggregate error
	verb               string // "move" / "copy"; used in per-operation error messages and dry-run text
	missingArgsMessage string // shown when fewer than two positional args are given
	successStatus      string // relocationJSONStatusMoved / relocationJSONStatusCopied
	execute            func(dbx filesClient, arg *files.RelocationArg) (*files.RelocationResult, error)
}

var moveSpec = relocationSpec{
	command:            "mv",
	verb:               "move",
	missingArgsMessage: "mv command requires a source and a destination",
	successStatus:      relocationJSONStatusMoved,
	execute: func(dbx filesClient, arg *files.RelocationArg) (*files.RelocationResult, error) {
		return dbx.MoveV2Context(currentContext(), arg)
	},
}

var copySpec = relocationSpec{
	command:            "cp",
	verb:               "copy",
	missingArgsMessage: "cp requires a source and a destination",
	successStatus:      relocationJSONStatusCopied,
	execute: func(dbx filesClient, arg *files.RelocationArg) (*files.RelocationResult, error) {
		return dbx.CopyV2Context(currentContext(), arg)
	},
}

// runRelocation implements the shared body of the mv and cp commands, using
// spec for the command-specific verb, status, error text, and API call.
func runRelocation(cmd *cobra.Command, args []string, spec relocationSpec) error {
	var destination string
	var argsToRelocate []string

	if len(args) > 2 {
		destination = args[len(args)-1]
		argsToRelocate = args[0 : len(args)-1]
	} else if len(args) == 2 {
		destination = args[1]
		argsToRelocate = append(argsToRelocate, args[0])
	} else {
		return invalidArgumentsErrorWithDetails(spec.missingArgsMessage, argumentsErrorDetails("source", "destination"))
	}

	opts, err := parseRelocationOptions(cmd)
	if err != nil {
		return err
	}

	var relocationErrors []error
	var relocationErrorDetails []map[string]any
	var relocationArgs []*files.RelocationArg
	var plannedResults []relocationResult
	var results []jsonOperationResult
	collectResults := commandOutputFormat(cmd) == output.FormatJSON

	dbx := filesNewFunc(config)
	destIsFolder := len(argsToRelocate) > 1 || strings.HasSuffix(destination, "/") || isRemoteFolder(dbx, destination)

	for _, argument := range argsToRelocate {
		dst := relocationDestination(argument, destination, destIsFolder)
		arg, err := makeRelocationArg(argument, dst)
		if err != nil {
			relocationErrors = append(relocationErrors, fmt.Errorf("Error validating %s for %s to %s: %v", spec.verb, argument, dst, err))
			relocationErrorDetails = append(relocationErrorDetails, relocationFailureDetails(argument, dst))
		} else {
			arg.Autorename = opts.ifExists == relocationIfExistsAutorename
			if opts.dryRun {
				result, err := plannedRelocationResult(dbx, arg)
				if err != nil {
					relocationErrors = append(relocationErrors, fmt.Errorf("%s %q to %q: %v", spec.verb, arg.FromPath, arg.ToPath, err))
					relocationErrorDetails = append(relocationErrorDetails, relocationFailureDetails(arg.FromPath, arg.ToPath))
					continue
				}
				plannedResults = append(plannedResults, result)
				if collectResults {
					results = append(results, relocationOperationResult(spec.successStatus, result))
				}
				continue
			}
			result, skipped, err := relocationSkipIfDestinationExists(dbx, arg, opts)
			if err != nil {
				relocationErrors = append(relocationErrors, fmt.Errorf("%s %q to %q: %v", spec.verb, arg.FromPath, arg.ToPath, err))
				relocationErrorDetails = append(relocationErrorDetails, relocationFailureDetails(arg.FromPath, arg.ToPath))
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
		res, err := spec.execute(dbx, arg)
		if err != nil {
			if result, skipped := relocationSkipAfterDestinationConflict(dbx, arg, err, opts); skipped {
				if collectResults {
					results = append(results, relocationOperationResult(relocationJSONStatusSkipped, result))
				}
				continue
			}
			relocationErrors = append(relocationErrors, fmt.Errorf("%s %q to %q: %v", spec.verb, arg.FromPath, arg.ToPath, err))
			relocationErrorDetails = append(relocationErrorDetails, relocationFailureDetails(arg.FromPath, arg.ToPath))
			continue
		}
		if collectResults {
			result, err := newRelocationResult(arg, res)
			if err != nil {
				relocationErrors = append(relocationErrors, fmt.Errorf("%s %q to %q: %v", spec.verb, arg.FromPath, arg.ToPath, err))
				relocationErrorDetails = append(relocationErrorDetails, relocationFailureDetails(arg.FromPath, arg.ToPath))
				continue
			}
			results = append(results, relocationOperationResult(relocationSuccessStatus(spec.successStatus, arg, result, opts), result))
		}
	}

	if len(relocationErrors) > 0 {
		for _, relocationError := range relocationErrors {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", relocationError)
		}
		return relocationAggregateError(spec.command, spec.verb, len(relocationErrors), relocationErrorDetails)
	}

	if opts.dryRun {
		return renderOperation(cmd, nil, results, nil, func(w io.Writer) error {
			return renderPlannedRelocationResults(w, spec.verb, plannedResults)
		})
	}

	if !collectResults {
		return nil
	}
	return renderJSONOperationOutput(cmd, nil, results)
}
