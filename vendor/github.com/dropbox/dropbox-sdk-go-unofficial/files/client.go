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

// This namespace contains endpoints and data types for basic file operations.
package files

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	dropbox "github.com/dropbox/dropbox-sdk-go-unofficial"
	"github.com/dropbox/dropbox-sdk-go-unofficial/async"
	"github.com/dropbox/dropbox-sdk-go-unofficial/properties"
)

type Client interface {
	// Returns the metadata for a file or folder. This is an alpha endpoint
	// compatible with the properties API. Note: Metadata for the root folder is
	// unsupported.
	AlphaGetMetadata(arg *AlphaGetMetadataArg) (res IsMetadata, err error)
	// Create a new file with the contents provided in the request. Note that
	// this endpoint is part of the properties API alpha and is slightly
	// different from `upload`. Do not use this to upload a file larger than 150
	// MB. Instead, create an upload session with `uploadSessionStart`.
	AlphaUpload(arg *CommitInfoWithProperties, content io.Reader) (res *FileMetadata, err error)
	// Copy a file or folder to a different location in the user's Dropbox. If
	// the source path is a folder all its contents will be copied.
	Copy(arg *RelocationArg) (res IsMetadata, err error)
	// Get a copy reference to a file or folder. This reference string can be
	// used to save that file or folder to another user's Dropbox by passing it
	// to `copyReferenceSave`.
	CopyReferenceGet(arg *GetCopyReferenceArg) (res *GetCopyReferenceResult, err error)
	// Save a copy reference returned by `copyReferenceGet` to the user's
	// Dropbox.
	CopyReferenceSave(arg *SaveCopyReferenceArg) (res *SaveCopyReferenceResult, err error)
	// Create a folder at a given path.
	CreateFolder(arg *CreateFolderArg) (res *FolderMetadata, err error)
	// Delete the file or folder at a given path. If the path is a folder, all
	// its contents will be deleted too. A successful response indicates that
	// the file or folder was deleted. The returned metadata will be the
	// corresponding `FileMetadata` or `FolderMetadata` for the item at time of
	// deletion, and not a `DeletedMetadata` object.
	Delete(arg *DeleteArg) (res IsMetadata, err error)
	// Download a file from a user's Dropbox.
	Download(arg *DownloadArg) (res *FileMetadata, content io.ReadCloser, err error)
	// Returns the metadata for a file or folder. Note: Metadata for the root
	// folder is unsupported.
	GetMetadata(arg *GetMetadataArg) (res IsMetadata, err error)
	// Get a preview for a file. Currently previews are only generated for the
	// files with  the following extensions: .doc, .docx, .docm, .ppt, .pps,
	// .ppsx, .ppsm, .pptx, .pptm,  .xls, .xlsx, .xlsm, .rtf
	GetPreview(arg *PreviewArg) (res *FileMetadata, content io.ReadCloser, err error)
	// Get a temporary link to stream content of a file. This link will expire
	// in four hours and afterwards you will get 410 Gone. Content-Type of the
	// link is determined automatically by the file's mime type.
	GetTemporaryLink(arg *GetTemporaryLinkArg) (res *GetTemporaryLinkResult, err error)
	// Get a thumbnail for an image. This method currently supports files with
	// the following file extensions: jpg, jpeg, png, tiff, tif, gif and bmp.
	// Photos that are larger than 20MB in size won't be converted to a
	// thumbnail.
	GetThumbnail(arg *ThumbnailArg) (res *FileMetadata, content io.ReadCloser, err error)
	// Returns the contents of a folder.
	ListFolder(arg *ListFolderArg) (res *ListFolderResult, err error)
	// Once a cursor has been retrieved from `listFolder`, use this to paginate
	// through all files and retrieve updates to the folder.
	ListFolderContinue(arg *ListFolderContinueArg) (res *ListFolderResult, err error)
	// A way to quickly get a cursor for the folder's state. Unlike
	// `listFolder`, `listFolderGetLatestCursor` doesn't return any entries.
	// This endpoint is for app which only needs to know about new files and
	// modifications and doesn't need to know about files that already exist in
	// Dropbox.
	ListFolderGetLatestCursor(arg *ListFolderArg) (res *ListFolderGetLatestCursorResult, err error)
	// A longpoll endpoint to wait for changes on an account. In conjunction
	// with `listFolderContinue`, this call gives you a low-latency way to
	// monitor an account for file changes. The connection will block until
	// there are changes available or a timeout occurs. This endpoint is useful
	// mostly for client-side apps. If you're looking for server-side
	// notifications, check out our `webhooks documentation`
	// <https://www.dropbox.com/developers/reference/webhooks>.
	ListFolderLongpoll(arg *ListFolderLongpollArg) (res *ListFolderLongpollResult, err error)
	// Return revisions of a file
	ListRevisions(arg *ListRevisionsArg) (res *ListRevisionsResult, err error)
	// Move a file or folder to a different location in the user's Dropbox. If
	// the source path is a folder all its contents will be moved.
	Move(arg *RelocationArg) (res IsMetadata, err error)
	// Permanently delete the file or folder at a given path (see
	// https://www.dropbox.com/en/help/40). Note: This endpoint is only
	// available for Dropbox Business apps.
	PermanentlyDelete(arg *DeleteArg) (err error)
	// Add custom properties to a file using a filled property template. See
	// properties/template/add to create new property templates.
	PropertiesAdd(arg *PropertyGroupWithPath) (err error)
	// Overwrite custom properties from a specified template associated with a
	// file.
	PropertiesOverwrite(arg *PropertyGroupWithPath) (err error)
	// Remove all custom properties from a specified template associated with a
	// file. To remove specific property key value pairs, see
	// `propertiesUpdate`. To update a property template, see
	// properties/template/update. Property templates can't be removed once
	// created.
	PropertiesRemove(arg *RemovePropertiesArg) (err error)
	// Get the schema for a specified template.
	PropertiesTemplateGet(arg *properties.GetPropertyTemplateArg) (res *properties.GetPropertyTemplateResult, err error)
	// Get the property template identifiers for a user. To get the schema of
	// each template use `propertiesTemplateGet`.
	PropertiesTemplateList() (res *properties.ListPropertyTemplateIds, err error)
	// Add, update or remove custom properties from a specified template
	// associated with a file. Fields that already exist and not described in
	// the request will not be modified.
	PropertiesUpdate(arg *UpdatePropertyGroupArg) (err error)
	// Restore a file to a specific revision
	Restore(arg *RestoreArg) (res *FileMetadata, err error)
	// Save a specified URL into a file in user's Dropbox. If the given path
	// already exists, the file will be renamed to avoid the conflict (e.g.
	// myfile (1).txt).
	SaveUrl(arg *SaveUrlArg) (res *SaveUrlResult, err error)
	// Check the status of a `saveUrl` job.
	SaveUrlCheckJobStatus(arg *async.PollArg) (res *SaveUrlJobStatus, err error)
	// Searches for files and folders. Note: Recent changes may not immediately
	// be reflected in search results due to a short delay in indexing.
	Search(arg *SearchArg) (res *SearchResult, err error)
	// Create a new file with the contents provided in the request. Do not use
	// this to upload a file larger than 150 MB. Instead, create an upload
	// session with `uploadSessionStart`.
	Upload(arg *CommitInfo, content io.Reader) (res *FileMetadata, err error)
	// Append more data to an upload session. A single request should not upload
	// more than 150 MB of file contents.
	UploadSessionAppend(arg *UploadSessionCursor, content io.Reader) (err error)
	// Append more data to an upload session. When the parameter close is set,
	// this call will close the session. A single request should not upload more
	// than 150 MB of file contents.
	UploadSessionAppendV2(arg *UploadSessionAppendArg, content io.Reader) (err error)
	// Finish an upload session and save the uploaded data to the given file
	// path. A single request should not upload more than 150 MB of file
	// contents.
	UploadSessionFinish(arg *UploadSessionFinishArg, content io.Reader) (res *FileMetadata, err error)
	// Upload sessions allow you to upload a single file using multiple
	// requests. This call starts a new upload session with the given data.  You
	// can then use `uploadSessionAppendV2` to add more data and
	// `uploadSessionFinish` to save all the data to a file in Dropbox. A single
	// request should not upload more than 150 MB of file contents.
	UploadSessionStart(arg *UploadSessionStartArg, content io.Reader) (res *UploadSessionStartResult, err error)
}

type apiImpl dropbox.Context
type AlphaGetMetadataApiError struct {
	dropbox.ApiError
	EndpointError *AlphaGetMetadataError `json:"error"`
}

func (dbx *apiImpl) AlphaGetMetadata(arg *AlphaGetMetadataArg) (res IsMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "alpha/get_metadata"), bytes.NewReader(b))
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
			var apiError AlphaGetMetadataApiError
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
	var tmp metadataUnion
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return
	}
	switch tmp.Tag {
	case "file":
		res = tmp.File

	case "folder":
		res = tmp.Folder

	case "deleted":
		res = tmp.Deleted

	}
	return
}

type AlphaUploadApiError struct {
	dropbox.ApiError
	EndpointError *UploadErrorWithProperties `json:"error"`
}

func (dbx *apiImpl) AlphaUpload(arg *CommitInfoWithProperties, content io.Reader) (res *FileMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("content", "files", "alpha/upload"), content)
	if err != nil {
		return
	}

	req.Header.Set("Dropbox-API-Arg", string(b))
	req.Header.Set("Content-Type", "application/octet-stream")
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
			var apiError AlphaUploadApiError
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

type CopyApiError struct {
	dropbox.ApiError
	EndpointError *RelocationError `json:"error"`
}

func (dbx *apiImpl) Copy(arg *RelocationArg) (res IsMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "copy"), bytes.NewReader(b))
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
			var apiError CopyApiError
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
	var tmp metadataUnion
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return
	}
	switch tmp.Tag {
	case "file":
		res = tmp.File

	case "folder":
		res = tmp.Folder

	case "deleted":
		res = tmp.Deleted

	}
	return
}

type CopyReferenceGetApiError struct {
	dropbox.ApiError
	EndpointError *GetCopyReferenceError `json:"error"`
}

func (dbx *apiImpl) CopyReferenceGet(arg *GetCopyReferenceArg) (res *GetCopyReferenceResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "copy_reference/get"), bytes.NewReader(b))
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
			var apiError CopyReferenceGetApiError
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

type CopyReferenceSaveApiError struct {
	dropbox.ApiError
	EndpointError *SaveCopyReferenceError `json:"error"`
}

func (dbx *apiImpl) CopyReferenceSave(arg *SaveCopyReferenceArg) (res *SaveCopyReferenceResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "copy_reference/save"), bytes.NewReader(b))
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
			var apiError CopyReferenceSaveApiError
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

type CreateFolderApiError struct {
	dropbox.ApiError
	EndpointError *CreateFolderError `json:"error"`
}

func (dbx *apiImpl) CreateFolder(arg *CreateFolderArg) (res *FolderMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "create_folder"), bytes.NewReader(b))
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
			var apiError CreateFolderApiError
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

type DeleteApiError struct {
	dropbox.ApiError
	EndpointError *DeleteError `json:"error"`
}

func (dbx *apiImpl) Delete(arg *DeleteArg) (res IsMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "delete"), bytes.NewReader(b))
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
			var apiError DeleteApiError
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
	var tmp metadataUnion
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return
	}
	switch tmp.Tag {
	case "file":
		res = tmp.File

	case "folder":
		res = tmp.Folder

	case "deleted":
		res = tmp.Deleted

	}
	return
}

type DownloadApiError struct {
	dropbox.ApiError
	EndpointError *DownloadError `json:"error"`
}

func (dbx *apiImpl) Download(arg *DownloadArg) (res *FileMetadata, content io.ReadCloser, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("content", "files", "download"), bytes.NewReader(b))
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

	body := []byte(resp.Header.Get("Dropbox-API-Result"))
	content = resp.Body
	if dbx.Config.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var apiError DownloadApiError
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

type GetMetadataApiError struct {
	dropbox.ApiError
	EndpointError *GetMetadataError `json:"error"`
}

func (dbx *apiImpl) GetMetadata(arg *GetMetadataArg) (res IsMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "get_metadata"), bytes.NewReader(b))
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
			var apiError GetMetadataApiError
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
	var tmp metadataUnion
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return
	}
	switch tmp.Tag {
	case "file":
		res = tmp.File

	case "folder":
		res = tmp.Folder

	case "deleted":
		res = tmp.Deleted

	}
	return
}

type GetPreviewApiError struct {
	dropbox.ApiError
	EndpointError *PreviewError `json:"error"`
}

func (dbx *apiImpl) GetPreview(arg *PreviewArg) (res *FileMetadata, content io.ReadCloser, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("content", "files", "get_preview"), bytes.NewReader(b))
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

	body := []byte(resp.Header.Get("Dropbox-API-Result"))
	content = resp.Body
	if dbx.Config.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var apiError GetPreviewApiError
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

type GetTemporaryLinkApiError struct {
	dropbox.ApiError
	EndpointError *GetTemporaryLinkError `json:"error"`
}

func (dbx *apiImpl) GetTemporaryLink(arg *GetTemporaryLinkArg) (res *GetTemporaryLinkResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "get_temporary_link"), bytes.NewReader(b))
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
			var apiError GetTemporaryLinkApiError
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

type GetThumbnailApiError struct {
	dropbox.ApiError
	EndpointError *ThumbnailError `json:"error"`
}

func (dbx *apiImpl) GetThumbnail(arg *ThumbnailArg) (res *FileMetadata, content io.ReadCloser, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("content", "files", "get_thumbnail"), bytes.NewReader(b))
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

	body := []byte(resp.Header.Get("Dropbox-API-Result"))
	content = resp.Body
	if dbx.Config.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 409 {
			var apiError GetThumbnailApiError
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

type ListFolderApiError struct {
	dropbox.ApiError
	EndpointError *ListFolderError `json:"error"`
}

func (dbx *apiImpl) ListFolder(arg *ListFolderArg) (res *ListFolderResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "list_folder"), bytes.NewReader(b))
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
			var apiError ListFolderApiError
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

type ListFolderContinueApiError struct {
	dropbox.ApiError
	EndpointError *ListFolderContinueError `json:"error"`
}

func (dbx *apiImpl) ListFolderContinue(arg *ListFolderContinueArg) (res *ListFolderResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "list_folder/continue"), bytes.NewReader(b))
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
			var apiError ListFolderContinueApiError
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

type ListFolderGetLatestCursorApiError struct {
	dropbox.ApiError
	EndpointError *ListFolderError `json:"error"`
}

func (dbx *apiImpl) ListFolderGetLatestCursor(arg *ListFolderArg) (res *ListFolderGetLatestCursorResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "list_folder/get_latest_cursor"), bytes.NewReader(b))
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
			var apiError ListFolderGetLatestCursorApiError
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

type ListFolderLongpollApiError struct {
	dropbox.ApiError
	EndpointError *ListFolderLongpollError `json:"error"`
}

func (dbx *apiImpl) ListFolderLongpoll(arg *ListFolderLongpollArg) (res *ListFolderLongpollResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("notify", "files", "list_folder/longpoll"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Del("Authorization")
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
			var apiError ListFolderLongpollApiError
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

type ListRevisionsApiError struct {
	dropbox.ApiError
	EndpointError *ListRevisionsError `json:"error"`
}

func (dbx *apiImpl) ListRevisions(arg *ListRevisionsArg) (res *ListRevisionsResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "list_revisions"), bytes.NewReader(b))
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
			var apiError ListRevisionsApiError
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

type MoveApiError struct {
	dropbox.ApiError
	EndpointError *RelocationError `json:"error"`
}

func (dbx *apiImpl) Move(arg *RelocationArg) (res IsMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "move"), bytes.NewReader(b))
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
			var apiError MoveApiError
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
	var tmp metadataUnion
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return
	}
	switch tmp.Tag {
	case "file":
		res = tmp.File

	case "folder":
		res = tmp.Folder

	case "deleted":
		res = tmp.Deleted

	}
	return
}

type PermanentlyDeleteApiError struct {
	dropbox.ApiError
	EndpointError *DeleteError `json:"error"`
}

func (dbx *apiImpl) PermanentlyDelete(arg *DeleteArg) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "permanently_delete"), bytes.NewReader(b))
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
			var apiError PermanentlyDeleteApiError
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
	return
}

type PropertiesAddApiError struct {
	dropbox.ApiError
	EndpointError *AddPropertiesError `json:"error"`
}

func (dbx *apiImpl) PropertiesAdd(arg *PropertyGroupWithPath) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "properties/add"), bytes.NewReader(b))
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
			var apiError PropertiesAddApiError
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
	return
}

type PropertiesOverwriteApiError struct {
	dropbox.ApiError
	EndpointError *InvalidPropertyGroupError `json:"error"`
}

func (dbx *apiImpl) PropertiesOverwrite(arg *PropertyGroupWithPath) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "properties/overwrite"), bytes.NewReader(b))
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
			var apiError PropertiesOverwriteApiError
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
	return
}

type PropertiesRemoveApiError struct {
	dropbox.ApiError
	EndpointError *RemovePropertiesError `json:"error"`
}

func (dbx *apiImpl) PropertiesRemove(arg *RemovePropertiesArg) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "properties/remove"), bytes.NewReader(b))
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
			var apiError PropertiesRemoveApiError
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
	return
}

type PropertiesTemplateGetApiError struct {
	dropbox.ApiError
	EndpointError *properties.PropertyTemplateError `json:"error"`
}

func (dbx *apiImpl) PropertiesTemplateGet(arg *properties.GetPropertyTemplateArg) (res *properties.GetPropertyTemplateResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "properties/template/get"), bytes.NewReader(b))
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
			var apiError PropertiesTemplateGetApiError
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

type PropertiesTemplateListApiError struct {
	dropbox.ApiError
	EndpointError *properties.PropertyTemplateError `json:"error"`
}

func (dbx *apiImpl) PropertiesTemplateList() (res *properties.ListPropertyTemplateIds, err error) {
	cli := dbx.Client

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "properties/template/list"), nil)
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
			var apiError PropertiesTemplateListApiError
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

type PropertiesUpdateApiError struct {
	dropbox.ApiError
	EndpointError *UpdatePropertiesError `json:"error"`
}

func (dbx *apiImpl) PropertiesUpdate(arg *UpdatePropertyGroupArg) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "properties/update"), bytes.NewReader(b))
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
			var apiError PropertiesUpdateApiError
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
	return
}

type RestoreApiError struct {
	dropbox.ApiError
	EndpointError *RestoreError `json:"error"`
}

func (dbx *apiImpl) Restore(arg *RestoreArg) (res *FileMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "restore"), bytes.NewReader(b))
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
			var apiError RestoreApiError
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

type SaveUrlApiError struct {
	dropbox.ApiError
	EndpointError *SaveUrlError `json:"error"`
}

func (dbx *apiImpl) SaveUrl(arg *SaveUrlArg) (res *SaveUrlResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "save_url"), bytes.NewReader(b))
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
			var apiError SaveUrlApiError
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

type SaveUrlCheckJobStatusApiError struct {
	dropbox.ApiError
	EndpointError *async.PollError `json:"error"`
}

func (dbx *apiImpl) SaveUrlCheckJobStatus(arg *async.PollArg) (res *SaveUrlJobStatus, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "save_url/check_job_status"), bytes.NewReader(b))
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
			var apiError SaveUrlCheckJobStatusApiError
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

type SearchApiError struct {
	dropbox.ApiError
	EndpointError *SearchError `json:"error"`
}

func (dbx *apiImpl) Search(arg *SearchArg) (res *SearchResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "search"), bytes.NewReader(b))
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
			var apiError SearchApiError
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

type UploadApiError struct {
	dropbox.ApiError
	EndpointError *UploadError `json:"error"`
}

func (dbx *apiImpl) Upload(arg *CommitInfo, content io.Reader) (res *FileMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("content", "files", "upload"), content)
	if err != nil {
		return
	}

	req.Header.Set("Dropbox-API-Arg", string(b))
	req.Header.Set("Content-Type", "application/octet-stream")
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
			var apiError UploadApiError
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

type UploadSessionAppendApiError struct {
	dropbox.ApiError
	EndpointError *UploadSessionLookupError `json:"error"`
}

func (dbx *apiImpl) UploadSessionAppend(arg *UploadSessionCursor, content io.Reader) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("content", "files", "upload_session/append"), content)
	if err != nil {
		return
	}

	req.Header.Set("Dropbox-API-Arg", string(b))
	req.Header.Set("Content-Type", "application/octet-stream")
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
			var apiError UploadSessionAppendApiError
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
	return
}

type UploadSessionAppendV2ApiError struct {
	dropbox.ApiError
	EndpointError *UploadSessionLookupError `json:"error"`
}

func (dbx *apiImpl) UploadSessionAppendV2(arg *UploadSessionAppendArg, content io.Reader) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("content", "files", "upload_session/append_v2"), content)
	if err != nil {
		return
	}

	req.Header.Set("Dropbox-API-Arg", string(b))
	req.Header.Set("Content-Type", "application/octet-stream")
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
			var apiError UploadSessionAppendV2ApiError
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
	return
}

type UploadSessionFinishApiError struct {
	dropbox.ApiError
	EndpointError *UploadSessionFinishError `json:"error"`
}

func (dbx *apiImpl) UploadSessionFinish(arg *UploadSessionFinishArg, content io.Reader) (res *FileMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("content", "files", "upload_session/finish"), content)
	if err != nil {
		return
	}

	req.Header.Set("Dropbox-API-Arg", string(b))
	req.Header.Set("Content-Type", "application/octet-stream")
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
			var apiError UploadSessionFinishApiError
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

type UploadSessionStartApiError struct {
	dropbox.ApiError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) UploadSessionStart(arg *UploadSessionStartArg, content io.Reader) (res *UploadSessionStartResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("content", "files", "upload_session/start"), content)
	if err != nil {
		return
	}

	req.Header.Set("Dropbox-API-Arg", string(b))
	req.Header.Set("Content-Type", "application/octet-stream")
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
			var apiError UploadSessionStartApiError
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
