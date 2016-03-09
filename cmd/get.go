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
	"path"

	"github.com/dropbox/dropbox-sdk-go-unofficial/files"
	"github.com/mitchellh/ioprogress"
	"github.com/spf13/cobra"
)

func get(cmd *cobra.Command, args []string) (err error) {
	src, err := validatePath(args[0])
	dst := args[1]

	if dst == "" {
		dst = path.Base(src)
	}

	arg := files.NewDownloadArg()
	arg.Path = src

	res, contents, err := dbx.Download(arg)
	defer contents.Close()
	if err != nil {
		return
	}

	f, err := os.Create(dst)
	defer f.Close()
	if err != nil {
		return
	}

	progressbar := &ioprogress.Reader{
		Reader: contents,
		DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
			return fmt.Sprintf("Downloading %s/%s",
				humanizeSize(uint64(progress)),
				humanizeSize(uint64(total)))
		}),
		Size: int64(res.Size),
	}

	if _, err = io.Copy(f, progressbar); err != nil {
		return
	}

	return
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Download a file",
	RunE:  get,
}

func init() {
	RootCmd.AddCommand(getCmd)
}
