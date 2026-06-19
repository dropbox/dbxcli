package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var removeFile = os.Remove

func spoolStdinToTemp(r io.Reader) (path string, size int64, cleanup func() error, err error) {
	f, err := os.CreateTemp("", "dbxcli-stdin-*")
	if err != nil {
		return "", 0, nil, fmt.Errorf("create temp file: %w", err)
	}
	if err := os.Chmod(f.Name(), 0600); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return "", 0, nil, fmt.Errorf("chmod temp file: %w", err)
	}

	tmpPath := f.Name()
	removeTmp := func() error { return removeFile(tmpPath) }

	n, copyErr := io.Copy(f, r)
	if closeErr := f.Close(); closeErr != nil {
		if copyErr != nil {
			return "", 0, nil, stdinSpoolFailureError(tmpPath, copyErr, removeTmp())
		}
		return "", 0, nil, stdinSpoolFailureError(tmpPath, closeErr, removeTmp())
	}
	if copyErr != nil {
		return "", 0, nil, stdinSpoolFailureError(tmpPath, copyErr, removeTmp())
	}

	return tmpPath, n, removeTmp, nil
}

func stdinSpoolFailureError(tmpPath string, cause error, cleanupErr error) error {
	if cleanupErr == nil {
		return cause
	}
	return errors.Join(
		cause,
		fmt.Errorf("failed to remove temp file %s after stdin spool failure; sensitive stdin data may remain on disk: %w", tmpPath, cleanupErr),
	)
}
