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
	"math"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/dropbox/dropbox-sdk-go-unofficial/files"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

// Sends a get_metadata request for a given path and returns the response
func getFileMetadata(path string) (*files.Metadata, error) {
	arg := files.NewGetMetadataArg(path)

	res, err := dbx.GetMetadata(arg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func printFolderMetadata(w io.Writer, e *files.FolderMetadata, longFormat bool) {
	if longFormat {
		fmt.Fprintf(w, "-\t-\t-\t")
	}
	fmt.Fprintf(w, "%s\n", e.Name)
}

func printFileMetadata(w io.Writer, e *files.FileMetadata, longFormat bool) {
	if longFormat {
		fmt.Fprintf(w, "%s\t%s\t%s\t", e.Rev, humanize.IBytes(e.Size), humanize.Time(e.ServerModified))
	}
	fmt.Fprintf(w, "%s\n", e.Name)
}

func ls(cmd *cobra.Command, args []string) (err error) {
	path := ""
	if len(args) > 0 {
		if path, err = validatePath(args[0]); err != nil {
			return err
		}
	}

	arg := files.NewListFolderArg(path)

	res, err := dbx.ListFolder(arg)
	var entries []*files.Metadata
	if err != nil {
		// Don't treat a "not_folder" error as fatal; recover by sending a
		// get_metadata request for the same path and using that response instead.
		if strings.Contains(err.Error(), "path/not_folder/") {
			var metaRes *files.Metadata
			metaRes, err = getFileMetadata(path)
			entries = []*files.Metadata{metaRes}
		}

		// Return if there's an error other than "not_folder" or if the follow-up
		// metadata request fails.
		if err != nil {
			return err
		}
	} else {
		entries = res.Entries

		for res.HasMore {
			arg := files.NewListFolderContinueArg(res.Cursor)

			res, err = dbx.ListFolderContinue(arg)
			if err != nil {
				return err
			}

			entries = append(entries, res.Entries...)
		}
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 8, 1, ' ', 0)
	long, _ := cmd.Flags().GetBool("long")
	if long {
		fmt.Fprintf(w, "Revision\tSize\tLast modified\tPath\n")
	}

	entryNames := listOfEntryNames(entries)
	lengths := lengthOfEntryNames(entryNames)
	columnLength := columnLength(lengths)
	entryNames = resizeEntries(columnLength, entryNames)
	termWidth, err := terminalWidth()
	if err != nil {
		return err
	}
	columns := termWidth / columnLength
	if columns == 0 {
		columns++
	}
	for i, entry := range entryNames {
		if i%columns == 0 {
			if i != 0 {
				fmt.Fprintf(w, "\n")
			}
		}
		fmt.Fprintf(w, "%s", entry)
	}
	fmt.Fprintf(w, "%s", "\n")
	w.Flush()
	return err
}

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls [flags] [<path>]",
	Short: "List files and folders",
	Example: `  dbxcli ls / # Or just 'ls'
  dbxcli ls /some-folder # Or 'ls some-folder'
  dbxcli ls /some-folder/some-file.pdf
  dbxcli ls -l`,
	RunE: ls,
}

func init() {
	RootCmd.AddCommand(lsCmd)

	lsCmd.Flags().BoolP("long", "l", false, "Long listing")
}

func listOfEntryNames(entries []*files.Metadata) []string {
	listOfEntryNames := []string{}

	for _, entry := range entries {
		switch entry.Tag {
		case "folder":
			listOfEntryNames = append(listOfEntryNames, entry.Folder.Name+"    ")
		case "file":
			listOfEntryNames = append(listOfEntryNames, entry.File.Name+"    ")
		}
	}

	sort.Strings(listOfEntryNames)
	return listOfEntryNames
}

func lengthOfEntryNames(listOfEntryNames []string) []int {
	lengths := []int{}
	for _, entry := range listOfEntryNames {
		lengths = append(lengths, len(entry))
	}
	return lengths
}

func fewestColumns(lengths []int) int {
	termWidth, _ := terminalWidth()

	columns := 1
	prevLength := 0
	for _, length := range lengths {
		if length+prevLength+4 < termWidth {
			columns++
			prevLength += length + 4
		} else {
			break
		}
	}
	return columns
}

func columnLength(lengths []int) int {
	sort.Ints(lengths)
	return reverse(lengths)[0] + 4
}

func resizeEntries(fullSize int, entryNames []string) []string {
	for i, entry := range entryNames {
		for len(entry) != fullSize {
			entry += " "
		}
		entryNames[i] = entry
	}
	return entryNames
}

func numberOfRows(numberOfRows, numberOfNames int) int {
	return int(math.Ceil(float64(numberOfNames) / float64(numberOfRows)))
}

func reverse(input []int) []int {
	if len(input) == 0 {
		return input
	}
	return append(reverse(input[1:]), input[0])
}

func terminalWidth() (int, error) {
	width, _, err := terminal.GetSize(0)
	return width, err
}
