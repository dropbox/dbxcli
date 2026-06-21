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
	if err != wantErr {
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
}

func newShareLinkUpdateTestCommand(stdout, stderr *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("expires", "", "")
	cmd.Flags().Bool("remove-expiration", false, "")
	cmd.Flags().Bool("allow-download", false, "")
	cmd.Flags().Bool("verbose", false, "")
	if stdout != nil {
		cmd.SetOut(stdout)
	}
	if stderr != nil {
		cmd.SetErr(stderr)
	}
	return cmd
}
