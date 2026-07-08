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
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestAddDryRunFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}

	addDryRunFlag(cmd)

	flag := cmd.Flags().Lookup(dryRunFlagName)
	if flag == nil {
		t.Fatalf("%q flag not registered", dryRunFlagName)
	}
	if flag.Shorthand != "" {
		t.Fatalf("dry-run shorthand = %q, want none", flag.Shorthand)
	}
	enabled, err := dryRunEnabled(cmd)
	if err != nil {
		t.Fatalf("dryRunEnabled error: %v", err)
	}
	if enabled {
		t.Fatal("dry-run default = true, want false")
	}

	if err := cmd.Flags().Set(dryRunFlagName, "true"); err != nil {
		t.Fatal(err)
	}
	enabled, err = dryRunEnabled(cmd)
	if err != nil {
		t.Fatalf("dryRunEnabled error after set: %v", err)
	}
	if !enabled {
		t.Fatal("dry-run = false after setting flag")
	}
}

func TestPlannedStatus(t *testing.T) {
	if got, want := plannedStatus(false, "deleted"), "deleted"; got != want {
		t.Fatalf("plannedStatus(false) = %q, want %q", got, want)
	}
	if got, want := plannedStatus(true, "deleted"), jsonStatusPlanned; got != want {
		t.Fatalf("plannedStatus(true) = %q, want %q", got, want)
	}
}

func TestWriteDryRunLine(t *testing.T) {
	var stdout bytes.Buffer
	if err := writeDryRunLine(&stdout, "delete", "/File.txt"); err != nil {
		t.Fatalf("writeDryRunLine error: %v", err)
	}
	if got, want := stdout.String(), "Would delete /File.txt\n"; got != want {
		t.Fatalf("writeDryRunLine output = %q, want %q", got, want)
	}

	stdout.Reset()
	if err := writeDryRunLine(&stdout, "permanently delete", "/File.txt"); err != nil {
		t.Fatalf("writeDryRunLine permanent error: %v", err)
	}
	if got, want := stdout.String(), "Would permanently delete /File.txt\n"; got != want {
		t.Fatalf("writeDryRunLine permanent output = %q, want %q", got, want)
	}
}

func TestDryRunDisplayPath(t *testing.T) {
	if got, want := dryRunDisplayPath(jsonMetadata{PathDisplay: "/Display.txt"}, "/fallback.txt"), "/Display.txt"; got != want {
		t.Fatalf("dryRunDisplayPath with display path = %q, want %q", got, want)
	}
	if got, want := dryRunDisplayPath(jsonMetadata{}, "/fallback.txt"), "/fallback.txt"; got != want {
		t.Fatalf("dryRunDisplayPath fallback = %q, want %q", got, want)
	}
}
