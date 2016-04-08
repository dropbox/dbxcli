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
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dropbox/dropbox-sdk-go-unofficial/apierror"
	"github.com/dropbox/dropbox-sdk-go-unofficial/async"
	"github.com/dropbox/dropbox-sdk-go-unofficial/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/sharing"
	"github.com/dropbox/dropbox-sdk-go-unofficial/team"
	"github.com/dropbox/dropbox-sdk-go-unofficial/users"
)

type Api interface {
	files.Files
	sharing.Sharing
	team.Team
	users.Users
}

type wrapCopy struct {
	apierror.ApiError
	EndpointError *files.RelocationError `json:"error"`
}

func (dbx *apiImpl) Copy(arg *files.RelocationArg) (res *files.Metadata, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "files", "copy"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapCopy
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapCreateFolder struct {
	apierror.ApiError
	EndpointError *files.CreateFolderError `json:"error"`
}

func (dbx *apiImpl) CreateFolder(arg *files.CreateFolderArg) (res *files.FolderMetadata, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "files", "create_folder"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapCreateFolder
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapDelete struct {
	apierror.ApiError
	EndpointError *files.DeleteError `json:"error"`
}

func (dbx *apiImpl) Delete(arg *files.DeleteArg) (res *files.Metadata, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "files", "delete"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapDelete
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapDownload struct {
	apierror.ApiError
	EndpointError *files.DownloadError `json:"error"`
}

func (dbx *apiImpl) Download(arg *files.DownloadArg) (res *files.FileMetadata, content io.ReadCloser, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("content", "files", "download"), nil)
	if err != nil {
		return
	}

	req.Header.Set("Dropbox-API-Arg", string(b))
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
		log.Printf("resp: %v", resp)
	}
	if err != nil {
		return
	}

	body := []byte(resp.Header.Get("Dropbox-API-Result"))
	content = resp.Body
	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapDownload
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGetMetadata struct {
	apierror.ApiError
	EndpointError *files.GetMetadataError `json:"error"`
}

func (dbx *apiImpl) GetMetadata(arg *files.GetMetadataArg) (res *files.Metadata, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "files", "get_metadata"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGetMetadata
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGetPreview struct {
	apierror.ApiError
	EndpointError *files.PreviewError `json:"error"`
}

func (dbx *apiImpl) GetPreview(arg *files.PreviewArg) (res *files.FileMetadata, content io.ReadCloser, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("content", "files", "get_preview"), nil)
	if err != nil {
		return
	}

	req.Header.Set("Dropbox-API-Arg", string(b))
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
		log.Printf("resp: %v", resp)
	}
	if err != nil {
		return
	}

	body := []byte(resp.Header.Get("Dropbox-API-Result"))
	content = resp.Body
	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGetPreview
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGetThumbnail struct {
	apierror.ApiError
	EndpointError *files.ThumbnailError `json:"error"`
}

func (dbx *apiImpl) GetThumbnail(arg *files.ThumbnailArg) (res *files.FileMetadata, content io.ReadCloser, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("content", "files", "get_thumbnail"), nil)
	if err != nil {
		return
	}

	req.Header.Set("Dropbox-API-Arg", string(b))
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
		log.Printf("resp: %v", resp)
	}
	if err != nil {
		return
	}

	body := []byte(resp.Header.Get("Dropbox-API-Result"))
	content = resp.Body
	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGetThumbnail
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapListFolder struct {
	apierror.ApiError
	EndpointError *files.ListFolderError `json:"error"`
}

func (dbx *apiImpl) ListFolder(arg *files.ListFolderArg) (res *files.ListFolderResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "files", "list_folder"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapListFolder
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapListFolderContinue struct {
	apierror.ApiError
	EndpointError *files.ListFolderContinueError `json:"error"`
}

func (dbx *apiImpl) ListFolderContinue(arg *files.ListFolderContinueArg) (res *files.ListFolderResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "files", "list_folder/continue"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapListFolderContinue
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapListFolderGetLatestCursor struct {
	apierror.ApiError
	EndpointError *files.ListFolderError `json:"error"`
}

func (dbx *apiImpl) ListFolderGetLatestCursor(arg *files.ListFolderArg) (res *files.ListFolderGetLatestCursorResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "files", "list_folder/get_latest_cursor"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapListFolderGetLatestCursor
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapListFolderLongpoll struct {
	apierror.ApiError
	EndpointError *files.ListFolderLongpollError `json:"error"`
}

func (dbx *apiImpl) ListFolderLongpoll(arg *files.ListFolderLongpollArg) (res *files.ListFolderLongpollResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("notify", "files", "list_folder/longpoll"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Del("Authorization")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapListFolderLongpoll
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapListRevisions struct {
	apierror.ApiError
	EndpointError *files.ListRevisionsError `json:"error"`
}

func (dbx *apiImpl) ListRevisions(arg *files.ListRevisionsArg) (res *files.ListRevisionsResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "files", "list_revisions"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapListRevisions
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapMove struct {
	apierror.ApiError
	EndpointError *files.RelocationError `json:"error"`
}

func (dbx *apiImpl) Move(arg *files.RelocationArg) (res *files.Metadata, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "files", "move"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapMove
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapPermanentlyDelete struct {
	apierror.ApiError
	EndpointError *files.DeleteError `json:"error"`
}

func (dbx *apiImpl) PermanentlyDelete(arg *files.DeleteArg) (err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "files", "permanently_delete"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapPermanentlyDelete
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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
	return
}

type wrapRestore struct {
	apierror.ApiError
	EndpointError *files.RestoreError `json:"error"`
}

func (dbx *apiImpl) Restore(arg *files.RestoreArg) (res *files.FileMetadata, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "files", "restore"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapRestore
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapSearch struct {
	apierror.ApiError
	EndpointError *files.SearchError `json:"error"`
}

func (dbx *apiImpl) Search(arg *files.SearchArg) (res *files.SearchResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "files", "search"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapSearch
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapUpload struct {
	apierror.ApiError
	EndpointError *files.UploadError `json:"error"`
}

func (dbx *apiImpl) Upload(arg *files.CommitInfo, content io.Reader) (res *files.FileMetadata, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("content", "files", "upload"), content)
	if err != nil {
		return
	}

	req.Header.Set("Dropbox-API-Arg", string(b))
	req.Header.Set("Content-Type", "application/octet-stream")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapUpload
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapUploadSessionAppend struct {
	apierror.ApiError
	EndpointError *files.UploadSessionLookupError `json:"error"`
}

func (dbx *apiImpl) UploadSessionAppend(arg *files.UploadSessionCursor, content io.Reader) (err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("content", "files", "upload_session/append"), content)
	if err != nil {
		return
	}

	req.Header.Set("Dropbox-API-Arg", string(b))
	req.Header.Set("Content-Type", "application/octet-stream")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapUploadSessionAppend
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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
	return
}

type wrapUploadSessionFinish struct {
	apierror.ApiError
	EndpointError *files.UploadSessionFinishError `json:"error"`
}

func (dbx *apiImpl) UploadSessionFinish(arg *files.UploadSessionFinishArg, content io.Reader) (res *files.FileMetadata, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("content", "files", "upload_session/finish"), content)
	if err != nil {
		return
	}

	req.Header.Set("Dropbox-API-Arg", string(b))
	req.Header.Set("Content-Type", "application/octet-stream")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapUploadSessionFinish
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapUploadSessionStart struct {
	apierror.ApiError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) UploadSessionStart(content io.Reader) (res *files.UploadSessionStartResult, err error) {
	cli := dbx.client

	req, err := http.NewRequest("POST", dbx.generateURL("content", "files", "upload_session/start"), content)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapUploadSessionStart
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapAddFolderMember struct {
	apierror.ApiError
	EndpointError *sharing.AddFolderMemberError `json:"error"`
}

func (dbx *apiImpl) AddFolderMember(arg *sharing.AddFolderMemberArg) (err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "add_folder_member"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapAddFolderMember
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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
	return
}

type wrapCheckJobStatus struct {
	apierror.ApiError
	EndpointError *async.PollError `json:"error"`
}

func (dbx *apiImpl) CheckJobStatus(arg *async.PollArg) (res *sharing.JobStatus, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "check_job_status"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapCheckJobStatus
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapCheckShareJobStatus struct {
	apierror.ApiError
	EndpointError *async.PollError `json:"error"`
}

func (dbx *apiImpl) CheckShareJobStatus(arg *async.PollArg) (res *sharing.ShareFolderJobStatus, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "check_share_job_status"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapCheckShareJobStatus
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapCreateSharedLink struct {
	apierror.ApiError
	EndpointError *sharing.CreateSharedLinkError `json:"error"`
}

func (dbx *apiImpl) CreateSharedLink(arg *sharing.CreateSharedLinkArg) (res *sharing.PathLinkMetadata, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "create_shared_link"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapCreateSharedLink
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapCreateSharedLinkWithSettings struct {
	apierror.ApiError
	EndpointError *sharing.CreateSharedLinkWithSettingsError `json:"error"`
}

func (dbx *apiImpl) CreateSharedLinkWithSettings(arg *sharing.CreateSharedLinkWithSettingsArg) (res *sharing.SharedLinkMetadata, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "create_shared_link_with_settings"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapCreateSharedLinkWithSettings
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGetFolderMetadata struct {
	apierror.ApiError
	EndpointError *sharing.SharedFolderAccessError `json:"error"`
}

func (dbx *apiImpl) GetFolderMetadata(arg *sharing.GetMetadataArgs) (res *sharing.SharedFolderMetadata, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "get_folder_metadata"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGetFolderMetadata
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGetSharedLinkFile struct {
	apierror.ApiError
	EndpointError *sharing.GetSharedLinkFileError `json:"error"`
}

func (dbx *apiImpl) GetSharedLinkFile(arg *sharing.GetSharedLinkMetadataArg) (res *sharing.SharedLinkMetadata, content io.ReadCloser, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("content", "sharing", "get_shared_link_file"), nil)
	if err != nil {
		return
	}

	req.Header.Set("Dropbox-API-Arg", string(b))
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
		log.Printf("resp: %v", resp)
	}
	if err != nil {
		return
	}

	body := []byte(resp.Header.Get("Dropbox-API-Result"))
	content = resp.Body
	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGetSharedLinkFile
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGetSharedLinkMetadata struct {
	apierror.ApiError
	EndpointError *sharing.SharedLinkError `json:"error"`
}

func (dbx *apiImpl) GetSharedLinkMetadata(arg *sharing.GetSharedLinkMetadataArg) (res *sharing.SharedLinkMetadata, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "get_shared_link_metadata"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGetSharedLinkMetadata
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGetSharedLinks struct {
	apierror.ApiError
	EndpointError *sharing.GetSharedLinksError `json:"error"`
}

func (dbx *apiImpl) GetSharedLinks(arg *sharing.GetSharedLinksArg) (res *sharing.GetSharedLinksResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "get_shared_links"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGetSharedLinks
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapListFolderMembers struct {
	apierror.ApiError
	EndpointError *sharing.SharedFolderAccessError `json:"error"`
}

func (dbx *apiImpl) ListFolderMembers(arg *sharing.ListFolderMembersArgs) (res *sharing.SharedFolderMembers, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "list_folder_members"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapListFolderMembers
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapListFolderMembersContinue struct {
	apierror.ApiError
	EndpointError *sharing.ListFolderMembersContinueError `json:"error"`
}

func (dbx *apiImpl) ListFolderMembersContinue(arg *sharing.ListFolderMembersContinueArg) (res *sharing.SharedFolderMembers, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "list_folder_members/continue"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapListFolderMembersContinue
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapListFolders struct {
	apierror.ApiError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) ListFolders(arg *sharing.ListFoldersArgs) (res *sharing.ListFoldersResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "list_folders"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapListFolders
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapListFoldersContinue struct {
	apierror.ApiError
	EndpointError *sharing.ListFoldersContinueError `json:"error"`
}

func (dbx *apiImpl) ListFoldersContinue(arg *sharing.ListFoldersContinueArg) (res *sharing.ListFoldersResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "list_folders/continue"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapListFoldersContinue
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapListMountableFolders struct {
	apierror.ApiError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) ListMountableFolders(arg *sharing.ListFoldersArgs) (res *sharing.ListFoldersResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "list_mountable_folders"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapListMountableFolders
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapListMountableFoldersContinue struct {
	apierror.ApiError
	EndpointError *sharing.ListFoldersContinueError `json:"error"`
}

func (dbx *apiImpl) ListMountableFoldersContinue(arg *sharing.ListFoldersContinueArg) (res *sharing.ListFoldersResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "list_mountable_folders/continue"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapListMountableFoldersContinue
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapListSharedLinks struct {
	apierror.ApiError
	EndpointError *sharing.ListSharedLinksError `json:"error"`
}

func (dbx *apiImpl) ListSharedLinks(arg *sharing.ListSharedLinksArg) (res *sharing.ListSharedLinksResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "list_shared_links"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapListSharedLinks
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapModifySharedLinkSettings struct {
	apierror.ApiError
	EndpointError *sharing.ModifySharedLinkSettingsError `json:"error"`
}

func (dbx *apiImpl) ModifySharedLinkSettings(arg *sharing.ModifySharedLinkSettingsArgs) (res *sharing.SharedLinkMetadata, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "modify_shared_link_settings"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapModifySharedLinkSettings
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapMountFolder struct {
	apierror.ApiError
	EndpointError *sharing.MountFolderError `json:"error"`
}

func (dbx *apiImpl) MountFolder(arg *sharing.MountFolderArg) (res *sharing.SharedFolderMetadata, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "mount_folder"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapMountFolder
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapRelinquishFolderMembership struct {
	apierror.ApiError
	EndpointError *sharing.RelinquishFolderMembershipError `json:"error"`
}

func (dbx *apiImpl) RelinquishFolderMembership(arg *sharing.RelinquishFolderMembershipArg) (err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "relinquish_folder_membership"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapRelinquishFolderMembership
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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
	return
}

type wrapRemoveFolderMember struct {
	apierror.ApiError
	EndpointError *sharing.RemoveFolderMemberError `json:"error"`
}

func (dbx *apiImpl) RemoveFolderMember(arg *sharing.RemoveFolderMemberArg) (res *async.LaunchEmptyResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "remove_folder_member"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapRemoveFolderMember
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapRevokeSharedLink struct {
	apierror.ApiError
	EndpointError *sharing.RevokeSharedLinkError `json:"error"`
}

func (dbx *apiImpl) RevokeSharedLink(arg *sharing.RevokeSharedLinkArg) (err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "revoke_shared_link"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapRevokeSharedLink
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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
	return
}

type wrapShareFolder struct {
	apierror.ApiError
	EndpointError *sharing.ShareFolderError `json:"error"`
}

func (dbx *apiImpl) ShareFolder(arg *sharing.ShareFolderArg) (res *sharing.ShareFolderLaunch, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "share_folder"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapShareFolder
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapTransferFolder struct {
	apierror.ApiError
	EndpointError *sharing.TransferFolderError `json:"error"`
}

func (dbx *apiImpl) TransferFolder(arg *sharing.TransferFolderArg) (err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "transfer_folder"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapTransferFolder
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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
	return
}

type wrapUnmountFolder struct {
	apierror.ApiError
	EndpointError *sharing.UnmountFolderError `json:"error"`
}

func (dbx *apiImpl) UnmountFolder(arg *sharing.UnmountFolderArg) (err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "unmount_folder"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapUnmountFolder
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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
	return
}

type wrapUnshareFolder struct {
	apierror.ApiError
	EndpointError *sharing.UnshareFolderError `json:"error"`
}

func (dbx *apiImpl) UnshareFolder(arg *sharing.UnshareFolderArg) (res *async.LaunchEmptyResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "unshare_folder"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapUnshareFolder
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapUpdateFolderMember struct {
	apierror.ApiError
	EndpointError *sharing.UpdateFolderMemberError `json:"error"`
}

func (dbx *apiImpl) UpdateFolderMember(arg *sharing.UpdateFolderMemberArg) (err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "update_folder_member"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapUpdateFolderMember
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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
	return
}

type wrapUpdateFolderPolicy struct {
	apierror.ApiError
	EndpointError *sharing.UpdateFolderPolicyError `json:"error"`
}

func (dbx *apiImpl) UpdateFolderPolicy(arg *sharing.UpdateFolderPolicyArg) (res *sharing.SharedFolderMetadata, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "sharing", "update_folder_policy"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapUpdateFolderPolicy
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapDevicesListMemberDevices struct {
	apierror.ApiError
	EndpointError *team.ListMemberDevicesError `json:"error"`
}

func (dbx *apiImpl) DevicesListMemberDevices(arg *team.ListMemberDevicesArg) (res *team.ListMemberDevicesResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "devices/list_member_devices"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapDevicesListMemberDevices
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapDevicesListTeamDevices struct {
	apierror.ApiError
	EndpointError *team.ListTeamDevicesError `json:"error"`
}

func (dbx *apiImpl) DevicesListTeamDevices(arg *team.ListTeamDevicesArg) (res *team.ListTeamDevicesResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "devices/list_team_devices"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapDevicesListTeamDevices
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapDevicesRevokeDeviceSession struct {
	apierror.ApiError
	EndpointError *team.RevokeDeviceSessionError `json:"error"`
}

func (dbx *apiImpl) DevicesRevokeDeviceSession(arg *team.RevokeDeviceSessionArg) (err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "devices/revoke_device_session"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapDevicesRevokeDeviceSession
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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
	return
}

type wrapDevicesRevokeDeviceSessionBatch struct {
	apierror.ApiError
	EndpointError *team.RevokeDeviceSessionBatchError `json:"error"`
}

func (dbx *apiImpl) DevicesRevokeDeviceSessionBatch(arg *team.RevokeDeviceSessionBatchArg) (res *team.RevokeDeviceSessionBatchResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "devices/revoke_device_session_batch"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapDevicesRevokeDeviceSessionBatch
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGetInfo struct {
	apierror.ApiError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) GetInfo() (res *team.TeamGetInfoResult, err error) {
	cli := dbx.client

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "get_info"), nil)
	if err != nil {
		return
	}

	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGetInfo
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGroupsCreate struct {
	apierror.ApiError
	EndpointError *team.GroupCreateError `json:"error"`
}

func (dbx *apiImpl) GroupsCreate(arg *team.GroupCreateArg) (res *team.GroupFullInfo, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "groups/create"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGroupsCreate
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGroupsDelete struct {
	apierror.ApiError
	EndpointError *team.GroupDeleteError `json:"error"`
}

func (dbx *apiImpl) GroupsDelete(arg *team.GroupSelector) (res *async.LaunchEmptyResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "groups/delete"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGroupsDelete
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGroupsGetInfo struct {
	apierror.ApiError
	EndpointError *team.GroupsGetInfoError `json:"error"`
}

func (dbx *apiImpl) GroupsGetInfo(arg *team.GroupsSelector) (res []*team.GroupsGetInfoItem, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "groups/get_info"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGroupsGetInfo
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGroupsJobStatusGet struct {
	apierror.ApiError
	EndpointError *team.GroupsPollError `json:"error"`
}

func (dbx *apiImpl) GroupsJobStatusGet(arg *async.PollArg) (res *async.PollEmptyResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "groups/job_status/get"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGroupsJobStatusGet
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGroupsList struct {
	apierror.ApiError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) GroupsList(arg *team.GroupsListArg) (res *team.GroupsListResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "groups/list"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGroupsList
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGroupsListContinue struct {
	apierror.ApiError
	EndpointError *team.GroupsListContinueError `json:"error"`
}

func (dbx *apiImpl) GroupsListContinue(arg *team.GroupsListContinueArg) (res *team.GroupsListResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "groups/list/continue"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGroupsListContinue
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGroupsMembersAdd struct {
	apierror.ApiError
	EndpointError *team.GroupMembersAddError `json:"error"`
}

func (dbx *apiImpl) GroupsMembersAdd(arg *team.GroupMembersAddArg) (res *team.GroupMembersChangeResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "groups/members/add"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGroupsMembersAdd
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGroupsMembersRemove struct {
	apierror.ApiError
	EndpointError *team.GroupMembersRemoveError `json:"error"`
}

func (dbx *apiImpl) GroupsMembersRemove(arg *team.GroupMembersRemoveArg) (res *team.GroupMembersChangeResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "groups/members/remove"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGroupsMembersRemove
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGroupsMembersSetAccessType struct {
	apierror.ApiError
	EndpointError *team.GroupMemberSelectorError `json:"error"`
}

func (dbx *apiImpl) GroupsMembersSetAccessType(arg *team.GroupMembersSetAccessTypeArg) (res []*team.GroupsGetInfoItem, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "groups/members/set_access_type"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGroupsMembersSetAccessType
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGroupsUpdate struct {
	apierror.ApiError
	EndpointError *team.GroupUpdateError `json:"error"`
}

func (dbx *apiImpl) GroupsUpdate(arg *team.GroupUpdateArgs) (res *team.GroupFullInfo, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "groups/update"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGroupsUpdate
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapLinkedAppsListMemberLinkedApps struct {
	apierror.ApiError
	EndpointError *team.ListMemberAppsError `json:"error"`
}

func (dbx *apiImpl) LinkedAppsListMemberLinkedApps(arg *team.ListMemberAppsArg) (res *team.ListMemberAppsResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "linked_apps/list_member_linked_apps"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapLinkedAppsListMemberLinkedApps
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapLinkedAppsListTeamLinkedApps struct {
	apierror.ApiError
	EndpointError *team.ListTeamAppsError `json:"error"`
}

func (dbx *apiImpl) LinkedAppsListTeamLinkedApps(arg *team.ListTeamAppsArg) (res *team.ListTeamAppsResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "linked_apps/list_team_linked_apps"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapLinkedAppsListTeamLinkedApps
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapLinkedAppsRevokeLinkedApp struct {
	apierror.ApiError
	EndpointError *team.RevokeLinkedAppError `json:"error"`
}

func (dbx *apiImpl) LinkedAppsRevokeLinkedApp(arg *team.RevokeLinkedApiAppArg) (err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "linked_apps/revoke_linked_app"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapLinkedAppsRevokeLinkedApp
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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
	return
}

type wrapLinkedAppsRevokeLinkedAppBatch struct {
	apierror.ApiError
	EndpointError *team.RevokeLinkedAppBatchError `json:"error"`
}

func (dbx *apiImpl) LinkedAppsRevokeLinkedAppBatch(arg *team.RevokeLinkedApiAppBatchArg) (res *team.RevokeLinkedAppBatchResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "linked_apps/revoke_linked_app_batch"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapLinkedAppsRevokeLinkedAppBatch
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapMembersAdd struct {
	apierror.ApiError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) MembersAdd(arg *team.MembersAddArg) (res *team.MembersAddLaunch, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "members/add"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapMembersAdd
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapMembersAddJobStatusGet struct {
	apierror.ApiError
	EndpointError *async.PollError `json:"error"`
}

func (dbx *apiImpl) MembersAddJobStatusGet(arg *async.PollArg) (res *team.MembersAddJobStatus, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "members/add/job_status/get"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapMembersAddJobStatusGet
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapMembersGetInfo struct {
	apierror.ApiError
	EndpointError *team.MembersGetInfoError `json:"error"`
}

func (dbx *apiImpl) MembersGetInfo(arg *team.MembersGetInfoArgs) (res []*team.MembersGetInfoItem, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "members/get_info"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapMembersGetInfo
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapMembersList struct {
	apierror.ApiError
	EndpointError *team.MembersListError `json:"error"`
}

func (dbx *apiImpl) MembersList(arg *team.MembersListArg) (res *team.MembersListResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "members/list"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapMembersList
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapMembersListContinue struct {
	apierror.ApiError
	EndpointError *team.MembersListContinueError `json:"error"`
}

func (dbx *apiImpl) MembersListContinue(arg *team.MembersListContinueArg) (res *team.MembersListResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "members/list/continue"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapMembersListContinue
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapMembersRemove struct {
	apierror.ApiError
	EndpointError *team.MembersRemoveError `json:"error"`
}

func (dbx *apiImpl) MembersRemove(arg *team.MembersRemoveArg) (res *async.LaunchEmptyResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "members/remove"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapMembersRemove
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapMembersRemoveJobStatusGet struct {
	apierror.ApiError
	EndpointError *async.PollError `json:"error"`
}

func (dbx *apiImpl) MembersRemoveJobStatusGet(arg *async.PollArg) (res *async.PollEmptyResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "members/remove/job_status/get"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapMembersRemoveJobStatusGet
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapMembersSendWelcomeEmail struct {
	apierror.ApiError
	EndpointError *team.MembersSendWelcomeError `json:"error"`
}

func (dbx *apiImpl) MembersSendWelcomeEmail(arg *team.UserSelectorArg) (err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "members/send_welcome_email"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapMembersSendWelcomeEmail
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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
	return
}

type wrapMembersSetAdminPermissions struct {
	apierror.ApiError
	EndpointError *team.MembersSetPermissionsError `json:"error"`
}

func (dbx *apiImpl) MembersSetAdminPermissions(arg *team.MembersSetPermissionsArg) (res *team.MembersSetPermissionsResult, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "members/set_admin_permissions"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapMembersSetAdminPermissions
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapMembersSetProfile struct {
	apierror.ApiError
	EndpointError *team.MembersSetProfileError `json:"error"`
}

func (dbx *apiImpl) MembersSetProfile(arg *team.MembersSetProfileArg) (res *team.TeamMemberInfo, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "members/set_profile"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapMembersSetProfile
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapMembersSuspend struct {
	apierror.ApiError
	EndpointError *team.MembersSuspendError `json:"error"`
}

func (dbx *apiImpl) MembersSuspend(arg *team.MembersDeactivateArg) (err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "members/suspend"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapMembersSuspend
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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
	return
}

type wrapMembersUnsuspend struct {
	apierror.ApiError
	EndpointError *team.MembersUnsuspendError `json:"error"`
}

func (dbx *apiImpl) MembersUnsuspend(arg *team.MembersUnsuspendArg) (err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "members/unsuspend"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapMembersUnsuspend
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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
	return
}

type wrapReportsGetActivity struct {
	apierror.ApiError
	EndpointError *team.DateRangeError `json:"error"`
}

func (dbx *apiImpl) ReportsGetActivity(arg *team.DateRange) (res *team.GetActivityReport, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "reports/get_activity"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapReportsGetActivity
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapReportsGetDevices struct {
	apierror.ApiError
	EndpointError *team.DateRangeError `json:"error"`
}

func (dbx *apiImpl) ReportsGetDevices(arg *team.DateRange) (res *team.GetDevicesReport, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "reports/get_devices"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapReportsGetDevices
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapReportsGetMembership struct {
	apierror.ApiError
	EndpointError *team.DateRangeError `json:"error"`
}

func (dbx *apiImpl) ReportsGetMembership(arg *team.DateRange) (res *team.GetMembershipReport, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "reports/get_membership"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapReportsGetMembership
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapReportsGetStorage struct {
	apierror.ApiError
	EndpointError *team.DateRangeError `json:"error"`
}

func (dbx *apiImpl) ReportsGetStorage(arg *team.DateRange) (res *team.GetStorageReport, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "team", "reports/get_storage"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapReportsGetStorage
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGetAccount struct {
	apierror.ApiError
	EndpointError *users.GetAccountError `json:"error"`
}

func (dbx *apiImpl) GetAccount(arg *users.GetAccountArg) (res *users.BasicAccount, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "users", "get_account"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGetAccount
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGetAccountBatch struct {
	apierror.ApiError
	EndpointError *users.GetAccountBatchError `json:"error"`
}

func (dbx *apiImpl) GetAccountBatch(arg *users.GetAccountBatchArg) (res []*users.BasicAccount, err error) {
	cli := dbx.client

	if dbx.options.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", dbx.generateURL("api", "users", "get_account_batch"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGetAccountBatch
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGetCurrentAccount struct {
	apierror.ApiError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) GetCurrentAccount() (res *users.FullAccount, err error) {
	cli := dbx.client

	req, err := http.NewRequest("POST", dbx.generateURL("api", "users", "get_current_account"), nil)
	if err != nil {
		return
	}

	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGetCurrentAccount
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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

type wrapGetSpaceUsage struct {
	apierror.ApiError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) GetSpaceUsage() (res *users.SpaceUsage, err error) {
	cli := dbx.client

	req, err := http.NewRequest("POST", dbx.generateURL("api", "users", "get_space_usage"), nil)
	if err != nil {
		return
	}

	if dbx.options.AsMemberId != "" {
		req.Header.Set("Dropbox-API-Select-User", dbx.options.AsMemberId)
	}
	if dbx.options.Verbose {
		log.Printf("req: %v", req)
	}
	resp, err := cli.Do(req)
	if dbx.options.Verbose {
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

	if dbx.options.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var errWrap wrapGetSpaceUsage
			err = json.Unmarshal(body, &errWrap)
			if err != nil {
				return
			}
			err = errWrap
			return
		}
		var apiError apierror.ApiError
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
