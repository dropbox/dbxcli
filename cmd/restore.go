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
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

type restoreInput struct {
	Path     string `json:"path"`
	Revision string `json:"revision"`
}

const (
	restoreStatusRestored = "restored"
	restoreKindFile       = "file"
)

type restoreResult struct {
	Status string       `json:"status"`
	Kind   string       `json:"kind"`
	Input  restoreInput `json:"input"`
	Result jsonMetadata `json:"result"`
}

func restore(cmd *cobra.Command, args []string) (err error) {
	if len(args) != 2 {
		return invalidArgumentsError("`restore` requires `target-path` and `revision` arguments")
	}

	path, err := validatePath(args[0])
	if err != nil {
		return
	}

	rev := args[1]

	arg := files.NewRestoreArg(path, rev)

	dbx := filesNewFunc(config)
	metadata, err := dbx.Restore(arg)
	if err != nil {
		return
	}

	verbose, _ := cmd.Flags().GetBool("verbose")
	result := newRestoreResult(path, rev, metadata)

	return commandOutput(cmd).Render(func(w io.Writer) error {
		if !verbose {
			return nil
		}
		return renderRestoreResult(w, result)
	}, newJSONOperationOutput(result.Input, []jsonOperationResult{restoreOperationResult(result)}, nil))
}

func newRestoreResult(path, revision string, metadata *files.FileMetadata) restoreResult {
	return restoreResult{
		Status: restoreStatusRestored,
		Kind:   restoreKindFile,
		Input: restoreInput{
			Path:     path,
			Revision: revision,
		},
		Result: restoreMetadataFromDropbox(path, metadata),
	}
}

func restoreOperationResult(result restoreResult) jsonOperationResult {
	return newJSONOperationResult(result.Status, result.Kind, result.Input, result.Result)
}

func restoreMetadataFromDropbox(path string, metadata *files.FileMetadata) jsonMetadata {
	if metadata == nil {
		return jsonMetadata{
			Type:        "file",
			PathDisplay: path,
		}
	}

	result := jsonMetadataFromDropbox(metadata)
	result.PathDisplay = metadataDisplayPath(path, result.PathDisplay)
	return result
}

func renderRestoreResult(w io.Writer, result restoreResult) error {
	path := result.Result.PathDisplay
	if path == "" {
		path = result.Input.Path
	}

	if result.Result.Rev != "" && result.Result.Rev != result.Input.Revision {
		_, err := fmt.Fprintf(w, "Restored %s to revision %s (current revision %s, server modified %s)\n",
			path, result.Input.Revision, result.Result.Rev, restoreResultServerModified(result))
		return err
	}

	_, err := fmt.Fprintf(w, "Restored %s to revision %s (server modified %s)\n",
		path, result.Input.Revision, restoreResultServerModified(result))
	return err
}

func restoreResultServerModified(result restoreResult) string {
	if result.Result.ServerModified != nil {
		return *result.Result.ServerModified
	}
	return time.Time{}.Format(time.RFC3339)
}

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore [flags] <target-path> <revision>",
	Short: "Restore a file revision",
	Long: `Restore a Dropbox file at <target-path> to the supplied revision.

The target path is the Dropbox path where the restored file is saved.
Use "dbxcli revs <target-path>" to list available revisions.`,
	Example: `  dbxcli revs /Reports/old.pdf
  dbxcli restore /Reports/old.pdf 015f...`,
	RunE: restore,
}

func init() {
	RootCmd.AddCommand(restoreCmd)
	enableStructuredOutput(restoreCmd)
}
