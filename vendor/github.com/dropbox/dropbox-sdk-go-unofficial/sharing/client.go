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

// This namespace contains endpoints and data types for creating and managing
// shared links and shared folders.
package sharing

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	dropbox "github.com/dropbox/dropbox-sdk-go-unofficial"
	"github.com/dropbox/dropbox-sdk-go-unofficial/async"
)

type Client interface {
	// Adds specified members to a file.
	AddFileMember(arg *AddFileMemberArgs) (res []*FileMemberActionResult, err error)
	// Allows an owner or editor (if the ACL update policy allows) of a shared
	// folder to add another member. For the new member to get access to all the
	// functionality for this folder, you will need to call `mountFolder` on
	// their behalf. Apps must have full Dropbox access to use this endpoint.
	AddFolderMember(arg *AddFolderMemberArg) (err error)
	// Returns the status of an asynchronous job. Apps must have full Dropbox
	// access to use this endpoint.
	CheckJobStatus(arg *async.PollArg) (res *JobStatus, err error)
	// Returns the status of an asynchronous job for sharing a folder. Apps must
	// have full Dropbox access to use this endpoint.
	CheckRemoveMemberJobStatus(arg *async.PollArg) (res *RemoveMemberJobStatus, err error)
	// Returns the status of an asynchronous job for sharing a folder. Apps must
	// have full Dropbox access to use this endpoint.
	CheckShareJobStatus(arg *async.PollArg) (res *ShareFolderJobStatus, err error)
	// Create a shared link. If a shared link already exists for the given path,
	// that link is returned. Note that in the returned `PathLinkMetadata`, the
	// `PathLinkMetadata.url` field is the shortened URL if
	// `CreateSharedLinkArg.short_url` argument is set to `True`. Previously, it
	// was technically possible to break a shared link by moving or renaming the
	// corresponding file or folder. In the future, this will no longer be the
	// case, so your app shouldn't rely on this behavior. Instead, if your app
	// needs to revoke a shared link, use `revokeSharedLink`.
	CreateSharedLink(arg *CreateSharedLinkArg) (res *PathLinkMetadata, err error)
	// Create a shared link with custom settings. If no settings are given then
	// the default visibility is `RequestedVisibility.public` (The resolved
	// visibility, though, may depend on other aspects such as team and shared
	// folder settings).
	CreateSharedLinkWithSettings(arg *CreateSharedLinkWithSettingsArg) (res IsSharedLinkMetadata, err error)
	// Returns shared file metadata.
	GetFileMetadata(arg *GetFileMetadataArg) (res *SharedFileMetadata, err error)
	// Returns shared file metadata.
	GetFileMetadataBatch(arg *GetFileMetadataBatchArg) (res []*GetFileMetadataBatchResult, err error)
	// Returns shared folder metadata by its folder ID. Apps must have full
	// Dropbox access to use this endpoint.
	GetFolderMetadata(arg *GetMetadataArgs) (res *SharedFolderMetadata, err error)
	// Download the shared link's file from a user's Dropbox.
	GetSharedLinkFile(arg *GetSharedLinkMetadataArg) (res IsSharedLinkMetadata, content io.ReadCloser, err error)
	// Get the shared link's metadata.
	GetSharedLinkMetadata(arg *GetSharedLinkMetadataArg) (res IsSharedLinkMetadata, err error)
	// Returns a list of `LinkMetadata` objects for this user, including
	// collection links. If no path is given or the path is empty, returns a
	// list of all shared links for the current user, including collection
	// links. If a non-empty path is given, returns a list of all shared links
	// that allow access to the given path.  Collection links are never returned
	// in this case. Note that the url field in the response is never the
	// shortened URL.
	GetSharedLinks(arg *GetSharedLinksArg) (res *GetSharedLinksResult, err error)
	// Use to obtain the members who have been invited to a file, both inherited
	// and uninherited members.
	ListFileMembers(arg *ListFileMembersArg) (res *SharedFileMembers, err error)
	// Get members of multiple files at once. The arguments to this route are
	// more limited, and the limit on query result size per file is more strict.
	// To customize the results more, use the individual file endpoint.
	// Inherited users are not included in the result, and permissions are not
	// returned for this endpoint.
	ListFileMembersBatch(arg *ListFileMembersBatchArg) (res []*ListFileMembersBatchResult, err error)
	// Once a cursor has been retrieved from `listFileMembers` or
	// `listFileMembersBatch`, use this to paginate through all shared file
	// members.
	ListFileMembersContinue(arg *ListFileMembersContinueArg) (res *SharedFileMembers, err error)
	// Returns shared folder membership by its folder ID. Apps must have full
	// Dropbox access to use this endpoint.
	ListFolderMembers(arg *ListFolderMembersArgs) (res *SharedFolderMembers, err error)
	// Once a cursor has been retrieved from `listFolderMembers`, use this to
	// paginate through all shared folder members. Apps must have full Dropbox
	// access to use this endpoint.
	ListFolderMembersContinue(arg *ListFolderMembersContinueArg) (res *SharedFolderMembers, err error)
	// Return the list of all shared folders the current user has access to.
	// Apps must have full Dropbox access to use this endpoint.
	ListFolders(arg *ListFoldersArgs) (res *ListFoldersResult, err error)
	// Once a cursor has been retrieved from `listFolders`, use this to paginate
	// through all shared folders. The cursor must come from a previous call to
	// `listFolders` or `listFoldersContinue`. Apps must have full Dropbox
	// access to use this endpoint.
	ListFoldersContinue(arg *ListFoldersContinueArg) (res *ListFoldersResult, err error)
	// Return the list of all shared folders the current user can mount or
	// unmount. Apps must have full Dropbox access to use this endpoint.
	ListMountableFolders(arg *ListFoldersArgs) (res *ListFoldersResult, err error)
	// Once a cursor has been retrieved from `listMountableFolders`, use this to
	// paginate through all mountable shared folders. The cursor must come from
	// a previous call to `listMountableFolders` or
	// `listMountableFoldersContinue`. Apps must have full Dropbox access to use
	// this endpoint.
	ListMountableFoldersContinue(arg *ListFoldersContinueArg) (res *ListFoldersResult, err error)
	// Returns a list of all files shared with current user.  Does not include
	// files the user has received via shared folders, and does  not include
	// unclaimed invitations.
	ListReceivedFiles(arg *ListFilesArg) (res *ListFilesResult, err error)
	// Get more results with a cursor from `listReceivedFiles`.
	ListReceivedFilesContinue(arg *ListFilesContinueArg) (res *ListFilesResult, err error)
	// List shared links of this user. If no path is given or the path is empty,
	// returns a list of all shared links for the current user. If a non-empty
	// path is given, returns a list of all shared links that allow access to
	// the given path - direct links to the given path and links to parent
	// folders of the given path. Links to parent folders can be suppressed by
	// setting direct_only to true.
	ListSharedLinks(arg *ListSharedLinksArg) (res *ListSharedLinksResult, err error)
	// Modify the shared link's settings. If the requested visibility conflict
	// with the shared links policy of the team or the shared folder (in case
	// the linked file is part of a shared folder) then the
	// `LinkPermissions.resolved_visibility` of the returned
	// `SharedLinkMetadata` will reflect the actual visibility of the shared
	// link and the `LinkPermissions.requested_visibility` will reflect the
	// requested visibility.
	ModifySharedLinkSettings(arg *ModifySharedLinkSettingsArgs) (res IsSharedLinkMetadata, err error)
	// The current user mounts the designated folder. Mount a shared folder for
	// a user after they have been added as a member. Once mounted, the shared
	// folder will appear in their Dropbox. Apps must have full Dropbox access
	// to use this endpoint.
	MountFolder(arg *MountFolderArg) (res *SharedFolderMetadata, err error)
	// The current user relinquishes their membership in the designated file.
	// Note that the current user may still have inherited access to this file
	// through the parent folder. Apps must have full Dropbox access to use this
	// endpoint.
	RelinquishFileMembership(arg *RelinquishFileMembershipArg) (err error)
	// The current user relinquishes their membership in the designated shared
	// folder and will no longer have access to the folder.  A folder owner
	// cannot relinquish membership in their own folder. This will run
	// synchronously if leave_a_copy is false, and asynchronously if
	// leave_a_copy is true. Apps must have full Dropbox access to use this
	// endpoint.
	RelinquishFolderMembership(arg *RelinquishFolderMembershipArg) (res *async.LaunchEmptyResult, err error)
	// Identical to remove_file_member_2 but with less information returned.
	RemoveFileMember(arg *RemoveFileMemberArg) (res *FileMemberActionIndividualResult, err error)
	// Removes a specified member from the file.
	RemoveFileMember2(arg *RemoveFileMemberArg) (res *FileMemberRemoveActionResult, err error)
	// Allows an owner or editor (if the ACL update policy allows) of a shared
	// folder to remove another member. Apps must have full Dropbox access to
	// use this endpoint.
	RemoveFolderMember(arg *RemoveFolderMemberArg) (res *async.LaunchResultBase, err error)
	// Revoke a shared link. Note that even after revoking a shared link to a
	// file, the file may be accessible if there are shared links leading to any
	// of the file parent folders. To list all shared links that enable access
	// to a specific file, you can use the `listSharedLinks` with the file as
	// the `ListSharedLinksArg.path` argument.
	RevokeSharedLink(arg *RevokeSharedLinkArg) (err error)
	// Share a folder with collaborators. Most sharing will be completed
	// synchronously. Large folders will be completed asynchronously. To make
	// testing the async case repeatable, set `ShareFolderArg.force_async`. If a
	// `ShareFolderLaunch.async_job_id` is returned, you'll need to call
	// `checkShareJobStatus` until the action completes to get the metadata for
	// the folder. Apps must have full Dropbox access to use this endpoint.
	ShareFolder(arg *ShareFolderArg) (res *ShareFolderLaunch, err error)
	// Transfer ownership of a shared folder to a member of the shared folder.
	// User must have `AccessLevel.owner` access to the shared folder to perform
	// a transfer. Apps must have full Dropbox access to use this endpoint.
	TransferFolder(arg *TransferFolderArg) (err error)
	// The current user unmounts the designated folder. They can re-mount the
	// folder at a later time using `mountFolder`. Apps must have full Dropbox
	// access to use this endpoint.
	UnmountFolder(arg *UnmountFolderArg) (err error)
	// Remove all members from this file. Does not remove inherited members.
	UnshareFile(arg *UnshareFileArg) (err error)
	// Allows a shared folder owner to unshare the folder. You'll need to call
	// `checkJobStatus` to determine if the action has completed successfully.
	// Apps must have full Dropbox access to use this endpoint.
	UnshareFolder(arg *UnshareFolderArg) (res *async.LaunchEmptyResult, err error)
	// Allows an owner or editor of a shared folder to update another member's
	// permissions. Apps must have full Dropbox access to use this endpoint.
	UpdateFolderMember(arg *UpdateFolderMemberArg) (res *MemberAccessLevelResult, err error)
	// Update the sharing policies for a shared folder. User must have
	// `AccessLevel.owner` access to the shared folder to update its policies.
	// Apps must have full Dropbox access to use this endpoint.
	UpdateFolderPolicy(arg *UpdateFolderPolicyArg) (res *SharedFolderMetadata, err error)
}

type apiImpl dropbox.Context
type AddFileMemberApiError struct {
	dropbox.ApiError
	EndpointError *AddFileMemberError `json:"error"`
}

func (dbx *apiImpl) AddFileMember(arg *AddFileMemberArgs) (res []*FileMemberActionResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "add_file_member"), bytes.NewReader(b))
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
			var apiError AddFileMemberApiError
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

type AddFolderMemberApiError struct {
	dropbox.ApiError
	EndpointError *AddFolderMemberError `json:"error"`
}

func (dbx *apiImpl) AddFolderMember(arg *AddFolderMemberArg) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "add_folder_member"), bytes.NewReader(b))
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
			var apiError AddFolderMemberApiError
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

type CheckJobStatusApiError struct {
	dropbox.ApiError
	EndpointError *async.PollError `json:"error"`
}

func (dbx *apiImpl) CheckJobStatus(arg *async.PollArg) (res *JobStatus, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "check_job_status"), bytes.NewReader(b))
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
			var apiError CheckJobStatusApiError
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

type CheckRemoveMemberJobStatusApiError struct {
	dropbox.ApiError
	EndpointError *async.PollError `json:"error"`
}

func (dbx *apiImpl) CheckRemoveMemberJobStatus(arg *async.PollArg) (res *RemoveMemberJobStatus, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "check_remove_member_job_status"), bytes.NewReader(b))
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
			var apiError CheckRemoveMemberJobStatusApiError
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

type CheckShareJobStatusApiError struct {
	dropbox.ApiError
	EndpointError *async.PollError `json:"error"`
}

func (dbx *apiImpl) CheckShareJobStatus(arg *async.PollArg) (res *ShareFolderJobStatus, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "check_share_job_status"), bytes.NewReader(b))
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
			var apiError CheckShareJobStatusApiError
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

type CreateSharedLinkApiError struct {
	dropbox.ApiError
	EndpointError *CreateSharedLinkError `json:"error"`
}

func (dbx *apiImpl) CreateSharedLink(arg *CreateSharedLinkArg) (res *PathLinkMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "create_shared_link"), bytes.NewReader(b))
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
			var apiError CreateSharedLinkApiError
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

type CreateSharedLinkWithSettingsApiError struct {
	dropbox.ApiError
	EndpointError *CreateSharedLinkWithSettingsError `json:"error"`
}

func (dbx *apiImpl) CreateSharedLinkWithSettings(arg *CreateSharedLinkWithSettingsArg) (res IsSharedLinkMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "create_shared_link_with_settings"), bytes.NewReader(b))
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
			var apiError CreateSharedLinkWithSettingsApiError
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
	var tmp sharedLinkMetadataUnion
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return
	}
	switch tmp.Tag {
	case "file":
		res = tmp.File

	case "folder":
		res = tmp.Folder

	}
	return
}

type GetFileMetadataApiError struct {
	dropbox.ApiError
	EndpointError *GetFileMetadataError `json:"error"`
}

func (dbx *apiImpl) GetFileMetadata(arg *GetFileMetadataArg) (res *SharedFileMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "get_file_metadata"), bytes.NewReader(b))
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
			var apiError GetFileMetadataApiError
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

type GetFileMetadataBatchApiError struct {
	dropbox.ApiError
	EndpointError *SharingUserError `json:"error"`
}

func (dbx *apiImpl) GetFileMetadataBatch(arg *GetFileMetadataBatchArg) (res []*GetFileMetadataBatchResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "get_file_metadata/batch"), bytes.NewReader(b))
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
			var apiError GetFileMetadataBatchApiError
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

type GetFolderMetadataApiError struct {
	dropbox.ApiError
	EndpointError *SharedFolderAccessError `json:"error"`
}

func (dbx *apiImpl) GetFolderMetadata(arg *GetMetadataArgs) (res *SharedFolderMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "get_folder_metadata"), bytes.NewReader(b))
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
			var apiError GetFolderMetadataApiError
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

type GetSharedLinkFileApiError struct {
	dropbox.ApiError
	EndpointError *GetSharedLinkFileError `json:"error"`
}

func (dbx *apiImpl) GetSharedLinkFile(arg *GetSharedLinkMetadataArg) (res IsSharedLinkMetadata, content io.ReadCloser, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("content", "sharing", "get_shared_link_file"), bytes.NewReader(b))
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
			var apiError GetSharedLinkFileApiError
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
	var tmp sharedLinkMetadataUnion
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return
	}
	switch tmp.Tag {
	case "file":
		res = tmp.File

	case "folder":
		res = tmp.Folder

	}
	return
}

type GetSharedLinkMetadataApiError struct {
	dropbox.ApiError
	EndpointError *SharedLinkError `json:"error"`
}

func (dbx *apiImpl) GetSharedLinkMetadata(arg *GetSharedLinkMetadataArg) (res IsSharedLinkMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "get_shared_link_metadata"), bytes.NewReader(b))
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
			var apiError GetSharedLinkMetadataApiError
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
	var tmp sharedLinkMetadataUnion
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return
	}
	switch tmp.Tag {
	case "file":
		res = tmp.File

	case "folder":
		res = tmp.Folder

	}
	return
}

type GetSharedLinksApiError struct {
	dropbox.ApiError
	EndpointError *GetSharedLinksError `json:"error"`
}

func (dbx *apiImpl) GetSharedLinks(arg *GetSharedLinksArg) (res *GetSharedLinksResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "get_shared_links"), bytes.NewReader(b))
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
			var apiError GetSharedLinksApiError
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

type ListFileMembersApiError struct {
	dropbox.ApiError
	EndpointError *ListFileMembersError `json:"error"`
}

func (dbx *apiImpl) ListFileMembers(arg *ListFileMembersArg) (res *SharedFileMembers, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "list_file_members"), bytes.NewReader(b))
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
			var apiError ListFileMembersApiError
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

type ListFileMembersBatchApiError struct {
	dropbox.ApiError
	EndpointError *SharingUserError `json:"error"`
}

func (dbx *apiImpl) ListFileMembersBatch(arg *ListFileMembersBatchArg) (res []*ListFileMembersBatchResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "list_file_members/batch"), bytes.NewReader(b))
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
			var apiError ListFileMembersBatchApiError
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

type ListFileMembersContinueApiError struct {
	dropbox.ApiError
	EndpointError *ListFileMembersContinueError `json:"error"`
}

func (dbx *apiImpl) ListFileMembersContinue(arg *ListFileMembersContinueArg) (res *SharedFileMembers, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "list_file_members/continue"), bytes.NewReader(b))
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
			var apiError ListFileMembersContinueApiError
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

type ListFolderMembersApiError struct {
	dropbox.ApiError
	EndpointError *SharedFolderAccessError `json:"error"`
}

func (dbx *apiImpl) ListFolderMembers(arg *ListFolderMembersArgs) (res *SharedFolderMembers, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "list_folder_members"), bytes.NewReader(b))
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
			var apiError ListFolderMembersApiError
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

type ListFolderMembersContinueApiError struct {
	dropbox.ApiError
	EndpointError *ListFolderMembersContinueError `json:"error"`
}

func (dbx *apiImpl) ListFolderMembersContinue(arg *ListFolderMembersContinueArg) (res *SharedFolderMembers, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "list_folder_members/continue"), bytes.NewReader(b))
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
			var apiError ListFolderMembersContinueApiError
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

type ListFoldersApiError struct {
	dropbox.ApiError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) ListFolders(arg *ListFoldersArgs) (res *ListFoldersResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "list_folders"), bytes.NewReader(b))
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
			var apiError ListFoldersApiError
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

type ListFoldersContinueApiError struct {
	dropbox.ApiError
	EndpointError *ListFoldersContinueError `json:"error"`
}

func (dbx *apiImpl) ListFoldersContinue(arg *ListFoldersContinueArg) (res *ListFoldersResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "list_folders/continue"), bytes.NewReader(b))
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
			var apiError ListFoldersContinueApiError
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

type ListMountableFoldersApiError struct {
	dropbox.ApiError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) ListMountableFolders(arg *ListFoldersArgs) (res *ListFoldersResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "list_mountable_folders"), bytes.NewReader(b))
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
			var apiError ListMountableFoldersApiError
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

type ListMountableFoldersContinueApiError struct {
	dropbox.ApiError
	EndpointError *ListFoldersContinueError `json:"error"`
}

func (dbx *apiImpl) ListMountableFoldersContinue(arg *ListFoldersContinueArg) (res *ListFoldersResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "list_mountable_folders/continue"), bytes.NewReader(b))
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
			var apiError ListMountableFoldersContinueApiError
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

type ListReceivedFilesApiError struct {
	dropbox.ApiError
	EndpointError *SharingUserError `json:"error"`
}

func (dbx *apiImpl) ListReceivedFiles(arg *ListFilesArg) (res *ListFilesResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "list_received_files"), bytes.NewReader(b))
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
			var apiError ListReceivedFilesApiError
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

type ListReceivedFilesContinueApiError struct {
	dropbox.ApiError
	EndpointError *ListFilesContinueError `json:"error"`
}

func (dbx *apiImpl) ListReceivedFilesContinue(arg *ListFilesContinueArg) (res *ListFilesResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "list_received_files/continue"), bytes.NewReader(b))
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
			var apiError ListReceivedFilesContinueApiError
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

type ListSharedLinksApiError struct {
	dropbox.ApiError
	EndpointError *ListSharedLinksError `json:"error"`
}

func (dbx *apiImpl) ListSharedLinks(arg *ListSharedLinksArg) (res *ListSharedLinksResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "list_shared_links"), bytes.NewReader(b))
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
			var apiError ListSharedLinksApiError
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

type ModifySharedLinkSettingsApiError struct {
	dropbox.ApiError
	EndpointError *ModifySharedLinkSettingsError `json:"error"`
}

func (dbx *apiImpl) ModifySharedLinkSettings(arg *ModifySharedLinkSettingsArgs) (res IsSharedLinkMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "modify_shared_link_settings"), bytes.NewReader(b))
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
			var apiError ModifySharedLinkSettingsApiError
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
	var tmp sharedLinkMetadataUnion
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return
	}
	switch tmp.Tag {
	case "file":
		res = tmp.File

	case "folder":
		res = tmp.Folder

	}
	return
}

type MountFolderApiError struct {
	dropbox.ApiError
	EndpointError *MountFolderError `json:"error"`
}

func (dbx *apiImpl) MountFolder(arg *MountFolderArg) (res *SharedFolderMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "mount_folder"), bytes.NewReader(b))
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
			var apiError MountFolderApiError
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

type RelinquishFileMembershipApiError struct {
	dropbox.ApiError
	EndpointError *RelinquishFileMembershipError `json:"error"`
}

func (dbx *apiImpl) RelinquishFileMembership(arg *RelinquishFileMembershipArg) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "relinquish_file_membership"), bytes.NewReader(b))
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
			var apiError RelinquishFileMembershipApiError
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

type RelinquishFolderMembershipApiError struct {
	dropbox.ApiError
	EndpointError *RelinquishFolderMembershipError `json:"error"`
}

func (dbx *apiImpl) RelinquishFolderMembership(arg *RelinquishFolderMembershipArg) (res *async.LaunchEmptyResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "relinquish_folder_membership"), bytes.NewReader(b))
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
			var apiError RelinquishFolderMembershipApiError
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

type RemoveFileMemberApiError struct {
	dropbox.ApiError
	EndpointError *RemoveFileMemberError `json:"error"`
}

func (dbx *apiImpl) RemoveFileMember(arg *RemoveFileMemberArg) (res *FileMemberActionIndividualResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "remove_file_member"), bytes.NewReader(b))
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
			var apiError RemoveFileMemberApiError
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

type RemoveFileMember2ApiError struct {
	dropbox.ApiError
	EndpointError *RemoveFileMemberError `json:"error"`
}

func (dbx *apiImpl) RemoveFileMember2(arg *RemoveFileMemberArg) (res *FileMemberRemoveActionResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "remove_file_member_2"), bytes.NewReader(b))
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
			var apiError RemoveFileMember2ApiError
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

type RemoveFolderMemberApiError struct {
	dropbox.ApiError
	EndpointError *RemoveFolderMemberError `json:"error"`
}

func (dbx *apiImpl) RemoveFolderMember(arg *RemoveFolderMemberArg) (res *async.LaunchResultBase, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "remove_folder_member"), bytes.NewReader(b))
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
			var apiError RemoveFolderMemberApiError
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

type RevokeSharedLinkApiError struct {
	dropbox.ApiError
	EndpointError *RevokeSharedLinkError `json:"error"`
}

func (dbx *apiImpl) RevokeSharedLink(arg *RevokeSharedLinkArg) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "revoke_shared_link"), bytes.NewReader(b))
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
			var apiError RevokeSharedLinkApiError
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

type ShareFolderApiError struct {
	dropbox.ApiError
	EndpointError *ShareFolderError `json:"error"`
}

func (dbx *apiImpl) ShareFolder(arg *ShareFolderArg) (res *ShareFolderLaunch, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "share_folder"), bytes.NewReader(b))
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
			var apiError ShareFolderApiError
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

type TransferFolderApiError struct {
	dropbox.ApiError
	EndpointError *TransferFolderError `json:"error"`
}

func (dbx *apiImpl) TransferFolder(arg *TransferFolderArg) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "transfer_folder"), bytes.NewReader(b))
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
			var apiError TransferFolderApiError
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

type UnmountFolderApiError struct {
	dropbox.ApiError
	EndpointError *UnmountFolderError `json:"error"`
}

func (dbx *apiImpl) UnmountFolder(arg *UnmountFolderArg) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "unmount_folder"), bytes.NewReader(b))
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
			var apiError UnmountFolderApiError
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

type UnshareFileApiError struct {
	dropbox.ApiError
	EndpointError *UnshareFileError `json:"error"`
}

func (dbx *apiImpl) UnshareFile(arg *UnshareFileArg) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "unshare_file"), bytes.NewReader(b))
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
			var apiError UnshareFileApiError
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

type UnshareFolderApiError struct {
	dropbox.ApiError
	EndpointError *UnshareFolderError `json:"error"`
}

func (dbx *apiImpl) UnshareFolder(arg *UnshareFolderArg) (res *async.LaunchEmptyResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "unshare_folder"), bytes.NewReader(b))
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
			var apiError UnshareFolderApiError
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

type UpdateFolderMemberApiError struct {
	dropbox.ApiError
	EndpointError *UpdateFolderMemberError `json:"error"`
}

func (dbx *apiImpl) UpdateFolderMember(arg *UpdateFolderMemberArg) (res *MemberAccessLevelResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "update_folder_member"), bytes.NewReader(b))
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
			var apiError UpdateFolderMemberApiError
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

type UpdateFolderPolicyApiError struct {
	dropbox.ApiError
	EndpointError *UpdateFolderPolicyError `json:"error"`
}

func (dbx *apiImpl) UpdateFolderPolicy(arg *UpdateFolderPolicyArg) (res *SharedFolderMetadata, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "sharing", "update_folder_policy"), bytes.NewReader(b))
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
			var apiError UpdateFolderPolicyApiError
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
