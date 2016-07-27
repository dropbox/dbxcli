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

	"golang.org/x/oauth2"
)

const (
	apiVersion    = 2
	defaultDomain = ".dropboxapi.com"
	hostAPI       = "api"
	hostContent   = "content"
	hostNotify    = "notify"
)

// Config contains parameters for configuring the SDK.
type Config struct {
	Token      string
	Verbose    bool
	AsMemberID string
	Domain     string
}

// Context is the base client context used to implement per-namespace clients.
type Context struct {
	Client  *http.Client
	Config  Config
	hostMap map[string]string
}

// GenerateURL returns the appropriate URL for given namespace/route.
func (c *Context) GenerateURL(host string, namespace string, route string) string {
	fqHost := c.hostMap[host]
	return fmt.Sprintf("https://%s/%d/%s/%s", fqHost, apiVersion, namespace, route)
}

// NewContext returns a new Context with the given Config.
func NewContext(c Config) Context {
	domain := c.Domain
	if domain == "" {
		domain = defaultDomain
	}

	hostMap := map[string]string{
		hostAPI:     hostAPI + domain,
		hostContent: hostContent + domain,
		hostNotify:  hostNotify + domain,
	}
	var conf = &oauth2.Config{Endpoint: OAuthEndpoint(domain)}
	tok := &oauth2.Token{AccessToken: c.Token}
	return Context{conf.Client(oauth2.NoContext, tok), c, hostMap}
}

// OAuthEndpoint constructs an `oauth2.Endpoint` for the given domain
func OAuthEndpoint(domain string) oauth2.Endpoint {
	if domain == "" {
		domain = defaultDomain
	}
	authURL := fmt.Sprintf("https://meta%s/1/oauth2/authorize", domain)
	tokenURL := fmt.Sprintf("https://api%s/1/oauth2/token", domain)
	if domain == defaultDomain {
		authURL = "https://www.dropbox.com/1/oauth2/authorize"
	}
	return oauth2.Endpoint{AuthURL: authURL, TokenURL: tokenURL}
}

// Tagged is used for tagged unions.
type Tagged struct {
	Tag string `json:".tag"`
}

// APIError is the base type for endpoint-specific errors.
type APIError struct {
	ErrorSummary string `json:"error_summary"`
}

func (e APIError) Error() string {
	return e.ErrorSummary
}

func init() {
	// These are not registered in the oauth library by default
	oauth2.RegisterBrokenAuthHeaderProvider("https://api.dropboxapi.com")
	oauth2.RegisterBrokenAuthHeaderProvider("https://api-dbdev.dev.corp.dropbox.com")
}
