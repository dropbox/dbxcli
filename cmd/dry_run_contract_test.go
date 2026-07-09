// Copyright © 2026 Dropbox, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// dryRunCommands is the expected set of structured-output commands that expose
// a --dry-run flag. It anchors the cross-command dry-run contract: adding
// --dry-run to one of those commands means adding it here, which forces the
// coupled edits (flag registration, input schema, planned status) to be
// verified together instead of drifting apart.
//
// Per-command tests already cover the runtime behavior (planned JSON/text
// output, and that no write API is called). This registry guards the
// invariants that span all of them at once, so a new dry-run command cannot
// silently skip one of the conventions.
var dryRunCommands = []string{
	"cp",
	"mkdir",
	"mv",
	"put",
	"restore",
	"rm",
	"share-link create",
	"share-link revoke",
	"share-link update",
}

// TestDryRunRegistryMatchesRegisteredFlags asserts the registry and the set of
// structured-output commands that actually register --dry-run are exactly
// equal. This catches both directions of drift for dry-run commands that are
// expected to publish the JSON contract checked below.
func TestDryRunRegistryMatchesRegisteredFlags(t *testing.T) {
	registered := commandsWithDryRunFlag(t)

	want := append([]string{}, dryRunCommands...)
	sort.Strings(want)
	got := append([]string{}, registered...)
	sort.Strings(got)

	if !slices.Equal(got, want) {
		t.Fatalf("commands with --dry-run = %v, want registry %v", got, want)
	}
}

// TestDryRunCommandsDeclarePlannedStatus asserts every dry-run command allows
// the "planned" result status in both its manifest contract metadata and its
// code-derived JSON schema. Without this a dry-run branch could emit a
// "planned" result that fails schema validation for consumers.
func TestDryRunCommandsDeclarePlannedStatus(t *testing.T) {
	schemas := jsonCommandSchemas()

	for _, command := range dryRunCommands {
		contract, ok := commandContractRegistry[command]
		if !ok {
			t.Errorf("dry-run command %q has no contract registry entry", command)
		} else if !slices.Contains(contract.Statuses, jsonStatusPlanned) {
			t.Errorf("dry-run command %q contract statuses = %v, want to include %q", command, contract.Statuses, jsonStatusPlanned)
		}

		schema, ok := schemas[command]
		if !ok {
			t.Errorf("dry-run command %q has no code-derived schema", command)
			continue
		}
		if !slices.Contains(schema.Statuses, jsonStatusPlanned) {
			t.Errorf("dry-run command %q schema statuses = %v, want to include %q", command, schema.Statuses, jsonStatusPlanned)
		}
	}
}

// TestDryRunCommandsExposeDryRunInInputSchema asserts every dry-run command
// advertises a dry_run property in its published input schema, so machine
// consumers can discover the flag.
func TestDryRunCommandsExposeDryRunInInputSchema(t *testing.T) {
	for _, command := range dryRunCommands {
		cmd := findCommandByPath(t, command)
		meta := commandManifestMetadataFor(command)
		flags := jsonCommandFlags(cmd, meta.Flags)
		schema := commandInputSchemaFor(meta.Args, commandInputSchemaFlags(cmd, flags))
		if _, ok := schema.Properties["dry_run"]; !ok {
			t.Errorf("dry-run command %q input schema properties = %v, want to include dry_run", command, inputSchemaPropertyNames(schema))
		}
	}
}

// TestDryRunCommandsExposeDryRunInResultInputSchema asserts every dry-run
// command uses a result input schema that carries dry_run too. This keeps the
// runtime JSON convention machine-checkable: dry-run is visible both at the
// command input level and on each planned operation result.
func TestDryRunCommandsExposeDryRunInResultInputSchema(t *testing.T) {
	schemas := jsonCommandSchemas()
	definitions := jsonContractDefinitions()

	for _, command := range dryRunCommands {
		schema, ok := schemas[command]
		if !ok {
			t.Errorf("dry-run command %q has no code-derived schema", command)
			continue
		}
		if schema.ResultInput == nil {
			t.Errorf("dry-run command %q result input schema = nil, want schema with dry_run", command)
			continue
		}

		resultInput := *schema.ResultInput
		fields, ok := definitions[resultInput]
		if !ok {
			t.Errorf("dry-run command %q result input schema %q has no definition", command, resultInput)
			continue
		}
		if !slices.Contains(fields, "dry_run") {
			t.Errorf("dry-run command %q result input schema %q fields = %v, want to include dry_run", command, resultInput, fields)
		}
	}
}

func commandsWithDryRunFlag(t *testing.T) []string {
	t.Helper()

	var commands []string
	// Dry-run currently belongs to structured-output mutation commands. If a
	// non-structured command grows --dry-run, that command should first define a
	// JSON contract so the dry-run schema invariants below can cover it.
	for _, path := range structuredOutputCommandPaths(RootCmd) {
		cmd := findCommandByPath(t, path)
		if cmd.Flags().Lookup(dryRunFlagName) != nil {
			commands = append(commands, path)
		}
	}
	return commands
}

func findCommandByPath(t *testing.T, path string) *cobra.Command {
	t.Helper()

	cmd, _, err := RootCmd.Find(strings.Split(path, " "))
	if err != nil {
		t.Fatalf("find command %q: %v", path, err)
	}
	if cmd == RootCmd {
		t.Fatalf("command %q resolved to root", path)
	}
	return cmd
}

func inputSchemaPropertyNames(schema jsonCommandInputSchema) []string {
	names := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
