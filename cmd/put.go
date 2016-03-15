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

	"github.com/dropbox/dropbox-sdk-go-unofficial/files"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/ioprogress"
	"github.com/spf13/cobra"
)

const chunkSize int64 = 1 << 24

func uploadChunked(r io.Reader, commitInfo *files.CommitInfo, sizeTotal int64) (err error) {
	res, err := dbx.UploadSessionStart(&io.LimitedReader{R: r, N: chunkSize})
	if err != nil {
		return
	}

	written := chunkSize

	for (sizeTotal - written) > chunkSize {
		args := files.NewUploadSessionCursor(res.SessionId, uint64(written))

		err = dbx.UploadSessionAppend(args, &io.LimitedReader{R: r, N: chunkSize})
		if err != nil {
			return
		}
		written += chunkSize
	}

	cursor := files.NewUploadSessionCursor(res.SessionId, uint64(written))
	args := files.NewUploadSessionFinishArg(cursor, commitInfo)

	if _, err = dbx.UploadSessionFinish(args, r); err != nil {
		return
	}

	return
}

func put(cmd *cobra.Command, args []string) (err error) {
	src := args[0]
	dst, err := validatePath(args[1])
	if err != nil {
		return
	}

	contents, err := os.Open(src)
	defer contents.Close()
	if err != nil {
		return
	}

	contentsInfo, err := contents.Stat()
	if err != nil {
		return
	}

	progressbar := &ioprogress.Reader{
		Reader: contents,
		DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
			return fmt.Sprintf("Uploading %s/%s",
				humanize.Bytes(uint64(progress)), humanize.Bytes(uint64(total)))
		}),
		Size: contentsInfo.Size(),
	}

	commitInfo := files.NewCommitInfo(dst)
	commitInfo.Mode.Tag = "overwrite"

	if contentsInfo.Size() > chunkSize {
		return uploadChunked(progressbar, commitInfo, contentsInfo.Size())
	}

	if _, err = dbx.Upload(commitInfo, progressbar); err != nil {
		return
	}

	return
}

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put",
	Short: "Upload files",
	RunE:  put,
}

func init() {
	RootCmd.AddCommand(putCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// putCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// putCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
