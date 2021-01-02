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
	"os"
	"text/tabwriter"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

const deletedItemFormatString = "<<%s>>"

const colorReset = "\033[0m"
const colorRed = "\033[31m"
const colorGreen = "\033[32m"
const colorYellow = "\033[33m"
const colorBlue = "\033[34m"
const colorPurple = "\033[35m"
const colorCyan = "\033[36m"
const colorWhite = "\033[37m"

// Sends a get_metadata request for a given path and returns the response
func getFileMetadata(c files.Client, path string) (files.IsMetadata, error) {
	arg := files.NewGetMetadataArg(path)

	arg.IncludeDeleted = true

	res, err := c.GetMetadata(arg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func formatFolderMetadata(w io.Writer, e *files.FolderMetadata, longFormat bool) (string) {
	text := fmt.Sprintf("%s/\t", e.PathDisplay)
	if longFormat {
		text = fmt.Sprintf("-\t-\t-\t") + text
	}
	return text
}

func printFileMetadata(w io.Writer, e *files.FileMetadata, longFormat bool) {
	if longFormat {
		fmt.Fprintf(w, "%s\t%s\t%s\t", e.Rev, humanize.IBytes(e.Size), humanize.Time(e.ServerModified))
	}
	fmt.Fprintf(w, "%s\t", e.PathDisplay)
}

func printDeletedMetadata(w io.Writer, e *files.DeletedMetadata, longFormat bool) {
	if longFormat {
		fmt.Fprintf(w, "-\t-\t-\t")
	}
	fmt.Fprintf(w, "%s\t", e.PathDisplay)
}

func typeof(v interface{}) string {
	return fmt.Sprintf("%T", v)
}

func ls(cmd *cobra.Command, args []string) (err error) {

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 8, 1, ' ', 0)

	logOut := new(tabwriter.Writer)
	logOut.Init(os.Stderr, 4, 8, 1, ' ', 0)
	debug := func(message string) {
		fmt.Fprint(w, string(colorCyan)+"DEBUG: "+message+string(colorReset)+"\n")
	}
	error := func(message string) {
		fmt.Fprint(w, string(colorRed)+"DEBUG: "+message+string(colorReset)+"\n")
	}

	print := func(message string) {
		fmt.Fprint(w, message)
	}

	path := ""
	if len(args) > 0 {
		if path, err = validatePath(args[0]); err != nil {
			return err
		}
	}
	dbx := files.New(config)

	arg := files.NewListFolderArg(path)
	arg.Recursive, _ = cmd.Flags().GetBool("recurse")
	arg.IncludeDeleted, _ = cmd.Flags().GetBool("includeDeleted")
	onlyDeleted, _ := cmd.Flags().GetBool("onlyDeleted")
	arg.IncludeDeleted = arg.IncludeDeleted || onlyDeleted

	res, err := dbx.ListFolder(arg)

	var entries []files.IsMetadata
	if err != nil {
		switch e := err.(type) {
		case files.ListFolderAPIError:
			// Don't treat a "not_folder" error as fatal; recover by sending a
			// get_metadata request for the same path and using that response instead.
			if e.EndpointError.Path.Tag == files.LookupErrorNotFolder {
				var metaRes files.IsMetadata
				metaRes, err = getFileMetadata(dbx, path)
				entries = []files.IsMetadata{metaRes}
			} else {
				return err
			}
		default:
			return err
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

	// debug(fmt.Sprintf("entries.Count = %v", len(entries)))

	// var items = make(map[string]files.IsMetadata)
	// // See ListFolder documentation at https://godoc.org/github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files#Client
	// for _, entry := range entries {
	// 	switch f := entry.(type) {
	// 	case *files.FileMetadata:
	// 		if !onlyDeleted {
	// 			items[f.PathLower] = entry
	// 		}
	// 	case *files.FolderMetadata:
	// 		if !onlyDeleted {
	// 			items[f.PathLower] = entry
	// 		}
	// 	case *files.DeletedMetadata:
	// 		if !onlyDeleted {
	// 			f.PathDisplay = fmt.Sprintf(deletedItemFormatString, f.PathDisplay)
	// 			items[f.PathLower] = entry
	// 		}
	// 	default:
	// 		error("Converting Array to Map on item was type other than FileMetadata, FolderMetadata, or DeletedMetadata")
	// 		// Ignore
	// 	}
	// }

	// debug(fmt.Sprintf("items.Count = %v", len(items)))

	long, _ := cmd.Flags().GetBool("long")

	itemCounter := 0
	newLineCounter := 0
	if long {
		fmt.Fprint(w, "Revision\tSize\tLast modified\tPath\n")
	}

	for _, entry := range entries {
		switch f := entry.(type) {
		case *files.FileMetadata:
			if !onlyDeleted {
				printFileMetadata(w, f, long)
				itemCounter = itemCounter + 1
			}
		case *files.FolderMetadata:
			if !onlyDeleted {
				printFolderMetadata(w, f, long)
				itemCounter = itemCounter + 1
			}
		case *files.DeletedMetadata:
			revisionArg := files.NewListRevisionsArg(f.PathLower)
			res, err := dbx.ListRevisions(revisionArg)
			if err != nil {
				listRevisionError, ok := err.(files.ListRevisionsAPIError)
				if ok {
					// We have a ListRevisionsAPIERror
					if listRevisionError.EndpointError.Path.Tag == files.LookupErrorNotFile {
						// Don't treat a "not_file" error as fatal; recover by sending a
						// get_metadata request for the same path and using that response instead.
						var metaRes files.IsMetadata
						metaRes, err = getFileMetadata(dbx, f.PathLower)
						if err != nil {
							error("getFileMetadata failed with error:" + string(err.Error()))
							break
						}
						debug("metaRes.(Type) = " + typeof(metaRes))
						switch x := entry.(type) {
						case *files.FileMetadata:
							if !onlyDeleted {
								debug("Retrieved revision for deleted File: " + x.PathLower)
								printFileMetadata(w, x, long)
								itemCounter = itemCounter + 1
							}
						case *files.FolderMetadata:
							if !onlyDeleted {
								debug("Retrieved revision for deleted Folder: " + x.PathLower)
								printFolderMetadata(w, x, long)
								itemCounter = itemCounter + 1
							}
						case *files.DeletedMetadata:
							debug("Retrieved revision for deleted DeletedMetadata: " + x.PathLower)
							f.PathDisplay = fmt.Sprintf(deletedItemFormatString, x.PathDisplay)
							printDeletedMetadata(w, x, long)
							itemCounter = itemCounter + 1
						default:
							error("Unexpected Type for mesRest = " + typeof(metaRes))
						}
					} else {
						debug("listRevisionError.EndpointError.Path.Tag = " + listRevisionError.EndpointError.Path.Tag)
					}
				} else {
					error("Unexpected Error type calling ListRevisions: " + err.Error())
					break
				}
			} else if len(res.Entries) == 0 {
				// Occasionally revisions will be returned with an empty Entries list.
				f.PathDisplay = fmt.Sprintf(deletedItemFormatString, f.PathDisplay)
				printDeletedMetadata(w, f, long)
				itemCounter = itemCounter + 1
			} else {
				res.Entries[0].PathDisplay = fmt.Sprintf(deletedItemFormatString, res.Entries[0].PathDisplay)
				printFileMetadata(w, res.Entries[0], long)
				itemCounter = itemCounter + 1
			}
		default:
			debug("Item of unknown type (not FileMetadata, FolderMetadata, or DeletedMetadata) when iterating over all items")
			// Ignore
		}
		// debug(fmt.Sprintf("itemCounter=%v; newLineCounter=%v", itemCounter, newLineCounter))
		if (itemCounter%4 == 0) || (long && (itemCounter > newLineCounter)) {
			fmt.Fprintln(w)
			newLineCounter = newLineCounter + 1
		}
	}

	err = w.Flush()
	if err != nil {
		error("w.Flush():" + string(err.Error()))
	}
	err = logOut.Flush()
	if err != nil {
		error("logOut.Flush():" + string(err.Error()))
	}
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
	lsCmd.Flags().BoolP("recurse", "R", false, "Recursively list all subfolders")
	lsCmd.Flags().BoolP("includeDeleted", "d", false, "Include deleted files")
	lsCmd.Flags().BoolP("onlyDeleted", "D", false, "Only show deleted files")
}
