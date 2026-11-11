package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/contenthash"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func samePutMetadata(t *testing.T, data []byte) *files.FileMetadata {
	t.Helper()
	hash, err := contenthash.Compute(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return &files.FileMetadata{Size: uint64(len(data)), ContentHash: hash}
}

func forbidPutUploads(t *testing.T) *mockFilesClient {
	t.Helper()
	return &mockFilesClient{
		uploadFn: func(*files.UploadArg, io.Reader) (*files.FileMetadata, error) {
			t.Fatal("unexpected upload")
			return nil, nil
		},
		uploadSessionStartFn: func(*files.UploadSessionStartArg, io.Reader) (*files.UploadSessionStartResult, error) {
			t.Fatal("unexpected session start")
			return nil, nil
		},
		uploadSessionAppendV2Fn: func(*files.UploadSessionAppendArg, io.Reader) error {
			t.Error("unexpected session append")
			return errors.New("unexpected append")
		},
		uploadSessionFinishFn: func(*files.UploadSessionFinishArg, io.Reader) (*files.FileMetadata, error) {
			t.Fatal("unexpected session finish")
			return nil, nil
		},
	}
}

func TestPutSkipSameExisting(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
		kind string
		want string
	}{
		{name: "empty", data: nil},
		{name: "same", data: []byte("hello")},
		{name: "large same", data: bytes.Repeat([]byte("x"), int(singleShotUploadSizeCutoff+1))},
		{name: "different size", data: []byte("hello"), kind: "size", want: "different contents"},
		{name: "different hash", data: []byte("hello"), kind: "hash", want: "different contents"},
		{name: "missing hash", data: []byte("hello"), kind: "no hash", want: "unavailable"},
		{name: "folder", data: []byte("hello"), kind: "folder", want: "folder"},
		{name: "metadata error", data: []byte("hello"), kind: "error", want: "metadata failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := filepath.Join(t.TempDir(), "source")
			if err := os.WriteFile(src, tc.data, 0600); err != nil {
				t.Fatal(err)
			}
			meta := samePutMetadata(t, tc.data)
			switch tc.kind {
			case "size":
				meta.Size++
			case "hash":
				meta.ContentHash = strings.Repeat("0", 64)
			case "no hash":
				meta.ContentHash = ""
			}
			mock := forbidPutUploads(t)
			mock.getMetadataFn = func(*files.GetMetadataArg) (files.IsMetadata, error) {
				if tc.kind == "folder" {
					return &files.FolderMetadata{}, nil
				}
				if tc.kind == "error" {
					return nil, errors.New("metadata failed")
				}
				return meta, nil
			}
			stubFilesClient(t, mock)
			result, err := putFileWithResult(src, "/target", putOptions{ifExists: putIfExistsSkipSame})
			if tc.want != "" {
				if err == nil || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("error = %v, want %q", err, tc.want)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != putStatusSkipped || result.Result == nil {
				t.Fatalf("result = %+v", result)
			}
		})
	}
}

func TestPutSkipSameMissingAndConcurrentCreation(t *testing.T) {
	for _, race := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing", true: "concurrent creation"}[race], func(t *testing.T) {
			data := []byte("entire source contents")
			src := filepath.Join(t.TempDir(), "source")
			if err := os.WriteFile(src, data, 0600); err != nil {
				t.Fatal(err)
			}
			calls := 0
			stubFilesClient(t, &mockFilesClient{
				getMetadataFn: func(*files.GetMetadataArg) (files.IsMetadata, error) { return nil, getMetadataNotFoundError() },
				uploadFn: func(arg *files.UploadArg, r io.Reader) (*files.FileMetadata, error) {
					calls++
					if arg.Mode.Tag != files.WriteModeAdd || !arg.StrictConflict || arg.Autorename {
						t.Fatalf("unsafe upload args: %+v", arg)
					}
					got, err := io.ReadAll(r)
					if err != nil || !bytes.Equal(got, data) {
						t.Fatalf("uploaded %q, error %v", got, err)
					}
					if race {
						return nil, uploadPathConflictError(files.WriteConflictErrorFile)
					}
					return samePutMetadata(t, data), nil
				},
			})
			result, err := putFileWithResult(src, "/target", putOptions{ifExists: putIfExistsSkipSame})
			if calls != 1 {
				t.Fatalf("upload calls = %d", calls)
			}
			if race {
				if err == nil {
					t.Fatal("expected conflict")
				}
				return
			}
			if err != nil || result.Status != putStatusUploaded {
				t.Fatalf("result = %+v, error = %v", result, err)
			}
		})
	}
}

type putHashErrorReader struct{ err error }

func (r putHashErrorReader) Read([]byte) (int, error) { return 0, r.err }
func TestComparePutContentsReadError(t *testing.T) {
	want := errors.New("read failed")
	err := comparePutContents(putHashErrorReader{want}, 1, samePutMetadata(t, []byte("x")), "/target")
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}

func TestPutSkipSameRecursiveMixed(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"same", "new", "different"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var uploaded []string
	mock := forbidPutUploads(t)
	mock.getMetadataFn = func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
		switch arg.Path {
		case "/backup/same":
			return samePutMetadata(t, []byte("same")), nil
		case "/backup/different":
			return samePutMetadata(t, []byte("DIFFERENT")), nil
		}
		return nil, getMetadataNotFoundError()
	}
	mock.uploadFn = func(arg *files.UploadArg, r io.Reader) (*files.FileMetadata, error) {
		uploaded = append(uploaded, arg.Path)
		return &files.FileMetadata{}, nil
	}
	mock.createFolderV2Fn = func(*files.CreateFolderArg) (*files.CreateFolderResult, error) {
		return &files.CreateFolderResult{}, nil
	}
	stubFilesClient(t, mock)
	err := putRecursive(dir, "/backup", putOptions{ifExists: putIfExistsSkipSame})
	if err == nil || !strings.Contains(err.Error(), "different contents") {
		t.Fatalf("error = %v", err)
	}
	if len(uploaded) != 1 || uploaded[0] != "/backup/new" {
		t.Fatalf("uploaded = %v", uploaded)
	}
}

func TestPutSkipSameStdinJSONAndCleanup(t *testing.T) {
	for _, same := range []bool{true, false} {
		t.Run(map[bool]string{true: "same", false: "different"}[same], func(t *testing.T) {
			tmp := t.TempDir()
			t.Setenv("TMPDIR", tmp)
			t.Setenv("TMP", tmp)
			t.Setenv("TEMP", tmp)
			mock := forbidPutUploads(t)
			mock.getMetadataFn = func(*files.GetMetadataArg) (files.IsMetadata, error) {
				data := "hello"
				if !same {
					data = "other"
				}
				return samePutMetadata(t, []byte(data)), nil
			}
			stubFilesClient(t, mock)
			var stdout, stderr bytes.Buffer
			cmd := testPutJSONCmd(&stdout, &stderr)
			cmd.Use = "put"
			cmd.SetIn(strings.NewReader("hello"))
			if err := cmd.Flags().Set("if-exists", putIfExistsSkipSame); err != nil {
				t.Fatal(err)
			}
			err := put(cmd, []string{"-", "/target"})
			if same {
				if err != nil {
					t.Fatal(err)
				}
				got := decodePutOutput(t, &stdout)
				if got.Input.IfExists != putIfExistsSkipSame || len(got.Results) != 1 || got.Results[0].Status != putStatusSkipped || got.Results[0].Input.Source != "-" {
					t.Fatalf("output = %+v", got)
				}
				if err := compileJSONSchemaFile(t, "../docs/json-schema/v1/commands.schema.json").Validate(decodeJSONValueForSchema(t, stdout.Bytes())); err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("expected conflict")
			}
			entries, err := os.ReadDir(tmp)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Fatalf("temp files remain: %v", entries)
			}
		})
	}
}

func TestPutSkipSameDryRun(t *testing.T) {
	src := filepath.Join(t.TempDir(), "source")
	if err := os.WriteFile(src, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	mock := forbidPutUploads(t)
	mock.getMetadataFn = func(*files.GetMetadataArg) (files.IsMetadata, error) {
		t.Fatal("unexpected destination comparison")
		return nil, nil
	}
	stubFilesClient(t, mock)
	var stdout bytes.Buffer
	cmd := testPutJSONCmd(&stdout, nil)
	for flag, value := range map[string]string{"if-exists": putIfExistsSkipSame, dryRunFlagName: "true"} {
		if err := cmd.Flags().Set(flag, value); err != nil {
			t.Fatal(err)
		}
	}
	// A directory suffix resolves the target without a metadata lookup.
	if err := put(cmd, []string{src, "/backup/"}); err != nil {
		t.Fatal(err)
	}
	got := decodePutOutput(t, &stdout)
	if got.Input.IfExists != putIfExistsSkipSame || got.Results[0].Status != jsonStatusPlanned {
		t.Fatalf("output = %+v", got)
	}
}

func TestComparePutContentsSizeMismatchDoesNotRead(t *testing.T) {
	err := comparePutContents(putHashErrorReader{errors.New("must not read")}, 2, samePutMetadata(t, []byte("x")), "/target")
	if err == nil || !strings.Contains(err.Error(), "different contents") {
		t.Fatalf("error = %v", err)
	}
}

func TestPutSkipSameChunkedConcurrentCreation(t *testing.T) {
	src := filepath.Join(t.TempDir(), "large")
	f, err := os.Create(src)
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
	finished := false
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(*files.GetMetadataArg) (files.IsMetadata, error) { return nil, getMetadataNotFoundError() },
		uploadSessionStartFn: func(*files.UploadSessionStartArg, io.Reader) (*files.UploadSessionStartResult, error) {
			return &files.UploadSessionStartResult{SessionId: "session"}, nil
		},
		uploadSessionFinishFn: func(arg *files.UploadSessionFinishArg, _ io.Reader) (*files.FileMetadata, error) {
			finished = true
			if arg.Commit.Mode.Tag != files.WriteModeAdd || !arg.Commit.StrictConflict || arg.Commit.Autorename {
				t.Errorf("unsafe commit: %+v", arg.Commit)
			}
			return nil, uploadSessionFinishPathConflictError(files.WriteConflictErrorFile)
		},
	})
	_, err = putFileWithResult(src, "/target", putOptions{ifExists: putIfExistsSkipSame, workers: 1, chunkSize: putChunkSizeUnit, errOut: io.Discard})
	if err == nil || !finished {
		t.Fatalf("finished=%v, error=%v, want conflict", finished, err)
	}
}
