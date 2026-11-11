package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func TestUploadChunkedCancelsHTTPRequests(t *testing.T) {
	for _, scenario := range []string{"source error", "source truncation", "sibling error", "parent cancellation", "parent deadline"} {
		t.Run(scenario, func(t *testing.T) {
			parent, cancel := context.WithCancel(context.Background())
			if scenario == "parent deadline" {
				cancel()
				parent, cancel = context.WithTimeout(context.Background(), time.Second)
			}
			old := commandContext
			commandContext = parent
			defer func() { commandContext = old }()
			started := make(chan struct{})
			canceled := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case strings.HasSuffix(r.URL.Path, "upload_session/start"):
					w.Header().Set("Content-Type", "application/json")
					_, _ = io.WriteString(w, `{"session_id":"session"}`)
				case strings.HasSuffix(r.URL.Path, "upload_session/append_v2"):
					_, _ = io.Copy(io.Discard, r.Body)
					var arg files.UploadSessionAppendArg
					if err := json.Unmarshal([]byte(r.Header.Get("Dropbox-API-Arg")), &arg); err != nil {
						t.Error(err)
						w.WriteHeader(400)
						return
					}
					if arg.Cursor.Offset != 0 {
						<-started
						w.WriteHeader(http.StatusBadRequest)
						_, _ = io.WriteString(w, "permanent sibling append error")
						return
					}
					close(started)
					<-r.Context().Done()
					close(canceled)
				default:
					t.Errorf("unexpected request %s", r.URL.Path)
					w.WriteHeader(http.StatusBadRequest)
				}
			}))
			defer server.Close()
			defer cancel() // Also releases a stalled handler when a regression fails.
			dbx := files.NewContext(dropbox.Config{Client: server.Client(), URLGenerator: func(_, ns, route string) string { return server.URL + "/" + ns + "/" + route }})
			var source io.Reader = bytes.NewReader([]byte("a"))
			size, workers := int64(1), 1
			sourceErr := errors.New("source read failed")
			switch scenario {
			case "source error":
				size = 2
				source = &uploadSourceFailure{started: started, failure: sourceErr}
			case "source truncation":
				size = 2
				source = &uploadSourceFailure{started: started, failure: io.EOF}
			case "sibling error":
				size = 2
				workers = 2
				source = bytes.NewReader([]byte("ab"))
			}
			done := make(chan error, 1)
			go func() {
				_, err := uploadChunked(dbx, source, files.NewCommitInfo("/file"), size, workers, 1, false)
				done <- err
			}()
			select {
			case <-started:
			case <-time.After(5 * time.Second):
				t.Fatal("append request did not start")
			}
			if scenario == "parent cancellation" {
				cancel()
			}
			select {
			case err := <-done:
				if err == nil {
					t.Fatal("expected upload failure")
				}
				switch scenario {
				case "source error":
					if !errors.Is(err, sourceErr) {
						t.Fatalf("error=%v, want original source error", err)
					}
				case "source truncation":
					if !strings.Contains(err.Error(), "failed to read") {
						t.Fatalf("error=%v, want short-read error", err)
					}
				case "sibling error":
					if !strings.Contains(err.Error(), "permanent sibling append error") {
						t.Fatalf("error=%v, want original append error", err)
					}
				case "parent cancellation":
					if !errors.Is(err, context.Canceled) {
						t.Fatalf("error=%v, want cancellation", err)
					}
				case "parent deadline":
					if !errors.Is(err, context.DeadlineExceeded) {
						t.Fatalf("error=%v, want deadline exceeded", err)
					}
				}
			case <-time.After(5 * time.Second):
				t.Fatal("upload did not cancel its stalled request")
			}
			select {
			case <-canceled:
			case <-time.After(time.Second):
				t.Fatal("HTTP server did not observe request cancellation")
			}
			if scenario != "parent cancellation" && scenario != "parent deadline" && parent.Err() != nil {
				t.Fatalf("upload canceled the command context: %v", parent.Err())
			}
		})
	}
}

func TestUploadChunkedCancelsRetryWaitOnSourceError(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	oldContext := commandContext
	commandContext = parent
	defer func() { commandContext = oldContext }()
	waiting := make(chan struct{})
	oldSleep := retrySleep
	retrySleep = func(ctx context.Context, _ time.Duration) error { close(waiting); <-ctx.Done(); return ctx.Err() }
	defer func() { retrySleep = oldSleep }()
	sourceErr := errors.New("source read failed")
	mock := &mockFilesClient{
		uploadSessionStartFn: func(*files.UploadSessionStartArg, io.Reader) (*files.UploadSessionStartResult, error) {
			return &files.UploadSessionStartResult{SessionId: "session"}, nil
		},
		uploadSessionAppendV2Fn: func(*files.UploadSessionAppendArg, io.Reader) error { return io.ErrUnexpectedEOF },
	}
	done := make(chan error, 1)
	go func() {
		_, err := uploadChunked(mock, &uploadSourceFailure{started: waiting, failure: sourceErr}, files.NewCommitInfo("/file"), 2, 1, 1, false)
		done <- err
	}()
	select {
	case err := <-done:
		if !errors.Is(err, sourceErr) {
			t.Fatalf("error=%v, want original source error", err)
		}
	case <-time.After(5 * time.Second):
		cancel()
		<-done
		t.Fatal("source failure did not cancel retry wait")
	}
	if parent.Err() != nil {
		t.Fatalf("command context canceled: %v", parent.Err())
	}
}

// Each call represents one source chunk, allowing tests to cancel at precise
// points without filesystem timing or a blocked read that needs manual release.
type uploadReadObserver struct {
	reads      int
	beforeRead func(int)
}

func (r *uploadReadObserver) Read(p []byte) (int, error) {
	r.reads++
	if r.beforeRead != nil {
		r.beforeRead(r.reads)
	}
	p[0] = 'a'
	return 1, nil
}

func TestUploadChunkedDoesNotReadAfterCancellation(t *testing.T) {
	for _, when := range []string{"before first read", "during first read"} {
		t.Run(when, func(t *testing.T) {
			// Repeat the second case to exercise both simultaneously-ready select arms.
			for i := 0; i < 100; i++ {
				ctx, cancel := context.WithCancel(context.Background())
				old := commandContext
				commandContext = ctx
				source := &uploadReadObserver{}
				mock := &mockFilesClient{uploadSessionStartFn: func(*files.UploadSessionStartArg, io.Reader) (*files.UploadSessionStartResult, error) {
					if when == "before first read" {
						cancel()
					}
					return &files.UploadSessionStartResult{SessionId: "session"}, nil
				}}
				wantReads := 0
				if when == "during first read" {
					wantReads = 1
					source.beforeRead = func(int) { cancel() }
				}
				_, err := uploadChunked(mock, source, files.NewCommitInfo("/file"), 2, 1, 1, false)
				cancel()
				commandContext = old
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("error=%v, want cancellation", err)
				}
				if source.reads != wantReads {
					t.Fatalf("trial %d: source reads=%d, want %d", i, source.reads, wantReads)
				}
			}
		})
	}
}

type uploadContextClient struct {
	*mockFilesClient
	appendContext func(context.Context, *files.UploadSessionAppendArg, io.Reader) error
}

func (c *uploadContextClient) UploadSessionAppendV2Context(ctx context.Context, arg *files.UploadSessionAppendArg, r io.Reader) error {
	return c.appendContext(ctx, arg, r)
}

func TestUploadChunkedCancellationPreservesWorkerError(t *testing.T) {
	parent := context.Background()
	old := commandContext
	commandContext = parent
	defer func() { commandContext = old }()
	workerContext := make(chan context.Context, 1)
	failure := errors.New("permanent worker error")
	mock := &uploadContextClient{
		mockFilesClient: &mockFilesClient{uploadSessionStartFn: func(*files.UploadSessionStartArg, io.Reader) (*files.UploadSessionStartResult, error) {
			return &files.UploadSessionStartResult{SessionId: "session"}, nil
		}},
		appendContext: func(ctx context.Context, _ *files.UploadSessionAppendArg, _ io.Reader) error {
			workerContext <- ctx
			return failure
		},
	}
	source := &uploadReadObserver{beforeRead: func(read int) {
		if read == 2 {
			ctx := <-workerContext
			<-ctx.Done()
		}
	}}
	_, err := uploadChunked(mock, source, files.NewCommitInfo("/file"), 3, 1, 1, false)
	if !errors.Is(err, failure) {
		t.Fatalf("error=%v, want original worker failure", err)
	}
	if source.reads > 2 {
		t.Fatalf("source reads=%d, want at most 2", source.reads)
	}
	if parent.Err() != nil {
		t.Fatalf("parent canceled: %v", parent.Err())
	}
}

func TestPutRecursiveContinuesAfterSessionCancellation(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a-fail", "b-good"} {
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Truncate(singleShotUploadSizeCutoff + 1); err != nil {
			f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	sessions := 0
	var committed []string
	failure := errors.New("first file append failed")
	stubFilesClient(t, &mockFilesClient{
		uploadSessionStartFn: func(*files.UploadSessionStartArg, io.Reader) (*files.UploadSessionStartResult, error) {
			sessions++
			return &files.UploadSessionStartResult{SessionId: strconv.Itoa(sessions)}, nil
		},
		uploadSessionAppendV2Fn: func(arg *files.UploadSessionAppendArg, _ io.Reader) error {
			if arg.Cursor.SessionId == "1" {
				return failure
			}
			return nil
		},
		uploadSessionFinishFn: func(arg *files.UploadSessionFinishArg, _ io.Reader) (*files.FileMetadata, error) {
			committed = append(committed, arg.Commit.Path)
			return &files.FileMetadata{}, nil
		},
		createFolderV2Fn: func(*files.CreateFolderArg) (*files.CreateFolderResult, error) {
			return &files.CreateFolderResult{}, nil
		},
	})
	err := putRecursive(dir, "/target", putOptions{workers: 1, chunkSize: putChunkSizeUnit, errOut: io.Discard})
	if err == nil || !strings.Contains(err.Error(), failure.Error()) {
		t.Fatalf("error=%v, want aggregate failure", err)
	}
	if sessions != 2 || len(committed) != 1 || committed[0] != "/target/b-good" {
		t.Fatalf("sessions=%d committed=%v, want second file committed", sessions, committed)
	}
}
