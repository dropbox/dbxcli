package cmd

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func TestUploadChunkedReturnsConcurrentAppendErrors(t *testing.T) {
	for _, workers := range []int{1, 2, 4} {
		t.Run(strconv.Itoa(workers)+" workers", func(t *testing.T) {
			failure := errors.New("permanent append error")
			var started sync.WaitGroup
			started.Add(workers)
			mock := &mockFilesClient{
				uploadSessionStartFn: func(*files.UploadSessionStartArg, io.Reader) (*files.UploadSessionStartResult, error) {
					return &files.UploadSessionStartResult{SessionId: "session"}, nil
				},
				uploadSessionAppendV2Fn: func(*files.UploadSessionAppendArg, io.Reader) error { started.Done(); started.Wait(); return failure },
				uploadSessionFinishFn: func(*files.UploadSessionFinishArg, io.Reader) (*files.FileMetadata, error) {
					t.Error("finish called after append failure")
					return nil, nil
				},
			}
			done := make(chan error, 1)
			go func() {
				_, err := uploadChunked(mock, bytes.NewReader(make([]byte, workers)), files.NewCommitInfo("/file"), int64(workers), workers, 1, false)
				done <- err
			}()
			select {
			case err := <-done:
				if !errors.Is(err, failure) {
					t.Fatalf("error=%v, want %v", err, failure)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("upload hung while workers reported append errors")
			}
		})
	}
}

type uploadSourceFailure struct {
	started <-chan struct{}
	failure error
	read    bool
}

func (r *uploadSourceFailure) Read(p []byte) (int, error) {
	if !r.read {
		r.read = true
		p[0] = 'a'
		return 1, nil
	}
	<-r.started
	return 0, r.failure
}

func TestUploadChunkedJoinsWorkersOnSourceFailure(t *testing.T) {
	for _, tc := range []struct {
		name    string
		failure error
		want    string
	}{
		{"read error", errors.New("source read failed"), "source read failed"},
		{"truncation", io.EOF, "failed to read"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			started := make(chan struct{})
			release := make(chan struct{})
			mock := &mockFilesClient{
				uploadSessionStartFn: func(*files.UploadSessionStartArg, io.Reader) (*files.UploadSessionStartResult, error) {
					return &files.UploadSessionStartResult{SessionId: "session"}, nil
				},
				uploadSessionAppendV2Fn: func(*files.UploadSessionAppendArg, io.Reader) error {
					close(started)
					<-release
					return errors.New("append failed too")
				},
				uploadSessionFinishFn: func(*files.UploadSessionFinishArg, io.Reader) (*files.FileMetadata, error) {
					t.Error("finish called after source failure")
					return nil, nil
				},
			}
			done := make(chan error, 1)
			go func() {
				_, err := uploadChunked(mock, &uploadSourceFailure{started: started, failure: tc.failure}, files.NewCommitInfo("/file"), 2, 1, 1, false)
				done <- err
			}()
			select {
			case err := <-done:
				close(release)
				t.Fatalf("returned before joining in-flight worker: %v", err)
			case <-time.After(50 * time.Millisecond):
			}
			close(release)
			select {
			case err := <-done:
				if err == nil || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("error=%v, want %q", err, tc.want)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("upload did not return after source and append errors")
			}
		})
	}
}
