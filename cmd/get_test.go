package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	dbxauth "github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/auth"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

type failingReadCloser struct {
	data []byte
	read bool
}

func (r *failingReadCloser) Read(p []byte) (int, error) {
	if r.read {
		return 0, io.EOF
	}
	r.read = true
	n := copy(p, r.data)
	return n, io.ErrUnexpectedEOF
}

func (r *failingReadCloser) Close() error { return nil }

func testGetJSONCmd(stdout, stderr *bytes.Buffer) *cobra.Command {
	cmd := testGetCmd()
	cmd.Flags().String(outputFlag, "text", "")
	if err := cmd.Flags().Set(outputFlag, "json"); err != nil {
		panic(err)
	}
	if stdout != nil {
		cmd.SetOut(stdout)
	}
	if stderr != nil {
		cmd.SetErr(stderr)
	}
	return cmd
}

type getOutputData struct {
	Input    getCommandInput `json:"input"`
	Results  []getResult     `json:"results"`
	Warnings []jsonWarning   `json:"warnings"`
}

func decodeGetOutput(t *testing.T, stdout *bytes.Buffer) getOutputData {
	t.Helper()

	var got getOutputData
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode get JSON output: %v\noutput: %s", err, stdout.String())
	}
	if got.Warnings == nil {
		t.Fatalf("warnings = nil, want empty array")
	}
	if len(got.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want empty", got.Warnings)
	}
	return got
}

func getTestFileMetadata(path string, size uint64) *files.FileMetadata {
	return &files.FileMetadata{
		Metadata: files.Metadata{
			Name:        strings.TrimPrefix(path, "/"),
			PathDisplay: path,
			PathLower:   strings.ToLower(path),
		},
		Id:   "id:" + strings.TrimPrefix(path, "/"),
		Rev:  "rev:" + strings.TrimPrefix(path, "/"),
		Size: size,
	}
}

func getTestFolderMetadata(path string) *files.FolderMetadata {
	return &files.FolderMetadata{
		Metadata: files.Metadata{
			Name:        strings.TrimPrefix(path, "/"),
			PathDisplay: path,
			PathLower:   strings.ToLower(path),
		},
		Id: "id:" + strings.TrimPrefix(path, "/"),
	}
}

func TestGetArgValidation(t *testing.T) {
	err := get(getCmd, []string{})
	if err == nil {
		t.Error("expected error for no args")
	}

	err = get(getCmd, []string{"a", "b", "c"})
	if err == nil {
		t.Error("expected error for too many args")
	}
}

func TestGetDstDefaultsToBasename(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	content := "downloaded content"
	mock := &mockFilesClient{
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			if arg.Path != "/some/path/file.txt" {
				t.Errorf("download path = %q, want %q", arg.Path, "/some/path/file.txt")
			}
			meta := &files.FileMetadata{
				Metadata: files.Metadata{PathDisplay: "/some/path/file.txt"},
				Size:     uint64(len(content)),
			}
			return meta, io.NopCloser(strings.NewReader(content)), nil
		},
	}

	err := getWithClient(mock, []string{"/some/path/file.txt"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	got, err := os.ReadFile(filepath.Join(tmpDir, "file.txt"))
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(got) != content {
		t.Errorf("got %q, want %q", string(got), content)
	}
}

func TestGetDstAppendsToDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	content := "pdf content"
	mock := &mockFilesClient{
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			if arg.Path != "/remote/doc.pdf" {
				t.Errorf("download path = %q, want %q", arg.Path, "/remote/doc.pdf")
			}
			meta := &files.FileMetadata{
				Metadata: files.Metadata{PathDisplay: "/remote/doc.pdf"},
				Size:     uint64(len(content)),
			}
			return meta, io.NopCloser(strings.NewReader(content)), nil
		},
	}

	err := getWithClient(mock, []string{"/remote/doc.pdf", tmpDir})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	got, err := os.ReadFile(filepath.Join(tmpDir, "doc.pdf"))
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(got) != content {
		t.Errorf("got %q, want %q", string(got), content)
	}
}

func TestGetDownloadWithRetry(t *testing.T) {
	stubRetrySleep(t)
	tmpDir := t.TempDir()
	dst := filepath.Join(tmpDir, "downloaded.txt")
	content := "file content here"

	calls := 0
	mock := &mockFilesClient{
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			calls++
			if calls < 3 {
				return nil, nil, dbxauth.ServerError{APIError: dropbox.APIError{ErrorSummary: "500"}}
			}
			meta := &files.FileMetadata{
				Metadata: files.Metadata{PathDisplay: "/test.txt"},
				Size:     uint64(len(content)),
			}
			return meta, io.NopCloser(strings.NewReader(content)), nil
		},
	}

	err := downloadFile(mock, "/test.txt", dst)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(got) != content {
		t.Errorf("got %q, want %q", string(got), content)
	}
}

func TestGetDownloadRetriesBodyReadError(t *testing.T) {
	delays := stubRetrySleep(t)
	tmpDir := t.TempDir()
	dst := filepath.Join(tmpDir, "downloaded.txt")
	content := "complete file content"

	calls := 0
	mock := &mockFilesClient{
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			calls++
			meta := &files.FileMetadata{
				Metadata: files.Metadata{PathDisplay: "/test.txt"},
				Size:     uint64(len(content)),
			}
			if calls == 1 {
				return meta, &failingReadCloser{data: []byte("partial")}, nil
			}
			return meta, io.NopCloser(strings.NewReader(content)), nil
		},
	}

	err := downloadFile(mock, "/test.txt", dst)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
	if len(*delays) != 1 {
		t.Fatalf("expected 1 sleep, got %d", len(*delays))
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(got) != content {
		t.Errorf("got %q, want %q", string(got), content)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read temp dir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(dst) {
		t.Fatalf("expected only final destination file, got %v", entries)
	}
}

func TestGetDownloadPreservesDestinationSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target.txt")
	link := filepath.Join(tmpDir, "link.txt")
	if err := os.WriteFile(target, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	content := "new content"
	mock := &mockFilesClient{
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			meta := &files.FileMetadata{
				Metadata: files.Metadata{PathDisplay: "/test.txt"},
				Size:     uint64(len(content)),
			}
			return meta, io.NopCloser(strings.NewReader(content)), nil
		},
	}

	err := downloadFile(mock, "/test.txt", link)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("failed to stat symlink: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected %s to remain a symlink, mode %v", link, info.Mode())
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read symlink target: %v", err)
	}
	if string(got) != content {
		t.Errorf("target content = %q, want %q", string(got), content)
	}
}

func TestGetDownloadPermanentError(t *testing.T) {
	tmpDir := t.TempDir()
	dst := filepath.Join(tmpDir, "downloaded.txt")

	mock := &mockFilesClient{
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			return nil, nil, &files.DownloadAPIError{}
		},
	}

	err := downloadFile(mock, "/nonexistent.txt", dst)
	if err == nil {
		t.Error("expected error for permanent failure, got nil")
	}
}

func TestGetRecursive_DownloadsDirectoryStructure(t *testing.T) {
	tmpDir := t.TempDir()
	dst := filepath.Join(tmpDir, "output")

	mock := &mockFilesClient{
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			if arg.Path != "/remote" {
				t.Errorf("ListFolder path = %q, want /remote", arg.Path)
			}
			if !arg.Recursive {
				t.Error("expected Recursive = true")
			}
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/remote/sub"}},
					&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/remote/root.txt"}, Size: 4},
					&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/remote/sub/deep.txt"}, Size: 4},
				},
				HasMore: false,
			}, nil
		},
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			meta := &files.FileMetadata{
				Metadata: files.Metadata{PathDisplay: arg.Path},
				Size:     4,
			}
			return meta, io.NopCloser(strings.NewReader("data")), nil
		},
	}

	err := getRecursive(mock, "/remote", dst)
	if err != nil {
		t.Fatalf("getRecursive error: %v", err)
	}

	for _, rel := range []string{"root.txt", "sub/deep.txt"} {
		p := filepath.Join(dst, filepath.FromSlash(rel))
		got, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("failed to read %s: %v", rel, err)
			continue
		}
		if string(got) != "data" {
			t.Errorf("%s content = %q, want %q", rel, string(got), "data")
		}
	}

	info, err := os.Stat(filepath.Join(dst, "sub"))
	if err != nil {
		t.Fatalf("sub directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("sub is not a directory")
	}
}

func TestGetRecursive_CreatesEmptyDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	dst := filepath.Join(tmpDir, "output")

	mock := &mockFilesClient{
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/remote/empty"}},
					&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/remote/empty/nested"}},
				},
				HasMore: false,
			}, nil
		},
	}

	err := getRecursive(mock, "/remote", dst)
	if err != nil {
		t.Fatalf("getRecursive error: %v", err)
	}

	for _, dir := range []string{"empty", "empty/nested"} {
		p := filepath.Join(dst, filepath.FromSlash(dir))
		info, err := os.Stat(p)
		if err != nil {
			t.Errorf("directory %s not created: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}

func TestGetRecursive_HandlesPagination(t *testing.T) {
	tmpDir := t.TempDir()
	dst := filepath.Join(tmpDir, "output")

	mock := &mockFilesClient{
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/remote/a.txt"}, Size: 1},
				},
				HasMore: true,
				Cursor:  "page2",
			}, nil
		},
		listFolderContinueFn: func(arg *files.ListFolderContinueArg) (*files.ListFolderResult, error) {
			if arg.Cursor != "page2" {
				t.Errorf("cursor = %q, want page2", arg.Cursor)
			}
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/remote/b.txt"}, Size: 1},
				},
				HasMore: false,
			}, nil
		},
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			meta := &files.FileMetadata{
				Metadata: files.Metadata{PathDisplay: arg.Path},
				Size:     1,
			}
			return meta, io.NopCloser(strings.NewReader("x")), nil
		},
	}

	err := getRecursive(mock, "/remote", dst)
	if err != nil {
		t.Fatalf("getRecursive error: %v", err)
	}

	for _, name := range []string{"a.txt", "b.txt"} {
		if _, err := os.Stat(filepath.Join(dst, name)); err != nil {
			t.Errorf("file %s not downloaded: %v", name, err)
		}
	}
}

func TestGetRecursive_ReportsDownloadErrors(t *testing.T) {
	tmpDir := t.TempDir()
	dst := filepath.Join(tmpDir, "output")

	mock := &mockFilesClient{
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/remote/good.txt"}, Size: 4},
					&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/remote/bad.txt"}, Size: 4},
				},
				HasMore: false,
			}, nil
		},
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			if strings.Contains(arg.Path, "bad.txt") {
				return nil, nil, &files.DownloadAPIError{}
			}
			meta := &files.FileMetadata{
				Metadata: files.Metadata{PathDisplay: arg.Path},
				Size:     4,
			}
			return meta, io.NopCloser(strings.NewReader("data")), nil
		},
	}

	var err error
	captureStderr(t, func() {
		err = getRecursive(mock, "/remote", dst)
	})
	if err == nil {
		t.Fatal("expected error for failed downloads")
	}
	if !strings.Contains(err.Error(), "1 error") {
		t.Errorf("error = %q, want mention of 1 error", err.Error())
	}

	// Good file should still have been downloaded
	if _, statErr := os.Stat(filepath.Join(dst, "good.txt")); statErr != nil {
		t.Errorf("good.txt not downloaded: %v", statErr)
	}
}

func TestGetFolderWithoutRecursiveFlag(t *testing.T) {
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return &files.FolderMetadata{Metadata: files.Metadata{PathDisplay: arg.Path}}, nil
		},
	}
	stubFilesClient(t, mock)

	err := get(getCmd, []string{"/remote-folder"})
	if err == nil {
		t.Fatal("expected error for folder without --recursive")
	}
	if !strings.Contains(err.Error(), "--recursive") {
		t.Errorf("error = %q, want mention of --recursive", err.Error())
	}
}

func TestGetRecursiveCommandGetsMetadataThenListsFolder(t *testing.T) {
	tmpDir := t.TempDir()
	var calls []string

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			calls = append(calls, "metadata:"+arg.Path)
			return &files.FolderMetadata{Metadata: files.Metadata{PathDisplay: arg.Path}}, nil
		},
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			calls = append(calls, "list:"+arg.Path)
			if !arg.Recursive {
				t.Error("expected Recursive = true")
			}
			return &files.ListFolderResult{HasMore: false}, nil
		},
	}
	stubFilesClient(t, mock)

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("recursive", "r", true, "")

	err := get(cmd, []string{"/remote-folder", filepath.Join(tmpDir, "out")})
	if err != nil {
		t.Fatalf("get error: %v", err)
	}

	want := []string{"metadata:/remote-folder", "list:/remote-folder"}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("calls = %v, want %v", calls, want)
	}
}

func TestGetFileCommandDownloadsAfterMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	dst := filepath.Join(tmpDir, "file.txt")
	var calls []string

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			calls = append(calls, "metadata:"+arg.Path)
			return &files.FileMetadata{
				Metadata: files.Metadata{PathDisplay: arg.Path},
				Size:     4,
			}, nil
		},
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			calls = append(calls, "download:"+arg.Path)
			return &files.FileMetadata{
				Metadata: files.Metadata{PathDisplay: arg.Path},
				Size:     4,
			}, io.NopCloser(strings.NewReader("data")), nil
		},
	}
	stubFilesClient(t, mock)

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("recursive", "r", false, "")

	err := get(cmd, []string{"/remote-file.txt", dst})
	if err != nil {
		t.Fatalf("get error: %v", err)
	}

	want := []string{"metadata:/remote-file.txt", "download:/remote-file.txt"}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("calls = %v, want %v", calls, want)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(got) != "data" {
		t.Errorf("downloaded file = %q, want data", got)
	}
}

func TestGetJSONFileOutputsDownloadedResult(t *testing.T) {
	tmpDir := t.TempDir()
	dst := filepath.Join(tmpDir, "file.txt")

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return getTestFileMetadata(arg.Path, 4), nil
		},
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			if arg.Path != "/remote/file.txt" {
				t.Fatalf("download path = %q, want /remote/file.txt", arg.Path)
			}
			return getTestFileMetadata(arg.Path, 4), io.NopCloser(strings.NewReader("data")), nil
		},
	}
	stubFilesClient(t, mock)

	var stdout, stderr bytes.Buffer
	cmd := testGetJSONCmd(&stdout, &stderr)
	if err := get(cmd, []string{"/remote/file.txt", dst}); err != nil {
		t.Fatalf("get error: %v", err)
	}

	got := decodeGetOutput(t, &stdout)
	if got.Input.Source != "/remote/file.txt" || got.Input.Target != dst || got.Input.Recursive || got.Input.Stdout {
		t.Fatalf("input = %+v", got.Input)
	}
	if len(got.Results) != 1 {
		t.Fatalf("results len = %d, want 1", len(got.Results))
	}
	result := got.Results[0]
	if result.Status != getStatusDownloaded || result.Kind != getKindFile {
		t.Fatalf("result status/kind = %s/%s", result.Status, result.Kind)
	}
	if result.Input.Source != "/remote/file.txt" || result.Input.Target != dst {
		t.Fatalf("result input = %+v", result.Input)
	}
	if result.Result == nil || result.Result.Type != "file" || result.Result.PathDisplay != "/remote/file.txt" {
		t.Fatalf("metadata = %+v", result.Result)
	}
	if result.Result.Size == nil || *result.Result.Size != 4 {
		t.Fatalf("size = %v, want 4", result.Result.Size)
	}
	if strings.Contains(stdout.String(), "Downloading ") {
		t.Fatalf("stdout contains progress: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Downloading ") {
		t.Fatalf("stderr = %q, want download progress", stderr.String())
	}
	if content, err := os.ReadFile(dst); err != nil || string(content) != "data" {
		t.Fatalf("downloaded file = %q, %v; want data", string(content), err)
	}
}

func TestGetJSONDefaultTargetUsesBasename(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return getTestFileMetadata(arg.Path, 4), nil
		},
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			return getTestFileMetadata(arg.Path, 4), io.NopCloser(strings.NewReader("data")), nil
		},
	}
	stubFilesClient(t, mock)

	var stdout bytes.Buffer
	cmd := testGetJSONCmd(&stdout, nil)
	if err := get(cmd, []string{"/remote/file.txt"}); err != nil {
		t.Fatalf("get error: %v", err)
	}

	got := decodeGetOutput(t, &stdout)
	if got.Input.Target != "file.txt" {
		t.Fatalf("target = %q, want file.txt", got.Input.Target)
	}
	if got.Results[0].Input.Target != "file.txt" {
		t.Fatalf("result target = %q, want file.txt", got.Results[0].Input.Target)
	}
}

func TestGetJSONExistingDirectoryTargetAppendsBasename(t *testing.T) {
	tmpDir := t.TempDir()

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return getTestFileMetadata(arg.Path, 4), nil
		},
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			return getTestFileMetadata(arg.Path, 4), io.NopCloser(strings.NewReader("data")), nil
		},
	}
	stubFilesClient(t, mock)

	var stdout bytes.Buffer
	cmd := testGetJSONCmd(&stdout, nil)
	if err := get(cmd, []string{"/remote/file.txt", tmpDir}); err != nil {
		t.Fatalf("get error: %v", err)
	}

	wantTarget := filepath.Join(tmpDir, "file.txt")
	got := decodeGetOutput(t, &stdout)
	if got.Input.Target != wantTarget {
		t.Fatalf("target = %q, want %q", got.Input.Target, wantTarget)
	}
	if got.Results[0].Input.Target != wantTarget {
		t.Fatalf("result target = %q, want %q", got.Results[0].Input.Target, wantTarget)
	}
}

func TestGetJSONRejectsStdoutTarget(t *testing.T) {
	var stdout bytes.Buffer
	cmd := testGetJSONCmd(&stdout, nil)

	err := get(cmd, []string{"/remote/file.txt", "-"})
	if err == nil {
		t.Fatal("expected JSON stdout target error")
	}
	if !strings.Contains(err.Error(), "stdout target") {
		t.Fatalf("error = %q, want stdout target", err.Error())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestGetTextStdoutTargetWritesOnlyFileBytes(t *testing.T) {
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return getTestFileMetadata(arg.Path, 4), nil
		},
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			return getTestFileMetadata(arg.Path, 4), io.NopCloser(strings.NewReader("data")), nil
		},
	}
	stubFilesClient(t, mock)

	var stdout bytes.Buffer
	cmd := testGetCmd()
	cmd.SetOut(&stdout)
	if err := get(cmd, []string{"/remote/file.txt", "-"}); err != nil {
		t.Fatalf("get error: %v", err)
	}
	if stdout.String() != "data" {
		t.Fatalf("stdout = %q, want data", stdout.String())
	}
}

func TestGetJSONRecursiveOutputsDirectoryAndFileResults(t *testing.T) {
	tmpDir := t.TempDir()
	dst := filepath.Join(tmpDir, "out")

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return getTestFolderMetadata(arg.Path), nil
		},
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					getTestFolderMetadata("/remote"),
					getTestFolderMetadata("/remote/empty"),
					getTestFileMetadata("/remote/file.txt", 4),
				},
				HasMore: false,
			}, nil
		},
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			return getTestFileMetadata(arg.Path, 4), io.NopCloser(strings.NewReader("data")), nil
		},
	}
	stubFilesClient(t, mock)

	var stdout, stderr bytes.Buffer
	cmd := testGetJSONCmd(&stdout, &stderr)
	if err := cmd.Flags().Set("recursive", "true"); err != nil {
		t.Fatal(err)
	}
	if err := get(cmd, []string{"/remote", dst}); err != nil {
		t.Fatalf("get error: %v", err)
	}

	got := decodeGetOutput(t, &stdout)
	if got.Input.Source != "/remote" || got.Input.Target != dst || !got.Input.Recursive || got.Input.Stdout {
		t.Fatalf("input = %+v", got.Input)
	}
	want := map[string]string{
		dst:                            getStatusCreated,
		filepath.Join(dst, "empty"):    getStatusCreated,
		filepath.Join(dst, "file.txt"): getStatusDownloaded,
	}
	if len(got.Results) != len(want) {
		t.Fatalf("results len = %d, want %d: %+v", len(got.Results), len(want), got.Results)
	}
	for _, result := range got.Results {
		wantStatus, ok := want[result.Input.Target]
		if !ok {
			t.Fatalf("unexpected target %q", result.Input.Target)
		}
		if result.Status != wantStatus {
			t.Fatalf("status for %s = %s, want %s", result.Input.Target, result.Status, wantStatus)
		}
		if result.Result == nil {
			t.Fatalf("result metadata for %s is nil", result.Input.Target)
		}
	}
	if !strings.Contains(stderr.String(), "Downloading /remote/file.txt") {
		t.Fatalf("stderr = %q, want recursive download status", stderr.String())
	}
}

func TestGetJSONRecursiveErrorEmitsNoSuccessJSON(t *testing.T) {
	tmpDir := t.TempDir()
	dst := filepath.Join(tmpDir, "out")

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return getTestFolderMetadata(arg.Path), nil
		},
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					getTestFileMetadata("/remote/good.txt", 4),
					getTestFileMetadata("/remote/bad.txt", 4),
				},
				HasMore: false,
			}, nil
		},
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			if strings.Contains(arg.Path, "bad.txt") {
				return nil, nil, &files.DownloadAPIError{}
			}
			return getTestFileMetadata(arg.Path, 4), io.NopCloser(strings.NewReader("data")), nil
		},
	}
	stubFilesClient(t, mock)

	var stdout bytes.Buffer
	cmd := testGetJSONCmd(&stdout, nil)
	if err := cmd.Flags().Set("recursive", "true"); err != nil {
		t.Fatal(err)
	}
	err := get(cmd, []string{"/remote", dst})
	if err == nil {
		t.Fatal("expected recursive error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty on recursive error", stdout.String())
	}
	if _, statErr := os.Stat(filepath.Join(dst, "good.txt")); statErr != nil {
		t.Fatalf("good file was not downloaded before failure: %v", statErr)
	}
}

func TestGetRecursive_AppendsSourceBaseWhenDstIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	var listPath string
	var downloadPath string

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return &files.FolderMetadata{Metadata: files.Metadata{PathDisplay: arg.Path}}, nil
		},
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			listPath = arg.Path
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/remote/folder/file.txt"}, Size: 4},
				},
				HasMore: false,
			}, nil
		},
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			downloadPath = arg.Path
			return &files.FileMetadata{
				Metadata: files.Metadata{PathDisplay: arg.Path},
				Size:     4,
			}, io.NopCloser(strings.NewReader("data")), nil
		},
	}
	stubFilesClient(t, mock)

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("recursive", "r", true, "")

	err := get(cmd, []string{"/remote/folder", tmpDir})
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if listPath != "/remote/folder" {
		t.Errorf("ListFolder path = %q, want /remote/folder", listPath)
	}
	if downloadPath != "/remote/folder/file.txt" {
		t.Errorf("Download path = %q, want /remote/folder/file.txt", downloadPath)
	}
	got, err := os.ReadFile(filepath.Join(tmpDir, "folder", "file.txt"))
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(got) != "data" {
		t.Errorf("downloaded file = %q, want data", got)
	}
}

func TestRelativeTo(t *testing.T) {
	tests := []struct {
		base, full, want string
	}{
		{"/remote", "/remote", ""},
		{"/remote", "/remote/file.txt", "file.txt"},
		{"/remote", "/remote/sub/deep.txt", "sub/deep.txt"},
		{"/Remote", "/remote/file.txt", "file.txt"},
		{"/remote", "/Remote/File.TXT", "File.TXT"},
	}
	for _, tt := range tests {
		got, err := relativeTo(tt.base, tt.full)
		if err != nil {
			t.Errorf("relativeTo(%q, %q) error: %v", tt.base, tt.full, err)
			continue
		}
		if got != tt.want {
			t.Errorf("relativeTo(%q, %q) = %q, want %q", tt.base, tt.full, got, tt.want)
		}
	}
}

func TestRelativeToRejectsSiblingPrefix(t *testing.T) {
	_, err := relativeTo("/remote", "/remote2/file.txt")
	if err == nil {
		t.Fatal("expected error for sibling path with shared prefix")
	}
}
