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

func downloadToStdout(dbx files.Client, src string, w io.Writer) error {
	ignoreBrokenPipeSignal()

	arg := files.NewDownloadArg(src)
	var bytesWritten int64

	err := retryWithBackoff(func() error {
		if bytesWritten > 0 {
			return partialStdoutError(bytesWritten)
		}

		_, contents, err := dbx.Download(arg)
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

func partialStdoutError(bytesWritten int64) error {
	return fmt.Errorf("download failed after writing %d bytes to stdout; cannot retry partial output", bytesWritten)
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
