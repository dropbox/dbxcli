package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"golang.org/x/oauth2"

	"github.com/codegangsta/cli"
	"github.com/dropbox/dropbox-sdk-go"
)

const (
	configFileName = "auth.conf"
	appKey         = "7fz0ag7t20fc0nv"
	appSecret      = "1x7b0lb2mulwmrb"
)

var dbx dropbox.Api

func readToken(filePath string) (*oauth2.Token, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var tok oauth2.Token
	if json.Unmarshal(b, &tok) != nil {
		return nil, err
	}

	if !tok.Valid() {
		return nil, fmt.Errorf("Token %v is no longer valid", tok)
	}

	return &tok, nil
}

func saveToken(filePath string, token *oauth2.Token) {
	if _, err := os.Stat(filePath); err != nil {
		if !os.IsNotExist(err) {
			return
		}
		// create file
		b, err := json.Marshal(token)
		if err != nil {
			return
		}
		if err = ioutil.WriteFile(filePath, b, 0644); err != nil {
			return
		}
	}
}

func initDbx(c *cli.Context) (err error) {
	if c.GlobalString("token") != "" {
		dbx = dropbox.Client(c.GlobalString("token"), c.GlobalBool("verbose"))
		return
	}

	u, err := user.Current()
	if u == nil || err != nil {
		return
	}
	filePath := path.Join(u.HomeDir, ".config", "dbxcli", configFileName)

	conf := &oauth2.Config{
		ClientID:     appKey,
		ClientSecret: appSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.dropbox.com/1/oauth2/authorize",
			TokenURL: "https://api.dropbox.com/1/oauth2/token",
		},
	}
	token, err := readToken(filePath)
	if err != nil {
		fmt.Printf("1. Go to %v\n", conf.AuthCodeURL("state"))
		fmt.Printf("2. Click \"Allow\" (you might have to log in first).\n")
		fmt.Printf("3. Copy the authorization code.\n")
		fmt.Printf("Enter the authorization code here: ")

		var code string
		if _, err = fmt.Scan(&code); err != nil {
			return
		}
		token, err = conf.Exchange(oauth2.NoContext, code)
		if err != nil {
			return
		}
		saveToken(filePath, token)
	}

	dbx = dropbox.Client(token.AccessToken, c.GlobalBool("verbose"))

	return
}

func main() {
	app := cli.NewApp()
	app.Name = "dbxcli"
	app.Usage = "Dropbox command-line client."
	app.Before = initDbx
	app.Commands = setupCommands()
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "token", Usage: "OAuth token (bypasses OAuth flow)"},
		cli.BoolFlag{Name: "verbose", Usage: "Enable verbose logging in the SDK"},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "dbxcli: %s\n", err)
		os.Exit(1)
	}
}
