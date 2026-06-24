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

	"github.com/dropbox/dbxcli/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
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

func TestCpJSONOutputsRelocationResults(t *testing.T) {
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, fmt.Errorf("path/not_found/")
		},
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			metadata := files.NewFileMetadata("file-copy.txt", "id:file-copy", time.Time{}, time.Time{}, "rev-copy", 42)
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

func TestCpJSONMultipleSourcesOutputsMultipleResults(t *testing.T) {
	stubFilesClient(t, &mockFilesClient{
		copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
			name := path.Base(arg.ToPath)
			metadata := files.NewFileMetadata(name, "id:"+name, time.Time{}, time.Time{}, "rev", 1)
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
}

func TestCpCommandSupportsStructuredOutput(t *testing.T) {
	if !commandSupportsStructuredOutput(cpCmd) {
		t.Fatal("cp should support structured output")
	}
}

func newRelocationTestCommand(stdout, stderr *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String(outputFlag, string(output.FormatText), "")
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

type relocationOutput struct {
	Input    map[string]any     `json:"input"`
	Results  []relocationResult `json:"results"`
	Warnings []jsonWarning      `json:"warnings"`
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
