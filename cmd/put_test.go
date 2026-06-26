package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	dbxauth "github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/auth"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

func TestUploadOneChunk_Success(t *testing.T) {
	mock := &mockFilesClient{
		uploadSessionAppendV2Fn: func(arg *files.UploadSessionAppendArg, content io.Reader) error {
			return nil
		},
	}

	cursor := files.NewUploadSessionCursor("session123", 0)
	args := files.NewUploadSessionAppendArg(cursor)
	data := []byte("hello chunk")

	err := uploadOneChunk(mock, args, data)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestUploadOneChunk_RetryOnServerError(t *testing.T) {
	stubRetrySleep(t)
	calls := 0
	mock := &mockFilesClient{
		uploadSessionAppendV2Fn: func(arg *files.UploadSessionAppendArg, content io.Reader) error {
			calls++
			if calls < 3 {
				return dbxauth.ServerError{APIError: dropbox.APIError{ErrorSummary: "500"}}
			}
			return nil
		},
	}

	cursor := files.NewUploadSessionCursor("session123", 0)
	args := files.NewUploadSessionAppendArg(cursor)
	data := []byte("hello chunk")

	err := uploadOneChunk(mock, args, data)
	if err != nil {
		t.Errorf("expected nil error after retry, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestUploadOneChunk_RateLimitUsesSingleRetryAfterDelay(t *testing.T) {
	delays := stubRetrySleep(t)
	rateLimit := dbxauth.NewRateLimitError(nil)
	rateLimit.RetryAfter = 4

	calls := 0
	mock := &mockFilesClient{
		uploadSessionAppendV2Fn: func(arg *files.UploadSessionAppendArg, content io.Reader) error {
			calls++
			if calls == 1 {
				return dbxauth.RateLimitAPIError{RateLimitError: rateLimit}
			}
			return nil
		},
	}

	cursor := files.NewUploadSessionCursor("session123", 0)
	args := files.NewUploadSessionAppendArg(cursor)
	data := []byte("hello chunk")

	err := uploadOneChunk(mock, args, data)
	if err != nil {
		t.Errorf("expected nil error after rate-limit retry, got %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
	if len(*delays) != 1 {
		t.Fatalf("expected 1 sleep, got %d", len(*delays))
	}
	if (*delays)[0] != 4*time.Second {
		t.Errorf("sleep = %v, want %v", (*delays)[0], 4*time.Second)
	}
}

func appendIncorrectOffsetError(correctOffset uint64) files.UploadSessionAppendV2APIError {
	return files.UploadSessionAppendV2APIError{
		EndpointError: &files.UploadSessionAppendError{
			Tagged:          dropbox.Tagged{Tag: files.UploadSessionAppendErrorIncorrectOffset},
			IncorrectOffset: files.NewUploadSessionOffsetError(correctOffset),
		},
	}
}

func TestUploadOneChunk_TreatsExpectedIncorrectOffsetAsSuccess(t *testing.T) {
	stubRetrySleep(t)
	calls := 0
	data := []byte("hello chunk")
	cursor := files.NewUploadSessionCursor("session123", 64)
	args := files.NewUploadSessionAppendArg(cursor)
	expectedOffset := cursor.Offset + uint64(len(data))

	mock := &mockFilesClient{
		uploadSessionAppendV2Fn: func(arg *files.UploadSessionAppendArg, content io.Reader) error {
			calls++
			if calls == 1 {
				return dbxauth.ServerError{APIError: dropbox.APIError{ErrorSummary: "500"}}
			}
			return appendIncorrectOffsetError(expectedOffset)
		},
	}

	err := uploadOneChunk(mock, args, data)
	if err != nil {
		t.Errorf("expected nil error for already-accepted chunk, got %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestUploadOneChunk_ReturnsUnexpectedIncorrectOffset(t *testing.T) {
	data := []byte("hello chunk")
	cursor := files.NewUploadSessionCursor("session123", 64)
	args := files.NewUploadSessionAppendArg(cursor)
	unexpectedOffset := cursor.Offset + uint64(len(data)) - 1

	mock := &mockFilesClient{
		uploadSessionAppendV2Fn: func(arg *files.UploadSessionAppendArg, content io.Reader) error {
			return appendIncorrectOffsetError(unexpectedOffset)
		},
	}

	err := uploadOneChunk(mock, args, data)
	if err == nil {
		t.Error("expected unexpected incorrect_offset error, got nil")
	}
}

func TestUploadOneChunk_PermanentError(t *testing.T) {
	mock := &mockFilesClient{
		uploadSessionAppendV2Fn: func(arg *files.UploadSessionAppendArg, content io.Reader) error {
			return &files.UploadSessionAppendAPIError{}
		},
	}

	cursor := files.NewUploadSessionCursor("session123", 0)
	args := files.NewUploadSessionAppendArg(cursor)
	data := []byte("hello chunk")

	err := uploadOneChunk(mock, args, data)
	if err == nil {
		t.Error("expected error for permanent failure, got nil")
	}
}

func TestUploadSingleShot_RetryOnServerErrorResetsReader(t *testing.T) {
	stubRetrySleep(t)
	data := []byte("small file contents")
	reader := bytes.NewReader(data)
	uploadArg := &files.UploadArg{CommitInfo: *files.NewCommitInfo("/test.txt")}

	calls := 0
	var bodies []string
	mock := &mockFilesClient{
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			calls++
			body, err := io.ReadAll(content)
			if err != nil {
				return nil, err
			}
			bodies = append(bodies, string(body))
			if calls < 3 {
				return nil, dbxauth.ServerError{APIError: dropbox.APIError{ErrorSummary: "500"}}
			}
			return &files.FileMetadata{}, nil
		},
	}

	_, err := uploadSingleShot(mock, reader, uploadArg, int64(len(data)), nil)
	if err != nil {
		t.Errorf("expected nil error after retry, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
	if len(bodies) != 3 {
		t.Fatalf("expected 3 uploaded bodies, got %d", len(bodies))
	}
	for i, body := range bodies {
		if body != string(data) {
			t.Errorf("body %d = %q, want %q", i, body, string(data))
		}
	}
}

func TestUploadSingleShot_RetriesTooManyWriteOperations(t *testing.T) {
	stubRetrySleep(t)
	data := []byte("small file contents")
	reader := bytes.NewReader(data)
	uploadArg := &files.UploadArg{CommitInfo: *files.NewCommitInfo("/test.txt")}

	calls := 0
	mock := &mockFilesClient{
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			calls++
			if _, err := io.ReadAll(content); err != nil {
				return nil, err
			}
			if calls == 1 {
				return nil, uploadWriteThrottleError()
			}
			return &files.FileMetadata{}, nil
		},
	}

	_, err := uploadSingleShot(mock, reader, uploadArg, int64(len(data)), nil)
	if err != nil {
		t.Errorf("expected nil error after write-throttle retry, got %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func testConfig() dropbox.Config {
	return dropbox.Config{Token: "test-token"}
}

func testPutCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().BoolP("recursive", "r", false, "Recursively upload directories")
	cmd.Flags().IntP("workers", "w", 4, "Number of concurrent upload workers to use")
	cmd.Flags().Int64P("chunksize", "c", 1<<24, "Chunk size to use (should be multiple of 4MiB)")
	cmd.Flags().BoolP("debug", "d", false, "Print debug timing")
	cmd.Flags().String("if-exists", putIfExistsOverwrite, "What to do when the destination file exists: overwrite, skip, or fail")
	return cmd
}

func testPutJSONCmd(stdout, stderr *bytes.Buffer) *cobra.Command {
	cmd := testPutCmd()
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

type putOutputData struct {
	Input    putCommandInput `json:"input"`
	Results  []putResult     `json:"results"`
	Warnings []jsonWarning   `json:"warnings"`
}

func decodePutOutput(t *testing.T, stdout *bytes.Buffer) putOutputData {
	t.Helper()

	got := decodePutOutputWithWarnings(t, stdout)
	if len(got.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want empty", got.Warnings)
	}
	return got
}

func decodePutOutputWithWarnings(t *testing.T, stdout *bytes.Buffer) putOutputData {
	t.Helper()

	var got putOutputData
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode put JSON output: %v\noutput: %s", err, stdout.String())
	}
	if got.Warnings == nil {
		t.Fatalf("warnings = nil, want empty array")
	}
	return got
}

func putFileMetadata(path string, size uint64) *files.FileMetadata {
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

func putFolderMetadata(path string) *files.FolderMetadata {
	return &files.FolderMetadata{
		Metadata: files.Metadata{
			Name:        strings.TrimPrefix(path, "/"),
			PathDisplay: path,
			PathLower:   strings.ToLower(path),
		},
		Id: "id:" + strings.TrimPrefix(path, "/"),
	}
}

func getMetadataNotFoundError() files.GetMetadataAPIError {
	return files.GetMetadataAPIError{
		EndpointError: &files.GetMetadataError{
			Tagged: dropbox.Tagged{Tag: files.GetMetadataErrorPath},
			Path: &files.LookupError{
				Tagged: dropbox.Tagged{Tag: files.LookupErrorNotFound},
			},
		},
	}
}

func writeConflictError(conflictTag string) *files.WriteError {
	return &files.WriteError{
		Tagged: dropbox.Tagged{Tag: files.WriteErrorConflict},
		Conflict: &files.WriteConflictError{
			Tagged: dropbox.Tagged{Tag: conflictTag},
		},
	}
}

func uploadPathConflictError(conflictTag string) files.UploadAPIError {
	return files.UploadAPIError{
		EndpointError: &files.UploadError{
			Tagged: dropbox.Tagged{Tag: files.UploadErrorPath},
			Path: files.NewUploadWriteFailed(
				writeConflictError(conflictTag),
				"session123",
			),
		},
	}
}

func uploadSessionFinishPathConflictError(conflictTag string) files.UploadSessionFinishAPIError {
	return files.UploadSessionFinishAPIError{
		EndpointError: &files.UploadSessionFinishError{
			Tagged: dropbox.Tagged{Tag: files.UploadSessionFinishErrorPath},
			Path:   writeConflictError(conflictTag),
		},
	}
}

func uploadPathConflictErrorPtr(conflictTag string) *files.UploadAPIError {
	err := uploadPathConflictError(conflictTag)
	return &err
}

func uploadSessionFinishPathConflictErrorPtr(conflictTag string) *files.UploadSessionFinishAPIError {
	err := uploadSessionFinishPathConflictError(conflictTag)
	return &err
}

func TestResolveDestination_TargetIsFolder(t *testing.T) {
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return &files.FolderMetadata{Metadata: files.Metadata{PathDisplay: arg.Path}}, nil
		},
	}
	got := resolveDestination(mock, "/local/video.mp4", "/Videos", false)
	if got != "/Videos/video.mp4" {
		t.Errorf("resolveDestination = %q, want /Videos/video.mp4", got)
	}
}

func TestResolveDestination_TargetIsFile(t *testing.T) {
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return &files.FileMetadata{}, nil
		},
	}
	got := resolveDestination(mock, "/local/a.txt", "/existing.txt", false)
	if got != "/existing.txt" {
		t.Errorf("resolveDestination = %q, want /existing.txt", got)
	}
}

func TestResolveDestination_TargetNotFound(t *testing.T) {
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, fmt.Errorf("path/not_found/")
		},
	}
	got := resolveDestination(mock, "/local/a.txt", "/new-name", false)
	if got != "/new-name" {
		t.Errorf("resolveDestination = %q, want /new-name", got)
	}
}

func TestResolveDestination_ExplicitDirectory(t *testing.T) {
	mock := &mockFilesClient{}
	got := resolveDestination(mock, "/local/a.txt", "/folder", true)
	if got != "/folder/a.txt" {
		t.Errorf("resolveDestination = %q, want /folder/a.txt", got)
	}
}

func TestResolveDestination_ExplicitRootDirectory(t *testing.T) {
	mock := &mockFilesClient{}
	got := resolveDestination(mock, "/local/a.txt", "", true)
	if got != "/a.txt" {
		t.Errorf("resolveDestination = %q, want /a.txt", got)
	}
}

func TestResolveDestination_MetadataError(t *testing.T) {
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, fmt.Errorf("server error 500")
		},
	}
	got := resolveDestination(mock, "/local/a.txt", "/path", false)
	if got != "/path" {
		t.Errorf("resolveDestination = %q, want /path (fallback on error)", got)
	}
}

func TestPutFileDestinationTrailingSlash(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "a.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	var uploadedPath string
	origConfig := config
	defer func() { config = origConfig }()

	config = testConfig()
	testClient := &mockFilesClient{
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			uploadedPath = arg.Path
			return &files.FileMetadata{}, nil
		},
	}
	origNew := filesNewFunc
	filesNewFunc = func(_ dropbox.Config) files.Client { return testClient }
	defer func() { filesNewFunc = origNew }()

	err := put(testPutCmd(), []string{tmpFile, "/folder/"})
	if err != nil {
		t.Fatalf("put error: %v", err)
	}
	if uploadedPath != "/folder/a.txt" {
		t.Errorf("uploaded path = %q, want /folder/a.txt", uploadedPath)
	}
}

func TestPutArgValidation(t *testing.T) {
	err := put(testPutCmd(), []string{})
	if err == nil {
		t.Error("expected error for no args")
	}
}

func TestPutDirectoryWithoutRecursiveFlag(t *testing.T) {
	dir := t.TempDir()
	err := put(testPutCmd(), []string{dir, "/dest"})
	if err == nil {
		t.Fatal("expected error when putting directory without --recursive")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("use --recursive")) {
		t.Errorf("error = %q, want mention of --recursive", err.Error())
	}
}

func TestPutRecursive_WalksDirectoryStructure(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "sub", "deep"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "root.txt"), []byte("root"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "mid.txt"), []byte("mid"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "deep", "leaf.txt"), []byte("leaf"), 0644); err != nil {
		t.Fatal(err)
	}

	var uploaded []string
	origConfig := config
	defer func() { config = origConfig }()

	config = testConfig()
	testClient := &mockFilesClient{
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			uploaded = append(uploaded, arg.Path)
			return &files.FileMetadata{}, nil
		},
	}
	origNew := filesNewFunc
	filesNewFunc = func(_ dropbox.Config) files.Client { return testClient }
	defer func() { filesNewFunc = origNew }()

	opts := putOptions{chunkSize: 1 << 24, workers: 4}
	err := putRecursive(dir, "/backup", opts)
	if err != nil {
		t.Fatalf("putRecursive error: %v", err)
	}

	expected := map[string]bool{
		"/backup/root.txt":          true,
		"/backup/sub/mid.txt":       true,
		"/backup/sub/deep/leaf.txt": true,
	}
	if len(uploaded) != len(expected) {
		t.Fatalf("uploaded %d files, want %d: %v", len(uploaded), len(expected), uploaded)
	}
	for _, path := range uploaded {
		if !expected[path] {
			t.Errorf("unexpected upload path: %s", path)
		}
	}
}

func TestPutRecursive_CreatesEmptyDirectories(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "has-files"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "empty"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "empty", "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "has-files", "a.txt"), []byte("a"), 0644); err != nil {
		t.Fatal(err)
	}

	var uploaded []string
	var createdDirs []string
	origConfig := config
	defer func() { config = origConfig }()

	config = testConfig()
	testClient := &mockFilesClient{
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			uploaded = append(uploaded, arg.Path)
			return &files.FileMetadata{}, nil
		},
		createFolderV2Fn: func(arg *files.CreateFolderArg) (*files.CreateFolderResult, error) {
			createdDirs = append(createdDirs, arg.Path)
			return &files.CreateFolderResult{}, nil
		},
	}
	origNew := filesNewFunc
	filesNewFunc = func(_ dropbox.Config) files.Client { return testClient }
	defer func() { filesNewFunc = origNew }()

	opts := putOptions{chunkSize: 1 << 24, workers: 4}
	err := putRecursive(dir, "/dest", opts)
	if err != nil {
		t.Fatalf("putRecursive error: %v", err)
	}

	if len(uploaded) != 1 || uploaded[0] != "/dest/has-files/a.txt" {
		t.Errorf("uploaded = %v, want [/dest/has-files/a.txt]", uploaded)
	}

	expectedDirs := map[string]bool{
		"/dest":              true,
		"/dest/empty":        true,
		"/dest/empty/nested": true,
	}
	if len(createdDirs) != len(expectedDirs) {
		t.Fatalf("created %d dirs, want %d: %v", len(createdDirs), len(expectedDirs), createdDirs)
	}
	for _, d := range createdDirs {
		if !expectedDirs[d] {
			t.Errorf("unexpected created dir: %s", d)
		}
	}
}

func TestPutRecursive_CreatesEmptyRootDirectory(t *testing.T) {
	dir := t.TempDir()

	var uploaded []string
	var createdDirs []string
	origConfig := config
	defer func() { config = origConfig }()

	config = testConfig()
	testClient := &mockFilesClient{
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			uploaded = append(uploaded, arg.Path)
			return &files.FileMetadata{}, nil
		},
		createFolderV2Fn: func(arg *files.CreateFolderArg) (*files.CreateFolderResult, error) {
			createdDirs = append(createdDirs, arg.Path)
			return &files.CreateFolderResult{}, nil
		},
	}
	origNew := filesNewFunc
	filesNewFunc = func(_ dropbox.Config) files.Client { return testClient }
	defer func() { filesNewFunc = origNew }()

	opts := putOptions{chunkSize: 1 << 24, workers: 4}
	err := putRecursive(dir, "/empty-root", opts)
	if err != nil {
		t.Fatalf("putRecursive error: %v", err)
	}

	if len(uploaded) != 0 {
		t.Fatalf("uploaded = %v, want no files", uploaded)
	}
	if len(createdDirs) != 1 || createdDirs[0] != "/empty-root" {
		t.Fatalf("createdDirs = %v, want [/empty-root]", createdDirs)
	}
}

func TestPutRecursive_SkipsSymlinks(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "real.txt"), []byte("real"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "real.txt"), filepath.Join(dir, "link.txt")); err != nil {
		t.Fatal(err)
	}

	var uploaded []string
	origConfig := config
	defer func() { config = origConfig }()

	config = testConfig()
	testClient := &mockFilesClient{
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			uploaded = append(uploaded, arg.Path)
			return &files.FileMetadata{}, nil
		},
	}
	origNew := filesNewFunc
	filesNewFunc = func(_ dropbox.Config) files.Client { return testClient }
	defer func() { filesNewFunc = origNew }()

	opts := putOptions{chunkSize: 1 << 24, workers: 4}
	err := putRecursive(dir, "/dest", opts)
	if err != nil {
		t.Fatalf("putRecursive error: %v", err)
	}

	if len(uploaded) != 1 {
		t.Fatalf("uploaded %d files, want 1: %v", len(uploaded), uploaded)
	}
	if uploaded[0] != "/dest/real.txt" {
		t.Errorf("uploaded[0] = %q, want /dest/real.txt", uploaded[0])
	}
}

func TestPutChunkSizeValidation(t *testing.T) {
	tests := []struct {
		name      string
		chunkSize string
		want      string
	}{
		{
			name:      "below 4MiB",
			chunkSize: "100",
			want:      "`put` requires chunk size to be at least 4MiB",
		},
		{
			name:      "not multiple of 4MiB",
			chunkSize: "6291456",
			want:      "`put` requires chunk size to be a multiple of 4MiB",
		},
		{
			name:      "above Dropbox request limit",
			chunkSize: "268435456",
			want:      "`put` requires chunk size to be no more than 128MiB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := testPutCmd()
			_ = cmd.Flags().Set("chunksize", tt.chunkSize)
			_, err := parsePutOptions(cmd)
			if err == nil || err.Error() != tt.want {
				t.Errorf("expected chunk size validation error %q, got %v", tt.want, err)
			}
		})
	}
}

func TestPutIfExistsValidation(t *testing.T) {
	cmd := testPutCmd()
	_ = cmd.Flags().Set("if-exists", "replace")

	_, err := parsePutOptions(cmd)
	if err == nil {
		t.Fatal("expected invalid --if-exists error")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("overwrite, skip, or fail")) {
		t.Errorf("error = %q, want valid option list", err.Error())
	}
}

func TestPutJSONSingleFileOutputsUploadedResult(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	mock := &mockFilesClient{
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			if arg.Path != "/uploaded.txt" {
				t.Fatalf("upload path = %q, want /uploaded.txt", arg.Path)
			}
			if _, err := io.ReadAll(content); err != nil {
				t.Fatal(err)
			}
			return putFileMetadata(arg.Path, 4), nil
		},
	}
	stubFilesClient(t, mock)

	var stdout, stderr bytes.Buffer
	cmd := testPutJSONCmd(&stdout, &stderr)
	if err := put(cmd, []string{tmpFile, "/uploaded.txt"}); err != nil {
		t.Fatalf("put error: %v", err)
	}

	got := decodePutOutput(t, &stdout)
	if got.Input.Source != tmpFile || got.Input.Target != "/uploaded.txt" || got.Input.Recursive || got.Input.Stdin {
		t.Fatalf("input = %+v", got.Input)
	}
	if got.Input.IfExists != putIfExistsOverwrite {
		t.Fatalf("if_exists = %q", got.Input.IfExists)
	}
	if len(got.Results) != 1 {
		t.Fatalf("results len = %d, want 1", len(got.Results))
	}
	result := got.Results[0]
	if result.Status != putStatusUploaded || result.Kind != putKindFile {
		t.Fatalf("result status/kind = %s/%s", result.Status, result.Kind)
	}
	if result.Input.Source != tmpFile || result.Input.Target != "/uploaded.txt" {
		t.Fatalf("result input = %+v", result.Input)
	}
	if result.Result == nil {
		t.Fatal("result metadata is nil")
	}
	if result.Result.Type != "file" || result.Result.PathDisplay != "/uploaded.txt" || result.Result.PathLower != "/uploaded.txt" {
		t.Fatalf("metadata = %+v", result.Result)
	}
	if result.Result.Size == nil || *result.Result.Size != 4 {
		t.Fatalf("size = %v, want 4", result.Result.Size)
	}
}

func TestPutJSONDefaultTargetUsesComputedPath(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "local.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	mock := &mockFilesClient{
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			if arg.Path != "/local.txt" {
				t.Fatalf("upload path = %q, want /local.txt", arg.Path)
			}
			if _, err := io.ReadAll(content); err != nil {
				t.Fatal(err)
			}
			return putFileMetadata(arg.Path, 4), nil
		},
	}
	stubFilesClient(t, mock)

	var stdout bytes.Buffer
	cmd := testPutJSONCmd(&stdout, nil)
	if err := put(cmd, []string{tmpFile}); err != nil {
		t.Fatalf("put error: %v", err)
	}

	got := decodePutOutput(t, &stdout)
	if got.Input.Target != "/local.txt" {
		t.Fatalf("target = %q, want /local.txt", got.Input.Target)
	}
	if got.Results[0].Input.Target != "/local.txt" {
		t.Fatalf("result target = %q, want /local.txt", got.Results[0].Input.Target)
	}
}

func TestPutJSONIfExistsSkipOutputsSkippedResult(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return putFileMetadata(arg.Path, 12), nil
		},
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			t.Fatal("upload should not be called for skipped destination")
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	var stdout, stderr bytes.Buffer
	cmd := testPutJSONCmd(&stdout, &stderr)
	if err := cmd.Flags().Set("if-exists", putIfExistsSkip); err != nil {
		t.Fatal(err)
	}
	if err := put(cmd, []string{tmpFile, "/existing.txt"}); err != nil {
		t.Fatalf("put error: %v", err)
	}

	got := decodePutOutput(t, &stdout)
	if len(got.Results) != 1 {
		t.Fatalf("results len = %d, want 1", len(got.Results))
	}
	result := got.Results[0]
	if result.Status != putStatusSkipped || result.Kind != putKindFile {
		t.Fatalf("result status/kind = %s/%s", result.Status, result.Kind)
	}
	if result.Result == nil || result.Result.PathDisplay != "/existing.txt" {
		t.Fatalf("metadata = %+v", result.Result)
	}
	if !strings.Contains(stderr.String(), "Skipping /existing.txt") {
		t.Fatalf("stderr = %q, want skip status", stderr.String())
	}
}

func TestPutJSONIfExistsFailReturnsErrorWithoutStdout(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return putFileMetadata(arg.Path, 12), nil
		},
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			t.Fatal("upload should not be called for failed destination")
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	var stdout bytes.Buffer
	cmd := testPutJSONCmd(&stdout, nil)
	if err := cmd.Flags().Set("if-exists", putIfExistsFail); err != nil {
		t.Fatal(err)
	}
	err := put(cmd, []string{tmpFile, "/existing.txt"})
	if err == nil {
		t.Fatal("expected error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty on error", stdout.String())
	}
}

func TestPutJSONStdinDoesNotExposeTempPath(t *testing.T) {
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, getMetadataNotFoundError()
		},
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			body, err := io.ReadAll(content)
			if err != nil {
				t.Fatal(err)
			}
			if string(body) != "hello" {
				t.Fatalf("uploaded body = %q, want hello", string(body))
			}
			return putFileMetadata(arg.Path, uint64(len(body))), nil
		},
	}
	stubFilesClient(t, mock)

	var stdout bytes.Buffer
	cmd := testPutJSONCmd(&stdout, nil)
	cmd.SetIn(strings.NewReader("hello"))
	if err := put(cmd, []string{"-", "/stdin.txt"}); err != nil {
		t.Fatalf("put error: %v", err)
	}

	output := stdout.String()
	if strings.Contains(output, "dbxcli-stdin") {
		t.Fatalf("stdout exposes temp path: %s", output)
	}
	got := decodePutOutput(t, &stdout)
	if got.Input.Source != "-" || !got.Input.Stdin || got.Input.Target != "/stdin.txt" {
		t.Fatalf("input = %+v", got.Input)
	}
	if got.Results[0].Input.Source != "-" || got.Results[0].Input.Target != "/stdin.txt" {
		t.Fatalf("result input = %+v", got.Results[0].Input)
	}
}

func TestPutJSONRecursiveOutputsDirectoryAndFileResults(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "empty"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("file"), 0644); err != nil {
		t.Fatal(err)
	}

	mock := &mockFilesClient{
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			if _, err := io.ReadAll(content); err != nil {
				t.Fatal(err)
			}
			return putFileMetadata(arg.Path, 4), nil
		},
		createFolderV2Fn: func(arg *files.CreateFolderArg) (*files.CreateFolderResult, error) {
			return files.NewCreateFolderResult(putFolderMetadata(arg.Path)), nil
		},
	}
	stubFilesClient(t, mock)

	var stdout, stderr bytes.Buffer
	cmd := testPutJSONCmd(&stdout, &stderr)
	if err := cmd.Flags().Set("recursive", "true"); err != nil {
		t.Fatal(err)
	}
	if err := put(cmd, []string{dir, "/remote"}); err != nil {
		t.Fatalf("put error: %v", err)
	}

	got := decodePutOutput(t, &stdout)
	if !got.Input.Recursive || got.Input.Source != dir || got.Input.Target != "/remote" {
		t.Fatalf("input = %+v", got.Input)
	}
	want := map[string]string{
		"/remote/file.txt": putStatusUploaded,
		"/remote":          putStatusCreated,
		"/remote/empty":    putStatusCreated,
	}
	if len(got.Results) != len(want) {
		t.Fatalf("results len = %d, want %d: %+v", len(got.Results), len(want), got.Results)
	}
	for _, result := range got.Results {
		wantStatus, ok := want[result.Input.Target]
		if !ok {
			t.Fatalf("unexpected result target %q", result.Input.Target)
		}
		if result.Status != wantStatus {
			t.Fatalf("status for %s = %s, want %s", result.Input.Target, result.Status, wantStatus)
		}
		if result.Result == nil {
			t.Fatalf("result metadata for %s is nil", result.Input.Target)
		}
	}
	if !strings.Contains(stderr.String(), "Processing ") || !strings.Contains(stderr.String(), "Creating directory ") {
		t.Fatalf("stderr = %q, want recursive status", stderr.String())
	}
	if !json.Valid(stdout.Bytes()) {
		t.Fatalf("stdout is not valid JSON: %s", stdout.String())
	}
}

func TestPutJSONRecursiveWarnsForSkippedSymlink(t *testing.T) {
	dir := t.TempDir()
	realPath := filepath.Join(dir, "real.txt")
	linkPath := filepath.Join(dir, "link.txt")
	if err := os.WriteFile(realPath, []byte("real"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realPath, linkPath); err != nil {
		t.Fatal(err)
	}

	mock := &mockFilesClient{
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			if arg.Path != "/remote/real.txt" {
				t.Fatalf("upload path = %q, want /remote/real.txt", arg.Path)
			}
			if _, err := io.ReadAll(content); err != nil {
				t.Fatal(err)
			}
			return putFileMetadata(arg.Path, 4), nil
		},
		createFolderV2Fn: func(arg *files.CreateFolderArg) (*files.CreateFolderResult, error) {
			return files.NewCreateFolderResult(putFolderMetadata(arg.Path)), nil
		},
	}
	stubFilesClient(t, mock)

	var stdout bytes.Buffer
	cmd := testPutJSONCmd(&stdout, nil)
	if err := cmd.Flags().Set("recursive", "true"); err != nil {
		t.Fatal(err)
	}
	if err := put(cmd, []string{dir, "/remote"}); err != nil {
		t.Fatalf("put error: %v", err)
	}

	got := decodePutOutputWithWarnings(t, &stdout)
	if len(got.Warnings) != 1 {
		t.Fatalf("warnings = %+v, want one skipped symlink warning", got.Warnings)
	}
	warning := got.Warnings[0]
	if warning.Code != jsonWarningCodeSkippedSymlink || warning.Path != linkPath {
		t.Fatalf("warning = %+v, want skipped symlink at %s", warning, linkPath)
	}
	if !strings.Contains(warning.Message, "skipped symlink") {
		t.Fatalf("warning message = %q, want skipped symlink", warning.Message)
	}
	for _, result := range got.Results {
		if result.Input.Source == linkPath || result.Input.Target == "/remote/link.txt" {
			t.Fatalf("symlink should not be uploaded: %+v", result)
		}
	}
	if !json.Valid(stdout.Bytes()) {
		t.Fatalf("stdout is not valid JSON: %s", stdout.String())
	}
}

func TestPutJSONRecursiveFailsWhenDirectoryTargetIsExistingFile(t *testing.T) {
	dir := t.TempDir()

	mock := &mockFilesClient{
		createFolderV2Fn: func(arg *files.CreateFolderArg) (*files.CreateFolderResult, error) {
			return nil, createFolderConflictError(files.WriteConflictErrorFile)
		},
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			t.Fatal("GetMetadata should not be called for typed file conflicts")
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	var stdout bytes.Buffer
	cmd := testPutJSONCmd(&stdout, nil)
	if err := cmd.Flags().Set("recursive", "true"); err != nil {
		t.Fatal(err)
	}
	err := put(cmd, []string{dir, "/remote-file"})
	if err == nil {
		t.Fatal("expected existing file conflict")
	}
	if !strings.Contains(err.Error(), "path exists and is not a folder: /remote-file") {
		t.Fatalf("error = %q, want not-a-folder error", err.Error())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty on error", stdout.String())
	}
}

func TestPutRecursiveTextModeDoesNotFetchExistingFolderMetadata(t *testing.T) {
	dir := t.TempDir()

	mock := &mockFilesClient{
		createFolderV2Fn: func(arg *files.CreateFolderArg) (*files.CreateFolderResult, error) {
			return nil, createFolderConflictError(files.WriteConflictErrorFolder)
		},
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			t.Fatal("GetMetadata should not be called in text mode for existing folders")
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	cmd := testPutCmd()
	if err := cmd.Flags().Set("recursive", "true"); err != nil {
		t.Fatal(err)
	}
	err := put(cmd, []string{dir, "/existing-folder"})
	if err != nil {
		t.Fatalf("put error: %v", err)
	}
}

func TestPutTextModeWritesNoStdoutOnSuccess(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	mock := &mockFilesClient{
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			if _, err := io.ReadAll(content); err != nil {
				t.Fatal(err)
			}
			return putFileMetadata(arg.Path, 4), nil
		},
	}
	stubFilesClient(t, mock)

	var stdout bytes.Buffer
	cmd := testPutCmd()
	cmd.SetOut(&stdout)
	if err := put(cmd, []string{tmpFile, "/text.txt"}); err != nil {
		t.Fatalf("put error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestPutFileIfExistsSkipSkipsExistingFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return &files.FileMetadata{Metadata: files.Metadata{PathDisplay: arg.Path}}, nil
		},
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			t.Fatal("upload should not be called for existing destination with --if-exists skip")
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	stderr := captureStderr(t, func() {
		err := putFile(tmpFile, "/existing.txt", putOptions{
			chunkSize: 1 << 24,
			workers:   4,
			ifExists:  putIfExistsSkip,
		})
		if err != nil {
			t.Fatalf("putFile error: %v", err)
		}
	})
	if !bytes.Contains([]byte(stderr), []byte("Skipping /existing.txt")) {
		t.Errorf("stderr = %q, want skip message", stderr)
	}
}

func TestPutFileIfExistsFailFailsExistingFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return &files.FileMetadata{Metadata: files.Metadata{PathDisplay: arg.Path}}, nil
		},
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			t.Fatal("upload should not be called for existing destination with --if-exists fail")
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	err := putFile(tmpFile, "/existing.txt", putOptions{
		chunkSize: 1 << 24,
		workers:   4,
		ifExists:  putIfExistsFail,
	})
	if err == nil {
		t.Fatal("expected existing destination error")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("already exists")) {
		t.Errorf("error = %q, want already exists", err.Error())
	}
}

func TestPutFileIfExistsFailUploadsMissingFileWithAddMode(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	var mode string
	var strictConflict bool
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, getMetadataNotFoundError()
		},
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			mode = arg.Mode.Tag
			strictConflict = arg.StrictConflict
			return &files.FileMetadata{}, nil
		},
	}
	stubFilesClient(t, mock)

	err := putFile(tmpFile, "/new.txt", putOptions{
		chunkSize: 1 << 24,
		workers:   4,
		ifExists:  putIfExistsFail,
	})
	if err != nil {
		t.Fatalf("putFile error: %v", err)
	}
	if mode != files.WriteModeAdd {
		t.Errorf("write mode = %q, want %q", mode, files.WriteModeAdd)
	}
	if !strictConflict {
		t.Error("strict conflict = false, want true")
	}
}

func TestPutFileIfExistsOverwriteUsesOverwriteMode(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	metadataCalls := 0
	var mode string
	var strictConflict bool
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			metadataCalls++
			return &files.FileMetadata{}, nil
		},
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			mode = arg.Mode.Tag
			strictConflict = arg.StrictConflict
			return &files.FileMetadata{}, nil
		},
	}
	stubFilesClient(t, mock)

	err := putFile(tmpFile, "/existing.txt", putOptions{
		chunkSize: 1 << 24,
		workers:   4,
		ifExists:  putIfExistsOverwrite,
	})
	if err != nil {
		t.Fatalf("putFile error: %v", err)
	}
	if metadataCalls != 0 {
		t.Errorf("metadata calls = %d, want 0 for overwrite", metadataCalls)
	}
	if mode != files.WriteModeOverwrite {
		t.Errorf("write mode = %q, want %q", mode, files.WriteModeOverwrite)
	}
	if strictConflict {
		t.Error("strict conflict = true, want false for overwrite")
	}
}

func TestPutFileIfExistsSkipTreatsUploadFileConflictAsSkip(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	var strictConflict bool
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, getMetadataNotFoundError()
		},
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			strictConflict = arg.StrictConflict
			return nil, uploadPathConflictError(files.WriteConflictErrorFile)
		},
	}
	stubFilesClient(t, mock)

	stderr := captureStderr(t, func() {
		err := putFile(tmpFile, "/race.txt", putOptions{
			chunkSize: 1 << 24,
			workers:   4,
			ifExists:  putIfExistsSkip,
		})
		if err != nil {
			t.Fatalf("putFile error: %v", err)
		}
	})
	if !strings.Contains(stderr, "Skipping /race.txt") {
		t.Fatalf("stderr = %q, want skip message", stderr)
	}
	if !strictConflict {
		t.Error("strict conflict = false, want true")
	}
}

func TestPutFileIfExistsSkipReturnsNonFileConflicts(t *testing.T) {
	tests := []struct {
		name        string
		conflictTag string
	}{
		{"folder", files.WriteConflictErrorFolder},
		{"file ancestor", files.WriteConflictErrorFileAncestor},
		{"other", files.WriteConflictErrorOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "test.txt")
			if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
				t.Fatal(err)
			}

			mock := &mockFilesClient{
				getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
					return nil, getMetadataNotFoundError()
				},
				uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
					return nil, uploadPathConflictError(tt.conflictTag)
				},
			}
			stubFilesClient(t, mock)

			stderr := captureStderr(t, func() {
				err := putFile(tmpFile, "/bad-conflict.txt", putOptions{
					chunkSize: 1 << 24,
					workers:   4,
					ifExists:  putIfExistsSkip,
				})
				if err == nil {
					t.Fatal("expected non-file conflict error")
				}
			})
			if strings.Contains(stderr, "Skipping ") {
				t.Fatalf("stderr = %q, should not skip non-file conflicts", stderr)
			}
		})
	}
}

func TestIsUploadDestinationFileConflict(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"upload file conflict", uploadPathConflictError(files.WriteConflictErrorFile), true},
		{"upload folder conflict", uploadPathConflictError(files.WriteConflictErrorFolder), false},
		{"upload pointer file conflict", uploadPathConflictErrorPtr(files.WriteConflictErrorFile), true},
		{"finish file conflict", uploadSessionFinishPathConflictError(files.WriteConflictErrorFile), true},
		{"finish file ancestor conflict", uploadSessionFinishPathConflictError(files.WriteConflictErrorFileAncestor), false},
		{"finish pointer file conflict", uploadSessionFinishPathConflictErrorPtr(files.WriteConflictErrorFile), true},
		{"generic write conflict without subtype", uploadWriteConflictError(), false},
		{"plain conflict string", fmt.Errorf("path/conflict/file/"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUploadDestinationFileConflict(tt.err)
			if got != tt.want {
				t.Fatalf("isUploadDestinationFileConflict() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPutRecursiveIfExistsSkipContinuesPastExistingFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	var uploaded []string
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			if arg.Path == "/backup/existing.txt" {
				return &files.FileMetadata{Metadata: files.Metadata{PathDisplay: arg.Path}}, nil
			}
			return nil, getMetadataNotFoundError()
		},
		uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
			uploaded = append(uploaded, arg.Path)
			return &files.FileMetadata{}, nil
		},
		createFolderV2Fn: func(arg *files.CreateFolderArg) (*files.CreateFolderResult, error) {
			return &files.CreateFolderResult{}, nil
		},
	}
	stubFilesClient(t, mock)

	var stderr bytes.Buffer
	cmd := testPutCmd()
	cmd.SetErr(&stderr)

	err := putRecursive(dir, "/backup", putOptions{
		chunkSize: 1 << 24,
		workers:   4,
		ifExists:  putIfExistsSkip,
		output:    commandOutput(cmd),
	})
	if err != nil {
		t.Fatalf("putRecursive error: %v", err)
	}

	if len(uploaded) != 1 || uploaded[0] != "/backup/new.txt" {
		t.Fatalf("uploaded = %v, want [/backup/new.txt]", uploaded)
	}
	if strings.Contains(stderr.String(), "Uploading ") {
		t.Fatalf("stderr = %q, should not print uploading before skip decision", stderr.String())
	}
	if !strings.Contains(stderr.String(), "Processing ") {
		t.Fatalf("stderr = %q, want processing status", stderr.String())
	}
	if !strings.Contains(stderr.String(), "Skipping /backup/existing.txt") {
		t.Fatalf("stderr = %q, want skip status", stderr.String())
	}
}

func TestSingleShotUploadSizeCutoff(t *testing.T) {
	if singleShotUploadSizeCutoff != 32*1024*1024 {
		t.Errorf("singleShotUploadSizeCutoff = %d, want %d", singleShotUploadSizeCutoff, 32*1024*1024)
	}
}

func TestUploadChunked_RetriesSessionStart(t *testing.T) {
	stubRetrySleep(t)
	startCalls := 0
	var appendCalls int
	var finishCalled bool
	mock := &mockFilesClient{
		uploadSessionStartFn: func(arg *files.UploadSessionStartArg, content io.Reader) (*files.UploadSessionStartResult, error) {
			startCalls++
			if startCalls < 3 {
				return nil, dbxauth.ServerError{APIError: dropbox.APIError{ErrorSummary: "500"}}
			}
			return &files.UploadSessionStartResult{SessionId: "test-session"}, nil
		},
		uploadSessionAppendV2Fn: func(arg *files.UploadSessionAppendArg, content io.Reader) error {
			appendCalls++
			return nil
		},
		uploadSessionFinishFn: func(arg *files.UploadSessionFinishArg, content io.Reader) (*files.FileMetadata, error) {
			finishCalled = true
			return &files.FileMetadata{}, nil
		},
	}

	data := bytes.Repeat([]byte("x"), 1024)
	reader := bytes.NewReader(data)
	commitInfo := files.NewCommitInfo("/test.txt")

	_, err := uploadChunked(mock, reader, commitInfo, int64(len(data)), 1, 512, false)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if startCalls != 3 {
		t.Errorf("expected 3 session start calls, got %d", startCalls)
	}
	if appendCalls != 2 {
		t.Errorf("expected 2 append calls, got %d", appendCalls)
	}
	if !finishCalled {
		t.Error("expected finish to be called")
	}
}

func TestUploadChunked_RetriesFinishTooManyWriteOperations(t *testing.T) {
	stubRetrySleep(t)
	var appendCalls int
	var finishCalls int
	mock := &mockFilesClient{
		uploadSessionStartFn: func(arg *files.UploadSessionStartArg, content io.Reader) (*files.UploadSessionStartResult, error) {
			return &files.UploadSessionStartResult{SessionId: "test-session"}, nil
		},
		uploadSessionAppendV2Fn: func(arg *files.UploadSessionAppendArg, content io.Reader) error {
			appendCalls++
			return nil
		},
		uploadSessionFinishFn: func(arg *files.UploadSessionFinishArg, content io.Reader) (*files.FileMetadata, error) {
			finishCalls++
			if finishCalls == 1 {
				return nil, finishTooManyWriteOperationsError()
			}
			return &files.FileMetadata{}, nil
		},
	}

	data := bytes.Repeat([]byte("x"), 1024)
	reader := bytes.NewReader(data)
	commitInfo := files.NewCommitInfo("/test.txt")

	_, err := uploadChunked(mock, reader, commitInfo, int64(len(data)), 1, 512, false)
	if err != nil {
		t.Errorf("expected nil error after finish retry, got %v", err)
	}
	if appendCalls != 2 {
		t.Errorf("expected 2 append calls, got %d", appendCalls)
	}
	if finishCalls != 2 {
		t.Errorf("expected 2 finish calls, got %d", finishCalls)
	}
}

func TestUploadChunked_Success(t *testing.T) {
	var appendCalls int
	var finishCalled bool
	mock := &mockFilesClient{
		uploadSessionStartFn: func(arg *files.UploadSessionStartArg, content io.Reader) (*files.UploadSessionStartResult, error) {
			return &files.UploadSessionStartResult{SessionId: "test-session"}, nil
		},
		uploadSessionAppendV2Fn: func(arg *files.UploadSessionAppendArg, content io.Reader) error {
			appendCalls++
			return nil
		},
		uploadSessionFinishFn: func(arg *files.UploadSessionFinishArg, content io.Reader) (*files.FileMetadata, error) {
			finishCalled = true
			return &files.FileMetadata{}, nil
		},
	}

	data := bytes.Repeat([]byte("x"), 1024)
	reader := bytes.NewReader(data)
	commitInfo := files.NewCommitInfo("/test.txt")

	_, err := uploadChunked(mock, reader, commitInfo, int64(len(data)), 1, 512, false)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if appendCalls != 2 {
		t.Errorf("expected 2 append calls, got %d", appendCalls)
	}
	if !finishCalled {
		t.Error("expected finish to be called")
	}
}
