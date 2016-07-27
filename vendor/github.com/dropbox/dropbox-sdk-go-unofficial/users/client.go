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

// This namespace contains endpoints and data types for user management.
package users

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	dropbox "github.com/dropbox/dropbox-sdk-go-unofficial"
)

type Client interface {
	// Get information about a user's account.
	GetAccount(arg *GetAccountArg) (res *BasicAccount, err error)
	// Get information about multiple user accounts.  At most 300 accounts may
	// be queried per request.
	GetAccountBatch(arg *GetAccountBatchArg) (res []*BasicAccount, err error)
	// Get information about the current user's account.
	GetCurrentAccount() (res *FullAccount, err error)
	// Get the space usage information for the current user's account.
	GetSpaceUsage() (res *SpaceUsage, err error)
}

type apiImpl dropbox.Context
type GetAccountApiError struct {
	dropbox.ApiError
	EndpointError *GetAccountError `json:"error"`
}

func (dbx *apiImpl) GetAccount(arg *GetAccountArg) (res *BasicAccount, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "users", "get_account"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.Config.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.Config.AsMemberId)
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
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var apiError GetAccountApiError
			err = json.Unmarshal(body, &apiError)
			if err != nil {
				return
			}
			err = apiError
			return
		}
		var apiError dropbox.ApiError
		if resp.StatusCode == 400 {
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
	err = json.Unmarshal(body, &res)
	if err != nil {
		return
	}

	return
}

type GetAccountBatchApiError struct {
	dropbox.ApiError
	EndpointError *GetAccountBatchError `json:"error"`
}

func (dbx *apiImpl) GetAccountBatch(arg *GetAccountBatchArg) (res []*BasicAccount, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "users", "get_account_batch"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.Config.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.Config.AsMemberId)
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
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var apiError GetAccountBatchApiError
			err = json.Unmarshal(body, &apiError)
			if err != nil {
				return
			}
			err = apiError
			return
		}
		var apiError dropbox.ApiError
		if resp.StatusCode == 400 {
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
	err = json.Unmarshal(body, &res)
	if err != nil {
		return
	}

	return
}

type GetCurrentAccountApiError struct {
	dropbox.ApiError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) GetCurrentAccount() (res *FullAccount, err error) {
	cli := dbx.Client

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "users", "get_current_account"), nil)
	if err != nil {
		return
	}

	if dbx.Config.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.Config.AsMemberId)
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
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var apiError GetCurrentAccountApiError
			err = json.Unmarshal(body, &apiError)
			if err != nil {
				return
			}
			err = apiError
			return
		}
		var apiError dropbox.ApiError
		if resp.StatusCode == 400 {
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
	err = json.Unmarshal(body, &res)
	if err != nil {
		return
	}

	return
}

type GetSpaceUsageApiError struct {
	dropbox.ApiError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) GetSpaceUsage() (res *SpaceUsage, err error) {
	cli := dbx.Client

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "users", "get_space_usage"), nil)
	if err != nil {
		return
	}

	if dbx.Config.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.Config.AsMemberId)
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
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var apiError GetSpaceUsageApiError
			err = json.Unmarshal(body, &apiError)
			if err != nil {
				return
			}
			err = apiError
			return
		}
		var apiError dropbox.ApiError
		if resp.StatusCode == 400 {
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
	err = json.Unmarshal(body, &res)
	if err != nil {
		return
	}

	return
}

func New(c dropbox.Config) *apiImpl {
	ctx := apiImpl(dropbox.NewContext(c))
	return &ctx
}
