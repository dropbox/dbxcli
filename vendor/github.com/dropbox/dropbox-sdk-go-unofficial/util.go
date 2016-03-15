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
	hostApi       = "api"
	hostContent   = "content"
	hostNotify    = "notify"
)

type apiImpl struct {
	client  *http.Client
	verbose bool
	hostMap map[string]string
}

func getenv(key string, defVal string) string {
	val := os.Getenv(key)
	if val == "" {
		val = defVal
	}
	return val
}

func (dbx *apiImpl) generateUrl(host string, namespace string, route string) string {
	fqHost := dbx.hostMap[host]
	return fmt.Sprintf("https://%s/%d/%s/%s", fqHost, apiVersion, namespace, route)
}

func Client(token string, verbose bool) Api {
	var conf = &oauth2.Config{
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.dropbox.com/1/oauth2/authorize",
			TokenURL: "https://api.dropbox.com/1/oauth2/token",
		},
	}
	tok := &oauth2.Token{AccessToken: token}
	domain := getenv("DROPBOX_DOMAIN", defaultDomain)
	hostMap := map[string]string{
		hostApi:     hostApi + domain,
		hostContent: hostContent + domain,
		hostNotify:  hostNotify + domain,
	}
	return &apiImpl{conf.Client(oauth2.NoContext, tok), verbose, hostMap}
}
