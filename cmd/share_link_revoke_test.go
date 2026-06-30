package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

func TestShareLinkRevokeRequiresExactlyOneURL(t *testing.T) {
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
				revokeSharedLinkFn: func(arg *sharing.RevokeSharedLinkArg) error {
					called = true
					return nil
				},
			})

			err := shareLinkRevoke(&cobra.Command{}, tt.args)
			if err == nil || !strings.Contains(err.Error(), "requires a `url` argument") {
				t.Fatalf("error = %v, want url argument error", err)
			}
			if called {
				t.Fatal("RevokeSharedLink should not be called")
			}
		})
	}
}

func TestShareLinkRevokeRejectsEmptyURL(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		revokeSharedLinkFn: func(arg *sharing.RevokeSharedLinkArg) error {
			called = true
			return nil
		},
	})

	err := shareLinkRevoke(&cobra.Command{}, []string{""})
	if err == nil || !strings.Contains(err.Error(), "non-empty URL") {
		t.Fatalf("error = %v, want non-empty URL error", err)
	}
	if called {
		t.Fatal("RevokeSharedLink should not be called")
	}
}

func TestShareLinkRevokeCallsAPIWithURL(t *testing.T) {
	var revokedURL string
	stubSharedLinkClient(t, &mockSharedLinkClient{
		revokeSharedLinkFn: func(arg *sharing.RevokeSharedLinkArg) error {
			revokedURL = arg.Url
			return nil
		},
	})

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	err := shareLinkRevoke(cmd, []string{"https://www.dropbox.com/s/abc123"})
	if err != nil {
		t.Fatalf("shareLinkRevoke error: %v", err)
	}
	if revokedURL != "https://www.dropbox.com/s/abc123" {
		t.Fatalf("revoked URL = %q, want https://www.dropbox.com/s/abc123", revokedURL)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestShareLinkRevokeVerboseWritesStatusToStderr(t *testing.T) {
	stubSharedLinkClient(t, &mockSharedLinkClient{
		revokeSharedLinkFn: func(arg *sharing.RevokeSharedLinkArg) error {
			return nil
		},
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().Bool("verbose", true, "")
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := shareLinkRevoke(cmd, []string{"https://www.dropbox.com/s/abc123"})
	if err != nil {
		t.Fatalf("shareLinkRevoke error: %v", err)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if got, want := stderr.String(), "Revoked shared link https://www.dropbox.com/s/abc123\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestShareLinkRevokeReturnsAPIError(t *testing.T) {
	wantErr := fmt.Errorf("shared_link_not_found")
	stubSharedLinkClient(t, &mockSharedLinkClient{
		revokeSharedLinkFn: func(arg *sharing.RevokeSharedLinkArg) error {
			return wantErr
		},
	})

	err := shareLinkRevoke(&cobra.Command{}, []string{"https://www.dropbox.com/s/abc123"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want API error", err)
	}
}

func TestShareLinkRevokePathRequiresNoURL(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			called = true
			return nil, nil
		},
		revokeSharedLinkFn: func(arg *sharing.RevokeSharedLinkArg) error {
			called = true
			return nil
		},
	})

	cmd := newShareLinkRevokeTestCommand(nil, nil)
	if err := cmd.Flags().Set("path", "/docs/file.txt"); err != nil {
		t.Fatalf("set path: %v", err)
	}

	err := shareLinkRevoke(cmd, []string{"https://www.dropbox.com/s/abc123"})
	if err == nil || !strings.Contains(err.Error(), "`--path` cannot be used with a shared link URL") {
		t.Fatalf("error = %v, want path and URL conflict error", err)
	}
	details := jsonErrorDetails(err)
	if details["flag"] != "path" || details["argument"] != "url" || details["path"] != "/docs/file.txt" {
		t.Fatalf("details = %#v, want flag, url argument, and path", details)
	}
	if called {
		t.Fatal("shared link API should not be called")
	}
}

func TestShareLinkRevokePathRejectsEmptyPath(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			called = true
			return nil, nil
		},
	})

	cmd := newShareLinkRevokeTestCommand(nil, nil)
	if err := cmd.Flags().Set("path", ""); err != nil {
		t.Fatalf("set path: %v", err)
	}

	err := shareLinkRevoke(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "`--path` requires a non-empty path") {
		t.Fatalf("error = %v, want empty path error", err)
	}
	if called {
		t.Fatal("ListSharedLinks should not be called")
	}
}

func TestShareLinkRevokePathRejectsRoot(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			called = true
			return nil, nil
		},
	})

	cmd := newShareLinkRevokeTestCommand(nil, nil)
	if err := cmd.Flags().Set("path", "/"); err != nil {
		t.Fatalf("set path: %v", err)
	}

	err := shareLinkRevoke(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "cannot revoke shared links for Dropbox root") {
		t.Fatalf("error = %v, want root path error", err)
	}
	details := jsonErrorDetails(err)
	if details["operation"] != "share_link_revoke" || details["path"] != "/" {
		t.Fatalf("details = %#v, want share_link_revoke operation and root path", details)
	}
	if called {
		t.Fatal("ListSharedLinks should not be called")
	}
}

func TestShareLinkRevokePathListsDirectLinksAndRevokesAll(t *testing.T) {
	var listArg *sharing.ListSharedLinksArg
	var revoked []string
	stubSharedLinkClient(t, &mockSharedLinkClient{
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			listArg = arg
			return sharing.NewListSharedLinksResult([]sharing.IsSharedLinkMetadata{
				sharedLinkFile("/docs/file.txt", "https://example.com/one"),
				sharedLinkFile("/docs/file.txt", "https://example.com/two"),
			}, false), nil
		},
		revokeSharedLinkFn: func(arg *sharing.RevokeSharedLinkArg) error {
			revoked = append(revoked, arg.Url)
			return nil
		},
	})

	var stdout bytes.Buffer
	cmd := newShareLinkRevokeTestCommand(&stdout, nil)
	if err := cmd.Flags().Set("path", "docs/file.txt"); err != nil {
		t.Fatalf("set path: %v", err)
	}

	if err := shareLinkRevoke(cmd, nil); err != nil {
		t.Fatalf("shareLinkRevoke error: %v", err)
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
	if got := strings.Join(revoked, ","); got != "https://example.com/one,https://example.com/two" {
		t.Fatalf("revoked URLs = %q, want both direct links", got)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestShareLinkRevokePathFollowsPagination(t *testing.T) {
	var cursors []string
	var revoked []string
	stubSharedLinkClient(t, &mockSharedLinkClient{
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			cursors = append(cursors, arg.Cursor)
			if arg.Cursor == "" {
				res := sharing.NewListSharedLinksResult([]sharing.IsSharedLinkMetadata{
					sharedLinkFile("/docs/file.txt", "https://example.com/one"),
				}, true)
				res.Cursor = "next-page"
				return res, nil
			}
			return sharing.NewListSharedLinksResult([]sharing.IsSharedLinkMetadata{
				sharedLinkFile("/docs/file.txt", "https://example.com/two"),
			}, false), nil
		},
		revokeSharedLinkFn: func(arg *sharing.RevokeSharedLinkArg) error {
			revoked = append(revoked, arg.Url)
			return nil
		},
	})

	cmd := newShareLinkRevokeTestCommand(nil, nil)
	if err := cmd.Flags().Set("path", "/docs/file.txt"); err != nil {
		t.Fatalf("set path: %v", err)
	}

	if err := shareLinkRevoke(cmd, nil); err != nil {
		t.Fatalf("shareLinkRevoke error: %v", err)
	}
	if got := strings.Join(cursors, ","); got != ",next-page" {
		t.Fatalf("cursors = %q, want first call then next-page", got)
	}
	if got := strings.Join(revoked, ","); got != "https://example.com/one,https://example.com/two" {
		t.Fatalf("revoked URLs = %q, want both pages", got)
	}
}

func TestShareLinkRevokePathReturnsErrorWhenNoLinksFound(t *testing.T) {
	calledRevoke := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			return sharing.NewListSharedLinksResult(nil, false), nil
		},
		revokeSharedLinkFn: func(arg *sharing.RevokeSharedLinkArg) error {
			calledRevoke = true
			return nil
		},
	})

	cmd := newShareLinkRevokeTestCommand(nil, nil)
	if err := cmd.Flags().Set("path", "/docs/file.txt"); err != nil {
		t.Fatalf("set path: %v", err)
	}

	err := shareLinkRevoke(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), `no direct shared links found for "/docs/file.txt"`) {
		t.Fatalf("error = %v, want no links found error", err)
	}
	details := jsonErrorDetails(err)
	if details["operation"] != "share_link_revoke" || details["path"] != "/docs/file.txt" {
		t.Fatalf("details = %#v, want share_link_revoke operation and path", details)
	}
	if calledRevoke {
		t.Fatal("RevokeSharedLink should not be called")
	}
}

func TestShareLinkRevokePathReturnsListError(t *testing.T) {
	wantErr := fmt.Errorf("list failed")
	stubSharedLinkClient(t, &mockSharedLinkClient{
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			return nil, wantErr
		},
	})

	cmd := newShareLinkRevokeTestCommand(nil, nil)
	if err := cmd.Flags().Set("path", "/docs/file.txt"); err != nil {
		t.Fatalf("set path: %v", err)
	}

	err := shareLinkRevoke(cmd, nil)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want list error", err)
	}
	details := jsonErrorDetails(err)
	if details["operation"] != "share_link_revoke" || details["path"] != "/docs/file.txt" {
		t.Fatalf("details = %#v, want share_link_revoke operation and path", details)
	}
}

func TestShareLinkRevokePathReturnsRevokeError(t *testing.T) {
	wantErr := fmt.Errorf("revoke failed")
	stubSharedLinkClient(t, &mockSharedLinkClient{
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			return sharing.NewListSharedLinksResult([]sharing.IsSharedLinkMetadata{
				sharedLinkFile("/docs/file.txt", "https://example.com/one"),
			}, false), nil
		},
		revokeSharedLinkFn: func(arg *sharing.RevokeSharedLinkArg) error {
			return wantErr
		},
	})

	cmd := newShareLinkRevokeTestCommand(nil, nil)
	if err := cmd.Flags().Set("path", "/docs/file.txt"); err != nil {
		t.Fatalf("set path: %v", err)
	}

	err := shareLinkRevoke(cmd, nil)
	if err == nil {
		t.Fatal("expected wrapped revoke error")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapped revoke error", err)
	}
	if !strings.Contains(err.Error(), "revoke shared link https://example.com/one") {
		t.Fatalf("error = %v, want failing URL context", err)
	}
	details := jsonErrorDetails(err)
	if details["operation"] != "share_link_revoke" || details["url"] != "https://example.com/one" {
		t.Fatalf("details = %#v, want share_link_revoke URL context", details)
	}
}

func TestShareLinkRevokePathVerboseWritesStatusToStderr(t *testing.T) {
	stubSharedLinkClient(t, &mockSharedLinkClient{
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			return sharing.NewListSharedLinksResult([]sharing.IsSharedLinkMetadata{
				sharedLinkFile("/docs/file.txt", "https://example.com/one"),
				sharedLinkFile("/docs/file.txt", "https://example.com/two"),
			}, false), nil
		},
		revokeSharedLinkFn: func(arg *sharing.RevokeSharedLinkArg) error {
			return nil
		},
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newShareLinkRevokeTestCommand(&stdout, &stderr)
	if err := cmd.Flags().Set("path", "/docs/file.txt"); err != nil {
		t.Fatalf("set path: %v", err)
	}
	if err := cmd.Flags().Set("verbose", "true"); err != nil {
		t.Fatalf("set verbose: %v", err)
	}

	if err := shareLinkRevoke(cmd, nil); err != nil {
		t.Fatalf("shareLinkRevoke error: %v", err)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if got, want := stderr.String(), "Revoked 2 shared links for /docs/file.txt\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestShareLinkRevokeDoesNotBreakOtherCommands(t *testing.T) {
	cmd, _, err := RootCmd.Find([]string{"share-link", "revoke"})
	if err != nil {
		t.Fatalf("find share-link revoke: %v", err)
	}
	if cmd != shareLinkRevokeCmd {
		t.Fatalf("share-link revoke resolved to %q", cmd.CommandPath())
	}
	if shareLinkRevokeCmd.Flags().Lookup("path") == nil {
		t.Fatal("share-link revoke should define --path")
	}
	if shareLinkRevokeCmd.Use != "revoke [url]" {
		t.Fatalf("share-link revoke use = %q, want optional URL", shareLinkRevokeCmd.Use)
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
}

func newShareLinkRevokeTestCommand(stdout, stderr *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("path", "", "")
	cmd.Flags().Bool("verbose", false, "")
	if stdout != nil {
		cmd.SetOut(stdout)
	}
	if stderr != nil {
		cmd.SetErr(stderr)
	}
	return cmd
}
