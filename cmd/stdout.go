package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

var errStdoutBrokenPipe = errors.New("stdout broken pipe")

type partialTransferError struct {
	bytesWritten int64
}

func (e partialTransferError) Error() string {
	return fmt.Sprintf("download failed after writing %d bytes to stdout; cannot retry partial output", e.bytesWritten)
}

func (e partialTransferError) JSONErrorCode() string {
	return jsonErrorCodePartialTransfer
}

func (e partialTransferError) JSONErrorDetails() map[string]any {
	return map[string]any{"bytes_written": e.bytesWritten}
}

func downloadToStdout(dbx filesClient, src string, w io.Writer) error {
	return downloadToStdoutWithMetadata(dbx, src, nil, w)
}

func downloadToStdoutWithMetadata(dbx filesClient, src string, metadata *files.FileMetadata, w io.Writer) error {
	ignoreBrokenPipeSignal()

	arg := files.NewDownloadArg(src)
	var bytesWritten int64

	err := retryWithBackoff(func() error {
		if bytesWritten > 0 {
			return partialStdoutError(bytesWritten)
		}

		contents, err := stdoutReadCloser(dbx, arg, metadata)
		if err != nil {
			return err
		}
		defer func() { _ = contents.Close() }()

		n, copyErr := io.Copy(stdoutBrokenPipeWriter{w: w}, contents)
		bytesWritten += n

		if errors.Is(copyErr, errStdoutBrokenPipe) {
			return nil
		}
		if copyErr != nil && bytesWritten > 0 {
			return partialStdoutError(bytesWritten)
		}
		return copyErr
	})

	return err
}

func stdoutReadCloser(dbx filesClient, downloadArg *files.DownloadArg, metadata *files.FileMetadata) (io.ReadCloser, error) {
	if isExportOnlyFile(metadata) {
		_, contents, err := exportFile(dbx, downloadArg.Path)
		return contents, err
	}

	_, contents, err := dbx.DownloadContext(currentContext(), downloadArg)
	return contents, err
}

func partialStdoutError(bytesWritten int64) error {
	return partialTransferError{bytesWritten: bytesWritten}
}

type stdoutBrokenPipeWriter struct {
	w io.Writer
}

func (w stdoutBrokenPipeWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	if err != nil && isEPIPE(err) {
		return n, errStdoutBrokenPipe
	}
	return n, err
}

func isEPIPE(err error) bool {
	if errors.Is(err, syscall.EPIPE) {
		return true
	}
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return errors.Is(pathErr.Err, syscall.EPIPE)
	}
	return false
}
