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
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/ioprogress"
	"github.com/spf13/cobra"
)

func get(cmd *cobra.Command, args []string) (err error) {
	dbx := files.New(config)
	return getWithClient(dbx, args)
}

func getWithClient(dbx files.Client, args []string) (err error) {
	if len(args) == 0 || len(args) > 2 {
		return errors.New("`get` requires `src` and/or `dst` arguments")
	}

	src, err := validatePath(args[0])
	if err != nil {
		return
	}

	// Default `dst` to the base segment of the source path; use the second argument if provided.
	dst := path.Base(src)
	if len(args) == 2 {
		dst = args[1]
	}
	// If `dst` is a directory, append the source filename.
	if f, err := os.Stat(dst); err == nil && f.IsDir() {
		dst = path.Join(dst, path.Base(src))
	}

	return downloadFile(dbx, src, dst)
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
	defer contents.Close()

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
	Short: "Download a file",
	RunE:  get,
}

func init() {
	RootCmd.AddCommand(getCmd)
}
