package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

func TestRenderRevisionResultsPrintsRevisionIDs(t *testing.T) {
	entries := []*files.FileMetadata{
		{Rev: "rev-a"},
		{Rev: "rev-b"},
	}

	var out bytes.Buffer
	if err := renderRevisionResults(&out, entries, listOptions{}); err != nil {
		t.Fatalf("renderRevisionResults returned error: %v", err)
	}

	if got, want := out.String(), "rev-a\nrev-b\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestRenderRevisionResultsLongModeUsesTimeOptions(t *testing.T) {
	serverModified := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)
	clientModified := time.Date(2026, 5, 1, 10, 30, 0, 0, time.UTC)
	entries := []*files.FileMetadata{
		{
			Metadata:       files.Metadata{PathDisplay: "/report.pdf"},
			Rev:            "rev-a",
			Size:           4096,
			ServerModified: serverModified,
			ClientModified: clientModified,
		},
	}

	var out bytes.Buffer
	err := renderRevisionResults(&out, entries, listOptions{
		long:       true,
		timeField:  "client",
		timeFormat: "rfc3339",
	})
	if err != nil {
		t.Fatalf("renderRevisionResults returned error: %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"Revision",
		"Size",
		"Last modified",
		"Path",
		"rev-a",
		"4.0 KiB",
		"2026-05-01T10:30:00Z",
		"/report.pdf",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output = %q, want to contain %q", got, want)
		}
	}
	if strings.Contains(got, "2026-05-01T09:00:00Z") {
		t.Errorf("output = %q, should use client-modified time", got)
	}
}

func TestRevsUsesListRevisionsAndCommandOutput(t *testing.T) {
	cmd, stdout := testRevsCmd()
	var gotPath string
	var gotLimit uint64

	stubFilesClient(t, &mockFilesClient{
		listRevisionsFn: func(arg *files.ListRevisionsArg) (*files.ListRevisionsResult, error) {
			gotPath = arg.Path
			gotLimit = arg.Limit
			return files.NewListRevisionsResult(false, []*files.FileMetadata{
				{Rev: "rev-c"},
			}), nil
		},
	})

	if err := revs(cmd, []string{"/report.pdf"}); err != nil {
		t.Fatalf("revs returned error: %v", err)
	}

	if gotPath != "/report.pdf" {
		t.Fatalf("ListRevisions path = %q, want %q", gotPath, "/report.pdf")
	}
	if gotLimit != 10 {
		t.Fatalf("ListRevisions limit = %d, want SDK default 10", gotLimit)
	}
	if got, want := stdout.String(), "rev-c\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRevsJSONOutputsInputAndResults(t *testing.T) {
	cmd, stdout := testRevsCmd()
	setRevsOutputJSON(t, cmd)
	setRevsFlag(t, cmd, "long", "true")
	setRevsFlag(t, cmd, "limit", "25")
	setRevsFlag(t, cmd, "time", "client")
	setRevsFlag(t, cmd, "time-format", "rfc3339")
	var gotPath string
	var gotLimit uint64
	clientModified := time.Date(2026, 6, 22, 10, 0, 0, 0, time.UTC)

	stubFilesClient(t, &mockFilesClient{
		listRevisionsFn: func(arg *files.ListRevisionsArg) (*files.ListRevisionsResult, error) {
			gotPath = arg.Path
			gotLimit = arg.Limit
			return files.NewListRevisionsResult(false, []*files.FileMetadata{
				{
					Metadata: files.Metadata{
						PathDisplay: "/report.pdf",
						PathLower:   "/report.pdf",
					},
					Id:             "id:file",
					Rev:            "rev-a",
					Size:           42,
					ClientModified: clientModified,
				},
			}), nil
		},
	})

	if err := revs(cmd, []string{"/report.pdf"}); err != nil {
		t.Fatalf("revs returned error: %v", err)
	}
	if gotPath != "/report.pdf" {
		t.Fatalf("ListRevisions path = %q, want /report.pdf", gotPath)
	}
	if gotLimit != 25 {
		t.Fatalf("ListRevisions limit = %d, want 25", gotLimit)
	}

	got := decodeRevsOutput(t, stdout)
	if got.Input.Path != "/report.pdf" || got.Input.Limit != 25 || !got.Input.Long || got.Input.Time != "client" || got.Input.TimeFormat != "rfc3339" {
		t.Fatalf("input = %#v, want path/long/time/time_format", got.Input)
	}
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Results))
	}
	if got.Results[0].Status != revsJSONStatusRevision || got.Results[0].Kind != "file" {
		t.Fatalf("result = %#v, want revision file", got.Results[0])
	}
	entry := got.Results[0].Result
	if entry.Type != "file" || entry.PathDisplay != "/report.pdf" || entry.Rev != "rev-a" || entry.Size == nil || *entry.Size != 42 {
		t.Fatalf("entry = %#v, want file revision metadata", entry)
	}
	if entry.ClientModified == nil || *entry.ClientModified != "2026-06-22T10:00:00Z" {
		t.Fatalf("client_modified = %#v, want RFC3339 timestamp", entry.ClientModified)
	}
	if strings.Contains(stdout.String(), `"entries"`) {
		t.Fatalf("JSON output = %s, want operation results and no entries key", stdout.String())
	}
}

func TestRevsJSONErrorWritesNoOutput(t *testing.T) {
	cmd, stdout := testRevsCmd()
	setRevsOutputJSON(t, cmd)

	stubFilesClient(t, &mockFilesClient{
		listRevisionsFn: func(arg *files.ListRevisionsArg) (*files.ListRevisionsResult, error) {
			return nil, fmt.Errorf("revs failed")
		},
	})

	if err := revs(cmd, []string{"/report.pdf"}); err == nil {
		t.Fatal("expected revs error")
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty output on error", got)
	}
}

func TestRevsRejectsInvalidListOptions(t *testing.T) {
	tests := []struct {
		name  string
		flag  string
		value string
	}{
		{name: "time", flag: "time", value: "created"},
		{name: "time-format", flag: "time-format", value: "unix"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, stdout := testRevsCmd()
			setRevsFlag(t, cmd, tt.flag, tt.value)
			stubFilesClient(t, &mockFilesClient{
				listRevisionsFn: func(arg *files.ListRevisionsArg) (*files.ListRevisionsResult, error) {
					t.Fatalf("ListRevisions called for invalid --%s", tt.flag)
					return nil, nil
				},
			})

			err := revs(cmd, []string{"/report.pdf"})
			if err == nil {
				t.Fatalf("expected invalid --%s error", tt.flag)
			}
			if !strings.Contains(err.Error(), tt.flag) || !strings.Contains(err.Error(), tt.value) {
				t.Fatalf("error = %v, want flag and value", err)
			}
			if got := stdout.String(); got != "" {
				t.Fatalf("stdout = %q, want empty output on validation error", got)
			}
		})
	}
}

func TestRevsCommandSupportsStructuredOutput(t *testing.T) {
	if !commandSupportsStructuredOutput(revsCmd) {
		t.Fatal("revs command should support structured output")
	}
}

func testRevsCmd() (*cobra.Command, *bytes.Buffer) {
	var stdout bytes.Buffer
	cmd := &cobra.Command{Use: "revs"}
	cmd.SetOut(&stdout)
	cmd.Flags().BoolP("long", "l", false, "")
	cmd.Flags().Uint64("limit", 0, "")
	cmd.Flags().String("time", "server", "")
	cmd.Flags().String("time-format", "", "")
	cmd.Flags().String(outputFlag, "text", "")
	return cmd, &stdout
}

func setRevsOutputJSON(t *testing.T, cmd *cobra.Command) {
	t.Helper()
	setRevsFlag(t, cmd, outputFlag, "json")
}

func setRevsFlag(t *testing.T, cmd *cobra.Command, name, value string) {
	t.Helper()
	if err := cmd.Flags().Set(name, value); err != nil {
		t.Fatalf("set %s: %v", name, err)
	}
}

func decodeRevsOutput(t *testing.T, out *bytes.Buffer) metadataOperationOutputForTest[revsInput] {
	t.Helper()

	return decodeMetadataOperationOutput[revsInput](t, out)
}
