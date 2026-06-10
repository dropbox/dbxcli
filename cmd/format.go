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
	"sort"
	"strings"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dustin/go-humanize"
)

type listOptions struct {
	long       bool
	timeField  string
	timeFormat string
	sortBy     string
	reverse    bool
}

func formatTime(t time.Time, opts listOptions) string {
	switch opts.timeFormat {
	case "short":
		return t.Format("2006-01-02 15:04")
	case "rfc3339":
		return t.Format(time.RFC3339)
	default:
		return humanize.Time(t)
	}
}

func getTime(e *files.FileMetadata, opts listOptions) time.Time {
	if opts.timeField == "client" {
		return e.ClientModified
	}
	return e.ServerModified
}

func sortEntries(entries []files.IsMetadata, opts listOptions) {
	if opts.sortBy == "" {
		return
	}

	sort.SliceStable(entries, func(i, j int) bool {
		less := compareLess(entries[i], entries[j], opts)
		if opts.reverse {
			return !less
		}
		return less
	})
}

func compareLess(a, b files.IsMetadata, opts listOptions) bool {
	switch opts.sortBy {
	case "name":
		return strings.ToLower(entryPath(a)) < strings.ToLower(entryPath(b))
	case "size":
		return entrySize(a) < entrySize(b)
	case "time":
		return entryTime(a, opts).Before(entryTime(b, opts))
	case "type":
		return entryTypeOrder(a) < entryTypeOrder(b)
	default:
		return false
	}
}

func entryPath(e files.IsMetadata) string {
	switch f := e.(type) {
	case *files.FileMetadata:
		return f.PathDisplay
	case *files.FolderMetadata:
		return f.PathDisplay
	case *files.DeletedMetadata:
		return f.PathDisplay
	}
	return ""
}

func entrySize(e files.IsMetadata) uint64 {
	if f, ok := e.(*files.FileMetadata); ok {
		return f.Size
	}
	return 0
}

func entryTime(e files.IsMetadata, opts listOptions) time.Time {
	if f, ok := e.(*files.FileMetadata); ok {
		return getTime(f, opts)
	}
	return time.Time{}
}

func entryTypeOrder(e files.IsMetadata) int {
	switch e.(type) {
	case *files.FolderMetadata:
		return 0
	case *files.FileMetadata:
		return 1
	case *files.DeletedMetadata:
		return 2
	}
	return 3
}
