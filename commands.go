package main

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/codegangsta/cli"
	"github.com/dropbox/dropbox-sdk-go/files"
	"github.com/mitchellh/ioprogress"
)

const (
	longFormatFlag = "l"
	// 16 MB chunks
	chunkSize int64 = 1 << 24
)

func printFolderMetadata(e *files.FolderMetadata, longFormat bool) {
	if longFormat {
		fmt.Printf("-\t\t-\t-\t\t")
	}
	fmt.Printf("%s\n", e.Name)
}

func printFileMetadata(e *files.FileMetadata, longFormat bool) {
	if longFormat {
		fmt.Printf("%s\t%s\t%s\t", e.Rev, humanizeSize(e.Size), humanizeDate(e.ServerModified))
	}
	fmt.Printf("%s\n", e.Name)
}

func Ls(ctx *cli.Context) (err error) {
	path := ""

	if ctx.Args().Present() {
		if path, err = parseDropboxUri(ctx.Args().First()); err != nil {
			return
		}
	}

	args := files.NewListFolderArg()
	args.Path = path

	res, err := dbx.ListFolder(args)
	if err != nil {
		return
	}

	if ctx.Bool(longFormatFlag) {
		fmt.Printf("Revision\tSize\tLast modified\tPath\n")
	}

	for _, e := range res.Entries {
		switch e.Tag {
		case "folder":
			printFolderMetadata(e.Folder, ctx.Bool(longFormatFlag))
		case "file":
			printFileMetadata(e.File, ctx.Bool(longFormatFlag))
		}
	}

	return
}

func uploadChunked(r io.Reader, commitInfo *files.CommitInfo, sizeTotal int64) (err error) {
	res, err := dbx.UploadSessionStart(&io.LimitedReader{R: r, N: chunkSize})
	if err != nil {
		return
	}

	written := chunkSize

	for (sizeTotal - written) > chunkSize {
		args := files.NewUploadSessionCursor()
		args.SessionId = res.SessionId
		args.Offset = uint64(written)

		err = dbx.UploadSessionAppend(args, &io.LimitedReader{R: r, N: chunkSize})
		if err != nil {
			return
		}
		written += chunkSize
	}

	args := files.NewUploadSessionFinishArg()
	args.Cursor = files.NewUploadSessionCursor()
	args.Cursor.SessionId = res.SessionId
	args.Cursor.Offset = uint64(written)
	args.Commit = commitInfo

	if _, err = dbx.UploadSessionFinish(args, r); err != nil {
		return
	}

	return
}

func Put(ctx *cli.Context) (err error) {
	src := ctx.Args().Get(0)
	dst, err := parseDropboxUri(ctx.Args().Get(1))
	if err != nil {
		return
	}

	contents, err := os.Open(src)
	defer contents.Close()
	if err != nil {
		return
	}

	contentsInfo, err := contents.Stat()
	if err != nil {
		return
	}

	progressbar := &ioprogress.Reader{
		Reader: contents,
		DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
			return fmt.Sprintf("Uploading %s/%s",
				humanizeSize(uint64(progress)),
				humanizeSize(uint64(total)))
		}),
		Size: contentsInfo.Size(),
	}

	commitInfo := files.NewCommitInfo()
	commitInfo.Path = dst
	commitInfo.Mode.Tag = "overwrite"

	if contentsInfo.Size() > chunkSize {
		return uploadChunked(progressbar, commitInfo, contentsInfo.Size())
	}

	if _, err = dbx.Upload(commitInfo, progressbar); err != nil {
		return
	}

	return
}

func Get(ctx *cli.Context) (err error) {
	src, err := parseDropboxUri(ctx.Args().Get(0))
	dst := ctx.Args().Get(1)

	if dst == "" {
		dst = path.Base(src)
	}

	args := files.NewDownloadArg()
	args.Path = src

	res, contents, err := dbx.Download(args)
	defer contents.Close()
	if err != nil {
		return
	}

	f, err := os.Create(dst)
	defer f.Close()
	if err != nil {
		return
	}

	progressbar := &ioprogress.Reader{
		Reader: contents,
		DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
			return fmt.Sprintf("Downloading %s/%s",
				humanizeSize(uint64(progress)),
				humanizeSize(uint64(total)))
		}),
		Size: int64(res.Size),
	}

	if _, err = io.Copy(f, progressbar); err != nil {
		return
	}

	return
}

func Rm(ctx *cli.Context) (err error) {
	path, err := parseDropboxUri(ctx.Args().First())
	if err != nil {
		return
	}

	args := files.NewDeleteArg()
	args.Path = path

	if _, err = dbx.Delete(args); err != nil {
		return
	}

	return
}

func parseRelocationArgs(ctx *cli.Context) (args *files.RelocationArg, err error) {
	src, err := parseDropboxUri(ctx.Args().Get(0))
	if err != nil {
		return
	}
	dst, err := parseDropboxUri(ctx.Args().Get(1))
	if err != nil {
		return
	}

	args = files.NewRelocationArg()
	args.FromPath = src
	args.ToPath = dst

	return
}

func Cp(ctx *cli.Context) (err error) {
	args, err := parseRelocationArgs(ctx)
	if err != nil {
		return
	}

	if _, err = dbx.Copy(args); err != nil {
		return
	}

	return
}

func Mv(ctx *cli.Context) (err error) {
	args, err := parseRelocationArgs(ctx)

	if _, err = dbx.Move(args); err != nil {
		return
	}

	return
}

func Revs(ctx *cli.Context) (err error) {
	path, err := parseDropboxUri(ctx.Args().First())
	if err != nil {
		return
	}

	args := files.NewListRevisionsArg()
	args.Path = path

	res, err := dbx.ListRevisions(args)
	if err != nil {
		return
	}

	if ctx.Bool(longFormatFlag) {
		fmt.Printf("Revision\tSize\tLast modified\tPath\n")
	}

	for _, e := range res.Entries {
		if ctx.Bool(longFormatFlag) {
			printFileMetadata(e, true)
		} else {
			fmt.Printf("%s\n", e.Rev)
		}
	}

	return
}

func Restore(ctx *cli.Context) (err error) {
	path, err := parseDropboxUri(ctx.Args().Get(0))
	if err != nil {
		return
	}

	rev := ctx.Args().Get(1)

	args := files.NewRestoreArg()
	args.Path = path
	args.Rev = rev

	if _, err = dbx.Restore(args); err != nil {
		return
	}

	return
}

func Mkdir(ctx *cli.Context) (err error) {
	dst, err := parseDropboxUri(ctx.Args().First())
	if err != nil {
		return
	}

	args := files.NewCreateFolderArg()
	args.Path = dst

	if _, err = dbx.CreateFolder(args); err != nil {
		return
	}

	return
}

func Find(ctx *cli.Context) (err error) {
	args := files.NewSearchArg()
	args.Query = ctx.Args().First()

	res, err := dbx.Search(args)
	if err != nil {
		return
	}

	if ctx.Bool(longFormatFlag) {
		fmt.Printf("Revision\tSize\tLast modified\tPath\n")
	}

	for _, m := range res.Matches {
		e := m.Metadata
		switch e.Tag {
		case "folder":
			printFolderMetadata(e.Folder, ctx.Bool(longFormatFlag))
		case "file":
			printFileMetadata(e.File, ctx.Bool(longFormatFlag))
		}
	}

	return
}

func Du(ctx *cli.Context) (err error) {
	usage, err := dbx.GetSpaceUsage()
	if err != nil {
		return
	}

	fmt.Printf("Used: %s\n", humanizeSize(usage.Used))
	fmt.Printf("Type: %s\n", usage.Allocation.Tag)

	allocation := usage.Allocation

	switch allocation.Tag {
	case "individual":
		fmt.Printf("Allocated: %s\n", humanizeSize(allocation.Individual.Allocated))
	case "team":
		fmt.Printf("Allocated: %s (Used: %s)\n",
			humanizeSize(allocation.Team.Allocated),
			humanizeSize(allocation.Team.Used))
	}

	return
}

func setupCommands() []cli.Command {
	var err error

	after := func(ctx *cli.Context) error {
		return err
	}

	longFormat := cli.BoolFlag{
		Name:  longFormatFlag,
		Usage: "use a long listing format",
	}

	return []cli.Command{
		{
			Name: "ls",
			Action: func(ctx *cli.Context) {
				err = Ls(ctx)
			},
			After:     after,
			Usage:     "List files in folder",
			ArgsUsage: "[dropbox://PATH]",
			Flags:     []cli.Flag{longFormat},
			HideHelp:  true,
		},
		{
			Name: "get",
			Action: func(ctx *cli.Context) {
				err = Get(ctx)
			},
			After:     after,
			Usage:     "Get file from Dropbox",
			ArgsUsage: "dropbox://SRC [DEST]",
		},
		{
			Name: "put",
			Action: func(ctx *cli.Context) {
				err = Put(ctx)
			},
			After:     after,
			Usage:     "Put file into Dropbox",
			ArgsUsage: "SRC dropbox://DEST",
		},
		{
			Name: "cp",
			Action: func(ctx *cli.Context) {
				err = Cp(ctx)
			},
			After:     after,
			Usage:     "Copy file",
			ArgsUsage: "dropbox://SRC dropbox://DEST",
		},
		{
			Name: "mv",
			Action: func(ctx *cli.Context) {
				err = Mv(ctx)
			},
			After:     after,
			Usage:     "Move file",
			ArgsUsage: "dropbox://SRC dropbox://DEST",
		},
		{
			Name: "rm",
			Action: func(ctx *cli.Context) {
				err = Rm(ctx)
			},
			After:     after,
			Usage:     "Remove file",
			ArgsUsage: "dropbox://PATH",
		},
		{
			Name: "mkdir",
			Action: func(ctx *cli.Context) {
				err = Mkdir(ctx)
			},
			After:     after,
			Usage:     "Create directory",
			ArgsUsage: "dropbox://PATH",
		},
		{
			Name: "find",
			Action: func(ctx *cli.Context) {
				err = Find(ctx)
			},
			After:     after,
			ArgsUsage: "PATTERN",
			Flags:     []cli.Flag{longFormat},
		},
		{
			Name: "revs",
			Action: func(ctx *cli.Context) {
				err = Revs(ctx)
			},
			After:     after,
			Usage:     "List file revisions",
			ArgsUsage: "dropbox://PATH",
			Flags:     []cli.Flag{longFormat},
		},
		{
			Name: "restore",
			Action: func(ctx *cli.Context) {
				err = Restore(ctx)
			},
			After:     after,
			Usage:     "Restore file",
			ArgsUsage: "dropbox://PATH REVISION",
		},
		{
			Name: "du",
			Action: func(ctx *cli.Context) {
				err = Du(ctx)
			},
			After: after,
			Usage: "Print Dropbox space usage",
		},
	}
}
