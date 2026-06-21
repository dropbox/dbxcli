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
	"path"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

type mockSharedLinkClient struct {
	createSharedLinkWithSettingsFn func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error)
	listSharedLinksFn              func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error)
	revokeSharedLinkFn             func(arg *sharing.RevokeSharedLinkArg) error
}

func (m *mockSharedLinkClient) CreateSharedLinkWithSettings(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
	if m.createSharedLinkWithSettingsFn != nil {
		return m.createSharedLinkWithSettingsFn(arg)
	}
	return nil, nil
}

func (m *mockSharedLinkClient) ListSharedLinks(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
	if m.listSharedLinksFn != nil {
		return m.listSharedLinksFn(arg)
	}
	return &sharing.ListSharedLinksResult{}, nil
}

func (m *mockSharedLinkClient) RevokeSharedLink(arg *sharing.RevokeSharedLinkArg) error {
	if m.revokeSharedLinkFn != nil {
		return m.revokeSharedLinkFn(arg)
	}
	return nil
}

func stubSharedLinkClient(t *testing.T, client sharedLinkClient) {
	t.Helper()

	orig := newSharedLinkClient
	newSharedLinkClient = func(_ dropbox.Config) sharedLinkClient { return client }
	t.Cleanup(func() { newSharedLinkClient = orig })
}

func TestSharedLinkCreateRequiresExactlyOnePath(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing path", args: nil},
		{name: "too many paths", args: []string{"/one", "/two"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			stubSharedLinkClient(t, &mockSharedLinkClient{
				createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
					called = true
					return nil, nil
				},
			})

			err := shareLinkCreate(&cobra.Command{}, tt.args)
			if err == nil || !strings.Contains(err.Error(), "`share-link create` requires a `path` argument") {
				t.Fatalf("error = %v, want path argument error", err)
			}
			if called {
				t.Fatal("CreateSharedLinkWithSettings should not be called")
			}
		})
	}
}

func TestSharedLinkCreateRejectsRootPath(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			called = true
			return nil, nil
		},
	})

	err := shareLinkCreate(&cobra.Command{}, []string{"/"})
	if err == nil || !strings.Contains(err.Error(), "cannot create a shared link for Dropbox root") {
		t.Fatalf("error = %v, want root path error", err)
	}
	if called {
		t.Fatal("CreateSharedLinkWithSettings should not be called")
	}
}

func TestSharedLinkCreatePrintsURLAndUsesDefaultSettings(t *testing.T) {
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			if arg.Path != "/file.txt" {
				t.Fatalf("create path = %q, want /file.txt", arg.Path)
			}
			if arg.Settings != nil {
				t.Fatalf("settings = %#v, want nil", arg.Settings)
			}
			return sharedLinkFile("/file.txt", "https://example.com/file"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	if err := shareLinkCreate(cmd, []string{"file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}

	if got := stdout.String(); got != "https://example.com/file\n" {
		t.Fatalf("stdout = %q, want URL only", got)
	}
}

func TestSharedLinkCreateVerboseStillPrintsURLOnly(t *testing.T) {
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			return sharedLinkFile("/file.txt", "https://example.com/file"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().Bool("verbose", true, "")
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := shareLinkCreate(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}

	if got := stdout.String(); got != "https://example.com/file\n" {
		t.Fatalf("stdout = %q, want URL only", got)
	}
	if got, want := stderr.String(), "Created shared link for /file.txt\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestSharedLinkCreatePrintsFolderURL(t *testing.T) {
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			return sharedLinkFolder("/docs", "https://example.com/docs"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	if err := shareLinkCreate(cmd, []string{"/docs"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}

	if got := stdout.String(); got != "https://example.com/docs\n" {
		t.Fatalf("stdout = %q, want folder URL only", got)
	}
}

func TestSharedLinkCreateReturnsNonAlreadyExistsError(t *testing.T) {
	wantErr := fmt.Errorf("access_denied")
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			return nil, wantErr
		},
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			t.Fatal("ListSharedLinks should not be called for non-conflict errors")
			return nil, nil
		},
	}
	stubSharedLinkClient(t, mock)

	err := shareLinkCreate(&cobra.Command{}, []string{"/docs"})
	if err != wantErr {
		t.Fatalf("error = %v, want original error", err)
	}
}

func TestSharedLinkCreateExistingMetadataPrintsURLWithoutList(t *testing.T) {
	existing := sharedLinkFolder("/docs", "https://example.com/docs")
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			return nil, alreadyExistsError(existing)
		},
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			t.Fatal("ListSharedLinks should not be called when conflict metadata is returned")
			return nil, nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	if err := shareLinkCreate(cmd, []string{"/docs"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}

	if got := stdout.String(); got != "https://example.com/docs\n" {
		t.Fatalf("stdout = %q, want existing URL only", got)
	}
}

func TestSharedLinkCreateVerboseReportsExistingLinkOnStderr(t *testing.T) {
	existing := sharedLinkFolder("/docs", "https://example.com/docs")
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			return nil, alreadyExistsError(existing)
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().Bool("verbose", true, "")
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := shareLinkCreate(cmd, []string{"/docs"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}

	if got := stdout.String(); got != "https://example.com/docs\n" {
		t.Fatalf("stdout = %q, want existing URL only", got)
	}
	if got, want := stderr.String(), "Using existing shared link for /docs\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestSharedLinkCreateFallbackPrefersExactPathLower(t *testing.T) {
	var listArg *sharing.ListSharedLinksArg
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			return nil, alreadyExistsOtherError()
		},
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			listArg = arg
			return sharing.NewListSharedLinksResult([]sharing.IsSharedLinkMetadata{
				sharedLinkFile("/docs/other.txt", "https://example.com/wrong"),
				sharedLinkFile("/docs/file.txt", "https://example.com/right"),
			}, false), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	if err := shareLinkCreate(cmd, []string{"/docs/file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}

	if listArg == nil {
		t.Fatal("expected ListSharedLinks to be called")
	}
	if listArg.Path != "/docs/file.txt" {
		t.Fatalf("ListSharedLinks path = %q, want /docs/file.txt", listArg.Path)
	}
	if !listArg.DirectOnly {
		t.Fatal("ListSharedLinks DirectOnly = false, want true")
	}
	if got := stdout.String(); got != "https://example.com/right\n" {
		t.Fatalf("stdout = %q, want exact path URL", got)
	}
}

func TestSharedLinkCreateFallbackFollowsPagination(t *testing.T) {
	var cursors []string
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			return nil, alreadyExistsOtherError()
		},
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			cursors = append(cursors, arg.Cursor)
			if arg.Cursor == "" {
				res := sharing.NewListSharedLinksResult(nil, true)
				res.Cursor = "next-page"
				return res, nil
			}
			return sharing.NewListSharedLinksResult([]sharing.IsSharedLinkMetadata{
				sharedLinkFile("/docs/file.txt", "https://example.com/page-two"),
			}, false), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	if err := shareLinkCreate(cmd, []string{"/docs/file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}

	if got := strings.Join(cursors, ","); got != ",next-page" {
		t.Fatalf("cursors = %q, want first call then next-page", got)
	}
	if got := stdout.String(); got != "https://example.com/page-two\n" {
		t.Fatalf("stdout = %q, want second-page URL", got)
	}
}

func TestSharedLinkCreateFallbackErrorsWhenNoDirectLinkFound(t *testing.T) {
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			return nil, alreadyExistsOtherError()
		},
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			return sharing.NewListSharedLinksResult([]sharing.IsSharedLinkMetadata{
				sharedLinkFolder("/docs", "https://example.com/parent"),
			}, false), nil
		},
	}
	stubSharedLinkClient(t, mock)

	err := shareLinkCreate(&cobra.Command{}, []string{"/docs/file.txt"})
	if err == nil || !strings.Contains(err.Error(), "no direct link was found") {
		t.Fatalf("error = %v, want no direct link error", err)
	}
}

func TestSharedLinkCreateFallbackPaginationRequiresCursor(t *testing.T) {
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			return nil, alreadyExistsOtherError()
		},
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			return sharing.NewListSharedLinksResult(nil, true), nil
		},
	}
	stubSharedLinkClient(t, mock)

	cmd := &cobra.Command{}
	err := shareLinkCreate(cmd, []string{"/docs/file.txt"})
	if err == nil || !strings.Contains(err.Error(), "more results but no cursor") {
		t.Fatalf("error = %v, want missing cursor error", err)
	}
}

func TestShareLinkCreateDoesNotBreakShareListLinkCommand(t *testing.T) {
	cmd, _, err := RootCmd.Find([]string{"share-link", "create", "/file.txt"})
	if err != nil {
		t.Fatalf("find share-link create: %v", err)
	}
	if cmd != shareLinkCreateCmd {
		t.Fatalf("share-link create resolved to %q", cmd.CommandPath())
	}

	cmd, _, err = RootCmd.Find([]string{"share", "list", "link"})
	if err != nil {
		t.Fatalf("find share list link: %v", err)
	}
	if cmd != shareListLinksCmd {
		t.Fatalf("share list link resolved to %q", cmd.CommandPath())
	}
	if shareListLinksCmd.Deprecated == "" {
		t.Fatal("share list link should be deprecated")
	}
	if !strings.Contains(shareListLinksCmd.Deprecated, "share-link list") {
		t.Fatalf("deprecation message = %q, want share-link list replacement", shareListLinksCmd.Deprecated)
	}
}

func TestShareLinkListListsAllLinks(t *testing.T) {
	var listArg *sharing.ListSharedLinksArg
	stubSharedLinkClient(t, &mockSharedLinkClient{
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			listArg = arg
			return sharing.NewListSharedLinksResult([]sharing.IsSharedLinkMetadata{
				sharedLinkFile("/docs/file.txt", "https://example.com/file"),
				sharedLinkFolder("/docs", "https://example.com/docs"),
			}, false), nil
		},
	})

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	if err := shareLinkList(cmd, nil); err != nil {
		t.Fatalf("shareLinkList error: %v", err)
	}

	if listArg == nil {
		t.Fatal("expected ListSharedLinks to be called")
	}
	if listArg.Path != "" {
		t.Fatalf("ListSharedLinks path = %q, want empty", listArg.Path)
	}
	if listArg.DirectOnly {
		t.Fatal("ListSharedLinks DirectOnly = true, want false")
	}
	want := "file.txt\thttps://example.com/file\n" +
		"docs\thttps://example.com/docs\n"
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestShareLinkListVerboseWritesStatusToStderr(t *testing.T) {
	stubSharedLinkClient(t, &mockSharedLinkClient{
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			return sharing.NewListSharedLinksResult([]sharing.IsSharedLinkMetadata{
				sharedLinkFile("/docs/file.txt", "https://example.com/file"),
			}, false), nil
		},
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().Bool("verbose", true, "")
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := shareLinkList(cmd, []string{"/docs/file.txt"}); err != nil {
		t.Fatalf("shareLinkList error: %v", err)
	}

	if got, want := stdout.String(), "file.txt\thttps://example.com/file\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if got, want := stderr.String(), "Listed 1 shared links for /docs/file.txt\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestShareLinkListPathFilterUsesDirectOnly(t *testing.T) {
	var listArg *sharing.ListSharedLinksArg
	stubSharedLinkClient(t, &mockSharedLinkClient{
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			listArg = arg
			return sharing.NewListSharedLinksResult([]sharing.IsSharedLinkMetadata{
				sharedLinkFile("/docs/file.txt", "https://example.com/file"),
			}, false), nil
		},
	})

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	if err := shareLinkList(cmd, []string{"docs/file.txt"}); err != nil {
		t.Fatalf("shareLinkList error: %v", err)
	}

	if listArg == nil {
		t.Fatal("expected ListSharedLinks to be called")
	}
	if listArg.Path != "/docs/file.txt" {
		t.Fatalf("ListSharedLinks path = %q, want /docs/file.txt", listArg.Path)
	}
	if !listArg.DirectOnly {
		t.Fatal("ListSharedLinks DirectOnly = false, want true")
	}
	want := "file.txt\thttps://example.com/file\n"
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestShareLinkListFollowsPagination(t *testing.T) {
	var cursors []string
	stubSharedLinkClient(t, &mockSharedLinkClient{
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			cursors = append(cursors, arg.Cursor)
			if arg.Cursor == "" {
				res := sharing.NewListSharedLinksResult([]sharing.IsSharedLinkMetadata{
					sharedLinkFile("/docs/one.txt", "https://example.com/one"),
				}, true)
				res.Cursor = "next-page"
				return res, nil
			}
			return sharing.NewListSharedLinksResult([]sharing.IsSharedLinkMetadata{
				sharedLinkFile("/docs/two.txt", "https://example.com/two"),
			}, false), nil
		},
	})

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	if err := shareLinkList(cmd, nil); err != nil {
		t.Fatalf("shareLinkList error: %v", err)
	}

	if got := strings.Join(cursors, ","); got != ",next-page" {
		t.Fatalf("cursors = %q, want first call then next-page", got)
	}
	got := stdout.String()
	for _, want := range []string{"https://example.com/one", "https://example.com/two"} {
		if !strings.Contains(got, want) {
			t.Fatalf("stdout = %q, missing %q", got, want)
		}
	}
}

func TestShareLinkListPaginationRequiresCursor(t *testing.T) {
	stubSharedLinkClient(t, &mockSharedLinkClient{
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			return sharing.NewListSharedLinksResult(nil, true), nil
		},
	})

	err := shareLinkList(&cobra.Command{}, nil)
	if err == nil || !strings.Contains(err.Error(), "more results but no cursor") {
		t.Fatalf("error = %v, want missing cursor error", err)
	}
}

func sharedLinkFile(pathLower string, url string) *sharing.FileLinkMetadata {
	link := sharing.NewFileLinkMetadata(url, path.Base(pathLower), nil, time.Time{}, time.Time{}, "rev", 1)
	link.PathLower = strings.ToLower(pathLower)
	return link
}

func sharedLinkFolder(pathLower string, url string) *sharing.FolderLinkMetadata {
	link := sharing.NewFolderLinkMetadata(url, path.Base(pathLower), nil)
	link.PathLower = strings.ToLower(pathLower)
	return link
}

func alreadyExistsError(link sharing.IsSharedLinkMetadata) error {
	return sharing.CreateSharedLinkWithSettingsAPIError{
		APIError: dropbox.APIError{ErrorSummary: "shared_link_already_exists"},
		EndpointError: &sharing.CreateSharedLinkWithSettingsError{
			Tagged: dropbox.Tagged{Tag: sharing.CreateSharedLinkWithSettingsErrorSharedLinkAlreadyExists},
			SharedLinkAlreadyExists: &sharing.SharedLinkAlreadyExistsMetadata{
				Tagged:   dropbox.Tagged{Tag: sharing.SharedLinkAlreadyExistsMetadataMetadata},
				Metadata: link,
			},
		},
	}
}

func alreadyExistsOtherError() error {
	return fmt.Errorf("wrapped: %w", sharing.CreateSharedLinkWithSettingsAPIError{
		APIError: dropbox.APIError{ErrorSummary: "shared_link_already_exists"},
		EndpointError: &sharing.CreateSharedLinkWithSettingsError{
			Tagged: dropbox.Tagged{Tag: sharing.CreateSharedLinkWithSettingsErrorSharedLinkAlreadyExists},
			SharedLinkAlreadyExists: &sharing.SharedLinkAlreadyExistsMetadata{
				Tagged: dropbox.Tagged{Tag: sharing.SharedLinkAlreadyExistsMetadataOther},
			},
		},
	})
}
