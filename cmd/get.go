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
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/dropbox/dbxcli/v3/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/ioprogress"
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

	src, err := validatePath(args[0])
	if err != nil {
		return
	}

	dst := path.Base(src)
	if len(args) == 2 {
		dst = args[1]
	}

	recursive, _ := cmd.Flags().GetBool("recursive")
	opts := parseGetOptions(cmd)

	if dst == "-" {
		if commandOutputFormat(cmd) == output.FormatJSON {
			return invalidArgumentsErrorWithDetails("`get --output=json` cannot be used with stdout target `-`", mergeJSONErrorDetails(operationErrorDetails("download"), argumentErrorDetails("dst"), flagErrorDetails("output")))
		}
		return getStdout(cmd, src, recursive)
	}

	dbx := filesNewFunc(config)

	meta, err := dbx.GetMetadata(files.NewGetMetadataArg(src))
	if err != nil {
		if recursive {
			return withJSONErrorDetails(fmt.Errorf("get metadata for %s: %v", src, err), operationErrorDetails("download"), pathErrorDetails(src))
		}
		// For non-recursive, fall through to download (will fail with proper error)
		if f, statErr := os.Stat(dst); statErr == nil && f.IsDir() {
			dst = filepath.Join(dst, path.Base(src))
		}
		result, err := downloadFileWithResult(dbx, src, dst, opts)
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

	if _, ok := meta.(*files.FolderMetadata); ok {
		if !recursive {
			return invalidArgumentsErrorfWithDetails("%s is a folder (use --recursive to download folders)", mergeJSONErrorDetails(operationErrorDetails("download"), pathErrorDetails(src)), src)
		}
		if f, statErr := os.Stat(dst); statErr == nil && f.IsDir() {
			dst = filepath.Join(dst, path.Base(src))
		}
		if commandOutputFormat(cmd) == output.FormatText {
			return withJSONErrorDetails(getRecursive(dbx, src, dst), operationErrorDetails("download"), pathErrorDetails(src), relocationErrorDetails(src, dst))
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

	if f, statErr := os.Stat(dst); statErr == nil && f.IsDir() {
		dst = filepath.Join(dst, path.Base(src))
	}

	result, err := downloadFileWithResult(dbx, src, dst, opts)
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

func parseGetOptions(cmd *cobra.Command) getOptions {
	return getOptions{
		errOut: cmd.ErrOrStderr(),
	}
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

	meta, err := dbx.GetMetadata(files.NewGetMetadataArg(src))
	if err == nil {
		if _, ok := meta.(*files.FolderMetadata); ok {
			return invalidArgumentsErrorfWithDetails("%s is a folder; cannot download folder to stdout", mergeJSONErrorDetails(operationErrorDetails("download"), pathErrorDetails(src)), src)
		}
	}

	return withJSONErrorDetails(downloadToStdout(dbx, src, cmd.OutOrStdout()), operationErrorDetails("download"), pathErrorDetails(src))
}

func getWithClient(dbx files.Client, args []string) (err error) {
	if len(args) == 0 || len(args) > 2 {
		return invalidArgumentsErrorWithDetails("`get` requires `src` and/or `dst` arguments", argumentsErrorDetails("src", "dst"))
	}

	src, err := validatePath(args[0])
	if err != nil {
		return
	}

	dst := path.Base(src)
	if len(args) == 2 {
		dst = args[1]
	}
	if f, err := os.Stat(dst); err == nil && f.IsDir() {
		dst = filepath.Join(dst, path.Base(src))
	}

	return withJSONErrorDetails(downloadFile(dbx, src, dst), operationErrorDetails("download"), pathErrorDetails(src), relocationErrorDetails(src, dst))
}

func getRecursive(dbx files.Client, src, dst string) error {
	_, err := getRecursiveInternal(dbx, src, dst, nil, getOptions{}, false)
	return err
}

func getRecursiveWithResults(dbx files.Client, src, dst string, rootMeta files.IsMetadata, opts getOptions) ([]getResult, error) {
	return getRecursiveInternal(dbx, src, dst, rootMeta, opts, true)
}

func getRecursiveInternal(dbx files.Client, src, dst string, rootMeta files.IsMetadata, opts getOptions, collectResults bool) ([]getResult, error) {
	arg := files.NewListFolderArg(src)
	arg.Recursive = true

	res, err := dbx.ListFolder(arg)
	if err != nil {
		return nil, withJSONErrorDetails(fmt.Errorf("list folder %s: %v", src, err), operationErrorDetails("download"), pathErrorDetails(src))
	}

	var entries []files.IsMetadata
	entries = append(entries, res.Entries...)
	for res.HasMore {
		cont := files.NewListFolderContinueArg(res.Cursor)
		res, err = dbx.ListFolderContinue(cont)
		if err != nil {
			return nil, withJSONErrorDetails(fmt.Errorf("list folder continue: %v", err), operationErrorDetails("download"), pathErrorDetails(src))
		}
		entries = append(entries, res.Entries...)
	}

	var results []getResult

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

	var downloadErrors []error

	for _, entry := range entries {
		switch f := entry.(type) {
		case *files.FolderMetadata:
			relPath, err := relativeTo(src, f.PathDisplay)
			if err != nil {
				downloadErrors = append(downloadErrors, err)
				continue
			}
			if relPath == "" {
				continue
			}
			localDir := filepath.Join(dst, filepath.FromSlash(relPath))
			if collectResults {
				result, err := ensureLocalDirectoryResult(f.PathDisplay, localDir, f)
				if err != nil {
					downloadErrors = append(downloadErrors, fmt.Errorf("mkdir %s: %w", localDir, err))
					continue
				}
				results = append(results, result)
			} else {
				if err := os.MkdirAll(localDir, 0755); err != nil {
					downloadErrors = append(downloadErrors, fmt.Errorf("mkdir %s: %w", localDir, err))
				}
			}
		case *files.FileMetadata:
			relPath, err := relativeTo(src, f.PathDisplay)
			if err != nil {
				downloadErrors = append(downloadErrors, err)
				continue
			}
			localPath := filepath.Join(dst, filepath.FromSlash(relPath))
			if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
				downloadErrors = append(downloadErrors, fmt.Errorf("mkdir %s: %w", filepath.Dir(localPath), err))
				continue
			}
			fmt.Fprintf(getErrorOutput(opts), "Downloading %s -> %s\n", f.PathDisplay, localPath)
			if collectResults {
				result, err := downloadFileWithResult(dbx, f.PathDisplay, localPath, opts)
				if err != nil {
					downloadErrors = append(downloadErrors, fmt.Errorf("%s: %w", f.PathDisplay, err))
					continue
				}
				results = append(results, result)
				continue
			}
			if err := downloadFile(dbx, f.PathDisplay, localPath); err != nil {
				downloadErrors = append(downloadErrors, fmt.Errorf("%s: %w", f.PathDisplay, err))
			}
		}
	}

	if len(downloadErrors) > 0 {
		for _, e := range downloadErrors {
			fmt.Fprintf(getErrorOutput(opts), "Error: %v\n", e)
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

func downloadFile(dbx files.Client, src string, dst string) error {
	_, err := downloadFileWithMetadata(dbx, src, dst, os.Stderr)
	return err
}

func downloadFileWithResult(dbx files.Client, src string, dst string, opts getOptions) (getResult, error) {
	metadata, err := downloadFileWithMetadata(dbx, src, dst, getErrorOutput(opts))
	if err != nil {
		return getResult{}, err
	}
	return newGetResult(getStatusDownloaded, getKindFile, src, dst, metadata)
}

func downloadFileWithMetadata(dbx files.Client, src string, dst string, errOut io.Writer) (*files.FileMetadata, error) {
	arg := files.NewDownloadArg(src)
	var metadata *files.FileMetadata

	err := retryWithBackoff(func() error {
		var err error
		metadata, err = downloadFileOnce(dbx, arg, dst, errOut)
		return err
	})
	return metadata, err
}

func createDownloadTemp(dst string) (*os.File, string, error) {
	dir := filepath.Dir(dst)
	base := filepath.Base(dst)
	for i := 0; i < 100; i++ {
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
	for i := 0; i < 255; i++ {
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

func downloadFileOnce(dbx files.Client, arg *files.DownloadArg, dst string, errOut io.Writer) (*files.FileMetadata, error) {
	res, contents, err := dbx.Download(arg)
	if err != nil {
		return nil, err
	}
	defer func() { _ = contents.Close() }()

	finalDst, err := downloadDestinationPath(dst)
	if err != nil {
		return nil, err
	}

	f, tmp, err := createDownloadTemp(finalDst)
	if err != nil {
		return nil, err
	}
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tmp)
		}
	}()

	progressbar := &ioprogress.Reader{
		Reader: contents,
		DrawFunc: ioprogress.DrawTerminalf(errOut, func(progress, total int64) string {
			return fmt.Sprintf("Downloading %s/%s",
				humanize.IBytes(uint64(progress)), humanize.IBytes(uint64(total)))
		}),
		Size: downloadMetadataSize(res),
	}

	_, copyErr := io.Copy(f, progressbar)
	closeErr := f.Close()
	if copyErr != nil {
		return nil, copyErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if err := os.Rename(tmp, finalDst); err != nil {
		return nil, err
	}
	removeTemp = false
	return res, nil
}

func downloadMetadataSize(metadata *files.FileMetadata) int64 {
	if metadata == nil {
		return 0
	}
	return int64(metadata.Size)
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [flags] <source> [<target>]",
	Short: "Download a file or folder",
	Long: `Download a file or folder from Dropbox.
  - Use --recursive (-r) to download entire directories.
  - Use - as target to write file bytes to stdout.
    Stdout is byte-clean: all progress and errors go to stderr.
`,
	Example: `  dbxcli get /remote/file.txt ./local-file.txt
  dbxcli get -r /remote/folder ./local-folder
  dbxcli get /backups/src.tgz - | tar tz
  dbxcli get /file.txt - > local-copy.txt`,
	RunE: get,
}

func init() {
	RootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolP("recursive", "r", false, "Recursively download a folder")
	enableStructuredOutput(getCmd)
}
