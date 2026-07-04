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
	"path"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

// Dropbox API paths are slash-separated regardless of the local OS. Use package
// path for Dropbox paths and filepath only for local filesystem paths.
func cleanDropboxPath(p string) string {
	p = path.Clean(p)
	if p == "/" {
		return ""
	}
	return p
}

func relocationDestination(source, destination string, destinationIsFolder bool) string {
	if destinationIsFolder {
		return path.Join(destination, path.Base(source))
	}
	return destination
}

func isRemoteFolder(dbx filesClient, dst string) bool {
	p, err := validatePath(dst)
	if err != nil {
		return false
	}
	meta, err := dbx.GetMetadataContext(currentContext(), files.NewGetMetadataArg(p))
	if err != nil {
		return false
	}
	_, ok := meta.(*files.FolderMetadata)
	return ok
}
