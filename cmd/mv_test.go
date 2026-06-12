package cmd

import (
	"fmt"
	"strings"
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

func TestMvMultipleSourcesTreatsDestinationAsFolder(t *testing.T) {
	var moved []*files.RelocationArg
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			t.Fatalf("GetMetadata called for multiple sources: %v", arg)
			return nil, nil
		},
		moveV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			moved = append(moved, arg)
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	err := mv(mvCmd, []string{"/src/a.txt", "/src/b.txt", "/dest"})
	if err != nil {
		t.Fatalf("mv error: %v", err)
	}
	if len(moved) != 2 {
		t.Fatalf("moved %d items, want 2", len(moved))
	}
	if moved[0].ToPath != "/dest/a.txt" {
		t.Errorf("first ToPath = %q, want /dest/a.txt", moved[0].ToPath)
	}
	if moved[1].ToPath != "/dest/b.txt" {
		t.Errorf("second ToPath = %q, want /dest/b.txt", moved[1].ToPath)
	}
}

func TestMvMoveErrorIncludesAPIError(t *testing.T) {
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, fmt.Errorf("path/not_found/")
		},
		moveV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			return nil, fmt.Errorf("path/malformed_path/")
		},
	}
	stubFilesClient(t, mock)

	var err error
	stderr := captureStderr(t, func() {
		err = mv(mvCmd, []string{"/src/file.txt", "/dest/file.txt"})
	})
	if err == nil {
		t.Fatal("expected mv error")
	}
	if !strings.Contains(stderr, `move "/src/file.txt" to "/dest/file.txt": path/malformed_path/`) {
		t.Errorf("stderr = %q, want move path and API error", stderr)
	}
	if strings.Contains(stderr, "&{{") {
		t.Errorf("stderr still contains formatted relocation arg: %q", stderr)
	}
}
