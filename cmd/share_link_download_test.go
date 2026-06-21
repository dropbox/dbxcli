// Copyright © 2016 Dropbox, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

func TestShareLinkDownloadRequiresURLAndOptionalTarget(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing URL", args: nil},
		{name: "too many args", args: []string{"https://example.com/one", "target", "extra"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			stubSharedLinkClient(t, &mockSharedLinkClient{
				getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
					called = true
					return nil, nil, nil
				},
			})

			err := shareLinkDownload(newShareLinkDownloadTestCommand(nil, nil), tt.args)
			if err == nil || !strings.Contains(err.Error(), "requires a `url` and optional `target` argument") {
				t.Fatalf("error = %v, want url/target argument error", err)
			}
			if called {
				t.Fatal("GetSharedLinkFile should not be called")
			}
		})
	}
}

func TestShareLinkDownloadRejectsEmptyURL(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			called = true
			return nil, nil, nil
		},
	})

	err := shareLinkDownload(newShareLinkDownloadTestCommand(nil, nil), []string{""})
	if err == nil || !strings.Contains(err.Error(), "requires a non-empty URL") {
		t.Fatalf("error = %v, want non-empty URL error", err)
	}
	if called {
		t.Fatal("GetSharedLinkFile should not be called")
	}
}

func TestShareLinkDownloadRejectsEmptyTarget(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			called = true
			return nil, nil, nil
		},
	})

	err := shareLinkDownload(newShareLinkDownloadTestCommand(nil, nil), []string{"https://example.com/link", ""})
	if err == nil || !strings.Contains(err.Error(), "requires a non-empty target") {
		t.Fatalf("error = %v, want non-empty target error", err)
	}
	if called {
		t.Fatal("GetSharedLinkFile should not be called")
	}
}

func TestShareLinkDownloadUsesMetadataNameAndPassword(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	content := "shared content"
	var requested *sharing.GetSharedLinkMetadataArg
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			requested = arg
			return downloadableSharedLinkFile("report.txt", "/docs/report.txt", "https://example.com/link", uint64(len(content))),
				io.NopCloser(strings.NewReader(content)), nil
		},
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newShareLinkDownloadTestCommand(&stdout, &stderr)
	if err := cmd.Flags().Set("password", "secret"); err != nil {
		t.Fatalf("set password: %v", err)
	}

	if err := shareLinkDownload(cmd, []string{"https://example.com/link"}); err != nil {
		t.Fatalf("shareLinkDownload error: %v", err)
	}

	if requested == nil {
		t.Fatal("GetSharedLinkFile was not called")
	}
	if requested.Url != "https://example.com/link" {
		t.Fatalf("url = %q, want https://example.com/link", requested.Url)
	}
	if requested.LinkPassword != "secret" {
		t.Fatalf("password = %q, want secret", requested.LinkPassword)
	}
	assertFileContent(t, filepath.Join(tmp, "report.txt"), content)
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if got := stderr.String(); !strings.Contains(got, "Downloading ") {
		t.Fatalf("stderr = %q, want progress", got)
	}
}

func TestShareLinkDownloadUsesExplicitTarget(t *testing.T) {
	tmp := t.TempDir()
	content := "shared content"
	target := filepath.Join(tmp, "local.txt")

	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			return downloadableSharedLinkFile("report.txt", "/docs/report.txt", "https://example.com/link", uint64(len(content))),
				io.NopCloser(strings.NewReader(content)), nil
		},
	})

	if err := shareLinkDownload(newShareLinkDownloadTestCommand(nil, nil), []string{"https://example.com/link", target}); err != nil {
		t.Fatalf("shareLinkDownload error: %v", err)
	}
	assertFileContent(t, target, content)
}

func TestShareLinkDownloadUsesTargetDirectory(t *testing.T) {
	tmp := t.TempDir()
	targetDir := filepath.Join(tmp, "downloads")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	content := "shared content"
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			return downloadableSharedLinkFile("report.txt", "/docs/report.txt", "https://example.com/link", uint64(len(content))),
				io.NopCloser(strings.NewReader(content)), nil
		},
	})

	if err := shareLinkDownload(newShareLinkDownloadTestCommand(nil, nil), []string{"https://example.com/link", targetDir}); err != nil {
		t.Fatalf("shareLinkDownload error: %v", err)
	}
	assertFileContent(t, filepath.Join(targetDir, "report.txt"), content)
}

func TestShareLinkDownloadToStdoutIsByteClean(t *testing.T) {
	content := "shared stdout content"
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			return downloadableSharedLinkFile("report.txt", "/docs/report.txt", "https://example.com/link", uint64(len(content))),
				io.NopCloser(strings.NewReader(content)), nil
		},
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newShareLinkDownloadTestCommand(&stdout, &stderr)

	if err := shareLinkDownload(cmd, []string{"https://example.com/link", "-"}); err != nil {
		t.Fatalf("shareLinkDownload error: %v", err)
	}
	if stdout.String() != content {
		t.Fatalf("stdout = %q, want file bytes", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty without verbose", stderr.String())
	}
}

func TestShareLinkDownloadToStdoutDoesNotRetryAfterPartialOutput(t *testing.T) {
	retryDelays := stubRetrySleep(t)
	calls := 0
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			calls++
			return downloadableSharedLinkFile("report.txt", "/docs/report.txt", "https://example.com/link", 100),
				&failingReadCloser{data: []byte("partial")}, nil
		},
	})

	var stdout bytes.Buffer
	err := shareLinkDownload(newShareLinkDownloadTestCommand(&stdout, nil), []string{"https://example.com/link", "-"})
	if err == nil {
		t.Fatal("expected error for partial stdout failure")
	}
	if !strings.Contains(err.Error(), "cannot retry") {
		t.Fatalf("error = %q, want cannot retry", err.Error())
	}
	if calls != 1 {
		t.Fatalf("GetSharedLinkFile calls = %d, want 1", calls)
	}
	if len(*retryDelays) != 0 {
		t.Fatalf("retry delays = %v, want none", *retryDelays)
	}
	if stdout.String() != "partial" {
		t.Fatalf("stdout = %q, want partial bytes", stdout.String())
	}
}

func TestShareLinkDownloadToStdoutBrokenPipeReturnsNil(t *testing.T) {
	mock := &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			return downloadableSharedLinkFile("report.txt", "/docs/report.txt", "https://example.com/link", 100),
				io.NopCloser(strings.NewReader("some data")), nil
		},
	}

	err := downloadSharedLinkToStdout(mock, sharing.NewGetSharedLinkMetadataArg("https://example.com/link"), epipeWriter{})
	if err != nil {
		t.Fatalf("downloadSharedLinkToStdout error: %v", err)
	}
}

func TestShareLinkDownloadReturnsAPIErrors(t *testing.T) {
	wantErr := fmt.Errorf("shared_link_not_found")
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			return nil, nil, wantErr
		},
	})

	err := shareLinkDownload(newShareLinkDownloadTestCommand(nil, nil), []string{"https://example.com/link", "-"})
	if err != wantErr {
		t.Fatalf("error = %v, want original API error", err)
	}
}

func TestShareLinkDownloadRejectsMissingContent(t *testing.T) {
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			return downloadableSharedLinkFile("report.txt", "/docs/report.txt", "https://example.com/link", 0), nil, nil
		},
	})

	err := shareLinkDownload(newShareLinkDownloadTestCommand(nil, nil), []string{"https://example.com/link", filepath.Join(t.TempDir(), "target")})
	if err == nil || !strings.Contains(err.Error(), "did not include file content") {
		t.Fatalf("error = %v, want missing content error", err)
	}
}

func TestShareLinkDownloadToStdoutRejectsMissingContent(t *testing.T) {
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			return downloadableSharedLinkFile("report.txt", "/docs/report.txt", "https://example.com/link", 0), nil, nil
		},
	})

	var stdout bytes.Buffer
	err := shareLinkDownload(newShareLinkDownloadTestCommand(&stdout, nil), []string{"https://example.com/link", "-"})
	if err == nil || !strings.Contains(err.Error(), "did not include file content") {
		t.Fatalf("error = %v, want missing content error", err)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestShareLinkDownloadRejectsNonFileMetadata(t *testing.T) {
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			return sharedLinkFolder("/docs", "https://example.com/link"), io.NopCloser(strings.NewReader("content")), nil
		},
	})

	err := shareLinkDownload(newShareLinkDownloadTestCommand(nil, nil), []string{"https://example.com/link", filepath.Join(t.TempDir(), "target")})
	if err == nil || !strings.Contains(err.Error(), "not a downloadable file") {
		t.Fatalf("error = %v, want non-file error", err)
	}
}

func TestShareLinkDownloadRejectsUnsafeMetadataName(t *testing.T) {
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			return downloadableSharedLinkFile("..", "/docs/..", "https://example.com/link", 7),
				io.NopCloser(strings.NewReader("content")), nil
		},
	})

	err := shareLinkDownload(newShareLinkDownloadTestCommand(nil, nil), []string{"https://example.com/link", t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "did not include a name") {
		t.Fatalf("error = %v, want invalid name error", err)
	}
}

func TestShareLinkDownloadVerboseWritesStatusToStderr(t *testing.T) {
	content := "shared stdout content"
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			return downloadableSharedLinkFile("report.txt", "/docs/report.txt", "https://example.com/link", uint64(len(content))),
				io.NopCloser(strings.NewReader(content)), nil
		},
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newShareLinkDownloadTestCommand(&stdout, &stderr)
	if err := cmd.Flags().Set("verbose", "true"); err != nil {
		t.Fatalf("set verbose: %v", err)
	}

	if err := shareLinkDownload(cmd, []string{"https://example.com/link", "-"}); err != nil {
		t.Fatalf("shareLinkDownload error: %v", err)
	}
	if stdout.String() != content {
		t.Fatalf("stdout = %q, want file bytes", stdout.String())
	}
	if got := stderr.String(); got != "Downloaded shared link to stdout\n" {
		t.Fatalf("stderr = %q, want verbose status", got)
	}
}

func TestShareLinkDownloadCommandIsRegistered(t *testing.T) {
	cmd, _, err := RootCmd.Find([]string{"share-link", "download", "https://example.com/link"})
	if err != nil {
		t.Fatalf("find share-link download: %v", err)
	}
	if cmd != shareLinkDownloadCmd {
		t.Fatalf("share-link download resolved to %q", cmd.CommandPath())
	}
}

func newShareLinkDownloadTestCommand(stdout, stderr *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("password", "", "")
	cmd.Flags().Bool("verbose", false, "")
	if stdout != nil {
		cmd.SetOut(stdout)
	}
	if stderr != nil {
		cmd.SetErr(stderr)
	}
	return cmd
}

func downloadableSharedLinkFile(name string, pathLower string, url string, size uint64) *sharing.FileLinkMetadata {
	link := sharing.NewFileLinkMetadata(url, name, nil, time.Time{}, time.Time{}, "rev", size)
	link.PathLower = strings.ToLower(pathLower)
	return link
}

func assertFileContent(t *testing.T, path string, want string) {
	t.Helper()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(got) != want {
		t.Fatalf("%s = %q, want %q", path, string(got), want)
	}
}
