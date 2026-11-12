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

	"github.com/dropbox/dbxcli/v3/internal/output"
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

func shareLinkDownload(cmd *cobra.Command, args []string) (retErr error) {
	if len(args) == 0 || len(args) > 2 {
		return invalidArgumentsErrorWithDetails("`share-link download` requires a `url` and optional `target` argument", argumentsErrorDetails("url", "target"))
	}

	url := args[0]
	if url == "" {
		return invalidArgumentsErrorWithDetails("`share-link download` requires a non-empty URL", mergeJSONErrorDetails(argumentErrorDetails("url"), urlErrorDetails(url)))
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
		return withJSONErrorDetails(err, urlErrorDetails(url), operationErrorDetails("share_link_download"))
	}
	defer func() {
		retErr = withJSONErrorDetails(retErr, urlErrorDetails(url), operationErrorDetails("share_link_download"))
		if opts.path != "" {
			flags := []string{"path"}
			if opts.recursive {
				flags = append(flags, "recursive")
			}
			retErr = withJSONErrorDetails(retErr, pathErrorDetails(opts.path), flagsErrorDetails(flags...))
		}
	}()
	if opts.password.set {
		arg.LinkPassword = opts.password.password
	}
	if target == "-" && commandOutputFormat(cmd) == output.FormatJSON {
		return invalidArgumentsErrorWithDetails("`share-link download -` cannot be used with --output=json", mergeJSONErrorDetails(operationErrorDetails("share_link_download"), argumentErrorDetails("target"), flagErrorDetails("output"), urlErrorDetails(url)))
	}

	dbx := newSharedLinkClient(config)
	if opts.path != "" {
		arg.Path = opts.path
	}

	if err := currentContext().Err(); err != nil {
		return err
	}
	link, err := dbx.GetSharedLinkMetadataContext(currentContext(), arg)
	if err != nil {
		return err
	}

	if folder, ok := link.(*sharing.FolderLinkMetadata); ok && folder != nil {
		if !opts.recursive {
			return invalidArgumentsErrorWithDetails("shared link is a folder (use --recursive to download folders)", mergeJSONErrorDetails(operationErrorDetails("share_link_download"), flagErrorDetails("recursive"), urlErrorDetails(url)))
		}
		if target == "-" {
			return invalidArgumentsErrorWithDetails("cannot download shared-link folder to stdout", mergeJSONErrorDetails(operationErrorDetails("share_link_download"), argumentErrorDetails("target"), urlErrorDetails(url)))
		}

		dst, err := sharedLinkFolderDownloadTarget(target, folder)
		if err != nil {
			return err
		}
		if err := downloadSharedLinkFolder(filesNewFunc(config), dbx, arg, dst, cmd.ErrOrStderr()); err != nil {
			return err
		}
		commandVerboseStatus(cmd, "Downloaded shared link folder to %s", dst)
		return renderShareLinkDownloadOutput(cmd, newShareLinkDownloadInput(url, target, opts), dst, folder)
	}

	if opts.path != "" {
		if file, ok := link.(*sharing.FileLinkMetadata); !ok || file == nil {
			return errors.New("shared link path is not a downloadable file or folder")
		}
	}

	if target == "-" {
		if opts.recursive {
			return invalidArgumentsErrorWithDetails("`share-link download -` cannot be used with --recursive", mergeJSONErrorDetails(operationErrorDetails("share_link_download"), flagErrorDetails("recursive"), urlErrorDetails(url)))
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
		// Reject traversal before validatePath can clean it away, as share-link ls does.
		for _, name := range strings.Split(pathArg, "/") {
			if name == "" || name == "." {
				continue
			}
			if _, err := sharedLinkEntryRelativePath("", name); err != nil {
				return opts, invalidArgumentsErrorWithDetails(err.Error(), mergeJSONErrorDetails(flagErrorDetails("path"), pathErrorDetails(pathArg)))
			}
		}
		dropboxPath, err := validatePath(pathArg)
		if err != nil {
			return opts, err
		}
		if dropboxPath == "" {
			return opts, invalidArgumentsErrorWithDetails("cannot download shared-link root with `--path`", mergeJSONErrorDetails(operationErrorDetails("share_link_download"), pathErrorDetails("/")))
		}
		opts.path = dropboxPath
	}

	return opts, nil
}

func sharedLinkFolderDownloadTarget(target string, link *sharing.FolderLinkMetadata) (string, error) {
	name := link.Name
	if _, err := sharedLinkEntryRelativePath("", name); err != nil {
		return "", fmt.Errorf("invalid shared link folder name: %w", err)
	}
	if _, err := sharedLinkLocalPath(".", name); err != nil {
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

func downloadSharedLinkFolder(filesDbx filesClient, dbx sharedLinkClient, arg *sharing.GetSharedLinkMetadataArg, dst string, errOut io.Writer) error {
	if errOut == nil {
		errOut = io.Discard
	}

	// Queue and local paths are relative to the selected folder. API paths
	// remain relative to the original shared-link root, even for unmounted links.
	start := strings.TrimPrefix(arg.Path, "/")
	var downloadErrors []error
	queue := []string{""}

	for len(queue) > 0 {
		if err := currentContext().Err(); err != nil {
			return err
		}
		relFolder := queue[0]
		queue = queue[1:]

		apiFolder := relFolder
		if start != "" {
			apiFolder = path.Join(start, relFolder)
		}
		entries, err := listSharedLinkFolderEntries(filesDbx, arg, apiFolder)
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
			if err := currentContext().Err(); err != nil {
				return err
			}
			switch f := entry.(type) {
			case *files.FolderMetadata:
				if f == nil {
					downloadErrors = append(downloadErrors, errors.New("shared link listing returned nil folder metadata"))
					continue
				}
				relPath, err := sharedLinkEntryRelativePath(relFolder, f.Name)
				if err != nil {
					downloadErrors = append(downloadErrors, err)
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
				if f == nil {
					downloadErrors = append(downloadErrors, errors.New("shared link listing returned nil file metadata"))
					continue
				}
				relPath, err := sharedLinkEntryRelativePath(relFolder, f.Name)
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
				apiPath := relPath
				if start != "" {
					apiPath = path.Join(start, relPath)
				}
				if err := downloadSharedLinkRelativeFile(dbx, arg, apiPath, localPath, errOut); err != nil {
					downloadErrors = append(downloadErrors, fmt.Errorf("%s: %w", relPath, err))
				}
			default:
				downloadErrors = append(downloadErrors, fmt.Errorf("unknown shared link entry type in %q", sharedLinkAPIPath(apiFolder)))
			}
		}
	}

	if len(downloadErrors) > 0 {
		for _, e := range downloadErrors {
			fmt.Fprintf(errOut, "Error: %v\n", e)
		}
		return fmt.Errorf("share-link download: %d error(s): %w", len(downloadErrors), errors.Join(downloadErrors...))
	}

	return currentContext().Err()
}

func listSharedLinkFolderEntries(dbx filesClient, arg *sharing.GetSharedLinkMetadataArg, relFolder string) ([]files.IsMetadata, error) {
	listArg := files.NewListFolderArg(sharedLinkAPIPath(relFolder))
	listArg.SharedLink = files.NewSharedLink(arg.Url)
	listArg.SharedLink.Password = arg.LinkPassword

	if err := currentContext().Err(); err != nil {
		return nil, err
	}
	res, err := dbx.ListFolderContext(currentContext(), listArg)
	if err != nil {
		return nil, fmt.Errorf("list shared link folder %q: %w", relFolder, err)
	}

	entries := append([]files.IsMetadata{}, res.Entries...)
	for res.HasMore {
		if err := currentContext().Err(); err != nil {
			return entries, err
		}
		if res.Cursor == "" {
			return entries, errors.New("list shared link folder has more results but no cursor")
		}
		cont := files.NewListFolderContinueArg(res.Cursor)
		res, err = dbx.ListFolderContinueContext(currentContext(), cont)
		if err != nil {
			return entries, fmt.Errorf("list shared link folder continue: %w", err)
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
		link, contents, err := dbx.GetSharedLinkFileContext(currentContext(), arg)
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

// Shared-link listings may omit path_display and path_lower when the folder is
// not mounted in the caller's Dropbox. Names are relative to the folder listed.
// Validate Dropbox names independently of the local OS; Windows device names
// and colons are not filesystem paths until sharedLinkLocalPath is called.
func sharedLinkEntryRelativePath(parentRel, name string) (string, error) {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00") {
		return "", fmt.Errorf("invalid shared link entry name %q", name)
	}
	return path.Join(parentRel, name), nil
}

func sharedLinkAPIPath(rel string) string {
	if rel == "" {
		return ""
	}
	return "/" + strings.TrimPrefix(rel, "/")
}

func sharedLinkLocalPath(root, rel string) (string, error) {
	localRel := filepath.Clean(filepath.FromSlash(rel))
	if localRel == "." || !filepath.IsLocal(localRel) {
		return "", fmt.Errorf("invalid shared link relative path %q", rel)
	}
	return filepath.Join(root, localRel), nil
}

func downloadSharedLinkToFile(dbx sharedLinkClient, arg *sharing.GetSharedLinkMetadataArg, target string, errOut io.Writer) (string, sharing.IsSharedLinkMetadata, error) {
	var dst string
	var downloaded sharing.IsSharedLinkMetadata
	err := retryWithBackoff(func() error {
		link, contents, err := dbx.GetSharedLinkFileContext(currentContext(), arg)
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

		_, contents, err := dbx.GetSharedLinkFileContext(currentContext(), arg)
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
  - Use --path to select a file or folder inside a folder shared link.
    Selected folders require --recursive; files download with or without it.
  - Use - as target to write file bytes to stdout.
    Stdout is byte-clean: all progress and errors go to stderr.
  - Use --recursive (-r) to download folder shared links or selected subfolders.
    Existing target directories receive a subdirectory named after the selected folder.
    Otherwise, target is used directly as the destination directory.
`,
	Example: `  dbxcli share-link download https://www.dropbox.com/s/example/file.txt
  dbxcli share-link download https://www.dropbox.com/s/example/file.txt ./local-file.txt
  dbxcli share-link download https://www.dropbox.com/s/example/folder --path /nested/file.txt
  dbxcli share-link download https://www.dropbox.com/s/example/folder ./out --path /2027/Logoset --recursive
  dbxcli share-link download https://www.dropbox.com/s/example/file.txt - | tar tz`,
	RunE: shareLinkDownload,
}

func init() {
	addSharedLinkPasswordFlags(shareLinkDownloadCmd)
	shareLinkDownloadCmd.Flags().String("path", "", "Select a file or folder path inside a folder shared link")
	shareLinkDownloadCmd.Flags().BoolP("recursive", "r", false, "Recursively download a folder shared link")
	shareLinkCmd.AddCommand(shareLinkDownloadCmd)
	enableStructuredOutput(shareLinkDownloadCmd)
}
