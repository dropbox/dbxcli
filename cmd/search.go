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
	"strings"
	"text/tabwriter"

	"github.com/dropbox/dbxcli/v3/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

type searchInput struct {
	Query      string `json:"query"`
	Path       string `json:"path,omitempty"`
	Content    bool   `json:"content"`
	Limit      uint64 `json:"limit,omitempty"`
	OrderBy    string `json:"order_by,omitempty"`
	Long       bool   `json:"long"`
	Sort       string `json:"sort,omitempty"`
	Reverse    bool   `json:"reverse"`
	Time       string `json:"time,omitempty"`
	TimeFormat string `json:"time_format,omitempty"`
}

const searchJSONStatusFound = "found"

type searchCommandOptions struct {
	list    listOptions
	content bool
	limit   uint64
	orderBy string
}

func search(cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 {
		return invalidArgumentsErrorWithDetails("`search` requires a `query` argument", argumentErrorDetails("query"))
	}
	if len(args) > 2 {
		return invalidArgumentsErrorWithDetails("`search` accepts at most one optional `path-scope` argument", mergeJSONErrorDetails(argumentErrorDetails("path-scope"), pathErrorDetails(args[2])))
	}

	var scope string
	if len(args) == 2 {
		scope = args[1]
		if !strings.HasPrefix(scope, "/") {
			return invalidArgumentsErrorWithDetails("`search` `path-scope` must begin with \"/\"", mergeJSONErrorDetails(argumentErrorDetails("path-scope"), pathErrorDetails(scope)))
		}
	}

	opts, err := parseSearchOptions(cmd)
	if err != nil {
		return err
	}
	arg := newSearchV2Arg(args[0], scope, opts)

	dbx := filesNewFunc(config)
	res, err := dbx.SearchV2(arg)
	if err != nil {
		return err
	}

	var entries []files.IsMetadata
	entries = appendSearchMatches(entries, res.Matches, opts.limit)

	for res.HasMore && !searchLimitReached(entries, opts.limit) {
		contArg := files.NewSearchV2ContinueArg(res.Cursor)
		res, err = dbx.SearchContinueV2(contArg)
		if err != nil {
			return err
		}
		entries = appendSearchMatches(entries, res.Matches, opts.limit)
	}

	sortEntries(entries, opts.list)

	return renderSearchOutput(cmd, args[0], scope, entries, opts)
}

func parseSearchOptions(cmd *cobra.Command) (searchCommandOptions, error) {
	content, _ := cmd.Flags().GetBool("content")
	limit, _ := cmd.Flags().GetUint64("limit")
	orderBy, _ := cmd.Flags().GetString("order-by")
	listOpts, err := parseListOptions(cmd)
	if err != nil {
		return searchCommandOptions{}, err
	}
	if !validSearchOrderBy(orderBy) {
		return searchCommandOptions{}, invalidArgumentsErrorWithDetails("`search --order-by` must be one of: relevance, modified", flagErrorDetails("order-by"))
	}

	return searchCommandOptions{
		list:    listOpts,
		content: content,
		limit:   limit,
		orderBy: orderBy,
	}, nil
}

func newSearchV2Arg(query, scope string, opts searchCommandOptions) *files.SearchV2Arg {
	arg := files.NewSearchV2Arg(query)
	searchOpts := files.NewSearchOptions()
	searchOpts.Path = scope
	searchOpts.FilenameOnly = !opts.content
	if opts.limit > 0 && opts.limit < searchOpts.MaxResults {
		searchOpts.MaxResults = opts.limit
	}
	if opts.orderBy != "" {
		searchOpts.OrderBy = searchOrderBy(opts.orderBy)
	}
	arg.Options = searchOpts
	return arg
}

func validSearchOrderBy(orderBy string) bool {
	switch orderBy {
	case "", "relevance", "modified":
		return true
	default:
		return false
	}
}

func searchOrderBy(orderBy string) *files.SearchOrderBy {
	switch orderBy {
	case "modified":
		return &files.SearchOrderBy{Tagged: dropbox.Tagged{Tag: files.SearchOrderByLastModifiedTime}}
	default:
		return &files.SearchOrderBy{Tagged: dropbox.Tagged{Tag: files.SearchOrderByRelevance}}
	}
}

func appendSearchMatches(entries []files.IsMetadata, matches []*files.SearchMatchV2, limit uint64) []files.IsMetadata {
	for _, m := range matches {
		if searchLimitReached(entries, limit) {
			break
		}
		if m.Metadata != nil && m.Metadata.Metadata != nil {
			entries = append(entries, m.Metadata.Metadata)
		}
	}
	return entries
}

func searchLimitReached(entries []files.IsMetadata, limit uint64) bool {
	return limit > 0 && uint64(len(entries)) >= limit
}

func renderSearchOutput(cmd *cobra.Command, query, scope string, entries []files.IsMetadata, opts searchCommandOptions) error {
	out := commandOutput(cmd)
	if commandOutputFormat(cmd) != output.FormatJSON {
		return out.RenderText(func(w io.Writer) error {
			return renderSearchResults(w, entries, opts.list)
		})
	}

	input := newSearchInput(query, scope, opts)
	metadata, err := jsonMetadataListFromDropbox(entries)
	if err != nil {
		return err
	}
	results := newJSONMetadataOperationResults(searchJSONStatusFound, metadata)
	return renderJSONOperationOutput(cmd, input, results)
}

func newSearchInput(query, scope string, opts searchCommandOptions) searchInput {
	return searchInput{
		Query:      query,
		Path:       scope,
		Content:    opts.content,
		Limit:      opts.limit,
		OrderBy:    opts.orderBy,
		Long:       opts.list.long,
		Sort:       opts.list.sortBy,
		Reverse:    opts.list.reverse,
		Time:       opts.list.timeField,
		TimeFormat: opts.list.timeFormat,
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
	searchCmd.Flags().BoolP("content", "c", false, "Search file contents in addition to filenames")
	searchCmd.Flags().Uint64("limit", 0, "Maximum number of matches to return")
	searchCmd.Flags().String("order-by", "", "Server-side search ordering: relevance, modified")
	searchCmd.Flags().BoolP("long", "l", false, "Long listing")
	searchCmd.Flags().String("sort", "", "Sort by: name, size, time, type")
	searchCmd.Flags().BoolP("reverse", "r", false, "Reverse sort order")
	searchCmd.Flags().String("time", "server", "Time field: server, client")
	searchCmd.Flags().String("time-format", "", "Time format: short (2006-01-02 15:04), rfc3339")
	enableStructuredOutput(searchCmd)
}
