// Copyright © 2026 Dropbox, Inc.
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
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
)

func TestShareLinkDownloadSelectedFolder(t *testing.T) {
	schema := compileJSONSchemaFile(t, "../docs/json-schema/v1/commands.schema.json")
	for _, mounted := range []bool{false, true} {
		for _, existing := range []bool{false, true} {
			t.Run(fmt.Sprintf("mounted=%t/existing=%t", mounted, existing), func(t *testing.T) {
				const selected = "/2027/Logoset"
				target := filepath.Join(t.TempDir(), "out")
				dst := target
				if existing {
					if err := os.Mkdir(target, 0755); err != nil {
						t.Fatal(err)
					}
					dst = filepath.Join(target, "Logoset")
				}
				folder := sharedLinkFolder("/Logoset", listingURL)
				folder.PathLower = ""
				if mounted {
					folder.PathLower = "/mounted/2027/logoset"
				}
				var listed, continued, downloaded []string
				stubSharedLinkClient(t, &mockSharedLinkClient{
					getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
						if arg.Url != listingURL || arg.Path != selected || arg.LinkPassword != "secret" {
							t.Fatalf("metadata arg = %#v", arg)
						}
						return folder, nil
					},
					getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
						if arg.Url != listingURL || arg.LinkPassword != "secret" {
							t.Fatalf("download arg = %#v", arg)
						}
						downloaded = append(downloaded, arg.Path)
						return downloadableSharedLinkFile(path.Base(arg.Path), "", arg.Url, uint64(len(arg.Path))), io.NopCloser(strings.NewReader(arg.Path)), nil
					},
				})
				metadata := func(name, parent string) files.Metadata {
					m := files.Metadata{Name: name}
					if mounted {
						m.PathDisplay = "/mounted" + parent + "/" + name
						m.PathLower = strings.ToLower(m.PathDisplay)
					}
					return m
				}
				stubFilesClient(t, &mockFilesClient{
					listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
						if arg.SharedLink == nil || arg.SharedLink.Url != listingURL || arg.SharedLink.Password != "secret" || arg.Recursive {
							t.Fatalf("listing arg = %#v", arg)
						}
						listed = append(listed, arg.Path)
						switch arg.Path {
						case selected:
							return &files.ListFolderResult{Entries: []files.IsMetadata{&files.FileMetadata{Metadata: metadata("logo.png", arg.Path)}}, HasMore: true, Cursor: "selected-next"}, nil
						case selected + "/icons":
							return &files.ListFolderResult{Entries: []files.IsMetadata{&files.FileMetadata{Metadata: metadata("é-資料.svg", arg.Path)}}, HasMore: true, Cursor: "icons-next"}, nil
						case selected + "/icons/deep", selected + "/empty":
							return &files.ListFolderResult{}, nil
						default:
							t.Fatalf("listed outside selected folder: %q", arg.Path)
							return nil, nil
						}
					},
					listFolderContinueFn: func(arg *files.ListFolderContinueArg) (*files.ListFolderResult, error) {
						continued = append(continued, arg.Cursor)
						switch arg.Cursor {
						case "selected-next":
							return &files.ListFolderResult{Entries: []files.IsMetadata{listingFolder("icons"), listingFolder("empty")}}, nil
						case "icons-next":
							return &files.ListFolderResult{Entries: []files.IsMetadata{listingFolder("deep"), listingFile("e\u0301.txt", 1)}}, nil
						default:
							t.Fatalf("unexpected cursor %q", arg.Cursor)
							return nil, nil
						}
					},
				})
				var out, stderr bytes.Buffer
				cmd := newShareLinkDownloadTestCommand(&out, &stderr)
				setShareLinkOutputJSON(t, cmd)
				if err := cmd.ParseFlags([]string{"--path", "//2027//./Logoset/", "-r", "--password", "secret"}); err != nil {
					t.Fatal(err)
				}
				if err := shareLinkDownload(cmd, []string{listingURL, target}); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(listed, []string{selected, selected + "/icons", selected + "/empty", selected + "/icons/deep"}) {
					t.Fatalf("listed = %v", listed)
				}
				if !reflect.DeepEqual(continued, []string{"selected-next", "icons-next"}) {
					t.Fatalf("continued = %v", continued)
				}
				wantFiles := []string{selected + "/logo.png", selected + "/icons/é-資料.svg", selected + "/icons/e\u0301.txt"}
				if !reflect.DeepEqual(downloaded, wantFiles) {
					t.Fatalf("downloaded = %v", downloaded)
				}
				for _, file := range wantFiles {
					assertFileContent(t, filepath.Join(dst, filepath.FromSlash(strings.TrimPrefix(file, selected+"/"))), file)
				}
				for _, dir := range []string{"empty", "icons/deep"} {
					info, err := os.Stat(filepath.Join(dst, filepath.FromSlash(dir)))
					if err != nil || !info.IsDir() {
						t.Fatalf("missing directory %s: %v", dir, err)
					}
				}
				got := decodeShareLinkOperationOutput[shareLinkDownloadInput, shareLinkDownloadResult](t, out.Bytes())
				if got.Input.URL != listingURL || got.Input.Path != selected || got.Input.Target != target || !got.Input.Recursive || !got.Input.Password {
					t.Fatalf("input = %#v", got.Input)
				}
				if len(got.Results) != 1 || got.Results[0].Status != "downloaded" || got.Results[0].Kind != "folder" || got.Results[0].Result.Target != dst || got.Results[0].Result.Link.Name != "Logoset" {
					t.Fatalf("results = %#v", got.Results)
				}
				if strings.Contains(out.String(), "secret") {
					t.Fatal("password leaked")
				}
				if err := schema.Validate(decodeJSONValueForSchema(t, out.Bytes())); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestShareLinkDownloadSelectedEmptyFolderTargets(t *testing.T) {
	for _, targetKind := range []string{"omitted", "new", "existing"} {
		t.Run(targetKind, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			dst := "Empty"
			args := []string{listingURL}
			switch targetKind {
			case "new":
				args = append(args, "out")
				dst = "out"
			case "existing":
				args = append(args, root)
				dst = filepath.Join(root, "Empty")
			}
			stubSharedLinkClient(t, &mockSharedLinkClient{getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
				folder := sharedLinkFolder("/Empty", arg.Url)
				folder.PathLower = ""
				return folder, nil
			}})
			stubFilesClient(t, &mockFilesClient{listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
				if arg.Path != "/nested/Empty" {
					t.Fatalf("listing = %q", arg.Path)
				}
				return &files.ListFolderResult{}, nil
			}})
			cmd := newShareLinkDownloadTestCommand(nil, nil)
			if err := cmd.ParseFlags([]string{"--path", "/nested/Empty", "-r"}); err != nil {
				t.Fatal(err)
			}
			if err := shareLinkDownload(cmd, args); err != nil {
				t.Fatal(err)
			}
			entries, err := os.ReadDir(dst)
			if err != nil || len(entries) != 0 {
				t.Fatalf("empty folder = %v, %v", entries, err)
			}
		})
	}
}

func TestShareLinkDownloadSelectedFileModes(t *testing.T) {
	schema := compileJSONSchemaFile(t, "../docs/json-schema/v1/commands.schema.json")
	for _, recursive := range []bool{false, true} {
		for _, stdout := range []bool{false, true} {
			t.Run(fmt.Sprintf("recursive=%t/stdout=%t", recursive, stdout), func(t *testing.T) {
				calls := 0
				stubSharedLinkClient(t, &mockSharedLinkClient{
					getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
						if arg.Path != "/nested/file.txt" {
							t.Fatalf("metadata arg = %#v", arg)
						}
						return downloadableSharedLinkFile("file.txt", "", arg.Url, 4), nil
					},
					getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
						calls++
						if arg.Path != "/nested/file.txt" {
							t.Fatalf("download arg = %#v", arg)
						}
						return downloadableSharedLinkFile("file.txt", "", arg.Url, 4), io.NopCloser(strings.NewReader("data")), nil
					},
				})
				target := filepath.Join(t.TempDir(), "file.txt")
				if stdout {
					target = "-"
				}
				var out, stderr bytes.Buffer
				cmd := newShareLinkDownloadTestCommand(&out, &stderr)
				if !stdout {
					setShareLinkOutputJSON(t, cmd)
				}
				if err := cmd.ParseFlags([]string{"--path", "/nested/file.txt", fmt.Sprintf("--recursive=%t", recursive)}); err != nil {
					t.Fatal(err)
				}
				err := shareLinkDownload(cmd, []string{listingURL, target})
				if stdout && recursive {
					if err == nil || !strings.Contains(err.Error(), "--recursive") || calls != 0 || out.Len() != 0 {
						t.Fatalf("err=%v calls=%d out=%s", err, calls, &out)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if calls != 1 {
					t.Fatalf("download calls=%d", calls)
				}
				if stdout {
					if out.String() != "data" || stderr.Len() != 0 {
						t.Fatalf("out=%s stderr=%s", &out, &stderr)
					}
					return
				}
				assertFileContent(t, target, "data")
				got := decodeShareLinkOperationOutput[shareLinkDownloadInput, shareLinkDownloadResult](t, out.Bytes())
				if got.Input.Path != "/nested/file.txt" || got.Input.Recursive != recursive || got.Results[0].Kind != "file" {
					t.Fatalf("output=%#v", got)
				}
				if err := schema.Validate(decodeJSONValueForSchema(t, out.Bytes())); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestShareLinkDownloadSelectedFolderRestrictions(t *testing.T) {
	schema := compileJSONSchemaFile(t, "../docs/json-schema/v1/error.schema.json")
	for _, mode := range []string{"recursive required", "stdout", "JSON stdout", "unknown metadata"} {
		t.Run(mode, func(t *testing.T) {
			called := false
			stubSharedLinkClient(t, &mockSharedLinkClient{
				getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
					if mode == "unknown metadata" {
						return &sharing.SharedLinkMetadata{}, nil
					}
					return sharedLinkFolder("/folder", arg.Url), nil
				},
				getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
					called = true
					return nil, nil, nil
				},
			})
			stubFilesClient(t, &mockFilesClient{listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) { called = true; return nil, nil }})
			var out bytes.Buffer
			cmd := newShareLinkDownloadTestCommand(&out, nil)
			flags := []string{"--path", "/nested/folder"}
			target := filepath.Join(t.TempDir(), "out")
			want := "--recursive"
			if mode != "recursive required" {
				flags = append(flags, "-r")
			}
			if strings.Contains(mode, "stdout") {
				target = "-"
				want = "stdout"
			}
			if mode == "JSON stdout" {
				setShareLinkOutputJSON(t, cmd)
				want = "--output=json"
			}
			if mode == "unknown metadata" {
				want = "not a downloadable file or folder"
			}
			if err := cmd.ParseFlags(flags); err != nil {
				t.Fatal(err)
			}
			err := shareLinkDownload(cmd, []string{listingURL, target})
			if err == nil || !strings.Contains(err.Error(), want) || called || out.Len() != 0 {
				t.Fatalf("err=%v called=%t out=%s", err, called, &out)
			}
			if err := schema.Validate(normalizeJSONValueForSchema(t, newJSONErrorResponse(cmd, err))); err != nil {
				t.Fatalf("download error violates JSON v1 schema: %v", err)
			}
			details := jsonErrorDetails(err)
			if details["url"] != listingURL || details["path"] != "/nested/folder" || details["operation"] != "share_link_download" {
				t.Fatalf("details=%#v", details)
			}
		})
	}
}

func TestShareLinkDownloadSelectedFolderFailures(t *testing.T) {
	schema := compileJSONSchemaFile(t, "../docs/json-schema/v1/error.schema.json")
	for _, stage := range []string{"metadata", "password required", "incorrect password", "expired", "not found", "list", "nested list", "continue", "nested continue", "missing cursor", "unknown entry", "nil file", "nil folder", "download", "local", "canceled", "cancel listing", "cancel pagination", "cancel download"} {
		t.Run(stage, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			old := commandContext
			commandContext = ctx
			t.Cleanup(func() { commandContext = old })
			want := errors.New(stage)
			if stage == "canceled" || strings.HasPrefix(stage, "cancel ") {
				want = context.Canceled
			}
			if stage == "canceled" {
				cancel()
			}
			target := filepath.Join(t.TempDir(), "out")
			if stage == "local" {
				if err := os.WriteFile(target, []byte("existing"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			stubSharedLinkClient(t, &mockSharedLinkClient{
				getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
					switch stage {
					case "metadata", "password required", "incorrect password", "expired", "not found":
						return nil, want
					}
					return sharedLinkFolder("/folder", arg.Url), nil
				},
				getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
					if strings.HasSuffix(arg.Path, "/bad") {
						return nil, nil, want
					}
					if stage == "cancel download" {
						cancel()
					}
					return downloadableSharedLinkFile("good", "", arg.Url, 4), io.NopCloser(strings.NewReader("data")), nil
				},
			})
			stubFilesClient(t, &mockFilesClient{
				listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
					switch stage {
					case "unknown entry":
						return &files.ListFolderResult{Entries: []files.IsMetadata{nil}}, nil
					case "nil file":
						return &files.ListFolderResult{Entries: []files.IsMetadata{(*files.FileMetadata)(nil)}}, nil
					case "nil folder":
						return &files.ListFolderResult{Entries: []files.IsMetadata{(*files.FolderMetadata)(nil)}}, nil
					}
					if stage == "list" || (stage == "nested list" && arg.Path == "/nested/folder/child") {
						return nil, want
					}
					if stage == "cancel listing" {
						cancel()
						return &files.ListFolderResult{}, nil
					}
					if stage == "continue" || (stage == "nested continue" && strings.HasSuffix(arg.Path, "/child")) || stage == "missing cursor" || stage == "cancel pagination" {
						cursor := "next"
						if stage == "missing cursor" {
							cursor = ""
						}
						if stage == "cancel pagination" {
							cancel()
						}
						return &files.ListFolderResult{HasMore: true, Cursor: cursor}, nil
					}
					if strings.HasSuffix(arg.Path, "/child") {
						return &files.ListFolderResult{}, nil
					}
					entries := []files.IsMetadata{listingFile("good", 4), listingFolder("child")}
					if stage == "download" {
						entries = append(entries, listingFile("bad", 0))
					}
					return &files.ListFolderResult{Entries: entries}, nil
				},
				listFolderContinueFn: func(arg *files.ListFolderContinueArg) (*files.ListFolderResult, error) { return nil, want },
			})
			var out, stderr bytes.Buffer
			cmd := newShareLinkDownloadTestCommand(&out, &stderr)
			setShareLinkOutputJSON(t, cmd)
			if err := cmd.ParseFlags([]string{"--path", "/nested/folder", "-r"}); err != nil {
				t.Fatal(err)
			}
			err := shareLinkDownload(cmd, []string{listingURL, target})
			if err == nil || out.Len() != 0 {
				t.Fatalf("err=%v out=%s", err, &out)
			}
			if stage != "local" && stage != "missing cursor" && stage != "unknown entry" && stage != "nil file" && stage != "nil folder" && !errors.Is(err, want) {
				t.Fatalf("error=%v, want wrapped %v", err, want)
			}
			if err := schema.Validate(normalizeJSONValueForSchema(t, newJSONErrorResponse(cmd, err))); err != nil {
				t.Fatalf("download error violates JSON v1 schema: %v", err)
			}
			details := jsonErrorDetails(err)
			if details["url"] != listingURL || details["path"] != "/nested/folder" || details["operation"] != "share_link_download" || !reflect.DeepEqual(details["flags"], []string{"path", "recursive"}) {
				t.Fatalf("details=%#v", details)
			}
			if stage == "download" {
				assertFileContent(t, filepath.Join(target, "good"), "data")
				entries, readErr := os.ReadDir(target)
				if readErr != nil {
					t.Fatal(readErr)
				}
				for _, entry := range entries {
					if strings.Contains(entry.Name(), ".tmp-") || entry.Name() == "bad" {
						t.Fatalf("failed file left behind: %s", entry.Name())
					}
				}
			}
		})
	}
}

func TestShareLinkDownloadSelectedFolderUnsafeNames(t *testing.T) {
	names := []string{"", ".", "..", "../escape", "/absolute", `a\b`, `C:\escape`, "a/b", "a\x00b"}
	if runtime.GOOS == "windows" {
		names = append(names, "CON", "NUL", "C:notes")
	}
	for _, name := range names {
		for _, source := range []string{"metadata", "entry"} {
			t.Run(source+"/"+name, func(t *testing.T) {
				target := filepath.Join(t.TempDir(), "out")
				stubSharedLinkClient(t, &mockSharedLinkClient{
					getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
						folder := sharedLinkFolder("/safe", arg.Url)
						if source == "metadata" {
							folder.Name = name
						}
						return folder, nil
					},
					getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
						t.Fatal("unsafe file downloaded")
						return nil, nil, nil
					},
				})
				stubFilesClient(t, &mockFilesClient{listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
					if arg.Path != "/nested/folder" {
						t.Fatalf("unsafe folder listed: %q", arg.Path)
					}
					return &files.ListFolderResult{Entries: []files.IsMetadata{listingFolder(name), listingFile(name, 0)}}, nil
				}})
				cmd := newShareLinkDownloadTestCommand(nil, nil)
				if err := cmd.ParseFlags([]string{"--path", "/nested/folder", "-r"}); err != nil {
					t.Fatal(err)
				}
				if err := shareLinkDownload(cmd, []string{listingURL, target}); err == nil {
					t.Fatal("unsafe name accepted")
				}
			})
		}
	}
}

func TestShareLinkDownloadPathsFromLS(t *testing.T) {
	const selected = "/2027/Logoset"
	filesDbx := &mockFilesClient{listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
		switch arg.Path {
		case "/2027":
			return &files.ListFolderResult{Entries: []files.IsMetadata{listingFolder("Logoset")}}, nil
		case selected:
			return &files.ListFolderResult{Entries: []files.IsMetadata{listingFile("logo.png", 4)}}, nil
		default:
			t.Fatalf("unexpected listing %q", arg.Path)
			return nil, nil
		}
	}}
	arg := sharing.NewGetSharedLinkMetadataArg(listingURL)
	results, err := listSharedLinkContents(filesDbx, arg, "2027", true)
	if err != nil {
		t.Fatal(err)
	}
	stubFilesClient(t, filesDbx)
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
			if arg.Path == selected {
				return sharedLinkFolder("/Logoset", arg.Url), nil
			}
			return downloadableSharedLinkFile("logo.png", "", arg.Url, 4), nil
		},
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			if arg.Path != selected+"/logo.png" {
				t.Fatalf("download path=%q", arg.Path)
			}
			return downloadableSharedLinkFile("logo.png", "", arg.Url, 4), io.NopCloser(strings.NewReader("data")), nil
		},
	})
	for _, result := range results {
		target := filepath.Join(t.TempDir(), "out")
		cmd := newShareLinkDownloadTestCommand(nil, nil)
		if err := cmd.ParseFlags([]string{"--path", result.Result.(shareLinkLSEntry).Path, "-r"}); err != nil {
			t.Fatal(err)
		}
		if err := shareLinkDownload(cmd, []string{listingURL, target}); err != nil {
			t.Fatal(err)
		}
		if result.Kind == "folder" {
			target = filepath.Join(target, "logo.png")
		}
		assertFileContent(t, target, "data")
	}
}

func TestShareLinkDownloadSelectedFolderRetriesAndCleansPartialFiles(t *testing.T) {
	for _, recover := range []bool{false, true} {
		t.Run(fmt.Sprintf("recover=%t", recover), func(t *testing.T) {
			delays := stubRetrySleep(t)
			root := t.TempDir()
			dst := filepath.Join(root, "Selected")
			if err := os.Mkdir(dst, 0755); err != nil {
				t.Fatal(err)
			}
			file := filepath.Join(dst, "file.txt")
			if err := os.WriteFile(file, []byte("original"), 0600); err != nil {
				t.Fatal(err)
			}
			calls := 0
			stubSharedLinkClient(t, &mockSharedLinkClient{
				getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
					return sharedLinkFolder("/Selected", arg.Url), nil
				},
				getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
					calls++
					if arg.Path != "/nested/Selected/file.txt" || arg.LinkPassword != "secret" {
						t.Fatalf("download arg=%#v", arg)
					}
					link := downloadableSharedLinkFile("file.txt", "", arg.Url, 7)
					if recover && calls > 1 {
						return link, io.NopCloser(strings.NewReader("complete")), nil
					}
					return link, &failingReadCloser{data: []byte("partial")}, nil
				},
			})
			stubFilesClient(t, &mockFilesClient{listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
				return &files.ListFolderResult{Entries: []files.IsMetadata{listingFile("file.txt", 7)}}, nil
			}})
			var out, stderr bytes.Buffer
			cmd := newShareLinkDownloadTestCommand(&out, &stderr)
			setShareLinkOutputJSON(t, cmd)
			if err := cmd.ParseFlags([]string{"--path", "/nested/Selected", "-r", "--password", "secret"}); err != nil {
				t.Fatal(err)
			}
			err := shareLinkDownload(cmd, []string{listingURL, root})
			if recover {
				if err != nil || calls != 2 || len(*delays) != 1 {
					t.Fatalf("error=%v calls=%d delays=%v", err, calls, *delays)
				}
				assertFileContent(t, file, "complete")
			} else {
				if !errors.Is(err, io.ErrUnexpectedEOF) || out.Len() != 0 || calls != maxRetries+1 {
					t.Fatalf("error=%v calls=%d output=%s", err, calls, &out)
				}
				assertFileContent(t, file, "original")
			}
			entries, err := os.ReadDir(dst)
			if err != nil || len(entries) != 1 || entries[0].Name() != "file.txt" {
				t.Fatalf("leftover files=%v error=%v", entries, err)
			}
		})
	}
}
