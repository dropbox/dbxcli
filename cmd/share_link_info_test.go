package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

func TestShareLinkInfoRequiresExactlyOneURL(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing url", args: nil},
		{name: "too many urls", args: []string{"http://a", "http://b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			stubSharedLinkClient(t, &mockSharedLinkClient{
				getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
					called = true
					return nil, nil
				},
			})

			err := shareLinkInfo(&cobra.Command{}, tt.args)
			if err == nil || !strings.Contains(err.Error(), "requires a `url` argument") {
				t.Fatalf("error = %v, want url argument error", err)
			}
			if called {
				t.Fatal("GetSharedLinkMetadata should not be called")
			}
		})
	}
}

func TestShareLinkInfoRejectsEmptyURL(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
			called = true
			return nil, nil
		},
	})

	err := shareLinkInfo(&cobra.Command{}, []string{""})
	if err == nil || !strings.Contains(err.Error(), "non-empty URL") {
		t.Fatalf("error = %v, want non-empty URL error", err)
	}
	if called {
		t.Fatal("GetSharedLinkMetadata should not be called")
	}
}

func TestShareLinkInfoCallsAPIWithURLAndPrintsFileInfo(t *testing.T) {
	serverModified := time.Date(2026, 6, 20, 12, 30, 0, 0, time.UTC)
	expires := time.Date(2026, 7, 1, 8, 0, 0, 0, time.UTC)
	permissions := sharing.NewLinkPermissions(true, nil, false, false, true, false, false, false, false)
	permissions.ResolvedVisibility = &sharing.ResolvedVisibility{Tagged: dropbox.Tagged{Tag: sharing.ResolvedVisibilityPublic}}

	link := sharedLinkFile("/docs/report.txt", "https://www.dropbox.com/s/abc123")
	link.Id = "id:file123"
	link.Expires = &expires
	link.LinkPermissions = permissions
	link.ServerModified = serverModified
	link.Rev = "rev123"
	link.Size = 2048

	var requestedURL string
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
			requestedURL = arg.Url
			return link, nil
		},
	})

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	err := shareLinkInfo(cmd, []string{"https://www.dropbox.com/s/abc123"})
	if err != nil {
		t.Fatalf("shareLinkInfo error: %v", err)
	}
	if requestedURL != "https://www.dropbox.com/s/abc123" {
		t.Fatalf("requested URL = %q, want https://www.dropbox.com/s/abc123", requestedURL)
	}

	got := stdout.String()
	for _, want := range []string{
		"Type:                file\n",
		"Name:                report.txt\n",
		"URL:                 https://www.dropbox.com/s/abc123\n",
		"Path:                /docs/report.txt\n",
		"ID:                  id:file123\n",
		"Expires:             2026-07-01T08:00:00Z\n",
		"Resolved Visibility: public\n",
		"Can Revoke:          true\n",
		"Allow Download:      true\n",
		"Revision:            rev123\n",
		"Size:                2.0 KiB\n",
		"Server Modified:     2026-06-20T12:30:00Z\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("stdout = %q, missing %q", got, want)
		}
	}
}

func TestShareLinkInfoPassesPathAndPassword(t *testing.T) {
	var requestedPath string
	var requestedPassword string
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
			requestedPath = arg.Path
			requestedPassword = arg.LinkPassword
			return sharedLinkFile("/docs/report.txt", "https://www.dropbox.com/s/abc123"), nil
		},
	})

	var stdout bytes.Buffer
	cmd := newShareLinkInfoTestCommand(&stdout)
	if err := cmd.Flags().Set("path", "docs/report.txt"); err != nil {
		t.Fatalf("set path: %v", err)
	}
	if err := cmd.Flags().Set("password", "secret"); err != nil {
		t.Fatalf("set password: %v", err)
	}

	if err := shareLinkInfo(cmd, []string{"https://www.dropbox.com/s/abc123"}); err != nil {
		t.Fatalf("shareLinkInfo error: %v", err)
	}
	if requestedPath != "/docs/report.txt" {
		t.Fatalf("path = %q, want /docs/report.txt", requestedPath)
	}
	if requestedPassword != "secret" {
		t.Fatalf("password = %q, want secret", requestedPassword)
	}
	if !strings.Contains(stdout.String(), "URL:") {
		t.Fatalf("stdout = %q, want rendered metadata", stdout.String())
	}
}

func TestShareLinkInfoRejectsEmptyPathFlag(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
			called = true
			return nil, nil
		},
	})

	cmd := newShareLinkInfoTestCommand(nil)
	if err := cmd.Flags().Set("path", ""); err != nil {
		t.Fatalf("set path: %v", err)
	}

	err := shareLinkInfo(cmd, []string{"https://www.dropbox.com/s/abc123"})
	if err == nil || !strings.Contains(err.Error(), "`--path` requires a non-empty path") {
		t.Fatalf("error = %v, want empty path error", err)
	}
	if called {
		t.Fatal("GetSharedLinkMetadata should not be called")
	}
}

func TestShareLinkInfoPrintsFolderInfo(t *testing.T) {
	link := sharedLinkFolder("/docs", "https://www.dropbox.com/s/folder")
	link.Id = "id:folder123"

	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
			return link, nil
		},
	})

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	if err := shareLinkInfo(cmd, []string{"https://www.dropbox.com/s/folder"}); err != nil {
		t.Fatalf("shareLinkInfo error: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{
		"Type: folder\n",
		"Name: docs\n",
		"URL:  https://www.dropbox.com/s/folder\n",
		"Path: /docs\n",
		"ID:   id:folder123\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("stdout = %q, missing %q", got, want)
		}
	}
}

func TestShareLinkInfoReturnsAPIError(t *testing.T) {
	wantErr := fmt.Errorf("shared_link_not_found")
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
			return nil, wantErr
		},
	})

	err := shareLinkInfo(&cobra.Command{}, []string{"https://www.dropbox.com/s/abc123"})
	if err != wantErr {
		t.Fatalf("error = %v, want API error", err)
	}
}

func TestShareLinkInfoDoesNotBreakOtherCommands(t *testing.T) {
	cmd, _, err := RootCmd.Find([]string{"share-link", "info"})
	if err != nil {
		t.Fatalf("find share-link info: %v", err)
	}
	if cmd != shareLinkInfoCmd {
		t.Fatalf("share-link info resolved to %q", cmd.CommandPath())
	}
	if shareLinkInfoCmd.Flags().Lookup("path") == nil {
		t.Fatal("share-link info should define --path")
	}
	if shareLinkInfoCmd.Flags().Lookup("password") == nil {
		t.Fatal("share-link info should define --password")
	}
	if shareLinkInfoCmd.Flags().Lookup("password-prompt") == nil {
		t.Fatal("share-link info should define --password-prompt")
	}
	if shareLinkInfoCmd.Flags().Lookup("password-file") == nil {
		t.Fatal("share-link info should define --password-file")
	}

	cmd, _, err = RootCmd.Find([]string{"share-link", "create"})
	if err != nil {
		t.Fatalf("find share-link create: %v", err)
	}
	if cmd != shareLinkCreateCmd {
		t.Fatalf("share-link create resolved to %q", cmd.CommandPath())
	}

	cmd, _, err = RootCmd.Find([]string{"share-link", "list"})
	if err != nil {
		t.Fatalf("find share-link list: %v", err)
	}
	if cmd != shareLinkListCmd {
		t.Fatalf("share-link list resolved to %q", cmd.CommandPath())
	}

	cmd, _, err = RootCmd.Find([]string{"share-link", "revoke"})
	if err != nil {
		t.Fatalf("find share-link revoke: %v", err)
	}
	if cmd != shareLinkRevokeCmd {
		t.Fatalf("share-link revoke resolved to %q", cmd.CommandPath())
	}

	cmd, _, err = RootCmd.Find([]string{"share-link", "download"})
	if err != nil {
		t.Fatalf("find share-link download: %v", err)
	}
	if cmd != shareLinkDownloadCmd {
		t.Fatalf("share-link download resolved to %q", cmd.CommandPath())
	}
}

func newShareLinkInfoTestCommand(stdout *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("path", "", "")
	addSharedLinkPasswordFlags(cmd)
	if stdout != nil {
		cmd.SetOut(stdout)
	}
	return cmd
}
