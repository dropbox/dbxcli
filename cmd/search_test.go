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

func TestSearchJSONOutputsInputAndEntries(t *testing.T) {
	cmd, stdout := testSearchCmd()
	setSearchOutputJSON(t, cmd)
	setSearchFlag(t, cmd, "long", "true")
	setSearchFlag(t, cmd, "sort", "name")
	setSearchFlag(t, cmd, "reverse", "true")
	setSearchFlag(t, cmd, "time", "client")
	setSearchFlag(t, cmd, "time-format", "rfc3339")
	var firstArg *files.SearchV2Arg

	mock := &mockFilesClient{
		searchV2Fn: func(arg *files.SearchV2Arg) (*files.SearchV2Result, error) {
			firstArg = arg
			return files.NewSearchV2Result([]*files.SearchMatchV2{
				searchMatch(&files.FileMetadata{
					Metadata: files.Metadata{
						PathDisplay: "/docs/report.txt",
						PathLower:   "/docs/report.txt",
					},
					Id:   "id:file",
					Rev:  "rev-file",
					Size: 42,
				}),
				searchMatch(&files.FolderMetadata{
					Metadata: files.Metadata{
						PathDisplay: "/docs/archive",
						PathLower:   "/docs/archive",
					},
					Id: "id:folder",
				}),
			}, false), nil
		},
	}
	stubFilesClient(t, mock)

	if err := search(cmd, []string{"report", "/docs"}); err != nil {
		t.Fatalf("search error: %v", err)
	}
	if firstArg == nil {
		t.Fatal("SearchV2 was not called")
	}
	if firstArg.Query != "report" {
		t.Fatalf("query = %q, want report", firstArg.Query)
	}
	if firstArg.Options == nil || firstArg.Options.Path != "/docs" {
		t.Fatalf("options path = %#v, want /docs", firstArg.Options)
	}

	got := decodeSearchOutput(t, stdout)
	if got.Input.Query != "report" || got.Input.Path != "/docs" {
		t.Fatalf("input = %#v, want query report path /docs", got.Input)
	}
	if !got.Input.Long || got.Input.Sort != "name" || !got.Input.Reverse || got.Input.Time != "client" || got.Input.TimeFormat != "rfc3339" {
		t.Fatalf("input options = %#v, want long/sort/reverse/time/time-format", got.Input)
	}
	if len(got.Entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(got.Entries))
	}
	if got.Entries[0].Type != "file" || got.Entries[0].Rev != "rev-file" || got.Entries[0].Size == nil || *got.Entries[0].Size != 42 {
		t.Fatalf("first entry = %#v, want file metadata", got.Entries[0])
	}
	if got.Entries[1].Type != "folder" || got.Entries[1].ID != "id:folder" {
		t.Fatalf("second entry = %#v, want folder metadata", got.Entries[1])
	}
}

func TestSearchJSONOmitsPathWithoutScope(t *testing.T) {
	cmd, stdout := testSearchCmd()
	setSearchOutputJSON(t, cmd)

	mock := &mockFilesClient{
		searchV2Fn: func(arg *files.SearchV2Arg) (*files.SearchV2Result, error) {
			if arg.Options != nil && arg.Options.Path != "" {
				t.Fatalf("options path = %q, want empty", arg.Options.Path)
			}
			return files.NewSearchV2Result(nil, false), nil
		},
	}
	stubFilesClient(t, mock)

	if err := search(cmd, []string{"report"}); err != nil {
		t.Fatalf("search error: %v", err)
	}
	output := append([]byte(nil), stdout.Bytes()...)
	got := decodeSearchOutput(t, stdout)
	if got.Input.Query != "report" || got.Input.Path != "" {
		t.Fatalf("input = %#v, want query report and empty path", got.Input)
	}
	var raw map[string]any
	if err := json.Unmarshal(output, &raw); err != nil {
		t.Fatalf("decode raw JSON output: %v\noutput: %s", err, string(output))
	}
	input, ok := raw["input"].(map[string]any)
	if !ok {
		t.Fatalf("raw input = %#v, want object", raw["input"])
	}
	if _, ok := input["path"]; ok {
		t.Fatalf("input path key is present in %s, want omitted", string(output))
	}
}

func TestSearchJSONErrorWritesNoOutput(t *testing.T) {
	cmd, stdout := testSearchCmd()
	setSearchOutputJSON(t, cmd)

	mock := &mockFilesClient{
		searchV2Fn: func(arg *files.SearchV2Arg) (*files.SearchV2Result, error) {
			return nil, fmt.Errorf("search failed")
		},
	}
	stubFilesClient(t, mock)

	if err := search(cmd, []string{"report"}); err == nil {
		t.Fatal("expected search error")
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty output on error", got)
	}
}

func TestSearchCommandSupportsStructuredOutput(t *testing.T) {
	if !commandSupportsStructuredOutput(searchCmd) {
		t.Fatal("search command should support structured output")
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
	cmd.Flags().String(outputFlag, "text", "")
	return cmd, &stdout
}

func setSearchOutputJSON(t *testing.T, cmd *cobra.Command) {
	t.Helper()
	setSearchFlag(t, cmd, outputFlag, "json")
}

func setSearchFlag(t *testing.T, cmd *cobra.Command, name, value string) {
	t.Helper()
	if err := cmd.Flags().Set(name, value); err != nil {
		t.Fatalf("set %s: %v", name, err)
	}
}

func decodeSearchOutput(t *testing.T, out *bytes.Buffer) searchOutput {
	t.Helper()

	var got searchOutput
	if err := json.NewDecoder(out).Decode(&got); err != nil {
		t.Fatalf("decode JSON output: %v\noutput: %s", err, out.String())
	}
	return got
}

func searchMatch(metadata files.IsMetadata) *files.SearchMatchV2 {
	return files.NewSearchMatchV2(&files.MetadataV2{
		Tagged:   dropbox.Tagged{Tag: files.MetadataV2Metadata},
		Metadata: metadata,
	})
}
