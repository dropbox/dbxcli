package main

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/codegangsta/cli"
	"github.com/dropbox/dropbox-sdk-go/files"
)

func printFolderMetadata(e *files.FolderMetadata) {
	fmt.Printf("-\t\t-\t-\t\t%s\n", e.Name)
}

func printFileMetadata(e *files.FileMetadata) {
	fmt.Printf("%s\t%s\t%s\t%s\n", e.Rev, humanizeSize(e.Size), humanizeDate(e.ServerModified), e.Name)
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

	fmt.Printf("Revision\tSize\tLast modified\tPath\n")

	for _, e := range res.Entries {
		switch e.Tag {
		case "folder":
			printFolderMetadata(e.Folder)
		case "file":
			printFileMetadata(e.File)
		}
	}

	return
}

func Put(ctx *cli.Context) (err error) {
	src := ctx.Args().Get(0)
	dst, err := parseDropboxUri(ctx.Args().Get(1))
	if err != nil {
		return
	}

	args := files.NewCommitInfo()
	args.Path = dst
	args.Mode.Tag = "overwrite"

	f, err := os.Open(src)
	if err != nil {
		return
	}

	if _, err = dbx.Upload(args, f); err != nil {
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

	_, contents, err := dbx.Download(args)
	defer contents.Close()
	if err != nil {
		return
	}

	f, err := os.Create(dst)
	defer f.Close()
	if err != nil {
		return
	}

	if _, err = io.Copy(f, contents); err != nil {
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

	fmt.Printf("Revision\tSize\tLast modified\tPath\n")
	for _, e := range res.Entries {
		printFileMetadata(e)
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

	fmt.Printf("Revision\tSize\tLast modified\n")

	for _, m := range res.Matches {
		e := m.Metadata
		switch e.Tag {
		case "folder":
			printFolderMetadata(e.Folder)
		case "file":
			printFileMetadata(e.File)
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
		Name:  "l",
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
			After:     after,
			Usage:     "",
			ArgsUsage: "",
		},
	}
}
