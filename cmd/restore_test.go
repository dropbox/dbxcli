package cmd

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

func TestRestoreArgValidation(t *testing.T) {
	err := restore(restoreCmd, []string{})
	if err == nil {
		t.Fatal("expected error for missing args")
	}
	for _, want := range []string{"target-path", "revision"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, want mention of %q", err.Error(), want)
		}
	}
}

func TestRestoreHelpClarifiesTargetPath(t *testing.T) {
	if !strings.Contains(restoreCmd.Use, "<target-path> <revision>") {
		t.Fatalf("Use = %q, want target-path and revision", restoreCmd.Use)
	}

	for _, want := range []string{
		"where the restored file is saved",
		"dbxcli revs <target-path>",
	} {
		if !strings.Contains(restoreCmd.Long, want) {
			t.Fatalf("Long = %q, want mention of %q", restoreCmd.Long, want)
		}
	}
}

func TestRestoreQuietByDefault(t *testing.T) {
	cmd, stdout := testRestoreCmd()
	var restoreArg *files.RestoreArg
	serverModified := time.Date(2026, 6, 17, 12, 30, 0, 0, time.UTC)
	mock := &mockFilesClient{
		restoreFn: func(arg *files.RestoreArg) (*files.FileMetadata, error) {
			restoreArg = arg
			return &files.FileMetadata{
				Metadata:       files.Metadata{PathDisplay: "/Reports/old.pdf"},
				Rev:            "current-rev",
				ServerModified: serverModified,
			}, nil
		},
	}
	stubFilesClient(t, mock)

	if err := restore(cmd, []string{"/Reports/old.pdf", "target-rev"}); err != nil {
		t.Fatalf("restore error: %v", err)
	}
	if restoreArg == nil {
		t.Fatal("Restore was not called")
	}
	if restoreArg.Path != "/Reports/old.pdf" || restoreArg.Rev != "target-rev" {
		t.Fatalf("restore arg = %#v, want path /Reports/old.pdf and rev target-rev", restoreArg)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want quiet success", got)
	}
}

func TestRestoreVerbosePrintsRevisionAndServerModifiedTime(t *testing.T) {
	cmd, stdout := testRestoreCmd()
	if err := cmd.Flags().Set("verbose", "true"); err != nil {
		t.Fatalf("set verbose: %v", err)
	}

	serverModified := time.Date(2026, 6, 17, 12, 30, 0, 0, time.UTC)
	mock := &mockFilesClient{
		restoreFn: func(arg *files.RestoreArg) (*files.FileMetadata, error) {
			return &files.FileMetadata{
				Metadata:       files.Metadata{PathDisplay: "/Reports/old.pdf"},
				Rev:            "current-rev",
				ServerModified: serverModified,
			}, nil
		},
	}
	stubFilesClient(t, mock)

	if err := restore(cmd, []string{"/Reports/old.pdf", "target-rev"}); err != nil {
		t.Fatalf("restore error: %v", err)
	}

	want := "Restored /Reports/old.pdf to revision target-rev (current revision current-rev, server modified 2026-06-17T12:30:00Z)\n"
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestNewRestoreResultKeepsInputAndMetadata(t *testing.T) {
	clientModified := time.Date(2026, 6, 16, 10, 0, 0, 0, time.UTC)
	serverModified := time.Date(2026, 6, 17, 12, 30, 0, 0, time.UTC)
	result := newRestoreResult("/Reports/old.pdf", "target-rev", &files.FileMetadata{
		Metadata: files.Metadata{
			PathDisplay: "/Reports/old.pdf",
		},
		Id:             "id:abc",
		Rev:            "current-rev",
		Size:           123,
		ClientModified: clientModified,
		ServerModified: serverModified,
	})

	if result.Input.Path != "/Reports/old.pdf" || result.Input.Revision != "target-rev" {
		t.Fatalf("input = %#v, want path and target revision", result.Input)
	}
	if result.Result.Type != "file" || result.Result.PathDisplay != "/Reports/old.pdf" {
		t.Fatalf("metadata = %#v, want file path metadata", result.Result)
	}
	if result.Result.ID != "id:abc" || result.Result.Rev != "current-rev" || result.Result.Size != 123 {
		t.Fatalf("metadata = %#v, want id, current rev, and size", result.Result)
	}
	if !result.Result.ClientModified.Equal(clientModified) || !result.Result.ServerModified.Equal(serverModified) {
		t.Fatalf("metadata times = %#v, want client and server modified times", result.Result)
	}
}

func testRestoreCmd() (*cobra.Command, *bytes.Buffer) {
	var stdout bytes.Buffer
	cmd := &cobra.Command{Use: "restore"}
	cmd.SetOut(&stdout)
	cmd.Flags().BoolP("verbose", "v", false, "")
	return cmd, &stdout
}
