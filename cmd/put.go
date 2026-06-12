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
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/ioprogress"
	"github.com/spf13/cobra"
)

const singleShotUploadSizeCutoff int64 = 32 * (1 << 20)

var filesNewFunc = func(cfg dropbox.Config) files.Client {
	return files.New(cfg)
}

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

type putOptions struct {
	chunkSize int64
	workers   int
	debug     bool
}

func put(cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 || len(args) > 2 {
		return errors.New("`put` requires `src` and/or `dst` arguments")
	}

	opts, err := parsePutOptions(cmd)
	if err != nil {
		return err
	}

	recursive, _ := cmd.Flags().GetBool("recursive")

	src := args[0]

	srcInfo, err := os.Stat(src)
	if err != nil {
		return
	}

	if srcInfo.IsDir() && !recursive {
		return fmt.Errorf("%s is a directory (use --recursive to upload directories)", src)
	}

	// Default `dst` to the base segment of the source path; use the second argument if provided.
	dst := "/" + filepath.Base(src)
	dstIsDir := false
	if len(args) == 2 {
		dstIsDir = strings.HasSuffix(args[1], "/")
		dst, err = validatePath(args[1])
		if err != nil {
			return
		}
	}

	if !srcInfo.IsDir() {
		dst = resolveDestination(filesNewFunc(config), src, dst, dstIsDir)
	}

	if srcInfo.IsDir() {
		return putRecursive(src, dst, opts)
	}

	return putFile(src, dst, opts)
}

func resolveDestination(dbx files.Client, src, dst string, dstIsDir bool) string {
	if dstIsDir {
		return path.Join("/", dst, filepath.Base(src))
	}
	meta, err := dbx.GetMetadata(files.NewGetMetadataArg(dst))
	if err != nil {
		return dst
	}
	if _, ok := meta.(*files.FolderMetadata); ok {
		return path.Join("/", dst, filepath.Base(src))
	}
	return dst
}

func parsePutOptions(cmd *cobra.Command) (putOptions, error) {
	chunkSize, err := cmd.Flags().GetInt64("chunksize")
	if err != nil {
		return putOptions{}, err
	}
	if chunkSize%(1<<22) != 0 {
		return putOptions{}, errors.New("`put` requires chunk size to be multiple of 4MiB")
	}
	workers, err := cmd.Flags().GetInt("workers")
	if err != nil {
		return putOptions{}, err
	}
	if workers < 1 {
		workers = 1
	}
	debug, _ := cmd.Flags().GetBool("debug")
	return putOptions{chunkSize: chunkSize, workers: workers, debug: debug}, nil
}

func putFile(src, dst string, opts putOptions) error {
	contents, err := os.Open(src)
	if err != nil {
		return err
	}
	defer contents.Close()

	contentsInfo, err := contents.Stat()
	if err != nil {
		return err
	}

	commitInfo := files.NewCommitInfo(dst)
	commitInfo.Mode.Tag = "overwrite"

	// The Dropbox API only accepts timestamps in UTC with second precision.
	ts := time.Now().UTC().Round(time.Second)
	commitInfo.ClientModified = &ts

	dbx := filesNewFunc(config)
	if contentsInfo.Size() > singleShotUploadSizeCutoff {
		return uploadChunked(dbx, uploadProgressReader(contents, contentsInfo.Size()), commitInfo, contentsInfo.Size(), opts.workers, opts.chunkSize, opts.debug)
	}

	uploadArg := &files.UploadArg{CommitInfo: *commitInfo}
	return uploadSingleShot(dbx, contents, uploadArg, contentsInfo.Size())
}

func putRecursive(src, dst string, opts putOptions) error {
	src = filepath.Clean(src)
	var uploadErrors []error
	dirsWithFiles := make(map[string]bool)

	err := filepath.WalkDir(src, func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}

		relPath, err := filepath.Rel(src, filePath)
		if err != nil {
			return err
		}

		dirsWithFiles[filepath.Dir(filePath)] = true

		remotePath := path.Join(dst, filepath.ToSlash(relPath))
		fmt.Fprintf(os.Stderr, "Uploading %s -> %s\n", filePath, remotePath)

		if err := putFile(filePath, remotePath, opts); err != nil {
			uploadErrors = append(uploadErrors, fmt.Errorf("%s: %w", filePath, err))
		}
		return nil
	})
	if err != nil {
		return err
	}

	dbx := filesNewFunc(config)

	fmt.Fprintf(os.Stderr, "Creating directory %s\n", dst)
	arg := files.NewCreateFolderArg(dst)
	if _, mkdirErr := dbx.CreateFolderV2(arg); mkdirErr != nil {
		if !isConflictError(mkdirErr) {
			uploadErrors = append(uploadErrors, fmt.Errorf("mkdir %s: %w", dst, mkdirErr))
		}
	}

	err = filepath.WalkDir(src, func(dirPath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if dirsWithFiles[dirPath] {
			return nil
		}

		relPath, err := filepath.Rel(src, dirPath)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		remotePath := path.Join(dst, filepath.ToSlash(relPath))
		fmt.Fprintf(os.Stderr, "Creating directory %s\n", remotePath)
		arg := files.NewCreateFolderArg(remotePath)
		if _, mkdirErr := dbx.CreateFolderV2(arg); mkdirErr != nil {
			if !isConflictError(mkdirErr) {
				uploadErrors = append(uploadErrors, fmt.Errorf("mkdir %s: %w", remotePath, mkdirErr))
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(uploadErrors) > 0 {
		return fmt.Errorf("failed to upload %d file(s): %v", len(uploadErrors), uploadErrors[0])
	}
	return nil
}

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put [flags] <source> [<target>]",
	Short: "Upload files or directories",
	Long: `Upload files or directories to Dropbox.
  - If target is not provided, uploads to the root of your Dropbox.
  - Use --recursive (-r) to upload entire directories.
`,
	RunE: put,
}

func init() {
	RootCmd.AddCommand(putCmd)
	putCmd.Flags().BoolP("recursive", "r", false, "Recursively upload directories")
	putCmd.Flags().IntP("workers", "w", 4, "Number of concurrent upload workers to use")
	putCmd.Flags().Int64P("chunksize", "c", 1<<24, "Chunk size to use (should be multiple of 4MiB)")
	putCmd.Flags().BoolP("debug", "d", false, "Print debug timing")
}
