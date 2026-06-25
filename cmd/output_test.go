package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/dropbox/dbxcli/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	dropboxauth "github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/auth"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

type metadataOperationOutputForTest[I any] struct {
	Input    I                                `json:"input"`
	Results  []metadataOperationResultForTest `json:"results"`
	Warnings []jsonWarning                    `json:"warnings"`
}

type metadataOperationResultForTest struct {
	Status string       `json:"status"`
	Kind   string       `json:"kind"`
	Result jsonMetadata `json:"result"`
}

func decodeMetadataOperationOutput[I any](t *testing.T, out *bytes.Buffer) metadataOperationOutputForTest[I] {
	t.Helper()

	var got metadataOperationOutputForTest[I]
	if err := json.NewDecoder(bytes.NewReader(out.Bytes())).Decode(&got); err != nil {
		t.Fatalf("decode JSON output: %v\noutput: %s", err, out.String())
	}
	if got.Warnings == nil {
		t.Fatalf("warnings = nil, want empty array")
	}
	if len(got.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want empty", got.Warnings)
	}
	return got
}

func TestCommandOutputUsesCobraWriters(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	out := commandOutput(cmd)
	out.Info("done")
	out.Error("failed: %d", 1)

	if got, want := stdout.String(), "done\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if got, want := stderr.String(), "Error: failed: 1\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestCommandOutputHonorsOutputJSON(t *testing.T) {
	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.Flags().String(outputFlag, "json", "")

	out := commandOutput(cmd)
	err := out.Render(func(w io.Writer) error {
		t.Fatal("text renderer should not be called for JSON output")
		return nil
	}, struct {
		Status string `json:"status"`
	}{Status: "ok"})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if got, want := stdout.String(), "{\"status\":\"ok\"}\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestCommandOutputHonorsInheritedOutputJSON(t *testing.T) {
	var stdout bytes.Buffer
	root := &cobra.Command{}
	root.PersistentFlags().String(outputFlag, "json", "")

	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	root.AddCommand(cmd)

	out := commandOutput(cmd)
	err := out.Render(func(w io.Writer) error {
		t.Fatal("text renderer should not be called for JSON output")
		return nil
	}, struct {
		Status string `json:"status"`
	}{Status: "ok"})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if got, want := stdout.String(), "{\"status\":\"ok\"}\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestCommandOutputHonorsRootPersistentOutputJSON(t *testing.T) {
	var stdout bytes.Buffer
	root := &cobra.Command{}
	root.SetOut(&stdout)
	root.PersistentFlags().String(outputFlag, "json", "")

	out := commandOutput(root)
	err := out.Render(func(w io.Writer) error {
		t.Fatal("text renderer should not be called for JSON output")
		return nil
	}, struct {
		Status string `json:"status"`
	}{Status: "ok"})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if got, want := stdout.String(), "{\"status\":\"ok\"}\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestValidateOutputFormatRejectsInvalidValue(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String(outputFlag, "yaml", "")

	err := validateOutputFormat(cmd)
	if err == nil {
		t.Fatal("expected invalid output format to fail")
	}
	if !strings.Contains(err.Error(), `unsupported output format "yaml": use text or json`) {
		t.Fatalf("error = %q, want unsupported output format", err.Error())
	}
}

func TestValidateOutputFormatRejectsUnsupportedJSONCommand(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String(outputFlag, "json", "")

	err := validateOutputFormat(cmd)
	if !errors.Is(err, output.ErrStructuredOutputUnsupported) {
		t.Fatalf("error = %v, want ErrStructuredOutputUnsupported", err)
	}
}

func TestValidateOutputFormatAllowsSupportedJSONCommand(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String(outputFlag, "json", "")
	enableStructuredOutput(cmd)

	if err := validateOutputFormat(cmd); err != nil {
		t.Fatalf("validateOutputFormat returned error: %v", err)
	}
}

func TestValidateOutputFormatHonorsInheritedOutput(t *testing.T) {
	root := &cobra.Command{}
	root.PersistentFlags().String(outputFlag, "json", "")

	cmd := &cobra.Command{}
	enableStructuredOutput(cmd)
	root.AddCommand(cmd)

	if err := validateOutputFormat(cmd); err != nil {
		t.Fatalf("validateOutputFormat returned error: %v", err)
	}
}

func TestStructuredOutputSupportDoesNotInheritFromParent(t *testing.T) {
	root := &cobra.Command{}
	root.PersistentFlags().String(outputFlag, "json", "")
	enableStructuredOutput(root)

	cmd := &cobra.Command{}
	root.AddCommand(cmd)

	err := validateOutputFormat(cmd)
	if !errors.Is(err, output.ErrStructuredOutputUnsupported) {
		t.Fatalf("error = %v, want ErrStructuredOutputUnsupported", err)
	}
}

func TestRootCommandDefinesOutputFlag(t *testing.T) {
	flag := RootCmd.PersistentFlags().Lookup(outputFlag)
	if flag == nil {
		t.Fatal("root command should define --output")
	}
	if got, want := flag.DefValue, "text"; got != want {
		t.Fatalf("--output default = %q, want %q", got, want)
	}
}

func TestCommandVerboseHonorsInheritedVerboseFlag(t *testing.T) {
	root := &cobra.Command{}
	root.PersistentFlags().BoolP("verbose", "v", false, "")

	cmd := &cobra.Command{}
	root.AddCommand(cmd)

	if err := root.PersistentFlags().Set("verbose", "true"); err != nil {
		t.Fatalf("set verbose: %v", err)
	}

	if !commandVerbose(cmd) {
		t.Fatal("commandVerbose = false, want true")
	}
}

func TestCommandVerboseStatusWritesOnlyWhenVerbose(t *testing.T) {
	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().Bool("verbose", false, "")
	cmd.SetErr(&stderr)

	commandVerboseStatus(cmd, "done %s", "quietly")
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}

	if err := cmd.Flags().Set("verbose", "true"); err != nil {
		t.Fatalf("set verbose: %v", err)
	}
	commandVerboseStatus(cmd, "done %s", "loudly")
	if got, want := stderr.String(), "done loudly\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestRenderCommandErrorWritesTextErrorToStderr(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.Flags().String(outputFlag, "text", "")

	renderCommandError(cmd, errors.New("failed"))

	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
	if got, want := stderr.String(), "Error: failed\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestRenderCommandErrorTextUnknownCommandIncludesUsageHint(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := &cobra.Command{Use: "dbxcli"}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.Flags().String(outputFlag, "text", "")

	renderCommandError(cmd, errors.New(`unknown command "missing" for "dbxcli"`))

	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
	want := "Error: unknown command \"missing\" for \"dbxcli\"\nRun 'dbxcli --help' for usage.\n"
	if got := stderr.String(); got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestRenderCommandErrorWritesJSONErrorToStdout(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.Flags().String(outputFlag, "json", "")

	renderCommandError(cmd, errors.New("failed"))

	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
	rendered := stdout.String()
	got := decodeJSONErrorResponse(t, rendered)
	if got.OK {
		t.Fatalf("ok = true, want false")
	}
	if got.Error.Message != "failed" {
		t.Fatalf("message = %q, want failed", got.Error.Message)
	}
	if got.Error.Code != "command_failed" {
		t.Fatalf("code = %q, want command_failed", got.Error.Code)
	}
	if len(got.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want empty", got.Warnings)
	}
	if !strings.Contains(rendered, `"warnings":[]`) {
		t.Fatalf("output = %q, want warnings array", rendered)
	}
}

func TestRenderCommandErrorWritesUnsupportedStructuredOutputAsJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.Flags().String(outputFlag, "json", "")

	renderCommandError(cmd, output.ErrStructuredOutputUnsupported)

	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
	got := decodeJSONErrorResponse(t, stdout.String())
	if got.Error.Code != "structured_output_unsupported" {
		t.Fatalf("code = %q, want structured_output_unsupported", got.Error.Code)
	}
	if !strings.Contains(got.Error.Message, "structured output is not supported") {
		t.Fatalf("message = %q, want structured output error", got.Error.Message)
	}
	if len(got.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want empty", got.Warnings)
	}
}

func TestRenderCommandErrorWithJSONForcesJSONError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.Flags().String(outputFlag, "text", "")

	renderCommandErrorWithJSON(cmd, errors.New(`unknown command "missing" for "dbxcli"`), true)

	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
	got := decodeJSONErrorResponse(t, stdout.String())
	if got.Error.Code != "unknown_command" {
		t.Fatalf("code = %q, want unknown_command", got.Error.Code)
	}
	if len(got.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want empty", got.Warnings)
	}
}

func TestNewJSONOperationOutputNormalizesWarnings(t *testing.T) {
	got := newJSONOperationOutput(
		struct {
			Path string `json:"path"`
		}{Path: "/file.txt"},
		[]jsonOperationResult{
			{
				Status: "downloaded",
				Kind:   "file",
				Input: struct {
					Path string `json:"path"`
				}{Path: "/file.txt"},
				Result: struct {
					Type string `json:"type"`
				}{Type: "file"},
			},
		},
		nil,
	)

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal operation output: %v", err)
	}
	if !strings.Contains(string(encoded), `"warnings":[]`) {
		t.Fatalf("encoded output = %s, want warnings array", encoded)
	}
	if got.Warnings == nil {
		t.Fatal("warnings is nil, want empty slice")
	}
}

func TestNewJSONOperationOutputNormalizesResults(t *testing.T) {
	got := newJSONOperationOutput(nil, nil, nil)

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal operation output: %v", err)
	}
	if !strings.Contains(string(encoded), `"results":[]`) {
		t.Fatalf("encoded output = %s, want results array", encoded)
	}
	if !strings.Contains(string(encoded), `"input":{}`) {
		t.Fatalf("encoded output = %s, want input object", encoded)
	}
	if got.Results == nil {
		t.Fatal("results is nil, want empty slice")
	}
}

func TestRenderJSONOperationOutput(t *testing.T) {
	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.Flags().String(outputFlag, "json", "")

	err := renderJSONOperationOutput(
		cmd,
		struct {
			Path string `json:"path"`
		}{Path: "/file.txt"},
		[]jsonOperationResult{
			newJSONOperationResult(
				"downloaded",
				"file",
				struct {
					Source string `json:"source"`
				}{Source: "/file.txt"},
				struct {
					Type string `json:"type"`
				}{Type: "file"},
			),
		},
	)
	if err != nil {
		t.Fatalf("renderJSONOperationOutput returned error: %v", err)
	}

	rendered := stdout.String()
	for _, want := range []string{
		`"input":{"path":"/file.txt"}`,
		`"results":[{"status":"downloaded","kind":"file"`,
		`"warnings":[]`,
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("output = %s, want %s", rendered, want)
		}
	}
}

func TestNewJSONOperationResultOmitsEmptyFields(t *testing.T) {
	got := newJSONOperationResult(
		"",
		"",
		struct {
			Path string `json:"path"`
		}{Path: "/file.txt"},
		nil,
	)

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal operation result: %v", err)
	}
	rendered := string(encoded)
	for _, unwanted := range []string{`"status"`, `"kind"`, `"result"`} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("encoded output = %s, did not expect %s", rendered, unwanted)
		}
	}
	if !strings.Contains(rendered, `"input":{"path":"/file.txt"}`) {
		t.Fatalf("encoded output = %s, want input", rendered)
	}
}

func TestNewJSONMetadataOperationResults(t *testing.T) {
	size := uint64(42)
	results := newJSONMetadataOperationResults("listed", []jsonMetadata{
		{
			Type:        "file",
			PathDisplay: "/file.txt",
			Size:        &size,
		},
		{
			Type:        "folder",
			PathDisplay: "/folder",
		},
	})

	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
	if results[0].Status != "listed" || results[0].Kind != "file" || results[0].Input != nil {
		t.Fatalf("first result = %#v, want listed file with no per-result input", results[0])
	}
	first, ok := results[0].Result.(jsonMetadata)
	if !ok {
		t.Fatalf("first result metadata type = %T, want jsonMetadata", results[0].Result)
	}
	if first.PathDisplay != "/file.txt" || first.Size == nil || *first.Size != 42 {
		t.Fatalf("first metadata = %#v, want file metadata", first)
	}
	if results[1].Status != "listed" || results[1].Kind != "folder" {
		t.Fatalf("second result = %#v, want listed folder", results[1])
	}
}

func TestRenderCommandErrorInvalidOutputFormatFallsBackToText(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.Flags().String(outputFlag, "yaml", "")

	err := fmt.Errorf(`unsupported output format "yaml": use text or json`)
	renderCommandError(cmd, err)

	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
	if got, want := stderr.String(), "Error: unsupported output format \"yaml\": use text or json\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestOutputJSONRequested(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{
			name: "equals",
			args: []string{"--output=json", "missing"},
			want: true,
		},
		{
			name: "separate",
			args: []string{"--output", "json", "missing"},
			want: true,
		},
		{
			name: "text",
			args: []string{"--output=text", "missing"},
			want: false,
		},
		{
			name: "invalid format",
			args: []string{"--output", "yaml", "missing"},
			want: false,
		},
		{
			name: "invalid format before json",
			args: []string{"--output", "yaml", "--output", "json", "missing"},
			want: true,
		},
		{
			name: "invalid format after json",
			args: []string{"--output", "json", "--output", "yaml", "missing"},
			want: false,
		},
		{
			name: "last separate flag wins text",
			args: []string{"--output", "json", "--output", "text", "missing"},
			want: false,
		},
		{
			name: "last separate flag wins json",
			args: []string{"--output", "text", "--output", "json", "missing"},
			want: true,
		},
		{
			name: "last equals flag wins text",
			args: []string{"--output=json", "--output=text", "missing"},
			want: false,
		},
		{
			name: "last equals flag wins json",
			args: []string{"--output=text", "--output=json", "missing"},
			want: true,
		},
		{
			name: "after double dash",
			args: []string{"mkdir", "--", "--output=json"},
			want: false,
		},
		{
			name: "output before double dash",
			args: []string{"--output=json", "mkdir", "--", "--output=text"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := outputJSONRequested(tt.args); got != tt.want {
				t.Fatalf("outputJSONRequested(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestJSONErrorCodeUsesCodedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "path conflict",
			err:  pathConflictErrorf("path exists and is not a folder: /file"),
			want: jsonErrorCodePathConflict,
		},
		{
			name: "optional argument validation",
			err:  invalidArgumentsError("`account` accepts an optional `id` argument"),
			want: jsonErrorCodeInvalidArguments,
		},
		{
			name: "required argument validation",
			err:  invalidArgumentsError("`add-member` requires `email`, `first`, and `last` arguments"),
			want: jsonErrorCodeInvalidArguments,
		},
		{
			name: "unsupported output format",
			err:  unsupportedOutputFormatErrorf("unsupported output format %q: use text or json", "yaml"),
			want: jsonErrorCodeUnsupportedOutputFormat,
		},
		{
			name: "auth required",
			err:  authRequiredErrorf("no saved Dropbox credentials"),
			want: jsonErrorCodeAuthRequired,
		},
		{
			name: "app key required",
			err:  appKeyRequiredError("Dropbox app key is required"),
			want: jsonErrorCodeAppKeyRequired,
		},
		{
			name: "auth exchange failed",
			err:  authExchangeFailedError("authorization did not return an access token"),
			want: jsonErrorCodeAuthExchangeFailed,
		},
		{
			name: "auth refresh failed",
			err:  authRefreshFailedErrorf("refresh saved Dropbox credentials: failed"),
			want: jsonErrorCodeAuthRefreshFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jsonErrorCode(tt.err); got != tt.want {
				t.Fatalf("jsonErrorCode = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestJSONErrorCodeClassifiesDropboxAPIErrors(t *testing.T) {
	expiredAuth := dropboxauth.AuthAPIError{AuthError: &dropboxauth.AuthError{}}
	expiredAuth.AuthError.Tag = dropboxauth.AuthErrorExpiredAccessToken
	missingScope := dropboxauth.AuthAPIError{AuthError: &dropboxauth.AuthError{}}
	missingScope.AuthError.Tag = dropboxauth.AuthErrorMissingScope

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "rate limit api error",
			err:  dropboxauth.RateLimitAPIError{},
			want: jsonErrorCodeRateLimited,
		},
		{
			name: "expired access token",
			err:  expiredAuth,
			want: jsonErrorCodeAuthRequired,
		},
		{
			name: "missing scope",
			err:  missingScope,
			want: jsonErrorCodePermissionDenied,
		},
		{
			name: "access api error",
			err:  dropboxauth.AccessAPIError{},
			want: jsonErrorCodePermissionDenied,
		},
		{
			name: "typed files api error",
			err:  files.GetMetadataAPIError{APIError: dropbox.APIError{ErrorSummary: "path/not_found/"}},
			want: jsonErrorCodeNotFound,
		},
		{
			name: "wrapped typed files api error",
			err:  fmt.Errorf("get metadata: %w", files.GetMetadataAPIError{APIError: dropbox.APIError{ErrorSummary: "path/not_found/"}}),
			want: jsonErrorCodeNotFound,
		},
		{
			name: "typed path conflict api error",
			err:  files.CreateFolderV2APIError{APIError: dropbox.APIError{ErrorSummary: "path/conflict/folder/"}},
			want: jsonErrorCodePathConflict,
		},
		{
			name: "relocation conflict summary",
			err:  errors.New(`Error in call to API function "files/move_v2": to/conflict/file/.`),
			want: jsonErrorCodePathConflict,
		},
		{
			name: "not found summary",
			err:  errors.New(`Error in call to API function "files/get_metadata": path/not_found/.`),
			want: jsonErrorCodeNotFound,
		},
		{
			name: "exact api summary",
			err:  errors.New("path/not_found/"),
			want: jsonErrorCodeNotFound,
		},
		{
			name: "exact auth summary",
			err:  errors.New("invalid_access_token/"),
			want: jsonErrorCodeAuthRequired,
		},
		{
			name: "wrapped api summary text",
			err:  errors.New("get metadata for /missing: path/not_found/"),
			want: jsonErrorCodeNotFound,
		},
		{
			name: "permission summary",
			err:  errors.New(`Error in call to API function "files/list_folder": path/no_permission/.`),
			want: jsonErrorCodePermissionDenied,
		},
		{
			name: "rate limit summary",
			err:  errors.New(`Error in call to API function "files/upload": too_many_requests/...`),
			want: jsonErrorCodeRateLimited,
		},
		{
			name: "generic api function error",
			err:  errors.New(`Error in call to API function "sharing/create_shared_link_with_settings": shared_link_already_exists/.`),
			want: jsonErrorCodeDropboxAPIError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jsonErrorCode(tt.err); got != tt.want {
				t.Fatalf("jsonErrorCode = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestJSONErrorCodeDoesNotClassifyPlainValidationStrings(t *testing.T) {
	for _, err := range []error{
		errors.New("path exists and is not a folder: /file"),
		errors.New("Dropbox API requires team admin permissions"),
		errors.New("`account` accepts an optional `id` argument"),
		errors.New("local cache not found"),
		errors.New("config missing_scope marker"),
		errors.New("local path/not_found/ marker"),
	} {
		if got := jsonErrorCode(err); got != jsonErrorCodeCommandFailed {
			t.Fatalf("jsonErrorCode(%q) = %q, want %q", err.Error(), got, jsonErrorCodeCommandFailed)
		}
	}
}

func decodeJSONErrorResponse(t *testing.T, value string) jsonErrorResponse {
	t.Helper()

	var got jsonErrorResponse
	if err := json.Unmarshal([]byte(value), &got); err != nil {
		t.Fatalf("decode JSON error response: %v\noutput: %s", err, value)
	}
	return got
}
