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
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/dropbox/dbxcli/v3/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/filetransfer"
	"github.com/spf13/cobra"
)

const (
	getStatusDownloaded = "downloaded"
	getStatusCreated    = "created"
	getStatusExisting   = "existing"

	getKindFile   = "file"
	getKindFolder = "folder"
)

type getOptions struct {
	errOut io.Writer
	// workers is the number of files downloaded concurrently by recursive
	// downloads; downloadWorkersAuto tunes it from measured throughput.
	workers int
}

type getCommandInput struct {
	Source    string `json:"source"`
	Target    string `json:"target"`
	Recursive bool   `json:"recursive"`
	Stdout    bool   `json:"stdout"`
}

type getResultInput struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type getResult struct {
	Status string         `json:"status"`
	Kind   string         `json:"kind"`
	Input  getResultInput `json:"input"`
	Result *jsonMetadata  `json:"result,omitempty"`
}

func get(cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 || len(args) > 2 {
		return invalidArgumentsErrorWithDetails("`get` requires `src` and/or `dst` arguments", argumentsErrorDetails("src", "dst"))
	}

	srcRef := newDropboxReference(args[0])
	src := srcRef.String()

	dst := ""
	if srcRef.isPath() {
		dst = path.Base(src)
	}

	dstExplicit := len(args) == 2
	if dstExplicit {
		dst = args[1]
	}

	recursive, _ := cmd.Flags().GetBool("recursive")
	opts, err := parseGetOptions(cmd)
	if err != nil {
		return err
	}

	if dst == "-" {
		if commandOutputFormat(cmd) == output.FormatJSON {
			return invalidArgumentsErrorWithDetails("`get --output=json` cannot be used with stdout target `-`", mergeJSONErrorDetails(operationErrorDetails("download"), argumentErrorDetails("dst"), flagErrorDetails("output")))
		}
		return getStdout(cmd, src, recursive)
	}

	dbx := filesNewFunc(config)

	meta, err := dbx.GetMetadataContext(currentContext(), files.NewGetMetadataArg(src))
	if err != nil {
		metadataErr := withJSONErrorDetails(fmt.Errorf("get metadata for %s: %v", src, err), operationErrorDetails("download"), pathErrorDetails(src))
		if recursive || dst == "" {
			return metadataErr
		}
		// For non-recursive, fall through to download (will fail with proper error)
		if f, statErr := os.Stat(dst); statErr == nil && f.IsDir() {
			if !srcRef.isPath() {
				return metadataErr
			}
			dst = filepath.Join(dst, path.Base(src))
		}
		result, err := downloadFileWithResult(dbx, src, dst, nil, true, opts)
		if err != nil {
			return withJSONErrorDetails(err, operationErrorDetails("download"), pathErrorDetails(src), relocationErrorDetails(src, dst))
		}
		return renderGetResults(cmd, getCommandInput{
			Source:    src,
			Target:    dst,
			Recursive: false,
			Stdout:    false,
		}, []getResult{result})
	}

	sourceName := path.Base(src)
	if !srcRef.isPath() {
		sourceName = metadataName(meta)
		if sourceName == "" {
			return withJSONErrorDetails(fmt.Errorf("get metadata for %s did not include a name", src), operationErrorDetails("download"), pathErrorDetails(src))
		}
		if dst == "" {
			dst = sourceName
		}
	}

	if _, ok := meta.(*files.FolderMetadata); ok {
		if !recursive {
			return invalidArgumentsErrorfWithDetails("%s is a folder (use --recursive to download folders)", mergeJSONErrorDetails(operationErrorDetails("download"), pathErrorDetails(src)), src)
		}
		if f, statErr := os.Stat(dst); statErr == nil && f.IsDir() {
			dst = filepath.Join(dst, sourceName)
		}
		if commandOutputFormat(cmd) == output.FormatText {
			return withJSONErrorDetails(getRecursiveWithRootMetadata(dbx, src, dst, meta, opts), operationErrorDetails("download"), pathErrorDetails(src), relocationErrorDetails(src, dst))
		}
		results, err := getRecursiveWithResults(dbx, src, dst, meta, opts)
		if err != nil {
			return withJSONErrorDetails(err, operationErrorDetails("download"), pathErrorDetails(src), relocationErrorDetails(src, dst))
		}
		return renderGetResults(cmd, getCommandInput{
			Source:    src,
			Target:    dst,
			Recursive: true,
			Stdout:    false,
		}, results)
	}

	dstFilenameExplicit := dstExplicit
	if f, statErr := os.Stat(dst); statErr == nil && f.IsDir() {
		dst = filepath.Join(dst, sourceName)
		dstFilenameExplicit = false
	}

	fileMeta, ok := meta.(*files.FileMetadata)
	if !ok {
		return fmt.Errorf("unexpected metadata type for %s", src)
	}
	result, err := downloadFileWithResult(dbx, src, dst, fileMeta, dstFilenameExplicit, opts)
	if err != nil {
		return withJSONErrorDetails(err, operationErrorDetails("download"), pathErrorDetails(src), relocationErrorDetails(src, dst))
	}
	return renderGetResults(cmd, getCommandInput{
		Source:    src,
		Target:    result.Input.Target,
		Recursive: false,
		Stdout:    false,
	}, []getResult{result})
}

func parseGetOptions(cmd *cobra.Command) (getOptions, error) {
	opts := getOptions{
		errOut:  cmd.ErrOrStderr(),
		workers: downloadWorkersAuto,
	}
	if cmd.Flags().Lookup("workers") != nil {
		workers, err := cmd.Flags().GetInt("workers")
		if err != nil {
			return getOptions{}, err
		}
		if workers < 0 {
			return getOptions{}, invalidArgumentsErrorWithDetails("`--workers` must be greater than or equal to 0 (0 selects automatic tuning)", flagErrorDetails("workers"))
		}
		opts.workers = workers
	}
	return opts, nil
}

func getErrorOutput(opts getOptions) io.Writer {
	if opts.errOut != nil {
		return opts.errOut
	}
	return os.Stderr
}

func newGetResult(status, kind, source, target string, metadata files.IsMetadata) (getResult, error) {
	result := getResult{
		Status: status,
		Kind:   kind,
		Input: getResultInput{
			Source: source,
			Target: target,
		},
	}
	if metadata != nil {
		jsonResult, err := jsonMetadataFromDropbox(metadata)
		if err != nil {
			return getResult{}, err
		}
		result.Result = &jsonResult
	}
	return result, nil
}

func renderGetResults(cmd *cobra.Command, input getCommandInput, results []getResult) error {
	return renderJSONOperationOutput(cmd, input, getOperationResults(results))
}

func getOperationResults(results []getResult) []jsonOperationResult {
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

func getStdout(cmd *cobra.Command, src string, recursive bool) error {
	if recursive {
		return invalidArgumentsErrorWithDetails("`get -` cannot be used with --recursive", mergeJSONErrorDetails(operationErrorDetails("download"), flagErrorDetails("recursive")))
	}

	dbx := filesNewFunc(config)

	meta, err := dbx.GetMetadataContext(currentContext(), files.NewGetMetadataArg(src))
	if err == nil {
		if _, ok := meta.(*files.FolderMetadata); ok {
			return invalidArgumentsErrorfWithDetails("%s is a folder; cannot download folder to stdout", mergeJSONErrorDetails(operationErrorDetails("download"), pathErrorDetails(src)), src)
		}
	}

	fileMeta, _ := meta.(*files.FileMetadata)
	return withJSONErrorDetails(downloadToStdoutWithMetadata(dbx, src, fileMeta, cmd.OutOrStdout()), operationErrorDetails("download"), pathErrorDetails(src))
}

func getRecursive(dbx filesClient, src, dst string) error {
	_, err := getRecursiveInternal(dbx, src, dst, nil, getOptions{}, false)
	return err
}

func getRecursiveWithRootMetadata(dbx filesClient, src, dst string, rootMeta files.IsMetadata, opts getOptions) error {
	_, err := getRecursiveInternal(dbx, src, dst, rootMeta, opts, false)
	return err
}

func getRecursiveWithResults(dbx filesClient, src, dst string, rootMeta files.IsMetadata, opts getOptions) ([]getResult, error) {
	return getRecursiveInternal(dbx, src, dst, rootMeta, opts, true)
}

func getRecursiveInternal(dbx filesClient, src, dst string, rootMeta files.IsMetadata, opts getOptions, collectResults bool) ([]getResult, error) {
	arg := files.NewListFolderArg(src)
	arg.Recursive = true

	res, err := dbx.ListFolderContext(currentContext(), arg)
	if err != nil {
		return nil, withJSONErrorDetails(fmt.Errorf("list folder %s: %v", src, err), operationErrorDetails("download"), pathErrorDetails(src))
	}

	var entries []files.IsMetadata
	entries = append(entries, res.Entries...)
	for res.HasMore {
		cont := files.NewListFolderContinueArg(res.Cursor)
		res, err = dbx.ListFolderContinueContext(currentContext(), cont)
		if err != nil {
			return nil, withJSONErrorDetails(fmt.Errorf("list folder continue: %v", err), operationErrorDetails("download"), pathErrorDetails(src))
		}
		entries = append(entries, res.Entries...)
	}

	var results []getResult
	rootPath := src
	if metadataPath := metadataPathDisplay(rootMeta); metadataPath != "" {
		rootPath = metadataPath
	}

	if collectResults {
		result, err := ensureLocalDirectoryResult(src, dst, rootMeta)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	} else {
		if err := os.MkdirAll(dst, 0755); err != nil {
			return nil, err
		}
	}

	// Folders are created up front and in listing order; files are queued as
	// jobs and downloaded concurrently. Results and errors are kept in listing
	// order so output is stable regardless of completion order.
	entryResults := make([]*getResult, len(entries))
	entryErrors := make([]error, len(entries))
	var jobIndexes []int
	var jobFiles []*files.FileMetadata
	var jobTargets []string
	var totalBytes int64

	for i, entry := range entries {
		switch f := entry.(type) {
		case *files.FolderMetadata:
			relPath, err := relativeTo(rootPath, f.PathDisplay)
			if err != nil {
				entryErrors[i] = err
				continue
			}
			if relPath == "" {
				continue
			}
			localDir := filepath.Join(dst, filepath.FromSlash(relPath))
			if collectResults {
				result, err := ensureLocalDirectoryResult(f.PathDisplay, localDir, f)
				if err != nil {
					entryErrors[i] = fmt.Errorf("mkdir %s: %w", localDir, err)
					continue
				}
				entryResults[i] = &result
			} else {
				if err := os.MkdirAll(localDir, 0755); err != nil {
					entryErrors[i] = fmt.Errorf("mkdir %s: %w", localDir, err)
				}
			}
		case *files.FileMetadata:
			relPath, err := relativeTo(rootPath, f.PathDisplay)
			if err != nil {
				entryErrors[i] = err
				continue
			}
			jobIndexes = append(jobIndexes, i)
			jobFiles = append(jobFiles, f)
			jobTargets = append(jobTargets, filepath.Join(dst, filepath.FromSlash(relPath)))
			if f.Size <= math.MaxInt64 {
				totalBytes += int64(f.Size)
			}
		}
	}

	errOut := getErrorOutput(opts)
	status := newDownloadStatusWriter(errOut)
	pool := newDownloadPool(opts.workers, len(jobFiles), totalBytes, status)
	concurrent := pool.concurrent()

	jobs := make([]func(), len(jobFiles))
	for j := range jobFiles {
		i, f, localPath := jobIndexes[j], jobFiles[j], jobTargets[j]
		jobs[j] = func() {
			if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
				entryErrors[i] = fmt.Errorf("mkdir %s: %w", filepath.Dir(localPath), err)
				return
			}

			// With several downloads in flight the per-file progress bars
			// would overwrite each other, so concurrent mode reports only
			// one line per file plus the pool's aggregate status line.
			fileErrOut := errOut
			var progress downloadProgressFunc
			if concurrent {
				status.message("Downloading %s -> %s\n", f.PathDisplay, localPath)
				fileErrOut = io.Discard
				if !isExportOnlyFile(f) {
					progress = pool.fileProgress()
				}
			} else {
				fmt.Fprintf(errOut, "Downloading %s -> %s\n", f.PathDisplay, localPath)
			}

			metadata, actualDst, err := downloadFileWithProgress(dbx, f.PathDisplay, localPath, f, false, fileErrOut, progress)
			if err != nil {
				entryErrors[i] = fmt.Errorf("%s: %w", f.PathDisplay, err)
				return
			}
			// Exported files report no streaming progress, so count them once
			// they have been written in full.
			if concurrent && progress == nil && metadata != nil && metadata.Size <= math.MaxInt64 {
				pool.addBytes(int64(metadata.Size))
			}
			if collectResults {
				result, err := newGetResult(getStatusDownloaded, getKindFile, f.PathDisplay, actualDst, metadata)
				if err != nil {
					entryErrors[i] = fmt.Errorf("%s: %w", f.PathDisplay, err)
					return
				}
				entryResults[i] = &result
			}
		}
	}

	pool.run(currentContext(), jobs, func(j int, err error) {
		entryErrors[jobIndexes[j]] = fmt.Errorf("%s: %w", jobFiles[j].PathDisplay, err)
	})

	var downloadErrors []error
	for i := range entries {
		if entryErrors[i] != nil {
			downloadErrors = append(downloadErrors, entryErrors[i])
			continue
		}
		if entryResults[i] != nil {
			results = append(results, *entryResults[i])
		}
	}

	if len(downloadErrors) > 0 {
		for _, e := range downloadErrors {
			fmt.Fprintf(errOut, "Error: %v\n", e)
		}
		return nil, commandFailedErrorfWithDetails("get: %d error(s)", mergeJSONErrorDetails(operationErrorDetails("download"), pathErrorDetails(src), relocationErrorDetails(src, dst)), len(downloadErrors))
	}

	return results, nil
}

func ensureLocalDirectoryResult(source, target string, metadata files.IsMetadata) (getResult, error) {
	status := getStatusCreated
	if info, err := os.Stat(target); err == nil {
		if !info.IsDir() {
			return getResult{}, pathConflictErrorWithPath(target, "path exists and is not a folder: %s", target)
		}
		status = getStatusExisting
	} else if !os.IsNotExist(err) {
		return getResult{}, err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return getResult{}, err
	}
	return newGetResult(status, getKindFolder, source, target, metadata)
}

func relativeTo(base, full string) (string, error) {
	baseLower := strings.ToLower(base)
	fullLower := strings.ToLower(full)
	if fullLower != baseLower && !strings.HasPrefix(fullLower, baseLower+"/") {
		return "", fmt.Errorf("path %q is not under %q", full, base)
	}
	rel := full[len(base):]
	rel = strings.TrimPrefix(rel, "/")
	return rel, nil
}

func downloadFile(dbx filesClient, src string, dst string) error {
	_, _, err := downloadFileWithMetadata(dbx, src, dst, nil, false, os.Stderr)
	return err
}

func downloadFileWithResult(
	dbx filesClient,
	src string,
	dst string,
	metadata *files.FileMetadata,
	dstExplicit bool,
	opts getOptions,
) (getResult, error) {
	metadata, actualDst, err := downloadFileWithMetadata(dbx, src, dst, metadata, dstExplicit, getErrorOutput(opts))
	if err != nil {
		return getResult{}, err
	}
	return newGetResult(getStatusDownloaded, getKindFile, src, actualDst, metadata)
}

func downloadFileWithMetadata(
	dbx filesClient,
	src string,
	dst string,
	metadata *files.FileMetadata,
	dstExplicit bool,
	errOut io.Writer,
) (*files.FileMetadata, string, error) {
	return downloadFileWithProgress(dbx, src, dst, metadata, dstExplicit, errOut, nil)
}

// downloadProgressFunc receives the monotonic number of bytes committed so
// far for one file and the file's total size.
type downloadProgressFunc func(committed, total int64)

func downloadFileWithProgress(
	dbx filesClient,
	src string,
	dst string,
	metadata *files.FileMetadata,
	dstExplicit bool,
	errOut io.Writer,
	progress downloadProgressFunc,
) (*files.FileMetadata, string, error) {
	if !isExportOnlyFile(metadata) {
		result, err := downloadFileOnce(dbx, src, dst, errOut, progress)
		return result, dst, err
	}

	var result *files.FileMetadata
	actualDst := dst
	err := retryWithBackoff(func() error {
		var err error
		result, actualDst, err = exportFileToPath(dbx, src, dst, dstExplicit)
		return err
	})
	return result, actualDst, err
}

func createDownloadTemp(dst string) (*os.File, string, error) {
	dir := filepath.Dir(dst)
	base := filepath.Base(dst)
	for i := range 100 {
		tmp := filepath.Join(dir, fmt.Sprintf(".%s.tmp-%d-%d", base, os.Getpid(), time.Now().UnixNano()+int64(i)))
		f, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		return f, tmp, err
	}
	return nil, "", fmt.Errorf("failed to create temporary file for %s", dst)
}

func downloadDestinationPath(dst string) (string, error) {
	for range 255 {
		info, err := os.Lstat(dst)
		if err != nil {
			if os.IsNotExist(err) {
				return dst, nil
			}
			return "", err
		}
		if info.Mode()&os.ModeSymlink == 0 {
			return dst, nil
		}

		target, err := os.Readlink(dst)
		if err != nil {
			return "", err
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(dst), target)
		}
		dst = target
	}

	return "", fmt.Errorf("too many symlinks resolving %s", dst)
}

func downloadFileOnce(dbx filesClient, src string, dst string, errOut io.Writer, onProgress downloadProgressFunc) (*files.FileMetadata, error) {
	finalDst, err := downloadDestinationPath(dst)
	if err != nil {
		return nil, err
	}
	if errOut == nil {
		errOut = io.Discard
	}

	drawer := newTransferProgressDrawer(errOut, "Downloading ")
	defer drawer.finish()

	result, err := filetransfer.NewDownloader(dbx).Download(
		currentContext(),
		src,
		filetransfer.File(finalDst),
		filetransfer.DownloadOptions{
			MaxAttempts: maxRetries + 1,
			Progress: func(progress filetransfer.DownloadProgress) {
				drawer.update(progress.BytesCommitted, progress.TotalBytes)
				if onProgress != nil {
					onProgress(progress.BytesCommitted, progress.TotalBytes)
				}
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return result.Metadata, nil
}

func exportFile(
	dbx filesClient,
	src string,
) (*files.ExportResult, io.ReadCloser, error) {
	return dbx.ExportContext(
		currentContext(),
		files.NewExportArg(src),
	)
}

func exportFileToPath(dbx filesClient, src string, dst string, dstExplicit bool) (*files.FileMetadata, string, error) {
	res, contents, err := exportFile(dbx, src)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = contents.Close() }()

	if !dstExplicit {
		dst = filepath.Join(filepath.Dir(dst), res.ExportMetadata.Name)
	}

	finalDst, err := downloadDestinationPath(dst)
	if err != nil {
		return nil, "", err
	}

	f, tmp, err := createDownloadTemp(finalDst)
	if err != nil {
		return nil, "", err
	}

	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tmp)
		}
	}()

	_, copyErr := io.Copy(f, contents)
	closeErr := f.Close()

	if copyErr != nil {
		return nil, "", copyErr
	}
	if closeErr != nil {
		return nil, "", closeErr
	}

	if err := os.Rename(tmp, finalDst); err != nil {
		return nil, "", err
	}

	removeTemp = false
	return res.FileMetadata, dst, nil
}

func isExportOnlyFile(metadata *files.FileMetadata) bool {
	return metadata != nil &&
		metadata.ExportInfo != nil &&
		metadata.ExportInfo.ExportAs != ""
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [flags] <source> [<target>]",
	Short: "Download a file or folder",
	Long: `Download a file or folder from Dropbox.
  - Source may be a Dropbox path, file ID (id:), revision (rev:), or
    namespace-relative path (ns:).
  - Use --recursive (-r) to download entire directories.
  - Recursive downloads fetch several files in parallel. By default the
    number of concurrent downloads is tuned automatically from the measured
    throughput; use --workers (-w) to set a fixed number instead.
  - Use - as target to write file bytes to stdout.
    Stdout is byte-clean: all progress and errors go to stderr.
`,
	Example: `  dbxcli get /remote/file.txt ./local-file.txt
  dbxcli get rev:a1c10ce0dd78 ./historical-file.txt
  dbxcli get -r /remote/folder ./local-folder
  dbxcli get -r -w 8 /remote/folder ./local-folder
  dbxcli get /backups/src.tgz - | tar tz
  dbxcli get /file.txt - > local-copy.txt`,
	RunE: get,
}

func init() {
	RootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolP("recursive", "r", false, "Recursively download a folder")
	getCmd.Flags().IntP("workers", "w", downloadWorkersAuto, "Number of files to download concurrently with --recursive (0 = auto-tune from measured bandwidth)")
	enableStructuredOutput(getCmd)
}
