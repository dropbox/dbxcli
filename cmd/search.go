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
	"os"
	"strings"
	"text/tabwriter"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

func search(cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 {
		return errors.New("`search` requires a `query` argument")
	}

	// Parse path scope, if provided.
	var scope string
	if len(args) == 2 {
		scope = args[1]
		if !strings.HasPrefix(scope, "/") {
			return errors.New("`search` `path-scope` must begin with \"/\"")
		}
	}

	arg := files.NewSearchArg(scope, args[0])

	dbx := files.New(config)
	res, err := dbx.Search(arg)
	if err != nil {
		return
	}

	long, _ := cmd.Flags().GetBool("long")

	return renderSearchResults(os.Stdout, res, long)
}

func renderSearchResults(out io.Writer, res *files.SearchResult, long bool) error {
	w := new(tabwriter.Writer)
	w.Init(out, 4, 8, 1, ' ', 0)

	if long {
		_, _ = fmt.Fprint(w, "Revision\tSize\tLast modified\tPath\n")
	}

	for _, m := range res.Matches {
		switch f := m.Metadata.(type) {
		case *files.FileMetadata:
			printFileMetadata(w, f, long)
			_, _ = fmt.Fprintln(w)
		case *files.FolderMetadata:
			printFolderMetadata(w, f, long)
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
}
