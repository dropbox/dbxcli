package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	dbxauth "github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/auth"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
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
