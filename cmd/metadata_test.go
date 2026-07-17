// Copyright © 2026 Dropbox, Inc.
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
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func TestMetadataCommonFields(t *testing.T) {
	tests := []struct {
		name     string
		metadata files.IsMetadata
	}{
		{name: "metadata", metadata: &files.Metadata{Name: "metadata", PathDisplay: "/metadata"}},
		{name: "file", metadata: &files.FileMetadata{Metadata: files.Metadata{Name: "file", PathDisplay: "/file"}}},
		{name: "folder", metadata: &files.FolderMetadata{Metadata: files.Metadata{Name: "folder", PathDisplay: "/folder"}}},
		{name: "deleted", metadata: &files.DeletedMetadata{Metadata: files.Metadata{Name: "deleted", PathDisplay: "/deleted"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := metadataName(tt.metadata); got != tt.name {
				t.Fatalf("metadataName() = %q, want %q", got, tt.name)
			}
			if got := metadataPathDisplay(tt.metadata); got != "/"+tt.name {
				t.Fatalf("metadataPathDisplay() = %q, want %q", got, "/"+tt.name)
			}
		})
	}
}

func TestMetadataCommonFieldsHandleNil(t *testing.T) {
	var metadata *files.FileMetadata
	if got := metadataName(metadata); got != "" {
		t.Fatalf("metadataName() = %q, want empty", got)
	}
	if got := metadataPathDisplay(metadata); got != "" {
		t.Fatalf("metadataPathDisplay() = %q, want empty", got)
	}
}
