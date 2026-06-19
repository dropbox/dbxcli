package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestSpoolStdinToTemp_Success(t *testing.T) {
	data := []byte("hello from stdin")
	r := bytes.NewReader(data)

	path, size, cleanup, err := spoolStdinToTemp(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = cleanup() }()

	if size != int64(len(data)) {
		t.Errorf("size = %d, want %d", size, len(data))
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("temp file content = %q, want %q", got, data)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat temp file: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Errorf("temp file permissions = %o, want 0600", info.Mode().Perm())
	}
}

func TestSpoolStdinToTemp_Empty(t *testing.T) {
	r := bytes.NewReader(nil)

	path, size, cleanup, err := spoolStdinToTemp(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = cleanup() }()

	if size != 0 {
		t.Errorf("size = %d, want 0", size)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("temp file content = %q, want empty", got)
	}
}

func TestSpoolStdinToTemp_ReadError(t *testing.T) {
	r := &failingReader{err: io.ErrUnexpectedEOF}

	_, _, _, err := spoolStdinToTemp(r)
	if err == nil {
		t.Fatal("expected error for failing reader")
	}
}

func TestSpoolStdinToTemp_ReadErrorReportsCleanupFailure(t *testing.T) {
	readErr := io.ErrUnexpectedEOF
	cleanupErr := errors.New("remove failed")
	var tempPath string
	origRemoveFile := removeFile
	removeFile = func(path string) error {
		tempPath = path
		return cleanupErr
	}
	t.Cleanup(func() {
		removeFile = origRemoveFile
		if tempPath != "" {
			_ = origRemoveFile(tempPath)
		}
	})

	_, _, _, err := spoolStdinToTemp(&failingReader{err: readErr})
	if err == nil {
		t.Fatal("expected error for failing reader")
	}
	if !errors.Is(err, readErr) {
		t.Fatalf("error = %v, want read error", err)
	}
	if !errors.Is(err, cleanupErr) {
		t.Fatalf("error = %v, want cleanup error", err)
	}
	if tempPath == "" {
		t.Fatal("expected temp path to be captured")
	}
	if !strings.Contains(err.Error(), "sensitive stdin data may remain on disk") {
		t.Errorf("error = %q, want sensitive-data warning", err.Error())
	}
}

func TestSpoolStdinToTemp_CleanupRemovesFile(t *testing.T) {
	r := bytes.NewReader([]byte("data"))

	path, _, cleanup, err := spoolStdinToTemp(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := cleanup(); err != nil {
		t.Fatalf("cleanup error: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected temp file to be removed after cleanup, stat err = %v", err)
	}
}

type failingReader struct {
	err error
}

func (r *failingReader) Read(p []byte) (int, error) {
	return 0, r.err
}
