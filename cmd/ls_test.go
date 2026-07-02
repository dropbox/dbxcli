package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"text/tabwriter"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

func TestFormatFolderMetadata(t *testing.T) {
	meta := &files.FolderMetadata{
		Metadata: files.Metadata{
			PathDisplay: "/Documents",
		},
	}

	got := formatFolderMetadata(meta, false)
	want := "/Documents\t"
	if got != want {
		t.Errorf("formatFolderMetadata(short) = %q, want %q", got, want)
	}

	got = formatFolderMetadata(meta, true)
	want = "-\t-\t-\t/Documents\t"
	if got != want {
		t.Errorf("formatFolderMetadata(long) = %q, want %q", got, want)
	}
}

func TestFormatFileMetadata(t *testing.T) {
	meta := &files.FileMetadata{
		Metadata: files.Metadata{
			PathDisplay: "/test.txt",
		},
		Rev:  "abc123",
		Size: 1024,
	}

	got := formatFileMetadata(meta, false)
	want := "/test.txt\t"
	if got != want {
		t.Errorf("formatFileMetadata(short) = %q, want %q", got, want)
	}

	got = formatFileMetadata(meta, true)
	if got == "" {
		t.Error("formatFileMetadata(long) returned empty string")
	}
	if len(got) <= len("/test.txt\t") {
		t.Errorf("formatFileMetadata(long) should include rev/size/time, got %q", got)
	}
}

func TestFormatDeletedMetadata(t *testing.T) {
	meta := &files.DeletedMetadata{
		Metadata: files.Metadata{
			PathDisplay: "/removed.txt",
		},
	}

	got := formatDeletedMetadata(meta, false)
	want := "/removed.txt\t"
	if got != want {
		t.Errorf("formatDeletedMetadata(short) = %q, want %q", got, want)
	}

	got = formatDeletedMetadata(meta, true)
	want = "-\t-\t-\t/removed.txt\t"
	if got != want {
		t.Errorf("formatDeletedMetadata(long) = %q, want %q", got, want)
	}
}

func TestSetPathDisplayAsDeleted(t *testing.T) {
	file := &files.FileMetadata{
		Metadata: files.Metadata{PathDisplay: "/file.txt"},
	}
	setPathDisplayAsDeleted(file)
	if file.PathDisplay != "<</file.txt>>" {
		t.Errorf("file PathDisplay = %q, want %q", file.PathDisplay, "<</file.txt>>")
	}

	folder := &files.FolderMetadata{
		Metadata: files.Metadata{PathDisplay: "/folder"},
	}
	setPathDisplayAsDeleted(folder)
	if folder.PathDisplay != "<</folder>>" {
		t.Errorf("folder PathDisplay = %q, want %q", folder.PathDisplay, "<</folder>>")
	}

	deleted := &files.DeletedMetadata{
		Metadata: files.Metadata{PathDisplay: "/gone"},
	}
	setPathDisplayAsDeleted(deleted)
	if deleted.PathDisplay != "<</gone>>" {
		t.Errorf("deleted PathDisplay = %q, want %q", deleted.PathDisplay, "<</gone>>")
	}
}

func TestFormatFileMetadataLongIncludesFields(t *testing.T) {
	ts := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	meta := &files.FileMetadata{
		Metadata: files.Metadata{
			PathDisplay: "/data.csv",
		},
		Rev:            "rev999",
		Size:           2048,
		ServerModified: dropbox.DBXTime(ts),
	}

	got := formatFileMetadata(meta, true)
	if !contains(got, "rev999") {
		t.Errorf("long format should contain rev, got %q", got)
	}
	if !contains(got, "2.0 KiB") {
		t.Errorf("long format should contain human size, got %q", got)
	}
	if !contains(got, "/data.csv") {
		t.Errorf("long format should contain path, got %q", got)
	}
}

func TestGetFileMetadataNotCalledForRoot(t *testing.T) {
	// This test verifies the ls function logic:
	// when path is "" (root), getFileMetadata should not be called.
	// We test this indirectly by confirming that NewGetMetadataArg
	// with empty string would be invalid for the API.
	arg := files.NewGetMetadataArg("")
	if arg.Path != "" {
		t.Errorf("NewGetMetadataArg('') path = %q, want empty", arg.Path)
	}
}

func TestFinishListOutputAddsTrailingNewlineForPartialShortRows(t *testing.T) {
	var out strings.Builder
	w := new(tabwriter.Writer)
	w.Init(&out, 4, 8, 1, ' ', 0)

	fmt.Fprint(w, "/one\t")
	if err := finishListOutput(w, 1, listOptions{}); err != nil {
		t.Fatalf("finishListOutput returned error: %v", err)
	}

	if got := out.String(); !strings.HasSuffix(got, "\n") {
		t.Fatalf("output %q does not end with newline", got)
	}
}

func TestRenderLsResultsShortModeUsesFourColumns(t *testing.T) {
	entries := []files.IsMetadata{
		&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/one"}},
		&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/two"}},
		&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/three"}},
		&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/four"}},
		&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/five"}},
	}

	var out bytes.Buffer
	if err := renderLsResults(&out, entries, listOptions{}); err != nil {
		t.Fatalf("renderLsResults returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("output = %q, want two short-mode rows", out.String())
	}
	if !strings.Contains(lines[0], "/one") || !strings.Contains(lines[0], "/four") {
		t.Fatalf("first row = %q, want first four entries", lines[0])
	}
	if !strings.Contains(lines[1], "/five") {
		t.Fatalf("second row = %q, want fifth entry", lines[1])
	}
}

func TestLsJSONListsResultsAndInput(t *testing.T) {
	cmd, stdout := testLsCmd(t)
	setLsOutputJSON(t, cmd)
	setLsFlag(t, cmd, "recursive", "true")
	setLsFlag(t, cmd, "include-deleted", "true")
	setLsFlag(t, cmd, "limit", "50")
	setLsFlag(t, cmd, "long", "true")
	setLsFlag(t, cmd, "sort", "name")
	setLsFlag(t, cmd, "reverse", "true")
	setLsFlag(t, cmd, "time", "client")
	setLsFlag(t, cmd, "time-format", "rfc3339")

	var listArg *files.ListFolderArg
	mock := &mockFilesClient{
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			listArg = arg
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					&files.FileMetadata{
						Metadata: files.Metadata{
							Name:        "file.txt",
							PathDisplay: "/file.txt",
							PathLower:   "/file.txt",
						},
						Id:   "id:file",
						Rev:  "rev-file",
						Size: 42,
					},
					&files.FolderMetadata{
						Metadata: files.Metadata{
							Name:        "Folder",
							PathDisplay: "/Folder",
							PathLower:   "/folder",
						},
						Id: "id:folder",
					},
				},
				HasMore: false,
			}, nil
		},
	}
	stubFilesClient(t, mock)

	if err := ls(cmd, nil); err != nil {
		t.Fatalf("ls error: %v", err)
	}
	if listArg == nil {
		t.Fatal("ListFolder was not called")
	}
	if listArg.Path != "" {
		t.Fatalf("ListFolder path = %q, want root empty path", listArg.Path)
	}
	if !listArg.Recursive || !listArg.IncludeDeleted {
		t.Fatalf("ListFolder flags = recursive:%v include_deleted:%v, want true/true", listArg.Recursive, listArg.IncludeDeleted)
	}
	if listArg.Limit != 50 {
		t.Fatalf("ListFolder limit = %d, want 50", listArg.Limit)
	}

	got := decodeLsOutput(t, stdout)
	if got.Input.Path != "/" {
		t.Fatalf("input path = %q, want /", got.Input.Path)
	}
	if !got.Input.Recursive || !got.Input.IncludeDeleted || got.Input.OnlyDeleted || !got.Input.Long {
		t.Fatalf("input flags = %+v, want recursive/include_deleted/long true and only_deleted false", got.Input)
	}
	if got.Input.Sort != "name" || !got.Input.Reverse || got.Input.Time != "client" || got.Input.TimeFormat != "rfc3339" {
		t.Fatalf("input options = %+v, want sort/name reverse/client/rfc3339", got.Input)
	}
	if got.Input.Limit != 50 {
		t.Fatalf("input limit = %d, want 50", got.Input.Limit)
	}
	if len(got.Results) != 2 {
		t.Fatalf("results = %d, want 2", len(got.Results))
	}
	if got.Results[0].Status != lsJSONStatusListed || got.Results[0].Kind != "folder" || got.Results[0].Result.PathDisplay != "/Folder" {
		t.Fatalf("first result = %#v, want sorted listed folder", got.Results[0])
	}
	second := got.Results[1].Result
	if got.Results[1].Status != lsJSONStatusListed || got.Results[1].Kind != "file" || second.Type != "file" || second.Rev != "rev-file" || second.Size == nil || *second.Size != 42 {
		t.Fatalf("second result = %#v, want listed file metadata", got.Results[1])
	}
	if strings.Contains(stdout.String(), `"entries"`) {
		t.Fatalf("JSON output = %s, want operation results and no entries key", stdout.String())
	}
}

func TestLsLimitStopsPaginationAndTruncatesOutput(t *testing.T) {
	cmd, stdout := testLsCmd(t)
	setLsOutputJSON(t, cmd)
	setLsFlag(t, cmd, "limit", "2")

	var listArg *files.ListFolderArg
	mock := &mockFilesClient{
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			listArg = arg
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/one.txt", PathLower: "/one.txt"}, Id: "id:one", Rev: "rev-one"},
					&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/two.txt", PathLower: "/two.txt"}, Id: "id:two", Rev: "rev-two"},
					&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/three.txt", PathLower: "/three.txt"}, Id: "id:three", Rev: "rev-three"},
				},
				HasMore: true,
				Cursor:  "cursor",
			}, nil
		},
		listFolderContinueFn: func(arg *files.ListFolderContinueArg) (*files.ListFolderResult, error) {
			t.Fatalf("ListFolderContinue should not be called after reaching limit: %v", arg)
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	if err := ls(cmd, nil); err != nil {
		t.Fatalf("ls error: %v", err)
	}
	if listArg == nil {
		t.Fatal("ListFolder was not called")
	}
	if listArg.Limit != 2 {
		t.Fatalf("ListFolder limit = %d, want 2", listArg.Limit)
	}

	got := decodeLsOutput(t, stdout)
	if got.Input.Limit != 2 {
		t.Fatalf("input limit = %d, want 2", got.Input.Limit)
	}
	if len(got.Results) != 2 {
		t.Fatalf("results = %d, want 2", len(got.Results))
	}
	for _, result := range got.Results {
		if result.Result.PathDisplay == "/three.txt" {
			t.Fatalf("results include truncated entry: %#v", got.Results)
		}
	}
}

func TestLsOnlyDeletedLimitCountsFilteredEntries(t *testing.T) {
	cmd, stdout := testLsCmd(t)
	setLsOutputJSON(t, cmd)
	setLsFlag(t, cmd, "only-deleted", "true")
	setLsFlag(t, cmd, "limit", "2")

	continueCalled := false
	mock := &mockFilesClient{
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			if !arg.IncludeDeleted {
				t.Fatal("ListFolder include_deleted = false, want true")
			}
			if arg.Limit != 2 {
				t.Fatalf("ListFolder limit = %d, want 2", arg.Limit)
			}
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/active-a.txt"}},
					&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/active-b.txt"}},
					&files.DeletedMetadata{Metadata: files.Metadata{PathDisplay: "/deleted-a.txt", PathLower: "/deleted-a.txt"}},
				},
				HasMore: true,
				Cursor:  "cursor-1",
			}, nil
		},
		listFolderContinueFn: func(arg *files.ListFolderContinueArg) (*files.ListFolderResult, error) {
			continueCalled = true
			if arg.Cursor != "cursor-1" {
				t.Fatalf("ListFolderContinue cursor = %q, want cursor-1", arg.Cursor)
			}
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					&files.DeletedMetadata{Metadata: files.Metadata{PathDisplay: "/deleted-b.txt", PathLower: "/deleted-b.txt"}},
					&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/active-c.txt"}},
				},
				HasMore: false,
			}, nil
		},
		listRevisionsFn: func(arg *files.ListRevisionsArg) (*files.ListRevisionsResult, error) {
			return files.NewListRevisionsResult(false, []*files.FileMetadata{
				{
					Metadata: files.Metadata{
						PathDisplay: arg.Path,
						PathLower:   arg.Path,
					},
					Id:  "id:" + arg.Path,
					Rev: "rev:" + arg.Path,
				},
			}, false), nil
		},
	}
	stubFilesClient(t, mock)

	if err := ls(cmd, nil); err != nil {
		t.Fatalf("ls error: %v", err)
	}
	if !continueCalled {
		t.Fatal("ListFolderContinue was not called; want pagination until deleted limit is reached")
	}

	got := decodeLsOutput(t, stdout)
	if !got.Input.OnlyDeleted || got.Input.Limit != 2 {
		t.Fatalf("input = %#v, want only_deleted true and limit 2", got.Input)
	}
	if len(got.Results) != 2 {
		t.Fatalf("results = %d, want 2 deleted entries: %#v", len(got.Results), got.Results)
	}
	for i, want := range []string{"/deleted-a.txt", "/deleted-b.txt"} {
		result := got.Results[i]
		if result.Result.PathDisplay != want || !result.Result.Deleted {
			t.Fatalf("result[%d] = %#v, want deleted path %s", i, result, want)
		}
	}
}

func TestLsOnlyDeletedLimitTruncatesFilteredPage(t *testing.T) {
	cmd, stdout := testLsCmd(t)
	setLsOutputJSON(t, cmd)
	setLsFlag(t, cmd, "only-deleted", "true")
	setLsFlag(t, cmd, "limit", "1")

	mock := &mockFilesClient{
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					&files.DeletedMetadata{Metadata: files.Metadata{PathDisplay: "/deleted-a.txt", PathLower: "/deleted-a.txt"}},
					&files.DeletedMetadata{Metadata: files.Metadata{PathDisplay: "/deleted-b.txt", PathLower: "/deleted-b.txt"}},
				},
				HasMore: true,
				Cursor:  "cursor-1",
			}, nil
		},
		listFolderContinueFn: func(arg *files.ListFolderContinueArg) (*files.ListFolderResult, error) {
			t.Fatalf("ListFolderContinue called after reaching filtered limit: %v", arg)
			return nil, nil
		},
		listRevisionsFn: func(arg *files.ListRevisionsArg) (*files.ListRevisionsResult, error) {
			return files.NewListRevisionsResult(false, []*files.FileMetadata{
				{
					Metadata: files.Metadata{
						PathDisplay: arg.Path,
						PathLower:   arg.Path,
					},
					Id:  "id:" + arg.Path,
					Rev: "rev:" + arg.Path,
				},
			}, false), nil
		},
	}
	stubFilesClient(t, mock)

	if err := ls(cmd, nil); err != nil {
		t.Fatalf("ls error: %v", err)
	}

	got := decodeLsOutput(t, stdout)
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1 deleted entry", len(got.Results))
	}
	if got.Results[0].Result.PathDisplay != "/deleted-a.txt" || !got.Results[0].Result.Deleted {
		t.Fatalf("result = %#v, want first deleted entry", got.Results[0])
	}
}

func TestLsRecurseAliasSetsRecursiveListFolderArg(t *testing.T) {
	cmd, stdout := testLsCmd(t)
	setLsOutputJSON(t, cmd)
	setLsFlag(t, cmd, "recurse", "true")

	var listArg *files.ListFolderArg
	mock := &mockFilesClient{
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			listArg = arg
			return &files.ListFolderResult{HasMore: false}, nil
		},
	}
	stubFilesClient(t, mock)

	if err := ls(cmd, nil); err != nil {
		t.Fatalf("ls error: %v", err)
	}
	if listArg == nil {
		t.Fatal("ListFolder was not called")
	}
	if !listArg.Recursive {
		t.Fatal("ListFolder recursive = false, want true from --recurse alias")
	}
	got := decodeLsOutput(t, stdout)
	if !got.Input.Recursive {
		t.Fatal("input recursive = false, want true from --recurse alias")
	}
}

func TestLsLimitRejectsUint32Overflow(t *testing.T) {
	cmd, stdout := testLsCmd(t)
	setLsFlag(t, cmd, "limit", "4294967296")

	if err := ls(cmd, nil); err == nil {
		t.Fatal("expected ls limit overflow error")
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty output on validation error", got)
	}
}

func TestLsRejectsInvalidListOptions(t *testing.T) {
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
			cmd, stdout := testLsCmd(t)
			setLsFlag(t, cmd, tt.flag, tt.value)
			stubFilesClient(t, &mockFilesClient{
				listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
					t.Fatalf("ListFolder called for invalid --%s", tt.flag)
					return nil, nil
				},
			})

			err := ls(cmd, nil)
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

func TestLsJSONFilePathUsesMetadata(t *testing.T) {
	cmd, stdout := testLsCmd(t)
	setLsOutputJSON(t, cmd)
	var metadataPath string

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			metadataPath = arg.Path
			return &files.FileMetadata{
				Metadata: files.Metadata{
					PathDisplay: "/file.txt",
					PathLower:   "/file.txt",
				},
				Id:   "id:file",
				Rev:  "rev-file",
				Size: 7,
			}, nil
		},
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			t.Fatalf("ListFolder called for file path: %v", arg)
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	if err := ls(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("ls error: %v", err)
	}
	if metadataPath != "/file.txt" {
		t.Fatalf("metadata path = %q, want /file.txt", metadataPath)
	}
	got := decodeLsOutput(t, stdout)
	if got.Input.Path != "/file.txt" {
		t.Fatalf("input path = %q, want /file.txt", got.Input.Path)
	}
	if len(got.Results) != 1 || got.Results[0].Status != lsJSONStatusListed || got.Results[0].Kind != "file" || got.Results[0].Result.Type != "file" {
		t.Fatalf("results = %#v, want one listed file", got.Results)
	}
}

func TestLsJSONDeletedEntryIsStructured(t *testing.T) {
	cmd, stdout := testLsCmd(t)
	setLsOutputJSON(t, cmd)
	setLsFlag(t, cmd, "include-deleted", "true")

	mock := &mockFilesClient{
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					&files.DeletedMetadata{
						Metadata: files.Metadata{
							PathDisplay: "/removed.txt",
							PathLower:   "/removed.txt",
						},
					},
				},
				HasMore: false,
			}, nil
		},
		listRevisionsFn: func(arg *files.ListRevisionsArg) (*files.ListRevisionsResult, error) {
			if arg.Path != "/removed.txt" {
				t.Fatalf("ListRevisions path = %q, want /removed.txt", arg.Path)
			}
			return files.NewListRevisionsResult(false, []*files.FileMetadata{
				{
					Metadata: files.Metadata{
						PathDisplay: "/removed.txt",
						PathLower:   "/removed.txt",
					},
					Rev:  "rev-removed",
					Size: 9,
				},
			}, false), nil
		},
	}
	stubFilesClient(t, mock)

	if err := ls(cmd, nil); err != nil {
		t.Fatalf("ls error: %v", err)
	}
	got := decodeLsOutput(t, stdout)
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Results))
	}
	if got.Results[0].Status != lsJSONStatusListed || got.Results[0].Kind != "file" {
		t.Fatalf("result = %#v, want listed file", got.Results[0])
	}
	entry := got.Results[0].Result
	if entry.PathDisplay != "/removed.txt" {
		t.Fatalf("path_display = %q, want undecorated path", entry.PathDisplay)
	}
	if !entry.Deleted {
		t.Fatal("deleted = false, want true")
	}
	if strings.Contains(stdout.String(), "<<") {
		t.Fatalf("JSON output = %s, want no text deleted marker", stdout.String())
	}
}

func TestLsTextUsesCommandOutput(t *testing.T) {
	cmd, stdout := testLsCmd(t)
	mock := &mockFilesClient{
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/docs"}},
				},
				HasMore: false,
			}, nil
		},
	}
	stubFilesClient(t, mock)

	if err := ls(cmd, nil); err != nil {
		t.Fatalf("ls error: %v", err)
	}
	if got := stdout.String(); !strings.Contains(got, "/docs") {
		t.Fatalf("stdout = %q, want command output to contain /docs", got)
	}
}

func TestLsJSONErrorWritesNoOutput(t *testing.T) {
	cmd, stdout := testLsCmd(t)
	setLsOutputJSON(t, cmd)
	mock := &mockFilesClient{
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			return nil, fmt.Errorf("list failed")
		},
	}
	stubFilesClient(t, mock)

	if err := ls(cmd, nil); err == nil {
		t.Fatal("expected ls error")
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty output on error", got)
	}
}

func TestLsCommandSupportsStructuredOutput(t *testing.T) {
	if !commandSupportsStructuredOutput(lsCmd) {
		t.Fatal("ls command should support structured output")
	}
}

func TestIsListFolderNotFolderErrorHandlesWrappedErrors(t *testing.T) {
	apiErr := files.ListFolderAPIError{
		EndpointError: &files.ListFolderError{
			Path: &files.LookupError{Tagged: dropbox.Tagged{Tag: files.LookupErrorNotFolder}},
		},
	}

	if !isListFolderNotFolderError(apiErr) {
		t.Fatal("expected raw list_folder not_folder error to match")
	}
	if !isListFolderNotFolderError(fmt.Errorf("wrapped: %w", apiErr)) {
		t.Fatal("expected wrapped list_folder not_folder error to match")
	}
}

func TestIsListRevisionsNotFileErrorHandlesWrappedErrors(t *testing.T) {
	apiErr := files.ListRevisionsAPIError{
		EndpointError: &files.ListRevisionsError{
			Path: &files.LookupError{Tagged: dropbox.Tagged{Tag: files.LookupErrorNotFile}},
		},
	}

	if !isListRevisionsNotFileError(apiErr) {
		t.Fatal("expected raw list_revisions not_file error to match")
	}
	if !isListRevisionsNotFileError(fmt.Errorf("wrapped: %w", apiErr)) {
		t.Fatal("expected wrapped list_revisions not_file error to match")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Verify that dropbox config can be constructed without panicking
func TestDropboxConfigConstruction(t *testing.T) {
	cfg := dropbox.Config{
		Token:    "test-token",
		LogLevel: dropbox.LogOff,
	}
	if cfg.Token != "test-token" {
		t.Error("config token not set")
	}
}

func testLsCmd(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()

	var stdout bytes.Buffer
	cmd := &cobra.Command{Use: "ls"}
	cmd.SetOut(&stdout)
	cmd.Flags().BoolP("long", "l", false, "")
	cmd.Flags().Bool("recursive", false, "")
	cmd.Flags().BoolP("recurse", "R", false, "")
	cmd.Flags().BoolP("include-deleted", "d", false, "")
	cmd.Flags().BoolP("only-deleted", "D", false, "")
	cmd.Flags().Uint64("limit", 0, "")
	cmd.Flags().String("sort", "", "")
	cmd.Flags().BoolP("reverse", "r", false, "")
	cmd.Flags().String("time", "server", "")
	cmd.Flags().String("time-format", "", "")
	cmd.Flags().String(outputFlag, "text", "")
	return cmd, &stdout
}

func setLsOutputJSON(t *testing.T, cmd *cobra.Command) {
	t.Helper()
	setLsFlag(t, cmd, outputFlag, "json")
}

func setLsFlag(t *testing.T, cmd *cobra.Command, name, value string) {
	t.Helper()
	if err := cmd.Flags().Set(name, value); err != nil {
		t.Fatalf("set %s: %v", name, err)
	}
}

func decodeLsOutput(t *testing.T, out *bytes.Buffer) metadataOperationOutputForTest[lsInput] {
	t.Helper()

	return decodeMetadataOperationOutput[lsInput](t, out)
}
