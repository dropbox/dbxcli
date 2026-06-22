package cmd

import (
	"bytes"
	"errors"
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
