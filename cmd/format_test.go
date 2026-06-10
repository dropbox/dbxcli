package cmd

import (
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func TestFormatTimeDefault(t *testing.T) {
	ts := time.Now().Add(-2 * time.Hour)
	got := formatTime(ts, listOptions{})
	if got == "" {
		t.Error("formatTime default returned empty")
	}
	if got == ts.Format(time.RFC3339) {
		t.Error("default should be relative, not rfc3339")
	}
}

func TestFormatTimeShort(t *testing.T) {
	ts := time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC)
	got := formatTime(ts, listOptions{timeFormat: "short"})
	want := "2025-03-15 10:30"
	if got != want {
		t.Errorf("formatTime short = %q, want %q", got, want)
	}
}

func TestFormatTimeRFC3339(t *testing.T) {
	ts := time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC)
	got := formatTime(ts, listOptions{timeFormat: "rfc3339"})
	want := "2025-03-15T10:30:00Z"
	if got != want {
		t.Errorf("formatTime rfc3339 = %q, want %q", got, want)
	}
}

func TestGetTimeServer(t *testing.T) {
	serverTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	clientTime := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	meta := &files.FileMetadata{
		ServerModified: serverTime,
		ClientModified: clientTime,
	}

	got := getTime(meta, listOptions{timeField: "server"})
	if !got.Equal(serverTime) {
		t.Errorf("getTime server = %v, want %v", got, serverTime)
	}

	got = getTime(meta, listOptions{})
	if !got.Equal(serverTime) {
		t.Errorf("getTime default = %v, want %v", got, serverTime)
	}
}

func TestGetTimeClient(t *testing.T) {
	serverTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	clientTime := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	meta := &files.FileMetadata{
		ServerModified: serverTime,
		ClientModified: clientTime,
	}

	got := getTime(meta, listOptions{timeField: "client"})
	if !got.Equal(clientTime) {
		t.Errorf("getTime client = %v, want %v", got, clientTime)
	}
}

func TestSortEntriesByName(t *testing.T) {
	entries := []files.IsMetadata{
		&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/Zebra"}},
		&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/apple.txt"}},
		&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/Banana"}},
	}

	sortEntries(entries, listOptions{sortBy: "name"})

	paths := []string{entryPath(entries[0]), entryPath(entries[1]), entryPath(entries[2])}
	if paths[0] != "/apple.txt" || paths[1] != "/Banana" || paths[2] != "/Zebra" {
		t.Errorf("sort by name = %v, want [/apple.txt /Banana /Zebra]", paths)
	}
}

func TestSortEntriesBySize(t *testing.T) {
	entries := []files.IsMetadata{
		&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/big"}, Size: 1000},
		&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/small"}, Size: 10},
		&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/folder"}},
	}

	sortEntries(entries, listOptions{sortBy: "size"})

	paths := []string{entryPath(entries[0]), entryPath(entries[1]), entryPath(entries[2])}
	if paths[0] != "/folder" || paths[1] != "/small" || paths[2] != "/big" {
		t.Errorf("sort by size = %v, want [/folder /small /big]", paths)
	}
}

func TestSortEntriesByTime(t *testing.T) {
	entries := []files.IsMetadata{
		&files.FileMetadata{
			Metadata:       files.Metadata{PathDisplay: "/new"},
			ServerModified: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		&files.FileMetadata{
			Metadata:       files.Metadata{PathDisplay: "/old"},
			ServerModified: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/folder"}},
	}

	sortEntries(entries, listOptions{sortBy: "time"})

	paths := []string{entryPath(entries[0]), entryPath(entries[1]), entryPath(entries[2])}
	if paths[0] != "/folder" || paths[1] != "/old" || paths[2] != "/new" {
		t.Errorf("sort by time = %v, want [/folder /old /new]", paths)
	}
}

func TestSortEntriesByType(t *testing.T) {
	entries := []files.IsMetadata{
		&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/file.txt"}},
		&files.DeletedMetadata{Metadata: files.Metadata{PathDisplay: "/gone"}},
		&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/folder"}},
	}

	sortEntries(entries, listOptions{sortBy: "type"})

	paths := []string{entryPath(entries[0]), entryPath(entries[1]), entryPath(entries[2])}
	if paths[0] != "/folder" || paths[1] != "/file.txt" || paths[2] != "/gone" {
		t.Errorf("sort by type = %v, want [/folder /file.txt /gone]", paths)
	}
}

func TestSortEntriesReverse(t *testing.T) {
	entries := []files.IsMetadata{
		&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/a"}},
		&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/b"}},
		&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/c"}},
	}

	sortEntries(entries, listOptions{sortBy: "name", reverse: true})

	paths := []string{entryPath(entries[0]), entryPath(entries[1]), entryPath(entries[2])}
	if paths[0] != "/c" || paths[1] != "/b" || paths[2] != "/a" {
		t.Errorf("sort by name reverse = %v, want [/c /b /a]", paths)
	}
}

func TestSortEntriesNoSort(t *testing.T) {
	entries := []files.IsMetadata{
		&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/z"}},
		&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/a"}},
	}

	sortEntries(entries, listOptions{})

	if entryPath(entries[0]) != "/z" || entryPath(entries[1]) != "/a" {
		t.Error("with no sort, order should be preserved")
	}
}

func TestFormatFileMetadataWithOptsShort(t *testing.T) {
	ts := time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC)
	meta := &files.FileMetadata{
		Metadata:       files.Metadata{PathDisplay: "/test.txt"},
		Rev:            "abc",
		Size:           4096,
		ServerModified: ts,
	}

	got := formatFileMetadataWithOpts(meta, listOptions{long: false})
	want := "/test.txt\t"
	if got != want {
		t.Errorf("short format = %q, want %q", got, want)
	}
}

func TestFormatFileMetadataWithOptsLongShortTime(t *testing.T) {
	ts := time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC)
	meta := &files.FileMetadata{
		Metadata:       files.Metadata{PathDisplay: "/test.txt"},
		Rev:            "abc",
		Size:           4096,
		ServerModified: ts,
	}

	got := formatFileMetadataWithOpts(meta, listOptions{long: true, timeFormat: "short"})
	if !stringContains(got, "2025-03-15 10:30") {
		t.Errorf("long short-time format should contain absolute time, got %q", got)
	}
	if !stringContains(got, "4.0 KiB") {
		t.Errorf("long format should contain size, got %q", got)
	}
}

func TestFormatFileMetadataWithOptsClientTime(t *testing.T) {
	server := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	client := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	meta := &files.FileMetadata{
		Metadata:       files.Metadata{PathDisplay: "/test.txt"},
		Rev:            "abc",
		Size:           1024,
		ServerModified: server,
		ClientModified: client,
	}

	got := formatFileMetadataWithOpts(meta, listOptions{long: true, timeField: "client", timeFormat: "short"})
	if !stringContains(got, "2024-06-15 12:00") {
		t.Errorf("client time format should show client modified, got %q", got)
	}
}
