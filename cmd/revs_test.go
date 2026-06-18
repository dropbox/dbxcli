package cmd

import (
	"bytes"
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

	stubFilesClient(t, &mockFilesClient{
		listRevisionsFn: func(arg *files.ListRevisionsArg) (*files.ListRevisionsResult, error) {
			gotPath = arg.Path
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
	if got, want := stdout.String(), "rev-c\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func testRevsCmd() (*cobra.Command, *bytes.Buffer) {
	var stdout bytes.Buffer
	cmd := &cobra.Command{Use: "revs"}
	cmd.SetOut(&stdout)
	cmd.Flags().BoolP("long", "l", false, "")
	cmd.Flags().String("time", "server", "")
	cmd.Flags().String("time-format", "", "")
	return cmd, &stdout
}
