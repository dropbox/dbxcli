package main

import (
	"fmt"
	"io"
	"os"

	"github.com/codegangsta/cli"
	"github.com/dropbox/dropbox-sdk-go/files"
)

type ActionFunc func(ctx *cli.Context) error

func NewCommand(name string, action ActionFunc) cli.Command {
	var err error

	return cli.Command{
		Action: func(ctx *cli.Context) {
			err = action(ctx)
		},
		After: func(ctx *cli.Context) error {
			return err
		},
		Name: name,
	}
}

func Ls(ctx *cli.Context) (err error) {
	path := ""

	if ctx.Args().Present() {
		path, err = parseDropboxUri(ctx.Args().First())
		if err != nil {
			return
		}
	}

	args := files.NewListFolderArg()
	args.Path = path

	res, err := dbx.ListFolder(args)
	if err != nil {
		return
	}

	for _, e := range res.Entries {
		switch e.Tag {
		case "folder":
			fmt.Printf("%s/\n", e.Folder.Name)
		case "file":
			fmt.Printf("%s\n", e.File.Name)
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

	_, err = dbx.Upload(args, f)
	if err != nil {
		return
	}

	return
}

func Get(ctx *cli.Context) (err error) {
	src, err := parseDropboxUri(ctx.Args().Get(0))
	dst := ctx.Args().Get(1)

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

	_, err = io.Copy(f, contents)
	if err != nil {
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

	_, err = dbx.Delete(args)
	if err != nil {
		return
	}

	return
}

func Cp(ctx *cli.Context) (err error) {
	src, err := parseDropboxUri(ctx.Args().Get(0))
	if err != nil {
		return
	}
	dst, err := parseDropboxUri(ctx.Args().Get(1))
	if err != nil {
		return
	}

	args := files.NewRelocationArg()
	args.FromPath = src
	args.ToPath = dst

	_, err = dbx.Copy(args)
	if err != nil {
		return
	}

	return
}

func Mv(ctx *cli.Context) (err error) {
	src, err := parseDropboxUri(ctx.Args().Get(0))
	if err != nil {
		return
	}
	dst, err := parseDropboxUri(ctx.Args().Get(1))
	if err != nil {
		return
	}

	args := files.NewRelocationArg()
	args.FromPath = src
	args.ToPath = dst

	_, err = dbx.Move(args)
	if err != nil {
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

	fmt.Printf("Revision\tModification time\t\tSize\n")
	for _, e := range res.Entries {
		fmt.Printf("%s\t%v\t%v\n", e.Rev, e.ServerModified, e.Size)
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

	_, err = dbx.Restore(args)
	if err != nil {
		return
	}

	return
}

func setupCommands() []cli.Command {
	return []cli.Command{
		NewCommand("ls", Ls),
		NewCommand("get", Get),
		NewCommand("put", Put),
		NewCommand("rm", Rm),
		NewCommand("cp", Cp),
		NewCommand("mv", Mv),
		NewCommand("revs", Revs),
		NewCommand("restore", Restore),
	}
}
