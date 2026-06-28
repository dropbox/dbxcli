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

	"github.com/dropbox/dbxcli/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/ioprogress"
	"github.com/spf13/cobra"
)

type shareLinkDownloadOptions struct {
	path      string
	password  sharedLinkPasswordOptions
	recursive bool
}

type shareLinkDownloadInput struct {
	URL       string `json:"url"`
	Target    string `json:"target,omitempty"`
	Path      string `json:"path,omitempty"`
	Recursive bool   `json:"recursive,omitempty"`
	Password  bool   `json:"password,omitempty"`
}

type shareLinkDownloadResult struct {
	Target string                `json:"target"`
	Link   shareLinkJSONMetadata `json:"link"`
}

func shareLinkDownload(cmd *cobra.Command, args []string) error {
	if len(args) == 0 || len(args) > 2 {
		return invalidArgumentsErrorWithDetails("`share-link download` requires a `url` and optional `target` argument", argumentsErrorDetails("url", "target"))
	}

	url := args[0]
	if url == "" {
		return invalidArgumentsErrorWithDetails("`share-link download` requires a non-empty URL", argumentErrorDetails("url"))
	}

	target := ""
	if len(args) == 2 {
		target = args[1]
		if target == "" {
			return invalidArgumentsErrorWithDetails("`share-link download` requires a non-empty target", argumentErrorDetails("target"))
		}
	}

	arg := sharing.NewGetSharedLinkMetadataArg(url)
	opts, err := parseShareLinkDownloadOptions(cmd)
	if err != nil {
		return err
	}
	if opts.password.set {
		arg.LinkPassword = opts.password.password
	}
	if target == "-" && commandOutputFormat(cmd) == output.FormatJSON {
		return invalidArgumentsErrorWithDetails("`share-link download -` cannot be used with --output=json", mergeJSONErrorDetails(argumentErrorDetails("target"), flagErrorDetails("output")))
	}

	dbx := newSharedLinkClient(config)
	if opts.path != "" {
		arg.Path = opts.path
		return downloadSharedLinkPath(cmd, dbx, arg, target, opts)
	}

	link, err := dbx.GetSharedLinkMetadata(arg)
	if err != nil {
		return err
	}

	if folder, ok := link.(*sharing.FolderLinkMetadata); ok {
		if !opts.recursive {
			return invalidArgumentsErrorWithDetails("shared link is a folder (use --recursive to download folders)", flagErrorDetails("recursive"))
		}
		if target == "-" {
			return invalidArgumentsErrorWithDetails("cannot download shared-link folder to stdout", argumentErrorDetails("target"))
		}

		dst, err := sharedLinkFolderDownloadTarget(target, folder)
		if err != nil {
			return err
		}
		if err := downloadSharedLinkFolder(filesNewFunc(config), dbx, arg, folder.Name, dst, cmd.ErrOrStderr()); err != nil {
			return err
		}
		commandVerboseStatus(cmd, "Downloaded shared link folder to %s", dst)
		return renderShareLinkDownloadOutput(cmd, newShareLinkDownloadInput(url, target, opts), dst, folder)
	}

	if target == "-" {
		if opts.recursive {
			return invalidArgumentsErrorWithDetails("`share-link download -` cannot be used with --recursive", flagErrorDetails("recursive"))
		}
		if err := downloadSharedLinkToStdout(dbx, arg, cmd.OutOrStdout()); err != nil {
			return err
		}
		commandVerboseStatus(cmd, "Downloaded shared link to stdout")
		return nil
	}

	dst, downloaded, err := downloadSharedLinkToFile(dbx, arg, target, cmd.ErrOrStderr())
	if err != nil {
		return err
	}
	commandVerboseStatus(cmd, "Downloaded shared link to %s", dst)
	return renderShareLinkDownloadOutput(cmd, newShareLinkDownloadInput(url, target, opts), dst, downloaded)
}

func parseShareLinkDownloadOptions(cmd *cobra.Command) (shareLinkDownloadOptions, error) {
	var opts shareLinkDownloadOptions

	password, err := sharedLinkPasswordFromFlags(cmd)
	if err != nil {
		return opts, err
	}
	opts.password = password

	recursive, err := cmd.Flags().GetBool("recursive")
	if err != nil {
		return opts, err
	}
	opts.recursive = recursive

	if localFlagChanged(cmd, "path") {
		pathArg, err := localStringFlag(cmd, "path")
		if err != nil {
			return opts, err
		}
		if pathArg == "" {
			return opts, invalidArgumentsErrorWithDetails("`--path` requires a non-empty path", flagErrorDetails("path"))
		}
		dropboxPath, err := validatePath(pathArg)
		if err != nil {
			return opts, err
		}
		if dropboxPath == "" {
			return opts, invalidArgumentsErrorWithDetails("cannot download shared-link root with `--path`", pathErrorDetails("/"))
		}
		opts.path = dropboxPath
	}

	if opts.path != "" && opts.recursive {
		return opts, invalidArgumentsErrorWithDetails("`--path` cannot be used with --recursive", flagsErrorDetails("path", "recursive"))
	}

	return opts, nil
}

func downloadSharedLinkPath(cmd *cobra.Command, dbx sharedLinkClient, arg *sharing.GetSharedLinkMetadataArg, target string, opts shareLinkDownloadOptions) error {
	if target == "-" {
		if err := downloadSharedLinkToStdout(dbx, arg, cmd.OutOrStdout()); err != nil {
			return err
		}
		commandVerboseStatus(cmd, "Downloaded shared link path %s to stdout", arg.Path)
		return nil
	}

	dst, downloaded, err := downloadSharedLinkToFile(dbx, arg, target, cmd.ErrOrStderr())
	if err != nil {
		return err
	}
	commandVerboseStatus(cmd, "Downloaded shared link path %s to %s", arg.Path, dst)
	return renderShareLinkDownloadOutput(cmd, newShareLinkDownloadInput(arg.Url, target, opts), dst, downloaded)
}

func sharedLinkFolderDownloadTarget(target string, link *sharing.FolderLinkMetadata) (string, error) {
	name := link.Name
	name = filepath.Base(filepath.FromSlash(name))
	if name == "" || name == "." || name == ".." || name == string(filepath.Separator) {
		return "", errors.New("shared link folder metadata did not include a name")
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

func downloadSharedLinkFolder(filesDbx files.Client, dbx sharedLinkClient, arg *sharing.GetSharedLinkMetadataArg, rootName, dst string, errOut io.Writer) error {
	if errOut == nil {
		errOut = io.Discard
	}

	var downloadErrors []error
	queue := []string{""}

	for len(queue) > 0 {
		relFolder := queue[0]
		queue = queue[1:]

		entries, err := listSharedLinkFolderEntries(filesDbx, arg, relFolder)
		if err != nil {
			if relFolder == "" {
				return err
			}
			downloadErrors = append(downloadErrors, err)
			continue
		}
		if relFolder == "" {
			if err := os.MkdirAll(dst, 0755); err != nil {
				return err
			}
		}

		for _, entry := range entries {
			switch f := entry.(type) {
			case *files.FolderMetadata:
				relPath, err := sharedLinkEntryRelativePath(f.PathDisplay, rootName)
				if err != nil {
					downloadErrors = append(downloadErrors, err)
					continue
				}
				if relPath == "" {
					continue
				}
				localDir, err := sharedLinkLocalPath(dst, relPath)
				if err != nil {
					downloadErrors = append(downloadErrors, err)
					continue
				}
				if err := os.MkdirAll(localDir, 0755); err != nil {
					downloadErrors = append(downloadErrors, fmt.Errorf("mkdir %s: %w", localDir, err))
					continue
				}
				queue = append(queue, relPath)
			case *files.FileMetadata:
				relPath, err := sharedLinkEntryRelativePath(f.PathDisplay, rootName)
				if err != nil {
					downloadErrors = append(downloadErrors, err)
					continue
				}
				localPath, err := sharedLinkLocalPath(dst, relPath)
				if err != nil {
					downloadErrors = append(downloadErrors, err)
					continue
				}
				if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
					downloadErrors = append(downloadErrors, fmt.Errorf("mkdir %s: %w", filepath.Dir(localPath), err))
					continue
				}
				fmt.Fprintf(errOut, "Downloading %s -> %s\n", relPath, localPath)
				if err := downloadSharedLinkRelativeFile(dbx, arg, relPath, localPath, errOut); err != nil {
					downloadErrors = append(downloadErrors, fmt.Errorf("%s: %w", relPath, err))
				}
			}
		}
	}

	if len(downloadErrors) > 0 {
		for _, e := range downloadErrors {
			fmt.Fprintf(errOut, "Error: %v\n", e)
		}
		return fmt.Errorf("share-link download: %d error(s)", len(downloadErrors))
	}

	return nil
}

func listSharedLinkFolderEntries(dbx files.Client, arg *sharing.GetSharedLinkMetadataArg, relFolder string) ([]files.IsMetadata, error) {
	listArg := files.NewListFolderArg(sharedLinkAPIPath(relFolder))
	listArg.SharedLink = files.NewSharedLink(arg.Url)
	listArg.SharedLink.Password = arg.LinkPassword

	res, err := dbx.ListFolder(listArg)
	if err != nil {
		return nil, fmt.Errorf("list shared link folder %q: %v", relFolder, err)
	}

	entries := append([]files.IsMetadata{}, res.Entries...)
	for res.HasMore {
		if res.Cursor == "" {
			return entries, errors.New("list shared link folder has more results but no cursor")
		}
		cont := files.NewListFolderContinueArg(res.Cursor)
		res, err = dbx.ListFolderContinue(cont)
		if err != nil {
			return entries, fmt.Errorf("list shared link folder continue: %v", err)
		}
		entries = append(entries, res.Entries...)
	}

	return entries, nil
}

func downloadSharedLinkRelativeFile(dbx sharedLinkClient, baseArg *sharing.GetSharedLinkMetadataArg, relPath, dst string, errOut io.Writer) error {
	arg := sharing.NewGetSharedLinkMetadataArg(baseArg.Url)
	arg.Path = sharedLinkAPIPath(relPath)
	arg.LinkPassword = baseArg.LinkPassword

	return retryWithBackoff(func() error {
		link, contents, err := dbx.GetSharedLinkFile(arg)
		if err != nil {
			return err
		}
		if contents == nil {
			return errors.New("shared link download response did not include file content")
		}
		defer func() { _ = contents.Close() }()

		return copySharedLinkContentToFile(contents, sharedLinkDownloadSize(link), dst, errOut)
	})
}

func sharedLinkEntryRelativePath(pathDisplay string, rootName string) (string, error) {
	rel := strings.TrimPrefix(pathDisplay, "/")
	if rootName != "" {
		rootName = strings.Trim(rootName, "/")
		first, rest, ok := strings.Cut(rel, "/")
		if strings.EqualFold(first, rootName) {
			if !ok {
				return "", nil
			}
			rel = rest
		}
	}
	parts := strings.Split(rel, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return "", fmt.Errorf("invalid shared link entry path %q", pathDisplay)
		}
	}
	rel = path.Clean(rel)
	if rel == "" || rel == "." || rel == ".." || strings.HasPrefix(rel, "../") {
		return "", fmt.Errorf("invalid shared link entry path %q", pathDisplay)
	}
	return rel, nil
}

func sharedLinkAPIPath(rel string) string {
	if rel == "" {
		return ""
	}
	return "/" + strings.TrimPrefix(rel, "/")
}

func sharedLinkLocalPath(root, rel string) (string, error) {
	localRel := filepath.Clean(filepath.FromSlash(rel))
	if localRel == "." || localRel == ".." || strings.HasPrefix(localRel, ".."+string(filepath.Separator)) || filepath.IsAbs(localRel) {
		return "", fmt.Errorf("invalid shared link relative path %q", rel)
	}
	return filepath.Join(root, localRel), nil
}

func downloadSharedLinkToFile(dbx sharedLinkClient, arg *sharing.GetSharedLinkMetadataArg, target string, errOut io.Writer) (string, sharing.IsSharedLinkMetadata, error) {
	var dst string
	var downloaded sharing.IsSharedLinkMetadata
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
		downloaded = link

		return copySharedLinkContentToFile(contents, sharedLinkDownloadSize(link), dst, errOut)
	})
	return dst, downloaded, err
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

func newShareLinkDownloadInput(url, target string, opts shareLinkDownloadOptions) shareLinkDownloadInput {
	return shareLinkDownloadInput{
		URL:       url,
		Target:    target,
		Path:      opts.path,
		Recursive: opts.recursive,
		Password:  opts.password.set,
	}
}

func renderShareLinkDownloadOutput(cmd *cobra.Command, input shareLinkDownloadInput, target string, link sharing.IsSharedLinkMetadata) error {
	metadata, ok := shareLinkJSONMetadataFromDropbox(link)
	if !ok {
		return errors.New("found unknown shared link type")
	}

	result := shareLinkDownloadResult{
		Target: target,
		Link:   metadata,
	}
	return renderJSONOperationOutput(
		cmd,
		input,
		[]jsonOperationResult{
			newJSONOperationResult(shareLinkJSONStatusDownloaded, result.Link.Type, nil, result),
		},
	)
}

var shareLinkDownloadCmd = &cobra.Command{
	Use:   "download <url> [target]",
	Short: "Download shared link content",
	Long: `Download content from a Dropbox shared link.
  - If target is omitted, the local filename comes from shared-link metadata.
  - Use --path to download a file inside a folder shared link.
  - Use - as target to write file bytes to stdout.
    Stdout is byte-clean: all progress and errors go to stderr.
  - Use --recursive (-r) to download folder shared links.
`,
	Example: `  dbxcli share-link download https://www.dropbox.com/s/example/file.txt
  dbxcli share-link download https://www.dropbox.com/s/example/file.txt ./local-file.txt
  dbxcli share-link download https://www.dropbox.com/s/example/folder --path /nested/file.txt
  dbxcli share-link download https://www.dropbox.com/s/example/file.txt - | tar tz`,
	RunE: shareLinkDownload,
}

func init() {
	addSharedLinkPasswordFlags(shareLinkDownloadCmd)
	shareLinkDownloadCmd.Flags().String("path", "", "Download a file path inside a folder shared link")
	shareLinkDownloadCmd.Flags().BoolP("recursive", "r", false, "Recursively download a folder shared link")
	shareLinkCmd.AddCommand(shareLinkDownloadCmd)
	enableStructuredOutput(shareLinkDownloadCmd)
}
