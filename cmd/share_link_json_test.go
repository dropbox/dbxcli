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

type shareLinkOperationOutputForTest[I, R any] struct {
	Input    I                                    `json:"input"`
	Results  []shareLinkOperationResultForTest[R] `json:"results"`
	Warnings []jsonWarning                        `json:"warnings"`
}

type shareLinkOperationResultForTest[R any] struct {
	Status string `json:"status"`
	Kind   string `json:"kind"`
	Result R      `json:"result"`
}

func TestShareLinkCreateJSONOutputsLinkMetadata(t *testing.T) {
	expires := dropbox.DBXTime(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))
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

	got := decodeShareLinkOperationOutput[shareLinkCreateInput, shareLinkJSONMetadata](t, stdout.Bytes())
	if got.Input.Path != "/docs/report.txt" {
		t.Fatalf("input.path = %q, want /docs/report.txt", got.Input.Path)
	}
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Results))
	}
	result := got.Results[0]
	if result.Status != shareLinkJSONStatusCreated || result.Kind != "file" {
		t.Fatalf("operation result = %#v, want created file", result)
	}
	if result.Result.Type != "file" || result.Result.URL != "https://example.com/report" || result.Result.PathLower != "/docs/report.txt" {
		t.Fatalf("result = %#v, want file shared link metadata", result.Result)
	}
	if result.Result.Size == nil || *result.Result.Size != 42 {
		t.Fatalf("result.size = %#v, want 42", result.Result.Size)
	}
	if result.Result.Expires == nil || *result.Result.Expires != "2026-07-01T00:00:00Z" {
		t.Fatalf("result.expires = %#v, want RFC3339 expiration", result.Result.Expires)
	}
	assertNoTopLevelJSONField(t, stdout.Bytes(), "existing")
}

func TestShareLinkCreateJSONReportsExistingAsStatus(t *testing.T) {
	existing := sharedLinkFolder("/docs", "https://example.com/docs")
	stubSharedLinkClient(t, &mockSharedLinkClient{
		createSharedLinkWithSettingsFn: func(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error) {
			return nil, alreadyExistsError(existing)
		},
	})

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	setShareLinkOutputJSON(t, cmd)

	if err := shareLinkCreate(cmd, []string{"/docs"}); err != nil {
		t.Fatalf("shareLinkCreate error: %v", err)
	}

	got := decodeShareLinkOperationOutput[shareLinkCreateInput, shareLinkJSONMetadata](t, stdout.Bytes())
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Results))
	}
	if got.Results[0].Status != shareLinkJSONStatusExisting || got.Results[0].Kind != "folder" {
		t.Fatalf("result = %#v, want existing folder", got.Results[0])
	}
	if got.Results[0].Result.URL != "https://example.com/docs" {
		t.Fatalf("result.url = %q, want existing URL", got.Results[0].Result.URL)
	}
	assertNoTopLevelJSONField(t, stdout.Bytes(), "existing")
}

func TestShareLinkListJSONOutputsResultsAndInput(t *testing.T) {
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

	got := decodeShareLinkOperationOutput[shareLinkListInput, shareLinkJSONMetadata](t, stdout.Bytes())
	if got.Input.Path != "/docs/report.txt" || !got.Input.DirectOnly {
		t.Fatalf("input = %#v, want direct path input", got.Input)
	}
	if len(got.Results) != 1 || got.Results[0].Status != shareLinkJSONStatusListed || got.Results[0].Kind != "file" || got.Results[0].Result.URL != "https://example.com/report" {
		t.Fatalf("results = %#v, want listed report shared link", got.Results)
	}
	if strings.Contains(stdout.String(), `"entries"`) {
		t.Fatalf("JSON output = %s, want operation results and no entries key", stdout.String())
	}
}

func TestDeprecatedShareListLinkJSONIncludesWarning(t *testing.T) {
	stubSharedLinkClient(t, &mockSharedLinkClient{
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			return sharing.NewListSharedLinksResult([]sharing.IsSharedLinkMetadata{
				sharedLinkFile("/docs/report.txt", "https://example.com/report"),
			}, false), nil
		},
	})

	var stdout bytes.Buffer
	cmd := &cobra.Command{Deprecated: shareListLinksDeprecatedMessage}
	cmd.SetOut(&stdout)
	setShareLinkOutputJSON(t, cmd)

	if err := shareListLinks(cmd, []string{"/docs/report.txt"}); err != nil {
		t.Fatalf("shareListLinks error: %v", err)
	}

	got := decodeShareLinkOperationOutputWithWarnings[shareLinkListInput, shareLinkJSONMetadata](t, stdout.Bytes())
	if len(got.Warnings) != 1 {
		t.Fatalf("warnings = %+v, want one deprecation warning", got.Warnings)
	}
	warning := got.Warnings[0]
	if warning.Code != jsonWarningCodeDeprecatedCommand {
		t.Fatalf("warning code = %q, want %q", warning.Code, jsonWarningCodeDeprecatedCommand)
	}
	if !strings.Contains(warning.Message, "share-link list") {
		t.Fatalf("warning message = %q, want share-link list", warning.Message)
	}
	if len(got.Results) != 1 || got.Results[0].Result.URL != "https://example.com/report" {
		t.Fatalf("results = %#v, want listed shared link", got.Results)
	}
}

func TestDeprecatedShareListLinkJSONKeepsDeprecationTextOffStdout(t *testing.T) {
	stubSharedLinkClient(t, &mockSharedLinkClient{
		listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
			return sharing.NewListSharedLinksResult([]sharing.IsSharedLinkMetadata{
				sharedLinkFile("/docs/report.txt", "https://example.com/report"),
			}, false), nil
		},
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := &cobra.Command{
		Use:        "link",
		Deprecated: shareListLinksDeprecatedMessage,
	}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	setShareLinkOutputJSON(t, cmd)

	if err := shareListLinksCmd.RunE(cmd, []string{"/docs/report.txt"}); err != nil {
		t.Fatalf("share list link error: %v", err)
	}

	if strings.Contains(stdout.String(), `Command "link" is deprecated`) {
		t.Fatalf("stdout = %q, want JSON only", stdout.String())
	}
	got := decodeShareLinkOperationOutputWithWarnings[shareLinkListInput, shareLinkJSONMetadata](t, stdout.Bytes())
	if len(got.Warnings) != 1 || got.Warnings[0].Code != jsonWarningCodeDeprecatedCommand {
		t.Fatalf("warnings = %+v, want deprecation warning", got.Warnings)
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, RunE should not print Cobra deprecation text directly", stderr.String())
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

	got := decodeShareLinkOperationOutput[shareLinkInfoInput, shareLinkJSONMetadata](t, stdout.Bytes())
	if got.Input.URL != "https://example.com/report" {
		t.Fatalf("input.url = %q, want https://example.com/report", got.Input.URL)
	}
	if len(got.Results) != 1 || got.Results[0].Status != shareLinkJSONStatusFound || got.Results[0].Kind != "file" {
		t.Fatalf("results = %#v, want found file", got.Results)
	}
	result := got.Results[0].Result
	if result.Permissions == nil {
		t.Fatal("result.permissions = nil, want permissions")
	}
	if result.Permissions.ResolvedVisibility != "public" || result.Permissions.AccessLevel != "viewer" {
		t.Fatalf("permissions = %#v, want visibility and access level", result.Permissions)
	}
	if !result.Permissions.AllowDownload || !result.Permissions.RequirePassword {
		t.Fatalf("permissions = %#v, want allow_download and require_password", result.Permissions)
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

	got := decodeShareLinkOperationOutput[shareLinkUpdateInput, shareLinkJSONMetadata](t, stdout.Bytes())
	if got.Input.URL != "https://example.com/report" || !got.Input.AllowDownload {
		t.Fatalf("input = %#v, want update input", got.Input)
	}
	if len(got.Results) != 1 || got.Results[0].Status != shareLinkJSONStatusUpdated || got.Results[0].Kind != "file" {
		t.Fatalf("results = %#v, want updated file", got.Results)
	}
	if got.Results[0].Result.URL != "https://example.com/report" {
		t.Fatalf("result.url = %q, want updated URL", got.Results[0].Result.URL)
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

	got := decodeShareLinkOperationOutput[shareLinkRevokeInput, shareLinkRevokeResult](t, stdout.Bytes())
	if got.Input.URL != "https://example.com/report" {
		t.Fatalf("input.url = %q, want revoked URL", got.Input.URL)
	}
	if len(got.Results) != 1 || got.Results[0].Status != shareLinkJSONStatusRevoked || got.Results[0].Kind != shareLinkJSONKindSharedLink || got.Results[0].Result.URL != "https://example.com/report" {
		t.Fatalf("results = %#v, want revoked URL", got.Results)
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

	got := decodeShareLinkOperationOutput[shareLinkDownloadInput, shareLinkDownloadResult](t, stdout.Bytes())
	if got.Input.URL != "https://example.com/report" || got.Input.Target != target {
		t.Fatalf("input = %#v, want download input", got.Input)
	}
	if len(got.Results) != 1 || got.Results[0].Status != shareLinkJSONStatusDownloaded || got.Results[0].Kind != "file" {
		t.Fatalf("results = %#v, want downloaded file", got.Results)
	}
	if got.Results[0].Result.Target != target || got.Results[0].Result.Link.URL != "https://example.com/report" {
		t.Fatalf("result = %#v, want target and link metadata", got.Results[0].Result)
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

func decodeShareLinkOperationOutput[I, R any](t *testing.T, data []byte) shareLinkOperationOutputForTest[I, R] {
	t.Helper()
	got := decodeShareLinkOperationOutputWithWarnings[I, R](t, data)
	if len(got.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want empty", got.Warnings)
	}
	return got
}

func decodeShareLinkOperationOutputWithWarnings[I, R any](t *testing.T, data []byte) shareLinkOperationOutputForTest[I, R] {
	t.Helper()
	var got shareLinkOperationOutputForTest[I, R]
	decodeJSONOutput(t, data, &got)
	if got.Warnings == nil {
		t.Fatalf("warnings = nil, want empty array")
	}
	return got
}

func assertNoTopLevelJSONField(t *testing.T, data []byte, field string) {
	t.Helper()
	var raw map[string]any
	decodeJSONOutput(t, data, &raw)
	if _, ok := raw[field]; ok {
		t.Fatalf("JSON output = %s, want no top-level %q field", string(data), field)
	}
}
