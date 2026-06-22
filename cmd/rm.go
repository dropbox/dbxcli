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
	"io"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"

	"github.com/spf13/cobra"
)

type removeOptions struct {
	force     bool
	recursive bool
	permanent bool
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
}

type removeResult struct {
	Input  removeInput  `json:"input"`
	Result jsonMetadata `json:"result"`
}

type removeOutput struct {
	Results []removeResult `json:"results"`
}

func rm(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("rm: missing operand")
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
		if !opts.verbose {
			return nil
		}
		return renderRemoveResults(w, results)
	}, removeOutput{Results: results})
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

	verbose, _ := cmd.Flags().GetBool("verbose")

	return removeOptions{
		force:     force,
		recursive: recursive,
		permanent: permanent,
		verbose:   verbose,
	}, nil
}

func validateRemoveTargets(dbx files.Client, args []string, opts removeOptions) ([]removeTarget, error) {
	var targets []removeTarget

	// Validate remove paths before executing removal
	for i := range args {
		path, err := validatePath(args[i])
		if err != nil {
			return nil, err
		}

		pathMetaData, err := getFileMetadata(dbx, path)
		if err != nil {
			return nil, err
		}

		if _, ok := pathMetaData.(*files.FileMetadata); !ok && !opts.allowNonEmptyFolder() {
			folderArg := files.NewListFolderArg(path)
			res, err := dbx.ListFolder(folderArg)
			if err != nil {
				return nil, err
			}
			if len(res.Entries) != 0 {
				return nil, fmt.Errorf("rm: cannot remove ‘%s’: Directory not empty, use `--force`/`-f` or `--recursive`/`-r` to proceed", path)
			}
		}
		targets = append(targets, removeTarget{path: path, metadata: pathMetaData})
	}

	return targets, nil
}

func removeTargets(dbx files.Client, targets []removeTarget, opts removeOptions) ([]removeResult, error) {
	results := make([]removeResult, 0, len(targets))

	for _, target := range targets {
		arg := files.NewDeleteArg(target.path)
		metadata := target.metadata

		if opts.permanent {
			if err := dbx.PermanentlyDelete(arg); err != nil {
				return nil, err
			}
		} else {
			res, err := dbx.DeleteV2(arg)
			if err != nil {
				return nil, err
			}
			if res != nil && res.Metadata != nil {
				metadata = res.Metadata
			}
		}

		results = append(results, newRemoveResult(target.path, metadata, opts))
	}

	return results, nil
}

func newRemoveResult(path string, metadata files.IsMetadata, opts removeOptions) removeResult {
	return removeResult{
		Input: removeInput{
			Path:      path,
			Permanent: opts.permanent,
			Recursive: opts.recursive,
			Force:     opts.force,
		},
		Result: removeMetadataFromDropbox(path, metadata),
	}
}

func removeMetadataFromDropbox(path string, metadata files.IsMetadata) jsonMetadata {
	result := jsonMetadataFromDropbox(metadata)
	result.PathDisplay = metadataDisplayPath(path, result.PathDisplay)
	return result
}

func renderRemoveResults(w io.Writer, results []removeResult) error {
	for _, result := range results {
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
	if r.Result.PathDisplay != "" {
		return r.Result.PathDisplay
	}
	return r.Input.Path
}

func (o removeOptions) allowNonEmptyFolder() bool {
	return o.force || o.recursive
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
	rmCmd.Flags().BoolP("force", "f", false, "Allow removing non-empty folders; same as --recursive")
	rmCmd.Flags().BoolP("recursive", "r", false, "Recursively remove folders")
	rmCmd.Flags().Bool("permanent", false, "Permanently delete instead of moving to Dropbox trash")
}
