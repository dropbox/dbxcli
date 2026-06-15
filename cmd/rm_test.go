package cmd

import (
	"bytes"
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
		},
		Id: "id:folder",
	}
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
