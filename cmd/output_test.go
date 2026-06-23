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
	"github.com/spf13/cobra"
)

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
	got := decodeJSONErrorResponse(t, stdout.String())
	if got.OK {
		t.Fatalf("ok = true, want false")
	}
	if got.Error.Message != "failed" {
		t.Fatalf("message = %q, want failed", got.Error.Message)
	}
	if got.Error.Code != "command_failed" {
		t.Fatalf("code = %q, want command_failed", got.Error.Code)
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

func TestJSONErrorCodePathConflict(t *testing.T) {
	err := errors.New("path exists and is not a folder: /file")
	if got, want := jsonErrorCode(err), "path_conflict"; got != want {
		t.Fatalf("jsonErrorCode = %q, want %q", got, want)
	}
}

func TestJSONErrorCodeOptionalArgumentValidation(t *testing.T) {
	err := errors.New("`account` accepts an optional `id` argument")
	if got, want := jsonErrorCode(err), "invalid_arguments"; got != want {
		t.Fatalf("jsonErrorCode = %q, want %q", got, want)
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
