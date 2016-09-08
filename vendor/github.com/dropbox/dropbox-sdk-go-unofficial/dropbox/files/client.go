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

package files

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/async"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/properties"
)

// Client interface describes all routes in this namespace
type Client interface {
	// AlphaGetMetadata : Returns the metadata for a file or folder. This is an
	// alpha endpoint compatible with the properties API. Note: Metadata for the
	// root folder is unsupported.
	AlphaGetMetadata(arg *AlphaGetMetadataArg) (res IsMetadata, err error)
	// AlphaUpload : Create a new file with the contents provided in the
	// request. Note that this endpoint is part of the properties API alpha and
	// is slightly different from `upload`. Do not use this to upload a file
	// larger than 150 MB. Instead, create an upload session with
	// `uploadSessionStart`.
	AlphaUpload(arg *CommitInfoWithProperties, content io.Reader) (res *FileMetadata, err error)
	// Copy : Copy a file or folder to a different location in the user's
	// Dropbox. If the source path is a folder all its contents will be copied.
	Copy(arg *RelocationArg) (res IsMetadata, err error)
	// CopyReferenceGet : Get a copy reference to a file or folder. This
	// reference string can be used to save that file or folder to another
	// user's Dropbox by passing it to `copyReferenceSave`.
	CopyReferenceGet(arg *GetCopyReferenceArg) (res *GetCopyReferenceResult, err error)
	// CopyReferenceSave : Save a copy reference returned by `copyReferenceGet`
	// to the user's Dropbox.
	CopyReferenceSave(arg *SaveCopyReferenceArg) (res *SaveCopyReferenceResult, err error)
	// CreateFolder : Create a folder at a given path.
	CreateFolder(arg *CreateFolderArg) (res *FolderMetadata, err error)
	// Delete : Delete the file or folder at a given path. If the path is a
	// folder, all its contents will be deleted too. A successful response
	// indicates that the file or folder was deleted. The returned metadata will
	// be the corresponding `FileMetadata` or `FolderMetadata` for the item at
	// time of deletion, and not a `DeletedMetadata` object.
	Delete(arg *DeleteArg) (res IsMetadata, err error)
	// Download : Download a file from a user's Dropbox.
	Download(arg *DownloadArg) (res *FileMetadata, content io.ReadCloser, err error)
	// GetMetadata : Returns the metadata for a file or folder. Note: Metadata
	// for the root folder is unsupported.
	GetMetadata(arg *GetMetadataArg) (res IsMetadata, err error)
	// GetPreview : Get a preview for a file. Currently previews are only
	// generated for the files with  the following extensions: .doc, .docx,
	// .docm, .ppt, .pps, .ppsx, .ppsm, .pptx, .pptm,  .xls, .xlsx, .xlsm, .rtf
	GetPreview(arg *PreviewArg) (res *FileMetadata, content io.ReadCloser, err error)
	// GetTemporaryLink : Get a temporary link to stream content of a file. This
	// link will expire in four hours and afterwards you will get 410 Gone.
	// Content-Type of the link is determined automatically by the file's mime
	// type.
	GetTemporaryLink(arg *GetTemporaryLinkArg) (res *GetTemporaryLinkResult, err error)
	// GetThumbnail : Get a thumbnail for an image. This method currently
	// supports files with the following file extensions: jpg, jpeg, png, tiff,
	// tif, gif and bmp. Photos that are larger than 20MB in size won't be
	// converted to a thumbnail.
	GetThumbnail(arg *ThumbnailArg) (res *FileMetadata, content io.ReadCloser, err error)
	// ListFolder : Returns the contents of a folder.
	ListFolder(arg *ListFolderArg) (res *ListFolderResult, err error)
	// ListFolderContinue : Once a cursor has been retrieved from `listFolder`,
	// use this to paginate through all files and retrieve updates to the
	// folder.
	ListFolderContinue(arg *ListFolderContinueArg) (res *ListFolderResult, err error)
	// ListFolderGetLatestCursor : A way to quickly get a cursor for the
	// folder's state. Unlike `listFolder`, `listFolderGetLatestCursor` doesn't
	// return any entries. This endpoint is for app which only needs to know
	// about new files and modifications and doesn't need to know about files
	// that already exist in Dropbox.
	ListFolderGetLatestCursor(arg *ListFolderArg) (res *ListFolderGetLatestCursorResult, err error)
	// ListFolderLongpoll : A longpoll endpoint to wait for changes on an
	// account. In conjunction with `listFolderContinue`, this call gives you a
	// low-latency way to monitor an account for file changes. The connection
	// will block until there are changes available or a timeout occurs. This
	// endpoint is useful mostly for client-side apps. If you're looking for
	// server-side notifications, check out our `webhooks documentation`
	// <https://www.dropbox.com/developers/reference/webhooks>.
	ListFolderLongpoll(arg *ListFolderLongpollArg) (res *ListFolderLongpollResult, err error)
	// ListRevisions : Return revisions of a file
	ListRevisions(arg *ListRevisionsArg) (res *ListRevisionsResult, err error)
	// Move : Move a file or folder to a different location in the user's
	// Dropbox. If the source path is a folder all its contents will be moved.
	Move(arg *RelocationArg) (res IsMetadata, err error)
	// PermanentlyDelete : Permanently delete the file or folder at a given path
	// (see https://www.dropbox.com/en/help/40). Note: This endpoint is only
	// available for Dropbox Business apps.
	PermanentlyDelete(arg *DeleteArg) (err error)
	// PropertiesAdd : Add custom properties to a file using a filled property
	// template. See properties/template/add to create new property templates.
	PropertiesAdd(arg *PropertyGroupWithPath) (err error)
	// PropertiesOverwrite : Overwrite custom properties from a specified
	// template associated with a file.
	PropertiesOverwrite(arg *PropertyGroupWithPath) (err error)
	// PropertiesRemove : Remove all custom properties from a specified template
	// associated with a file. To remove specific property key value pairs, see
	// `propertiesUpdate`. To update a property template, see
	// properties/template/update. Property templates can't be removed once
	// created.
	PropertiesRemove(arg *RemovePropertiesArg) (err error)
	// PropertiesTemplateGet : Get the schema for a specified template.
	PropertiesTemplateGet(arg *properties.GetPropertyTemplateArg) (res *properties.GetPropertyTemplateResult, err error)
	// PropertiesTemplateList : Get the property template identifiers for a
	// user. To get the schema of each template use `propertiesTemplateGet`.
	PropertiesTemplateList() (res *properties.ListPropertyTemplateIds, err error)
	// PropertiesUpdate : Add, update or remove custom properties from a
	// specified template associated with a file. Fields that already exist and
	// not described in the request will not be modified.
	PropertiesUpdate(arg *UpdatePropertyGroupArg) (err error)
	// Restore : Restore a file to a specific revision
	Restore(arg *RestoreArg) (res *FileMetadata, err error)
	// SaveUrl : Save a specified URL into a file in user's Dropbox. If the
	// given path already exists, the file will be renamed to avoid the conflict
	// (e.g. myfile (1).txt).
	SaveUrl(arg *SaveUrlArg) (res *SaveUrlResult, err error)
	// SaveUrlCheckJobStatus : Check the status of a `saveUrl` job.
	SaveUrlCheckJobStatus(arg *async.PollArg) (res *SaveUrlJobStatus, err error)
	// Search : Searches for files and folders. Note: Recent changes may not
	// immediately be reflected in search results due to a short delay in
	// indexing.
	Search(arg *SearchArg) (res *SearchResult, err error)
	// Upload : Create a new file with the contents provided in the request. Do
	// not use this to upload a file larger than 150 MB. Instead, create an
	// upload session with `uploadSessionStart`.
	Upload(arg *CommitInfo, content io.Reader) (res *FileMetadata, err error)
	// UploadSessionAppend : Append more data to an upload session. A single
	// request should not upload more than 150 MB of file contents.
	UploadSessionAppend(arg *UploadSessionCursor, content io.Reader) (err error)
	// UploadSessionAppendV2 : Append more data to an upload session. When the
	// parameter close is set, this call will close the session. A single
	// request should not upload more than 150 MB of file contents.
	UploadSessionAppendV2(arg *UploadSessionAppendArg, content io.Reader) (err error)
	// UploadSessionFinish : Finish an upload session and save the uploaded data
	// to the given file path. A single request should not upload more than 150
	// MB of file contents.
	UploadSessionFinish(arg *UploadSessionFinishArg, content io.Reader) (res *FileMetadata, err error)
	// UploadSessionFinishBatch : This route helps you commit many files at once
	// into a user's Dropbox. Use `uploadSessionStart` and
	// `uploadSessionAppendV2` to upload file contents. We recommend uploading
	// many files in parallel to increase throughput. Once the file contents
	// have been uploaded, rather than calling `uploadSessionFinish`, use this
	// route to finish all your upload sessions in a single request.
	// `UploadSessionStartArg.close` or `UploadSessionAppendArg.close` needs to
	// be true for last `uploadSessionStart` or `uploadSessionAppendV2` call.
	// This route will return job_id immediately and do the async commit job in
	// background. We have another route `uploadSessionFinishBatchCheck` to
	// check the job status. For the same account, this route should be executed
	// serially. That means you should not start next job before current job
	// finishes. Also we only allow up to 1000 entries in a single request
	UploadSessionFinishBatch(arg *UploadSessionFinishBatchArg) (res *async.LaunchEmptyResult, err error)
	// UploadSessionFinishBatchCheck : Returns the status of an asynchronous job
	// for `uploadSessionFinishBatch`. If success, it returns list of result for
	// each entry
	UploadSessionFinishBatchCheck(arg *async.PollArg) (res *UploadSessionFinishBatchJobStatus, err error)
	// UploadSessionStart : Upload sessions allow you to upload a single file
	// using multiple requests. This call starts a new upload session with the
	// given data.  You can then use `uploadSessionAppendV2` to add more data
	// and `uploadSessionFinish` to save all the data to a file in Dropbox. A
	// single request should not upload more than 150 MB of file contents.
	UploadSessionStart(arg *UploadSessionStartArg, content io.Reader) (res *UploadSessionStartResult, err error)
}

type apiImpl dropbox.Context

//AlphaGetMetadataAPIError is an error-wrapper for the alpha/get_metadata route
type AlphaGetMetadataAPIError struct {
	dropbox.APIError
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
	if resp.StatusCode == http.StatusConflict {
		var apiError AlphaGetMetadataAPIError
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

//AlphaUploadAPIError is an error-wrapper for the alpha/upload route
type AlphaUploadAPIError struct {
	dropbox.APIError
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError AlphaUploadAPIError
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

//CopyAPIError is an error-wrapper for the copy route
type CopyAPIError struct {
	dropbox.APIError
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
	if resp.StatusCode == http.StatusConflict {
		var apiError CopyAPIError
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

//CopyReferenceGetAPIError is an error-wrapper for the copy_reference/get route
type CopyReferenceGetAPIError struct {
	dropbox.APIError
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError CopyReferenceGetAPIError
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

//CopyReferenceSaveAPIError is an error-wrapper for the copy_reference/save route
type CopyReferenceSaveAPIError struct {
	dropbox.APIError
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError CopyReferenceSaveAPIError
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

//CreateFolderAPIError is an error-wrapper for the create_folder route
type CreateFolderAPIError struct {
	dropbox.APIError
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError CreateFolderAPIError
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

//DeleteAPIError is an error-wrapper for the delete route
type DeleteAPIError struct {
	dropbox.APIError
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
	if resp.StatusCode == http.StatusConflict {
		var apiError DeleteAPIError
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

//DownloadAPIError is an error-wrapper for the download route
type DownloadAPIError struct {
	dropbox.APIError
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

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("content", "files", "download"), nil)
	if err != nil {
		return
	}

	req.Header.Set("Dropbox-API-Arg", string(b))
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

	body := []byte(resp.Header.Get("Dropbox-API-Result"))
	content = resp.Body
	if dbx.Config.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError DownloadAPIError
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

//GetMetadataAPIError is an error-wrapper for the get_metadata route
type GetMetadataAPIError struct {
	dropbox.APIError
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
	if resp.StatusCode == http.StatusConflict {
		var apiError GetMetadataAPIError
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

//GetPreviewAPIError is an error-wrapper for the get_preview route
type GetPreviewAPIError struct {
	dropbox.APIError
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

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("content", "files", "get_preview"), nil)
	if err != nil {
		return
	}

	req.Header.Set("Dropbox-API-Arg", string(b))
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

	body := []byte(resp.Header.Get("Dropbox-API-Result"))
	content = resp.Body
	if dbx.Config.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError GetPreviewAPIError
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

//GetTemporaryLinkAPIError is an error-wrapper for the get_temporary_link route
type GetTemporaryLinkAPIError struct {
	dropbox.APIError
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError GetTemporaryLinkAPIError
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

//GetThumbnailAPIError is an error-wrapper for the get_thumbnail route
type GetThumbnailAPIError struct {
	dropbox.APIError
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

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("content", "files", "get_thumbnail"), nil)
	if err != nil {
		return
	}

	req.Header.Set("Dropbox-API-Arg", string(b))
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

	body := []byte(resp.Header.Get("Dropbox-API-Result"))
	content = resp.Body
	if dbx.Config.Verbose {
		log.Printf("body: %s", body)
	}
	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError GetThumbnailAPIError
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

//ListFolderAPIError is an error-wrapper for the list_folder route
type ListFolderAPIError struct {
	dropbox.APIError
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError ListFolderAPIError
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

//ListFolderContinueAPIError is an error-wrapper for the list_folder/continue route
type ListFolderContinueAPIError struct {
	dropbox.APIError
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError ListFolderContinueAPIError
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

//ListFolderGetLatestCursorAPIError is an error-wrapper for the list_folder/get_latest_cursor route
type ListFolderGetLatestCursorAPIError struct {
	dropbox.APIError
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError ListFolderGetLatestCursorAPIError
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

//ListFolderLongpollAPIError is an error-wrapper for the list_folder/longpoll route
type ListFolderLongpollAPIError struct {
	dropbox.APIError
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
	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError ListFolderLongpollAPIError
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

//ListRevisionsAPIError is an error-wrapper for the list_revisions route
type ListRevisionsAPIError struct {
	dropbox.APIError
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError ListRevisionsAPIError
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

//MoveAPIError is an error-wrapper for the move route
type MoveAPIError struct {
	dropbox.APIError
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
	if resp.StatusCode == http.StatusConflict {
		var apiError MoveAPIError
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

//PermanentlyDeleteAPIError is an error-wrapper for the permanently_delete route
type PermanentlyDeleteAPIError struct {
	dropbox.APIError
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
		var apiError PermanentlyDeleteAPIError
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

//PropertiesAddAPIError is an error-wrapper for the properties/add route
type PropertiesAddAPIError struct {
	dropbox.APIError
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
		var apiError PropertiesAddAPIError
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

//PropertiesOverwriteAPIError is an error-wrapper for the properties/overwrite route
type PropertiesOverwriteAPIError struct {
	dropbox.APIError
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
		var apiError PropertiesOverwriteAPIError
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

//PropertiesRemoveAPIError is an error-wrapper for the properties/remove route
type PropertiesRemoveAPIError struct {
	dropbox.APIError
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
		var apiError PropertiesRemoveAPIError
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

//PropertiesTemplateGetAPIError is an error-wrapper for the properties/template/get route
type PropertiesTemplateGetAPIError struct {
	dropbox.APIError
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError PropertiesTemplateGetAPIError
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

//PropertiesTemplateListAPIError is an error-wrapper for the properties/template/list route
type PropertiesTemplateListAPIError struct {
	dropbox.APIError
	EndpointError *properties.PropertyTemplateError `json:"error"`
}

func (dbx *apiImpl) PropertiesTemplateList() (res *properties.ListPropertyTemplateIds, err error) {
	cli := dbx.Client

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "properties/template/list"), nil)
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError PropertiesTemplateListAPIError
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

//PropertiesUpdateAPIError is an error-wrapper for the properties/update route
type PropertiesUpdateAPIError struct {
	dropbox.APIError
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
		var apiError PropertiesUpdateAPIError
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

//RestoreAPIError is an error-wrapper for the restore route
type RestoreAPIError struct {
	dropbox.APIError
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError RestoreAPIError
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

//SaveUrlAPIError is an error-wrapper for the save_url route
type SaveUrlAPIError struct {
	dropbox.APIError
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError SaveUrlAPIError
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

//SaveUrlCheckJobStatusAPIError is an error-wrapper for the save_url/check_job_status route
type SaveUrlCheckJobStatusAPIError struct {
	dropbox.APIError
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError SaveUrlCheckJobStatusAPIError
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

//SearchAPIError is an error-wrapper for the search route
type SearchAPIError struct {
	dropbox.APIError
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError SearchAPIError
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

//UploadAPIError is an error-wrapper for the upload route
type UploadAPIError struct {
	dropbox.APIError
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError UploadAPIError
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

//UploadSessionAppendAPIError is an error-wrapper for the upload_session/append route
type UploadSessionAppendAPIError struct {
	dropbox.APIError
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
		var apiError UploadSessionAppendAPIError
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

//UploadSessionAppendV2APIError is an error-wrapper for the upload_session/append_v2 route
type UploadSessionAppendV2APIError struct {
	dropbox.APIError
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
		var apiError UploadSessionAppendV2APIError
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

//UploadSessionFinishAPIError is an error-wrapper for the upload_session/finish route
type UploadSessionFinishAPIError struct {
	dropbox.APIError
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError UploadSessionFinishAPIError
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

//UploadSessionFinishBatchAPIError is an error-wrapper for the upload_session/finish_batch route
type UploadSessionFinishBatchAPIError struct {
	dropbox.APIError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) UploadSessionFinishBatch(arg *UploadSessionFinishBatchArg) (res *async.LaunchEmptyResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "upload_session/finish_batch"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError UploadSessionFinishBatchAPIError
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

//UploadSessionFinishBatchCheckAPIError is an error-wrapper for the upload_session/finish_batch/check route
type UploadSessionFinishBatchCheckAPIError struct {
	dropbox.APIError
	EndpointError *async.PollError `json:"error"`
}

func (dbx *apiImpl) UploadSessionFinishBatchCheck(arg *async.PollArg) (res *UploadSessionFinishBatchJobStatus, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "files", "upload_session/finish_batch/check"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError UploadSessionFinishBatchCheckAPIError
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

//UploadSessionStartAPIError is an error-wrapper for the upload_session/start route
type UploadSessionStartAPIError struct {
	dropbox.APIError
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
		err = json.Unmarshal(body, &res)
		if err != nil {
			return
		}

		return
	}
	if resp.StatusCode == http.StatusConflict {
		var apiError UploadSessionStartAPIError
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
