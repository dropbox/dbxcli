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
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

type restoreInput struct {
	Path     string `json:"path"`
	Revision string `json:"revision"`
}

type restoreMetadata struct {
	Type           string    `json:"type"`
	PathDisplay    string    `json:"path_display,omitempty"`
	ID             string    `json:"id,omitempty"`
	Rev            string    `json:"rev,omitempty"`
	Size           uint64    `json:"size,omitempty"`
	ClientModified time.Time `json:"client_modified"`
	ServerModified time.Time `json:"server_modified"`
}

type restoreResult struct {
	Input  restoreInput    `json:"input"`
	Result restoreMetadata `json:"result"`
}

func restore(cmd *cobra.Command, args []string) (err error) {
	if len(args) != 2 {
		return errors.New("`restore` requires `target-path` and `revision` arguments")
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
	if verbose {
		printRestoreResult(cmd, newRestoreResult(path, rev, metadata))
	}

	return
}

func newRestoreResult(path, revision string, metadata *files.FileMetadata) restoreResult {
	return restoreResult{
		Input: restoreInput{
			Path:     path,
			Revision: revision,
		},
		Result: restoreMetadataFromDropbox(path, metadata),
	}
}

func restoreMetadataFromDropbox(path string, metadata *files.FileMetadata) restoreMetadata {
	if metadata == nil {
		return restoreMetadata{
			Type:        "file",
			PathDisplay: path,
		}
	}
	return restoreMetadata{
		Type:           "file",
		PathDisplay:    metadataDisplayPath(path, metadata.PathDisplay),
		ID:             metadata.Id,
		Rev:            metadata.Rev,
		Size:           metadata.Size,
		ClientModified: metadata.ClientModified,
		ServerModified: metadata.ServerModified,
	}
}

func printRestoreResult(cmd *cobra.Command, result restoreResult) {
	path := result.Result.PathDisplay
	if path == "" {
		path = result.Input.Path
	}

	if result.Result.Rev != "" && result.Result.Rev != result.Input.Revision {
		commandOutput(cmd).Info("Restored %s to revision %s (current revision %s, server modified %s)",
			path, result.Input.Revision, result.Result.Rev, result.Result.ServerModified.Format(time.RFC3339))
		return
	}

	commandOutput(cmd).Info("Restored %s to revision %s (server modified %s)",
		path, result.Input.Revision, result.Result.ServerModified.Format(time.RFC3339))
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
}
