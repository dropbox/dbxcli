package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

func TestMkdirArgValidation(t *testing.T) {
	err := mkdir(mkdirCmd, []string{})
	if err == nil {
		t.Error("expected error for no args")
	}
}

func TestMkdirTooManyArgs(t *testing.T) {
	err := mkdir(mkdirCmd, []string{"/a", "/b"})
	if err == nil {
		t.Error("expected error for too many args")
	}
}

func TestMkdirQuietByDefault(t *testing.T) {
	cmd, stdout := testMkdirCmd(t)
	folder := mkdirFolderMetadata("/Projects")
	var createdPath string

	mock := &mockFilesClient{
		createFolderV2Fn: func(arg *files.CreateFolderArg) (*files.CreateFolderResult, error) {
			createdPath = arg.Path
			return files.NewCreateFolderResult(folder), nil
		},
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			t.Fatalf("GetMetadata called on create success: %v", arg)
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	if err := mkdir(cmd, []string{"/Projects"}); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	if createdPath != "/Projects" {
		t.Fatalf("created path = %q, want /Projects", createdPath)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want quiet success", got)
	}
}

func TestMkdirJSONOutputsCreatedFolder(t *testing.T) {
	cmd, stdout := testMkdirCmd(t)
	setMkdirOutputJSON(t, cmd)
	folder := mkdirFolderMetadata("/Projects")

	mock := &mockFilesClient{
		createFolderV2Fn: func(arg *files.CreateFolderArg) (*files.CreateFolderResult, error) {
			return files.NewCreateFolderResult(folder), nil
		},
	}
	stubFilesClient(t, mock)

	if err := mkdir(cmd, []string{"/Projects"}); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}

	got := decodeMkdirOutput(t, stdout)
	if got.Input.Path != "/Projects" || got.Input.Parents {
		t.Fatalf("input = %#v, want path /Projects and parents false", got.Input)
	}
	result := got.Results[0]
	if result.Status != mkdirStatusCreated || result.Kind != mkdirKindFolder {
		t.Fatalf("status/kind = %s/%s, want created/folder", result.Status, result.Kind)
	}
	if result.Input.Path != "/Projects" || result.Input.Parents {
		t.Fatalf("result input = %#v, want path /Projects and parents false", result.Input)
	}
	if result.Result.Type != "folder" {
		t.Fatalf("result type = %q, want folder", result.Result.Type)
	}
	if result.Result.PathDisplay != "/Projects" {
		t.Fatalf("path_display = %q, want /Projects", result.Result.PathDisplay)
	}
	if result.Result.PathLower != "/projects" {
		t.Fatalf("path_lower = %q, want /projects", result.Result.PathLower)
	}
	if result.Result.ID != "id:folder" {
		t.Fatalf("id = %q, want id:folder", result.Result.ID)
	}
	if strings.Contains(stdout.String(), `"rev"`) || strings.Contains(stdout.String(), `"size"`) {
		t.Fatalf("folder JSON output = %s, want no file-only fields", stdout.String())
	}
}

func TestMkdirJSONParentsReturnsExistingFolderMetadata(t *testing.T) {
	cmd, stdout := testMkdirCmd(t)
	setMkdirOutputJSON(t, cmd)
	setMkdirParents(t, cmd)
	var createPath string
	var getPath string

	mock := &mockFilesClient{
		createFolderV2Fn: func(arg *files.CreateFolderArg) (*files.CreateFolderResult, error) {
			createPath = arg.Path
			return nil, createFolderConflictError(files.WriteConflictErrorFolder)
		},
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			getPath = arg.Path
			return mkdirFolderMetadata("/Existing"), nil
		},
	}
	stubFilesClient(t, mock)

	if err := mkdir(cmd, []string{"/Existing"}); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	if createPath != "/Existing" || getPath != "/Existing" {
		t.Fatalf("createPath = %q, getPath = %q; want /Existing", createPath, getPath)
	}

	got := decodeMkdirOutput(t, stdout)
	if got.Input.Path != "/Existing" || !got.Input.Parents {
		t.Fatalf("input = %#v, want path /Existing and parents true", got.Input)
	}
	result := got.Results[0]
	if result.Status != mkdirStatusExisting || result.Kind != mkdirKindFolder {
		t.Fatalf("status/kind = %s/%s, want existing/folder", result.Status, result.Kind)
	}
	if result.Result.Type != "folder" || result.Result.PathDisplay != "/Existing" {
		t.Fatalf("result = %#v, want existing folder metadata", result.Result)
	}
}

func TestMkdirParentsExistingFolderTextDoesNotFetchMetadata(t *testing.T) {
	cmd, stdout := testMkdirCmd(t)
	setMkdirParents(t, cmd)

	mock := &mockFilesClient{
		createFolderV2Fn: func(arg *files.CreateFolderArg) (*files.CreateFolderResult, error) {
			return nil, createFolderConflictError(files.WriteConflictErrorFolder)
		},
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			t.Fatalf("GetMetadata called for text existing folder success: %v", arg)
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	if err := mkdir(cmd, []string{"/Existing"}); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want quiet success", got)
	}
}

func TestMkdirParentsExistingFileReturnsError(t *testing.T) {
	cmd, stdout := testMkdirCmd(t)
	setMkdirParents(t, cmd)

	mock := &mockFilesClient{
		createFolderV2Fn: func(arg *files.CreateFolderArg) (*files.CreateFolderResult, error) {
			return nil, createFolderConflictError(files.WriteConflictErrorFile)
		},
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			t.Fatalf("GetMetadata called for typed existing file conflict: %v", arg)
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	err := mkdir(cmd, []string{"/Existing"})
	if err == nil {
		t.Fatal("expected error for existing file")
	}
	if !strings.Contains(err.Error(), "not a folder") {
		t.Fatalf("error = %q, want not a folder", err.Error())
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty output on error", got)
	}
}

func TestMkdirJSONErrorWritesNoOutput(t *testing.T) {
	cmd, stdout := testMkdirCmd(t)
	setMkdirOutputJSON(t, cmd)

	mock := &mockFilesClient{
		createFolderV2Fn: func(arg *files.CreateFolderArg) (*files.CreateFolderResult, error) {
			return nil, fmt.Errorf("create failed")
		},
	}
	stubFilesClient(t, mock)

	if err := mkdir(cmd, []string{"/Projects"}); err == nil {
		t.Fatal("expected mkdir error")
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty output on error", got)
	}
}

func TestMkdirJSONUsesInputPathWhenMetadataPathDisplayMissing(t *testing.T) {
	cmd, stdout := testMkdirCmd(t)
	setMkdirOutputJSON(t, cmd)

	mock := &mockFilesClient{
		createFolderV2Fn: func(arg *files.CreateFolderArg) (*files.CreateFolderResult, error) {
			return files.NewCreateFolderResult(&files.FolderMetadata{}), nil
		},
	}
	stubFilesClient(t, mock)

	if err := mkdir(cmd, []string{"/Projects"}); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}

	got := decodeMkdirOutput(t, stdout)
	if got.Results[0].Result.PathDisplay != "/Projects" {
		t.Fatalf("path_display = %q, want fallback input path", got.Results[0].Result.PathDisplay)
	}
}

func TestMkdirCommandSupportsStructuredOutput(t *testing.T) {
	if !commandSupportsStructuredOutput(mkdirCmd) {
		t.Fatal("mkdir command should support structured output")
	}
}

func TestIsConflictError(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"path/conflict/folder/", true},
		{"path/conflict/file/", true},
		{"path/not_found/", false},
		{"some other error", false},
	}

	for _, tt := range tests {
		got := isConflictError(fmt.Errorf("%s", tt.msg))
		if got != tt.want {
			t.Errorf("isConflictError(%q) = %v, want %v", tt.msg, got, tt.want)
		}
	}
}

func testMkdirCmd(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()

	var stdout bytes.Buffer
	cmd := &cobra.Command{Use: "mkdir"}
	cmd.SetOut(&stdout)
	cmd.Flags().BoolP("parents", "p", false, "")
	cmd.Flags().String(outputFlag, "text", "")
	return cmd, &stdout
}

func setMkdirOutputJSON(t *testing.T, cmd *cobra.Command) {
	t.Helper()

	if err := cmd.Flags().Set(outputFlag, "json"); err != nil {
		t.Fatal(err)
	}
}

func setMkdirParents(t *testing.T, cmd *cobra.Command) {
	t.Helper()

	if err := cmd.Flags().Set("parents", "true"); err != nil {
		t.Fatal(err)
	}
}

func mkdirFolderMetadata(path string) *files.FolderMetadata {
	return &files.FolderMetadata{
		Metadata: files.Metadata{
			Name:        strings.TrimPrefix(path, "/"),
			PathDisplay: path,
			PathLower:   strings.ToLower(path),
		},
		Id: "id:folder",
	}
}

func createFolderConflictError(conflictTag string) files.CreateFolderV2APIError {
	return files.CreateFolderV2APIError{
		APIError: dropbox.APIError{ErrorSummary: "path/conflict/" + conflictTag + "/"},
		EndpointError: &files.CreateFolderError{
			Tagged: dropbox.Tagged{Tag: files.CreateFolderErrorPath},
			Path: &files.WriteError{
				Tagged: dropbox.Tagged{Tag: files.WriteErrorConflict},
				Conflict: &files.WriteConflictError{
					Tagged: dropbox.Tagged{Tag: conflictTag},
				},
			},
		},
	}
}

type mkdirOutput struct {
	Input    mkdirInput    `json:"input"`
	Results  []mkdirResult `json:"results"`
	Warnings []jsonWarning `json:"warnings"`
}

func decodeMkdirOutput(t *testing.T, stdout *bytes.Buffer) mkdirOutput {
	t.Helper()

	var got mkdirOutput
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode JSON output: %v\noutput: %s", err, stdout.String())
	}
	if got.Warnings == nil {
		t.Fatalf("warnings = nil, want empty array")
	}
	if len(got.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want empty", got.Warnings)
	}
	if len(got.Results) != 1 {
		t.Fatalf("results len = %d, want 1", len(got.Results))
	}
	return got
}
