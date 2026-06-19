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

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/ioprogress"
	"github.com/spf13/cobra"
)

func get(cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 || len(args) > 2 {
		return errors.New("`get` requires `src` and/or `dst` arguments")
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

	if dst == "-" {
		return getStdout(cmd, src, recursive)
	}

	dbx := filesNewFunc(config)

	meta, err := dbx.GetMetadata(files.NewGetMetadataArg(src))
	if err != nil {
		if recursive {
			return fmt.Errorf("get metadata for %s: %v", src, err)
		}
		// For non-recursive, fall through to download (will fail with proper error)
		if f, statErr := os.Stat(dst); statErr == nil && f.IsDir() {
			dst = filepath.Join(dst, path.Base(src))
		}
		return downloadFile(dbx, src, dst)
	}

	if _, ok := meta.(*files.FolderMetadata); ok {
		if !recursive {
			return fmt.Errorf("%s is a folder (use --recursive to download folders)", src)
		}
		if f, statErr := os.Stat(dst); statErr == nil && f.IsDir() {
			dst = filepath.Join(dst, path.Base(src))
		}
		return getRecursive(dbx, src, dst)
	}

	if f, statErr := os.Stat(dst); statErr == nil && f.IsDir() {
		dst = filepath.Join(dst, path.Base(src))
	}

	return downloadFile(dbx, src, dst)
}

func getStdout(cmd *cobra.Command, src string, recursive bool) error {
	if recursive {
		return errors.New("`get -` cannot be used with --recursive")
	}

	dbx := filesNewFunc(config)

	meta, err := dbx.GetMetadata(files.NewGetMetadataArg(src))
	if err == nil {
		if _, ok := meta.(*files.FolderMetadata); ok {
			return fmt.Errorf("%s is a folder; cannot download folder to stdout", src)
		}
	}

	return downloadToStdout(dbx, src, cmd.OutOrStdout())
}

func getWithClient(dbx files.Client, args []string) (err error) {
	if len(args) == 0 || len(args) > 2 {
		return errors.New("`get` requires `src` and/or `dst` arguments")
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

	return downloadFile(dbx, src, dst)
}

func getRecursive(dbx files.Client, src, dst string) error {
	arg := files.NewListFolderArg(src)
	arg.Recursive = true

	res, err := dbx.ListFolder(arg)
	if err != nil {
		return fmt.Errorf("list folder %s: %v", src, err)
	}

	var entries []files.IsMetadata
	entries = append(entries, res.Entries...)
	for res.HasMore {
		cont := files.NewListFolderContinueArg(res.Cursor)
		res, err = dbx.ListFolderContinue(cont)
		if err != nil {
			return fmt.Errorf("list folder continue: %v", err)
		}
		entries = append(entries, res.Entries...)
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
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
			localDir := filepath.Join(dst, filepath.FromSlash(relPath))
			if err := os.MkdirAll(localDir, 0755); err != nil {
				downloadErrors = append(downloadErrors, fmt.Errorf("mkdir %s: %w", localDir, err))
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
			fmt.Fprintf(os.Stderr, "Downloading %s -> %s\n", f.PathDisplay, localPath)
			if err := downloadFile(dbx, f.PathDisplay, localPath); err != nil {
				downloadErrors = append(downloadErrors, fmt.Errorf("%s: %w", f.PathDisplay, err))
			}
		}
	}

	if len(downloadErrors) > 0 {
		for _, e := range downloadErrors {
			fmt.Fprintf(os.Stderr, "Error: %v\n", e)
		}
		return fmt.Errorf("get: %d error(s)", len(downloadErrors))
	}

	return nil
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
	arg := files.NewDownloadArg(src)

	return retryWithBackoff(func() error {
		return downloadFileOnce(dbx, arg, dst)
	})
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

func downloadFileOnce(dbx files.Client, arg *files.DownloadArg, dst string) error {
	res, contents, err := dbx.Download(arg)
	if err != nil {
		return err
	}
	defer func() { _ = contents.Close() }()

	finalDst, err := downloadDestinationPath(dst)
	if err != nil {
		return err
	}

	f, tmp, err := createDownloadTemp(finalDst)
	if err != nil {
		return err
	}
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tmp)
		}
	}()

	progressbar := &ioprogress.Reader{
		Reader: contents,
		DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
			return fmt.Sprintf("Downloading %s/%s",
				humanize.IBytes(uint64(progress)), humanize.IBytes(uint64(total)))
		}),
		Size: int64(res.Size),
	}

	_, copyErr := io.Copy(f, progressbar)
	closeErr := f.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Rename(tmp, finalDst); err != nil {
		return err
	}
	removeTemp = false
	return nil
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
}
