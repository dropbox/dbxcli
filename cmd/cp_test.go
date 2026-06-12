package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func stubFilesClient(t *testing.T, client files.Client) {
	t.Helper()

	origNew := filesNewFunc
	filesNewFunc = func(_ dropbox.Config) files.Client { return client }
	t.Cleanup(func() { filesNewFunc = origNew })
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stderr = w
	fn()
	closeErr := w.Close()
	os.Stderr = oldStderr
	if closeErr != nil {
		t.Fatal(closeErr)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestCpArgValidation(t *testing.T) {
	err := cp(cpCmd, []string{})
	if err == nil {
		t.Error("expected error for no args")
	}

	err = cp(cpCmd, []string{"/only-one"})
	if err == nil {
		t.Error("expected error for single arg")
	}
}

func TestCpCopiesIntoExistingRemoteFolder(t *testing.T) {
	var copied []*files.RelocationArg
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			if arg.Path != "/dest" {
				t.Errorf("metadata path = %q, want /dest", arg.Path)
			}
			return &files.FolderMetadata{}, nil
		},
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			copied = append(copied, arg)
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	err := cp(cpCmd, []string{"/src/file.txt", "/dest"})
	if err != nil {
		t.Fatalf("cp error: %v", err)
	}
	if len(copied) != 1 {
		t.Fatalf("copied %d items, want 1", len(copied))
	}
	if copied[0].FromPath != "/src/file.txt" {
		t.Errorf("FromPath = %q, want /src/file.txt", copied[0].FromPath)
	}
	if copied[0].ToPath != "/dest/file.txt" {
		t.Errorf("ToPath = %q, want /dest/file.txt", copied[0].ToPath)
	}
}

func TestCpCopiesIntoTrailingSlashDestination(t *testing.T) {
	var copied []*files.RelocationArg
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			t.Fatalf("GetMetadata called for trailing slash destination: %v", arg)
			return nil, nil
		},
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			copied = append(copied, arg)
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	err := cp(cpCmd, []string{"/src/file.txt", "/dest/"})
	if err != nil {
		t.Fatalf("cp error: %v", err)
	}
	if len(copied) != 1 {
		t.Fatalf("copied %d items, want 1", len(copied))
	}
	if copied[0].ToPath != "/dest/file.txt" {
		t.Errorf("ToPath = %q, want /dest/file.txt", copied[0].ToPath)
	}
}

func TestCpSingleSourceUsesExactDestinationWhenNotFolder(t *testing.T) {
	var copied []*files.RelocationArg
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, fmt.Errorf("path/not_found/")
		},
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			copied = append(copied, arg)
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	err := cp(cpCmd, []string{"/src/file.txt", "/dest/file-copy.txt"})
	if err != nil {
		t.Fatalf("cp error: %v", err)
	}
	if len(copied) != 1 {
		t.Fatalf("copied %d items, want 1", len(copied))
	}
	if copied[0].ToPath != "/dest/file-copy.txt" {
		t.Errorf("ToPath = %q, want /dest/file-copy.txt", copied[0].ToPath)
	}
}

func TestCpMultipleSourcesTreatsDestinationAsFolder(t *testing.T) {
	var copied []*files.RelocationArg
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			t.Fatalf("GetMetadata called for multiple sources: %v", arg)
			return nil, nil
		},
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			copied = append(copied, arg)
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	err := cp(cpCmd, []string{"/src/a.txt", "/src/b.txt", "/dest"})
	if err != nil {
		t.Fatalf("cp error: %v", err)
	}
	if len(copied) != 2 {
		t.Fatalf("copied %d items, want 2", len(copied))
	}
	if copied[0].ToPath != "/dest/a.txt" {
		t.Errorf("first ToPath = %q, want /dest/a.txt", copied[0].ToPath)
	}
	if copied[1].ToPath != "/dest/b.txt" {
		t.Errorf("second ToPath = %q, want /dest/b.txt", copied[1].ToPath)
	}
}

func TestCpCopyErrorIncludesAPIError(t *testing.T) {
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, fmt.Errorf("path/not_found/")
		},
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			return nil, fmt.Errorf("path/malformed_path/")
		},
	}
	stubFilesClient(t, mock)

	var err error
	stderr := captureStderr(t, func() {
		err = cp(cpCmd, []string{"/src/file.txt", "/dest/file.txt"})
	})
	if err == nil {
		t.Fatal("expected cp error")
	}
	if !strings.Contains(stderr, `copy "/src/file.txt" to "/dest/file.txt": path/malformed_path/`) {
		t.Errorf("stderr = %q, want copy path and API error", stderr)
	}
	if strings.Contains(stderr, "&{{") {
		t.Errorf("stderr still contains formatted relocation arg: %q", stderr)
	}
}
