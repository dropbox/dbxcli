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
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/dropbox/dbxcli/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

type searchInput struct {
	Query      string `json:"query"`
	Path       string `json:"path,omitempty"`
	Long       bool   `json:"long"`
	Sort       string `json:"sort,omitempty"`
	Reverse    bool   `json:"reverse"`
	Time       string `json:"time,omitempty"`
	TimeFormat string `json:"time_format,omitempty"`
}

type searchOutput struct {
	Input   searchInput    `json:"input"`
	Entries []jsonMetadata `json:"entries"`
}

func search(cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 {
		return errors.New("`search` requires a `query` argument")
	}

	var scope string
	if len(args) == 2 {
		scope = args[1]
		if !strings.HasPrefix(scope, "/") {
			return errors.New("`search` `path-scope` must begin with \"/\"")
		}
	}

	arg := files.NewSearchV2Arg(args[0])
	if scope != "" {
		opts := files.NewSearchOptions()
		opts.Path = scope
		arg.Options = opts
	}

	dbx := filesNewFunc(config)
	res, err := dbx.SearchV2(arg)
	if err != nil {
		return err
	}

	var entries []files.IsMetadata
	for _, m := range res.Matches {
		if m.Metadata != nil && m.Metadata.Metadata != nil {
			entries = append(entries, m.Metadata.Metadata)
		}
	}

	for res.HasMore {
		contArg := files.NewSearchV2ContinueArg(res.Cursor)
		res, err = dbx.SearchContinueV2(contArg)
		if err != nil {
			return err
		}
		for _, m := range res.Matches {
			if m.Metadata != nil && m.Metadata.Metadata != nil {
				entries = append(entries, m.Metadata.Metadata)
			}
		}
	}

	opts := parseLsOptions(cmd)
	sortEntries(entries, opts)

	return renderSearchOutput(cmd, args[0], scope, entries, opts)
}

func renderSearchOutput(cmd *cobra.Command, query, scope string, entries []files.IsMetadata, opts listOptions) error {
	out := commandOutput(cmd)
	if commandOutputFormat(cmd) != output.FormatJSON {
		return out.RenderText(func(w io.Writer) error {
			return renderSearchResults(w, entries, opts)
		})
	}

	return out.Render(nil, searchOutput{
		Input:   newSearchInput(query, scope, opts),
		Entries: jsonMetadataListFromDropbox(entries),
	})
}

func newSearchInput(query, scope string, opts listOptions) searchInput {
	return searchInput{
		Query:      query,
		Path:       scope,
		Long:       opts.long,
		Sort:       opts.sortBy,
		Reverse:    opts.reverse,
		Time:       opts.timeField,
		TimeFormat: opts.timeFormat,
	}
}

func renderSearchResults(out io.Writer, entries []files.IsMetadata, opts listOptions) error {
	w := new(tabwriter.Writer)
	w.Init(out, 4, 8, 1, ' ', 0)

	if opts.long {
		_, _ = fmt.Fprint(w, "Revision\tSize\tLast modified\tPath\n")
	}

	for _, entry := range entries {
		switch f := entry.(type) {
		case *files.FileMetadata:
			_, _ = fmt.Fprint(w, formatFileMetadataWithOpts(f, opts))
			_, _ = fmt.Fprintln(w)
		case *files.FolderMetadata:
			printFolderMetadata(w, f, opts.long)
			_, _ = fmt.Fprintln(w)
		}
	}

	return w.Flush()
}

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [flags] <query> [path-scope]",
	Short: "Search",
	RunE:  search,
}

func init() {
	RootCmd.AddCommand(searchCmd)
	searchCmd.Flags().BoolP("long", "l", false, "Long listing")
	searchCmd.Flags().String("sort", "", "Sort by: name, size, time, type")
	searchCmd.Flags().BoolP("reverse", "r", false, "Reverse sort order")
	searchCmd.Flags().String("time", "server", "Time field: server, client")
	searchCmd.Flags().String("time-format", "", "Time format: short (2006-01-02 15:04), rfc3339")
	enableStructuredOutput(searchCmd)
}
