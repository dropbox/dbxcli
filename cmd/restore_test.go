package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
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
				Metadata:       files.Metadata{PathDisplay: "/Reports/old.pdf", PathLower: "/reports/old.pdf"},
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
				Metadata:       files.Metadata{PathDisplay: "/Reports/old.pdf", PathLower: "/reports/old.pdf"},
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
	result, err := newRestoreResult("/Reports/old.pdf", "target-rev", &files.FileMetadata{
		Metadata: files.Metadata{
			PathDisplay: "/Reports/old.pdf",
			PathLower:   "/reports/old.pdf",
		},
		Id:             "id:abc",
		Rev:            "current-rev",
		Size:           123,
		ClientModified: clientModified,
		ServerModified: serverModified,
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.Input.Path != "/Reports/old.pdf" || result.Input.Revision != "target-rev" {
		t.Fatalf("input = %#v, want path and target revision", result.Input)
	}
	if result.Status != restoreStatusRestored || result.Kind != restoreKindFile {
		t.Fatalf("status/kind = %s/%s, want restored/file", result.Status, result.Kind)
	}
	if result.Result.Type != "file" || result.Result.PathDisplay != "/Reports/old.pdf" {
		t.Fatalf("metadata = %#v, want file path metadata", result.Result)
	}
	if result.Result.ID != "id:abc" || result.Result.Rev != "current-rev" ||
		result.Result.Size == nil || *result.Result.Size != 123 {
		t.Fatalf("metadata = %#v, want id, current rev, and size", result.Result)
	}
	if result.Result.ServerModified == nil || *result.Result.ServerModified != "2026-06-17T12:30:00Z" {
		t.Fatalf("server modified = %v, want 2026-06-17T12:30:00Z", result.Result.ServerModified)
	}
	if result.Result.ClientModified == nil || *result.Result.ClientModified != "2026-06-16T10:00:00Z" {
		t.Fatalf("client modified = %v, want 2026-06-16T10:00:00Z", result.Result.ClientModified)
	}
}

func TestRestoreJSONOutputsInputAndMetadata(t *testing.T) {
	cmd, stdout := testRestoreCmd()
	setRestoreOutputJSON(t, cmd)
	var restoreArg *files.RestoreArg
	clientModified := time.Date(2026, 6, 16, 10, 0, 0, 0, time.UTC)
	serverModified := time.Date(2026, 6, 17, 12, 30, 0, 0, time.UTC)
	mock := &mockFilesClient{
		restoreFn: func(arg *files.RestoreArg) (*files.FileMetadata, error) {
			restoreArg = arg
			return &files.FileMetadata{
				Metadata: files.Metadata{
					PathDisplay: "/Reports/old.pdf",
					PathLower:   "/reports/old.pdf",
				},
				Id:             "id:abc",
				Rev:            "current-rev",
				Size:           123,
				ClientModified: clientModified,
				ServerModified: serverModified,
			}, nil
		},
	}
	stubFilesClient(t, mock)

	if err := restore(cmd, []string{"/Reports/old.pdf", "target-rev"}); err != nil {
		t.Fatalf("restore error: %v", err)
	}
	if restoreArg == nil || restoreArg.Path != "/Reports/old.pdf" || restoreArg.Rev != "target-rev" {
		t.Fatalf("restore arg = %#v, want path /Reports/old.pdf and rev target-rev", restoreArg)
	}

	got := decodeRestoreOutput(t, stdout)
	if got.Input.Path != "/Reports/old.pdf" || got.Input.Revision != "target-rev" {
		t.Fatalf("input = %#v, want path and target revision", got.Input)
	}
	result := got.Results[0]
	if result.Status != restoreStatusRestored || result.Kind != restoreKindFile {
		t.Fatalf("status/kind = %s/%s, want restored/file", result.Status, result.Kind)
	}
	if result.Input.Path != "/Reports/old.pdf" || result.Input.Revision != "target-rev" {
		t.Fatalf("result input = %#v, want path and target revision", result.Input)
	}
	if result.Result.Type != "file" || result.Result.PathDisplay != "/Reports/old.pdf" || result.Result.PathLower != "/reports/old.pdf" {
		t.Fatalf("metadata = %#v, want file path metadata", result.Result)
	}
	if result.Result.ID != "id:abc" || result.Result.Rev != "current-rev" {
		t.Fatalf("metadata = %#v, want returned id and current revision", result.Result)
	}
	if result.Result.Size == nil || *result.Result.Size != 123 {
		t.Fatalf("size = %v, want 123", result.Result.Size)
	}
	if result.Result.ServerModified == nil || *result.Result.ServerModified != "2026-06-17T12:30:00Z" {
		t.Fatalf("server modified = %v, want 2026-06-17T12:30:00Z", result.Result.ServerModified)
	}
	if result.Result.ClientModified == nil || *result.Result.ClientModified != "2026-06-16T10:00:00Z" {
		t.Fatalf("client modified = %v, want 2026-06-16T10:00:00Z", result.Result.ClientModified)
	}
}

func TestRestoreJSONUsesInputPathWhenMetadataPathDisplayMissing(t *testing.T) {
	cmd, stdout := testRestoreCmd()
	setRestoreOutputJSON(t, cmd)
	mock := &mockFilesClient{
		restoreFn: func(arg *files.RestoreArg) (*files.FileMetadata, error) {
			return &files.FileMetadata{Rev: "current-rev"}, nil
		},
	}
	stubFilesClient(t, mock)

	if err := restore(cmd, []string{"/Reports/old.pdf", "target-rev"}); err != nil {
		t.Fatalf("restore error: %v", err)
	}

	got := decodeRestoreOutput(t, stdout)
	if got.Results[0].Result.PathDisplay != "/Reports/old.pdf" {
		t.Fatalf("path_display = %q, want fallback input path", got.Results[0].Result.PathDisplay)
	}
}

func TestRestoreJSONVerboseDoesNotPrintText(t *testing.T) {
	cmd, stdout := testRestoreCmd()
	setRestoreOutputJSON(t, cmd)
	if err := cmd.Flags().Set("verbose", "true"); err != nil {
		t.Fatalf("set verbose: %v", err)
	}

	mock := &mockFilesClient{
		restoreFn: func(arg *files.RestoreArg) (*files.FileMetadata, error) {
			return &files.FileMetadata{
				Metadata:       files.Metadata{PathDisplay: "/Reports/old.pdf"},
				Rev:            "current-rev",
				ServerModified: time.Date(2026, 6, 17, 12, 30, 0, 0, time.UTC),
			}, nil
		},
	}
	stubFilesClient(t, mock)

	if err := restore(cmd, []string{"/Reports/old.pdf", "target-rev"}); err != nil {
		t.Fatalf("restore error: %v", err)
	}
	if strings.Contains(stdout.String(), "Restored ") {
		t.Fatalf("stdout = %q, want JSON only", stdout.String())
	}
	got := decodeRestoreOutput(t, stdout)
	if got.Results[0].Result.Rev != "current-rev" {
		t.Fatalf("rev = %q, want current-rev", got.Results[0].Result.Rev)
	}
}

func TestRestoreJSONErrorWritesNoOutput(t *testing.T) {
	cmd, stdout := testRestoreCmd()
	setRestoreOutputJSON(t, cmd)
	mock := &mockFilesClient{
		restoreFn: func(arg *files.RestoreArg) (*files.FileMetadata, error) {
			return nil, errors.New("restore failed")
		},
	}
	stubFilesClient(t, mock)

	err := restore(cmd, []string{"/Reports/old.pdf", "target-rev"})
	if err == nil {
		t.Fatal("expected restore error")
	}
	details := jsonErrorDetails(err)
	if details["operation"] != "restore" || details["path"] != "/Reports/old.pdf" || details["revision"] != "target-rev" {
		t.Fatalf("details = %#v, want restore operation, path, and revision", details)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty output on error", got)
	}
}

func TestRestoreCommandSupportsStructuredOutput(t *testing.T) {
	if !commandSupportsStructuredOutput(restoreCmd) {
		t.Fatal("restore command should support structured output")
	}
}

func testRestoreCmd() (*cobra.Command, *bytes.Buffer) {
	var stdout bytes.Buffer
	cmd := &cobra.Command{Use: "restore"}
	cmd.SetOut(&stdout)
	cmd.Flags().BoolP("verbose", "v", false, "")
	cmd.Flags().String(outputFlag, "text", "")
	return cmd, &stdout
}

func setRestoreOutputJSON(t *testing.T, cmd *cobra.Command) {
	t.Helper()

	if err := cmd.Flags().Set(outputFlag, "json"); err != nil {
		t.Fatal(err)
	}
}

type restoreOutput struct {
	Input    restoreInput    `json:"input"`
	Results  []restoreResult `json:"results"`
	Warnings []jsonWarning   `json:"warnings"`
}

func decodeRestoreOutput(t *testing.T, stdout *bytes.Buffer) restoreOutput {
	t.Helper()

	var got restoreOutput
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
