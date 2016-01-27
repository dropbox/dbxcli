package main

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/codegangsta/cli"
	"github.com/dropbox/dropbox-sdk-go"
	"github.com/dropbox/dropbox-sdk-go/files"
)

const (
	_          = iota // ignore first value by assigning to blank identifier
	KB float64 = 1 << (10 * iota)
	MB
	GB
	configFileName = "auth.conf"
	appKey         = "7fz0ag7t20fc0nv"
	appSecret      = "1x7b0lb2mulwmrb"
)

var dbx dropbox.Api

func initDbx(c *cli.Context) error {
	u, err := user.Current()
	if u == nil || err != nil {
		return err
	}
	filePath := path.Join(u.HomeDir, ".config", "dbxcli", configFileName)
	if dbx, err = dropbox.OauthClient(appKey, appSecret, filePath, false); err != nil {
		return err
	}
	return nil
}

func ls(c *cli.Context) {
	path := c.Args().First()
	arg := files.NewListFolderArg()
	arg.Path = path
	res, err := dbx.ListFolder(arg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	long := c.Bool("l")
	for _, e := range res.Entries {
		switch e.Tag {
		case "folder":
			fmt.Printf("%s/\n", e.Folder.Name)
		case "file":
			if long {
				fmt.Printf("%s\t%v\t%v\n", e.File.Name, e.File.ServerModified, e.File.Size)
			} else {
				fmt.Printf("%s\n", e.File.Name)
			}
		}
	}
}

func df(c *cli.Context) {
	usage, err := dbx.GetSpaceUsage()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Used: %.2f MB\n", float64(usage.Used)/MB)
	fmt.Printf("Type: %s\n", usage.Allocation.Tag)
	allocation := usage.Allocation
	switch allocation.Tag {
	case "individual":
		fmt.Printf("Allocated: %.2f MB\n", float64(allocation.Individual.Allocated)/MB)
	case "team":
		fmt.Printf("Allocated: %.2f MB (Used: %.2f)\n",
			float64(allocation.Team.Allocated)/MB,
			float64(allocation.Team.Used)/MB)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "dbxcli"
	app.Before = initDbx
	app.Commands = setupCommands()
	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}
