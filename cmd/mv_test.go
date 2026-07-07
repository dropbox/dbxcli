package cmd

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dbxcli/v3/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
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
			metadata := files.NewFileMetadata("file-moved.txt", "id:file-moved", dropbox.DBXTime(time.Time{}), dropbox.DBXTime(time.Time{}), "rev-moved", 64)
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
			metadata := files.NewFileMetadata(name, "id:"+name, dropbox.DBXTime(time.Time{}), dropbox.DBXTime(time.Time{}), "rev", 1)
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
	details := jsonErrorDetails(err)
	if details["operation"] != "move" || details["from_path"] != "/src/file.txt" || details["to_path"] != "/dest/file.txt" {
		t.Fatalf("details = %#v, want move from/to paths", details)
	}
}

func TestMvCommandDefinesIfExistsFlag(t *testing.T) {
	flag := mvCmd.Flags().Lookup("if-exists")
	if flag == nil {
		t.Fatal("mv should define --if-exists")
	}
	if flag.DefValue != relocationIfExistsFail {
		t.Fatalf("--if-exists default = %q, want %q", flag.DefValue, relocationIfExistsFail)
	}
}

func TestMvInvalidIfExistsReturnsInvalidArguments(t *testing.T) {
	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", "replace"); err != nil {
		t.Fatal(err)
	}

	err := mv(cmd, []string{"/src/file.txt", "/dest/file.txt"})
	if err == nil {
		t.Fatal("expected mv error")
	}
	if code := jsonErrorCode(err); code != jsonErrorCodeInvalidArguments {
		t.Fatalf("json error code = %q, want %q", code, jsonErrorCodeInvalidArguments)
	}
	details := jsonErrorDetails(err)
	if details["flag"] != "if-exists" || details["value"] != "replace" {
		t.Fatalf("details = %#v, want if-exists flag value", details)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestMvIfExistsFailCallsMove(t *testing.T) {
	var moved []*files.RelocationArg
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, relocationTestGetMetadataNotFoundError()
		},
		moveV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			moved = append(moved, arg)
			return files.NewRelocationResult(relocationTestFileMetadata(arg.ToPath, 1)), nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", relocationIfExistsFail); err != nil {
		t.Fatal(err)
	}
	if err := mv(cmd, []string{"/src/file.txt", "/dest/file.txt"}); err != nil {
		t.Fatalf("mv error: %v", err)
	}
	if len(moved) != 1 {
		t.Fatalf("moved = %d, want 1", len(moved))
	}
}

func TestMvIfExistsAutorenameSetsAutorenameFlag(t *testing.T) {
	var moved []*files.RelocationArg
	stubFilesClient(t, &mockFilesClient{
		moveV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			moved = append(moved, arg)
			return files.NewRelocationResult(relocationTestFileMetadata("/dest/file (1).txt", 1)), nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", relocationIfExistsAutorename); err != nil {
		t.Fatal(err)
	}
	if err := mv(cmd, []string{"/src/file.txt", "/dest/file.txt"}); err != nil {
		t.Fatalf("mv error: %v", err)
	}
	if len(moved) != 1 {
		t.Fatalf("moved = %d, want 1", len(moved))
	}
	if !moved[0].Autorename {
		t.Fatal("MoveV2 arg Autorename = false, want true")
	}
	got := decodeRelocationOutput(t, stdout.Bytes())
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Results))
	}
	if got.Results[0].Status != relocationJSONStatusAutorenamed {
		t.Fatalf("status = %q, want autorenamed", got.Results[0].Status)
	}
	if got.Results[0].Result.PathDisplay != "/dest/file (1).txt" {
		t.Fatalf("path_display = %q, want /dest/file (1).txt", got.Results[0].Result.PathDisplay)
	}
}

func TestMvIfExistsAutorenameCaseOnlyPathKeepsMovedStatus(t *testing.T) {
	stubFilesClient(t, &mockFilesClient{
		moveV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			return files.NewRelocationResult(relocationTestFileMetadata("/Dest/File.txt", 1)), nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", relocationIfExistsAutorename); err != nil {
		t.Fatal(err)
	}
	if err := mv(cmd, []string{"/src/file.txt", "/dest/file.txt"}); err != nil {
		t.Fatalf("mv error: %v", err)
	}
	got := decodeRelocationOutput(t, stdout.Bytes())
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Results))
	}
	if got.Results[0].Status != relocationJSONStatusMoved {
		t.Fatalf("status = %q, want moved", got.Results[0].Status)
	}
	if got.Results[0].Result.PathDisplay != "/Dest/File.txt" {
		t.Fatalf("path_display = %q, want /Dest/File.txt", got.Results[0].Result.PathDisplay)
	}
}

func TestMvIfExistsSkipExistingDestinationDoesNotMove(t *testing.T) {
	var moved bool
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			if arg.Path != "/dest/file.txt" {
				t.Fatalf("metadata path = %q, want /dest/file.txt", arg.Path)
			}
			return relocationTestFileMetadata(arg.Path, 8), nil
		},
		moveV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			moved = true
			return nil, nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", relocationIfExistsSkip); err != nil {
		t.Fatal(err)
	}

	if err := mv(cmd, []string{"/src/file.txt", "/dest/file.txt"}); err != nil {
		t.Fatalf("mv error: %v", err)
	}
	if moved {
		t.Fatal("MoveV2 called for skipped destination")
	}
	got := decodeRelocationOutput(t, stdout.Bytes())
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Results))
	}
	if got.Results[0].Status != relocationJSONStatusSkipped {
		t.Fatalf("status = %q, want skipped", got.Results[0].Status)
	}
}

func TestMvIfExistsSkipMissingDestinationMoves(t *testing.T) {
	var moved []*files.RelocationArg
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, relocationTestGetMetadataNotFoundError()
		},
		moveV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			moved = append(moved, arg)
			return files.NewRelocationResult(relocationTestFileMetadata(arg.ToPath, 3)), nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", relocationIfExistsSkip); err != nil {
		t.Fatal(err)
	}

	if err := mv(cmd, []string{"/src/file.txt", "/dest/file.txt"}); err != nil {
		t.Fatalf("mv error: %v", err)
	}
	if len(moved) != 1 {
		t.Fatalf("moved = %d, want 1", len(moved))
	}
	got := decodeRelocationOutput(t, stdout.Bytes())
	if got.Results[0].Status != relocationJSONStatusMoved {
		t.Fatalf("status = %q, want moved", got.Results[0].Status)
	}
}

func TestMvIfExistsSkipConvertsDestinationConflict(t *testing.T) {
	getMetadataCalls := 0
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			getMetadataCalls++
			if getMetadataCalls < 3 {
				return nil, relocationTestGetMetadataNotFoundError()
			}
			return relocationTestFileMetadata(arg.Path, 13), nil
		},
		moveV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			return nil, relocationTestMoveDestinationConflictError()
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", relocationIfExistsSkip); err != nil {
		t.Fatal(err)
	}

	if err := mv(cmd, []string{"/src/file.txt", "/dest/file.txt"}); err != nil {
		t.Fatalf("mv error: %v", err)
	}
	got := decodeRelocationOutput(t, stdout.Bytes())
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Results))
	}
	if got.Results[0].Status != relocationJSONStatusSkipped {
		t.Fatalf("status = %q, want skipped", got.Results[0].Status)
	}
}

func TestMvIfExistsSkipMultipleSourcesAppliesPerTarget(t *testing.T) {
	var moved []string
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			switch arg.Path {
			case "/dest/a.txt":
				return relocationTestFileMetadata(arg.Path, 1), nil
			case "/dest/b.txt":
				return nil, relocationTestGetMetadataNotFoundError()
			default:
				t.Fatalf("unexpected metadata path %q", arg.Path)
				return nil, nil
			}
		},
		moveV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			moved = append(moved, arg.ToPath)
			return files.NewRelocationResult(relocationTestFileMetadata(arg.ToPath, 2)), nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", relocationIfExistsSkip); err != nil {
		t.Fatal(err)
	}

	if err := mv(cmd, []string{"/src/a.txt", "/src/b.txt", "/dest"}); err != nil {
		t.Fatalf("mv error: %v", err)
	}
	if len(moved) != 1 || moved[0] != "/dest/b.txt" {
		t.Fatalf("moved = %#v, want only /dest/b.txt", moved)
	}
	got := decodeRelocationOutput(t, stdout.Bytes())
	if len(got.Results) != 2 {
		t.Fatalf("results = %d, want 2", len(got.Results))
	}
	if got.Results[0].Status != relocationJSONStatusSkipped || got.Results[1].Status != relocationJSONStatusMoved {
		t.Fatalf("statuses = %q, %q; want skipped, moved", got.Results[0].Status, got.Results[1].Status)
	}
}

func TestMvIfExistsSkipTextModeQuiet(t *testing.T) {
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return relocationTestFileMetadata(arg.Path, 8), nil
		},
		moveV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			t.Fatal("MoveV2 called for skipped destination")
			return nil, nil
		},
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().String(outputFlag, string(output.FormatText), "")
	cmd.Flags().String("if-exists", relocationIfExistsFail, "")
	if err := cmd.Flags().Set("if-exists", relocationIfExistsSkip); err != nil {
		t.Fatal(err)
	}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := mv(cmd, []string{"/src/file.txt", "/dest/file.txt"}); err != nil {
		t.Fatalf("mv error: %v", err)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestMvCommandSupportsStructuredOutput(t *testing.T) {
	if !commandSupportsStructuredOutput(mvCmd) {
		t.Fatal("mv should support structured output")
	}
}
