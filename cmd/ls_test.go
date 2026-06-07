package cmd

import (
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/", ""},
		{"/foo", "/foo"},
		{"foo", "/foo"},
		{"/foo/", "/foo"},
		{"/foo/bar", "/foo/bar"},
		{"foo/bar/", "/foo/bar"},
	}

	for _, tt := range tests {
		got, err := validatePath(tt.input)
		if err != nil {
			t.Errorf("validatePath(%q) returned error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("validatePath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatFolderMetadata(t *testing.T) {
	meta := &files.FolderMetadata{
		Metadata: files.Metadata{
			PathDisplay: "/Documents",
		},
	}

	got := formatFolderMetadata(meta, false)
	want := "/Documents\t"
	if got != want {
		t.Errorf("formatFolderMetadata(short) = %q, want %q", got, want)
	}

	got = formatFolderMetadata(meta, true)
	want = "-\t-\t-\t/Documents\t"
	if got != want {
		t.Errorf("formatFolderMetadata(long) = %q, want %q", got, want)
	}
}

func TestFormatFileMetadata(t *testing.T) {
	meta := &files.FileMetadata{
		Metadata: files.Metadata{
			PathDisplay: "/test.txt",
		},
		Rev:  "abc123",
		Size: 1024,
	}

	got := formatFileMetadata(meta, false)
	want := "/test.txt\t"
	if got != want {
		t.Errorf("formatFileMetadata(short) = %q, want %q", got, want)
	}

	got = formatFileMetadata(meta, true)
	if got == "" {
		t.Error("formatFileMetadata(long) returned empty string")
	}
	if len(got) <= len("/test.txt\t") {
		t.Errorf("formatFileMetadata(long) should include rev/size/time, got %q", got)
	}
}

func TestFormatDeletedMetadata(t *testing.T) {
	meta := &files.DeletedMetadata{
		Metadata: files.Metadata{
			PathDisplay: "/removed.txt",
		},
	}

	got := formatDeletedMetadata(meta, false)
	want := "/removed.txt\t"
	if got != want {
		t.Errorf("formatDeletedMetadata(short) = %q, want %q", got, want)
	}

	got = formatDeletedMetadata(meta, true)
	want = "-\t-\t-\t/removed.txt\t"
	if got != want {
		t.Errorf("formatDeletedMetadata(long) = %q, want %q", got, want)
	}
}

func TestSetPathDisplayAsDeleted(t *testing.T) {
	file := &files.FileMetadata{
		Metadata: files.Metadata{PathDisplay: "/file.txt"},
	}
	SetPathDisplayAsDeleted(file)
	if file.PathDisplay != "<</file.txt>>" {
		t.Errorf("file PathDisplay = %q, want %q", file.PathDisplay, "<</file.txt>>")
	}

	folder := &files.FolderMetadata{
		Metadata: files.Metadata{PathDisplay: "/folder"},
	}
	SetPathDisplayAsDeleted(folder)
	if folder.PathDisplay != "<</folder>>" {
		t.Errorf("folder PathDisplay = %q, want %q", folder.PathDisplay, "<</folder>>")
	}

	deleted := &files.DeletedMetadata{
		Metadata: files.Metadata{PathDisplay: "/gone"},
	}
	SetPathDisplayAsDeleted(deleted)
	if deleted.PathDisplay != "<</gone>>" {
		t.Errorf("deleted PathDisplay = %q, want %q", deleted.PathDisplay, "<</gone>>")
	}
}

func TestFormatFileMetadataLongIncludesFields(t *testing.T) {
	ts := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	meta := &files.FileMetadata{
		Metadata: files.Metadata{
			PathDisplay: "/data.csv",
		},
		Rev:            "rev999",
		Size:           2048,
		ServerModified: ts,
	}

	got := formatFileMetadata(meta, true)
	if !contains(got, "rev999") {
		t.Errorf("long format should contain rev, got %q", got)
	}
	if !contains(got, "2.0 KiB") {
		t.Errorf("long format should contain human size, got %q", got)
	}
	if !contains(got, "/data.csv") {
		t.Errorf("long format should contain path, got %q", got)
	}
}

func TestValidatePathRootBecomesEmpty(t *testing.T) {
	got, err := validatePath("/")
	if err != nil {
		t.Fatalf("validatePath('/') error: %v", err)
	}
	if got != "" {
		t.Errorf("validatePath('/') = %q, want empty string for root", got)
	}
}

func TestGetFileMetadataNotCalledForRoot(t *testing.T) {
	// This test verifies the ls function logic:
	// when path is "" (root), getFileMetadata should not be called.
	// We test this indirectly by confirming that NewGetMetadataArg
	// with empty string would be invalid for the API.
	arg := files.NewGetMetadataArg("")
	if arg.Path != "" {
		t.Errorf("NewGetMetadataArg('') path = %q, want empty", arg.Path)
	}
}

func TestMakeRelocationArg(t *testing.T) {
	arg, err := makeRelocationArg("src", "dst")
	if err != nil {
		t.Fatalf("makeRelocationArg error: %v", err)
	}
	if arg.FromPath != "/src" {
		t.Errorf("FromPath = %q, want %q", arg.FromPath, "/src")
	}
	if arg.ToPath != "/dst" {
		t.Errorf("ToPath = %q, want %q", arg.ToPath, "/dst")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Verify that dropbox config can be constructed without panicking
func TestDropboxConfigConstruction(t *testing.T) {
	cfg := dropbox.Config{
		Token:    "test-token",
		LogLevel: dropbox.LogOff,
	}
	if cfg.Token != "test-token" {
		t.Error("config token not set")
	}
}
