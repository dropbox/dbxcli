package cmd

import (
	"bytes"
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
	if err != wantErr {
		t.Fatalf("error = %v, want API error", err)
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
