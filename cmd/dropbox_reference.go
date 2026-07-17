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

import "strings"

type dropboxReferenceKind uint8

const (
	dropboxPathReference dropboxReferenceKind = iota
	dropboxIDReference
	dropboxRevisionReference
	dropboxNamespaceReference
)

// dropboxReference preserves Dropbox API references as opaque strings. Only
// ordinary paths are normalized; Dropbox remains responsible for validating
// identifier, revision, and namespace reference values.
type dropboxReference struct {
	value string
	kind  dropboxReferenceKind
}

func newDropboxReference(value string) dropboxReference {
	switch {
	case strings.HasPrefix(value, "id:"):
		return dropboxReference{value: value, kind: dropboxIDReference}
	case strings.HasPrefix(value, "rev:"):
		return dropboxReference{value: value, kind: dropboxRevisionReference}
	case strings.HasPrefix(value, "ns:"):
		return dropboxReference{value: value, kind: dropboxNamespaceReference}
	default:
		if !strings.HasPrefix(value, "/") {
			value = "/" + value
		}
		return dropboxReference{value: cleanDropboxPath(value), kind: dropboxPathReference}
	}
}

func (r dropboxReference) String() string {
	return r.value
}

func (r dropboxReference) isPath() bool {
	return r.kind == dropboxPathReference
}
