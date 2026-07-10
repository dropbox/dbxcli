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
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

const (
	dryRunFlagName    = "dry-run"
	jsonStatusPlanned = "planned"
)

// Dry-run support is limited here to shared flags and output conventions.
// Command implementations should keep their validate/plan/execute flow local
// until at least two commands need the same execution driver.
func addDryRunFlag(cmd *cobra.Command) {
	cmd.Flags().Bool(dryRunFlagName, false, "Preview intended writes without making changes")
}

func dryRunEnabled(cmd *cobra.Command) (bool, error) {
	return cmd.Flags().GetBool(dryRunFlagName)
}

// dryRunOptionalEnabled returns the dry-run flag state, or false when the flag
// is not registered. Option parsers shared with callers that do not register
// --dry-run use this instead of dryRunEnabled to avoid a missing-flag error.
func dryRunOptionalEnabled(cmd *cobra.Command) (bool, error) {
	if cmd == nil || cmd.Flags().Lookup(dryRunFlagName) == nil {
		return false, nil
	}
	return dryRunEnabled(cmd)
}

// Dry-run JSON results use status "planned" and include dry_run=true in the
// result input. They do not preserve the real mutation status, because no
// mutation happened.
func plannedStatus(dryRun bool, realStatus string) string {
	if dryRun {
		return jsonStatusPlanned
	}
	return realStatus
}

// renderOperation renders a command's operation output through the shared
// command output writer: the text closure produces the human-readable form and
// the JSON form is the standard operation envelope. Dry-run branches use this
// to avoid duplicating the commandOutput/newJSONCommandOperationOutput glue.
func renderOperation(cmd *cobra.Command, input any, results []jsonOperationResult, warnings []jsonWarning, text func(io.Writer) error) error {
	return commandOutput(cmd).Render(text, newJSONCommandOperationOutput(cmd, input, results, warnings))
}

func writeDryRunLine(w io.Writer, verb, path string) error {
	_, err := fmt.Fprintf(w, "Would %s %s\n", verb, path)
	return err
}

func writeDryRunRelocationLine(w io.Writer, verb, fromPath, toPath string) error {
	_, err := fmt.Fprintf(w, "Would %s %s to %s\n", verb, fromPath, toPath)
	return err
}

func renderPlannedRelocationResults(w io.Writer, verb string, results []relocationResult) error {
	for _, result := range results {
		if err := writeDryRunRelocationLine(w, verb, result.Input.FromPath, result.Input.ToPath); err != nil {
			return err
		}
	}
	return nil
}

func plannedMetadata(kind, path string) jsonMetadata {
	return jsonMetadata{
		Type:        kind,
		PathDisplay: path,
		PathLower:   strings.ToLower(path),
	}
}

func dryRunDisplayPath(metadata jsonMetadata, fallback string) string {
	if metadata.PathDisplay != "" {
		return metadata.PathDisplay
	}
	return fallback
}
