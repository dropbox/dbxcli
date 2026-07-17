package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

func testPutCmdWithStdin(stdin io.Reader) *cobra.Command {
	cmd := testPutCmd()
	cmd.SetIn(stdin)
	return cmd
}

type failReadReader struct {
	t *testing.T
}

func (r failReadReader) Read(_ []byte) (int, error) {
	r.t.Fatal("stdin should not be read")
	return 0, io.EOF
}

func TestPutStdin_RequiresTarget(t *testing.T) {
	cmd := testPutCmdWithStdin(strings.NewReader("data"))
	err := put(cmd, []string{"-"})
	if err == nil {
		t.Fatal("expected error for put - without target")
	}
	if !strings.Contains(err.Error(), "explicit target") {
		t.Errorf("error = %q, want mention of explicit target", err.Error())
	}
}

func TestPutStdin_RejectsRecursive(t *testing.T) {
	cmd := testPutCmdWithStdin(strings.NewReader("data"))
	_ = cmd.Flags().Set("recursive", "true")
	err := put(cmd, []string{"-", "/file.txt"})
	if err == nil {
		t.Fatal("expected error for put - --recursive")
	}
	if !strings.Contains(err.Error(), "--recursive") {
		t.Errorf("error = %q, want mention of --recursive", err.Error())
	}
}

func TestPutStdin_RejectsDirectoryStyleTarget(t *testing.T) {
	cmd := testPutCmdWithStdin(strings.NewReader("data"))
	err := put(cmd, []string{"-", "/folder/"})
	if err == nil {
		t.Fatal("expected error for directory-style target")
	}
	if !strings.Contains(err.Error(), "directory target") {
		t.Errorf("error = %q, want mention of directory target", err.Error())
	}
}

func TestPutStdin_RejectsExistingFolder(t *testing.T) {
	cmd := testPutCmdWithStdin(strings.NewReader("data"))

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return &files.FolderMetadata{Metadata: files.Metadata{PathDisplay: arg.Path}}, nil
		},
	}
	stubFilesClient(t, mock)

	err := put(cmd, []string{"-", "/existing-folder"})
	if err == nil {
		t.Fatal("expected error for existing folder target")
	}
	if !strings.Contains(err.Error(), "folder") {
		t.Errorf("error = %q, want mention of folder", err.Error())
	}
}

func TestPutStdin_UploadsContent(t *testing.T) {
	content := "hello from stdin"
	cmd := testPutCmdWithStdin(strings.NewReader(content))

	var uploadedPath string
	var uploadedContent []byte
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, &files.GetMetadataAPIError{}
		},
		uploadFn: func(arg *files.UploadArg, r io.Reader) (*files.FileMetadata, error) {
			uploadedPath = arg.Path
			data, _ := io.ReadAll(r)
			uploadedContent = data
			return &files.FileMetadata{}, nil
		},
	}
	stubFilesClient(t, mock)

	err := put(cmd, []string{"-", "/dest.txt"})
	if err != nil {
		t.Fatalf("put stdin error: %v", err)
	}
	if uploadedPath != "/dest.txt" {
		t.Errorf("uploaded path = %q, want /dest.txt", uploadedPath)
	}
	if !strings.Contains(string(uploadedContent), content) {
		t.Errorf("uploaded content = %q, want to contain %q", uploadedContent, content)
	}
}

func TestPutStdinSetsClientModified(t *testing.T) {
	content := "hello from stdin"
	cmd := testPutCmdWithStdin(strings.NewReader(content))

	start := time.Now().UTC()
	var uploadedClientModified *dropbox.DBXTime
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, &files.GetMetadataAPIError{}
		},
		uploadFn: func(arg *files.UploadArg, r io.Reader) (*files.FileMetadata, error) {
			uploadedClientModified = arg.ClientModified
			data, err := io.ReadAll(r)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != content {
				t.Fatalf("uploaded content = %q, want %q", string(data), content)
			}
			return &files.FileMetadata{}, nil
		},
	}
	stubFilesClient(t, mock)

	err := put(cmd, []string{"-", "/dest.txt"})
	end := time.Now().UTC()
	if err != nil {
		t.Fatalf("put stdin error: %v", err)
	}
	if uploadedClientModified == nil {
		t.Fatal("ClientModified = nil, want stdin spool modified time")
	}

	got := time.Time(*uploadedClientModified)
	lower := start.Add(-time.Second)
	upper := end.Add(time.Second)
	if got.Before(lower) || got.After(upper) {
		t.Fatalf("ClientModified = %s, want between %s and %s", got.Format(time.RFC3339Nano), lower.Format(time.RFC3339Nano), upper.Format(time.RFC3339Nano))
	}
	if got.Location() != time.UTC {
		t.Fatalf("ClientModified location = %v, want UTC", got.Location())
	}
}

func TestPutStdinIfExistsSkipDoesNotReadStdin(t *testing.T) {
	cmd := testPutCmdWithStdin(failReadReader{t: t})
	_ = cmd.Flags().Set("if-exists", "skip")
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return &files.FileMetadata{Metadata: files.Metadata{PathDisplay: arg.Path}}, nil
		},
		uploadFn: func(arg *files.UploadArg, r io.Reader) (*files.FileMetadata, error) {
			t.Fatal("upload should not be called for existing destination with --if-exists skip")
			return nil, nil
		},
	}
	stubFilesClient(t, mock)

	err := put(cmd, []string{"-", "/existing.txt"})
	if err != nil {
		t.Fatalf("put stdin error: %v", err)
	}
	if !strings.Contains(stderr.String(), "Skipping /existing.txt") {
		t.Errorf("stderr = %q, want skip message", stderr.String())
	}
}

func TestPutStdinDryRunDoesNotReadStdin(t *testing.T) {
	cmd := testPutCmdWithStdin(failReadReader{t: t})
	if err := cmd.Flags().Set(dryRunFlagName, "true"); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	stubFilesClient(t, &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			t.Fatalf("GetMetadata called during stdin dry-run: %v", arg)
			return nil, nil
		},
		uploadFn: func(arg *files.UploadArg, r io.Reader) (*files.FileMetadata, error) {
			t.Fatalf("Upload called during stdin dry-run: %v", arg)
			return nil, nil
		},
	})

	if err := put(cmd, []string{"-", "/stdin.txt"}); err != nil {
		t.Fatalf("put stdin dry-run error: %v", err)
	}

	const want = "Would upload - to /stdin.txt\n"
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestPutStdinJSONDryRunOutputsPlannedResult(t *testing.T) {
	cmd := testPutCmdWithStdin(failReadReader{t: t})
	cmd.Flags().String(outputFlag, "text", "")
	if err := cmd.Flags().Set(outputFlag, "json"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set(dryRunFlagName, "true"); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	stubFilesClient(t, &mockFilesClient{
		uploadFn: func(arg *files.UploadArg, r io.Reader) (*files.FileMetadata, error) {
			t.Fatalf("Upload called during stdin dry-run: %v", arg)
			return nil, nil
		},
	})

	if err := put(cmd, []string{"-", "/stdin.txt"}); err != nil {
		t.Fatalf("put stdin dry-run error: %v", err)
	}

	got := decodePutOutput(t, &stdout)
	if got.Input.Source != "-" || got.Input.Target != "/stdin.txt" || !got.Input.Stdin || !got.Input.DryRun {
		t.Fatalf("input = %+v, want stdin dry-run", got.Input)
	}
	result := got.Results[0]
	if result.Status != jsonStatusPlanned || result.Kind != putKindFile || !result.Input.DryRun {
		t.Fatalf("result = %+v, want planned file dry-run", result)
	}
	if result.Result == nil || result.Result.PathDisplay != "/stdin.txt" {
		t.Fatalf("metadata = %+v, want planned stdin target metadata", result.Result)
	}
}

func TestPutStdin_UploadsToDashPath(t *testing.T) {
	content := "dash path"
	cmd := testPutCmdWithStdin(strings.NewReader(content))

	var uploadedPath string
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, &files.GetMetadataAPIError{}
		},
		uploadFn: func(arg *files.UploadArg, r io.Reader) (*files.FileMetadata, error) {
			uploadedPath = arg.Path
			_, _ = io.ReadAll(r)
			return &files.FileMetadata{}, nil
		},
	}
	stubFilesClient(t, mock)

	err := put(cmd, []string{"-", "/-"})
	if err != nil {
		t.Fatalf("put stdin error: %v", err)
	}
	if uploadedPath != "/-" {
		t.Errorf("uploaded path = %q, want /-", uploadedPath)
	}
}

func TestPutStdin_EmptyUploadsZeroBytes(t *testing.T) {
	cmd := testPutCmdWithStdin(strings.NewReader(""))

	var uploadedSize int
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, &files.GetMetadataAPIError{}
		},
		uploadFn: func(arg *files.UploadArg, r io.Reader) (*files.FileMetadata, error) {
			data, _ := io.ReadAll(r)
			uploadedSize = len(data)
			return &files.FileMetadata{}, nil
		},
	}
	stubFilesClient(t, mock)

	err := put(cmd, []string{"-", "/empty.txt"})
	if err != nil {
		t.Fatalf("put stdin error: %v", err)
	}
	if uploadedSize != 0 {
		t.Errorf("uploaded size = %d, want 0", uploadedSize)
	}
}

func TestPutStdin_CleanupFailureAfterSuccessfulUpload(t *testing.T) {
	cmd := testPutCmdWithStdin(strings.NewReader("data"))

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, &files.GetMetadataAPIError{}
		},
		uploadFn: func(arg *files.UploadArg, r io.Reader) (*files.FileMetadata, error) {
			_, _ = io.ReadAll(r)
			return &files.FileMetadata{}, nil
		},
	}
	stubFilesClient(t, mock)

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

	origStderr := os.Stderr
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = stderrWriter
	t.Cleanup(func() {
		os.Stderr = origStderr
		_ = stderrReader.Close()
		_ = stderrWriter.Close()
	})

	err = put(cmd, []string{"-", "/dest.txt"})
	_ = stderrWriter.Close()
	stderr, _ := io.ReadAll(stderrReader)
	os.Stderr = origStderr

	if err == nil {
		t.Fatal("expected cleanup failure")
	}
	if !errors.Is(err, cleanupErr) {
		t.Fatalf("error = %v, want wrapped cleanup error", err)
	}
	if !strings.Contains(err.Error(), "sensitive stdin data may remain on disk") {
		t.Errorf("error = %q, want sensitive-data warning", err.Error())
	}
	if tempPath == "" {
		t.Fatal("expected temp path to be captured")
	}
	if !strings.Contains(string(stderr), "failed to remove temp file") || !strings.Contains(string(stderr), cleanupErr.Error()) {
		t.Errorf("stderr = %q, want cleanup failure with underlying error", string(stderr))
	}
}

func TestPutStdin_ReportsCleanupFailureAfterUploadFailure(t *testing.T) {
	cmd := testPutCmdWithStdin(strings.NewReader("data"))
	uploadErr := errors.New("upload failed")

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, &files.GetMetadataAPIError{}
		},
		uploadFn: func(arg *files.UploadArg, r io.Reader) (*files.FileMetadata, error) {
			_, _ = io.ReadAll(r)
			return nil, uploadErr
		},
	}
	stubFilesClient(t, mock)

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

	origStderr := os.Stderr
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = stderrWriter
	t.Cleanup(func() {
		os.Stderr = origStderr
		_ = stderrReader.Close()
		_ = stderrWriter.Close()
	})

	err = put(cmd, []string{"-", "/dest.txt"})
	_ = stderrWriter.Close()
	stderr, _ := io.ReadAll(stderrReader)
	os.Stderr = origStderr

	if !errors.Is(err, uploadErr) {
		t.Fatalf("error = %v, want upload error", err)
	}
	if errors.Is(err, cleanupErr) {
		t.Fatalf("error = %v, should preserve upload error as return value", err)
	}
	if tempPath == "" {
		t.Fatal("expected temp path to be captured")
	}
	if !strings.Contains(string(stderr), "failed to remove temp file") ||
		!strings.Contains(string(stderr), cleanupErr.Error()) ||
		!strings.Contains(string(stderr), "sensitive stdin data may remain on disk") {
		t.Errorf("stderr = %q, want cleanup failure warning", string(stderr))
	}
}

func TestPutLocalDashFile(t *testing.T) {
	// put ./- /remote uploads a local file named "-"
	tmpDir := t.TempDir()
	dashFile := tmpDir + "/-"
	if err := writeTestFile(dashFile, "local dash file"); err != nil {
		t.Fatal(err)
	}

	var uploadedPath string
	origConfig := config
	defer func() { config = origConfig }()
	config = testConfig()

	testClient := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return nil, &files.GetMetadataAPIError{}
		},
		uploadFn: func(arg *files.UploadArg, r io.Reader) (*files.FileMetadata, error) {
			uploadedPath = arg.Path
			_, _ = io.ReadAll(r)
			return &files.FileMetadata{}, nil
		},
	}
	origNew := filesNewFunc
	filesNewFunc = func(_ dropbox.Config) filesClient { return testClient }
	defer func() { filesNewFunc = origNew }()

	cmd := testPutCmd()
	err := put(cmd, []string{dashFile, "/remote"})
	if err != nil {
		t.Fatalf("put ./ - error: %v", err)
	}
	if uploadedPath != "/remote" {
		t.Errorf("uploaded path = %q, want /remote", uploadedPath)
	}
}

// --- get stdout tests ---

func testGetCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().BoolP("recursive", "r", false, "")
	return cmd
}

func TestGetStdout_WritesContent(t *testing.T) {
	content := "file content here"
	var stdout bytes.Buffer

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return &files.FileMetadata{
				Metadata: files.Metadata{PathDisplay: arg.Path},
				Size:     uint64(len(content)),
			}, nil
		},
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			meta := &files.FileMetadata{
				Metadata: files.Metadata{PathDisplay: arg.Path},
				Size:     uint64(len(content)),
			}
			return meta, io.NopCloser(strings.NewReader(content)), nil
		},
	}
	stubFilesClient(t, mock)

	cmd := testGetCmd()
	cmd.SetOut(&stdout)

	err := get(cmd, []string{"/file.txt", "-"})
	if err != nil {
		t.Fatalf("get stdout error: %v", err)
	}
	if stdout.String() != content {
		t.Errorf("stdout = %q, want %q", stdout.String(), content)
	}
}

func TestGetStdout_DashPathDownloads(t *testing.T) {
	content := "dash content"
	var stdout bytes.Buffer

	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return &files.FileMetadata{
				Metadata: files.Metadata{PathDisplay: arg.Path},
				Size:     uint64(len(content)),
			}, nil
		},
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			if arg.Path != "/-" {
				t.Errorf("download path = %q, want /-", arg.Path)
			}
			meta := &files.FileMetadata{
				Metadata: files.Metadata{PathDisplay: arg.Path},
				Size:     uint64(len(content)),
			}
			return meta, io.NopCloser(strings.NewReader(content)), nil
		},
	}
	stubFilesClient(t, mock)

	cmd := testGetCmd()
	cmd.SetOut(&stdout)

	err := get(cmd, []string{"/-", "-"})
	if err != nil {
		t.Fatalf("get /- - error: %v", err)
	}
	if stdout.String() != content {
		t.Errorf("stdout = %q, want %q", stdout.String(), content)
	}
}

func TestGetStdout_RejectsFolder(t *testing.T) {
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return &files.FolderMetadata{Metadata: files.Metadata{PathDisplay: arg.Path}}, nil
		},
	}
	stubFilesClient(t, mock)

	cmd := testGetCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := get(cmd, []string{"/folder", "-"})
	if err == nil {
		t.Fatal("expected error for folder download to stdout")
	}
	if !strings.Contains(err.Error(), "folder") {
		t.Errorf("error = %q, want mention of folder", err.Error())
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty, got %q", stdout.String())
	}
}

func TestGetStdout_RejectsRecursive(t *testing.T) {
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return &files.FolderMetadata{Metadata: files.Metadata{PathDisplay: arg.Path}}, nil
		},
	}
	stubFilesClient(t, mock)

	cmd := testGetCmd()
	_ = cmd.Flags().Set("recursive", "true")
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := get(cmd, []string{"/folder", "-"})
	if err == nil {
		t.Fatal("expected error for recursive download to stdout")
	}
	if !strings.Contains(err.Error(), "--recursive") {
		t.Errorf("error = %q, want mention of --recursive", err.Error())
	}
}

func TestGetStdout_NoLocalFilesCreated(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	content := "file data"
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return &files.FileMetadata{
				Metadata: files.Metadata{PathDisplay: arg.Path},
				Size:     uint64(len(content)),
			}, nil
		},
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			meta := &files.FileMetadata{
				Metadata: files.Metadata{PathDisplay: arg.Path},
				Size:     uint64(len(content)),
			}
			return meta, io.NopCloser(strings.NewReader(content)), nil
		},
	}
	stubFilesClient(t, mock)

	cmd := testGetCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := get(cmd, []string{"/file.txt", "-"})
	if err != nil {
		t.Fatalf("get stdout error: %v", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("expected no local files created, found: %v", names)
	}
}

func writeTestFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
