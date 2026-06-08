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
	"log"
	"os"
	"path"
	"sync"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/ioprogress"
	"github.com/spf13/cobra"
)

const singleShotUploadSizeCutoff int64 = 32 * (1 << 20)

type uploadChunk struct {
	data   []byte
	offset uint64
	close  bool
}

func uploadOneChunk(dbx files.Client, args *files.UploadSessionAppendArg, data []byte) error {
	return retryWithBackoff(func() error {
		err := dbx.UploadSessionAppendV2(args, bytes.NewReader(data))
		if uploadChunkAlreadyAccepted(err, args.Cursor.Offset+uint64(len(data))) {
			return nil
		}
		return err
	})
}

func uploadChunkAlreadyAccepted(err error, expectedOffset uint64) bool {
	var appendErr files.UploadSessionAppendV2APIError
	if !errors.As(err, &appendErr) || appendErr.EndpointError == nil {
		return false
	}
	endpointErr := appendErr.EndpointError
	return endpointErr.Tag == files.UploadSessionAppendErrorIncorrectOffset &&
		endpointErr.IncorrectOffset != nil &&
		endpointErr.IncorrectOffset.CorrectOffset == expectedOffset
}

func uploadProgressReader(r io.Reader, size int64) *ioprogress.Reader {
	return &ioprogress.Reader{
		Reader: r,
		DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
			return fmt.Sprintf("Uploading %s/%s",
				humanize.IBytes(uint64(progress)), humanize.IBytes(uint64(total)))
		}),
		Size: size,
	}
}

func uploadSingleShot(dbx files.Client, r io.ReadSeeker, uploadArg *files.UploadArg, size int64) error {
	return retryWithBackoff(func() error {
		if _, err := r.Seek(0, io.SeekStart); err != nil {
			return err
		}
		_, err := dbx.Upload(uploadArg, uploadProgressReader(r, size))
		return err
	})
}

func uploadChunked(dbx files.Client, r io.Reader, commitInfo *files.CommitInfo, sizeTotal int64, workers int, chunkSize int64, debug bool) (err error) {
	t0 := time.Now()
	startArgs := files.NewUploadSessionStartArg()
	startArgs.SessionType = &files.UploadSessionType{}
	startArgs.SessionType.Tag = files.UploadSessionTypeConcurrent
	var res *files.UploadSessionStartResult
	err = retryWithBackoff(func() error {
		var e error
		res, e = dbx.UploadSessionStart(startArgs, nil)
		return e
	})
	if err != nil {
		return
	}
	if debug {
		log.Printf("Start took: %v\n", time.Since(t0))
	}

	t1 := time.Now()
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

				t0 := time.Now()
				if err := uploadOneChunk(dbx, args, chunk.data); err != nil {
					errCh <- err
				}
				if debug {
					log.Printf("Chunk upload at offset %d took: %v\n", chunk.offset, time.Since(t0))
				}
			}
		}()
	}

	written := int64(0)
	for written < sizeTotal {
		data, err := io.ReadAll(&io.LimitedReader{R: r, N: chunkSize})
		if err != nil {
			return err
		}
		expectedLen := chunkSize
		if written+chunkSize > sizeTotal {
			expectedLen = sizeTotal - written
		}
		if len(data) != int(expectedLen) {
			return fmt.Errorf("failed to read %d bytes from source", expectedLen)
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
	if debug {
		log.Printf("Full upload took: %v\n", time.Since(t1))
	}

	t2 := time.Now()
	cursor := files.NewUploadSessionCursor(res.SessionId, uint64(written))
	finishArgs := files.NewUploadSessionFinishArg(cursor, commitInfo)
	err = retryWithBackoff(func() error {
		_, e := dbx.UploadSessionFinish(finishArgs, nil)
		return e
	})
	if debug {
		log.Printf("Finish took: %v\n", time.Since(t2))
	}
	return
}

func put(cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 || len(args) > 2 {
		return errors.New("`put` requires `src` and/or `dst` arguments")
	}

	chunkSize, err := cmd.Flags().GetInt64("chunksize")
	if err != nil {
		return err
	}
	if chunkSize%(1<<22) != 0 {
		return errors.New("`put` requires chunk size to be multiple of 4MiB")
	}
	workers, err := cmd.Flags().GetInt("workers")
	if err != nil {
		return err
	}
	if workers < 1 {
		workers = 1
	}
	debug, _ := cmd.Flags().GetBool("debug")

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

	commitInfo := files.NewCommitInfo(dst)
	commitInfo.Mode.Tag = "overwrite"

	// The Dropbox API only accepts timestamps in UTC with second precision.
	ts := time.Now().UTC().Round(time.Second)
	commitInfo.ClientModified = &ts

	dbx := files.New(config)
	if contentsInfo.Size() > singleShotUploadSizeCutoff {
		return uploadChunked(dbx, uploadProgressReader(contents, contentsInfo.Size()), commitInfo, contentsInfo.Size(), workers, chunkSize, debug)
	}

	uploadArg := &files.UploadArg{CommitInfo: *commitInfo}
	return uploadSingleShot(dbx, contents, uploadArg, contentsInfo.Size())
}

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put [flags] <source> [<target>]",
	Short: "Upload a single file",
	Long: `Upload a single file
	- If target is not provided puts the file in the root of your Dropbox directory.
	- If target is provided it must be the desired filename in the cloud (and not a directory).
	`,

	RunE: put,
}

func init() {
	RootCmd.AddCommand(putCmd)
	putCmd.Flags().IntP("workers", "w", 4, "Number of concurrent upload workers to use")
	putCmd.Flags().Int64P("chunksize", "c", 1<<24, "Chunk size to use (should be multiple of 4MiB)")
	putCmd.Flags().BoolP("debug", "d", false, "Print debug timing")
}
