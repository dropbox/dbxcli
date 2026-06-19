package cmd

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"syscall"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	dbxauth "github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/auth"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func TestDownloadToStdout_Success(t *testing.T) {
	content := "file content"
	mock := &mockFilesClient{
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			meta := &files.FileMetadata{Size: uint64(len(content))}
			return meta, io.NopCloser(strings.NewReader(content)), nil
		},
	}

	var buf bytes.Buffer
	err := downloadToStdout(mock, "/file.txt", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != content {
		t.Errorf("output = %q, want %q", buf.String(), content)
	}
}

func TestDownloadToStdout_RetriesBeforeFirstByte(t *testing.T) {
	stubRetrySleep(t)
	content := "file content"
	calls := 0

	mock := &mockFilesClient{
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			calls++
			if calls < 3 {
				return nil, nil, dbxauth.ServerError{APIError: dropbox.APIError{ErrorSummary: "500"}}
			}
			meta := &files.FileMetadata{Size: uint64(len(content))}
			return meta, io.NopCloser(strings.NewReader(content)), nil
		},
	}

	var buf bytes.Buffer
	err := downloadToStdout(mock, "/file.txt", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
	if buf.String() != content {
		t.Errorf("output = %q, want %q", buf.String(), content)
	}
}

func TestDownloadToStdout_NoRetryAfterPartialOutput(t *testing.T) {
	retryDelays := stubRetrySleep(t)
	downloadCalls := 0

	mock := &mockFilesClient{
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			downloadCalls++
			meta := &files.FileMetadata{Size: 100}
			return meta, &failingReadCloser{data: []byte("partial")}, nil
		},
	}

	var buf bytes.Buffer
	err := downloadToStdout(mock, "/file.txt", &buf)
	if err == nil {
		t.Fatal("expected error for partial output failure")
	}
	if !strings.Contains(err.Error(), "cannot retry") {
		t.Errorf("error = %q, want mention of cannot retry", err.Error())
	}
	if downloadCalls != 1 {
		t.Errorf("download calls = %d, want 1 (no second Download call after partial write)", downloadCalls)
	}
	if len(*retryDelays) != 0 {
		t.Errorf("retry delays = %v, want no retry sleep after partial output", *retryDelays)
	}
}

func TestDownloadToStdout_WriterEPIPEReturnsNil(t *testing.T) {
	mock := &mockFilesClient{
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			meta := &files.FileMetadata{Size: 100}
			return meta, io.NopCloser(strings.NewReader("some data")), nil
		},
	}

	err := downloadToStdout(mock, "/file.txt", epipeWriter{})
	if err != nil {
		t.Fatalf("expected nil error for writer EPIPE, got %v", err)
	}
}

func TestDownloadToStdout_ReaderEPIPEReturnsError(t *testing.T) {
	downloadCalls := 0
	mock := &mockFilesClient{
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			downloadCalls++
			meta := &files.FileMetadata{Size: 100}
			return meta, io.NopCloser(&epipeReader{data: []byte("some")}), nil
		},
	}

	var buf bytes.Buffer
	err := downloadToStdout(mock, "/file.txt", &buf)
	if err == nil {
		t.Fatal("expected error for reader EPIPE")
	}
	if !strings.Contains(err.Error(), "cannot retry") {
		t.Fatalf("error = %q, want partial-output no-retry error", err.Error())
	}
	if downloadCalls != 1 {
		t.Errorf("download calls = %d, want 1", downloadCalls)
	}
}

func TestIsEPIPE(t *testing.T) {
	if !isEPIPE(syscall.EPIPE) {
		t.Error("expected EPIPE to be recognized")
	}
	if isEPIPE(io.EOF) {
		t.Error("expected EOF to not be EPIPE")
	}
	if isEPIPE(nil) {
		t.Error("expected nil to not be EPIPE")
	}
	if isEPIPE(errors.New("random error")) {
		t.Error("expected random error to not be EPIPE")
	}
}

type epipeReader struct {
	data []byte
	sent bool
}

func (r *epipeReader) Read(p []byte) (int, error) {
	if !r.sent {
		r.sent = true
		n := copy(p, r.data)
		return n, nil
	}
	return 0, syscall.EPIPE
}

type epipeWriter struct{}

func (epipeWriter) Write(p []byte) (int, error) {
	return 0, syscall.EPIPE
}
