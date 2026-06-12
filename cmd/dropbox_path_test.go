package cmd

import (
	"fmt"
	"testing"

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
		{"/foo//bar", "/foo/bar"},
		{"foo//bar/", "/foo/bar"},
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

func TestValidatePathRootBecomesEmpty(t *testing.T) {
	got, err := validatePath("/")
	if err != nil {
		t.Fatalf("validatePath('/') error: %v", err)
	}
	if got != "" {
		t.Errorf("validatePath('/') = %q, want empty string for root", got)
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

func TestRelocationDestination(t *testing.T) {
	tests := []struct {
		name                string
		source              string
		destination         string
		destinationIsFolder bool
		want                string
	}{
		{
			name:                "exact destination",
			source:              "/src/file.txt",
			destination:         "/dest/file-copy.txt",
			destinationIsFolder: false,
			want:                "/dest/file-copy.txt",
		},
		{
			name:                "folder destination",
			source:              "/src/file.txt",
			destination:         "/dest",
			destinationIsFolder: true,
			want:                "/dest/file.txt",
		},
		{
			name:                "folder destination trailing slash",
			source:              "/src/file.txt",
			destination:         "/dest/",
			destinationIsFolder: true,
			want:                "/dest/file.txt",
		},
	}

	for _, tt := range tests {
		got := relocationDestination(tt.source, tt.destination, tt.destinationIsFolder)
		if got != tt.want {
			t.Errorf("%s: relocationDestination() = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestIsRemoteFolder_Folder(t *testing.T) {
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return &files.FolderMetadata{Metadata: files.Metadata{PathDisplay: arg.Path}}, nil
		},
	}
	if !isRemoteFolder(mock, "/Videos") {
		t.Error("expected true for folder metadata")
	}
}

func TestIsRemoteFolder_File(t *testing.T) {
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return &files.FileMetadata{}, nil
		},
	}
	if isRemoteFolder(mock, "/file.txt") {
		t.Error("expected false for file metadata")
	}
}

func TestIsRemoteFolder_NotFound(t *testing.T) {
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, fmt.Errorf("path/not_found/")
		},
	}
	if isRemoteFolder(mock, "/nonexistent") {
		t.Error("expected false for not found")
	}
}
