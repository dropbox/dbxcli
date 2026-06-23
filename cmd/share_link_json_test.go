package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

func TestShareLinkCreateJSONOutputsLinkMetadata(t *testing.T) {
	expires := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	stubSharedLinkClient(t, &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			link := sharedLinkFile("/docs/report.txt", "https://example.com/report")
			link.Id = "id:file"
			link.Expires = &expires
			link.Size = 42
			return link, nil
		},
	})

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	setShareLinkOutputJSON(t, cmd)

	if err := shareLinkCreate(cmd, []string{"/docs/report.txt"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}

	var got shareLinkCreateOutput
	decodeJSONOutput(t, stdout.Bytes(), &got)
	if got.Input.Path != "/docs/report.txt" {
		t.Fatalf("input.path = %q, want /docs/report.txt", got.Input.Path)
	}
	if got.Existing {
		t.Fatal("existing = true, want false")
	}
	if got.Result.Type != "file" || got.Result.URL != "https://example.com/report" || got.Result.PathLower != "/docs/report.txt" {
		t.Fatalf("result = %#v, want file shared link metadata", got.Result)
	}
	if got.Result.Size == nil || *got.Result.Size != 42 {
		t.Fatalf("result.size = %#v, want 42", got.Result.Size)
	}
	if got.Result.Expires == nil || *got.Result.Expires != "2026-07-01T00:00:00Z" {
		t.Fatalf("result.expires = %#v, want RFC3339 expiration", got.Result.Expires)
	}
}

func TestShareLinkListJSONOutputsEntriesAndInput(t *testing.T) {
	stubSharedLinkClient(t, &mockSharedLinkClient{
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			if arg.Path != "/docs/report.txt" || !arg.DirectOnly {
				t.Fatalf("ListSharedLinks arg = %#v, want direct path filter", arg)
			}
			return sharing.NewListSharedLinksResult([]sharing.IsSharedLinkMetadata{
				sharedLinkFile("/docs/report.txt", "https://example.com/report"),
			}, false), nil
		},
	})

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	setShareLinkOutputJSON(t, cmd)

	if err := shareLinkList(cmd, []string{"/docs/report.txt"}); err != nil {
		t.Fatalf("shareLinkList error: %v", err)
	}

	var got shareLinkListOutput
	decodeJSONOutput(t, stdout.Bytes(), &got)
	if got.Input.Path != "/docs/report.txt" || !got.Input.DirectOnly {
		t.Fatalf("input = %#v, want direct path input", got.Input)
	}
	if len(got.Entries) != 1 || got.Entries[0].URL != "https://example.com/report" {
		t.Fatalf("entries = %#v, want report shared link", got.Entries)
	}
}

func TestShareLinkInfoJSONOutputsPermissions(t *testing.T) {
	permissions := sharing.NewLinkPermissions(true, nil, true, true, true, true, true, false, false)
	permissions.ResolvedVisibility = &sharing.ResolvedVisibility{Tagged: dropbox.Tagged{Tag: sharing.ResolvedVisibilityPublic}}
	permissions.LinkAccessLevel = &sharing.LinkAccessLevel{Tagged: dropbox.Tagged{Tag: sharing.LinkAccessLevelViewer}}
	permissions.RequirePassword = true

	link := sharedLinkFile("/docs/report.txt", "https://example.com/report")
	link.LinkPermissions = permissions

	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
			return link, nil
		},
	})

	var stdout bytes.Buffer
	cmd := newShareLinkInfoTestCommand(&stdout)
	setShareLinkOutputJSON(t, cmd)

	if err := shareLinkInfo(cmd, []string{"https://example.com/report"}); err != nil {
		t.Fatalf("shareLinkInfo error: %v", err)
	}

	var got shareLinkInfoOutput
	decodeJSONOutput(t, stdout.Bytes(), &got)
	if got.Input.URL != "https://example.com/report" {
		t.Fatalf("input.url = %q, want https://example.com/report", got.Input.URL)
	}
	if got.Result.Permissions == nil {
		t.Fatal("result.permissions = nil, want permissions")
	}
	if got.Result.Permissions.ResolvedVisibility != "public" || got.Result.Permissions.AccessLevel != "viewer" {
		t.Fatalf("permissions = %#v, want visibility and access level", got.Result.Permissions)
	}
	if !got.Result.Permissions.AllowDownload || !got.Result.Permissions.RequirePassword {
		t.Fatalf("permissions = %#v, want allow_download and require_password", got.Result.Permissions)
	}
}

func TestShareLinkUpdateJSONOutputsUpdatedMetadata(t *testing.T) {
	stubSharedLinkClient(t, &mockSharedLinkClient{
		modifySharedLinkSettingsFn: func(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error) {
			return sharedLinkFile("/docs/report.txt", arg.Url), nil
		},
	})

	var stdout bytes.Buffer
	cmd := newShareLinkUpdateTestCommand(&stdout, nil)
	setShareLinkOutputJSON(t, cmd)
	if err := cmd.Flags().Set("allow-download", "true"); err != nil {
		t.Fatalf("set allow-download: %v", err)
	}

	if err := shareLinkUpdate(cmd, []string{"https://example.com/report"}); err != nil {
		t.Fatalf("shareLinkUpdate error: %v", err)
	}

	var got shareLinkUpdateOutput
	decodeJSONOutput(t, stdout.Bytes(), &got)
	if got.Input.URL != "https://example.com/report" || !got.Input.AllowDownload {
		t.Fatalf("input = %#v, want update input", got.Input)
	}
	if got.Result.URL != "https://example.com/report" {
		t.Fatalf("result.url = %q, want updated URL", got.Result.URL)
	}
}

func TestShareLinkRevokeJSONOutputsRevokedURL(t *testing.T) {
	stubSharedLinkClient(t, &mockSharedLinkClient{
		revokeSharedLinkFn: func(arg *sharing.RevokeSharedLinkArg) error {
			return nil
		},
	})

	var stdout bytes.Buffer
	cmd := newShareLinkRevokeTestCommand(&stdout, nil)
	setShareLinkOutputJSON(t, cmd)

	if err := shareLinkRevoke(cmd, []string{"https://example.com/report"}); err != nil {
		t.Fatalf("shareLinkRevoke error: %v", err)
	}

	var got shareLinkRevokeOutput
	decodeJSONOutput(t, stdout.Bytes(), &got)
	if got.Input.URL != "https://example.com/report" {
		t.Fatalf("input.url = %q, want revoked URL", got.Input.URL)
	}
	if len(got.Revoked) != 1 || got.Revoked[0].URL != "https://example.com/report" {
		t.Fatalf("revoked = %#v, want revoked URL", got.Revoked)
	}
}

func TestShareLinkDownloadJSONOutputsTargetAndMetadata(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "report.txt")
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkMetadataFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error) {
			return sharedLinkFile("/docs/report.txt", arg.Url), nil
		},
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			return downloadableSharedLinkFile("report.txt", "/docs/report.txt", arg.Url, 7),
				io.NopCloser(strings.NewReader("content")), nil
		},
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newShareLinkDownloadTestCommand(&stdout, &stderr)
	setShareLinkOutputJSON(t, cmd)

	if err := shareLinkDownload(cmd, []string{"https://example.com/report", target}); err != nil {
		t.Fatalf("shareLinkDownload error: %v", err)
	}
	assertFileContent(t, target, "content")

	var got shareLinkDownloadOutput
	decodeJSONOutput(t, stdout.Bytes(), &got)
	if got.Input.URL != "https://example.com/report" || got.Input.Target != target {
		t.Fatalf("input = %#v, want download input", got.Input)
	}
	if got.Result.Target != target || got.Result.Link.URL != "https://example.com/report" {
		t.Fatalf("result = %#v, want target and link metadata", got.Result)
	}
}

func TestShareLinkDownloadJSONRejectsStdoutTarget(t *testing.T) {
	called := false
	stubSharedLinkClient(t, &mockSharedLinkClient{
		getSharedLinkFileFn: func(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error) {
			called = true
			return nil, nil, nil
		},
	})

	var stdout bytes.Buffer
	cmd := newShareLinkDownloadTestCommand(&stdout, nil)
	setShareLinkOutputJSON(t, cmd)

	err := shareLinkDownload(cmd, []string{"https://example.com/report", "-"})
	if err == nil || !strings.Contains(err.Error(), "cannot be used with --output=json") {
		t.Fatalf("error = %v, want JSON stdout rejection", err)
	}
	if called {
		t.Fatal("GetSharedLinkFile should not be called")
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestShareLinkCommandsSupportStructuredOutput(t *testing.T) {
	for _, cmd := range []*cobra.Command{
		shareLinkCreateCmd,
		shareLinkListCmd,
		shareListLinksCmd,
		shareLinkInfoCmd,
		shareLinkUpdateCmd,
		shareLinkRevokeCmd,
		shareLinkDownloadCmd,
	} {
		if !commandSupportsStructuredOutput(cmd) {
			t.Fatalf("%s should support structured output", cmd.CommandPath())
		}
	}
}

func setShareLinkOutputJSON(t *testing.T, cmd *cobra.Command) {
	t.Helper()
	if cmd.Flags().Lookup(outputFlag) == nil {
		cmd.Flags().String(outputFlag, "text", "")
	}
	if err := cmd.Flags().Set(outputFlag, "json"); err != nil {
		t.Fatalf("set output: %v", err)
	}
}

func decodeJSONOutput(t *testing.T, data []byte, out any) {
	t.Helper()
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("decode JSON output: %v\noutput: %s", err, string(data))
	}
}
