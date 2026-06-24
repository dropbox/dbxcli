package cmd

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"testing"
	"time"

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

func TestMvJSONOutputsRelocationResults(t *testing.T) {
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, fmt.Errorf("path/not_found/")
		},
		moveV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			metadata := files.NewFileMetadata("file-moved.txt", "id:file-moved", time.Time{}, time.Time{}, "rev-moved", 64)
			metadata.PathDisplay = arg.ToPath
			metadata.PathLower = strings.ToLower(arg.ToPath)
			return files.NewRelocationResult(metadata), nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)

	if err := mv(cmd, []string{"/src/file.txt", "/dest/file-moved.txt"}); err != nil {
		t.Fatalf("mv error: %v", err)
	}

	got := decodeRelocationOutput(t, stdout.Bytes())
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Results))
	}
	result := got.Results[0]
	if result.Input.FromPath != "/src/file.txt" || result.Input.ToPath != "/dest/file-moved.txt" {
		t.Fatalf("input = %#v, want source and destination", result.Input)
	}
	if result.Result.Type != "file" || result.Result.PathDisplay != "/dest/file-moved.txt" {
		t.Fatalf("result = %#v, want moved file metadata", result.Result)
	}
	if result.Result.Size == nil || *result.Result.Size != 64 {
		t.Fatalf("size = %#v, want 64", result.Result.Size)
	}
}

func TestMvJSONMultipleSourcesOutputsMultipleResults(t *testing.T) {
	stubFilesClient(t, &mockFilesClient{
		moveV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			name := path.Base(arg.ToPath)
			metadata := files.NewFileMetadata(name, "id:"+name, time.Time{}, time.Time{}, "rev", 1)
			metadata.PathDisplay = arg.ToPath
			metadata.PathLower = strings.ToLower(arg.ToPath)
			return files.NewRelocationResult(metadata), nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)

	if err := mv(cmd, []string{"/src/a.txt", "/src/b.txt", "/dest"}); err != nil {
		t.Fatalf("mv error: %v", err)
	}

	got := decodeRelocationOutput(t, stdout.Bytes())
	if len(got.Results) != 2 {
		t.Fatalf("results = %d, want 2", len(got.Results))
	}
	if got.Results[0].Input.ToPath != "/dest/a.txt" || got.Results[1].Input.ToPath != "/dest/b.txt" {
		t.Fatalf("results = %#v, want folder destinations", got.Results)
	}
}

func TestMvJSONErrorUsesCommandStderr(t *testing.T) {
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, fmt.Errorf("path/not_found/")
		},
		moveV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			return nil, fmt.Errorf("path/malformed_path/")
		},
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, &stderr)

	err := mv(cmd, []string{"/src/file.txt", "/dest/file.txt"})
	if err == nil {
		t.Fatal("expected mv error")
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), `move "/src/file.txt" to "/dest/file.txt": path/malformed_path/`) {
		t.Fatalf("stderr = %q, want move API error", stderr.String())
	}
}

func TestMvCommandSupportsStructuredOutput(t *testing.T) {
	if !commandSupportsStructuredOutput(mvCmd) {
		t.Fatal("mv should support structured output")
	}
}
