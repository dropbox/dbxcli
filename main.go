package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/codegangsta/cli"
	"github.com/dropbox/dropbox-sdk-go/dropbox"
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
	if dbx, err = dropbox.OauthClient(appKey, appSecret, filePath); err != nil {
		return err
	}
	return nil
}

func ls(c *cli.Context) {
	path := c.Args().First()
	arg := dropbox.NewListFolderArg()
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

func put(c *cli.Context) {
	src := c.Args().Get(0)
	dst := c.Args().Get(1)
	f, _ := os.Open(src)
	arg := dropbox.NewCommitInfo()
	arg.Path = dst
	arg.Mode.Tag = "overwrite"
	res, err := dbx.Upload(arg, f)
	fmt.Printf("%v %v\n", res, err)
}

func get(c *cli.Context) {
	path := c.Args().First()
	arg := &dropbox.DownloadArg{Path: path}
	res, content, err := dbx.Download(arg)
	defer content.Close()
	buf, _ := ioutil.ReadAll(content)
	fmt.Printf("%v %s %v\n", res, buf, err)
}

func main() {
	app := cli.NewApp()
	app.Name = "dbxcli"
	app.Before = initDbx
	app.Commands = []cli.Command{
		{Name: "get_space_usage", Aliases: []string{"df"}, Action: df},
		{Name: "ls", Action: ls, Flags: []cli.Flag{
			cli.BoolFlag{Name: "l"},
		}},
		{Name: "get", Action: get},
		{Name: "put", Action: put},
	}
	app.Run(os.Args)
}
