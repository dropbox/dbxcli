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
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

func TestShareLinkUpdateRequiresExactlyOneURL(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing URL", args: nil},
		{name: "too many URLs", args: []string{"https://example.com/one", "https://example.com/two"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			stubSharedLinkClient(t, &mockSharedLinkClient{
				modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
					called = true
					return nil, nil
				},
			})

			err := shareLinkUpdate(newShareLinkUpdateTestCommand(nil, nil), tt.args)
			if err == nil || !strings.Contains(err.Error(), "requires a `url` argument") {
				t.Fatalf("error = %v, want URL argument error", err)
			}
			if called {
				t.Fatal("ModifySharedLinkSettings should not be called")
			}
		})
	}
}

func TestShareLinkUpdateRejectsEmptyURL(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			called = true
			return nil, nil
		},
	})

	err := shareLinkUpdate(newShareLinkUpdateTestCommand(nil, nil), []string{""})
	if err == nil || !strings.Contains(err.Error(), "requires a non-empty URL") {
		t.Fatalf("error = %v, want non-empty URL error", err)
	}
	if called {
		t.Fatal("ModifySharedLinkSettings should not be called")
	}
}

func TestShareLinkUpdateRequiresSettingFlag(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			called = true
			return nil, nil
		},
	})

	err := shareLinkUpdate(newShareLinkUpdateTestCommand(nil, nil), []string{"https://example.com/link"})
	if err == nil || !strings.Contains(err.Error(), "at least one shared link setting flag is required") {
		t.Fatalf("error = %v, want setting flag error", err)
	}
	if called {
		t.Fatal("ModifySharedLinkSettings should not be called")
	}
}

func TestShareLinkUpdateRejectsInvalidExpires(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			called = true
			return nil, nil
		},
	})

	cmd := newShareLinkUpdateTestCommand(nil, nil)
	if err := cmd.Flags().Set("expires", "tomorrow"); err != nil {
		t.Fatalf("set expires: %v", err)
	}

	err := shareLinkUpdate(cmd, []string{"https://example.com/link"})
	if err == nil || !strings.Contains(err.Error(), "invalid --expires") {
		t.Fatalf("error = %v, want invalid expires error", err)
	}
	if called {
		t.Fatal("ModifySharedLinkSettings should not be called")
	}
}

func TestShareLinkUpdateRejectsInvalidAudience(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			called = true
			return nil, nil
		},
	})

	cmd := newShareLinkUpdateTestCommand(nil, nil)
	if err := cmd.Flags().Set("audience", "password"); err != nil {
		t.Fatalf("set audience: %v", err)
	}

	err := shareLinkUpdate(cmd, []string{"https://example.com/link"})
	if err == nil || !strings.Contains(err.Error(), `invalid --audience "password": use public, team, members, or no-one`) {
		t.Fatalf("error = %v, want invalid audience error", err)
	}
	if called {
		t.Fatal("ModifySharedLinkSettings should not be called")
	}
}

func TestShareLinkUpdateRejectsExpiresAndRemoveExpiration(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			called = true
			return nil, nil
		},
	})

	cmd := newShareLinkUpdateTestCommand(nil, nil)
	if err := cmd.Flags().Set("expires", "2026-07-01T00:00:00Z"); err != nil {
		t.Fatalf("set expires: %v", err)
	}
	if err := cmd.Flags().Set("remove-expiration", "true"); err != nil {
		t.Fatalf("set remove-expiration: %v", err)
	}

	err := shareLinkUpdate(cmd, []string{"https://example.com/link"})
	if err == nil || !strings.Contains(err.Error(), "cannot be used together") {
		t.Fatalf("error = %v, want mutual exclusion error", err)
	}
	if called {
		t.Fatal("ModifySharedLinkSettings should not be called")
	}
}

func TestShareLinkUpdateSetsExpiration(t *testing.T) {
	expires := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	mock := &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			if arg.Url != "https://example.com/link" {
				t.Fatalf("url = %q, want https://example.com/link", arg.Url)
			}
			if arg.Settings == nil || arg.Settings.Expires == nil {
				t.Fatal("expires setting was not sent")
			}
			if !arg.Settings.Expires.Equal(expires) {
				t.Fatalf("expires = %v, want %v", arg.Settings.Expires, expires)
			}
			if arg.RemoveExpiration {
				t.Fatal("remove expiration = true, want false")
			}
			return sharedLinkFile("/file.txt", "https://example.com/link"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	cmd := newShareLinkUpdateTestCommand(&stdout, &stderr)
	if err := cmd.Flags().Set("expires", "2026-07-01T00:00:00Z"); err != nil {
		t.Fatalf("set expires: %v", err)
	}

	if err := shareLinkUpdate(cmd, []string{"https://example.com/link"}); err != nil {
		t.Fatalf("shareLinkUpdate error: %v", err)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestShareLinkUpdateRemovesExpiration(t *testing.T) {
	mock := &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			if !arg.RemoveExpiration {
				t.Fatal("remove expiration = false, want true")
			}
			if arg.Settings == nil {
				t.Fatal("settings = nil, want empty settings")
			}
			if arg.Settings.Expires != nil {
				t.Fatalf("expires = %v, want nil", arg.Settings.Expires)
			}
			return sharedLinkFile("/file.txt", "https://example.com/link"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	cmd := newShareLinkUpdateTestCommand(nil, nil)
	if err := cmd.Flags().Set("remove-expiration", "true"); err != nil {
		t.Fatalf("set remove-expiration: %v", err)
	}

	if err := shareLinkUpdate(cmd, []string{"https://example.com/link"}); err != nil {
		t.Fatalf("shareLinkUpdate error: %v", err)
	}
}

func TestShareLinkUpdateAllowsDownload(t *testing.T) {
	mock := &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			if arg.Settings == nil || !arg.Settings.AllowDownload {
				t.Fatalf("allow download = false, want true")
			}
			return sharedLinkFile("/file.txt", "https://example.com/link"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	cmd := newShareLinkUpdateTestCommand(nil, nil)
	if err := cmd.Flags().Set("allow-download", "true"); err != nil {
		t.Fatalf("set allow-download: %v", err)
	}

	if err := shareLinkUpdate(cmd, []string{"https://example.com/link"}); err != nil {
		t.Fatalf("shareLinkUpdate error: %v", err)
	}
}

func TestShareLinkUpdateDisallowsDownload(t *testing.T) {
	var rawURL string
	mock := &mockSharedLinkClient{
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

	cmd := newShareLinkUpdateTestCommand(nil, nil)
	if err := cmd.Flags().Set("disallow-download", "true"); err != nil {
		t.Fatalf("set disallow-download: %v", err)
	}

	if err := shareLinkUpdate(cmd, []string{"https://example.com/link"}); err != nil {
		t.Fatalf("shareLinkUpdate error: %v", err)
	}
	if rawURL != "https://example.com/link" {
		t.Fatalf("raw modify URL = %q, want https://example.com/link", rawURL)
	}
}

func TestShareLinkUpdateDisallowDownloadCombinesSettingsInOneRawCall(t *testing.T) {
	expires := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	var rawCalls int
	mock := &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			t.Fatal("ModifySharedLinkSettings should not be called when disallow-download is combined with other settings")
			return nil, nil
		},
		modifySharedLinkSettingsRawFn: func(url string, settings *rawSharedLinkSettings, removeExpiration bool) error {
			rawCalls++
			if url != "https://example.com/link" {
				t.Fatalf("url = %q, want https://example.com/link", url)
			}
			if removeExpiration {
				t.Fatal("remove expiration = true, want false")
			}
			if settings == nil || settings.AllowDownload == nil || *settings.AllowDownload {
				t.Fatalf("settings = %#v, want allow_download=false", settings)
			}
			if settings.Expires == nil || !settings.Expires.Equal(expires) {
				t.Fatalf("expires = %v, want %v", settings.Expires, expires)
			}
			if settings.Audience == nil || settings.Audience.Tag != sharing.LinkAudienceTeam {
				t.Fatalf("audience = %#v, want team", settings.Audience)
			}
			if settings.RequirePassword == nil || !*settings.RequirePassword || settings.LinkPassword != "secret" {
				t.Fatalf("password settings = require:%v value:%q, want password settings", settings.RequirePassword, settings.LinkPassword)
			}
			return nil
		},
	}
	stubSharedLinkClient(t, mock)

	cmd := newShareLinkUpdateTestCommand(nil, nil)
	if err := cmd.Flags().Set("disallow-download", "true"); err != nil {
		t.Fatalf("set disallow-download: %v", err)
	}
	if err := cmd.Flags().Set("expires", expires.Format(time.RFC3339)); err != nil {
		t.Fatalf("set expires: %v", err)
	}
	if err := cmd.Flags().Set("audience", "team"); err != nil {
		t.Fatalf("set audience: %v", err)
	}
	if err := cmd.Flags().Set("password", "secret"); err != nil {
		t.Fatalf("set password: %v", err)
	}

	if err := shareLinkUpdate(cmd, []string{"https://example.com/link"}); err != nil {
		t.Fatalf("shareLinkUpdate error: %v", err)
	}
	if rawCalls != 1 {
		t.Fatalf("raw modify calls = %d, want 1", rawCalls)
	}
}

func TestShareLinkUpdateRejectsAllowAndDisallowDownload(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			called = true
			return nil, nil
		},
		modifySharedLinkSettingsRawFn: func(url string, settings *rawSharedLinkSettings, removeExpiration bool) error {
			called = true
			return nil
		},
	})

	cmd := newShareLinkUpdateTestCommand(nil, nil)
	if err := cmd.Flags().Set("allow-download", "true"); err != nil {
		t.Fatalf("set allow-download: %v", err)
	}
	if err := cmd.Flags().Set("disallow-download", "true"); err != nil {
		t.Fatalf("set disallow-download: %v", err)
	}

	err := shareLinkUpdate(cmd, []string{"https://example.com/link"})
	if err == nil || !strings.Contains(err.Error(), "cannot be used together") {
		t.Fatalf("error = %v, want mutual exclusion error", err)
	}
	if called {
		t.Fatal("shared link API should not be called")
	}
}

func TestShareLinkUpdateSetsPassword(t *testing.T) {
	mock := &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			if arg.Url != "https://example.com/link" {
				t.Fatalf("url = %q, want https://example.com/link", arg.Url)
			}
			if arg.Settings == nil || !arg.Settings.RequirePassword || arg.Settings.LinkPassword != "secret" {
				t.Fatalf("settings = %#v, want password settings", arg.Settings)
			}
			return sharedLinkFile("/file.txt", "https://example.com/link"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	cmd := newShareLinkUpdateTestCommand(nil, nil)
	if err := cmd.Flags().Set("password", "secret"); err != nil {
		t.Fatalf("set password: %v", err)
	}

	if err := shareLinkUpdate(cmd, []string{"https://example.com/link"}); err != nil {
		t.Fatalf("shareLinkUpdate error: %v", err)
	}
}

func TestShareLinkUpdateReadsPasswordFile(t *testing.T) {
	passwordFile := t.TempDir() + "/password.txt"
	if err := os.WriteFile(passwordFile, []byte("file-secret\n"), 0600); err != nil {
		t.Fatalf("write password file: %v", err)
	}

	mock := &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			if arg.Settings == nil || !arg.Settings.RequirePassword || arg.Settings.LinkPassword != "file-secret" {
				t.Fatalf("settings = %#v, want password from file", arg.Settings)
			}
			return sharedLinkFile("/file.txt", "https://example.com/link"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	cmd := newShareLinkUpdateTestCommand(nil, nil)
	if err := cmd.Flags().Set("password-file", passwordFile); err != nil {
		t.Fatalf("set password-file: %v", err)
	}

	if err := shareLinkUpdate(cmd, []string{"https://example.com/link"}); err != nil {
		t.Fatalf("shareLinkUpdate error: %v", err)
	}
}

func TestShareLinkUpdateRemovesPassword(t *testing.T) {
	var rawURL string
	mock := &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			t.Fatal("ModifySharedLinkSettings should not be called for remove-password only")
			return nil, nil
		},
		modifySharedLinkSettingsRawFn: func(url string, settings *rawSharedLinkSettings, removeExpiration bool) error {
			rawURL = url
			if removeExpiration {
				t.Fatal("remove expiration = true, want false")
			}
			if settings == nil || settings.RequirePassword == nil {
				t.Fatalf("settings = %#v, want require_password setting", settings)
			}
			if *settings.RequirePassword {
				t.Fatal("require_password = true, want false")
			}
			return nil
		},
	}
	stubSharedLinkClient(t, mock)

	cmd := newShareLinkUpdateTestCommand(nil, nil)
	if err := cmd.Flags().Set("remove-password", "true"); err != nil {
		t.Fatalf("set remove-password: %v", err)
	}

	if err := shareLinkUpdate(cmd, []string{"https://example.com/link"}); err != nil {
		t.Fatalf("shareLinkUpdate error: %v", err)
	}
	if rawURL != "https://example.com/link" {
		t.Fatalf("raw modify URL = %q, want https://example.com/link", rawURL)
	}
}

func TestShareLinkUpdateRemovePasswordCombinesSettingsInOneRawCall(t *testing.T) {
	expires := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	var rawCalls int
	mock := &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			t.Fatal("ModifySharedLinkSettings should not be called when remove-password is combined with other settings")
			return nil, nil
		},
		modifySharedLinkSettingsRawFn: func(url string, settings *rawSharedLinkSettings, removeExpiration bool) error {
			rawCalls++
			if url != "https://example.com/link" {
				t.Fatalf("url = %q, want https://example.com/link", url)
			}
			if removeExpiration {
				t.Fatal("remove expiration = true, want false")
			}
			if settings == nil {
				t.Fatal("settings = nil, want raw settings")
			}
			if settings.RequirePassword == nil || *settings.RequirePassword {
				t.Fatalf("require_password = %v, want false", settings.RequirePassword)
			}
			if settings.AllowDownload == nil || !*settings.AllowDownload {
				t.Fatalf("allow_download = %v, want true", settings.AllowDownload)
			}
			if settings.Expires == nil || !settings.Expires.Equal(expires) {
				t.Fatalf("expires = %v, want %v", settings.Expires, expires)
			}
			if settings.Audience == nil || settings.Audience.Tag != sharing.LinkAudienceTeam {
				t.Fatalf("audience = %#v, want team", settings.Audience)
			}
			return nil
		},
	}
	stubSharedLinkClient(t, mock)

	cmd := newShareLinkUpdateTestCommand(nil, nil)
	if err := cmd.Flags().Set("remove-password", "true"); err != nil {
		t.Fatalf("set remove-password: %v", err)
	}
	if err := cmd.Flags().Set("allow-download", "true"); err != nil {
		t.Fatalf("set allow-download: %v", err)
	}
	if err := cmd.Flags().Set("expires", expires.Format(time.RFC3339)); err != nil {
		t.Fatalf("set expires: %v", err)
	}
	if err := cmd.Flags().Set("audience", "team"); err != nil {
		t.Fatalf("set audience: %v", err)
	}

	if err := shareLinkUpdate(cmd, []string{"https://example.com/link"}); err != nil {
		t.Fatalf("shareLinkUpdate error: %v", err)
	}
	if rawCalls != 1 {
		t.Fatalf("raw modify calls = %d, want 1", rawCalls)
	}
}

func TestShareLinkUpdateRejectsPasswordAndRemovePassword(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			called = true
			return nil, nil
		},
		modifySharedLinkSettingsRawFn: func(url string, settings *rawSharedLinkSettings, removeExpiration bool) error {
			called = true
			return nil
		},
	})

	cmd := newShareLinkUpdateTestCommand(nil, nil)
	if err := cmd.Flags().Set("password", "secret"); err != nil {
		t.Fatalf("set password: %v", err)
	}
	if err := cmd.Flags().Set("remove-password", "true"); err != nil {
		t.Fatalf("set remove-password: %v", err)
	}

	err := shareLinkUpdate(cmd, []string{"https://example.com/link"})
	if err == nil || !strings.Contains(err.Error(), "password-setting flags and `--remove-password` cannot be used together") {
		t.Fatalf("error = %v, want password mutual exclusion error", err)
	}
	if called {
		t.Fatal("shared link API should not be called")
	}
}

func TestShareLinkUpdateSetsAudience(t *testing.T) {
	mock := &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			if arg.Url != "https://example.com/link" {
				t.Fatalf("url = %q, want https://example.com/link", arg.Url)
			}
			if arg.Settings == nil || arg.Settings.Audience == nil {
				t.Fatal("audience setting was not sent")
			}
			if arg.Settings.Audience.Tag != sharing.LinkAudienceNoOne {
				t.Fatalf("audience = %q, want no_one", arg.Settings.Audience.Tag)
			}
			if arg.RemoveExpiration {
				t.Fatal("remove expiration = true, want false")
			}
			return sharedLinkFile("/file.txt", "https://example.com/link"), nil
		},
	}
	stubSharedLinkClient(t, mock)

	cmd := newShareLinkUpdateTestCommand(nil, nil)
	if err := cmd.Flags().Set("audience", "no-one"); err != nil {
		t.Fatalf("set audience: %v", err)
	}

	if err := shareLinkUpdate(cmd, []string{"https://example.com/link"}); err != nil {
		t.Fatalf("shareLinkUpdate error: %v", err)
	}
}

func TestShareLinkUpdateVerboseWritesStatusToStderr(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	stubSharedLinkClient(t, &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			return sharedLinkFile("/file.txt", "https://example.com/link"), nil
		},
	})

	cmd := newShareLinkUpdateTestCommand(&stdout, &stderr)
	if err := cmd.Flags().Set("allow-download", "true"); err != nil {
		t.Fatalf("set allow-download: %v", err)
	}
	if err := cmd.Flags().Set("verbose", "true"); err != nil {
		t.Fatalf("set verbose: %v", err)
	}

	if err := shareLinkUpdate(cmd, []string{"https://example.com/link"}); err != nil {
		t.Fatalf("shareLinkUpdate error: %v", err)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if got := stderr.String(); got != "Updated shared link https://example.com/link\n" {
		t.Fatalf("stderr = %q, want verbose status", got)
	}
}

func TestShareLinkUpdateReturnsAPIErrors(t *testing.T) {
	wantErr := fmt.Errorf("shared_link_not_found")
	stubSharedLinkClient(t, &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			return nil, wantErr
		},
	})

	cmd := newShareLinkUpdateTestCommand(nil, nil)
	if err := cmd.Flags().Set("allow-download", "true"); err != nil {
		t.Fatalf("set allow-download: %v", err)
	}

	err := shareLinkUpdate(cmd, []string{"https://example.com/link"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want original API error", err)
	}
}

func TestShareLinkUpdateCommandIsRegistered(t *testing.T) {
	cmd, _, err := RootCmd.Find([]string{"share-link", "update", "https://example.com/link"})
	if err != nil {
		t.Fatalf("find share-link update: %v", err)
	}
	if cmd != shareLinkUpdateCmd {
		t.Fatalf("share-link update resolved to %q", cmd.CommandPath())
	}
	if shareLinkUpdateCmd.Flags().Lookup("audience") == nil {
		t.Fatal("share-link update should define --audience")
	}
	if shareLinkUpdateCmd.Flags().Lookup("password") == nil {
		t.Fatal("share-link update should define --password")
	}
	if shareLinkUpdateCmd.Flags().Lookup("password-prompt") == nil {
		t.Fatal("share-link update should define --password-prompt")
	}
	if shareLinkUpdateCmd.Flags().Lookup("password-file") == nil {
		t.Fatal("share-link update should define --password-file")
	}
	if shareLinkUpdateCmd.Flags().Lookup("remove-password") == nil {
		t.Fatal("share-link update should define --remove-password")
	}
	if shareLinkUpdateCmd.Flags().Lookup("disallow-download") == nil {
		t.Fatal("share-link update should define --disallow-download")
	}
}

func newShareLinkUpdateTestCommand(stdout, stderr *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("audience", "", "")
	cmd.Flags().String("expires", "", "")
	cmd.Flags().Bool("remove-expiration", false, "")
	cmd.Flags().Bool("allow-download", false, "")
	cmd.Flags().Bool("disallow-download", false, "")
	addSharedLinkPasswordFlags(cmd)
	cmd.Flags().Bool("remove-password", false, "")
	cmd.Flags().Bool("verbose", false, "")
	if stdout != nil {
		cmd.SetOut(stdout)
	}
	if stderr != nil {
		cmd.SetErr(stderr)
	}
	return cmd
}
