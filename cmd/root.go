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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/dropbox/dropbox-sdk-go-unofficial"
	"github.com/dropbox/dropbox-sdk-go-unofficial/files"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

const (
	configFileName = "auth.conf"
	appKey         = "7fz0ag7t20fc0nv"
	appSecret      = "1x7b0lb2mulwmrb"
	dropboxScheme  = "dropbox"
	dateFormat     = "Jan 2 15:04"
)

var (
	sizeUnits = [...]string{"B", "K", "M", "G", "T", "P", "E", "Z"}
)

func validatePath(p string) (path string, err error) {
	path = p

	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}

	path = strings.TrimSuffix(path, "/")

	return
}

func makeRelocationArg(s string, d string) (arg *files.RelocationArg, err error) {
	src, err := validatePath(s)
	if err != nil {
		return
	}
	dst, err := validatePath(d)
	if err != nil {
		return
	}

	arg = files.NewRelocationArg()
	arg.FromPath = src
	arg.ToPath = dst

	return
}

func humanizeSize(size uint64) string {
	num := float64(size)
	for _, unit := range sizeUnits {
		if math.Abs(num) < 1024.0 {
			return fmt.Sprintf("%3.1f%s", num, unit)
		}
		num /= 1024.0
	}
	return fmt.Sprintf("%.1f%s", num, "Y")
}

func humanizeDate(t time.Time) string {
	return t.Format(dateFormat)
}

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

func initDbx(cmd *cobra.Command, args []string) (err error) {
	verbose, _ := cmd.Flags().GetBool("verbose")

	if token, _ := cmd.Flags().GetString("token"); token != "" {
		dbx = dropbox.Client(token, verbose)
		return
	}

	dir, err := homedir.Dir()
	if err != nil {
		return
	}
	filePath := path.Join(dir, ".config", "dbxcli", configFileName)

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

	dbx = dropbox.Client(token.AccessToken, verbose)

	return
}

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dbxcli",
	Short: "A command line tool for Dropbox users and team admins",
	Long: `Use dbxcli to quickly interact with your Dropbox, upload/download files,
manage your team and more. It is easy, scriptable and works on all platforms!`,
	PersistentPreRunE: initDbx,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
	RootCmd.PersistentFlags().String("token", "", "Access token")
}
