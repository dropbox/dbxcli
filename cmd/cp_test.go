package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dbxcli/v3/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

func stubFilesClient(t *testing.T, client filesClient) {
	t.Helper()

	origNew := filesNewFunc
	filesNewFunc = func(_ dropbox.Config) filesClient { return client }
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

func TestCpJSONOutputsRelocationResults(t *testing.T) {
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, fmt.Errorf("path/not_found/")
		},
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			metadata := files.NewFileMetadata("file-copy.txt", "id:file-copy", dropbox.DBXTime(time.Time{}), dropbox.DBXTime(time.Time{}), "rev-copy", 42)
			metadata.PathDisplay = arg.ToPath
			metadata.PathLower = strings.ToLower(arg.ToPath)
			return files.NewRelocationResult(metadata), nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)

	if err := cp(cmd, []string{"/src/file.txt", "/dest/file-copy.txt"}); err != nil {
		t.Fatalf("cp error: %v", err)
	}

	got := decodeRelocationOutput(t, stdout.Bytes())
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Results))
	}
	result := got.Results[0]
	if result.Input.FromPath != "/src/file.txt" || result.Input.ToPath != "/dest/file-copy.txt" {
		t.Fatalf("input = %#v, want source and destination", result.Input)
	}
	if result.Result.Type != "file" || result.Result.PathDisplay != "/dest/file-copy.txt" {
		t.Fatalf("result = %#v, want copied file metadata", result.Result)
	}
	if result.Result.Size == nil || *result.Result.Size != 42 {
		t.Fatalf("size = %#v, want 42", result.Result.Size)
	}
}

func TestCpDryRunTextOutputSnapshot(t *testing.T) {
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			switch arg.Path {
			case "/dest/file-copy.txt":
				return nil, relocationTestGetMetadataNotFoundError()
			case "/src/file.txt":
				return relocationTestFileMetadata("/src/file.txt", 42), nil
			default:
				t.Fatalf("unexpected GetMetadata path during dry-run: %q", arg.Path)
				return nil, nil
			}
		},
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			t.Fatalf("CopyV2 called during dry-run: %v", arg)
			return nil, nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTextTestCommand(&stdout, nil)
	if err := cmd.Flags().Set(dryRunFlagName, "true"); err != nil {
		t.Fatal(err)
	}

	if err := cp(cmd, []string{"/src/file.txt", "/dest/file-copy.txt"}); err != nil {
		t.Fatalf("cp error: %v", err)
	}

	const want = "Would copy /src/file.txt to /dest/file-copy.txt\n"
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestCpJSONDryRunOutputsPlannedResult(t *testing.T) {
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			switch arg.Path {
			case "/dest/file-copy.txt":
				return nil, relocationTestGetMetadataNotFoundError()
			case "/src/file.txt":
				return relocationTestFileMetadata("/src/file.txt", 42), nil
			default:
				t.Fatalf("unexpected GetMetadata path during dry-run: %q", arg.Path)
				return nil, nil
			}
		},
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			t.Fatalf("CopyV2 called during dry-run: %v", arg)
			return nil, nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set(dryRunFlagName, "true"); err != nil {
		t.Fatal(err)
	}

	if err := cp(cmd, []string{"/src/file.txt", "/dest/file-copy.txt"}); err != nil {
		t.Fatalf("cp error: %v", err)
	}

	got := decodeRelocationOutput(t, stdout.Bytes())
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Results))
	}
	result := got.Results[0]
	if result.Status != jsonStatusPlanned || result.Kind != "file" {
		t.Fatalf("status/kind = %s/%s, want planned/file", result.Status, result.Kind)
	}
	if result.Input.FromPath != "/src/file.txt" || result.Input.ToPath != "/dest/file-copy.txt" || !result.Input.DryRun {
		t.Fatalf("input = %#v, want source, destination, dry_run true", result.Input)
	}
	if result.Result.Type != "file" || result.Result.PathDisplay != "/dest/file-copy.txt" || result.Result.PathLower != "/dest/file-copy.txt" {
		t.Fatalf("result = %#v, want planned destination file metadata", result.Result)
	}
	if result.Result.Size == nil || *result.Result.Size != 42 {
		t.Fatalf("size = %#v, want 42", result.Result.Size)
	}
}

func TestCpDryRunMultipleSourcesOutputsPlansWithoutCopying(t *testing.T) {
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			switch arg.Path {
			case "/src/a.txt", "/src/b.txt":
				return relocationTestFileMetadata(arg.Path, 1), nil
			default:
				t.Fatalf("unexpected GetMetadata path during dry-run: %q", arg.Path)
				return nil, nil
			}
		},
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			t.Fatalf("CopyV2 called during dry-run: %v", arg)
			return nil, nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTextTestCommand(&stdout, nil)
	if err := cmd.Flags().Set(dryRunFlagName, "true"); err != nil {
		t.Fatal(err)
	}

	if err := cp(cmd, []string{"/src/a.txt", "/src/b.txt", "/dest"}); err != nil {
		t.Fatalf("cp error: %v", err)
	}

	const want = "Would copy /src/a.txt to /dest/a.txt\nWould copy /src/b.txt to /dest/b.txt\n"
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestCpDryRunSingleSourceExistingDestinationFolder(t *testing.T) {
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			switch arg.Path {
			case "/dest":
				return mkdirFolderMetadata("/dest"), nil
			case "/src/file.txt":
				return relocationTestFileMetadata("/src/file.txt", 42), nil
			default:
				t.Fatalf("unexpected GetMetadata path during dry-run: %q", arg.Path)
				return nil, nil
			}
		},
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			t.Fatalf("CopyV2 called during dry-run: %v", arg)
			return nil, nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set(dryRunFlagName, "true"); err != nil {
		t.Fatal(err)
	}

	if err := cp(cmd, []string{"/src/file.txt", "/dest"}); err != nil {
		t.Fatalf("cp error: %v", err)
	}

	got := decodeRelocationOutput(t, stdout.Bytes())
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Results))
	}
	if got.Results[0].Input.ToPath != "/dest/file.txt" {
		t.Fatalf("to_path = %q, want /dest/file.txt", got.Results[0].Input.ToPath)
	}
	if got.Results[0].Result.PathDisplay != "/dest/file.txt" {
		t.Fatalf("path_display = %q, want /dest/file.txt", got.Results[0].Result.PathDisplay)
	}
}

func TestCpJSONMultipleSourcesOutputsMultipleResults(t *testing.T) {
	stubFilesClient(t, &mockFilesClient{
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			name := path.Base(arg.ToPath)
			metadata := files.NewFileMetadata(name, "id:"+name, dropbox.DBXTime(time.Time{}), dropbox.DBXTime(time.Time{}), "rev", 1)
			metadata.PathDisplay = arg.ToPath
			metadata.PathLower = strings.ToLower(arg.ToPath)
			return files.NewRelocationResult(metadata), nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)

	if err := cp(cmd, []string{"/src/a.txt", "/src/b.txt", "/dest"}); err != nil {
		t.Fatalf("cp error: %v", err)
	}

	got := decodeRelocationOutput(t, stdout.Bytes())
	if len(got.Results) != 2 {
		t.Fatalf("results = %d, want 2", len(got.Results))
	}
	if got.Results[0].Input.ToPath != "/dest/a.txt" || got.Results[1].Input.ToPath != "/dest/b.txt" {
		t.Fatalf("results = %#v, want folder destinations", got.Results)
	}
}

func TestCpJSONErrorUsesCommandStderr(t *testing.T) {
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, fmt.Errorf("path/not_found/")
		},
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			return nil, fmt.Errorf("path/malformed_path/")
		},
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, &stderr)

	err := cp(cmd, []string{"/src/file.txt", "/dest/file.txt"})
	if err == nil {
		t.Fatal("expected cp error")
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), `copy "/src/file.txt" to "/dest/file.txt": path/malformed_path/`) {
		t.Fatalf("stderr = %q, want copy API error", stderr.String())
	}
	details := jsonErrorDetails(err)
	if details["operation"] != "copy" || details["from_path"] != "/src/file.txt" || details["to_path"] != "/dest/file.txt" {
		t.Fatalf("details = %#v, want copy from/to paths", details)
	}
}

func TestCpCommandDefinesIfExistsFlag(t *testing.T) {
	flag := cpCmd.Flags().Lookup("if-exists")
	if flag == nil {
		t.Fatal("cp should define --if-exists")
	}
	if flag.DefValue != relocationIfExistsFail {
		t.Fatalf("--if-exists default = %q, want %q", flag.DefValue, relocationIfExistsFail)
	}
	if cpCmd.Flags().Lookup(dryRunFlagName) == nil {
		t.Fatalf("cp should define --%s", dryRunFlagName)
	}
}

func TestCpInvalidIfExistsReturnsInvalidArguments(t *testing.T) {
	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", "replace"); err != nil {
		t.Fatal(err)
	}

	err := cp(cmd, []string{"/src/file.txt", "/dest/file.txt"})
	if err == nil {
		t.Fatal("expected cp error")
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

func TestCpIfExistsFailCallsCopy(t *testing.T) {
	var copied []*files.RelocationArg
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, relocationTestGetMetadataNotFoundError()
		},
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			copied = append(copied, arg)
			return files.NewRelocationResult(relocationTestFileMetadata(arg.ToPath, 1)), nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", relocationIfExistsFail); err != nil {
		t.Fatal(err)
	}
	if err := cp(cmd, []string{"/src/file.txt", "/dest/file.txt"}); err != nil {
		t.Fatalf("cp error: %v", err)
	}
	if len(copied) != 1 {
		t.Fatalf("copied = %d, want 1", len(copied))
	}
}

func TestCpIfExistsAutorenameSetsAutorenameFlag(t *testing.T) {
	var copied []*files.RelocationArg
	stubFilesClient(t, &mockFilesClient{
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			copied = append(copied, arg)
			metadata := relocationTestFileMetadata("/dest/file (1).txt", 1)
			return files.NewRelocationResult(metadata), nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", relocationIfExistsAutorename); err != nil {
		t.Fatal(err)
	}

	if err := cp(cmd, []string{"/src/file.txt", "/dest/file.txt"}); err != nil {
		t.Fatalf("cp error: %v", err)
	}
	if len(copied) != 1 {
		t.Fatalf("copied = %d, want 1", len(copied))
	}
	if !copied[0].Autorename {
		t.Fatal("CopyV2 arg Autorename = false, want true")
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

func TestCpIfExistsAutorenameNoConflictKeepsCopiedStatus(t *testing.T) {
	stubFilesClient(t, &mockFilesClient{
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			return files.NewRelocationResult(relocationTestFileMetadata(arg.ToPath, 1)), nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", relocationIfExistsAutorename); err != nil {
		t.Fatal(err)
	}

	if err := cp(cmd, []string{"/src/file.txt", "/dest/file.txt"}); err != nil {
		t.Fatalf("cp error: %v", err)
	}
	got := decodeRelocationOutput(t, stdout.Bytes())
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Results))
	}
	if got.Results[0].Status != relocationJSONStatusCopied {
		t.Fatalf("status = %q, want copied", got.Results[0].Status)
	}
}

func TestCpIfExistsAutorenameCaseOnlyPathKeepsCopiedStatus(t *testing.T) {
	stubFilesClient(t, &mockFilesClient{
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			return files.NewRelocationResult(relocationTestFileMetadata("/Dest/File.txt", 1)), nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", relocationIfExistsAutorename); err != nil {
		t.Fatal(err)
	}

	if err := cp(cmd, []string{"/src/file.txt", "/dest/file.txt"}); err != nil {
		t.Fatalf("cp error: %v", err)
	}
	got := decodeRelocationOutput(t, stdout.Bytes())
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Results))
	}
	if got.Results[0].Status != relocationJSONStatusCopied {
		t.Fatalf("status = %q, want copied", got.Results[0].Status)
	}
	if got.Results[0].Result.PathDisplay != "/Dest/File.txt" {
		t.Fatalf("path_display = %q, want /Dest/File.txt", got.Results[0].Result.PathDisplay)
	}
}

func TestCpIfExistsSkipExistingDestinationDoesNotCopy(t *testing.T) {
	var copied bool
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			if arg.Path != "/dest/file.txt" {
				t.Fatalf("metadata path = %q, want /dest/file.txt", arg.Path)
			}
			return relocationTestFileMetadata(arg.Path, 8), nil
		},
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			copied = true
			return nil, nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", relocationIfExistsSkip); err != nil {
		t.Fatal(err)
	}

	if err := cp(cmd, []string{"/src/file.txt", "/dest/file.txt"}); err != nil {
		t.Fatalf("cp error: %v", err)
	}
	if copied {
		t.Fatal("CopyV2 called for skipped destination")
	}
	got := decodeRelocationOutput(t, stdout.Bytes())
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Results))
	}
	if got.Results[0].Status != relocationJSONStatusSkipped {
		t.Fatalf("status = %q, want skipped", got.Results[0].Status)
	}
	if got.Results[0].Input.FromPath != "/src/file.txt" || got.Results[0].Input.ToPath != "/dest/file.txt" {
		t.Fatalf("input = %#v, want source and destination", got.Results[0].Input)
	}
}

func TestCpIfExistsSkipMissingDestinationCopies(t *testing.T) {
	var copied []*files.RelocationArg
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, relocationTestGetMetadataNotFoundError()
		},
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			copied = append(copied, arg)
			return files.NewRelocationResult(relocationTestFileMetadata(arg.ToPath, 3)), nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", relocationIfExistsSkip); err != nil {
		t.Fatal(err)
	}

	if err := cp(cmd, []string{"/src/file.txt", "/dest/file.txt"}); err != nil {
		t.Fatalf("cp error: %v", err)
	}
	if len(copied) != 1 {
		t.Fatalf("copied = %d, want 1", len(copied))
	}
	got := decodeRelocationOutput(t, stdout.Bytes())
	if got.Results[0].Status != relocationJSONStatusCopied {
		t.Fatalf("status = %q, want copied", got.Results[0].Status)
	}
}

func TestCpIfExistsSkipConvertsDestinationConflict(t *testing.T) {
	getMetadataCalls := 0
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			getMetadataCalls++
			if getMetadataCalls < 3 {
				return nil, relocationTestGetMetadataNotFoundError()
			}
			return relocationTestFileMetadata(arg.Path, 13), nil
		},
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			return nil, relocationTestCopyDestinationConflictError()
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", relocationIfExistsSkip); err != nil {
		t.Fatal(err)
	}

	if err := cp(cmd, []string{"/src/file.txt", "/dest/file.txt"}); err != nil {
		t.Fatalf("cp error: %v", err)
	}
	got := decodeRelocationOutput(t, stdout.Bytes())
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Results))
	}
	if got.Results[0].Status != relocationJSONStatusSkipped {
		t.Fatalf("status = %q, want skipped", got.Results[0].Status)
	}
}

func TestCpIfExistsSkipMultipleSourcesAppliesPerTarget(t *testing.T) {
	var copied []string
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
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			copied = append(copied, arg.ToPath)
			return files.NewRelocationResult(relocationTestFileMetadata(arg.ToPath, 2)), nil
		},
	})

	var stdout bytes.Buffer
	cmd := newRelocationTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", relocationIfExistsSkip); err != nil {
		t.Fatal(err)
	}

	if err := cp(cmd, []string{"/src/a.txt", "/src/b.txt", "/dest"}); err != nil {
		t.Fatalf("cp error: %v", err)
	}
	if len(copied) != 1 || copied[0] != "/dest/b.txt" {
		t.Fatalf("copied = %#v, want only /dest/b.txt", copied)
	}
	got := decodeRelocationOutput(t, stdout.Bytes())
	if len(got.Results) != 2 {
		t.Fatalf("results = %d, want 2", len(got.Results))
	}
	if got.Results[0].Status != relocationJSONStatusSkipped || got.Results[1].Status != relocationJSONStatusCopied {
		t.Fatalf("statuses = %q, %q; want skipped, copied", got.Results[0].Status, got.Results[1].Status)
	}
}

func TestCpIfExistsSkipTextModeQuiet(t *testing.T) {
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return relocationTestFileMetadata(arg.Path, 8), nil
		},
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			t.Fatal("CopyV2 called for skipped destination")
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

	if err := cp(cmd, []string{"/src/file.txt", "/dest/file.txt"}); err != nil {
		t.Fatalf("cp error: %v", err)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestCpCommandSupportsStructuredOutput(t *testing.T) {
	if !commandSupportsStructuredOutput(cpCmd) {
		t.Fatal("cp should support structured output")
	}
}

func newRelocationTestCommand(stdout, stderr *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String(outputFlag, string(output.FormatText), "")
	cmd.Flags().String("if-exists", relocationIfExistsFail, "")
	addDryRunFlag(cmd)
	if err := cmd.Flags().Set(outputFlag, string(output.FormatJSON)); err != nil {
		panic(err)
	}
	if stdout != nil {
		cmd.SetOut(stdout)
	}
	if stderr != nil {
		cmd.SetErr(stderr)
	}
	return cmd
}

func newRelocationTextTestCommand(stdout, stderr *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String(outputFlag, string(output.FormatText), "")
	cmd.Flags().String("if-exists", relocationIfExistsFail, "")
	addDryRunFlag(cmd)
	if stdout != nil {
		cmd.SetOut(stdout)
	}
	if stderr != nil {
		cmd.SetErr(stderr)
	}
	return cmd
}

type relocationOutput struct {
	Input    map[string]any         `json:"input"`
	Results  []relocationJSONResult `json:"results"`
	Warnings []jsonWarning          `json:"warnings"`
}

type relocationJSONResult struct {
	Status string          `json:"status"`
	Kind   string          `json:"kind"`
	Input  relocationInput `json:"input"`
	Result jsonMetadata    `json:"result"`
}

func decodeRelocationOutput(t *testing.T, data []byte) relocationOutput {
	t.Helper()
	var got relocationOutput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("decode JSON output: %v\noutput: %s", err, string(data))
	}
	if got.Input == nil {
		t.Fatalf("input = nil, want empty object")
	}
	if len(got.Input) != 0 {
		t.Fatalf("input = %+v, want empty object", got.Input)
	}
	if got.Warnings == nil {
		t.Fatalf("warnings = nil, want empty array")
	}
	if len(got.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want empty", got.Warnings)
	}
	return got
}

func relocationTestFileMetadata(pathDisplay string, size uint64) *files.FileMetadata {
	metadata := files.NewFileMetadata(path.Base(pathDisplay), "id:"+path.Base(pathDisplay), dropbox.DBXTime(time.Time{}), dropbox.DBXTime(time.Time{}), "rev", size)
	metadata.PathDisplay = pathDisplay
	metadata.PathLower = strings.ToLower(pathDisplay)
	return metadata
}

func relocationTestGetMetadataNotFoundError() error {
	return files.GetMetadataAPIError{
		EndpointError: &files.GetMetadataError{
			Tagged: dropbox.Tagged{Tag: files.GetMetadataErrorPath},
			Path:   &files.LookupError{Tagged: dropbox.Tagged{Tag: files.LookupErrorNotFound}},
		},
	}
}

func relocationTestCopyDestinationConflictError() error {
	return files.CopyV2APIError{
		EndpointError: relocationTestDestinationConflictError(),
	}
}

func relocationTestMoveDestinationConflictError() error {
	return files.MoveV2APIError{
		EndpointError: relocationTestDestinationConflictError(),
	}
}

func relocationTestDestinationConflictError() *files.RelocationError {
	return &files.RelocationError{
		Tagged: dropbox.Tagged{Tag: files.RelocationErrorTo},
		To: &files.WriteError{
			Tagged:   dropbox.Tagged{Tag: files.WriteErrorConflict},
			Conflict: &files.WriteConflictError{Tagged: dropbox.Tagged{Tag: files.WriteConflictErrorFile}},
		},
	}
}
