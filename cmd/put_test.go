package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
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

	err := uploadSingleShot(mock, reader, uploadArg, int64(len(data)))
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

	err := uploadSingleShot(mock, reader, uploadArg, int64(len(data)))
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
	return cmd
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
	os.MkdirAll(filepath.Join(dir, "sub", "deep"), 0755)
	os.WriteFile(filepath.Join(dir, "root.txt"), []byte("root"), 0644)
	os.WriteFile(filepath.Join(dir, "sub", "mid.txt"), []byte("mid"), 0644)
	os.WriteFile(filepath.Join(dir, "sub", "deep", "leaf.txt"), []byte("leaf"), 0644)

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
	os.MkdirAll(filepath.Join(dir, "has-files"), 0755)
	os.MkdirAll(filepath.Join(dir, "empty"), 0755)
	os.MkdirAll(filepath.Join(dir, "empty", "nested"), 0755)
	os.WriteFile(filepath.Join(dir, "has-files", "a.txt"), []byte("a"), 0644)

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
	os.WriteFile(filepath.Join(dir, "real.txt"), []byte("real"), 0644)
	os.Symlink(filepath.Join(dir, "real.txt"), filepath.Join(dir, "link.txt"))

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
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := testPutCmd()
	_ = cmd.Flags().Set("chunksize", "100")
	err := put(cmd, []string{tmpFile, "/test.txt"})
	if err == nil || err.Error() != "`put` requires chunk size to be multiple of 4MiB" {
		t.Errorf("expected chunk size validation error, got %v", err)
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

	err := uploadChunked(mock, reader, commitInfo, int64(len(data)), 1, 512, false)
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

	err := uploadChunked(mock, reader, commitInfo, int64(len(data)), 1, 512, false)
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

	err := uploadChunked(mock, reader, commitInfo, int64(len(data)), 1, 512, false)
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
