package main

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/codegangsta/cli"
	"github.com/dropbox/dropbox-sdk-go"
)

const (
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

func main() {
	app := cli.NewApp()
	app.Name = "dbxcli"
	app.Before = initDbx
	app.Commands = setupCommands()
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dbxcli: %s\n", err)
		os.Exit(1)
	}
}
