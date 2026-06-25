package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/dropbox/dbxcli/internal/output"
	"github.com/spf13/cobra"
)

func TestStructuredOutputCommandAudit(t *testing.T) {
	got := structuredOutputCommandPaths(RootCmd)
	got = append(got, NewVersionCommand("test").Name())
	sort.Strings(got)

	want := []string{
		"account",
		"cp",
		"du",
		"get",
		"ls",
		"mkdir",
		"mv",
		"put",
		"restore",
		"revs",
		"rm",
		"search",
		"share list folder",
		"share list link",
		"share-link create",
		"share-link download",
		"share-link info",
		"share-link list",
		"share-link revoke",
		"share-link update",
		"team add-member",
		"team info",
		"team list-groups",
		"team list-members",
		"team remove-member",
		"version",
	}
	sort.Strings(want)

	if len(got) != len(want) {
		t.Fatalf("structured commands = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("structured commands = %v, want %v", got, want)
		}
	}
}

func structuredOutputCommandPaths(root *cobra.Command) []string {
	var paths []string
	var walk func(*cobra.Command, []string)
	walk = func(cmd *cobra.Command, parents []string) {
		parts := parents
		if cmd != root {
			parts = append(append([]string{}, parents...), cmd.Name())
			if commandSupportsStructuredOutput(cmd) {
				paths = append(paths, strings.Join(parts, " "))
			}
		}
		for _, child := range cmd.Commands() {
			walk(child, parts)
		}
	}
	walk(root, nil)
	return paths
}

func TestJSONOperationOutputContractShape(t *testing.T) {
	encoded, err := json.Marshal(newJSONOperationOutput(nil, nil, nil))
	if err != nil {
		t.Fatalf("marshal operation output: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatalf("decode operation output: %v", err)
	}
	for _, key := range []string{"input", "results", "warnings"} {
		if _, ok := raw[key]; !ok {
			t.Fatalf("operation output = %s, missing %q", encoded, key)
		}
	}
	if len(raw) != 3 {
		t.Fatalf("operation output = %s, want only input/results/warnings", encoded)
	}
}

func TestUnsupportedCommandsReturnJSONErrorEnvelope(t *testing.T) {
	for _, name := range []string{"login", "logout", "completion"} {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := &cobra.Command{Use: name}
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.Flags().String(outputFlag, "json", "")

			err := validateOutputFormat(cmd)
			if !errors.Is(err, output.ErrStructuredOutputUnsupported) {
				t.Fatalf("validateOutputFormat error = %v, want structured output unsupported", err)
			}

			renderCommandError(cmd, err)
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}

			got := decodeJSONErrorResponse(t, stdout.String())
			if got.OK {
				t.Fatalf("ok = true, want false")
			}
			if got.Error.Code != jsonErrorCodeStructuredOutputUnsupported {
				t.Fatalf("code = %q, want %q", got.Error.Code, jsonErrorCodeStructuredOutputUnsupported)
			}
			if got.Warnings == nil || len(got.Warnings) != 0 {
				t.Fatalf("warnings = %+v, want empty array", got.Warnings)
			}
		})
	}
}
