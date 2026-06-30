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
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

type mockSharedLinkClient struct {
	createSharedLinkWithSettingsFn    func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error)
	createSharedLinkWithRawSettingsFn func(path string, settings *rawSharedLinkSettings) (sharing.IsSharedLinkMetadata, error)
	getSharedLinkFileFn               func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error)
	getSharedLinkMetadataFn           func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error)
	listSharedLinksFn                 func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error)
	modifySharedLinkSettingsFn        func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error)
	revokeSharedLinkFn                func(arg *sharing.RevokeSharedLinkArg) error
	modifySharedLinkSettingsRawFn     func(url string, settings *rawSharedLinkSettings, removeExpiration bool) error
}

func (m *mockSharedLinkClient) CreateSharedLinkWithSettings(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
	if m.createSharedLinkWithSettingsFn != nil {
		return m.createSharedLinkWithSettingsFn(arg)
	}
	return nil, nil
}

func (m *mockSharedLinkClient) CreateSharedLinkWithRawSettings(path string, settings *rawSharedLinkSettings) (sharing.IsSharedLinkMetadata, error) {
	if m.createSharedLinkWithRawSettingsFn != nil {
		return m.createSharedLinkWithRawSettingsFn(path, settings)
	}
	return nil, nil
}

func (m *mockSharedLinkClient) GetSharedLinkFile(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
	if m.getSharedLinkFileFn != nil {
		return m.getSharedLinkFileFn(arg)
	}
	return nil, nil, nil
}

func (m *mockSharedLinkClient) GetSharedLinkMetadata(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
	if m.getSharedLinkMetadataFn != nil {
		return m.getSharedLinkMetadataFn(arg)
	}
	return nil, nil
}

func (m *mockSharedLinkClient) ListSharedLinks(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
	if m.listSharedLinksFn != nil {
		return m.listSharedLinksFn(arg)
	}
	return &sharing.ListSharedLinksResult{}, nil
}

func (m *mockSharedLinkClient) ModifySharedLinkSettings(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
	if m.modifySharedLinkSettingsFn != nil {
		return m.modifySharedLinkSettingsFn(arg)
	}
	return nil, nil
}

func (m *mockSharedLinkClient) RevokeSharedLink(arg *sharing.RevokeSharedLinkArg) error {
	if m.revokeSharedLinkFn != nil {
		return m.revokeSharedLinkFn(arg)
	}
	return nil
}

func (m *mockSharedLinkClient) ModifySharedLinkSettingsRaw(url string, settings *rawSharedLinkSettings, removeExpiration bool) error {
	if m.modifySharedLinkSettingsRawFn != nil {
		return m.modifySharedLinkSettingsRawFn(url, settings, removeExpiration)
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

func TestSharedLinkCreateWithExpiresSetsExpiration(t *testing.T) {
	wantExpires := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			if arg.Settings == nil {
				t.Fatal("settings = nil, want expiration settings")
			}
			if arg.Settings.Expires == nil {
				t.Fatal("expires = nil, want expiration time")
			}
			if !arg.Settings.Expires.Equal(wantExpires) {
				t.Fatalf("expires = %s, want %s", arg.Settings.Expires.Format(time.RFC3339), wantExpires.Format(time.RFC3339))
			}
			return sharedLinkFile("/file.txt", "https://example.com/file"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().String("expires", "", "")
	cmd.SetOut(&stdout)
	if err := cmd.Flags().Set("expires", wantExpires.Format(time.RFC3339)); err != nil {
		t.Fatalf("set expires: %v", err)
	}

	if err := shareLinkCreate(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}

	if got := stdout.String(); got != "https://example.com/file\n" {
		t.Fatalf("stdout = %q, want URL only", got)
	}
}

func TestSharedLinkCreateWithAllowDownloadSetsAllowDownload(t *testing.T) {
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			if arg.Settings == nil {
				t.Fatal("settings = nil, want allow-download settings")
			}
			if !arg.Settings.AllowDownload {
				t.Fatal("AllowDownload = false, want true")
			}
			return sharedLinkFile("/file.txt", "https://example.com/file"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().Bool("allow-download", false, "")
	cmd.SetOut(&stdout)
	if err := cmd.Flags().Set("allow-download", "true"); err != nil {
		t.Fatalf("set allow-download: %v", err)
	}

	if err := shareLinkCreate(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}

	if got := stdout.String(); got != "https://example.com/file\n" {
		t.Fatalf("stdout = %q, want URL only", got)
	}
}

func TestSharedLinkCreateWithDisallowDownloadUpdatesNewLink(t *testing.T) {
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			t.Fatal("CreateSharedLinkWithSettings should not be called for disallow-download")
			return nil, nil
		},
		createSharedLinkWithRawSettingsFn: func(path string, settings *rawSharedLinkSettings) (sharing.IsSharedLinkMetadata, error) {
			if path != "/file.txt" {
				t.Fatalf("path = %q, want /file.txt", path)
			}
			if settings == nil || settings.AllowDownload == nil {
				t.Fatalf("settings = %#v, want allow_download setting", settings)
			}
			if *settings.AllowDownload {
				t.Fatal("allow_download = true, want false")
			}
			return sharedLinkFile("/file.txt", "https://example.com/file"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().Bool("disallow-download", false, "")
	cmd.SetOut(&stdout)
	if err := cmd.Flags().Set("disallow-download", "true"); err != nil {
		t.Fatalf("set disallow-download: %v", err)
	}

	if err := shareLinkCreate(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}
	if got := stdout.String(); got != "https://example.com/file\n" {
		t.Fatalf("stdout = %q, want URL only", got)
	}
}

func TestSharedLinkCreateWithDisallowDownloadCombinesRawCreateSettings(t *testing.T) {
	expires := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			t.Fatal("CreateSharedLinkWithSettings should not be called for disallow-download")
			return nil, nil
		},
		createSharedLinkWithRawSettingsFn: func(path string, settings *rawSharedLinkSettings) (sharing.IsSharedLinkMetadata, error) {
			if path != "/file.txt" {
				t.Fatalf("path = %q, want /file.txt", path)
			}
			if settings == nil {
				t.Fatal("settings = nil, want raw settings")
			}
			if settings.AllowDownload == nil || *settings.AllowDownload {
				t.Fatalf("allow_download = %v, want false", settings.AllowDownload)
			}
			if settings.Expires == nil || !settings.Expires.Equal(expires) {
				t.Fatalf("expires = %v, want %v", settings.Expires, expires)
			}
			if settings.Access == nil || settings.Access.Tag != sharing.RequestedLinkAccessLevelViewer {
				t.Fatalf("access = %#v, want viewer", settings.Access)
			}
			if settings.Audience == nil || settings.Audience.Tag != sharing.LinkAudienceTeam {
				t.Fatalf("audience = %#v, want team", settings.Audience)
			}
			if settings.RequirePassword == nil || !*settings.RequirePassword || settings.LinkPassword != "secret" {
				t.Fatalf("password settings = require:%v value:%q, want password settings", settings.RequirePassword, settings.LinkPassword)
			}
			return sharedLinkFile("/file.txt", "https://example.com/file"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().Bool("disallow-download", false, "")
	cmd.Flags().String("expires", "", "")
	cmd.Flags().String("access", "", "")
	cmd.Flags().String("audience", "", "")
	addSharedLinkPasswordFlags(cmd)
	cmd.SetOut(&stdout)
	if err := cmd.Flags().Set("disallow-download", "true"); err != nil {
		t.Fatalf("set disallow-download: %v", err)
	}
	if err := cmd.Flags().Set("expires", expires.Format(time.RFC3339)); err != nil {
		t.Fatalf("set expires: %v", err)
	}
	if err := cmd.Flags().Set("access", "viewer"); err != nil {
		t.Fatalf("set access: %v", err)
	}
	if err := cmd.Flags().Set("audience", "team"); err != nil {
		t.Fatalf("set audience: %v", err)
	}
	if err := cmd.Flags().Set("password", "secret"); err != nil {
		t.Fatalf("set password: %v", err)
	}

	if err := shareLinkCreate(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}
	if got := stdout.String(); got != "https://example.com/file\n" {
		t.Fatalf("stdout = %q, want URL only", got)
	}
}

func TestSharedLinkCreateRejectsAllowAndDisallowDownload(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			called = true
			return nil, nil
		},
	})

	cmd := &cobra.Command{}
	cmd.Flags().Bool("allow-download", false, "")
	cmd.Flags().Bool("disallow-download", false, "")
	if err := cmd.Flags().Set("allow-download", "true"); err != nil {
		t.Fatalf("set allow-download: %v", err)
	}
	if err := cmd.Flags().Set("disallow-download", "true"); err != nil {
		t.Fatalf("set disallow-download: %v", err)
	}

	err := shareLinkCreate(cmd, []string{"/file.txt"})
	if err == nil || !strings.Contains(err.Error(), "cannot be used together") {
		t.Fatalf("error = %v, want mutual exclusion error", err)
	}
	if called {
		t.Fatal("CreateSharedLinkWithSettings should not be called")
	}
}

func TestSharedLinkCreateWithPasswordSetsPassword(t *testing.T) {
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			if arg.Settings == nil {
				t.Fatal("settings = nil, want password settings")
			}
			if !arg.Settings.RequirePassword {
				t.Fatal("RequirePassword = false, want true")
			}
			if arg.Settings.LinkPassword != "secret" {
				t.Fatalf("LinkPassword = %q, want secret", arg.Settings.LinkPassword)
			}
			return sharedLinkFile("/file.txt", "https://example.com/file"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	addSharedLinkPasswordFlags(cmd)
	cmd.SetOut(&stdout)
	if err := cmd.Flags().Set("password", "secret"); err != nil {
		t.Fatalf("set password: %v", err)
	}

	if err := shareLinkCreate(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}
	if got := stdout.String(); got != "https://example.com/file\n" {
		t.Fatalf("stdout = %q, want URL only", got)
	}
}

func TestSharedLinkCreateWithPasswordPromptSetsPassword(t *testing.T) {
	orig := readSharedLinkPassword
	readSharedLinkPassword = func(prompt string, in io.Reader, errOut io.Writer) (string, error) {
		if prompt != "Shared link password: " {
			t.Fatalf("prompt = %q, want shared link password prompt", prompt)
		}
		return "prompt-secret", nil
	}
	t.Cleanup(func() { readSharedLinkPassword = orig })

	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			if arg.Settings == nil || !arg.Settings.RequirePassword || arg.Settings.LinkPassword != "prompt-secret" {
				t.Fatalf("settings = %#v, want prompted password", arg.Settings)
			}
			return sharedLinkFile("/file.txt", "https://example.com/file"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	addSharedLinkPasswordFlags(cmd)
	cmd.SetOut(&stdout)
	if err := cmd.Flags().Set("password-prompt", "true"); err != nil {
		t.Fatalf("set password-prompt: %v", err)
	}

	if err := shareLinkCreate(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}
}

func TestSharedLinkCreateRejectsMultiplePasswordSources(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			called = true
			return nil, nil
		},
	})

	cmd := &cobra.Command{}
	addSharedLinkPasswordFlags(cmd)
	if err := cmd.Flags().Set("password", "secret"); err != nil {
		t.Fatalf("set password: %v", err)
	}
	if err := cmd.Flags().Set("password-prompt", "true"); err != nil {
		t.Fatalf("set password-prompt: %v", err)
	}

	err := shareLinkCreate(cmd, []string{"/file.txt"})
	if err == nil || !strings.Contains(err.Error(), "use only one of `--password`, `--password-prompt`, or `--password-file`") {
		t.Fatalf("error = %v, want password source error", err)
	}
	if called {
		t.Fatal("CreateSharedLinkWithSettings should not be called")
	}
}

func TestSharedLinkCreateRejectsEmptyPassword(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			called = true
			return nil, nil
		},
	})

	cmd := &cobra.Command{}
	addSharedLinkPasswordFlags(cmd)
	if err := cmd.Flags().Set("password", ""); err != nil {
		t.Fatalf("set password: %v", err)
	}

	err := shareLinkCreate(cmd, []string{"/file.txt"})
	if err == nil || !strings.Contains(err.Error(), "shared link password cannot be empty") {
		t.Fatalf("error = %v, want empty password error", err)
	}
	if called {
		t.Fatal("CreateSharedLinkWithSettings should not be called")
	}
}

func TestSharedLinkCreateWithAccessSetsAccess(t *testing.T) {
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			if arg.Settings == nil {
				t.Fatal("settings = nil, want access settings")
			}
			if arg.Settings.Access == nil {
				t.Fatal("access = nil, want editor")
			}
			if arg.Settings.Access.Tag != sharing.RequestedLinkAccessLevelEditor {
				t.Fatalf("access = %q, want editor", arg.Settings.Access.Tag)
			}
			return sharedLinkFile("/file.txt", "https://example.com/file"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().String("access", "", "")
	cmd.SetOut(&stdout)
	if err := cmd.Flags().Set("access", "editor"); err != nil {
		t.Fatalf("set access: %v", err)
	}

	if err := shareLinkCreate(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}

	if got := stdout.String(); got != "https://example.com/file\n" {
		t.Fatalf("stdout = %q, want URL only", got)
	}
}

func TestSharedLinkCreateWithAudienceSetsAudience(t *testing.T) {
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			if arg.Settings == nil {
				t.Fatal("settings = nil, want audience settings")
			}
			if arg.Settings.Audience == nil {
				t.Fatal("audience = nil, want team")
			}
			if arg.Settings.Audience.Tag != sharing.LinkAudienceTeam {
				t.Fatalf("audience = %q, want team", arg.Settings.Audience.Tag)
			}
			return sharedLinkFile("/file.txt", "https://example.com/file"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().String("audience", "", "")
	cmd.SetOut(&stdout)
	if err := cmd.Flags().Set("audience", "team"); err != nil {
		t.Fatalf("set audience: %v", err)
	}

	if err := shareLinkCreate(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}

	if got := stdout.String(); got != "https://example.com/file\n" {
		t.Fatalf("stdout = %q, want URL only", got)
	}
}

func TestSharedLinkCreateWithInvalidExpiresReturnsError(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			called = true
			return nil, nil
		},
	})

	cmd := &cobra.Command{}
	cmd.Flags().String("expires", "", "")
	if err := cmd.Flags().Set("expires", "tomorrow"); err != nil {
		t.Fatalf("set expires: %v", err)
	}

	err := shareLinkCreate(cmd, []string{"/file.txt"})
	if err == nil || !strings.Contains(err.Error(), `invalid --expires "tomorrow": use RFC3339 timestamp`) {
		t.Fatalf("error = %v, want invalid expires error", err)
	}
	if called {
		t.Fatal("CreateSharedLinkWithSettings should not be called")
	}
}

func TestSharedLinkCreateWithInvalidAccessReturnsError(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			called = true
			return nil, nil
		},
	})

	cmd := &cobra.Command{}
	cmd.Flags().String("access", "", "")
	if err := cmd.Flags().Set("access", "owner"); err != nil {
		t.Fatalf("set access: %v", err)
	}

	err := shareLinkCreate(cmd, []string{"/file.txt"})
	if err == nil || !strings.Contains(err.Error(), `invalid --access "owner": use viewer, editor, or max`) {
		t.Fatalf("error = %v, want invalid access error", err)
	}
	if called {
		t.Fatal("CreateSharedLinkWithSettings should not be called")
	}
}

func TestSharedLinkCreateWithInvalidAudienceReturnsError(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			called = true
			return nil, nil
		},
	})

	cmd := &cobra.Command{}
	cmd.Flags().String("audience", "", "")
	if err := cmd.Flags().Set("audience", "password"); err != nil {
		t.Fatalf("set audience: %v", err)
	}

	err := shareLinkCreate(cmd, []string{"/file.txt"})
	if err == nil || !strings.Contains(err.Error(), `invalid --audience "password": use public, team, members, or no-one`) {
		t.Fatalf("error = %v, want invalid audience error", err)
	}
	if called {
		t.Fatal("CreateSharedLinkWithSettings should not be called")
	}
}

func TestSharedLinkCreateRejectsExpiresWithRemoveExpiration(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			called = true
			return nil, nil
		},
	})

	cmd := &cobra.Command{}
	cmd.Flags().String("expires", "", "")
	cmd.Flags().Bool("remove-expiration", false, "")
	if err := cmd.Flags().Set("expires", "2026-07-01T00:00:00Z"); err != nil {
		t.Fatalf("set expires: %v", err)
	}
	if err := cmd.Flags().Set("remove-expiration", "true"); err != nil {
		t.Fatalf("set remove-expiration: %v", err)
	}

	err := shareLinkCreate(cmd, []string{"/file.txt"})
	if err == nil || !strings.Contains(err.Error(), "`--expires` and `--remove-expiration` cannot be used together") {
		t.Fatalf("error = %v, want mutually exclusive error", err)
	}
	if called {
		t.Fatal("CreateSharedLinkWithSettings should not be called")
	}
}

func TestSharedLinkCreateWithRemoveExpirationUsesDefaultCreateSettings(t *testing.T) {
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			if arg.Settings != nil {
				t.Fatalf("settings = %#v, want nil for new link with remove-expiration", arg.Settings)
			}
			return sharedLinkFile("/file.txt", "https://example.com/file"), nil
		},
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			t.Fatal("ModifySharedLinkSettings should not be called for a newly created link")
			return nil, nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().Bool("remove-expiration", false, "")
	cmd.SetOut(&stdout)
	if err := cmd.Flags().Set("remove-expiration", "true"); err != nil {
		t.Fatalf("set remove-expiration: %v", err)
	}

	if err := shareLinkCreate(cmd, []string{"/file.txt"}); err != nil {
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
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want original error", err)
	}
	details := jsonErrorDetails(err)
	if details["operation"] != "share_link_create" || details["path"] != "/docs" {
		t.Fatalf("details = %#v, want share_link_create operation and path", details)
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

func TestSharedLinkCreateWithAccessErrorsForExistingLink(t *testing.T) {
	existing := sharedLinkFile("/file.txt", "https://example.com/file-old")
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			if arg.Settings == nil || arg.Settings.Access == nil || arg.Settings.Access.Tag != sharing.RequestedLinkAccessLevelMax {
				t.Fatalf("create settings = %#v, want max access", arg.Settings)
			}
			return nil, alreadyExistsError(existing)
		},
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			t.Fatal("ModifySharedLinkSettings should not be called because access cannot be modified")
			return nil, nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().String("access", "", "")
	cmd.SetOut(&stdout)
	if err := cmd.Flags().Set("access", "max"); err != nil {
		t.Fatalf("set access: %v", err)
	}

	err := shareLinkCreate(cmd, []string{"/file.txt"})
	if err == nil || !strings.Contains(err.Error(), "cannot apply `--access` because the shared link already exists") {
		t.Fatalf("error = %v, want existing link access error", err)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty output on error", got)
	}
}

func TestSharedLinkCreateWithAudienceUpdatesExistingLink(t *testing.T) {
	existing := sharedLinkFile("/file.txt", "https://example.com/file-old")
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			if arg.Settings == nil || arg.Settings.Audience == nil || arg.Settings.Audience.Tag != sharing.LinkAudienceMembers {
				t.Fatalf("create settings = %#v, want members audience", arg.Settings)
			}
			return nil, alreadyExistsError(existing)
		},
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			if arg.Url != "https://example.com/file-old" {
				t.Fatalf("modify URL = %q, want existing URL", arg.Url)
			}
			if arg.Settings == nil || arg.Settings.Audience == nil || arg.Settings.Audience.Tag != sharing.LinkAudienceMembers {
				t.Fatalf("modify settings = %#v, want members audience", arg.Settings)
			}
			return sharedLinkFile("/file.txt", "https://example.com/file-new"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().String("audience", "", "")
	cmd.SetOut(&stdout)
	if err := cmd.Flags().Set("audience", "members"); err != nil {
		t.Fatalf("set audience: %v", err)
	}

	if err := shareLinkCreate(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}
	if got := stdout.String(); got != "https://example.com/file-new\n" {
		t.Fatalf("stdout = %q, want updated URL", got)
	}
}

func TestSharedLinkCreateWithExpiresUpdatesExistingLink(t *testing.T) {
	wantExpires := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	existing := sharedLinkFile("/file.txt", "https://example.com/file-old")
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			if arg.Settings == nil || arg.Settings.Expires == nil || !arg.Settings.Expires.Equal(wantExpires) {
				t.Fatalf("create settings = %#v, want expires %s", arg.Settings, wantExpires.Format(time.RFC3339))
			}
			return nil, alreadyExistsError(existing)
		},
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			if arg.Url != "https://example.com/file-old" {
				t.Fatalf("modify URL = %q, want existing URL", arg.Url)
			}
			if arg.RemoveExpiration {
				t.Fatal("RemoveExpiration = true, want false")
			}
			if arg.Settings == nil || arg.Settings.Expires == nil || !arg.Settings.Expires.Equal(wantExpires) {
				t.Fatalf("modify settings = %#v, want expires %s", arg.Settings, wantExpires.Format(time.RFC3339))
			}
			return sharedLinkFile("/file.txt", "https://example.com/file-new"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().String("expires", "", "")
	cmd.SetOut(&stdout)
	if err := cmd.Flags().Set("expires", wantExpires.Format(time.RFC3339)); err != nil {
		t.Fatalf("set expires: %v", err)
	}

	if err := shareLinkCreate(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}
	if got := stdout.String(); got != "https://example.com/file-new\n" {
		t.Fatalf("stdout = %q, want updated URL", got)
	}
}

func TestSharedLinkCreateWithRemoveExpirationUpdatesExistingLink(t *testing.T) {
	existing := sharedLinkFile("/file.txt", "https://example.com/file-old")
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			if arg.Settings != nil {
				t.Fatalf("create settings = %#v, want nil", arg.Settings)
			}
			return nil, alreadyExistsError(existing)
		},
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			if arg.Url != "https://example.com/file-old" {
				t.Fatalf("modify URL = %q, want existing URL", arg.Url)
			}
			if !arg.RemoveExpiration {
				t.Fatal("RemoveExpiration = false, want true")
			}
			if arg.Settings == nil {
				t.Fatal("settings = nil, want empty settings object")
			}
			return sharedLinkFile("/file.txt", "https://example.com/file-new"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().Bool("remove-expiration", false, "")
	cmd.SetOut(&stdout)
	if err := cmd.Flags().Set("remove-expiration", "true"); err != nil {
		t.Fatalf("set remove-expiration: %v", err)
	}

	if err := shareLinkCreate(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}
	if got := stdout.String(); got != "https://example.com/file-new\n" {
		t.Fatalf("stdout = %q, want updated URL", got)
	}
}

func TestSharedLinkCreateWithAllowDownloadUpdatesExistingLink(t *testing.T) {
	existing := sharedLinkFile("/file.txt", "https://example.com/file-old")
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			if arg.Settings == nil || !arg.Settings.AllowDownload {
				t.Fatalf("create settings = %#v, want allow download", arg.Settings)
			}
			return nil, alreadyExistsError(existing)
		},
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			if arg.Url != "https://example.com/file-old" {
				t.Fatalf("modify URL = %q, want existing URL", arg.Url)
			}
			if arg.RemoveExpiration {
				t.Fatal("RemoveExpiration = true, want false")
			}
			if arg.Settings == nil || !arg.Settings.AllowDownload {
				t.Fatalf("modify settings = %#v, want allow download", arg.Settings)
			}
			return sharedLinkFile("/file.txt", "https://example.com/file-new"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().Bool("allow-download", false, "")
	cmd.SetOut(&stdout)
	if err := cmd.Flags().Set("allow-download", "true"); err != nil {
		t.Fatalf("set allow-download: %v", err)
	}

	if err := shareLinkCreate(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}
	if got := stdout.String(); got != "https://example.com/file-new\n" {
		t.Fatalf("stdout = %q, want updated URL", got)
	}
}

func TestSharedLinkCreateWithDisallowDownloadUpdatesExistingLink(t *testing.T) {
	existing := sharedLinkFile("/file.txt", "https://example.com/file-old")
	var rawURL string
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			t.Fatal("CreateSharedLinkWithSettings should not be called for disallow-download")
			return nil, nil
		},
		createSharedLinkWithRawSettingsFn: func(path string, settings *rawSharedLinkSettings) (sharing.IsSharedLinkMetadata, error) {
			if path != "/file.txt" {
				t.Fatalf("path = %q, want /file.txt", path)
			}
			if settings == nil || settings.AllowDownload == nil {
				t.Fatalf("create settings = %#v, want allow_download setting", settings)
			}
			if *settings.AllowDownload {
				t.Fatal("create allow_download = true, want false")
			}
			return nil, alreadyExistsError(existing)
		},
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			t.Fatal("ModifySharedLinkSettings should not be called for disallow-download only")
			return nil, nil
		},
		modifySharedLinkSettingsRawFn: func(url string, settings *rawSharedLinkSettings, removeExpiration bool) error {
			rawURL = url
			if removeExpiration {
				t.Fatal("remove expiration = true, want false")
			}
			if settings == nil || settings.AllowDownload == nil {
				t.Fatalf("settings = %#v, want allow_download setting", settings)
			}
			if *settings.AllowDownload {
				t.Fatal("allow_download = true, want false")
			}
			return nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().Bool("disallow-download", false, "")
	cmd.SetOut(&stdout)
	if err := cmd.Flags().Set("disallow-download", "true"); err != nil {
		t.Fatalf("set disallow-download: %v", err)
	}

	if err := shareLinkCreate(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}
	if rawURL != "https://example.com/file-old" {
		t.Fatalf("raw modify URL = %q, want existing URL", rawURL)
	}
	if got := stdout.String(); got != "https://example.com/file-old\n" {
		t.Fatalf("stdout = %q, want existing URL", got)
	}
}

func TestSharedLinkCreateWithPasswordUpdatesExistingLink(t *testing.T) {
	existing := sharedLinkFile("/file.txt", "https://example.com/file-old")
	mock := &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			if arg.Settings == nil || !arg.Settings.RequirePassword || arg.Settings.LinkPassword != "secret" {
				t.Fatalf("create settings = %#v, want password settings", arg.Settings)
			}
			return nil, alreadyExistsError(existing)
		},
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			if arg.Url != "https://example.com/file-old" {
				t.Fatalf("modify URL = %q, want existing URL", arg.Url)
			}
			if arg.Settings == nil || !arg.Settings.RequirePassword || arg.Settings.LinkPassword != "secret" {
				t.Fatalf("modify settings = %#v, want password settings", arg.Settings)
			}
			return sharedLinkFile("/file.txt", "https://example.com/file-new"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	addSharedLinkPasswordFlags(cmd)
	cmd.SetOut(&stdout)
	if err := cmd.Flags().Set("password", "secret"); err != nil {
		t.Fatalf("set password: %v", err)
	}

	if err := shareLinkCreate(cmd, []string{"/file.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}
	if got := stdout.String(); got != "https://example.com/file-new\n" {
		t.Fatalf("stdout = %q, want updated URL", got)
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
	if shareLinkCreateCmd.Flags().Lookup("access") == nil {
		t.Fatal("share-link create should define --access")
	}
	if shareLinkCreateCmd.Flags().Lookup("audience") == nil {
		t.Fatal("share-link create should define --audience")
	}
	if shareLinkCreateCmd.Flags().Lookup("allow-download") == nil {
		t.Fatal("share-link create should define --allow-download")
	}
	if shareLinkCreateCmd.Flags().Lookup("disallow-download") == nil {
		t.Fatal("share-link create should define --disallow-download")
	}
	if shareLinkCreateCmd.Flags().Lookup("expires") == nil {
		t.Fatal("share-link create should define --expires")
	}
	if shareLinkCreateCmd.Flags().Lookup("remove-expiration") == nil {
		t.Fatal("share-link create should define --remove-expiration")
	}
	if shareLinkCreateCmd.Flags().Lookup("password") == nil {
		t.Fatal("share-link create should define --password")
	}
	if shareLinkCreateCmd.Flags().Lookup("password-prompt") == nil {
		t.Fatal("share-link create should define --password-prompt")
	}
	if shareLinkCreateCmd.Flags().Lookup("password-file") == nil {
		t.Fatal("share-link create should define --password-file")
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

func TestShareLinkListRejectsExtraPathWithDetails(t *testing.T) {
	err := shareLinkList(&cobra.Command{}, []string{"/docs/one.txt", "/docs/two.txt"})
	if err == nil || !strings.Contains(err.Error(), "at most one `path` argument") {
		t.Fatalf("error = %v, want extra path error", err)
	}
	details := jsonErrorDetails(err)
	if details["operation"] != "share_link_list" || details["argument"] != "path" || details["path"] != "/docs/two.txt" {
		t.Fatalf("details = %#v, want share_link_list operation, path argument, and extra path", details)
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
