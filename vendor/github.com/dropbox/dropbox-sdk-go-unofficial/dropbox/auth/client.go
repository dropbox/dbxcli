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

package auth

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
)

// Client interface describes all routes in this namespace
type Client interface {
	// TokenRevoke : Disables the access token used to authenticate the call.
	TokenRevoke() (err error)
}

type apiImpl dropbox.Context

//TokenRevokeAPIError is an error-wrapper for the token/revoke route
type TokenRevokeAPIError struct {
	dropbox.APIError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) TokenRevoke() (err error) {
	cli := dbx.Client

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "auth", "token/revoke"), nil)
	if err != nil {
		return
	}

	if dbx.Config.AsMemberID != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.Config.AsMemberID)
	}
	if dbx.Config.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.Config.Verbose {
		log.Printf("resp: %v", resp)
	}
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if dbx.Config.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode == http.StatusOK {
		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError TokenRevokeAPIError
		err = json.Unmarshal(body, &apiError)
		if err != nil {
			return
		}
		err = apiError
		return
	}
	var apiError dropbox.APIError
	if resp.StatusCode == http.StatusBadRequest {
		apiError.ErrorSummary = string(body)
		err = apiError
		return
	}
	err = json.Unmarshal(body, &apiError)
	if err != nil {
		return
	}
	err = apiError
	return
}

// New returns a Client implementation for this namespace
func New(c dropbox.Config) *apiImpl {
	ctx := apiImpl(dropbox.NewContext(c))
	return &ctx
}
