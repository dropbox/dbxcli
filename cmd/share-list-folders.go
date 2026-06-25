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

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

type sharedFolderClient interface {
	ListFolders(*sharing.ListFoldersArgs) (*sharing.ListFoldersResult, error)
	ListFoldersContinue(*sharing.ListFoldersContinueArg) (*sharing.ListFoldersResult, error)
}

type shareFolderListInput struct{}

type shareFolderJSONMetadata struct {
	Type                 string   `json:"type"`
	Name                 string   `json:"name"`
	PathLower            string   `json:"path_lower,omitempty"`
	SharedFolderID       string   `json:"shared_folder_id"`
	PreviewURL           string   `json:"preview_url,omitempty"`
	AccessType           string   `json:"access_type,omitempty"`
	IsInsideTeamFolder   bool     `json:"is_inside_team_folder"`
	IsTeamFolder         bool     `json:"is_team_folder"`
	OwnerDisplayNames    []string `json:"owner_display_names,omitempty"`
	ParentSharedFolderID string   `json:"parent_shared_folder_id,omitempty"`
	ParentFolderName     string   `json:"parent_folder_name,omitempty"`
	TimeInvited          *string  `json:"time_invited,omitempty"`
	AccessInheritance    string   `json:"access_inheritance,omitempty"`
}

const (
	shareFolderJSONStatusListed = "listed"
	shareFolderJSONKindFolder   = "shared_folder"
)

var newSharedFolderClient = func(cfg dropbox.Config) sharedFolderClient {
	return sharing.New(cfg)
}

func shareListFolders(cmd *cobra.Command, args []string) (err error) {
	arg := sharing.NewListFoldersArgs()

	dbx := newSharedFolderClient(config)
	entries, err := listSharedFolders(dbx, arg)
	if err != nil {
		return
	}

	commandVerboseStatus(cmd, "Listed %d shared folders", len(entries))

	return commandOutput(cmd).Render(func(w io.Writer) error {
		return renderSharedFolders(w, entries)
	}, newJSONOperationOutput(
		shareFolderListInput{},
		shareFolderJSONOperationResults(shareFolderJSONMetadataListFromDropbox(entries)),
		nil,
	))
}

func listSharedFolders(dbx sharedFolderClient, arg *sharing.ListFoldersArgs) ([]*sharing.SharedFolderMetadata, error) {
	var entries []*sharing.SharedFolderMetadata
	res, err := dbx.ListFolders(arg)
	if err != nil {
		return nil, err
	}
	entries = append(entries, res.Entries...)

	for len(res.Cursor) > 0 {
		continueArg := sharing.NewListFoldersContinueArg(res.Cursor)

		res, err = dbx.ListFoldersContinue(continueArg)
		if err != nil {
			return nil, err
		}

		entries = append(entries, res.Entries...)
	}

	return entries, nil
}

func renderSharedFolders(out io.Writer, entries []*sharing.SharedFolderMetadata) error {
	for _, f := range entries {
		if _, err := fmt.Fprintf(out, "%v\t%v\n", f.PathLower, f.PreviewUrl); err != nil {
			return err
		}
	}

	return nil
}

func shareFolderJSONMetadataListFromDropbox(entries []*sharing.SharedFolderMetadata) []shareFolderJSONMetadata {
	result := make([]shareFolderJSONMetadata, 0, len(entries))
	for _, entry := range entries {
		result = append(result, shareFolderJSONMetadataFromDropbox(entry))
	}
	return result
}

func shareFolderJSONMetadataFromDropbox(entry *sharing.SharedFolderMetadata) shareFolderJSONMetadata {
	if entry == nil {
		return shareFolderJSONMetadata{Type: shareFolderJSONKindFolder}
	}

	result := shareFolderJSONMetadata{
		Type:                 shareFolderJSONKindFolder,
		Name:                 entry.Name,
		PathLower:            entry.PathLower,
		SharedFolderID:       entry.SharedFolderId,
		PreviewURL:           entry.PreviewUrl,
		IsInsideTeamFolder:   entry.IsInsideTeamFolder,
		IsTeamFolder:         entry.IsTeamFolder,
		OwnerDisplayNames:    entry.OwnerDisplayNames,
		ParentSharedFolderID: entry.ParentSharedFolderId,
		ParentFolderName:     entry.ParentFolderName,
		TimeInvited:          jsonTime(entry.TimeInvited),
	}
	if entry.AccessType != nil {
		result.AccessType = entry.AccessType.Tag
	}
	if entry.AccessInheritance != nil {
		result.AccessInheritance = entry.AccessInheritance.Tag
	}
	return result
}

func shareFolderJSONOperationResults(entries []shareFolderJSONMetadata) []jsonOperationResult {
	results := make([]jsonOperationResult, 0, len(entries))
	for _, entry := range entries {
		results = append(results, newJSONOperationResult(shareFolderJSONStatusListed, shareFolderJSONKindFolder, nil, entry))
	}
	return results
}

var shareListFoldersCmd = &cobra.Command{
	Use:   "folder",
	Short: "List shared folders",
	RunE:  shareListFolders,
}

func init() {
	shareListCmd.AddCommand(shareListFoldersCmd)
	enableStructuredOutput(shareListFoldersCmd)
}
