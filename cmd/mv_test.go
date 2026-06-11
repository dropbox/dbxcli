package cmd

import (
	"fmt"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func TestMvArgValidation(t *testing.T) {
	err := mv(mvCmd, []string{})
	if err == nil {
		t.Error("expected error for no args")
	}

	err = mv(mvCmd, []string{"/only-one"})
	if err == nil {
		t.Error("expected error for single arg")
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
