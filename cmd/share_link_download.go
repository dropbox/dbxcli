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

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/ioprogress"
	"github.com/spf13/cobra"
)

func shareLinkDownload(cmd *cobra.Command, args []string) error {
	if len(args) == 0 || len(args) > 2 {
		return errors.New("`share-link download` requires a `url` and optional `target` argument")
	}

	url := args[0]
	if url == "" {
		return errors.New("`share-link download` requires a non-empty URL")
	}

	target := ""
	if len(args) == 2 {
		target = args[1]
		if target == "" {
			return errors.New("`share-link download` requires a non-empty target")
		}
	}

	arg := sharing.NewGetSharedLinkMetadataArg(url)
	password, err := cmd.Flags().GetString("password")
	if err != nil {
		return err
	}
	arg.LinkPassword = password

	dbx := newSharedLinkClient(config)
	if target == "-" {
		if err := downloadSharedLinkToStdout(dbx, arg, cmd.OutOrStdout()); err != nil {
			return err
		}
		commandVerboseStatus(cmd, "Downloaded shared link to stdout")
		return nil
	}

	dst, err := downloadSharedLinkToFile(dbx, arg, target, cmd.ErrOrStderr())
	if err != nil {
		return err
	}
	commandVerboseStatus(cmd, "Downloaded shared link to %s", dst)
	return nil
}

func downloadSharedLinkToFile(dbx sharedLinkClient, arg *sharing.GetSharedLinkMetadataArg, target string, errOut io.Writer) (string, error) {
	var dst string
	err := retryWithBackoff(func() error {
		link, contents, err := dbx.GetSharedLinkFile(arg)
		if err != nil {
			return err
		}
		if contents == nil {
			return errors.New("shared link download response did not include file content")
		}
		defer func() { _ = contents.Close() }()

		dst, err = sharedLinkDownloadTarget(target, link)
		if err != nil {
			return err
		}

		return copySharedLinkContentToFile(contents, sharedLinkDownloadSize(link), dst, errOut)
	})
	return dst, err
}

func downloadSharedLinkToStdout(dbx sharedLinkClient, arg *sharing.GetSharedLinkMetadataArg, w io.Writer) error {
	ignoreBrokenPipeSignal()

	var bytesWritten int64
	return retryWithBackoff(func() error {
		if bytesWritten > 0 {
			return partialStdoutError(bytesWritten)
		}

		_, contents, err := dbx.GetSharedLinkFile(arg)
		if err != nil {
			return err
		}
		if contents == nil {
			return errors.New("shared link download response did not include file content")
		}
		defer func() { _ = contents.Close() }()

		n, copyErr := io.Copy(stdoutBrokenPipeWriter{w: w}, contents)
		bytesWritten += n

		if errors.Is(copyErr, errStdoutBrokenPipe) {
			return nil
		}
		if copyErr != nil && bytesWritten > 0 {
			return partialStdoutError(bytesWritten)
		}
		return copyErr
	})
}

func copySharedLinkContentToFile(contents io.Reader, size uint64, dst string, errOut io.Writer) error {
	if errOut == nil {
		errOut = io.Discard
	}

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
		DrawFunc: ioprogress.DrawTerminalf(errOut, func(progress, total int64) string {
			return fmt.Sprintf("Downloading %s/%s",
				humanize.IBytes(uint64(progress)), humanize.IBytes(uint64(total)))
		}),
		Size: int64(size),
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

func sharedLinkDownloadTarget(target string, link sharing.IsSharedLinkMetadata) (string, error) {
	name, err := sharedLinkDownloadName(link)
	if err != nil {
		return "", err
	}

	if target == "" {
		return name, nil
	}

	if info, err := os.Stat(target); err == nil && info.IsDir() {
		return filepath.Join(target, name), nil
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	return target, nil
}

func sharedLinkDownloadName(link sharing.IsSharedLinkMetadata) (string, error) {
	file, ok := link.(*sharing.FileLinkMetadata)
	if !ok {
		return "", errors.New("shared link is not a downloadable file")
	}

	name := file.Name
	if name == "" && file.PathLower != "" {
		name = path.Base(file.PathLower)
	}
	name = filepath.Base(filepath.FromSlash(name))
	if name == "" || name == "." || name == ".." || name == string(filepath.Separator) {
		return "", errors.New("shared link file metadata did not include a name")
	}
	return name, nil
}

func sharedLinkDownloadSize(link sharing.IsSharedLinkMetadata) uint64 {
	file, ok := link.(*sharing.FileLinkMetadata)
	if !ok {
		return 0
	}
	return file.Size
}

var shareLinkDownloadCmd = &cobra.Command{
	Use:   "download <url> [target]",
	Short: "Download a shared link file",
	Long: `Download a file from a Dropbox shared link.
  - If target is omitted, the local filename comes from shared-link metadata.
  - Use - as target to write file bytes to stdout.
    Stdout is byte-clean: all progress and errors go to stderr.
  - Folder-link recursive download is not supported by this command.
`,
	Example: `  dbxcli share-link download https://www.dropbox.com/s/example/file.txt
  dbxcli share-link download https://www.dropbox.com/s/example/file.txt ./local-file.txt
  dbxcli share-link download https://www.dropbox.com/s/example/file.txt - | tar tz`,
	RunE: shareLinkDownload,
}

func init() {
	shareLinkDownloadCmd.Flags().String("password", "", "Password for password-protected shared links")
	shareLinkCmd.AddCommand(shareLinkDownloadCmd)
}
