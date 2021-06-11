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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/auth"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/ioprogress"
	"github.com/spf13/cobra"
)

const chunkSize int64 = 1 << 24

type uploadChunk struct {
	data   []byte
	offset uint64
	close  bool
}

func uploadOneChunk(dbx files.Client, args *files.UploadSessionAppendArg, data []byte) error {
	for {
		err := dbx.UploadSessionAppendV2(args, bytes.NewReader(data))
		if err != nil {
			return nil
		}

		rl, ok := err.(auth.RateLimitAPIError)
		if !ok {
			return err
		}

		time.Sleep(time.Second * time.Duration(rl.RateLimitError.RetryAfter))
	}
}

func uploadChunked(dbx files.Client, r io.Reader, commitInfo *files.CommitInfo, sizeTotal int64, workers int) (err error) {
	startArgs := files.NewUploadSessionStartArg()
	startArgs.SessionType = &files.UploadSessionType{}
	startArgs.SessionType.Tag = files.UploadSessionTypeConcurrent
	res, err := dbx.UploadSessionStart(startArgs, nil)
	if err != nil {
		return
	}

	wg := sync.WaitGroup{}
	workCh := make(chan uploadChunk, workers)
	errCh := make(chan error, 1)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chunk := range workCh {
				cursor := files.NewUploadSessionCursor(res.SessionId, chunk.offset)
				args := files.NewUploadSessionAppendArg(cursor)
				args.Close = chunk.close

				if err := uploadOneChunk(dbx, args, chunk.data); err != nil {
					errCh <- err
				}
			}
		}()
	}

	written := int64(0)
	for written < sizeTotal {
		data, err := ioutil.ReadAll(&io.LimitedReader{R: r, N: chunkSize})
		if err != nil {
			return err
		}

		chunk := uploadChunk{
			data:   data,
			offset: uint64(written),
			close:  written+chunkSize >= sizeTotal,
		}

		select {
		case workCh <- chunk:
		case err := <-errCh:
			return err
		}

		written += int64(len(data))
	}

	close(workCh)
	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
	}

	cursor := files.NewUploadSessionCursor(res.SessionId, uint64(written))
	args := files.NewUploadSessionFinishArg(cursor, commitInfo)
	_, err = dbx.UploadSessionFinish(args, nil)
	return
}

func put(cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 || len(args) > 2 {
		return errors.New("`put` requires `src` and/or `dst` arguments")
	}

	workers, err := cmd.Flags().GetInt("workers")
	if err != nil {
		return err
	}
	if workers < 1 {
		workers = 1
	}

	src := args[0]

	// Default `dst` to the base segment of the source path; use the second argument if provided.
	dst := "/" + path.Base(src)
	if len(args) == 2 {
		dst, err = validatePath(args[1])
		if err != nil {
			return
		}
	}

	contents, err := os.Open(src)
	if err != nil {
		return
	}
	defer contents.Close()

	contentsInfo, err := contents.Stat()
	if err != nil {
		return
	}

	progressbar := &ioprogress.Reader{
		Reader: contents,
		DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
			return fmt.Sprintf("Uploading %s/%s",
				humanize.IBytes(uint64(progress)), humanize.IBytes(uint64(total)))
		}),
		Size: contentsInfo.Size(),
	}

	commitInfo := files.NewCommitInfo(dst)
	commitInfo.Mode.Tag = "overwrite"

	// The Dropbox API only accepts timestamps in UTC with second precision.
	ts := time.Now().UTC().Round(time.Second)
	commitInfo.ClientModified = &ts

	dbx := files.New(config)
	if contentsInfo.Size() > chunkSize {
		return uploadChunked(dbx, progressbar, commitInfo, contentsInfo.Size(), workers)
	}

	if _, err = dbx.Upload(commitInfo, progressbar); err != nil {
		return
	}

	return
}

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put [flags] <source> [<target>]",
	Short: "Upload files",
	RunE:  put,
}

func init() {
	RootCmd.AddCommand(putCmd)
	putCmd.Flags().IntP("workers", "w", 4, "Number of concurrent upload workers to use")
}
