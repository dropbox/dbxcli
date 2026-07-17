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

import "testing"

func TestNewDropboxReference(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		kind  dropboxReferenceKind
	}{
		{name: "absolute path", input: "/folder/file.txt", want: "/folder/file.txt", kind: dropboxPathReference},
		{name: "relative path", input: "folder/file.txt", want: "/folder/file.txt", kind: dropboxPathReference},
		{name: "file id", input: "id:opaque-value", want: "id:opaque-value", kind: dropboxIDReference},
		{name: "revision", input: "rev:opaque-value", want: "rev:opaque-value", kind: dropboxRevisionReference},
		{name: "namespace path", input: "ns:123/folder/file.txt", want: "ns:123/folder/file.txt", kind: dropboxNamespaceReference},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newDropboxReference(tt.input)
			if got.String() != tt.want {
				t.Fatalf("String() = %q, want %q", got.String(), tt.want)
			}
			if got.kind != tt.kind {
				t.Fatalf("kind = %d, want %d", got.kind, tt.kind)
			}
		})
	}
}
