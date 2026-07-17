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
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func metadataDisplayPath(inputPath, metadataPath string) string {
	if metadataPath != "" {
		return metadataPath
	}
	return inputPath
}

func sameDropboxPath(a string, b string) bool {
	return strings.EqualFold(cleanDropboxPath(a), cleanDropboxPath(b))
}

func sameDropboxMetadataPath(pathDisplay, pathLower, requested string) bool {
	if pathLower != "" {
		return sameDropboxPath(pathLower, requested)
	}
	if pathDisplay != "" {
		return sameDropboxPath(pathDisplay, requested)
	}
	return true
}

func baseMetadata(metadata files.IsMetadata) *files.Metadata {
	switch entry := metadata.(type) {
	case *files.Metadata:
		return entry
	case *files.FileMetadata:
		if entry != nil {
			return &entry.Metadata
		}
	case *files.FolderMetadata:
		if entry != nil {
			return &entry.Metadata
		}
	case *files.DeletedMetadata:
		if entry != nil {
			return &entry.Metadata
		}
	}
	return nil
}

func metadataName(metadata files.IsMetadata) string {
	if base := baseMetadata(metadata); base != nil {
		return base.Name
	}
	return ""
}

func metadataPathDisplay(metadata files.IsMetadata) string {
	if base := baseMetadata(metadata); base != nil {
		return base.PathDisplay
	}
	return ""
}
