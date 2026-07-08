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

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"

	"github.com/spf13/cobra"
)

type removeOptions struct {
	force     bool
	recursive bool
	permanent bool
	dryRun    bool
	verbose   bool
}

type removeTarget struct {
	path     string
	metadata files.IsMetadata
}

type removeInput struct {
	Path      string `json:"path"`
	Permanent bool   `json:"permanent"`
	Recursive bool   `json:"recursive"`
	Force     bool   `json:"force"`
	DryRun    bool   `json:"dry_run,omitempty"`
}

type removeResult struct {
	Input  removeInput  `json:"input"`
	Result jsonMetadata `json:"result"`
}

const (
	removeJSONStatusDeleted            = "deleted"
	removeJSONStatusPermanentlyDeleted = "permanently_deleted"
)

func rm(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return invalidArgumentsErrorWithDetails("rm: missing operand", argumentErrorDetails("path"))
	}

	opts, err := parseRemoveOptions(cmd)
	if err != nil {
		return err
	}

	dbx := filesNewFunc(config)

	targets, err := validateRemoveTargets(dbx, args, opts)
	if err != nil {
		return err
	}

	results, err := removeTargets(dbx, targets, opts)
	if err != nil {
		return err
	}

	return commandOutput(cmd).Render(func(w io.Writer) error {
		if !opts.dryRun && !opts.verbose {
			return nil
		}
		return renderRemoveResults(w, results)
	}, newJSONCommandOperationOutput(cmd, nil, removeOperationResults(results), nil))
}

func removeOperationResults(results []removeResult) []jsonOperationResult {
	operationResults := make([]jsonOperationResult, 0, len(results))
	for _, result := range results {
		operationResults = append(operationResults, newJSONOperationResult(removeJSONStatus(result), result.Result.Type, result.Input, result.Result))
	}
	return operationResults
}

func removeJSONStatus(result removeResult) string {
	if result.Input.Permanent {
		return plannedStatus(result.Input.DryRun, removeJSONStatusPermanentlyDeleted)
	}
	return plannedStatus(result.Input.DryRun, removeJSONStatusDeleted)
}

func parseRemoveOptions(cmd *cobra.Command) (removeOptions, error) {
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return removeOptions{}, err
	}

	recursive, err := cmd.Flags().GetBool("recursive")
	if err != nil {
		return removeOptions{}, err
	}

	permanent, err := cmd.Flags().GetBool("permanent")
	if err != nil {
		return removeOptions{}, err
	}

	dryRun, err := dryRunEnabled(cmd)
	if err != nil {
		return removeOptions{}, err
	}

	verbose, _ := cmd.Flags().GetBool("verbose")

	return removeOptions{
		force:     force,
		recursive: recursive,
		permanent: permanent,
		dryRun:    dryRun,
		verbose:   verbose,
	}, nil
}

func validateRemoveTargets(dbx filesClient, args []string, opts removeOptions) ([]removeTarget, error) {
	var targets []removeTarget

	// Validate remove paths before executing removal
	for i := range args {
		path, err := validatePath(args[i])
		if err != nil {
			return nil, err
		}

		pathMetaData, err := getFileMetadata(dbx, path)
		if err != nil {
			return nil, withJSONErrorDetails(err, operationErrorDetails(removeOperation(opts)), pathErrorDetails(path))
		}

		if _, ok := pathMetaData.(*files.FileMetadata); !ok && !opts.allowNonEmptyFolder() {
			folderArg := files.NewListFolderArg(path)
			res, err := dbx.ListFolderContext(currentContext(), folderArg)
			if err != nil {
				return nil, withJSONErrorDetails(err, operationErrorDetails(removeOperation(opts)), pathErrorDetails(path))
			}
			if len(res.Entries) != 0 {
				return nil, invalidArgumentsErrorfWithDetails("rm: cannot remove ‘%s’: Directory not empty, use `--force`/`-f` or `--recursive`/`-r` to proceed", mergeJSONErrorDetails(operationErrorDetails(removeOperation(opts)), pathErrorDetails(path)), path)
			}
		}
		targets = append(targets, removeTarget{path: path, metadata: pathMetaData})
	}

	return targets, nil
}

func removeTargets(dbx filesClient, targets []removeTarget, opts removeOptions) ([]removeResult, error) {
	results := make([]removeResult, 0, len(targets))

	for _, target := range targets {
		arg := files.NewDeleteArg(target.path)
		metadata := target.metadata

		if !opts.dryRun {
			if opts.permanent {
				if err := dbx.PermanentlyDeleteContext(currentContext(), arg); err != nil {
					return nil, withJSONErrorDetails(err, operationErrorDetails(removeOperation(opts)), pathErrorDetails(target.path))
				}
			} else {
				res, err := dbx.DeleteV2Context(currentContext(), arg)
				if err != nil {
					return nil, withJSONErrorDetails(err, operationErrorDetails(removeOperation(opts)), pathErrorDetails(target.path))
				}
				if res != nil && res.Metadata != nil {
					metadata = res.Metadata
				}
			}
		}

		result, err := newRemoveResult(target.path, metadata, opts)
		if err != nil {
			return nil, withJSONErrorDetails(err, operationErrorDetails(removeOperation(opts)), pathErrorDetails(target.path))
		}
		results = append(results, result)
	}

	return results, nil
}

func newRemoveResult(path string, metadata files.IsMetadata, opts removeOptions) (removeResult, error) {
	result, err := removeMetadataFromDropbox(path, metadata)
	if err != nil {
		return removeResult{}, err
	}
	return removeResult{
		Input: removeInput{
			Path:      path,
			Permanent: opts.permanent,
			Recursive: opts.recursive,
			Force:     opts.force,
			DryRun:    opts.dryRun,
		},
		Result: result,
	}, nil
}

func removeMetadataFromDropbox(path string, metadata files.IsMetadata) (jsonMetadata, error) {
	result, err := jsonMetadataFromDropbox(metadata)
	if err != nil {
		return jsonMetadata{}, err
	}
	result.PathDisplay = metadataDisplayPath(path, result.PathDisplay)
	return result, nil
}

func renderRemoveResults(w io.Writer, results []removeResult) error {
	for _, result := range results {
		if result.Input.DryRun {
			if result.Input.Permanent {
				if err := writeDryRunLine(w, "permanently delete", result.displayPath()); err != nil {
					return err
				}
				continue
			}
			if err := writeDryRunLine(w, "delete", result.displayPath()); err != nil {
				return err
			}
			continue
		}
		if result.Input.Permanent {
			if _, err := fmt.Fprintf(w, "Permanently deleted %s\n", result.displayPath()); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintf(w, "Deleted %s\n", result.displayPath()); err != nil {
			return err
		}
	}
	return nil
}

func (r removeResult) displayPath() string {
	return dryRunDisplayPath(r.Result, r.Input.Path)
}

func (o removeOptions) allowNonEmptyFolder() bool {
	return o.force || o.recursive
}

func removeOperation(opts removeOptions) string {
	if opts.permanent {
		return "permanent_delete"
	}
	return "delete"
}

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm [flags] <file>",
	Short: "Remove files or folders",
	RunE:  rm,
}

func init() {
	RootCmd.AddCommand(rmCmd)
	enableStructuredOutput(rmCmd)
	setCommandDestructiveLevel(rmCmd, destructiveLevelDelete)
	rmCmd.Flags().BoolP("force", "f", false, "Allow removing non-empty folders; same as --recursive")
	rmCmd.Flags().BoolP("recursive", "r", false, "Recursively remove folders")
	rmCmd.Flags().Bool("permanent", false, "Permanently delete instead of moving to Dropbox trash")
	addDryRunFlag(rmCmd)
}
