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

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
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

func TestShareLinkDownloadReadsPasswordPrompt(t *testing.T) {
	orig := readSharedLinkPassword
	readSharedLinkPassword = func(prompt string, in io.Reader, errOut io.Writer) (string, error) {
		if prompt != "Shared link password: " {
			t.Fatalf("prompt = %q, want shared link password prompt", prompt)
		}
		return "prompt-secret", nil
	}
	t.Cleanup(func() { readSharedLinkPassword = orig })

	content := "shared content"
	var requested *sharing.GetSharedLinkMetadataArg
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			requested = arg
			return downloadableSharedLinkFile("report.txt", "/docs/report.txt", "https://example.com/link", uint64(len(content))),
				io.NopCloser(strings.NewReader(content)), nil
		},
	})

	target := filepath.Join(t.TempDir(), "report.txt")
	cmd := newShareLinkDownloadTestCommand(nil, nil)
	if err := cmd.Flags().Set("password-prompt", "true"); err != nil {
		t.Fatalf("set password-prompt: %v", err)
	}

	if err := shareLinkDownload(cmd, []string{"https://example.com/link", target}); err != nil {
		t.Fatalf("shareLinkDownload error: %v", err)
	}
	if requested == nil {
		t.Fatal("GetSharedLinkFile was not called")
	}
	if requested.LinkPassword != "prompt-secret" {
		t.Fatalf("password = %q, want prompt-secret", requested.LinkPassword)
	}
}

func TestShareLinkDownloadReadsPasswordFile(t *testing.T) {
	passwordFile := filepath.Join(t.TempDir(), "password.txt")
	if err := os.WriteFile(passwordFile, []byte("file-secret\n"), 0600); err != nil {
		t.Fatalf("write password file: %v", err)
	}

	content := "shared content"
	var requested *sharing.GetSharedLinkMetadataArg
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			requested = arg
			return downloadableSharedLinkFile("report.txt", "/docs/report.txt", "https://example.com/link", uint64(len(content))),
				io.NopCloser(strings.NewReader(content)), nil
		},
	})

	target := filepath.Join(t.TempDir(), "report.txt")
	cmd := newShareLinkDownloadTestCommand(nil, nil)
	if err := cmd.Flags().Set("password-file", passwordFile); err != nil {
		t.Fatalf("set password-file: %v", err)
	}

	if err := shareLinkDownload(cmd, []string{"https://example.com/link", target}); err != nil {
		t.Fatalf("shareLinkDownload error: %v", err)
	}
	if requested == nil {
		t.Fatal("GetSharedLinkFile was not called")
	}
	if requested.LinkPassword != "file-secret" {
		t.Fatalf("password = %q, want file-secret", requested.LinkPassword)
	}
}

func TestShareLinkDownloadRejectsMultiplePasswordSources(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			called = true
			return nil, nil, nil
		},
	})

	cmd := newShareLinkDownloadTestCommand(nil, nil)
	if err := cmd.Flags().Set("password", "secret"); err != nil {
		t.Fatalf("set password: %v", err)
	}
	if err := cmd.Flags().Set("password-file", filepath.Join(t.TempDir(), "password.txt")); err != nil {
		t.Fatalf("set password-file: %v", err)
	}

	err := shareLinkDownload(cmd, []string{"https://example.com/link", filepath.Join(t.TempDir(), "target")})
	if err == nil || !strings.Contains(err.Error(), "use only one of `--password`, `--password-prompt`, or `--password-file`") {
		t.Fatalf("error = %v, want password source error", err)
	}
	if called {
		t.Fatal("GetSharedLinkFile should not be called")
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

func TestShareLinkDownloadPathDownloadsNestedFile(t *testing.T) {
	tmp := t.TempDir()
	targetDir := filepath.Join(tmp, "downloads")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	content := "nested content"
	var requested *sharing.GetSharedLinkMetadataArg
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
			t.Fatal("GetSharedLinkMetadata should not be called for --path file downloads")
			return nil, nil
		},
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			requested = arg
			return downloadableSharedLinkFile("nested.txt", "/docs/sub/nested.txt", "https://example.com/folder", uint64(len(content))),
				io.NopCloser(strings.NewReader(content)), nil
		},
	})

	cmd := newShareLinkDownloadTestCommand(nil, nil)
	if err := cmd.Flags().Set("path", "sub/nested.txt"); err != nil {
		t.Fatalf("set path: %v", err)
	}
	if err := cmd.Flags().Set("password", "secret"); err != nil {
		t.Fatalf("set password: %v", err)
	}

	if err := shareLinkDownload(cmd, []string{"https://example.com/folder", targetDir}); err != nil {
		t.Fatalf("shareLinkDownload error: %v", err)
	}

	if requested == nil {
		t.Fatal("GetSharedLinkFile was not called")
	}
	if requested.Url != "https://example.com/folder" {
		t.Fatalf("url = %q, want https://example.com/folder", requested.Url)
	}
	if requested.Path != "/sub/nested.txt" {
		t.Fatalf("path = %q, want /sub/nested.txt", requested.Path)
	}
	if requested.LinkPassword != "secret" {
		t.Fatalf("password = %q, want secret", requested.LinkPassword)
	}
	assertFileContent(t, filepath.Join(targetDir, "nested.txt"), content)
}

func TestShareLinkDownloadPathToStdoutIsByteClean(t *testing.T) {
	content := "nested stdout content"
	var requested *sharing.GetSharedLinkMetadataArg
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			requested = arg
			return downloadableSharedLinkFile("nested.txt", "/docs/sub/nested.txt", "https://example.com/folder", uint64(len(content))),
				io.NopCloser(strings.NewReader(content)), nil
		},
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newShareLinkDownloadTestCommand(&stdout, &stderr)
	if err := cmd.Flags().Set("path", "/sub/nested.txt"); err != nil {
		t.Fatalf("set path: %v", err)
	}

	if err := shareLinkDownload(cmd, []string{"https://example.com/folder", "-"}); err != nil {
		t.Fatalf("shareLinkDownload error: %v", err)
	}
	if requested == nil {
		t.Fatal("GetSharedLinkFile was not called")
	}
	if requested.Path != "/sub/nested.txt" {
		t.Fatalf("path = %q, want /sub/nested.txt", requested.Path)
	}
	if stdout.String() != content {
		t.Fatalf("stdout = %q, want file bytes", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty without verbose", stderr.String())
	}
}

func TestShareLinkDownloadPathRejectsInvalidCombinations(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		recursive bool
		want      string
	}{
		{name: "empty path", path: "", want: "`--path` requires a non-empty path"},
		{name: "root path", path: "/", want: "cannot download shared-link root with `--path`"},
		{name: "recursive path", path: "/sub/nested.txt", recursive: true, want: "`--path` cannot be used with --recursive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			stubSharedLinkClient(t, &mockSharedLinkClient{
				getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
					called = true
					return nil, nil, nil
				},
				getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
					called = true
					return nil, nil
				},
			})

			cmd := newShareLinkDownloadTestCommand(nil, nil)
			if err := cmd.Flags().Set("path", tt.path); err != nil {
				t.Fatalf("set path: %v", err)
			}
			if tt.recursive {
				if err := cmd.Flags().Set("recursive", "true"); err != nil {
					t.Fatalf("set recursive: %v", err)
				}
			}

			err := shareLinkDownload(cmd, []string{"https://example.com/folder", filepath.Join(t.TempDir(), "target")})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
			if called {
				t.Fatal("shared link API should not be called")
			}
		})
	}
}

func TestShareLinkDownloadFolderRequiresRecursive(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
			return sharedLinkFolder("/docs", "https://example.com/folder"), nil
		},
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			called = true
			return nil, nil, nil
		},
	})

	err := shareLinkDownload(newShareLinkDownloadTestCommand(nil, nil), []string{"https://example.com/folder", filepath.Join(t.TempDir(), "target")})
	if err == nil || !strings.Contains(err.Error(), "--recursive") {
		t.Fatalf("error = %v, want recursive error", err)
	}
	if called {
		t.Fatal("GetSharedLinkFile should not be called")
	}
}

func TestShareLinkDownloadFolderRejectsStdoutTarget(t *testing.T) {
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
			return sharedLinkFolder("/docs", "https://example.com/folder"), nil
		},
	})

	cmd := newShareLinkDownloadTestCommand(nil, nil)
	if err := cmd.Flags().Set("recursive", "true"); err != nil {
		t.Fatalf("set recursive: %v", err)
	}

	err := shareLinkDownload(cmd, []string{"https://example.com/folder", "-"})
	if err == nil || !strings.Contains(err.Error(), "stdout") {
		t.Fatalf("error = %v, want stdout folder error", err)
	}
}

func TestShareLinkDownloadFolderRecursiveDownloadsNestedFiles(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "out")
	var listed []string
	var listedPassword string
	var downloaded []string

	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
			if arg.LinkPassword != "secret" {
				t.Fatalf("metadata password = %q, want secret", arg.LinkPassword)
			}
			return sharedLinkFolder("/docs", "https://example.com/folder"), nil
		},
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			downloaded = append(downloaded, arg.Path)
			if arg.LinkPassword != "secret" {
				t.Fatalf("download password = %q, want secret", arg.LinkPassword)
			}
			contents := strings.TrimPrefix(arg.Path, "/")
			return downloadableSharedLinkFile(filepath.Base(arg.Path), arg.Path, "https://example.com/folder", uint64(len(contents))),
				io.NopCloser(strings.NewReader(contents)), nil
		},
	})
	stubFilesClient(t, &mockFilesClient{
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			listed = append(listed, arg.Path)
			if arg.SharedLink == nil {
				t.Fatal("SharedLink = nil, want shared-link listing")
			}
			if arg.SharedLink.Url != "https://example.com/folder" {
				t.Fatalf("shared link URL = %q, want https://example.com/folder", arg.SharedLink.Url)
			}
			listedPassword = arg.SharedLink.Password
			if arg.Recursive {
				t.Fatal("Recursive = true, want manual recursion for shared links")
			}
			switch arg.Path {
			case "":
				return &files.ListFolderResult{
					Entries: []files.IsMetadata{
						&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/docs/root.txt"}, Size: 8},
						&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/docs/sub"}},
					},
				}, nil
			case "/sub":
				return &files.ListFolderResult{
					Entries: []files.IsMetadata{
						&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/docs/sub/deep.txt"}, Size: 12},
					},
				}, nil
			default:
				t.Fatalf("unexpected list path %q", arg.Path)
			}
			return nil, nil
		},
	})

	cmd := newShareLinkDownloadTestCommand(nil, nil)
	if err := cmd.Flags().Set("recursive", "true"); err != nil {
		t.Fatalf("set recursive: %v", err)
	}
	if err := cmd.Flags().Set("password", "secret"); err != nil {
		t.Fatalf("set password: %v", err)
	}

	if err := shareLinkDownload(cmd, []string{"https://example.com/folder", target}); err != nil {
		t.Fatalf("shareLinkDownload error: %v", err)
	}

	if strings.Join(listed, ",") != ",/sub" {
		t.Fatalf("listed paths = %q, want root then sub", strings.Join(listed, ","))
	}
	if listedPassword != "secret" {
		t.Fatalf("listed password = %q, want secret", listedPassword)
	}
	if strings.Join(downloaded, ",") != "/root.txt,/sub/deep.txt" {
		t.Fatalf("downloaded paths = %q, want root and nested file", strings.Join(downloaded, ","))
	}
	assertFileContent(t, filepath.Join(target, "root.txt"), "root.txt")
	assertFileContent(t, filepath.Join(target, "sub", "deep.txt"), "sub/deep.txt")
}

func TestShareLinkDownloadFolderUsesExistingTargetDirectory(t *testing.T) {
	tmp := t.TempDir()
	targetDir := filepath.Join(tmp, "downloads")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
			return sharedLinkFolder("/docs", "https://example.com/folder"), nil
		},
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			return downloadableSharedLinkFile("root.txt", arg.Path, "https://example.com/folder", 4),
				io.NopCloser(strings.NewReader("data")), nil
		},
	})
	stubFilesClient(t, &mockFilesClient{
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/docs/root.txt"}, Size: 4},
				},
			}, nil
		},
	})

	cmd := newShareLinkDownloadTestCommand(nil, nil)
	if err := cmd.Flags().Set("recursive", "true"); err != nil {
		t.Fatalf("set recursive: %v", err)
	}

	if err := shareLinkDownload(cmd, []string{"https://example.com/folder", targetDir}); err != nil {
		t.Fatalf("shareLinkDownload error: %v", err)
	}
	assertFileContent(t, filepath.Join(targetDir, "docs", "root.txt"), "data")
}

func TestShareLinkDownloadFolderFollowsPagination(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "out")
	var continued string

	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
			return sharedLinkFolder("/docs", "https://example.com/folder"), nil
		},
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			return downloadableSharedLinkFile(filepath.Base(arg.Path), arg.Path, "https://example.com/folder", 4),
				io.NopCloser(strings.NewReader("data")), nil
		},
	})
	stubFilesClient(t, &mockFilesClient{
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/docs/page-one.txt"}, Size: 4},
				},
				HasMore: true,
				Cursor:  "cursor-2",
			}, nil
		},
		listFolderContinueFn: func(arg *files.ListFolderContinueArg) (*files.ListFolderResult, error) {
			continued = arg.Cursor
			return &files.ListFolderResult{
				Entries: []files.IsMetadata{
					&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/docs/page-two.txt"}, Size: 4},
				},
			}, nil
		},
	})

	cmd := newShareLinkDownloadTestCommand(nil, nil)
	if err := cmd.Flags().Set("recursive", "true"); err != nil {
		t.Fatalf("set recursive: %v", err)
	}

	if err := shareLinkDownload(cmd, []string{"https://example.com/folder", target}); err != nil {
		t.Fatalf("shareLinkDownload error: %v", err)
	}
	if continued != "cursor-2" {
		t.Fatalf("continue cursor = %q, want cursor-2", continued)
	}
	assertFileContent(t, filepath.Join(target, "page-one.txt"), "data")
	assertFileContent(t, filepath.Join(target, "page-two.txt"), "data")
}

func TestShareLinkDownloadFolderDoesNotCreateTargetWhenRootListFails(t *testing.T) {
	target := filepath.Join(t.TempDir(), "out")
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
			return sharedLinkFolder("/docs", "https://example.com/folder"), nil
		},
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			t.Fatal("GetSharedLinkFile should not be called when root listing fails")
			return nil, nil, nil
		},
	})
	stubFilesClient(t, &mockFilesClient{
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			return nil, fmt.Errorf("root list failed")
		},
	})

	cmd := newShareLinkDownloadTestCommand(nil, nil)
	if err := cmd.Flags().Set("recursive", "true"); err != nil {
		t.Fatalf("set recursive: %v", err)
	}

	err := shareLinkDownload(cmd, []string{"https://example.com/folder", target})
	if err == nil || !strings.Contains(err.Error(), "root list failed") {
		t.Fatalf("error = %v, want root list failure", err)
	}
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Fatalf("target stat error = %v, want target not created", statErr)
	}
}

func TestSharedLinkEntryRelativePathStripsRootBySegment(t *testing.T) {
	tests := []struct {
		name        string
		pathDisplay string
		rootName    string
		want        string
	}{
		{
			name:        "root itself",
			pathDisplay: "/docs",
			rootName:    "docs",
			want:        "",
		},
		{
			name:        "child",
			pathDisplay: "/docs/root.txt",
			rootName:    "docs",
			want:        "root.txt",
		},
		{
			name:        "case-insensitive root segment",
			pathDisplay: "/DOCS/root.txt",
			rootName:    "docs",
			want:        "root.txt",
		},
		{
			name:        "partial prefix is not stripped",
			pathDisplay: "/docs-other/root.txt",
			rootName:    "docs",
			want:        "docs-other/root.txt",
		},
		{
			name:        "unicode root segment",
			pathDisplay: "/\u0130/root.txt",
			rootName:    "\u0130",
			want:        "root.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sharedLinkEntryRelativePath(tt.pathDisplay, tt.rootName)
			if err != nil {
				t.Fatalf("sharedLinkEntryRelativePath error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("relative path = %q, want %q", got, tt.want)
			}
		})
	}
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
	if shareLinkDownloadCmd.Flags().Lookup("password") == nil {
		t.Fatal("share-link download should define --password")
	}
	if shareLinkDownloadCmd.Flags().Lookup("password-prompt") == nil {
		t.Fatal("share-link download should define --password-prompt")
	}
	if shareLinkDownloadCmd.Flags().Lookup("password-file") == nil {
		t.Fatal("share-link download should define --password-file")
	}
	if shareLinkDownloadCmd.Flags().Lookup("path") == nil {
		t.Fatal("share-link download should define --path")
	}
	if shareLinkDownloadCmd.Flags().Lookup("recursive") == nil {
		t.Fatal("share-link download should define --recursive")
	}
}

func newShareLinkDownloadTestCommand(stdout, stderr *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{}
	addSharedLinkPasswordFlags(cmd)
	cmd.Flags().String("path", "", "")
	cmd.Flags().BoolP("recursive", "r", false, "")
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
