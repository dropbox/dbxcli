package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

func testRmCmd(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()

	var stdout bytes.Buffer
	cmd := &cobra.Command{Use: "rm"}
	cmd.SetOut(&stdout)
	cmd.Flags().BoolP("force", "f", false, "")
	cmd.Flags().BoolP("recursive", "r", false, "")
	cmd.Flags().Bool("permanent", false, "")
	cmd.Flags().BoolP("verbose", "v", false, "")
	cmd.Flags().String(outputFlag, "text", "")
	return cmd, &stdout
}

func setRmFlag(t *testing.T, cmd *cobra.Command, name string) {
	t.Helper()

	if err := cmd.Flags().Set(name, "true"); err != nil {
		t.Fatal(err)
	}
}

func rmFileMetadata(path string) *files.FileMetadata {
	return &files.FileMetadata{
		Metadata: files.Metadata{
			Name:        strings.TrimPrefix(path, "/"),
			PathDisplay: path,
			PathLower:   strings.ToLower(path),
		},
		Id:   "id:file",
		Rev:  "rev",
		Size: 123,
	}
}

func rmFolderMetadata(path string) *files.FolderMetadata {
	return &files.FolderMetadata{
		Metadata: files.Metadata{
			Name:        strings.TrimPrefix(path, "/"),
			PathDisplay: path,
			PathLower:   strings.ToLower(path),
		},
		Id: "id:folder",
	}
}

func setRmOutputJSON(t *testing.T, cmd *cobra.Command) {
	t.Helper()

	if err := cmd.Flags().Set(outputFlag, "json"); err != nil {
		t.Fatal(err)
	}
}

func decodeRemoveOutput(t *testing.T, stdout *bytes.Buffer) removeOutput {
	t.Helper()

	return decodeRemoveOutputString(t, stdout.String())
}

func decodeRemoveOutputString(t *testing.T, output string) removeOutput {
	t.Helper()

	var got removeOutput
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatalf("decode JSON output: %v\noutput: %s", err, output)
	}
	return got
}

func rmNonEmptyFolderResult() *files.ListFolderResult {
	return &files.ListFolderResult{
		Entries: []files.IsMetadata{rmFileMetadata("/folder/file.txt")},
	}
}

func TestRmFileDeletesWithDeleteV2(t *testing.T) {
	cmd, stdout := testRmCmd(t)
	file := rmFileMetadata("/file.txt")
	var deleted []string

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return file, nil
		},
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			t.Fatalf("ListFolder called for file: %v", arg)
			return nil, nil
		},
		deleteV2Fn: func(arg *files.DeleteArg) (*files.DeleteResult, error) {
			deleted = append(deleted, arg.Path)
			return files.NewDeleteResult(file), nil
		},
		permanentlyDeleteFn: func(arg *files.DeleteArg) error {
			t.Fatalf("PermanentlyDelete called for normal delete: %v", arg)
			return nil
		},
	}
	stubFilesClient(t, mock)

	if err := rm(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("rm error: %v", err)
	}
	if len(deleted) != 1 || deleted[0] != "/file.txt" {
		t.Fatalf("deleted = %v, want [/file.txt]", deleted)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want quiet success", got)
	}
}

func TestRmNonEmptyFolderRequiresRecursiveOrForce(t *testing.T) {
	cmd, _ := testRmCmd(t)
	deleteCalled := false

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return rmFolderMetadata("/folder"), nil
		},
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			return rmNonEmptyFolderResult(), nil
		},
		deleteV2Fn: func(arg *files.DeleteArg) (*files.DeleteResult, error) {
			deleteCalled = true
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	err := rm(cmd, []string{"/folder"})
	if err == nil {
		t.Fatal("expected error for non-empty folder without recursive or force")
	}
	if !strings.Contains(err.Error(), "--recursive") || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("error = %q, want recursive and force guidance", err.Error())
	}
	if deleteCalled {
		t.Fatal("delete called after validation failure")
	}
}

func TestRmRecursiveDeletesNonEmptyFolder(t *testing.T) {
	cmd, _ := testRmCmd(t)
	setRmFlag(t, cmd, "recursive")
	var deleted []string
	folder := rmFolderMetadata("/folder")

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return folder, nil
		},
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			t.Fatalf("ListFolder called when --recursive already allows folder removal: %v", arg)
			return nil, nil
		},
		deleteV2Fn: func(arg *files.DeleteArg) (*files.DeleteResult, error) {
			deleted = append(deleted, arg.Path)
			return files.NewDeleteResult(folder), nil
		},
	}
	stubFilesClient(t, mock)

	if err := rm(cmd, []string{"/folder"}); err != nil {
		t.Fatalf("rm error: %v", err)
	}
	if len(deleted) != 1 || deleted[0] != "/folder" {
		t.Fatalf("deleted = %v, want [/folder]", deleted)
	}
}

func TestRmForceStillDeletesNonEmptyFolder(t *testing.T) {
	cmd, _ := testRmCmd(t)
	setRmFlag(t, cmd, "force")
	var deleted []string
	folder := rmFolderMetadata("/folder")

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return folder, nil
		},
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			t.Fatalf("ListFolder called when --force already allows folder removal: %v", arg)
			return nil, nil
		},
		deleteV2Fn: func(arg *files.DeleteArg) (*files.DeleteResult, error) {
			deleted = append(deleted, arg.Path)
			return files.NewDeleteResult(folder), nil
		},
	}
	stubFilesClient(t, mock)

	if err := rm(cmd, []string{"/folder"}); err != nil {
		t.Fatalf("rm error: %v", err)
	}
	if len(deleted) != 1 || deleted[0] != "/folder" {
		t.Fatalf("deleted = %v, want [/folder]", deleted)
	}
}

func TestRemoveResultInputKeepsForceAndRecursiveSeparate(t *testing.T) {
	result := newRemoveResult("/folder", rmFolderMetadata("/folder"), removeOptions{force: true})
	if !result.Input.Force {
		t.Fatal("Force = false, want true")
	}
	if result.Input.Recursive {
		t.Fatal("Recursive = true for force-only delete, want false")
	}

	result = newRemoveResult("/folder", rmFolderMetadata("/folder"), removeOptions{recursive: true})
	if result.Input.Force {
		t.Fatal("Force = true for recursive-only delete, want false")
	}
	if !result.Input.Recursive {
		t.Fatal("Recursive = false, want true")
	}
}

func TestRmPermanentCallsPermanentlyDelete(t *testing.T) {
	cmd, stdout := testRmCmd(t)
	setRmFlag(t, cmd, "permanent")
	file := rmFileMetadata("/file.txt")
	var permanentlyDeleted []string

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return file, nil
		},
		deleteV2Fn: func(arg *files.DeleteArg) (*files.DeleteResult, error) {
			t.Fatalf("DeleteV2 called for permanent delete: %v", arg)
			return nil, nil
		},
		permanentlyDeleteFn: func(arg *files.DeleteArg) error {
			permanentlyDeleted = append(permanentlyDeleted, arg.Path)
			return nil
		},
	}
	stubFilesClient(t, mock)

	if err := rm(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("rm error: %v", err)
	}
	if len(permanentlyDeleted) != 1 || permanentlyDeleted[0] != "/file.txt" {
		t.Fatalf("permanentlyDeleted = %v, want [/file.txt]", permanentlyDeleted)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want quiet success", got)
	}
}

func TestRmPermanentRecursiveDeletesNonEmptyFolder(t *testing.T) {
	cmd, _ := testRmCmd(t)
	setRmFlag(t, cmd, "permanent")
	setRmFlag(t, cmd, "recursive")
	folder := rmFolderMetadata("/folder")
	var permanentlyDeleted []string

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return folder, nil
		},
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			t.Fatalf("ListFolder called when --recursive already allows folder removal: %v", arg)
			return nil, nil
		},
		deleteV2Fn: func(arg *files.DeleteArg) (*files.DeleteResult, error) {
			t.Fatalf("DeleteV2 called for permanent delete: %v", arg)
			return nil, nil
		},
		permanentlyDeleteFn: func(arg *files.DeleteArg) error {
			permanentlyDeleted = append(permanentlyDeleted, arg.Path)
			return nil
		},
	}
	stubFilesClient(t, mock)

	if err := rm(cmd, []string{"/folder"}); err != nil {
		t.Fatalf("rm error: %v", err)
	}
	if len(permanentlyDeleted) != 1 || permanentlyDeleted[0] != "/folder" {
		t.Fatalf("permanentlyDeleted = %v, want [/folder]", permanentlyDeleted)
	}
}

func TestRmValidatesAllTargetsBeforeDeleting(t *testing.T) {
	cmd, _ := testRmCmd(t)
	deleted := false

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			switch arg.Path {
			case "/file.txt":
				return rmFileMetadata("/file.txt"), nil
			case "/folder":
				return rmFolderMetadata("/folder"), nil
			default:
				return nil, fmt.Errorf("unexpected metadata path %q", arg.Path)
			}
		},
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			return rmNonEmptyFolderResult(), nil
		},
		deleteV2Fn: func(arg *files.DeleteArg) (*files.DeleteResult, error) {
			deleted = true
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	err := rm(cmd, []string{"/file.txt", "/folder"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if deleted {
		t.Fatal("deleted first target before validating later target")
	}
}

func TestRmVerbosePrintsDeleteResults(t *testing.T) {
	cmd, stdout := testRmCmd(t)
	setRmFlag(t, cmd, "verbose")
	file := rmFileMetadata("/File.txt")

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return file, nil
		},
		deleteV2Fn: func(arg *files.DeleteArg) (*files.DeleteResult, error) {
			return files.NewDeleteResult(file), nil
		},
	}
	stubFilesClient(t, mock)

	if err := rm(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("rm error: %v", err)
	}
	if got, want := stdout.String(), "Deleted /File.txt\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRmVerboseUsesInheritedFlag(t *testing.T) {
	var stdout bytes.Buffer
	root := &cobra.Command{Use: "dbxcli"}
	root.PersistentFlags().BoolP("verbose", "v", false, "")
	cmd := &cobra.Command{
		Use:  "rm",
		RunE: rm,
	}
	cmd.SetOut(&stdout)
	cmd.Flags().BoolP("force", "f", false, "")
	cmd.Flags().BoolP("recursive", "r", false, "")
	cmd.Flags().Bool("permanent", false, "")
	root.AddCommand(cmd)
	root.SetArgs([]string{"rm", "--verbose", "/file.txt"})

	file := rmFileMetadata("/File.txt")
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return file, nil
		},
		deleteV2Fn: func(arg *files.DeleteArg) (*files.DeleteResult, error) {
			return files.NewDeleteResult(file), nil
		},
	}
	stubFilesClient(t, mock)

	if err := root.Execute(); err != nil {
		t.Fatalf("rm error: %v", err)
	}
	if got, want := stdout.String(), "Deleted /File.txt\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRmVerbosePrintsPermanentDeleteResults(t *testing.T) {
	cmd, stdout := testRmCmd(t)
	setRmFlag(t, cmd, "permanent")
	setRmFlag(t, cmd, "verbose")
	file := rmFileMetadata("/File.txt")

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return file, nil
		},
		permanentlyDeleteFn: func(arg *files.DeleteArg) error {
			return nil
		},
	}
	stubFilesClient(t, mock)

	if err := rm(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("rm error: %v", err)
	}
	if got, want := stdout.String(), "Permanently deleted /File.txt\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRmJSONDeletesFile(t *testing.T) {
	cmd, stdout := testRmCmd(t)
	setRmOutputJSON(t, cmd)
	file := rmFileMetadata("/File.txt")
	deletedFile := rmFileMetadata("/File.txt")
	deletedFile.Rev = "deleted-rev"
	deletedFile.Size = 456

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return file, nil
		},
		deleteV2Fn: func(arg *files.DeleteArg) (*files.DeleteResult, error) {
			return files.NewDeleteResult(deletedFile), nil
		},
	}
	stubFilesClient(t, mock)

	if err := rm(cmd, []string{"/File.txt"}); err != nil {
		t.Fatalf("rm error: %v", err)
	}

	got := decodeRemoveOutput(t, stdout)
	if len(got.Results) != 1 {
		t.Fatalf("results len = %d, want 1", len(got.Results))
	}
	result := got.Results[0]
	if result.Input.Path != "/File.txt" {
		t.Fatalf("input path = %q, want /File.txt", result.Input.Path)
	}
	if result.Input.Permanent || result.Input.Recursive || result.Input.Force {
		t.Fatalf("input flags = %+v, want all false", result.Input)
	}
	if result.Result.Type != "file" {
		t.Fatalf("result type = %q, want file", result.Result.Type)
	}
	if result.Result.PathDisplay != "/File.txt" {
		t.Fatalf("path_display = %q, want /File.txt", result.Result.PathDisplay)
	}
	if result.Result.PathLower != "/file.txt" {
		t.Fatalf("path_lower = %q, want /file.txt", result.Result.PathLower)
	}
	if result.Result.ID != "id:file" {
		t.Fatalf("id = %q, want id:file", result.Result.ID)
	}
	if result.Result.Rev != "deleted-rev" {
		t.Fatalf("rev = %q, want deleted-rev", result.Result.Rev)
	}
	if result.Result.Size == nil || *result.Result.Size != 456 {
		t.Fatalf("size = %v, want 456", result.Result.Size)
	}
}

func TestRmJSONFolderOmitsFileFields(t *testing.T) {
	cmd, stdout := testRmCmd(t)
	setRmOutputJSON(t, cmd)
	setRmFlag(t, cmd, "recursive")
	folder := rmFolderMetadata("/Folder")

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return folder, nil
		},
		deleteV2Fn: func(arg *files.DeleteArg) (*files.DeleteResult, error) {
			return files.NewDeleteResult(folder), nil
		},
	}
	stubFilesClient(t, mock)

	if err := rm(cmd, []string{"/Folder"}); err != nil {
		t.Fatalf("rm error: %v", err)
	}

	output := stdout.String()
	if strings.Contains(output, `"rev"`) || strings.Contains(output, `"size"`) {
		t.Fatalf("folder JSON output = %s, want no file-only fields", output)
	}
	got := decodeRemoveOutputString(t, output)
	if len(got.Results) != 1 {
		t.Fatalf("results len = %d, want 1", len(got.Results))
	}
	result := got.Results[0]
	if result.Result.Type != "folder" {
		t.Fatalf("result type = %q, want folder", result.Result.Type)
	}
	if !result.Input.Recursive {
		t.Fatalf("recursive = false, want true")
	}
}

func TestRmJSONPermanentUsesValidatedMetadata(t *testing.T) {
	cmd, stdout := testRmCmd(t)
	setRmOutputJSON(t, cmd)
	setRmFlag(t, cmd, "permanent")
	file := rmFileMetadata("/File.txt")
	file.Rev = "validated-rev"
	var permanentlyDeleted []string

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return file, nil
		},
		deleteV2Fn: func(arg *files.DeleteArg) (*files.DeleteResult, error) {
			t.Fatalf("DeleteV2 called for permanent delete: %v", arg)
			return nil, nil
		},
		permanentlyDeleteFn: func(arg *files.DeleteArg) error {
			permanentlyDeleted = append(permanentlyDeleted, arg.Path)
			return nil
		},
	}
	stubFilesClient(t, mock)

	if err := rm(cmd, []string{"/File.txt"}); err != nil {
		t.Fatalf("rm error: %v", err)
	}
	if len(permanentlyDeleted) != 1 || permanentlyDeleted[0] != "/File.txt" {
		t.Fatalf("permanentlyDeleted = %v, want [/File.txt]", permanentlyDeleted)
	}
	got := decodeRemoveOutput(t, stdout)
	if len(got.Results) != 1 {
		t.Fatalf("results len = %d, want 1", len(got.Results))
	}
	result := got.Results[0]
	if !result.Input.Permanent {
		t.Fatalf("permanent = false, want true")
	}
	if result.Result.Rev != "validated-rev" {
		t.Fatalf("rev = %q, want validated-rev", result.Result.Rev)
	}
}

func TestRmJSONMultipleTargets(t *testing.T) {
	cmd, stdout := testRmCmd(t)
	setRmOutputJSON(t, cmd)

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return rmFileMetadata(arg.Path), nil
		},
		deleteV2Fn: func(arg *files.DeleteArg) (*files.DeleteResult, error) {
			return files.NewDeleteResult(rmFileMetadata(arg.Path)), nil
		},
	}
	stubFilesClient(t, mock)

	if err := rm(cmd, []string{"/one.txt", "/two.txt"}); err != nil {
		t.Fatalf("rm error: %v", err)
	}

	got := decodeRemoveOutput(t, stdout)
	if len(got.Results) != 2 {
		t.Fatalf("results len = %d, want 2", len(got.Results))
	}
	if got.Results[0].Input.Path != "/one.txt" || got.Results[1].Input.Path != "/two.txt" {
		t.Fatalf("result paths = %q, %q; want /one.txt, /two.txt", got.Results[0].Input.Path, got.Results[1].Input.Path)
	}
}

func TestRmJSONVerboseDoesNotPrintText(t *testing.T) {
	cmd, stdout := testRmCmd(t)
	setRmOutputJSON(t, cmd)
	setRmFlag(t, cmd, "verbose")
	file := rmFileMetadata("/File.txt")

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return file, nil
		},
		deleteV2Fn: func(arg *files.DeleteArg) (*files.DeleteResult, error) {
			return files.NewDeleteResult(file), nil
		},
	}
	stubFilesClient(t, mock)

	if err := rm(cmd, []string{"/File.txt"}); err != nil {
		t.Fatalf("rm error: %v", err)
	}
	if strings.Contains(stdout.String(), "Deleted ") {
		t.Fatalf("stdout = %q, want JSON only", stdout.String())
	}
	got := decodeRemoveOutput(t, stdout)
	if len(got.Results) != 1 {
		t.Fatalf("results len = %d, want 1", len(got.Results))
	}
}

func TestRmJSONErrorWritesNoOutput(t *testing.T) {
	cmd, stdout := testRmCmd(t)
	setRmOutputJSON(t, cmd)

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, errors.New("metadata failed")
		},
	}
	stubFilesClient(t, mock)

	if err := rm(cmd, []string{"/File.txt"}); err == nil {
		t.Fatal("expected rm error")
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty output on error", got)
	}
}

func TestRmCommandSupportsStructuredOutput(t *testing.T) {
	if !commandSupportsStructuredOutput(rmCmd) {
		t.Fatal("rm command should support structured output")
	}
}
