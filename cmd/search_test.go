package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

func TestSearchArgValidation(t *testing.T) {
	err := search(searchCmd, []string{})
	if err == nil {
		t.Error("expected error for no args")
	}
}

func TestSearchPathScopeValidation(t *testing.T) {
	err := search(searchCmd, []string{"query", "no-slash"})
	if err == nil {
		t.Error("expected error for path-scope without leading slash")
	}
}

func TestRenderSearchResultsSeparatesMatchesWithNewlines(t *testing.T) {
	entries := []files.IsMetadata{
		&files.FileMetadata{
			Metadata: files.Metadata{PathDisplay: "/first.txt"},
		},
		&files.FolderMetadata{
			Metadata: files.Metadata{PathDisplay: "/second"},
		},
	}

	var out bytes.Buffer
	if err := renderSearchResults(&out, entries, listOptions{long: false}); err != nil {
		t.Fatalf("renderSearchResults returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected one output line per match, got %d lines in %q", len(lines), out.String())
	}

	if got, want := strings.TrimSpace(lines[0]), "/first.txt"; got != want {
		t.Errorf("first rendered match = %q, want %q", got, want)
	}
	if got, want := strings.TrimSpace(lines[1]), "/second"; got != want {
		t.Errorf("second rendered match = %q, want %q", got, want)
	}
}

func TestRenderSearchResultsLongModeIncludesHeader(t *testing.T) {
	entries := []files.IsMetadata{
		&files.FileMetadata{
			Metadata: files.Metadata{PathDisplay: "/first.txt"},
			Rev:      "abc123",
			Size:     42,
		},
	}

	var out bytes.Buffer
	if err := renderSearchResults(&out, entries, listOptions{long: true}); err != nil {
		t.Fatalf("renderSearchResults returned error: %v", err)
	}

	got := out.String()
	for _, want := range []string{"Revision", "Size", "Last modified", "Path", "abc123", "/first.txt"} {
		if !strings.Contains(got, want) {
			t.Errorf("output = %q, want to contain %q", got, want)
		}
	}
}

func TestSearchUsesSearchV2AndCommandOutput(t *testing.T) {
	cmd, stdout := testSearchCmd()
	var firstArg *files.SearchV2Arg
	var continueCursor string

	mock := &mockFilesClient{
		searchV2Fn: func(arg *files.SearchV2Arg) (*files.SearchV2Result, error) {
			firstArg = arg
			res := files.NewSearchV2Result([]*files.SearchMatchV2{
				searchMatch(&files.FileMetadata{
					Metadata: files.Metadata{PathDisplay: "/docs/first.txt"},
				}),
			}, true)
			res.Cursor = "cursor-1"
			return res, nil
		},
		searchContinueV2Fn: func(arg *files.SearchV2ContinueArg) (*files.SearchV2Result, error) {
			continueCursor = arg.Cursor
			return files.NewSearchV2Result([]*files.SearchMatchV2{
				searchMatch(&files.FolderMetadata{
					Metadata: files.Metadata{PathDisplay: "/docs/second"},
				}),
			}, false), nil
		},
	}
	stubFilesClient(t, mock)

	if err := search(cmd, []string{"needle", "/docs"}); err != nil {
		t.Fatalf("search error: %v", err)
	}

	if firstArg == nil {
		t.Fatal("SearchV2 was not called")
	}
	if firstArg.Query != "needle" {
		t.Errorf("query = %q, want %q", firstArg.Query, "needle")
	}
	if firstArg.Options == nil || firstArg.Options.Path != "/docs" {
		t.Fatalf("options path = %#v, want /docs", firstArg.Options)
	}
	if continueCursor != "cursor-1" {
		t.Errorf("continue cursor = %q, want cursor-1", continueCursor)
	}

	got := stdout.String()
	for _, want := range []string{"/docs/first.txt", "/docs/second"} {
		if !strings.Contains(got, want) {
			t.Errorf("stdout = %q, want to contain %q", got, want)
		}
	}
}

func testSearchCmd() (*cobra.Command, *bytes.Buffer) {
	var stdout bytes.Buffer
	cmd := &cobra.Command{Use: "search"}
	cmd.SetOut(&stdout)
	cmd.Flags().BoolP("long", "l", false, "")
	cmd.Flags().String("sort", "", "")
	cmd.Flags().BoolP("reverse", "r", false, "")
	cmd.Flags().String("time", "server", "")
	cmd.Flags().String("time-format", "", "")
	return cmd, &stdout
}

func searchMatch(metadata files.IsMetadata) *files.SearchMatchV2 {
	return files.NewSearchMatchV2(&files.MetadataV2{
		Tagged:   dropbox.Tagged{Tag: files.MetadataV2Metadata},
		Metadata: metadata,
	})
}
