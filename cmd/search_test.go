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

func TestSearchRejectsExtraArgs(t *testing.T) {
	err := search(searchCmd, []string{"query", "/docs", "extra"})
	if err == nil || !strings.Contains(err.Error(), "path-scope") {
		t.Fatalf("error = %v, want extra path-scope error", err)
	}
}

func TestSearchOrderByValidation(t *testing.T) {
	cmd, _ := testSearchCmd()
	setSearchFlag(t, cmd, "order-by", "name")

	err := search(cmd, []string{"query"})
	if err == nil || !strings.Contains(err.Error(), "order-by") {
		t.Fatalf("error = %v, want order-by validation error", err)
	}
}

func TestSearchRejectsInvalidListOptions(t *testing.T) {
	tests := []struct {
		name  string
		flag  string
		value string
	}{
		{name: "sort", flag: "sort", value: "date"},
		{name: "time", flag: "time", value: "created"},
		{name: "time-format", flag: "time-format", value: "unix"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, stdout := testSearchCmd()
			setSearchFlag(t, cmd, tt.flag, tt.value)
			stubFilesClient(t, &mockFilesClient{
				searchV2Fn: func(arg *files.SearchV2Arg) (*files.SearchV2Result, error) {
					t.Fatalf("SearchV2 called for invalid --%s", tt.flag)
					return nil, nil
				},
			})

			err := search(cmd, []string{"query"})
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
	if !firstArg.Options.FilenameOnly {
		t.Fatal("filename_only = false, want true by default")
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

func TestSearchJSONOutputsInputAndResults(t *testing.T) {
	cmd, stdout := testSearchCmd()
	setSearchOutputJSON(t, cmd)
	setSearchFlag(t, cmd, "long", "true")
	setSearchFlag(t, cmd, "sort", "name")
	setSearchFlag(t, cmd, "reverse", "true")
	setSearchFlag(t, cmd, "time", "client")
	setSearchFlag(t, cmd, "time-format", "rfc3339")
	setSearchFlag(t, cmd, "content", "true")
	setSearchFlag(t, cmd, "limit", "2")
	setSearchFlag(t, cmd, "order-by", "modified")
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
	if firstArg.Options.FilenameOnly {
		t.Fatal("filename_only = true, want false with --content")
	}
	if firstArg.Options.MaxResults != 2 {
		t.Fatalf("max_results = %d, want 2", firstArg.Options.MaxResults)
	}
	if firstArg.Options.OrderBy == nil || firstArg.Options.OrderBy.Tag != files.SearchOrderByLastModifiedTime {
		t.Fatalf("order_by = %#v, want last_modified_time", firstArg.Options.OrderBy)
	}

	got := decodeSearchOutput(t, stdout)
	if got.Input.Query != "report" || got.Input.Path != "/docs" {
		t.Fatalf("input = %#v, want query report path /docs", got.Input)
	}
	if !got.Input.Content || got.Input.Limit != 2 || got.Input.OrderBy != "modified" || !got.Input.Long || got.Input.Sort != "name" || !got.Input.Reverse || got.Input.Time != "client" || got.Input.TimeFormat != "rfc3339" {
		t.Fatalf("input options = %#v, want content/limit/order-by/long/sort/reverse/time/time-format", got.Input)
	}
	if len(got.Results) != 2 {
		t.Fatalf("results = %d, want 2", len(got.Results))
	}
	first := got.Results[0].Result
	if got.Results[0].Status != searchJSONStatusFound || got.Results[0].Kind != "file" || first.Type != "file" || first.Rev != "rev-file" || first.Size == nil || *first.Size != 42 {
		t.Fatalf("first result = %#v, want found file metadata", got.Results[0])
	}
	second := got.Results[1].Result
	if got.Results[1].Status != searchJSONStatusFound || got.Results[1].Kind != "folder" || second.Type != "folder" || second.ID != "id:folder" {
		t.Fatalf("second result = %#v, want found folder metadata", got.Results[1])
	}
	if strings.Contains(stdout.String(), `"entries"`) {
		t.Fatalf("JSON output = %s, want operation results and no entries key", stdout.String())
	}
}

func TestSearchJSONOmitsPathWithoutScope(t *testing.T) {
	cmd, stdout := testSearchCmd()
	setSearchOutputJSON(t, cmd)

	mock := &mockFilesClient{
		searchV2Fn: func(arg *files.SearchV2Arg) (*files.SearchV2Result, error) {
			if arg.Options == nil {
				t.Fatal("options = nil, want search options")
			}
			if arg.Options.Path != "" {
				t.Fatalf("options path = %q, want empty", arg.Options.Path)
			}
			if !arg.Options.FilenameOnly {
				t.Fatal("filename_only = false, want true by default")
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

func TestSearchLimitCapsResultsAndStopsPagination(t *testing.T) {
	cmd, stdout := testSearchCmd()
	setSearchFlag(t, cmd, "limit", "1")
	var firstArg *files.SearchV2Arg

	mock := &mockFilesClient{
		searchV2Fn: func(arg *files.SearchV2Arg) (*files.SearchV2Result, error) {
			firstArg = arg
			res := files.NewSearchV2Result([]*files.SearchMatchV2{
				searchMatch(&files.FileMetadata{
					Metadata: files.Metadata{PathDisplay: "/docs/first.txt"},
				}),
				searchMatch(&files.FileMetadata{
					Metadata: files.Metadata{PathDisplay: "/docs/second.txt"},
				}),
			}, true)
			res.Cursor = "cursor-1"
			return res, nil
		},
		searchContinueV2Fn: func(arg *files.SearchV2ContinueArg) (*files.SearchV2Result, error) {
			t.Fatalf("SearchContinueV2 was called with cursor %q, want stop after limit", arg.Cursor)
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	if err := search(cmd, []string{"needle", "/docs"}); err != nil {
		t.Fatalf("search error: %v", err)
	}
	if firstArg == nil || firstArg.Options == nil {
		t.Fatal("SearchV2 options were not set")
	}
	if firstArg.Options.MaxResults != 1 {
		t.Fatalf("max_results = %d, want 1", firstArg.Options.MaxResults)
	}
	got := stdout.String()
	if !strings.Contains(got, "/docs/first.txt") {
		t.Fatalf("stdout = %q, want first result", got)
	}
	if strings.Contains(got, "/docs/second.txt") {
		t.Fatalf("stdout = %q, want second result capped by --limit", got)
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
	cmd.Flags().BoolP("content", "c", false, "")
	cmd.Flags().Uint64("limit", 0, "")
	cmd.Flags().String("order-by", "", "")
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

func decodeSearchOutput(t *testing.T, out *bytes.Buffer) metadataOperationOutputForTest[searchInput] {
	t.Helper()

	return decodeMetadataOperationOutput[searchInput](t, out)
}

func searchMatch(metadata files.IsMetadata) *files.SearchMatchV2 {
	return files.NewSearchMatchV2(&files.MetadataV2{
		Tagged:   dropbox.Tagged{Tag: files.MetadataV2Metadata},
		Metadata: metadata,
	})
}
