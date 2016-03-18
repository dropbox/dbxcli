// Copyright (c) Dropbox, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package dropbox

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

const (
	apiVersion    = 2
	defaultDomain = ".dropboxapi.com"
	hostAPI       = "api"
	hostContent   = "content"
	hostNotify    = "notify"
)

type Options struct {
	Verbose    bool
	AsMemberId string
}

type apiImpl struct {
	client  *http.Client
	options Options
	hostMap map[string]string
}

func getenv(key string, defVal string) string {
	val := os.Getenv(key)
	if val == "" {
		val = defVal
	}
	return val
}

func (dbx *apiImpl) generateURL(host string, namespace string, route string) string {
	fqHost := dbx.hostMap[host]
	return fmt.Sprintf("https://%s/%d/%s/%s", fqHost, apiVersion, namespace, route)
}

// Client returns an `Api` instance for Dropbox using the given OAuth token.
func Client(token string, options Options) Api {
	domain := getenv("DROPBOX_DOMAIN", defaultDomain)
	hostMap := map[string]string{
		hostAPI:     hostAPI + domain,
		hostContent: hostContent + domain,
		hostNotify:  hostNotify + domain,
	}
	authDomain := getenv("DROPBOX_DOMAIN", ".dropbox.com")
	authUrl := fmt.Sprintf("https://www%s/1/oauth2/authorize", authDomain)
	tokenUrl := fmt.Sprintf("https://api%s/1/oauth2/token", authDomain)
	var conf = &oauth2.Config{
		Endpoint: oauth2.Endpoint{
			AuthURL:  authUrl,
			TokenURL: tokenUrl,
		},
	}
	tok := &oauth2.Token{AccessToken: token}
	return &apiImpl{conf.Client(oauth2.NoContext, tok), options, hostMap}
}
