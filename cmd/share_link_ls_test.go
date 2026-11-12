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
	"reflect"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

const listingURL = "https://www.dropbox.com/s/example/folder"

func newShareLinkLSTestCommand(out *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{Use: "ls"}
	root := &cobra.Command{Use: "dbxcli"}
	parent := &cobra.Command{Use: "share-link"}
	root.AddCommand(parent)
	parent.AddCommand(cmd)
	cmd.Flags().String("path", "", "")
	cmd.Flags().Bool("recursive", false, "")
	cmd.Flags().BoolP("long", "l", false, "")
	addSharedLinkPasswordFlags(cmd)
	cmd.SetOut(out)
	return cmd
}

func stubListingFolder(t *testing.T) {
	t.Helper()
	stubSharedLinkClient(t, &mockSharedLinkClient{getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
		return sharedLinkFolder("/mounted/elsewhere", arg.Url), nil
	}})
}

func listingFile(name string, size uint64) *files.FileMetadata {
	return &files.FileMetadata{Metadata: files.Metadata{Name: name}, Size: size}
}
func listingFolder(name string) *files.FolderMetadata {
	return &files.FolderMetadata{Metadata: files.Metadata{Name: name}}
}

func TestShareLinkLSJSON(t *testing.T) {
	for _, mounted := range []bool{false, true} {
		for _, recursive := range []bool{false, true} {
			for _, start := range []string{"", "/documents"} {
				t.Run(fmt.Sprintf("mounted=%t/recursive=%t/path=%s", mounted, recursive, start), func(t *testing.T) {
					stubSharedLinkClient(t, &mockSharedLinkClient{getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
						if arg.Url != listingURL || arg.Path != start || arg.LinkPassword != "secret" {
							t.Fatalf("metadata arg = %+v", arg)
						}
						return sharedLinkFolder("/unrelated/mount", arg.Url), nil
					}})
					var paths, cursors []string
					stubFilesClient(t, &mockFilesClient{
						listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
							paths = append(paths, arg.Path)
							if arg.SharedLink == nil || arg.SharedLink.Url != listingURL || arg.SharedLink.Password != "secret" || arg.Recursive {
								t.Fatalf("list arg = %+v", arg)
							}
							switch arg.Path {
							case start:
								f := listingFile("résumé.txt", 0)
								if mounted {
									f.PathDisplay, f.PathLower = "/wrong/mount/file", "/wrong/mount/file"
								}
								return &files.ListFolderResult{Entries: []files.IsMetadata{f}, HasMore: true, Cursor: "root"}, nil
							case start + "/nested":
								return &files.ListFolderResult{Entries: []files.IsMetadata{listingFolder("deep")}, HasMore: true, Cursor: "nested"}, nil
							case start + "/nested/deep":
								return &files.ListFolderResult{}, nil
							default:
								t.Fatalf("unexpected path %q", arg.Path)
								return nil, nil
							}
						},
						listFolderContinueFn: func(arg *files.ListFolderContinueArg) (*files.ListFolderResult, error) {
							cursors = append(cursors, arg.Cursor)
							if arg.Cursor == "root" {
								return &files.ListFolderResult{Entries: []files.IsMetadata{listingFolder("nested")}}, nil
							}
							if arg.Cursor == "nested" {
								return &files.ListFolderResult{Entries: []files.IsMetadata{listingFile("照片.jpg", 2048)}}, nil
							}
							t.Fatalf("unexpected cursor %q", arg.Cursor)
							return nil, nil
						},
					})
					var out bytes.Buffer
					cmd := newShareLinkLSTestCommand(&out)
					setShareLinkOutputJSON(t, cmd)
					if start != "" {
						_ = cmd.Flags().Set("path", start)
					}
					_ = cmd.Flags().Set("recursive", fmt.Sprint(recursive))
					_ = cmd.Flags().Set("password", "secret")
					if err := shareLinkLS(cmd, []string{listingURL}); err != nil {
						t.Fatal(err)
					}
					got := decodeShareLinkOperationOutput[shareLinkLSInput, shareLinkLSEntry](t, out.Bytes())
					if got.Input != (shareLinkLSInput{URL: listingURL, Path: start, Recursive: recursive, Password: true}) {
						t.Fatalf("input = %+v", got.Input)
					}
					if strings.Contains(out.String(), "secret") {
						t.Fatal("password leaked")
					}
					wantPaths := []string{start}
					wantCursors := []string{"root"}
					wantEntries := []string{start + "/nested", start + "/résumé.txt"}
					if recursive {
						wantPaths = append(wantPaths, start+"/nested", start+"/nested/deep")
						wantCursors = append(wantCursors, "nested")
						wantEntries = []string{start + "/nested", start + "/nested/deep", start + "/nested/照片.jpg", start + "/résumé.txt"}
					}
					if !reflect.DeepEqual(paths, wantPaths) || !reflect.DeepEqual(cursors, wantCursors) {
						t.Fatalf("calls = %v, %v", paths, cursors)
					}
					var entries []string
					for _, result := range got.Results {
						entries = append(entries, result.Result.Path)
						if result.Status != "listed" {
							t.Fatalf("status = %s", result.Status)
						}
						if (result.Kind == "file") != (result.Result.Size != nil) {
							t.Fatalf("size/kind = %+v", result)
						}
						rel, err := sharedLinkEntryRelativePath(strings.TrimPrefix(start, "/"), "résumé.txt")
						if err != nil {
							t.Fatal(err)
						}
						if result.Result.Name == "résumé.txt" && result.Result.Path != sharedLinkAPIPath(rel) {
							t.Fatal("download path mismatch")
						}
					}
					if !reflect.DeepEqual(entries, wantEntries) {
						t.Fatalf("entries = %v, want %v", entries, wantEntries)
					}
					schema := compileJSONSchemaFile(t, "../docs/json-schema/v1/commands.schema.json")
					if err := schema.Validate(decodeJSONValueForSchema(t, out.Bytes())); err != nil {
						t.Fatal(err)
					}
				})
			}
		}
	}
}

func TestShareLinkLSTextAndEmpty(t *testing.T) {
	for _, empty := range []bool{false, true} {
		for _, long := range []bool{false, true} {
			t.Run(fmt.Sprintf("empty=%t/long=%t", empty, long), func(t *testing.T) {
				stubListingFolder(t)
				stubFilesClient(t, &mockFilesClient{listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
					if empty {
						return &files.ListFolderResult{}, nil
					}
					return &files.ListFolderResult{Entries: []files.IsMetadata{listingFile("zero", 0), listingFolder("docs"), listingFile("photo", 2048)}}, nil
				}})
				var out bytes.Buffer
				cmd := newShareLinkLSTestCommand(&out)
				if long {
					if err := cmd.ParseFlags([]string{"-l"}); err != nil {
						t.Fatal(err)
					}
				}
				if err := shareLinkLS(cmd, []string{listingURL}); err != nil {
					t.Fatal(err)
				}
				want := "/docs /photo /zero"
				if empty {
					want = ""
				}
				if long {
					want = "TYPE SIZE PATH folder - /docs file 2.0 KiB /photo file 0 B /zero"
					if empty {
						want = "TYPE SIZE PATH"
					}
				}
				if got := strings.Join(strings.Fields(out.String()), " "); got != want {
					t.Fatalf("output = %q, want %q", out.String(), want)
				}
				setShareLinkOutputJSON(t, cmd)
				out.Reset()
				if err := shareLinkLS(cmd, []string{listingURL}); err != nil {
					t.Fatal(err)
				}
				got := decodeShareLinkOperationOutput[shareLinkLSInput, shareLinkLSEntry](t, out.Bytes())
				if got.Input.Long != long {
					t.Fatalf("input.long = %t, want %t", got.Input.Long, long)
				}
				if empty && len(got.Results) != 0 {
					t.Fatalf("results = %+v", got.Results)
				}
				if !empty {
					if len(got.Results) != 3 || got.Results[1].Kind != "file" || got.Results[1].Result.Size == nil || *got.Results[1].Result.Size != 2048 {
						t.Fatalf("JSON must retain file metadata in both text formats: %+v", got.Results)
					}
				}
				schema := compileJSONSchemaFile(t, "../docs/json-schema/v1/commands.schema.json")
				if err := schema.Validate(decodeJSONValueForSchema(t, out.Bytes())); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestShareLinkLSShortTextLayout(t *testing.T) {
	stubListingFolder(t)
	stubFilesClient(t, &mockFilesClient{listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
		return &files.ListFolderResult{Entries: []files.IsMetadata{
			listingFile("five", 1), listingFolder("four"), listingFile("one", 2), listingFile("three", 3), listingFile("two words", 4),
		}}, nil
	}})
	var out bytes.Buffer
	cmd := newShareLinkLSTestCommand(&out)
	if err := cmd.Flags().Set("path", "/nested"); err != nil {
		t.Fatal(err)
	}
	if err := shareLinkLS(cmd, []string{listingURL}); err != nil {
		t.Fatal(err)
	}
	want := "/nested/five /nested/four /nested/one /nested/three\n/nested/two words\n"
	var normalized []string
	for _, line := range strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n") {
		normalized = append(normalized, strings.Join(strings.Fields(line), " "))
		if strings.TrimRight(line, " \t") != line {
			t.Fatalf("trailing padding: %q", line)
		}
	}
	if got := strings.Join(normalized, "\n") + "\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestShareLinkLSInvalidArguments(t *testing.T) {
	for _, args := range [][]string{nil, {listingURL, "extra"}, {""}, {"bad url"}, {"ftp://example.com/folder"}, {"https:///folder"}, {"https://user:secret@example.com"}} {
		var out bytes.Buffer
		err := shareLinkLS(newShareLinkLSTestCommand(&out), args)
		if err == nil || jsonErrorCode(err) != "invalid_arguments" || out.Len() != 0 {
			t.Fatalf("args=%v, err=%v, out=%s", args, err, &out)
		}
	}
	for _, value := range []string{"", "../escape", "/a/../escape", "/a/./b", "/a//b", "/a/", "/a\\b", "/nul\x00"} {
		var out bytes.Buffer
		cmd := newShareLinkLSTestCommand(&out)
		_ = cmd.Flags().Set("path", value)
		err := shareLinkLS(cmd, []string{listingURL})
		if err == nil || jsonErrorCode(err) != "invalid_arguments" {
			t.Fatalf("path=%q, err=%v", value, err)
		}
	}
}

func TestShareLinkLSInvalidEntries(t *testing.T) {
	for _, name := range []string{"", ".", "..", "../escape", "/absolute", "a/b", "a\\b", "nul\x00"} {
		t.Run(fmt.Sprintf("%q", name), func(t *testing.T) {
			stubListingFolder(t)
			stubFilesClient(t, &mockFilesClient{listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
				return &files.ListFolderResult{Entries: []files.IsMetadata{listingFile("valid", 1), listingFolder(name)}}, nil
			}})
			var out bytes.Buffer
			err := shareLinkLS(newShareLinkLSTestCommand(&out), []string{listingURL})
			if err == nil || !strings.Contains(err.Error(), "invalid shared link entry name") || out.Len() != 0 {
				t.Fatalf("err=%v output=%s", err, &out)
			}
		})
	}
}

func TestShareLinkLSFailures(t *testing.T) {
	for _, stage := range []string{"metadata", "file link", "root", "nested", "continue", "no cursor", "unknown entry", "canceled", "cancel pagination"} {
		t.Run(stage, func(t *testing.T) {
			want := errors.New("API failed")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			old := commandContext
			commandContext = ctx
			t.Cleanup(func() { commandContext = old })
			if stage == "canceled" {
				cancel()
				want = context.Canceled
			}
			stubSharedLinkClient(t, &mockSharedLinkClient{getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
				if stage == "metadata" {
					return nil, want
				}
				if stage == "file link" {
					return sharedLinkFile("/file", arg.Url), nil
				}
				return sharedLinkFolder("/folder", arg.Url), nil
			}})
			stubFilesClient(t, &mockFilesClient{
				listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
					if stage == "root" || (stage == "nested" && arg.Path == "/nested") {
						return nil, want
					}
					if stage == "unknown entry" {
						return &files.ListFolderResult{Entries: []files.IsMetadata{nil}}, nil
					}
					if stage == "continue" {
						return &files.ListFolderResult{HasMore: true, Cursor: "next"}, nil
					}
					if stage == "cancel pagination" {
						cancel()
						return &files.ListFolderResult{HasMore: true, Cursor: "next"}, nil
					}
					if stage == "no cursor" {
						return &files.ListFolderResult{HasMore: true}, nil
					}
					return &files.ListFolderResult{Entries: []files.IsMetadata{listingFolder("nested")}}, nil
				},
				listFolderContinueFn: func(arg *files.ListFolderContinueArg) (*files.ListFolderResult, error) { return nil, want },
			})
			var out bytes.Buffer
			cmd := newShareLinkLSTestCommand(&out)
			setShareLinkOutputJSON(t, cmd)
			_ = cmd.Flags().Set("recursive", "true")
			err := shareLinkLS(cmd, []string{listingURL})
			if err == nil || out.Len() != 0 {
				t.Fatalf("error=%v output=%s", err, &out)
			}
			if stage == "metadata" || stage == "root" || stage == "nested" || stage == "continue" || stage == "canceled" {
				if !errors.Is(err, want) {
					t.Fatalf("error=%v, want %v", err, want)
				}
			}
			if stage == "cancel pagination" && !errors.Is(err, context.Canceled) {
				t.Fatalf("error=%v", err)
			}
			details := jsonErrorDetails(err)
			if details["url"] != listingURL || details["operation"] != "share_link_ls" {
				t.Fatalf("details=%v", details)
			}
		})
	}
}

func TestShareLinkLSHelp(t *testing.T) {
	command, _, err := RootCmd.Find([]string{"share-link", "ls"})
	if err != nil || command != shareLinkLSCmd {
		t.Fatalf("command=%v err=%v", command, err)
	}
	for _, name := range []string{"path", "recursive", "long", "password", "password-file", "password-prompt"} {
		if command.Flags().Lookup(name) == nil {
			t.Fatalf("missing --%s", name)
		}
	}
	if command.Flags().Lookup("long").Shorthand != "l" {
		t.Fatal("--long should have -l shorthand")
	}
	manifest := jsonCommandManifestFor(command)
	if !reflect.DeepEqual(manifest.ResultStatuses, []string{"listed"}) || !reflect.DeepEqual(manifest.ResultKinds, []string{"file", "folder"}) {
		t.Fatalf("manifest=%+v", manifest)
	}
	if !commandSupportsStructuredOutput(command) {
		t.Fatal("missing structured output")
	}
	list, _, err := RootCmd.Find([]string{"share-link", "list"})
	if err != nil || list != shareLinkListCmd {
		t.Fatal("share-link list changed")
	}
}

func TestShareLinkLSPasswordSources(t *testing.T) {
	for _, source := range []string{"none", "password-file", "password-prompt", "conflict", "empty"} {
		t.Run(source, func(t *testing.T) {
			var out bytes.Buffer
			cmd := newShareLinkLSTestCommand(&out)
			setShareLinkOutputJSON(t, cmd)
			want := "secret"
			switch source {
			case "none":
				want = ""
			case "password-file":
				p := t.TempDir() + "/password"
				if err := os.WriteFile(p, []byte("secret\n"), 0600); err != nil {
					t.Fatal(err)
				}
				_ = cmd.Flags().Set(source, p)
			case "password-prompt":
				old := readSharedLinkPassword
				readSharedLinkPassword = func(_ string, _ io.Reader, _ io.Writer) (string, error) { return "secret", nil }
				t.Cleanup(func() { readSharedLinkPassword = old })
				_ = cmd.Flags().Set(source, "true")
			case "conflict":
				_ = cmd.Flags().Set("password", "secret")
				_ = cmd.Flags().Set("password-prompt", "true")
			case "empty":
				_ = cmd.Flags().Set("password", "")
			}
			stubSharedLinkClient(t, &mockSharedLinkClient{getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
				if source == "conflict" || source == "empty" {
					t.Fatal("API called with invalid flags")
				}
				if arg.LinkPassword != want {
					t.Fatalf("metadata password = %q", arg.LinkPassword)
				}
				return sharedLinkFolder("/folder", arg.Url), nil
			}})
			stubFilesClient(t, &mockFilesClient{listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
				if arg.SharedLink.Password != want {
					t.Fatalf("list password = %q", arg.SharedLink.Password)
				}
				return &files.ListFolderResult{}, nil
			}})
			err := shareLinkLS(cmd, []string{listingURL})
			if source == "conflict" || source == "empty" {
				if err == nil || jsonErrorCode(err) != "invalid_arguments" {
					t.Fatalf("error = %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got := decodeShareLinkOperationOutput[shareLinkLSInput, shareLinkLSEntry](t, out.Bytes())
			if got.Input.Password != (want != "") || strings.Contains(out.String(), "secret") {
				t.Fatalf("output = %s", &out)
			}
		})
	}
}

func TestShareLinkLSAPIErrorDetails(t *testing.T) {
	for _, stage := range []string{"metadata", "listing", "pagination"} {
		for _, summary := range []string{"path/not_found/", "shared_link_not_found/", "shared_link_access_denied/", "shared_link/password_required/", "shared_link/password_mismatch/", "shared_link/other/"} {
			t.Run(stage+"/"+summary, func(t *testing.T) {
				var want error = files.ListFolderAPIError{APIError: dropbox.APIError{ErrorSummary: summary}}
				if stage == "metadata" {
					want = sharing.GetSharedLinkMetadataAPIError{APIError: dropbox.APIError{ErrorSummary: summary}}
				}
				if stage == "pagination" {
					want = files.ListFolderContinueAPIError{APIError: dropbox.APIError{ErrorSummary: summary}}
				}
				stubSharedLinkClient(t, &mockSharedLinkClient{getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
					if stage == "metadata" {
						return nil, want
					}
					return sharedLinkFolder("/folder", arg.Url), nil
				}})
				stubFilesClient(t, &mockFilesClient{
					listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
						if stage == "listing" {
							return nil, want
						}
						return &files.ListFolderResult{Entries: []files.IsMetadata{listingFile("partial", 1)}, HasMore: true, Cursor: "next"}, nil
					},
					listFolderContinueFn: func(arg *files.ListFolderContinueArg) (*files.ListFolderResult, error) { return nil, want },
				})
				var out bytes.Buffer
				cmd := newShareLinkLSTestCommand(&out)
				_ = cmd.Flags().Set("path", "/missing")
				err := shareLinkLS(cmd, []string{listingURL})
				if !errors.Is(err, want) || out.Len() != 0 {
					t.Fatalf("error=%v output=%s", err, &out)
				}
				response := newJSONErrorResponse(cmd, err)
				if response.Error.Code != jsonErrorCode(want) || response.Error.Details["api_summary"] != summary || response.Error.Details["path"] != "/missing" {
					t.Fatalf("error = %+v", response.Error)
				}
				schema := compileJSONSchemaFile(t, "../docs/json-schema/v1/error.schema.json")
				if err := schema.Validate(normalizeJSONValueForSchema(t, response)); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestShareLinkLSRootAndRelativePath(t *testing.T) {
	for _, value := range []string{"/", "documents/sub"} {
		t.Run(value, func(t *testing.T) {
			want := strings.TrimPrefix(value, "/")
			stubSharedLinkClient(t, &mockSharedLinkClient{getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
				if arg.Path != sharedLinkAPIPath(want) {
					t.Fatalf("metadata path=%s", arg.Path)
				}
				return sharedLinkFolder("/elsewhere", arg.Url), nil
			}})
			stubFilesClient(t, &mockFilesClient{listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
				if arg.Path != sharedLinkAPIPath(want) {
					t.Fatalf("listing path=%s", arg.Path)
				}
				return &files.ListFolderResult{}, nil
			}})
			var out bytes.Buffer
			cmd := newShareLinkLSTestCommand(&out)
			_ = cmd.Flags().Set("path", value)
			if err := shareLinkLS(cmd, []string{listingURL}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestShareLinkLSPathsCanBeDownloaded(t *testing.T) {
	stubListingFolder(t)
	stubFilesClient(t, &mockFilesClient{listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
		return &files.ListFolderResult{Entries: []files.IsMetadata{listingFile("照片.txt", 7)}}, nil
	}})
	var out bytes.Buffer
	cmd := newShareLinkLSTestCommand(&out)
	setShareLinkOutputJSON(t, cmd)
	_ = cmd.Flags().Set("path", "/documents")
	if err := shareLinkLS(cmd, []string{listingURL}); err != nil {
		t.Fatal(err)
	}
	got := decodeShareLinkOperationOutput[shareLinkLSInput, shareLinkLSEntry](t, out.Bytes())
	if len(got.Results) != 1 {
		t.Fatalf("results=%v", got.Results)
	}
	returnedPath := got.Results[0].Result.Path
	stubSharedLinkClient(t, &mockSharedLinkClient{getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
		if arg.Path != "/documents/照片.txt" || arg.Url != listingURL {
			t.Fatalf("download arg=%+v", arg)
		}
		return downloadableSharedLinkFile("照片.txt", "", arg.Url, 7), io.NopCloser(strings.NewReader("content")), nil
	}})
	var downloaded bytes.Buffer
	download := newShareLinkDownloadTestCommand(&downloaded, nil)
	_ = download.Flags().Set("path", returnedPath)
	target := t.TempDir() + "/download.txt"
	if err := shareLinkDownload(download, []string{listingURL, target}); err != nil {
		t.Fatal(err)
	}
	assertFileContent(t, target, "content")
}

func TestShareLinkLSDeterministicTraversal(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprint(reverse), func(t *testing.T) {
			stubListingFolder(t)
			var calls []string
			stubFilesClient(t, &mockFilesClient{listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
				calls = append(calls, arg.Path)
				if arg.Path == "" {
					entries := []files.IsMetadata{listingFolder("beta"), listingFolder("alpha")}
					if reverse {
						entries[0], entries[1] = entries[1], entries[0]
					}
					return &files.ListFolderResult{Entries: entries}, nil
				}
				return &files.ListFolderResult{Entries: []files.IsMetadata{listingFile("child", 1)}}, nil
			}})
			var out bytes.Buffer
			cmd := newShareLinkLSTestCommand(&out)
			setShareLinkOutputJSON(t, cmd)
			_ = cmd.Flags().Set("recursive", "true")
			if err := shareLinkLS(cmd, []string{listingURL}); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(calls, []string{"", "/alpha", "/beta"}) {
				t.Fatalf("calls=%v", calls)
			}
			got := decodeShareLinkOperationOutput[shareLinkLSInput, shareLinkLSEntry](t, out.Bytes())
			var paths []string
			for _, result := range got.Results {
				paths = append(paths, result.Result.Path)
			}
			if !reflect.DeepEqual(paths, []string{"/alpha", "/alpha/child", "/beta", "/beta/child"}) {
				t.Fatalf("paths=%v", paths)
			}
		})
	}
}
