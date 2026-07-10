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
	"io"
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

type mkdirInput struct {
	Path    string `json:"path"`
	Parents bool   `json:"parents"`
	DryRun  bool   `json:"dry_run,omitempty"`
}

type mkdirOptions struct {
	parents bool
	dryRun  bool
}

const (
	mkdirStatusCreated  = "created"
	mkdirStatusExisting = "existing"

	mkdirKindFolder = "folder"
)

type mkdirResult struct {
	Status string       `json:"status"`
	Kind   string       `json:"kind"`
	Input  mkdirInput   `json:"input"`
	Result jsonMetadata `json:"result"`
}

func mkdir(cmd *cobra.Command, args []string) (err error) {
	if len(args) != 1 {
		return invalidArgumentsErrorWithDetails("`mkdir` requires a `directory` argument", argumentErrorDetails("directory"))
	}

	dst, err := validatePath(args[0])
	if err != nil {
		return
	}

	opts, err := parseMkdirOptions(cmd)
	if err != nil {
		return err
	}

	if opts.dryRun {
		result := newPlannedMkdirResult(dst, opts)
		return renderOperation(cmd, result.Input, []jsonOperationResult{mkdirOperationResult(result)}, nil, func(w io.Writer) error {
			return writeDryRunLine(w, "create directory", result.displayPath())
		})
	}

	arg := files.NewCreateFolderArg(dst)
	dbx := filesNewFunc(config)
	created, err := dbx.CreateFolderV2Context(currentContext(), arg)
	var metadata *files.FolderMetadata
	status := mkdirStatusCreated
	if err != nil {
		if !opts.parents {
			return err
		}

		conflictTag, ok := createFolderConflictTag(err)
		switch {
		case ok && conflictTag == files.WriteConflictErrorFolder:
			if commandOutputFormat(cmd) == "text" {
				return nil
			}
			metadata, err = existingFolderMetadata(dbx, dst)
			if err != nil {
				return err
			}
			status = mkdirStatusExisting
		case ok && (conflictTag == files.WriteConflictErrorFile || conflictTag == files.WriteConflictErrorFileAncestor):
			return pathConflictErrorWithPath(dst, "path exists and is not a folder: %s", dst)
		case ok:
			return err
		case isConflictError(err):
			if commandOutputFormat(cmd) == "text" {
				return nil
			}
			metadata, err = existingFolderMetadata(dbx, dst)
			if err != nil {
				return err
			}
			status = mkdirStatusExisting
		default:
			return err
		}
	} else {
		if created == nil || created.Metadata == nil {
			return errors.New("create folder returned no metadata")
		}
		metadata = created.Metadata
	}

	result, err := newMkdirResult(status, dst, opts, metadata)
	if err != nil {
		return err
	}
	return renderJSONOperationOutput(cmd, result.Input, []jsonOperationResult{mkdirOperationResult(result)})
}

func parseMkdirOptions(cmd *cobra.Command) (mkdirOptions, error) {
	parents, err := cmd.Flags().GetBool("parents")
	if err != nil {
		return mkdirOptions{}, err
	}

	dryRun, err := dryRunEnabled(cmd)
	if err != nil {
		return mkdirOptions{}, err
	}

	return mkdirOptions{
		parents: parents,
		dryRun:  dryRun,
	}, nil
}

func existingFolderMetadata(dbx filesClient, dst string) (*files.FolderMetadata, error) {
	metadata, err := dbx.GetMetadataContext(currentContext(), files.NewGetMetadataArg(dst))
	if err != nil {
		return nil, err
	}
	folder, ok := metadata.(*files.FolderMetadata)
	if !ok || folder == nil {
		return nil, pathConflictErrorWithPath(dst, "path exists and is not a folder: %s", dst)
	}
	return folder, nil
}

func newMkdirResult(status, path string, opts mkdirOptions, metadata *files.FolderMetadata) (mkdirResult, error) {
	result, err := jsonMetadataFromDropbox(metadata)
	if err != nil {
		return mkdirResult{}, err
	}
	result.PathDisplay = metadataDisplayPath(path, result.PathDisplay)

	return mkdirResult{
		Status: status,
		Kind:   mkdirKindFolder,
		Input: mkdirInput{
			Path:    path,
			Parents: opts.parents,
			DryRun:  opts.dryRun,
		},
		Result: result,
	}, nil
}

func newPlannedMkdirResult(path string, opts mkdirOptions) mkdirResult {
	return mkdirResult{
		Status: mkdirStatusCreated,
		Kind:   mkdirKindFolder,
		Input: mkdirInput{
			Path:    path,
			Parents: opts.parents,
			DryRun:  opts.dryRun,
		},
		Result: plannedMetadata(mkdirKindFolder, path),
	}
}

func mkdirOperationResult(result mkdirResult) jsonOperationResult {
	return newJSONOperationResult(plannedStatus(result.Input.DryRun, result.Status), result.Kind, result.Input, result.Result)
}

func (r mkdirResult) displayPath() string {
	return dryRunDisplayPath(r.Result, r.Input.Path)
}

func createFolderConflictTag(err error) (string, bool) {
	var apiErrPtr *files.CreateFolderV2APIError
	if errors.As(err, &apiErrPtr) && apiErrPtr != nil {
		return createFolderEndpointConflictTag(apiErrPtr.EndpointError)
	}

	var apiErr files.CreateFolderV2APIError
	if errors.As(err, &apiErr) {
		return createFolderEndpointConflictTag(apiErr.EndpointError)
	}

	return "", false
}

func createFolderEndpointConflictTag(endpointErr *files.CreateFolderError) (string, bool) {
	if endpointErr == nil ||
		endpointErr.Tag != files.CreateFolderErrorPath ||
		endpointErr.Path == nil ||
		endpointErr.Path.Tag != files.WriteErrorConflict ||
		endpointErr.Path.Conflict == nil {
		return "", false
	}
	return endpointErr.Path.Conflict.Tag, true
}

func isConflictError(err error) bool {
	return strings.Contains(err.Error(), "path/conflict")
}

// mkdirCmd represents the mkdir command
var mkdirCmd = &cobra.Command{
	Use:   "mkdir [flags] <directory>",
	Short: "Create a new directory",
	RunE:  mkdir,
}

func init() {
	RootCmd.AddCommand(mkdirCmd)
	mkdirCmd.Flags().BoolP("parents", "p", false, "No error if existing, create parent directories as needed")
	addDryRunFlag(mkdirCmd)
	enableStructuredOutput(mkdirCmd)
}
