// Copyright © 2016 Dropbox, Inc.
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
	"fmt"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

// relocationCommand invokes mv or cp through its command function so the tests
// below exercise the shared runRelocation driver via both specs.
type relocationCommand struct {
	name          string
	verb          string // verb woven into per-operation error messages
	missingArgs   string // exact missing-args message
	run           func(cmd *cobra.Command, args []string) error
	successStatus string
}

// relocationCommands returns mv and cp so each behavior can be table-tested
// against both, confirming the spec substitution (verb, status, API call,
// error text) is correct for each command.
func relocationCommands() []relocationCommand {
	return []relocationCommand{
		{name: "mv", verb: "move", missingArgs: "mv command requires a source and a destination", run: mv, successStatus: relocationJSONStatusMoved},
		{name: "cp", verb: "copy", missingArgs: "cp requires a source and a destination", run: cp, successStatus: relocationJSONStatusCopied},
	}
}

func TestRunRelocationMissingArgsMessagePerCommand(t *testing.T) {
	for _, rc := range relocationCommands() {
		t.Run(rc.name, func(t *testing.T) {
			cmd := newRelocationTestCommand(nil, nil)
			for _, args := range [][]string{{}, {"/only-one"}} {
				err := rc.run(cmd, args)
				if err == nil {
					t.Fatalf("%s(%v): expected error, got nil", rc.name, args)
				}
				if !strings.Contains(err.Error(), rc.missingArgs) {
					t.Fatalf("%s(%v) error = %q, want to contain %q", rc.name, args, err.Error(), rc.missingArgs)
				}
			}
		})
	}
}

func TestRunRelocationDryRunSourceLookupErrorUsesVerb(t *testing.T) {
	for _, rc := range relocationCommands() {
		t.Run(rc.name, func(t *testing.T) {
			stubFilesClient(t, &mockFilesClient{
				getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
					return nil, fmt.Errorf("boom")
				},
				moveV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
					t.Fatalf("%s: MoveV2 called during dry-run", rc.name)
					return nil, nil
				},
				copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
					t.Fatalf("%s: CopyV2 called during dry-run", rc.name)
					return nil, nil
				},
			})

			var stderr bytes.Buffer
			cmd := newRelocationTextTestCommand(nil, &stderr)
			if err := cmd.Flags().Set(dryRunFlagName, "true"); err != nil {
				t.Fatal(err)
			}
			err := rc.run(cmd, []string{"/src/file.txt", "/dest/file.txt"})
			if err == nil {
				t.Fatalf("%s: expected error from failed source lookup, got nil", rc.name)
			}
			if !strings.Contains(stderr.String(), rc.verb+" ") {
				t.Fatalf("%s stderr = %q, want to contain verb %q", rc.name, stderr.String(), rc.verb)
			}
		})
	}
}

func TestRunRelocationSkipCheckErrorUsesVerb(t *testing.T) {
	for _, rc := range relocationCommands() {
		t.Run(rc.name, func(t *testing.T) {
			stubFilesClient(t, &mockFilesClient{
				// relocationSkipIfDestinationExists (via if-exists=skip) probes the
				// destination with GetMetadata; a non-not-found error surfaces here.
				getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
					return nil, fmt.Errorf("transient network failure")
				},
			})

			var stderr bytes.Buffer
			cmd := newRelocationTextTestCommand(nil, &stderr)
			if err := cmd.Flags().Set("if-exists", relocationIfExistsSkip); err != nil {
				t.Fatal(err)
			}
			err := rc.run(cmd, []string{"/src/file.txt", "/dest/file.txt"})
			if err == nil {
				t.Fatalf("%s: expected error from destination check, got nil", rc.name)
			}
			if !strings.Contains(stderr.String(), rc.verb+" ") {
				t.Fatalf("%s stderr = %q, want to contain verb %q", rc.name, stderr.String(), rc.verb)
			}
		})
	}
}

func TestRunRelocationResultBuildErrorUsesVerb(t *testing.T) {
	for _, rc := range relocationCommands() {
		t.Run(rc.name, func(t *testing.T) {
			// A successful API call that returns metadata newRelocationResult
			// cannot decode (a nil typed *FileMetadata) drives the post-execute
			// result-building error branch.
			okResult := &files.RelocationResult{Metadata: (*files.FileMetadata)(nil)}
			stubFilesClient(t, &mockFilesClient{
				getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
					return nil, relocationTestGetMetadataNotFoundError()
				},
				moveV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
					return okResult, nil
				},
				copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
					return okResult, nil
				},
			})

			var stderr bytes.Buffer
			cmd := newRelocationTestCommand(nil, &stderr) // JSON mode: exercises collectResults path
			err := rc.run(cmd, []string{"/src/file.txt", "/dest/file.txt"})
			if err == nil {
				t.Fatalf("%s: expected result-build error, got nil", rc.name)
			}
			if !strings.Contains(stderr.String(), rc.verb+" ") {
				t.Fatalf("%s stderr = %q, want to contain verb %q", rc.name, stderr.String(), rc.verb)
			}
		})
	}
}

func TestRunRelocationCallsCorrectAPIForSpec(t *testing.T) {
	for _, rc := range relocationCommands() {
		t.Run(rc.name, func(t *testing.T) {
			var moveCalls, copyCalls int
			stubFilesClient(t, &mockFilesClient{
				getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
					return nil, relocationTestGetMetadataNotFoundError()
				},
				moveV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
					moveCalls++
					return files.NewRelocationResult(relocationTestFileMetadata(arg.ToPath, 1)), nil
				},
				copyV2Fn: func(arg *files.RelocationArg) (*files.RelocationResult, error) {
					copyCalls++
					return files.NewRelocationResult(relocationTestFileMetadata(arg.ToPath, 1)), nil
				},
			})

			var stdout bytes.Buffer
			cmd := newRelocationTestCommand(&stdout, nil)
			if err := rc.run(cmd, []string{"/src/file.txt", "/dest/file.txt"}); err != nil {
				t.Fatalf("%s error: %v", rc.name, err)
			}

			wantMove, wantCopy := 0, 1
			if rc.name == "mv" {
				wantMove, wantCopy = 1, 0
			}
			if moveCalls != wantMove || copyCalls != wantCopy {
				t.Fatalf("%s: moveCalls=%d copyCalls=%d, want %d/%d", rc.name, moveCalls, copyCalls, wantMove, wantCopy)
			}

			got := decodeRelocationOutput(t, stdout.Bytes())
			if len(got.Results) != 1 {
				t.Fatalf("%s: results = %d, want 1", rc.name, len(got.Results))
			}
			if got.Results[0].Status != rc.successStatus {
				t.Fatalf("%s: status = %q, want %q", rc.name, got.Results[0].Status, rc.successStatus)
			}
		})
	}
}
