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

	"github.com/dropbox/dbxcli/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/ioprogress"
	"github.com/spf13/cobra"
)

const singleShotUploadSizeCutoff int64 = 32 * (1 << 20)

const (
	putIfExistsOverwrite = "overwrite"
	putIfExistsSkip      = "skip"
	putIfExistsFail      = "fail"
)

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

func uploadProgressReader(r io.Reader, size int64, errOut io.Writer) *ioprogress.Reader {
	if errOut == nil {
		errOut = os.Stderr
	}
	return &ioprogress.Reader{
		Reader: r,
		DrawFunc: ioprogress.DrawTerminalf(errOut, func(progress, total int64) string {
			return fmt.Sprintf("Uploading %s/%s",
				humanize.IBytes(uint64(progress)), humanize.IBytes(uint64(total)))
		}),
		Size: size,
	}
}

func uploadSingleShot(dbx files.Client, r io.ReadSeeker, uploadArg *files.UploadArg, size int64, errOut io.Writer) (*files.FileMetadata, error) {
	var metadata *files.FileMetadata
	err := retryWithBackoff(func() error {
		if _, err := r.Seek(0, io.SeekStart); err != nil {
			return err
		}
		var err error
		metadata, err = dbx.Upload(uploadArg, uploadProgressReader(r, size, errOut))
		return err
	})
	return metadata, err
}

func uploadChunked(dbx files.Client, r io.Reader, commitInfo *files.CommitInfo, sizeTotal int64, workers int, chunkSize int64, debug bool) (metadata *files.FileMetadata, err error) {
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
			return nil, err
		}
		expectedLen := chunkSize
		if written+chunkSize > sizeTotal {
			expectedLen = sizeTotal - written
		}
		if len(data) != int(expectedLen) {
			return nil, fmt.Errorf("failed to read %d bytes from source", expectedLen)
		}

		chunk := uploadChunk{
			data:   data,
			offset: uint64(written),
			close:  written+chunkSize >= sizeTotal,
		}

		select {
		case workCh <- chunk:
		case err := <-errCh:
			return nil, err
		}

		written += int64(len(data))
	}

	close(workCh)
	wg.Wait()
	select {
	case err := <-errCh:
		return nil, err
	default:
	}
	if debug {
		log.Printf("Full upload took: %v\n", time.Since(t1))
	}

	t2 := time.Now()
	cursor := files.NewUploadSessionCursor(res.SessionId, uint64(written))
	finishArgs := files.NewUploadSessionFinishArg(cursor, commitInfo)
	err = retryWithBackoff(func() error {
		var e error
		metadata, e = dbx.UploadSessionFinish(finishArgs, nil)
		return e
	})
	if debug {
		log.Printf("Finish took: %v\n", time.Since(t2))
	}
	return metadata, err
}

type putOptions struct {
	chunkSize int64
	workers   int
	debug     bool
	ifExists  string
	output    *output.Renderer
	errOut    io.Writer
}

const (
	putStatusUploaded = "uploaded"
	putStatusSkipped  = "skipped"
	putStatusCreated  = "created"
	putStatusExisting = "existing"

	putKindFile   = "file"
	putKindFolder = "folder"
)

type putCommandInput struct {
	Source    string `json:"source"`
	Target    string `json:"target"`
	Recursive bool   `json:"recursive"`
	IfExists  string `json:"if_exists"`
	Stdin     bool   `json:"stdin"`
}

type putResultInput struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type putResult struct {
	Status string         `json:"status"`
	Kind   string         `json:"kind"`
	Input  putResultInput `json:"input"`
	Result *jsonMetadata  `json:"result,omitempty"`
}

type putDestinationAction int

const (
	putDestinationUpload putDestinationAction = iota
	putDestinationSkip
)

func newPutResult(status, kind, source, target string, metadata files.IsMetadata) putResult {
	result := putResult{
		Status: status,
		Kind:   kind,
		Input: putResultInput{
			Source: source,
			Target: target,
		},
	}
	if metadata != nil {
		jsonResult := jsonMetadataFromDropbox(metadata)
		result.Result = &jsonResult
	}
	return result
}

func renderPutResults(cmd *cobra.Command, input putCommandInput, results []putResult) error {
	return renderJSONOperationOutput(cmd, input, putOperationResults(results))
}

func putOperationResults(results []putResult) []jsonOperationResult {
	operationResults := make([]jsonOperationResult, 0, len(results))
	for _, result := range results {
		var metadata any
		if result.Result != nil {
			metadata = result.Result
		}
		operationResults = append(operationResults, newJSONOperationResult(result.Status, result.Kind, result.Input, metadata))
	}
	return operationResults
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

	if src == "-" {
		return putStdin(cmd, args, opts, recursive)
	}

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
		if commandOutputFormat(cmd) == output.FormatText {
			return putRecursive(src, dst, opts)
		}
		results, err := putRecursiveWithResults(src, dst, opts)
		if err != nil {
			return err
		}
		return renderPutResults(cmd, putCommandInput{
			Source:    src,
			Target:    dst,
			Recursive: true,
			IfExists:  opts.ifExists,
			Stdin:     false,
		}, results)
	}

	result, err := putFileWithResult(src, dst, opts)
	if err != nil {
		return err
	}
	return renderPutResults(cmd, putCommandInput{
		Source:    src,
		Target:    dst,
		Recursive: false,
		IfExists:  opts.ifExists,
		Stdin:     false,
	}, []putResult{result})
}

func putStdin(cmd *cobra.Command, args []string, opts putOptions, recursive bool) error {
	if len(args) < 2 {
		return errors.New("`put -` requires an explicit target path")
	}
	if recursive {
		return errors.New("`put -` cannot be used with --recursive")
	}

	dst := args[1]
	if strings.HasSuffix(dst, "/") {
		return fmt.Errorf("cannot upload stdin to directory target %q; provide a full Dropbox file path", dst)
	}

	dstPath, err := validatePath(dst)
	if err != nil {
		return err
	}

	dbx := filesNewFunc(config)
	action, existingMetadata, err := checkPutStdinDestination(dbx, dstPath, opts.ifExists)
	if err != nil {
		return err
	}
	if action == putDestinationSkip {
		reportPutSkipped(opts, dstPath)
		result := newPutResult(putStatusSkipped, putKindFile, "-", dstPath, existingMetadata)
		return renderPutResults(cmd, putCommandInput{
			Source:    "-",
			Target:    dstPath,
			Recursive: false,
			IfExists:  opts.ifExists,
			Stdin:     true,
		}, []putResult{result})
	}

	tmpPath, _, cleanup, err := spoolStdinToTemp(cmd.InOrStdin())
	if err != nil {
		return err
	}

	result, uploadErr := putFileWithResult(tmpPath, dstPath, opts)
	cleanupErr := cleanup()

	if uploadErr != nil {
		if cleanupErr != nil {
			reportStdinCleanupFailure(opts, tmpPath, cleanupErr)
		}
		return uploadErr
	}

	if cleanupErr != nil {
		reportStdinCleanupFailure(opts, tmpPath, cleanupErr)
		return fmt.Errorf("failed to remove temp file %s after upload; sensitive stdin data may remain on disk: %w", tmpPath, cleanupErr)
	}

	result.Input.Source = "-"
	return renderPutResults(cmd, putCommandInput{
		Source:    "-",
		Target:    dstPath,
		Recursive: false,
		IfExists:  opts.ifExists,
		Stdin:     true,
	}, []putResult{result})
}

func reportStdinCleanupFailure(opts putOptions, tmpPath string, err error) {
	putOutput(opts).Status("error: failed to remove temp file %s: %v; sensitive stdin data may remain on disk", tmpPath, err)
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
	ifExists, err := parsePutIfExists(cmd)
	if err != nil {
		return putOptions{}, err
	}
	return putOptions{
		chunkSize: chunkSize,
		workers:   workers,
		debug:     debug,
		ifExists:  ifExists,
		output:    commandOutput(cmd),
		errOut:    cmd.ErrOrStderr(),
	}, nil
}

func parsePutIfExists(cmd *cobra.Command) (string, error) {
	ifExists, err := cmd.Flags().GetString("if-exists")
	if err != nil {
		return "", err
	}
	return normalizePutIfExists(ifExists)
}

func normalizePutIfExists(ifExists string) (string, error) {
	if ifExists == "" {
		return putIfExistsOverwrite, nil
	}
	switch ifExists {
	case putIfExistsOverwrite, putIfExistsSkip, putIfExistsFail:
		return ifExists, nil
	default:
		return "", fmt.Errorf("invalid --if-exists %q (use overwrite, skip, or fail)", ifExists)
	}
}

func putFile(src, dst string, opts putOptions) error {
	_, err := putFileWithResult(src, dst, opts)
	return err
}

func putFileWithResult(src, dst string, opts putOptions) (putResult, error) {
	ifExists, err := normalizePutIfExists(opts.ifExists)
	if err != nil {
		return putResult{}, err
	}

	dbx := filesNewFunc(config)
	action, existingMetadata, err := checkPutDestination(dbx, dst, ifExists)
	if err != nil {
		return putResult{}, err
	}
	if action == putDestinationSkip {
		reportPutSkipped(opts, dst)
		return newPutResult(putStatusSkipped, putKindFile, src, dst, existingMetadata), nil
	}

	contents, err := os.Open(src)
	if err != nil {
		return putResult{}, err
	}
	defer contents.Close()

	contentsInfo, err := contents.Stat()
	if err != nil {
		return putResult{}, err
	}

	commitInfo := files.NewCommitInfo(dst)
	commitInfo.Mode.Tag = writeModeForIfExists(ifExists)
	commitInfo.StrictConflict = ifExists != putIfExistsOverwrite

	// The Dropbox API only accepts timestamps in UTC with second precision.
	ts := time.Now().UTC().Round(time.Second)
	commitInfo.ClientModified = &ts

	if contentsInfo.Size() > singleShotUploadSizeCutoff {
		metadata, err := uploadChunked(dbx, uploadProgressReader(contents, contentsInfo.Size(), putErrorOutput(opts)), commitInfo, contentsInfo.Size(), opts.workers, opts.chunkSize, opts.debug)
		if err != nil && ifExists == putIfExistsSkip && isUploadDestinationFileConflict(err) {
			reportPutSkipped(opts, dst)
			return newPutResult(putStatusSkipped, putKindFile, src, dst, nil), nil
		}
		if err != nil {
			return putResult{}, err
		}
		return newPutResult(putStatusUploaded, putKindFile, src, dst, metadata), nil
	}

	uploadArg := &files.UploadArg{CommitInfo: *commitInfo}
	metadata, err := uploadSingleShot(dbx, contents, uploadArg, contentsInfo.Size(), putErrorOutput(opts))
	if err != nil && ifExists == putIfExistsSkip && isUploadDestinationFileConflict(err) {
		reportPutSkipped(opts, dst)
		return newPutResult(putStatusSkipped, putKindFile, src, dst, nil), nil
	}
	if err != nil {
		return putResult{}, err
	}
	return newPutResult(putStatusUploaded, putKindFile, src, dst, metadata), nil
}

func writeModeForIfExists(ifExists string) string {
	if ifExists == putIfExistsOverwrite {
		return files.WriteModeOverwrite
	}
	return files.WriteModeAdd
}

func checkPutStdinDestination(dbx files.Client, dst string, ifExists string) (putDestinationAction, files.IsMetadata, error) {
	ifExists, err := normalizePutIfExists(ifExists)
	if err != nil {
		return putDestinationUpload, nil, err
	}

	meta, exists, err := getDestinationMetadata(dbx, dst)
	if err != nil {
		if ifExists == putIfExistsOverwrite {
			return putDestinationUpload, nil, nil
		}
		return putDestinationUpload, nil, err
	}
	if !exists {
		return putDestinationUpload, nil, nil
	}
	if _, ok := meta.(*files.FolderMetadata); ok {
		return putDestinationUpload, nil, fmt.Errorf("cannot upload stdin to folder %q; provide a full Dropbox file path", dst)
	}
	return actionForExistingDestination(dst, ifExists, meta)
}

func checkPutDestination(dbx files.Client, dst string, ifExists string) (putDestinationAction, files.IsMetadata, error) {
	ifExists, err := normalizePutIfExists(ifExists)
	if err != nil {
		return putDestinationUpload, nil, err
	}
	if ifExists == putIfExistsOverwrite {
		return putDestinationUpload, nil, nil
	}

	meta, exists, err := getDestinationMetadata(dbx, dst)
	if err != nil {
		return putDestinationUpload, nil, err
	}
	if !exists {
		return putDestinationUpload, nil, nil
	}
	if _, ok := meta.(*files.FolderMetadata); ok {
		return putDestinationUpload, nil, fmt.Errorf("destination %q is a folder", dst)
	}
	return actionForExistingDestination(dst, ifExists, meta)
}

func actionForExistingDestination(dst string, ifExists string, metadata files.IsMetadata) (putDestinationAction, files.IsMetadata, error) {
	switch ifExists {
	case putIfExistsSkip:
		return putDestinationSkip, metadata, nil
	case putIfExistsFail:
		return putDestinationUpload, nil, fmt.Errorf("destination %q already exists", dst)
	default:
		return putDestinationUpload, nil, nil
	}
}

func getDestinationMetadata(dbx files.Client, dst string) (files.IsMetadata, bool, error) {
	meta, err := dbx.GetMetadata(files.NewGetMetadataArg(dst))
	if err != nil {
		if isGetMetadataNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return meta, true, nil
}

func isGetMetadataNotFoundError(err error) bool {
	var apiErr files.GetMetadataAPIError
	if errors.As(err, &apiErr) {
		return getMetadataAPIErrorIsNotFound(&apiErr)
	}
	var apiErrPtr *files.GetMetadataAPIError
	if errors.As(err, &apiErrPtr) {
		return getMetadataAPIErrorIsNotFound(apiErrPtr)
	}
	return false
}

func getMetadataAPIErrorIsNotFound(err *files.GetMetadataAPIError) bool {
	return err != nil &&
		err.EndpointError != nil &&
		err.EndpointError.Path != nil &&
		err.EndpointError.Path.Tag == files.LookupErrorNotFound
}

func isUploadDestinationFileConflict(err error) bool {
	var uploadErr files.UploadAPIError
	if errors.As(err, &uploadErr) &&
		uploadErr.EndpointError != nil &&
		uploadErr.EndpointError.Tag == files.UploadErrorPath &&
		uploadErr.EndpointError.Path != nil &&
		isWriteFileConflict(uploadErr.EndpointError.Path.Reason) {
		return true
	}

	var uploadErrPtr *files.UploadAPIError
	if errors.As(err, &uploadErrPtr) &&
		uploadErrPtr != nil &&
		uploadErrPtr.EndpointError != nil &&
		uploadErrPtr.EndpointError.Tag == files.UploadErrorPath &&
		uploadErrPtr.EndpointError.Path != nil &&
		isWriteFileConflict(uploadErrPtr.EndpointError.Path.Reason) {
		return true
	}

	var finishErr files.UploadSessionFinishAPIError
	if errors.As(err, &finishErr) &&
		finishErr.EndpointError != nil &&
		finishErr.EndpointError.Tag == files.UploadSessionFinishErrorPath &&
		isWriteFileConflict(finishErr.EndpointError.Path) {
		return true
	}

	var finishErrPtr *files.UploadSessionFinishAPIError
	if errors.As(err, &finishErrPtr) &&
		finishErrPtr != nil &&
		finishErrPtr.EndpointError != nil &&
		finishErrPtr.EndpointError.Tag == files.UploadSessionFinishErrorPath &&
		isWriteFileConflict(finishErrPtr.EndpointError.Path) {
		return true
	}

	return false
}

func isWriteFileConflict(err *files.WriteError) bool {
	return err != nil &&
		err.Tag == files.WriteErrorConflict &&
		err.Conflict != nil &&
		err.Conflict.Tag == files.WriteConflictErrorFile
}

func reportPutSkipped(opts putOptions, dst string) {
	putOutput(opts).Status("Skipping %s (already exists)", dst)
}

func putOutput(opts putOptions) *output.Renderer {
	if opts.output != nil {
		return opts.output
	}
	return output.New(nil, nil, output.FormatText)
}

func putErrorOutput(opts putOptions) io.Writer {
	if opts.errOut != nil {
		return opts.errOut
	}
	return os.Stderr
}

func putRecursive(src, dst string, opts putOptions) error {
	_, err := putRecursiveInternal(src, dst, opts, false)
	return err
}

func putRecursiveWithResults(src, dst string, opts putOptions) ([]putResult, error) {
	return putRecursiveInternal(src, dst, opts, true)
}

func putRecursiveInternal(src, dst string, opts putOptions, collectResults bool) ([]putResult, error) {
	src = filepath.Clean(src)
	var results []putResult
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
		putOutput(opts).Status("Processing %s -> %s", filePath, remotePath)

		if collectResults {
			result, err := putFileWithResult(filePath, remotePath, opts)
			if err != nil {
				uploadErrors = append(uploadErrors, fmt.Errorf("%s: %w", filePath, err))
				return nil
			}
			results = append(results, result)
			return nil
		}
		if err := putFile(filePath, remotePath, opts); err != nil {
			uploadErrors = append(uploadErrors, fmt.Errorf("%s: %w", filePath, err))
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	dbx := filesNewFunc(config)

	putOutput(opts).Status("Creating directory %s", dst)
	if collectResults {
		result, mkdirErr := putDirectoryWithResult(dbx, src, dst)
		if mkdirErr != nil {
			uploadErrors = append(uploadErrors, fmt.Errorf("mkdir %s: %w", dst, mkdirErr))
		} else {
			results = append(results, result)
		}
	} else {
		if mkdirErr := putDirectory(dbx, dst); mkdirErr != nil {
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
		putOutput(opts).Status("Creating directory %s", remotePath)
		if collectResults {
			result, mkdirErr := putDirectoryWithResult(dbx, dirPath, remotePath)
			if mkdirErr != nil {
				uploadErrors = append(uploadErrors, fmt.Errorf("mkdir %s: %w", remotePath, mkdirErr))
				return nil
			}
			results = append(results, result)
		} else {
			if mkdirErr := putDirectory(dbx, remotePath); mkdirErr != nil {
				uploadErrors = append(uploadErrors, fmt.Errorf("mkdir %s: %w", remotePath, mkdirErr))
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(uploadErrors) > 0 {
		return nil, fmt.Errorf("failed to upload %d file(s): %v", len(uploadErrors), uploadErrors[0])
	}
	return results, nil
}

func putDirectory(dbx files.Client, dst string) error {
	_, err := dbx.CreateFolderV2(files.NewCreateFolderArg(dst))
	if err == nil {
		return nil
	}
	return putDirectoryConflictError(dst, err)
}

func putDirectoryWithResult(dbx files.Client, src, dst string) (putResult, error) {
	arg := files.NewCreateFolderArg(dst)
	created, err := dbx.CreateFolderV2(arg)
	if err != nil {
		if conflictErr := putDirectoryConflictError(dst, err); conflictErr != nil {
			return putResult{}, conflictErr
		}
		metadata, metaErr := existingFolderMetadata(dbx, dst)
		if metaErr != nil {
			return putResult{}, metaErr
		}
		return newPutResult(putStatusExisting, putKindFolder, src, dst, metadata), nil
	}
	if created == nil {
		return newPutResult(putStatusCreated, putKindFolder, src, dst, nil), nil
	}
	return newPutResult(putStatusCreated, putKindFolder, src, dst, created.Metadata), nil
}

func putDirectoryConflictError(dst string, err error) error {
	conflictTag, ok := createFolderConflictTag(err)
	switch {
	case ok && conflictTag == files.WriteConflictErrorFolder:
		return nil
	case ok && (conflictTag == files.WriteConflictErrorFile || conflictTag == files.WriteConflictErrorFileAncestor):
		return fmt.Errorf("path exists and is not a folder: %s", dst)
	case ok:
		return err
	case isConflictError(err):
		return nil
	default:
		return err
	}
}

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put [flags] <source> [<target>]",
	Short: "Upload files or directories",
	Long: `Upload files or directories to Dropbox.
  - If target is not provided, uploads to the root of your Dropbox.
  - Use --recursive (-r) to upload entire directories.
  - Use - as source to read from stdin (target is required).
    Stdin is spooled to a temp file before upload and may use disk
    space up to the full input size.
`,
	Example: `  dbxcli put file.txt /destination/file.txt
  dbxcli put -r ./project /backup/project
  printf 'hello' | dbxcli put - /hello.txt
  tar cz ./src | dbxcli put - /backups/src.tgz`,
	RunE: put,
}

func init() {
	RootCmd.AddCommand(putCmd)
	enableStructuredOutput(putCmd)
	putCmd.Flags().BoolP("recursive", "r", false, "Recursively upload directories")
	putCmd.Flags().IntP("workers", "w", 4, "Number of concurrent upload workers to use")
	putCmd.Flags().Int64P("chunksize", "c", 1<<24, "Chunk size to use (should be multiple of 4MiB)")
	putCmd.Flags().BoolP("debug", "d", false, "Print debug timing")
	putCmd.Flags().String("if-exists", putIfExistsOverwrite, "What to do when the destination file exists: overwrite, skip, or fail")
}
