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
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/dropbox/dbxcli/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

type revsInput struct {
	Path       string `json:"path"`
	Long       bool   `json:"long"`
	Time       string `json:"time,omitempty"`
	TimeFormat string `json:"time_format,omitempty"`
}

const revsJSONStatusRevision = "revision"

func revs(cmd *cobra.Command, args []string) (err error) {
	if len(args) != 1 {
		return invalidArgumentsError("`revs` requires a `file` argument")
	}

	path, err := validatePath(args[0])
	if err != nil {
		return
	}

	arg := files.NewListRevisionsArg(path)

	dbx := filesNewFunc(config)
	res, err := dbx.ListRevisions(arg)
	if err != nil {
		return
	}

	opts := parseLsOptions(cmd)

	return renderRevisionsOutput(cmd, path, res.Entries, opts)
}

func renderRevisionsOutput(cmd *cobra.Command, path string, entries []*files.FileMetadata, opts listOptions) error {
	out := commandOutput(cmd)
	if commandOutputFormat(cmd) != output.FormatJSON {
		return out.RenderText(func(w io.Writer) error {
			return renderRevisionResults(w, entries, opts)
		})
	}

	input := newRevsInput(path, opts)
	results := newJSONMetadataOperationResults(revsJSONStatusRevision, jsonMetadataListFromRevisions(entries))
	return renderJSONOperationOutput(cmd, input, results)
}

func newRevsInput(path string, opts listOptions) revsInput {
	return revsInput{
		Path:       path,
		Long:       opts.long,
		Time:       opts.timeField,
		TimeFormat: opts.timeFormat,
	}
}

func jsonMetadataListFromRevisions(entries []*files.FileMetadata) []jsonMetadata {
	result := make([]jsonMetadata, 0, len(entries))
	for _, entry := range entries {
		result = append(result, jsonMetadataFromDropbox(entry))
	}
	return result
}

func renderRevisionResults(out io.Writer, entries []*files.FileMetadata, opts listOptions) error {
	w := new(tabwriter.Writer)
	w.Init(out, 4, 8, 1, ' ', 0)

	if opts.long {
		_, _ = fmt.Fprint(w, "Revision\tSize\tLast modified\tPath\n")
	}

	for _, entry := range entries {
		if opts.long {
			_, _ = fmt.Fprint(w, formatFileMetadataWithOpts(entry, opts))
			_, _ = fmt.Fprintln(w)
		} else {
			_, _ = fmt.Fprintln(w, entry.Rev)
		}
	}

	return w.Flush()
}

// revsCmd represents the revs command
var revsCmd = &cobra.Command{
	Use:   "revs [flags] <file>",
	Short: "List file revisions",
	RunE:  revs,
}

func init() {
	RootCmd.AddCommand(revsCmd)

	revsCmd.Flags().BoolP("long", "l", false, "Long listing")
	revsCmd.Flags().String("time", "server", "Time field: server, client")
	revsCmd.Flags().String("time-format", "", "Time format: short (2006-01-02 15:04), rfc3339")
	enableStructuredOutput(revsCmd)
}
