package cmd

import (
	"bytes"
	"io"
	"testing"

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

func TestCommandOutputHonorsJSONFlag(t *testing.T) {
	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.Flags().Bool(jsonOutputFlag, true, "")

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

func TestCommandOutputHonorsInheritedJSONFlag(t *testing.T) {
	var stdout bytes.Buffer
	root := &cobra.Command{}
	root.PersistentFlags().Bool(jsonOutputFlag, true, "")

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
