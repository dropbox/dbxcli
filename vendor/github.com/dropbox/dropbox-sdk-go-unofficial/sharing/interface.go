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
	"encoding/json"
	"io"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/async"
	"github.com/dropbox/dropbox-sdk-go-unofficial/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/team_common"
	"github.com/dropbox/dropbox-sdk-go-unofficial/users"
)

// Defines the access levels for collaborators.
type AccessLevel struct {
	Tag string `json:".tag"`
}

// Policy governing who can change a shared folder's access control list (ACL).
// In other words, who can add, remove, or change the privileges of members.
type AclUpdatePolicy struct {
	Tag string `json:".tag"`
}

// Arguments for `AddFileMember`.
type AddFileMemberArgs struct {
	// File to which to add members.
	File string `json:"file"`
	// Members to add. Note that even an email address is given, this may result in
	// a user being directy added to the membership if that email is the user's
	// main account email.
	Members []*MemberSelector `json:"members"`
	// Message to send to added members in their invitation.
	CustomMessage string `json:"custom_message,omitempty"`
	// Whether added members should be notified via device notifications of their
	// invitation.
	Quiet bool `json:"quiet"`
	// AccessLevel union object, describing what access level we want to give new
	// members.
	AccessLevel *AccessLevel `json:"access_level"`
	// If the custom message should be added as a comment on the file.
	AddMessageAsComment bool `json:"add_message_as_comment"`
}

func NewAddFileMemberArgs(File string, Members []*MemberSelector) *AddFileMemberArgs {
	s := new(AddFileMemberArgs)
	s.File = File
	s.Members = Members
	s.Quiet = false
	s.AccessLevel = &AccessLevel{Tag: "viewer"}
	s.AddMessageAsComment = false
	return s
}

// Errors for `AddFileMember`.
type AddFileMemberError struct {
	Tag         string                  `json:".tag"`
	UserError   *SharingUserError       `json:"user_error,omitempty"`
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
}

func (u *AddFileMemberError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag         string          `json:".tag"`
		UserError   json.RawMessage `json:"user_error"`
		AccessError json.RawMessage `json:"access_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "user_error":
		{
			if len(w.UserError) == 0 {
				break
			}
			if err := json.Unmarshal(w.UserError, &u.UserError); err != nil {
				return err
			}
		}
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	}
	return nil
}

type AddFolderMemberArg struct {
	// The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// The intended list of members to add.  Added members will receive invites to
	// join the shared folder.
	Members []*AddMember `json:"members"`
	// Whether added members should be notified via email and device notifications
	// of their invite.
	Quiet bool `json:"quiet"`
	// Optional message to display to added members in their invitation.
	CustomMessage string `json:"custom_message,omitempty"`
}

func NewAddFolderMemberArg(SharedFolderId string, Members []*AddMember) *AddFolderMemberArg {
	s := new(AddFolderMemberArg)
	s.SharedFolderId = SharedFolderId
	s.Members = Members
	s.Quiet = false
	return s
}

type AddFolderMemberError struct {
	Tag string `json:".tag"`
	// Unable to access shared folder.
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
	// `AddFolderMemberArg.members` contains a bad invitation recipient.
	BadMember *AddMemberSelectorError `json:"bad_member,omitempty"`
	// The value is the member limit that was reached.
	TooManyMembers uint64 `json:"too_many_members,omitempty"`
	// The value is the pending invite limit that was reached.
	TooManyPendingInvites uint64 `json:"too_many_pending_invites,omitempty"`
}

func (u *AddFolderMemberError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// Unable to access shared folder.
		AccessError json.RawMessage `json:"access_error"`
		// `AddFolderMemberArg.members` contains a bad invitation recipient.
		BadMember json.RawMessage `json:"bad_member"`
		// The value is the member limit that was reached.
		TooManyMembers json.RawMessage `json:"too_many_members"`
		// The value is the pending invite limit that was reached.
		TooManyPendingInvites json.RawMessage `json:"too_many_pending_invites"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	case "bad_member":
		{
			if len(w.BadMember) == 0 {
				break
			}
			if err := json.Unmarshal(w.BadMember, &u.BadMember); err != nil {
				return err
			}
		}
	case "too_many_members":
		{
			if len(w.TooManyMembers) == 0 {
				break
			}
			if err := json.Unmarshal(w.TooManyMembers, &u.TooManyMembers); err != nil {
				return err
			}
		}
	case "too_many_pending_invites":
		{
			if len(w.TooManyPendingInvites) == 0 {
				break
			}
			if err := json.Unmarshal(w.TooManyPendingInvites, &u.TooManyPendingInvites); err != nil {
				return err
			}
		}
	}
	return nil
}

// The member and type of access the member should have when added to a shared
// folder.
type AddMember struct {
	// The member to add to the shared folder.
	Member *MemberSelector `json:"member"`
	// The access level to grant `member` to the shared folder.
	// `AccessLevel.owner` is disallowed.
	AccessLevel *AccessLevel `json:"access_level"`
}

func NewAddMember(Member *MemberSelector) *AddMember {
	s := new(AddMember)
	s.Member = Member
	s.AccessLevel = &AccessLevel{Tag: "viewer"}
	return s
}

type AddMemberSelectorError struct {
	Tag string `json:".tag"`
	// The value is the ID that could not be identified.
	InvalidDropboxId string `json:"invalid_dropbox_id,omitempty"`
	// The value is the e-email address that is malformed.
	InvalidEmail string `json:"invalid_email,omitempty"`
	// The value is the ID of the Dropbox user with an unverified e-mail address.
	// Invite unverified users by e-mail address instead of by their Dropbox ID.
	UnverifiedDropboxId string `json:"unverified_dropbox_id,omitempty"`
}

func (u *AddMemberSelectorError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// The value is the ID that could not be identified.
		InvalidDropboxId json.RawMessage `json:"invalid_dropbox_id"`
		// The value is the e-email address that is malformed.
		InvalidEmail json.RawMessage `json:"invalid_email"`
		// The value is the ID of the Dropbox user with an unverified e-mail address.
		// Invite unverified users by e-mail address instead of by their Dropbox ID.
		UnverifiedDropboxId json.RawMessage `json:"unverified_dropbox_id"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "invalid_dropbox_id":
		{
			if len(w.InvalidDropboxId) == 0 {
				break
			}
			if err := json.Unmarshal(w.InvalidDropboxId, &u.InvalidDropboxId); err != nil {
				return err
			}
		}
	case "invalid_email":
		{
			if len(w.InvalidEmail) == 0 {
				break
			}
			if err := json.Unmarshal(w.InvalidEmail, &u.InvalidEmail); err != nil {
				return err
			}
		}
	case "unverified_dropbox_id":
		{
			if len(w.UnverifiedDropboxId) == 0 {
				break
			}
			if err := json.Unmarshal(w.UnverifiedDropboxId, &u.UnverifiedDropboxId); err != nil {
				return err
			}
		}
	}
	return nil
}

// Metadata for a shared link. This can be either a `PathLinkMetadata` or
// `CollectionLinkMetadata`.
type LinkMetadata struct {
	Tag        string                  `json:".tag"`
	Path       *PathLinkMetadata       `json:"path,omitempty"`
	Collection *CollectionLinkMetadata `json:"collection,omitempty"`
}

func (u *LinkMetadata) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag        string          `json:".tag"`
		Path       json.RawMessage `json:"path"`
		Collection json.RawMessage `json:"collection"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "path":
		{
			if err := json.Unmarshal(body, &u.Path); err != nil {
				return err
			}
		}
	case "collection":
		{
			if err := json.Unmarshal(body, &u.Collection); err != nil {
				return err
			}
		}
	}
	return nil
}

// Metadata for a collection-based shared link.
type CollectionLinkMetadata struct {
	// URL of the shared link.
	Url string `json:"url"`
	// Who can access the link.
	Visibility *Visibility `json:"visibility"`
	// Expiration time, if set. By default the link won't expire.
	Expires time.Time `json:"expires,omitempty"`
}

func NewCollectionLinkMetadata(Url string, Visibility *Visibility) *CollectionLinkMetadata {
	s := new(CollectionLinkMetadata)
	s.Url = Url
	s.Visibility = Visibility
	return s
}

type CreateSharedLinkArg struct {
	// The path to share.
	Path string `json:"path"`
	// Whether to return a shortened URL.
	ShortUrl bool `json:"short_url"`
	// If it's okay to share a path that does not yet exist, set this to either
	// `PendingUploadMode.file` or `PendingUploadMode.folder` to indicate whether
	// to assume it's a file or folder.
	PendingUpload *PendingUploadMode `json:"pending_upload,omitempty"`
}

func NewCreateSharedLinkArg(Path string) *CreateSharedLinkArg {
	s := new(CreateSharedLinkArg)
	s.Path = Path
	s.ShortUrl = false
	return s
}

type CreateSharedLinkError struct {
	Tag  string             `json:".tag"`
	Path *files.LookupError `json:"path,omitempty"`
}

func (u *CreateSharedLinkError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag  string          `json:".tag"`
		Path json.RawMessage `json:"path"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "path":
		{
			if len(w.Path) == 0 {
				break
			}
			if err := json.Unmarshal(w.Path, &u.Path); err != nil {
				return err
			}
		}
	}
	return nil
}

type CreateSharedLinkWithSettingsArg struct {
	// The path to be shared by the shared link
	Path string `json:"path"`
	// The requested settings for the newly created shared link
	Settings *SharedLinkSettings `json:"settings,omitempty"`
}

func NewCreateSharedLinkWithSettingsArg(Path string) *CreateSharedLinkWithSettingsArg {
	s := new(CreateSharedLinkWithSettingsArg)
	s.Path = Path
	return s
}

type CreateSharedLinkWithSettingsError struct {
	Tag  string             `json:".tag"`
	Path *files.LookupError `json:"path,omitempty"`
	// There is an error with the given settings
	SettingsError *SharedLinkSettingsError `json:"settings_error,omitempty"`
}

func (u *CreateSharedLinkWithSettingsError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag  string          `json:".tag"`
		Path json.RawMessage `json:"path"`
		// There is an error with the given settings
		SettingsError json.RawMessage `json:"settings_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "path":
		{
			if len(w.Path) == 0 {
				break
			}
			if err := json.Unmarshal(w.Path, &u.Path); err != nil {
				return err
			}
		}
	case "settings_error":
		{
			if len(w.SettingsError) == 0 {
				break
			}
			if err := json.Unmarshal(w.SettingsError, &u.SettingsError); err != nil {
				return err
			}
		}
	}
	return nil
}

// Sharing actions that may be taken on files.
type FileAction struct {
	Tag string `json:".tag"`
}

type FileErrorResult struct {
	Tag string `json:".tag"`
	// File specified by id was not found.
	FileNotFoundError string `json:"file_not_found_error,omitempty"`
	// User does not have permission to take the specified action on the file.
	InvalidFileActionError string `json:"invalid_file_action_error,omitempty"`
	// User does not have permission to access file specified by file.Id.
	PermissionDeniedError string `json:"permission_denied_error,omitempty"`
}

func (u *FileErrorResult) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// File specified by id was not found.
		FileNotFoundError json.RawMessage `json:"file_not_found_error"`
		// User does not have permission to take the specified action on the file.
		InvalidFileActionError json.RawMessage `json:"invalid_file_action_error"`
		// User does not have permission to access file specified by file.Id.
		PermissionDeniedError json.RawMessage `json:"permission_denied_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "file_not_found_error":
		{
			if len(w.FileNotFoundError) == 0 {
				break
			}
			if err := json.Unmarshal(w.FileNotFoundError, &u.FileNotFoundError); err != nil {
				return err
			}
		}
	case "invalid_file_action_error":
		{
			if len(w.InvalidFileActionError) == 0 {
				break
			}
			if err := json.Unmarshal(w.InvalidFileActionError, &u.InvalidFileActionError); err != nil {
				return err
			}
		}
	case "permission_denied_error":
		{
			if len(w.PermissionDeniedError) == 0 {
				break
			}
			if err := json.Unmarshal(w.PermissionDeniedError, &u.PermissionDeniedError); err != nil {
				return err
			}
		}
	}
	return nil
}

// The metadata of a shared link
type SharedLinkMetadata struct {
	Tag    string              `json:".tag"`
	File   *FileLinkMetadata   `json:"file,omitempty"`
	Folder *FolderLinkMetadata `json:"folder,omitempty"`
}

func (u *SharedLinkMetadata) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag    string          `json:".tag"`
		File   json.RawMessage `json:"file"`
		Folder json.RawMessage `json:"folder"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "file":
		{
			if err := json.Unmarshal(body, &u.File); err != nil {
				return err
			}
		}
	case "folder":
		{
			if err := json.Unmarshal(body, &u.Folder); err != nil {
				return err
			}
		}
	}
	return nil
}

// The metadata of a file shared link
type FileLinkMetadata struct {
	// URL of the shared link.
	Url string `json:"url"`
	// The linked file name (including extension). This never contains a slash.
	Name string `json:"name"`
	// The link's access permissions.
	LinkPermissions *LinkPermissions `json:"link_permissions"`
	// The modification time set by the desktop client when the file was added to
	// Dropbox. Since this time is not verified (the Dropbox server stores whatever
	// the desktop client sends up), this should only be used for display purposes
	// (such as sorting) and not, for example, to determine if a file has changed
	// or not.
	ClientModified time.Time `json:"client_modified"`
	// The last time the file was modified on Dropbox.
	ServerModified time.Time `json:"server_modified"`
	// A unique identifier for the current revision of a file. This field is the
	// same rev as elsewhere in the API and can be used to detect changes and avoid
	// conflicts.
	Rev string `json:"rev"`
	// The file size in bytes.
	Size uint64 `json:"size"`
	// A unique identifier for the linked file.
	Id string `json:"id,omitempty"`
	// Expiration time, if set. By default the link won't expire.
	Expires time.Time `json:"expires,omitempty"`
	// The lowercased full path in the user's Dropbox. This always starts with a
	// slash. This field will only be present only if the linked file is in the
	// authenticated user's  dropbox.
	PathLower string `json:"path_lower,omitempty"`
	// The team membership information of the link's owner.  This field will only
	// be present  if the link's owner is a team member.
	TeamMemberInfo *TeamMemberInfo `json:"team_member_info,omitempty"`
	// The team information of the content's owner. This field will only be present
	// if the content's owner is a team member and the content's owner team is
	// different from the link's owner team.
	ContentOwnerTeamInfo *users.Team `json:"content_owner_team_info,omitempty"`
}

func NewFileLinkMetadata(Url string, Name string, LinkPermissions *LinkPermissions, ClientModified time.Time, ServerModified time.Time, Rev string, Size uint64) *FileLinkMetadata {
	s := new(FileLinkMetadata)
	s.Url = Url
	s.Name = Name
	s.LinkPermissions = LinkPermissions
	s.ClientModified = ClientModified
	s.ServerModified = ServerModified
	s.Rev = Rev
	s.Size = Size
	return s
}

type FileMemberActionError struct {
	Tag string `json:".tag"`
}

type FileMemberActionIndividualResult struct {
	Tag string `json:".tag"`
	// Member was successfully removed from this file. If AccessLevel is given, the
	// member still has access via a parent shared folder.
	Success *AccessLevel `json:"success,omitempty"`
	// User was not able to remove this member.
	MemberError *FileMemberActionError `json:"member_error,omitempty"`
}

func (u *FileMemberActionIndividualResult) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// Member was successfully removed from this file. If AccessLevel is given,
		// the member still has access via a parent shared folder.
		Success json.RawMessage `json:"success,omitempty"`
		// User was not able to remove this member.
		MemberError json.RawMessage `json:"member_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "success":
		{
			if len(w.Success) == 0 {
				break
			}
			if err := json.Unmarshal(w.Success, &u.Success); err != nil {
				return err
			}
		}
	case "member_error":
		{
			if len(w.MemberError) == 0 {
				break
			}
			if err := json.Unmarshal(w.MemberError, &u.MemberError); err != nil {
				return err
			}
		}
	}
	return nil
}

// Per-member result for `RemoveFileMember2` or `AddFileMember`.
type FileMemberActionResult struct {
	// One of specified input members.
	Member *MemberSelector `json:"member"`
	// The outcome of the action on this member.
	Result *FileMemberActionIndividualResult `json:"result"`
}

func NewFileMemberActionResult(Member *MemberSelector, Result *FileMemberActionIndividualResult) *FileMemberActionResult {
	s := new(FileMemberActionResult)
	s.Member = Member
	s.Result = Result
	return s
}

type FileMemberRemoveActionResult struct {
	Tag string `json:".tag"`
	// Member was successfully removed from this file.
	Success *MemberAccessLevelResult `json:"success,omitempty"`
	// User was not able to remove this member.
	MemberError *FileMemberActionError `json:"member_error,omitempty"`
}

func (u *FileMemberRemoveActionResult) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// Member was successfully removed from this file.
		Success json.RawMessage `json:"success"`
		// User was not able to remove this member.
		MemberError json.RawMessage `json:"member_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "success":
		{
			if err := json.Unmarshal(body, &u.Success); err != nil {
				return err
			}
		}
	case "member_error":
		{
			if len(w.MemberError) == 0 {
				break
			}
			if err := json.Unmarshal(w.MemberError, &u.MemberError); err != nil {
				return err
			}
		}
	}
	return nil
}

// Whether the user is allowed to take the sharing action on the file.
type FilePermission struct {
	// The action that the user may wish to take on the file.
	Action *FileAction `json:"action"`
	// True if the user is allowed to take the action.
	Allow bool `json:"allow"`
	// The reason why the user is denied the permission. Not present if the action
	// is allowed
	Reason *PermissionDeniedReason `json:"reason,omitempty"`
}

func NewFilePermission(Action *FileAction, Allow bool) *FilePermission {
	s := new(FilePermission)
	s.Action = Action
	s.Allow = Allow
	return s
}

// Actions that may be taken on shared folders.
type FolderAction struct {
	Tag string `json:".tag"`
}

// The metadata of a folder shared link
type FolderLinkMetadata struct {
	// URL of the shared link.
	Url string `json:"url"`
	// The linked file name (including extension). This never contains a slash.
	Name string `json:"name"`
	// The link's access permissions.
	LinkPermissions *LinkPermissions `json:"link_permissions"`
	// A unique identifier for the linked file.
	Id string `json:"id,omitempty"`
	// Expiration time, if set. By default the link won't expire.
	Expires time.Time `json:"expires,omitempty"`
	// The lowercased full path in the user's Dropbox. This always starts with a
	// slash. This field will only be present only if the linked file is in the
	// authenticated user's  dropbox.
	PathLower string `json:"path_lower,omitempty"`
	// The team membership information of the link's owner.  This field will only
	// be present  if the link's owner is a team member.
	TeamMemberInfo *TeamMemberInfo `json:"team_member_info,omitempty"`
	// The team information of the content's owner. This field will only be present
	// if the content's owner is a team member and the content's owner team is
	// different from the link's owner team.
	ContentOwnerTeamInfo *users.Team `json:"content_owner_team_info,omitempty"`
}

func NewFolderLinkMetadata(Url string, Name string, LinkPermissions *LinkPermissions) *FolderLinkMetadata {
	s := new(FolderLinkMetadata)
	s.Url = Url
	s.Name = Name
	s.LinkPermissions = LinkPermissions
	return s
}

// Whether the user is allowed to take the action on the shared folder.
type FolderPermission struct {
	// The action that the user may wish to take on the folder.
	Action *FolderAction `json:"action"`
	// True if the user is allowed to take the action.
	Allow bool `json:"allow"`
	// The reason why the user is denied the permission. Not present if the action
	// is allowed, or if no reason is available.
	Reason *PermissionDeniedReason `json:"reason,omitempty"`
}

func NewFolderPermission(Action *FolderAction, Allow bool) *FolderPermission {
	s := new(FolderPermission)
	s.Action = Action
	s.Allow = Allow
	return s
}

// A set of policies governing membership and privileges for a shared folder.
type FolderPolicy struct {
	// Who can add and remove members from this shared folder.
	AclUpdatePolicy *AclUpdatePolicy `json:"acl_update_policy"`
	// Who links can be shared with.
	SharedLinkPolicy *SharedLinkPolicy `json:"shared_link_policy"`
	// Who can be a member of this shared folder, as set on the folder itself. The
	// effective policy may differ from this value if the team-wide policy is more
	// restrictive. Present only if the folder is owned by a team.
	MemberPolicy *MemberPolicy `json:"member_policy,omitempty"`
	// Who can be a member of this shared folder, taking into account both the
	// folder and the team-wide policy. This value may differ from that of
	// member_policy if the team-wide policy is more restrictive than the folder
	// policy. Present only if the folder is owned by a team.
	ResolvedMemberPolicy *MemberPolicy `json:"resolved_member_policy,omitempty"`
}

func NewFolderPolicy(AclUpdatePolicy *AclUpdatePolicy, SharedLinkPolicy *SharedLinkPolicy) *FolderPolicy {
	s := new(FolderPolicy)
	s.AclUpdatePolicy = AclUpdatePolicy
	s.SharedLinkPolicy = SharedLinkPolicy
	return s
}

// Arguments of `GetFileMetadata`
type GetFileMetadataArg struct {
	// The file to query.
	File string `json:"file"`
	// File actions to query.
	Actions []*FileAction `json:"actions,omitempty"`
}

func NewGetFileMetadataArg(File string) *GetFileMetadataArg {
	s := new(GetFileMetadataArg)
	s.File = File
	return s
}

// Arguments of `GetFileMetadataBatch`
type GetFileMetadataBatchArg struct {
	// The files to query.
	Files []string `json:"files"`
	// File actions to query.
	Actions []*FileAction `json:"actions,omitempty"`
}

func NewGetFileMetadataBatchArg(Files []string) *GetFileMetadataBatchArg {
	s := new(GetFileMetadataBatchArg)
	s.Files = Files
	return s
}

// Per file results of `GetFileMetadataBatch`
type GetFileMetadataBatchResult struct {
	// This is the input file identifier corresponding to one of
	// `GetFileMetadataBatchArg.files`.
	File string `json:"file"`
	// The result for this particular file
	Result *GetFileMetadataIndividualResult `json:"result"`
}

func NewGetFileMetadataBatchResult(File string, Result *GetFileMetadataIndividualResult) *GetFileMetadataBatchResult {
	s := new(GetFileMetadataBatchResult)
	s.File = File
	s.Result = Result
	return s
}

// Error result for `GetFileMetadata`.
type GetFileMetadataError struct {
	Tag         string                  `json:".tag"`
	UserError   *SharingUserError       `json:"user_error,omitempty"`
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
}

func (u *GetFileMetadataError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag         string          `json:".tag"`
		UserError   json.RawMessage `json:"user_error"`
		AccessError json.RawMessage `json:"access_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "user_error":
		{
			if len(w.UserError) == 0 {
				break
			}
			if err := json.Unmarshal(w.UserError, &u.UserError); err != nil {
				return err
			}
		}
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	}
	return nil
}

type GetFileMetadataIndividualResult struct {
	Tag string `json:".tag"`
	// The result for this file if it was successful.
	Metadata *SharedFileMetadata `json:"metadata,omitempty"`
	// The result for this file if it was an error.
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
}

func (u *GetFileMetadataIndividualResult) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// The result for this file if it was successful.
		Metadata json.RawMessage `json:"metadata"`
		// The result for this file if it was an error.
		AccessError json.RawMessage `json:"access_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "metadata":
		{
			if err := json.Unmarshal(body, &u.Metadata); err != nil {
				return err
			}
		}
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	}
	return nil
}

type GetMetadataArgs struct {
	// The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// This is a list indicating whether the returned folder data will include a
	// boolean value  `FolderPermission.allow` that describes whether the current
	// user can perform the  FolderAction on the folder.
	Actions []*FolderAction `json:"actions,omitempty"`
}

func NewGetMetadataArgs(SharedFolderId string) *GetMetadataArgs {
	s := new(GetMetadataArgs)
	s.SharedFolderId = SharedFolderId
	return s
}

type SharedLinkError struct {
	Tag string `json:".tag"`
}

type GetSharedLinkFileError struct {
	Tag string `json:".tag"`
}

type GetSharedLinkMetadataArg struct {
	// URL of the shared link.
	Url string `json:"url"`
	// If the shared link is to a folder, this parameter can be used to retrieve
	// the metadata for a specific file or sub-folder in this folder. A relative
	// path should be used.
	Path string `json:"path,omitempty"`
	// If the shared link has a password, this parameter can be used.
	LinkPassword string `json:"link_password,omitempty"`
}

func NewGetSharedLinkMetadataArg(Url string) *GetSharedLinkMetadataArg {
	s := new(GetSharedLinkMetadataArg)
	s.Url = Url
	return s
}

type GetSharedLinksArg struct {
	// See `GetSharedLinks` description.
	Path string `json:"path,omitempty"`
}

func NewGetSharedLinksArg() *GetSharedLinksArg {
	s := new(GetSharedLinksArg)
	return s
}

type GetSharedLinksError struct {
	Tag  string `json:".tag"`
	Path string `json:"path,omitempty"`
}

func (u *GetSharedLinksError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag  string          `json:".tag"`
		Path json.RawMessage `json:"path,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "path":
		{
			if len(w.Path) == 0 {
				break
			}
			if err := json.Unmarshal(w.Path, &u.Path); err != nil {
				return err
			}
		}
	}
	return nil
}

type GetSharedLinksResult struct {
	// Shared links applicable to the path argument.
	Links []*LinkMetadata `json:"links"`
}

func NewGetSharedLinksResult(Links []*LinkMetadata) *GetSharedLinksResult {
	s := new(GetSharedLinksResult)
	s.Links = Links
	return s
}

// The information about a group. Groups is a way to manage a list of users  who
// need same access permission to the shared folder.
type GroupInfo struct {
	GroupName string `json:"group_name"`
	GroupId   string `json:"group_id"`
	// The type of group.
	GroupType *team_common.GroupType `json:"group_type"`
	// If the current user is an owner of the group.
	IsOwner bool `json:"is_owner"`
	// If the group is owned by the current user's team.
	SameTeam bool `json:"same_team"`
	// External ID of group. This is an arbitrary ID that an admin can attach to a
	// group.
	GroupExternalId string `json:"group_external_id,omitempty"`
	// The number of members in the group.
	MemberCount uint32 `json:"member_count,omitempty"`
}

func NewGroupInfo(GroupName string, GroupId string, GroupType *team_common.GroupType, IsOwner bool, SameTeam bool) *GroupInfo {
	s := new(GroupInfo)
	s.GroupName = GroupName
	s.GroupId = GroupId
	s.GroupType = GroupType
	s.IsOwner = IsOwner
	s.SameTeam = SameTeam
	return s
}

// The information about a member of the shared content.
type MembershipInfo struct {
	// The access type for this member.
	AccessType *AccessLevel `json:"access_type"`
	// The permissions that requesting user has on this member. The set of
	// permissions corresponds to the MemberActions in the request.
	Permissions []*MemberPermission `json:"permissions,omitempty"`
	// Suggested name initials for a member.
	Initials string `json:"initials,omitempty"`
	// True if the member has access from a parent folder.
	IsInherited bool `json:"is_inherited"`
}

func NewMembershipInfo(AccessType *AccessLevel) *MembershipInfo {
	s := new(MembershipInfo)
	s.AccessType = AccessType
	s.IsInherited = false
	return s
}

// The information about a group member of the shared content.
type GroupMembershipInfo struct {
	// The access type for this member.
	AccessType *AccessLevel `json:"access_type"`
	// The information about the membership group.
	Group *GroupInfo `json:"group"`
	// The permissions that requesting user has on this member. The set of
	// permissions corresponds to the MemberActions in the request.
	Permissions []*MemberPermission `json:"permissions,omitempty"`
	// Suggested name initials for a member.
	Initials string `json:"initials,omitempty"`
	// True if the member has access from a parent folder.
	IsInherited bool `json:"is_inherited"`
}

func NewGroupMembershipInfo(AccessType *AccessLevel, Group *GroupInfo) *GroupMembershipInfo {
	s := new(GroupMembershipInfo)
	s.AccessType = AccessType
	s.Group = Group
	s.IsInherited = false
	return s
}

// Information about the recipient of a shared content invitation.
type InviteeInfo struct {
	Tag string `json:".tag"`
	// E-mail address of invited user.
	Email string `json:"email,omitempty"`
}

func (u *InviteeInfo) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// E-mail address of invited user.
		Email json.RawMessage `json:"email"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "email":
		{
			if len(w.Email) == 0 {
				break
			}
			if err := json.Unmarshal(w.Email, &u.Email); err != nil {
				return err
			}
		}
	}
	return nil
}

// Information about an invited member of a shared content.
type InviteeMembershipInfo struct {
	// The access type for this member.
	AccessType *AccessLevel `json:"access_type"`
	// Recipient of the invitation.
	Invitee *InviteeInfo `json:"invitee"`
	// The permissions that requesting user has on this member. The set of
	// permissions corresponds to the MemberActions in the request.
	Permissions []*MemberPermission `json:"permissions,omitempty"`
	// Suggested name initials for a member.
	Initials string `json:"initials,omitempty"`
	// True if the member has access from a parent folder.
	IsInherited bool `json:"is_inherited"`
	// The user this invitation is tied to, if available.
	User *UserInfo `json:"user,omitempty"`
}

func NewInviteeMembershipInfo(AccessType *AccessLevel, Invitee *InviteeInfo) *InviteeMembershipInfo {
	s := new(InviteeMembershipInfo)
	s.AccessType = AccessType
	s.Invitee = Invitee
	s.IsInherited = false
	return s
}

// Error occurred while performing an asynchronous job from `UnshareFolder` or
// `RemoveFolderMember`.
type JobError struct {
	Tag string `json:".tag"`
	// Error occurred while performing `UnshareFolder` action.
	UnshareFolderError *UnshareFolderError `json:"unshare_folder_error,omitempty"`
	// Error occurred while performing `RemoveFolderMember` action.
	RemoveFolderMemberError *RemoveFolderMemberError `json:"remove_folder_member_error,omitempty"`
	// Error occurred while performing `RelinquishFolderMembership` action.
	RelinquishFolderMembershipError *RelinquishFolderMembershipError `json:"relinquish_folder_membership_error,omitempty"`
}

func (u *JobError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// Error occurred while performing `UnshareFolder` action.
		UnshareFolderError json.RawMessage `json:"unshare_folder_error"`
		// Error occurred while performing `RemoveFolderMember` action.
		RemoveFolderMemberError json.RawMessage `json:"remove_folder_member_error"`
		// Error occurred while performing `RelinquishFolderMembership` action.
		RelinquishFolderMembershipError json.RawMessage `json:"relinquish_folder_membership_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "unshare_folder_error":
		{
			if len(w.UnshareFolderError) == 0 {
				break
			}
			if err := json.Unmarshal(w.UnshareFolderError, &u.UnshareFolderError); err != nil {
				return err
			}
		}
	case "remove_folder_member_error":
		{
			if len(w.RemoveFolderMemberError) == 0 {
				break
			}
			if err := json.Unmarshal(w.RemoveFolderMemberError, &u.RemoveFolderMemberError); err != nil {
				return err
			}
		}
	case "relinquish_folder_membership_error":
		{
			if len(w.RelinquishFolderMembershipError) == 0 {
				break
			}
			if err := json.Unmarshal(w.RelinquishFolderMembershipError, &u.RelinquishFolderMembershipError); err != nil {
				return err
			}
		}
	}
	return nil
}

type JobStatus struct {
	Tag string `json:".tag"`
	// The asynchronous job returned an error.
	Failed *JobError `json:"failed,omitempty"`
}

func (u *JobStatus) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// The asynchronous job returned an error.
		Failed json.RawMessage `json:"failed"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "failed":
		{
			if len(w.Failed) == 0 {
				break
			}
			if err := json.Unmarshal(w.Failed, &u.Failed); err != nil {
				return err
			}
		}
	}
	return nil
}

type LinkPermissions struct {
	// Whether the caller can revoke the shared link
	CanRevoke bool `json:"can_revoke"`
	// The current visibility of the link after considering the shared links
	// policies of the the team (in case the link's owner is part of a team) and
	// the shared folder (in case the linked file is part of a shared folder). This
	// field is shown only if the caller has access to this info (the link's owner
	// always has access to this data).
	ResolvedVisibility *ResolvedVisibility `json:"resolved_visibility,omitempty"`
	// The shared link's requested visibility. This can be overridden by the team
	// and shared folder policies. The final visibility, after considering these
	// policies, can be found in `resolved_visibility`. This is shown only if the
	// caller is the link's owner.
	RequestedVisibility *RequestedVisibility `json:"requested_visibility,omitempty"`
	// The failure reason for revoking the link. This field will only be present if
	// the `can_revoke` is `False`.
	RevokeFailureReason *SharedLinkAccessFailureReason `json:"revoke_failure_reason,omitempty"`
}

func NewLinkPermissions(CanRevoke bool) *LinkPermissions {
	s := new(LinkPermissions)
	s.CanRevoke = CanRevoke
	return s
}

// Arguments for `ListFileMembers`.
type ListFileMembersArg struct {
	// The file for which you want to see members.
	File string `json:"file"`
	// The actions for which to return permissions on a member
	Actions []*MemberAction `json:"actions,omitempty"`
	// Whether to include members who only have access from a parent shared folder.
	IncludeInherited bool `json:"include_inherited"`
	// Number of members to return max per query. Defaults to 100 if no limit is
	// specified.
	Limit uint32 `json:"limit"`
}

func NewListFileMembersArg(File string) *ListFileMembersArg {
	s := new(ListFileMembersArg)
	s.File = File
	s.IncludeInherited = true
	s.Limit = 100
	return s
}

// Arguments for `ListFileMembersBatch`.
type ListFileMembersBatchArg struct {
	// Files for which to return members.
	Files []string `json:"files"`
	// Number of members to return max per query. Defaults to 10 if no limit is
	// specified.
	Limit uint32 `json:"limit"`
}

func NewListFileMembersBatchArg(Files []string) *ListFileMembersBatchArg {
	s := new(ListFileMembersBatchArg)
	s.Files = Files
	s.Limit = 10
	return s
}

// Per-file result for `ListFileMembersBatch`.
type ListFileMembersBatchResult struct {
	// This is the input file identifier, whether an ID or a path.
	File string `json:"file"`
	// The result for this particular file
	Result *ListFileMembersIndividualResult `json:"result"`
}

func NewListFileMembersBatchResult(File string, Result *ListFileMembersIndividualResult) *ListFileMembersBatchResult {
	s := new(ListFileMembersBatchResult)
	s.File = File
	s.Result = Result
	return s
}

// Arguments for `ListFileMembersContinue`.
type ListFileMembersContinueArg struct {
	// The cursor returned by your last call to `ListFileMembers`,
	// `ListFileMembersContinue`, or `ListFileMembersBatch`.
	Cursor string `json:"cursor"`
}

func NewListFileMembersContinueArg(Cursor string) *ListFileMembersContinueArg {
	s := new(ListFileMembersContinueArg)
	s.Cursor = Cursor
	return s
}

// Error for `ListFileMembersContinue`.
type ListFileMembersContinueError struct {
	Tag         string                  `json:".tag"`
	UserError   *SharingUserError       `json:"user_error,omitempty"`
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
}

func (u *ListFileMembersContinueError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag         string          `json:".tag"`
		UserError   json.RawMessage `json:"user_error"`
		AccessError json.RawMessage `json:"access_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "user_error":
		{
			if len(w.UserError) == 0 {
				break
			}
			if err := json.Unmarshal(w.UserError, &u.UserError); err != nil {
				return err
			}
		}
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	}
	return nil
}

type ListFileMembersCountResult struct {
	// A list of members on this file.
	Members *SharedFileMembers `json:"members"`
	// The number of members on this file. This does not include inherited members
	MemberCount uint32 `json:"member_count"`
}

func NewListFileMembersCountResult(Members *SharedFileMembers, MemberCount uint32) *ListFileMembersCountResult {
	s := new(ListFileMembersCountResult)
	s.Members = Members
	s.MemberCount = MemberCount
	return s
}

// Error for `ListFileMembers`.
type ListFileMembersError struct {
	Tag         string                  `json:".tag"`
	UserError   *SharingUserError       `json:"user_error,omitempty"`
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
}

func (u *ListFileMembersError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag         string          `json:".tag"`
		UserError   json.RawMessage `json:"user_error"`
		AccessError json.RawMessage `json:"access_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "user_error":
		{
			if len(w.UserError) == 0 {
				break
			}
			if err := json.Unmarshal(w.UserError, &u.UserError); err != nil {
				return err
			}
		}
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	}
	return nil
}

type ListFileMembersIndividualResult struct {
	Tag string `json:".tag"`
	// The results of the query for this file if it was successful
	Result *ListFileMembersCountResult `json:"result,omitempty"`
	// The result of the query for this file if it was an error.
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
}

func (u *ListFileMembersIndividualResult) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// The results of the query for this file if it was successful
		Result json.RawMessage `json:"result"`
		// The result of the query for this file if it was an error.
		AccessError json.RawMessage `json:"access_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "result":
		{
			if err := json.Unmarshal(body, &u.Result); err != nil {
				return err
			}
		}
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	}
	return nil
}

// Arguments for `ListReceivedFiles`.
type ListFilesArg struct {
	// Number of files to return max per query. Defaults to 100 if no limit is
	// specified.
	Limit uint32 `json:"limit"`
	// File actions to query.
	Actions []*FileAction `json:"actions,omitempty"`
}

func NewListFilesArg() *ListFilesArg {
	s := new(ListFilesArg)
	s.Limit = 100
	return s
}

// Arguments for `ListReceivedFilesContinue`.
type ListFilesContinueArg struct {
	// Cursor in `ListFilesResult.cursor`
	Cursor string `json:"cursor"`
}

func NewListFilesContinueArg(Cursor string) *ListFilesContinueArg {
	s := new(ListFilesContinueArg)
	s.Cursor = Cursor
	return s
}

// Error results for `ListReceivedFilesContinue`.
type ListFilesContinueError struct {
	Tag string `json:".tag"`
	// User account had a problem.
	UserError *SharingUserError `json:"user_error,omitempty"`
}

func (u *ListFilesContinueError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// User account had a problem.
		UserError json.RawMessage `json:"user_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "user_error":
		{
			if len(w.UserError) == 0 {
				break
			}
			if err := json.Unmarshal(w.UserError, &u.UserError); err != nil {
				return err
			}
		}
	}
	return nil
}

// Success results for `ListReceivedFiles`.
type ListFilesResult struct {
	// Information about the files shared with current user.
	Entries []*SharedFileMetadata `json:"entries"`
	// Cursor used to obtain additional shared files.
	Cursor string `json:"cursor,omitempty"`
}

func NewListFilesResult(Entries []*SharedFileMetadata) *ListFilesResult {
	s := new(ListFilesResult)
	s.Entries = Entries
	return s
}

type ListFolderMembersCursorArg struct {
	// This is a list indicating whether each returned member will include a
	// boolean value `MemberPermission.allow` that describes whether the current
	// user can perform the MemberAction on the member.
	Actions []*MemberAction `json:"actions,omitempty"`
	// The maximum number of results that include members, groups and invitees to
	// return per request.
	Limit uint32 `json:"limit"`
}

func NewListFolderMembersCursorArg() *ListFolderMembersCursorArg {
	s := new(ListFolderMembersCursorArg)
	s.Limit = 1000
	return s
}

type ListFolderMembersArgs struct {
	// The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// This is a list indicating whether each returned member will include a
	// boolean value `MemberPermission.allow` that describes whether the current
	// user can perform the MemberAction on the member.
	Actions []*MemberAction `json:"actions,omitempty"`
	// The maximum number of results that include members, groups and invitees to
	// return per request.
	Limit uint32 `json:"limit"`
}

func NewListFolderMembersArgs(SharedFolderId string) *ListFolderMembersArgs {
	s := new(ListFolderMembersArgs)
	s.SharedFolderId = SharedFolderId
	s.Limit = 1000
	return s
}

type ListFolderMembersContinueArg struct {
	// The cursor returned by your last call to `ListFolderMembers` or
	// `ListFolderMembersContinue`.
	Cursor string `json:"cursor"`
}

func NewListFolderMembersContinueArg(Cursor string) *ListFolderMembersContinueArg {
	s := new(ListFolderMembersContinueArg)
	s.Cursor = Cursor
	return s
}

type ListFolderMembersContinueError struct {
	Tag         string                   `json:".tag"`
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
}

func (u *ListFolderMembersContinueError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag         string          `json:".tag"`
		AccessError json.RawMessage `json:"access_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	}
	return nil
}

type ListFoldersArgs struct {
	// The maximum number of results to return per request.
	Limit uint32 `json:"limit"`
	// This is a list indicating whether each returned folder data entry will
	// include a boolean field `FolderPermission.allow` that describes whether the
	// current user can perform the `FolderAction` on the folder.
	Actions []*FolderAction `json:"actions,omitempty"`
}

func NewListFoldersArgs() *ListFoldersArgs {
	s := new(ListFoldersArgs)
	s.Limit = 1000
	return s
}

type ListFoldersContinueArg struct {
	// The cursor returned by the previous API call specified in the endpoint
	// description.
	Cursor string `json:"cursor"`
}

func NewListFoldersContinueArg(Cursor string) *ListFoldersContinueArg {
	s := new(ListFoldersContinueArg)
	s.Cursor = Cursor
	return s
}

type ListFoldersContinueError struct {
	Tag string `json:".tag"`
}

// Result for `ListFolders` or `ListMountableFolders`, depending on which
// endpoint was requested. Unmounted shared folders can be identified by the
// absence of `SharedFolderMetadata.path_lower`.
type ListFoldersResult struct {
	// List of all shared folders the authenticated user has access to.
	Entries []*SharedFolderMetadata `json:"entries"`
	// Present if there are additional shared folders that have not been returned
	// yet. Pass the cursor into the corresponding continue endpoint (either
	// `ListFoldersContinue` or `ListMountableFoldersContinue`) to list additional
	// folders.
	Cursor string `json:"cursor,omitempty"`
}

func NewListFoldersResult(Entries []*SharedFolderMetadata) *ListFoldersResult {
	s := new(ListFoldersResult)
	s.Entries = Entries
	return s
}

type ListSharedLinksArg struct {
	// See `ListSharedLinks` description.
	Path string `json:"path,omitempty"`
	// The cursor returned by your last call to `ListSharedLinks`.
	Cursor string `json:"cursor,omitempty"`
	// See `ListSharedLinks` description.
	DirectOnly bool `json:"direct_only,omitempty"`
}

func NewListSharedLinksArg() *ListSharedLinksArg {
	s := new(ListSharedLinksArg)
	return s
}

type ListSharedLinksError struct {
	Tag  string             `json:".tag"`
	Path *files.LookupError `json:"path,omitempty"`
}

func (u *ListSharedLinksError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag  string          `json:".tag"`
		Path json.RawMessage `json:"path"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "path":
		{
			if len(w.Path) == 0 {
				break
			}
			if err := json.Unmarshal(w.Path, &u.Path); err != nil {
				return err
			}
		}
	}
	return nil
}

type ListSharedLinksResult struct {
	// Shared links applicable to the path argument.
	Links []*SharedLinkMetadata `json:"links"`
	// Is true if there are additional shared links that have not been returned
	// yet. Pass the cursor into `ListSharedLinks` to retrieve them.
	HasMore bool `json:"has_more"`
	// Pass the cursor into `ListSharedLinks` to obtain the additional links.
	// Cursor is returned only if no path is given or the path is empty.
	Cursor string `json:"cursor,omitempty"`
}

func NewListSharedLinksResult(Links []*SharedLinkMetadata, HasMore bool) *ListSharedLinksResult {
	s := new(ListSharedLinksResult)
	s.Links = Links
	s.HasMore = HasMore
	return s
}

// Contains information about a member's access level to content after an
// operation.
type MemberAccessLevelResult struct {
	// The member still has this level of access to the content through a parent
	// folder.
	AccessLevel *AccessLevel `json:"access_level,omitempty"`
	// A localized string with additional information about why the user has this
	// access level to the content.
	Warning string `json:"warning,omitempty"`
}

func NewMemberAccessLevelResult() *MemberAccessLevelResult {
	s := new(MemberAccessLevelResult)
	return s
}

// Actions that may be taken on members of a shared folder.
type MemberAction struct {
	Tag string `json:".tag"`
}

// Whether the user is allowed to take the action on the associated member.
type MemberPermission struct {
	// The action that the user may wish to take on the member.
	Action *MemberAction `json:"action"`
	// True if the user is allowed to take the action.
	Allow bool `json:"allow"`
	// The reason why the user is denied the permission. Not present if the action
	// is allowed
	Reason *PermissionDeniedReason `json:"reason,omitempty"`
}

func NewMemberPermission(Action *MemberAction, Allow bool) *MemberPermission {
	s := new(MemberPermission)
	s.Action = Action
	s.Allow = Allow
	return s
}

// Policy governing who can be a member of a shared folder. Only applicable to
// folders owned by a user on a team.
type MemberPolicy struct {
	Tag string `json:".tag"`
}

// Includes different ways to identify a member of a shared folder.
type MemberSelector struct {
	Tag string `json:".tag"`
	// Dropbox account, team member, or group ID of member.
	DropboxId string `json:"dropbox_id,omitempty"`
	// E-mail address of member.
	Email string `json:"email,omitempty"`
}

func (u *MemberSelector) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// Dropbox account, team member, or group ID of member.
		DropboxId json.RawMessage `json:"dropbox_id"`
		// E-mail address of member.
		Email json.RawMessage `json:"email"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "dropbox_id":
		{
			if len(w.DropboxId) == 0 {
				break
			}
			if err := json.Unmarshal(w.DropboxId, &u.DropboxId); err != nil {
				return err
			}
		}
	case "email":
		{
			if len(w.Email) == 0 {
				break
			}
			if err := json.Unmarshal(w.Email, &u.Email); err != nil {
				return err
			}
		}
	}
	return nil
}

type ModifySharedLinkSettingsArgs struct {
	// URL of the shared link to change its settings
	Url string `json:"url"`
	// Set of settings for the shared link.
	Settings *SharedLinkSettings `json:"settings"`
	// If set to true, removes the expiration of the shared link.
	RemoveExpiration bool `json:"remove_expiration"`
}

func NewModifySharedLinkSettingsArgs(Url string, Settings *SharedLinkSettings) *ModifySharedLinkSettingsArgs {
	s := new(ModifySharedLinkSettingsArgs)
	s.Url = Url
	s.Settings = Settings
	s.RemoveExpiration = false
	return s
}

type ModifySharedLinkSettingsError struct {
	Tag string `json:".tag"`
	// There is an error with the given settings
	SettingsError *SharedLinkSettingsError `json:"settings_error,omitempty"`
}

func (u *ModifySharedLinkSettingsError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// There is an error with the given settings
		SettingsError json.RawMessage `json:"settings_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "settings_error":
		{
			if len(w.SettingsError) == 0 {
				break
			}
			if err := json.Unmarshal(w.SettingsError, &u.SettingsError); err != nil {
				return err
			}
		}
	}
	return nil
}

type MountFolderArg struct {
	// The ID of the shared folder to mount.
	SharedFolderId string `json:"shared_folder_id"`
}

func NewMountFolderArg(SharedFolderId string) *MountFolderArg {
	s := new(MountFolderArg)
	s.SharedFolderId = SharedFolderId
	return s
}

type MountFolderError struct {
	Tag         string                   `json:".tag"`
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
}

func (u *MountFolderError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag         string          `json:".tag"`
		AccessError json.RawMessage `json:"access_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	}
	return nil
}

// Metadata for a path-based shared link.
type PathLinkMetadata struct {
	// URL of the shared link.
	Url string `json:"url"`
	// Who can access the link.
	Visibility *Visibility `json:"visibility"`
	// Path in user's Dropbox.
	Path string `json:"path"`
	// Expiration time, if set. By default the link won't expire.
	Expires time.Time `json:"expires,omitempty"`
}

func NewPathLinkMetadata(Url string, Visibility *Visibility, Path string) *PathLinkMetadata {
	s := new(PathLinkMetadata)
	s.Url = Url
	s.Visibility = Visibility
	s.Path = Path
	return s
}

// Flag to indicate pending upload default (for linking to not-yet-existing
// paths).
type PendingUploadMode struct {
	Tag string `json:".tag"`
}

// Possible reasons the user is denied a permission.
type PermissionDeniedReason struct {
	Tag string `json:".tag"`
}

type RelinquishFileMembershipArg struct {
	// The path or id for the file.
	File string `json:"file"`
}

func NewRelinquishFileMembershipArg(File string) *RelinquishFileMembershipArg {
	s := new(RelinquishFileMembershipArg)
	s.File = File
	return s
}

type RelinquishFileMembershipError struct {
	Tag         string                  `json:".tag"`
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
}

func (u *RelinquishFileMembershipError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag         string          `json:".tag"`
		AccessError json.RawMessage `json:"access_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	}
	return nil
}

type RelinquishFolderMembershipArg struct {
	// The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// Keep a copy of the folder's contents upon relinquishing membership.
	LeaveACopy bool `json:"leave_a_copy"`
}

func NewRelinquishFolderMembershipArg(SharedFolderId string) *RelinquishFolderMembershipArg {
	s := new(RelinquishFolderMembershipArg)
	s.SharedFolderId = SharedFolderId
	s.LeaveACopy = false
	return s
}

type RelinquishFolderMembershipError struct {
	Tag         string                   `json:".tag"`
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
}

func (u *RelinquishFolderMembershipError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag         string          `json:".tag"`
		AccessError json.RawMessage `json:"access_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	}
	return nil
}

// Arguments for `RemoveFileMember2`.
type RemoveFileMemberArg struct {
	// File from which to remove members.
	File string `json:"file"`
	// Member to remove from this file. Note that even if an email is specified, it
	// may result in the removal of a user (not an invitee) if the user's main
	// account corresponds to that email address.
	Member *MemberSelector `json:"member"`
}

func NewRemoveFileMemberArg(File string, Member *MemberSelector) *RemoveFileMemberArg {
	s := new(RemoveFileMemberArg)
	s.File = File
	s.Member = Member
	return s
}

// Errors for `RemoveFileMember2`.
type RemoveFileMemberError struct {
	Tag         string                  `json:".tag"`
	UserError   *SharingUserError       `json:"user_error,omitempty"`
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
}

func (u *RemoveFileMemberError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag         string          `json:".tag"`
		UserError   json.RawMessage `json:"user_error"`
		AccessError json.RawMessage `json:"access_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "user_error":
		{
			if len(w.UserError) == 0 {
				break
			}
			if err := json.Unmarshal(w.UserError, &u.UserError); err != nil {
				return err
			}
		}
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	}
	return nil
}

type RemoveFolderMemberArg struct {
	// The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// The member to remove from the folder.
	Member *MemberSelector `json:"member"`
	// If true, the removed user will keep their copy of the folder after it's
	// unshared, assuming it was mounted. Otherwise, it will be removed from their
	// Dropbox. Also, this must be set to false when kicking a group.
	LeaveACopy bool `json:"leave_a_copy"`
}

func NewRemoveFolderMemberArg(SharedFolderId string, Member *MemberSelector, LeaveACopy bool) *RemoveFolderMemberArg {
	s := new(RemoveFolderMemberArg)
	s.SharedFolderId = SharedFolderId
	s.Member = Member
	s.LeaveACopy = LeaveACopy
	return s
}

type RemoveFolderMemberError struct {
	Tag         string                   `json:".tag"`
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
	MemberError *SharedFolderMemberError `json:"member_error,omitempty"`
}

func (u *RemoveFolderMemberError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag         string          `json:".tag"`
		AccessError json.RawMessage `json:"access_error"`
		MemberError json.RawMessage `json:"member_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	case "member_error":
		{
			if len(w.MemberError) == 0 {
				break
			}
			if err := json.Unmarshal(w.MemberError, &u.MemberError); err != nil {
				return err
			}
		}
	}
	return nil
}

type RemoveMemberJobStatus struct {
	Tag string `json:".tag"`
	// Removing the folder member has finished. The value is information about
	// whether the member has another form of access.
	Complete *MemberAccessLevelResult `json:"complete,omitempty"`
	Failed   *RemoveFolderMemberError `json:"failed,omitempty"`
}

func (u *RemoveMemberJobStatus) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// Removing the folder member has finished. The value is information about
		// whether the member has another form of access.
		Complete json.RawMessage `json:"complete"`
		Failed   json.RawMessage `json:"failed"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "complete":
		{
			if err := json.Unmarshal(body, &u.Complete); err != nil {
				return err
			}
		}
	case "failed":
		{
			if len(w.Failed) == 0 {
				break
			}
			if err := json.Unmarshal(w.Failed, &u.Failed); err != nil {
				return err
			}
		}
	}
	return nil
}

// The access permission that can be requested by the caller for the shared
// link. Note that the final resolved visibility of the shared link takes into
// account other aspects, such as team and shared folder settings. Check the
// `ResolvedVisibility` for more info on the possible resolved visibility values
// of shared links.
type RequestedVisibility struct {
	Tag string `json:".tag"`
}

// The actual access permissions values of shared links after taking into
// account user preferences and the team and shared folder settings. Check the
// `RequestedVisibility` for more info on the possible visibility values that
// can be set by the shared link's owner.
type ResolvedVisibility struct {
	Tag string `json:".tag"`
}

type RevokeSharedLinkArg struct {
	// URL of the shared link.
	Url string `json:"url"`
}

func NewRevokeSharedLinkArg(Url string) *RevokeSharedLinkArg {
	s := new(RevokeSharedLinkArg)
	s.Url = Url
	return s
}

type RevokeSharedLinkError struct {
	Tag string `json:".tag"`
}

type ShareFolderArg struct {
	// The path to the folder to share. If it does not exist, then a new one is
	// created.
	Path string `json:"path"`
	// Who can be a member of this shared folder. Only applicable if the current
	// user is on a team.
	MemberPolicy *MemberPolicy `json:"member_policy"`
	// Who can add and remove members of this shared folder.
	AclUpdatePolicy *AclUpdatePolicy `json:"acl_update_policy"`
	// The policy to apply to shared links created for content inside this shared
	// folder.  The current user must be on a team to set this policy to
	// `SharedLinkPolicy.members`.
	SharedLinkPolicy *SharedLinkPolicy `json:"shared_link_policy"`
	// Whether to force the share to happen asynchronously.
	ForceAsync bool `json:"force_async"`
}

func NewShareFolderArg(Path string) *ShareFolderArg {
	s := new(ShareFolderArg)
	s.Path = Path
	s.MemberPolicy = &MemberPolicy{Tag: "anyone"}
	s.AclUpdatePolicy = &AclUpdatePolicy{Tag: "owner"}
	s.SharedLinkPolicy = &SharedLinkPolicy{Tag: "anyone"}
	s.ForceAsync = false
	return s
}

type ShareFolderErrorBase struct {
	Tag string `json:".tag"`
	// `ShareFolderArg.path` is invalid.
	BadPath *SharePathError `json:"bad_path,omitempty"`
}

func (u *ShareFolderErrorBase) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// `ShareFolderArg.path` is invalid.
		BadPath json.RawMessage `json:"bad_path"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "bad_path":
		{
			if len(w.BadPath) == 0 {
				break
			}
			if err := json.Unmarshal(w.BadPath, &u.BadPath); err != nil {
				return err
			}
		}
	}
	return nil
}

type ShareFolderError struct {
	Tag string `json:".tag"`
}

type ShareFolderJobStatus struct {
	Tag string `json:".tag"`
	// The share job has finished. The value is the metadata for the folder.
	Complete *SharedFolderMetadata `json:"complete,omitempty"`
	Failed   *ShareFolderError     `json:"failed,omitempty"`
}

func (u *ShareFolderJobStatus) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// The share job has finished. The value is the metadata for the folder.
		Complete json.RawMessage `json:"complete"`
		Failed   json.RawMessage `json:"failed"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "complete":
		{
			if err := json.Unmarshal(body, &u.Complete); err != nil {
				return err
			}
		}
	case "failed":
		{
			if len(w.Failed) == 0 {
				break
			}
			if err := json.Unmarshal(w.Failed, &u.Failed); err != nil {
				return err
			}
		}
	}
	return nil
}

type ShareFolderLaunch struct {
	Tag      string                `json:".tag"`
	Complete *SharedFolderMetadata `json:"complete,omitempty"`
}

func (u *ShareFolderLaunch) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag      string          `json:".tag"`
		Complete json.RawMessage `json:"complete"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "complete":
		{
			if err := json.Unmarshal(body, &u.Complete); err != nil {
				return err
			}
		}
	}
	return nil
}

type SharePathError struct {
	Tag string `json:".tag"`
	// Folder is already shared. Contains metadata about the existing shared
	// folder.
	AlreadyShared *SharedFolderMetadata `json:"already_shared,omitempty"`
}

func (u *SharePathError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// Folder is already shared. Contains metadata about the existing shared
		// folder.
		AlreadyShared json.RawMessage `json:"already_shared"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "already_shared":
		{
			if err := json.Unmarshal(body, &u.AlreadyShared); err != nil {
				return err
			}
		}
	}
	return nil
}

// Shared file user, group, and invitee membership. Used for the results of
// `ListFileMembers` and `ListFileMembersContinue`, and used as part of the
// results for `ListFileMembersBatch`.
type SharedFileMembers struct {
	// The list of user members of the shared file.
	Users []*UserMembershipInfo `json:"users"`
	// The list of group members of the shared file.
	Groups []*GroupMembershipInfo `json:"groups"`
	// The list of invited members of a file, but have not logged in and claimed
	// this.
	Invitees []*InviteeMembershipInfo `json:"invitees"`
	// Present if there are additional shared file members that have not been
	// returned yet. Pass the cursor into `ListFileMembersContinue` to list
	// additional members.
	Cursor string `json:"cursor,omitempty"`
}

func NewSharedFileMembers(Users []*UserMembershipInfo, Groups []*GroupMembershipInfo, Invitees []*InviteeMembershipInfo) *SharedFileMembers {
	s := new(SharedFileMembers)
	s.Users = Users
	s.Groups = Groups
	s.Invitees = Invitees
	return s
}

// Properties of the shared file.
type SharedFileMetadata struct {
	// Policies governing this shared file.
	Policy *FolderPolicy `json:"policy"`
	// URL for displaying a web preview of the shared file.
	PreviewUrl string `json:"preview_url"`
	// The name of this file.
	Name string `json:"name"`
	// The ID of the file.
	Id string `json:"id"`
	// The sharing permissions that requesting user has on this file. This
	// corresponds to the entries given in `GetFileMetadataBatchArg.actions` or
	// `GetFileMetadataArg.actions`.
	Permissions []*FilePermission `json:"permissions,omitempty"`
	// The team that owns the file. This field is not present if the file is not
	// owned by a team.
	OwnerTeam *users.Team `json:"owner_team,omitempty"`
	// The ID of the parent shared folder. This field is present only if the file
	// is contained within a shared folder.
	ParentSharedFolderId string `json:"parent_shared_folder_id,omitempty"`
	// The lower-case full path of this file. Absent for unmounted files.
	PathLower string `json:"path_lower,omitempty"`
	// The cased path to be used for display purposes only. In rare instances the
	// casing will not correctly match the user's filesystem, but this behavior
	// will match the path provided in the Core API v1. Absent for unmounted files.
	PathDisplay string `json:"path_display,omitempty"`
}

func NewSharedFileMetadata(Policy *FolderPolicy, PreviewUrl string, Name string, Id string) *SharedFileMetadata {
	s := new(SharedFileMetadata)
	s.Policy = Policy
	s.PreviewUrl = PreviewUrl
	s.Name = Name
	s.Id = Id
	return s
}

// There is an error accessing the shared folder.
type SharedFolderAccessError struct {
	Tag string `json:".tag"`
}

type SharedFolderMemberError struct {
	Tag string `json:".tag"`
}

// Shared folder user and group membership.
type SharedFolderMembers struct {
	// The list of user members of the shared folder.
	Users []*UserMembershipInfo `json:"users"`
	// The list of group members of the shared folder.
	Groups []*GroupMembershipInfo `json:"groups"`
	// The list of invitees to the shared folder.
	Invitees []*InviteeMembershipInfo `json:"invitees"`
	// Present if there are additional shared folder members that have not been
	// returned yet. Pass the cursor into `ListFolderMembersContinue` to list
	// additional members.
	Cursor string `json:"cursor,omitempty"`
}

func NewSharedFolderMembers(Users []*UserMembershipInfo, Groups []*GroupMembershipInfo, Invitees []*InviteeMembershipInfo) *SharedFolderMembers {
	s := new(SharedFolderMembers)
	s.Users = Users
	s.Groups = Groups
	s.Invitees = Invitees
	return s
}

// Properties of the shared folder.
type SharedFolderMetadataBase struct {
	// The current user's access level for this shared folder.
	AccessType *AccessLevel `json:"access_type"`
	// Whether this folder is a `team folder`
	// <https://www.dropbox.com/en/help/986>.
	IsTeamFolder bool `json:"is_team_folder"`
	// Policies governing this shared folder.
	Policy *FolderPolicy `json:"policy"`
	// The team that owns the folder. This field is not present if the folder is
	// not owned by a team.
	OwnerTeam *users.Team `json:"owner_team,omitempty"`
	// The ID of the parent shared folder. This field is present only if the folder
	// is contained within another shared folder.
	ParentSharedFolderId string `json:"parent_shared_folder_id,omitempty"`
}

func NewSharedFolderMetadataBase(AccessType *AccessLevel, IsTeamFolder bool, Policy *FolderPolicy) *SharedFolderMetadataBase {
	s := new(SharedFolderMetadataBase)
	s.AccessType = AccessType
	s.IsTeamFolder = IsTeamFolder
	s.Policy = Policy
	return s
}

// The metadata which includes basic information about the shared folder.
type SharedFolderMetadata struct {
	// The current user's access level for this shared folder.
	AccessType *AccessLevel `json:"access_type"`
	// Whether this folder is a `team folder`
	// <https://www.dropbox.com/en/help/986>.
	IsTeamFolder bool `json:"is_team_folder"`
	// Policies governing this shared folder.
	Policy *FolderPolicy `json:"policy"`
	// The name of the this shared folder.
	Name string `json:"name"`
	// The ID of the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// Timestamp indicating when the current user was invited to this shared
	// folder.
	TimeInvited time.Time `json:"time_invited"`
	// The team that owns the folder. This field is not present if the folder is
	// not owned by a team.
	OwnerTeam *users.Team `json:"owner_team,omitempty"`
	// The ID of the parent shared folder. This field is present only if the folder
	// is contained within another shared folder.
	ParentSharedFolderId string `json:"parent_shared_folder_id,omitempty"`
	// The lower-cased full path of this shared folder. Absent for unmounted
	// folders.
	PathLower string `json:"path_lower,omitempty"`
	// Actions the current user may perform on the folder and its contents. The set
	// of permissions corresponds to the FolderActions in the request.
	Permissions []*FolderPermission `json:"permissions,omitempty"`
}

func NewSharedFolderMetadata(AccessType *AccessLevel, IsTeamFolder bool, Policy *FolderPolicy, Name string, SharedFolderId string, TimeInvited time.Time) *SharedFolderMetadata {
	s := new(SharedFolderMetadata)
	s.AccessType = AccessType
	s.IsTeamFolder = IsTeamFolder
	s.Policy = Policy
	s.Name = Name
	s.SharedFolderId = SharedFolderId
	s.TimeInvited = TimeInvited
	return s
}

type SharedLinkAccessFailureReason struct {
	Tag string `json:".tag"`
}

// Policy governing who can view shared links.
type SharedLinkPolicy struct {
	Tag string `json:".tag"`
}

type SharedLinkSettings struct {
	// The requested access for this shared link.
	RequestedVisibility *RequestedVisibility `json:"requested_visibility,omitempty"`
	// If `requested_visibility` is `RequestedVisibility.password` this is needed
	// to specify the password to access the link.
	LinkPassword string `json:"link_password,omitempty"`
	// Expiration time of the shared link. By default the link won't expire.
	Expires time.Time `json:"expires,omitempty"`
}

func NewSharedLinkSettings() *SharedLinkSettings {
	s := new(SharedLinkSettings)
	return s
}

type SharedLinkSettingsError struct {
	Tag string `json:".tag"`
}

// User could not access this file.
type SharingFileAccessError struct {
	Tag string `json:".tag"`
}

// User account had a problem preventing this action.
type SharingUserError struct {
	Tag string `json:".tag"`
}

// Information about a team member.
type TeamMemberInfo struct {
	// Information about the member's team
	TeamInfo *users.Team `json:"team_info"`
	// The display name of the user.
	DisplayName string `json:"display_name"`
	// ID of user as a member of a team. This field will only be present if the
	// member is in the same team as current user.
	MemberId string `json:"member_id,omitempty"`
}

func NewTeamMemberInfo(TeamInfo *users.Team, DisplayName string) *TeamMemberInfo {
	s := new(TeamMemberInfo)
	s.TeamInfo = TeamInfo
	s.DisplayName = DisplayName
	return s
}

type TransferFolderArg struct {
	// The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// A account or team member ID to transfer ownership to.
	ToDropboxId string `json:"to_dropbox_id"`
}

func NewTransferFolderArg(SharedFolderId string, ToDropboxId string) *TransferFolderArg {
	s := new(TransferFolderArg)
	s.SharedFolderId = SharedFolderId
	s.ToDropboxId = ToDropboxId
	return s
}

type TransferFolderError struct {
	Tag         string                   `json:".tag"`
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
}

func (u *TransferFolderError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag         string          `json:".tag"`
		AccessError json.RawMessage `json:"access_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	}
	return nil
}

type UnmountFolderArg struct {
	// The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
}

func NewUnmountFolderArg(SharedFolderId string) *UnmountFolderArg {
	s := new(UnmountFolderArg)
	s.SharedFolderId = SharedFolderId
	return s
}

type UnmountFolderError struct {
	Tag         string                   `json:".tag"`
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
}

func (u *UnmountFolderError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag         string          `json:".tag"`
		AccessError json.RawMessage `json:"access_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	}
	return nil
}

// Arguments for `UnshareFile`.
type UnshareFileArg struct {
	// The file to unshare.
	File string `json:"file"`
}

func NewUnshareFileArg(File string) *UnshareFileArg {
	s := new(UnshareFileArg)
	s.File = File
	return s
}

// Error result for `UnshareFile`.
type UnshareFileError struct {
	Tag         string                  `json:".tag"`
	UserError   *SharingUserError       `json:"user_error,omitempty"`
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
}

func (u *UnshareFileError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag         string          `json:".tag"`
		UserError   json.RawMessage `json:"user_error"`
		AccessError json.RawMessage `json:"access_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "user_error":
		{
			if len(w.UserError) == 0 {
				break
			}
			if err := json.Unmarshal(w.UserError, &u.UserError); err != nil {
				return err
			}
		}
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	}
	return nil
}

type UnshareFolderArg struct {
	// The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// If true, members of this shared folder will get a copy of this folder after
	// it's unshared. Otherwise, it will be removed from their Dropbox. The current
	// user, who is an owner, will always retain their copy.
	LeaveACopy bool `json:"leave_a_copy"`
}

func NewUnshareFolderArg(SharedFolderId string) *UnshareFolderArg {
	s := new(UnshareFolderArg)
	s.SharedFolderId = SharedFolderId
	s.LeaveACopy = false
	return s
}

type UnshareFolderError struct {
	Tag         string                   `json:".tag"`
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
}

func (u *UnshareFolderError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag         string          `json:".tag"`
		AccessError json.RawMessage `json:"access_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	}
	return nil
}

type UpdateFolderMemberArg struct {
	// The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// The member of the shared folder to update.  Only the
	// `MemberSelector.dropbox_id` may be set at this time.
	Member *MemberSelector `json:"member"`
	// The new access level for `member`. `AccessLevel.owner` is disallowed.
	AccessLevel *AccessLevel `json:"access_level"`
}

func NewUpdateFolderMemberArg(SharedFolderId string, Member *MemberSelector, AccessLevel *AccessLevel) *UpdateFolderMemberArg {
	s := new(UpdateFolderMemberArg)
	s.SharedFolderId = SharedFolderId
	s.Member = Member
	s.AccessLevel = AccessLevel
	return s
}

type UpdateFolderMemberError struct {
	Tag         string                   `json:".tag"`
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
	MemberError *SharedFolderMemberError `json:"member_error,omitempty"`
	// If updating the access type required the member to be added to the shared
	// folder and there was an error when adding the member.
	NoExplicitAccess *AddFolderMemberError `json:"no_explicit_access,omitempty"`
}

func (u *UpdateFolderMemberError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag         string          `json:".tag"`
		AccessError json.RawMessage `json:"access_error"`
		MemberError json.RawMessage `json:"member_error"`
		// If updating the access type required the member to be added to the shared
		// folder and there was an error when adding the member.
		NoExplicitAccess json.RawMessage `json:"no_explicit_access"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	case "member_error":
		{
			if len(w.MemberError) == 0 {
				break
			}
			if err := json.Unmarshal(w.MemberError, &u.MemberError); err != nil {
				return err
			}
		}
	case "no_explicit_access":
		{
			if len(w.NoExplicitAccess) == 0 {
				break
			}
			if err := json.Unmarshal(w.NoExplicitAccess, &u.NoExplicitAccess); err != nil {
				return err
			}
		}
	}
	return nil
}

// If any of the policy's are unset, then they retain their current setting.
type UpdateFolderPolicyArg struct {
	// The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// Who can be a member of this shared folder. Only applicable if the current
	// user is on a team.
	MemberPolicy *MemberPolicy `json:"member_policy,omitempty"`
	// Who can add and remove members of this shared folder.
	AclUpdatePolicy *AclUpdatePolicy `json:"acl_update_policy,omitempty"`
	// The policy to apply to shared links created for content inside this shared
	// folder. The current user must be on a team to set this policy to
	// `SharedLinkPolicy.members`.
	SharedLinkPolicy *SharedLinkPolicy `json:"shared_link_policy,omitempty"`
}

func NewUpdateFolderPolicyArg(SharedFolderId string) *UpdateFolderPolicyArg {
	s := new(UpdateFolderPolicyArg)
	s.SharedFolderId = SharedFolderId
	return s
}

type UpdateFolderPolicyError struct {
	Tag         string                   `json:".tag"`
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
}

func (u *UpdateFolderPolicyError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag         string          `json:".tag"`
		AccessError json.RawMessage `json:"access_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "access_error":
		{
			if len(w.AccessError) == 0 {
				break
			}
			if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
				return err
			}
		}
	}
	return nil
}

// Basic information about a user. Use `UsersAccount` and `UsersAccountBatch` to
// obtain more detailed information.
type UserInfo struct {
	// The account ID of the user.
	AccountId string `json:"account_id"`
	// If the user is in the same team as current user.
	SameTeam bool `json:"same_team"`
	// The team member ID of the shared folder member. Only present if `same_team`
	// is true.
	TeamMemberId string `json:"team_member_id,omitempty"`
}

func NewUserInfo(AccountId string, SameTeam bool) *UserInfo {
	s := new(UserInfo)
	s.AccountId = AccountId
	s.SameTeam = SameTeam
	return s
}

// The information about a user member of the shared content.
type UserMembershipInfo struct {
	// The access type for this member.
	AccessType *AccessLevel `json:"access_type"`
	// The account information for the membership user.
	User *UserInfo `json:"user"`
	// The permissions that requesting user has on this member. The set of
	// permissions corresponds to the MemberActions in the request.
	Permissions []*MemberPermission `json:"permissions,omitempty"`
	// Suggested name initials for a member.
	Initials string `json:"initials,omitempty"`
	// True if the member has access from a parent folder.
	IsInherited bool `json:"is_inherited"`
}

func NewUserMembershipInfo(AccessType *AccessLevel, User *UserInfo) *UserMembershipInfo {
	s := new(UserMembershipInfo)
	s.AccessType = AccessType
	s.User = User
	s.IsInherited = false
	return s
}

// Who can access a shared link. The most open visibility is `public`. The
// default depends on many aspects, such as team and user preferences and shared
// folder settings.
type Visibility struct {
	Tag string `json:".tag"`
}

type Sharing interface {
	// Adds specified members to a file.
	AddFileMember(arg *AddFileMemberArgs) (res []*FileMemberActionResult, err error)
	// Allows an owner or editor (if the ACL update policy allows) of a shared
	// folder to add another member. For the new member to get access to all the
	// functionality for this folder, you will need to call `MountFolder` on their
	// behalf. Apps must have full Dropbox access to use this endpoint.
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
	// needs to revoke a shared link, use `RevokeSharedLink`.
	CreateSharedLink(arg *CreateSharedLinkArg) (res *PathLinkMetadata, err error)
	// Create a shared link with custom settings. If no settings are given then the
	// default visibility is `RequestedVisibility.public` (The resolved visibility,
	// though, may depend on other aspects such as team and shared folder
	// settings).
	CreateSharedLinkWithSettings(arg *CreateSharedLinkWithSettingsArg) (res *SharedLinkMetadata, err error)
	// Returns shared file metadata.
	GetFileMetadata(arg *GetFileMetadataArg) (res *SharedFileMetadata, err error)
	// Returns shared file metadata.
	GetFileMetadataBatch(arg *GetFileMetadataBatchArg) (res []*GetFileMetadataBatchResult, err error)
	// Returns shared folder metadata by its folder ID. Apps must have full Dropbox
	// access to use this endpoint.
	GetFolderMetadata(arg *GetMetadataArgs) (res *SharedFolderMetadata, err error)
	// Download the shared link's file from a user's Dropbox.
	GetSharedLinkFile(arg *GetSharedLinkMetadataArg) (res *SharedLinkMetadata, content io.ReadCloser, err error)
	// Get the shared link's metadata.
	GetSharedLinkMetadata(arg *GetSharedLinkMetadataArg) (res *SharedLinkMetadata, err error)
	// Returns a list of `LinkMetadata` objects for this user, including collection
	// links. If no path is given or the path is empty, returns a list of all
	// shared links for the current user, including collection links. If a
	// non-empty path is given, returns a list of all shared links that allow
	// access to the given path.  Collection links are never returned in this case.
	// Note that the url field in the response is never the shortened URL.
	GetSharedLinks(arg *GetSharedLinksArg) (res *GetSharedLinksResult, err error)
	// Use to obtain the members who have been invited to a file, both inherited
	// and uninherited members.
	ListFileMembers(arg *ListFileMembersArg) (res *SharedFileMembers, err error)
	// Get members of multiple files at once. The arguments to this route are more
	// limited, and the limit on query result size per file is more strict. To
	// customize the results more, use the individual file endpoint. Inherited
	// users are not included in the result, and permissions are not returned for
	// this endpoint.
	ListFileMembersBatch(arg *ListFileMembersBatchArg) (res []*ListFileMembersBatchResult, err error)
	// Once a cursor has been retrieved from `ListFileMembers` or
	// `ListFileMembersBatch`, use this to paginate through all shared file
	// members.
	ListFileMembersContinue(arg *ListFileMembersContinueArg) (res *SharedFileMembers, err error)
	// Returns shared folder membership by its folder ID. Apps must have full
	// Dropbox access to use this endpoint.
	ListFolderMembers(arg *ListFolderMembersArgs) (res *SharedFolderMembers, err error)
	// Once a cursor has been retrieved from `ListFolderMembers`, use this to
	// paginate through all shared folder members. Apps must have full Dropbox
	// access to use this endpoint.
	ListFolderMembersContinue(arg *ListFolderMembersContinueArg) (res *SharedFolderMembers, err error)
	// Return the list of all shared folders the current user has access to. Apps
	// must have full Dropbox access to use this endpoint.
	ListFolders(arg *ListFoldersArgs) (res *ListFoldersResult, err error)
	// Once a cursor has been retrieved from `ListFolders`, use this to paginate
	// through all shared folders. The cursor must come from a previous call to
	// `ListFolders` or `ListFoldersContinue`. Apps must have full Dropbox access
	// to use this endpoint.
	ListFoldersContinue(arg *ListFoldersContinueArg) (res *ListFoldersResult, err error)
	// Return the list of all shared folders the current user can mount or unmount.
	// Apps must have full Dropbox access to use this endpoint.
	ListMountableFolders(arg *ListFoldersArgs) (res *ListFoldersResult, err error)
	// Once a cursor has been retrieved from `ListMountableFolders`, use this to
	// paginate through all mountable shared folders. The cursor must come from a
	// previous call to `ListMountableFolders` or `ListMountableFoldersContinue`.
	// Apps must have full Dropbox access to use this endpoint.
	ListMountableFoldersContinue(arg *ListFoldersContinueArg) (res *ListFoldersResult, err error)
	// Returns a list of all files shared with current user.  Does not include
	// files the user has received via shared folders, and does  not include
	// unclaimed invitations.
	ListReceivedFiles(arg *ListFilesArg) (res *ListFilesResult, err error)
	// Get more results with a cursor from `ListReceivedFiles`.
	ListReceivedFilesContinue(arg *ListFilesContinueArg) (res *ListFilesResult, err error)
	// List shared links of this user. If no path is given or the path is empty,
	// returns a list of all shared links for the current user. If a non-empty path
	// is given, returns a list of all shared links that allow access to the given
	// path - direct links to the given path and links to parent folders of the
	// given path. Links to parent folders can be suppressed by setting direct_only
	// to true.
	ListSharedLinks(arg *ListSharedLinksArg) (res *ListSharedLinksResult, err error)
	// Modify the shared link's settings. If the requested visibility conflict with
	// the shared links policy of the team or the shared folder (in case the linked
	// file is part of a shared folder) then the
	// `LinkPermissions.resolved_visibility` of the returned `SharedLinkMetadata`
	// will reflect the actual visibility of the shared link and the
	// `LinkPermissions.requested_visibility` will reflect the requested
	// visibility.
	ModifySharedLinkSettings(arg *ModifySharedLinkSettingsArgs) (res *SharedLinkMetadata, err error)
	// The current user mounts the designated folder. Mount a shared folder for a
	// user after they have been added as a member. Once mounted, the shared folder
	// will appear in their Dropbox. Apps must have full Dropbox access to use this
	// endpoint.
	MountFolder(arg *MountFolderArg) (res *SharedFolderMetadata, err error)
	// The current user relinquishes their membership in the designated file. Note
	// that the current user may still have inherited access to this file through
	// the parent folder. Apps must have full Dropbox access to use this endpoint.
	RelinquishFileMembership(arg *RelinquishFileMembershipArg) (err error)
	// The current user relinquishes their membership in the designated shared
	// folder and will no longer have access to the folder.  A folder owner cannot
	// relinquish membership in their own folder. This will run synchronously if
	// leave_a_copy is false, and asynchronously if leave_a_copy is true. Apps must
	// have full Dropbox access to use this endpoint.
	RelinquishFolderMembership(arg *RelinquishFolderMembershipArg) (res *async.LaunchEmptyResult, err error)
	// Identical to remove_file_member_2 but with less information returned.
	RemoveFileMember(arg *RemoveFileMemberArg) (res *FileMemberActionIndividualResult, err error)
	// Removes a specified member from the file.
	RemoveFileMember2(arg *RemoveFileMemberArg) (res *FileMemberRemoveActionResult, err error)
	// Allows an owner or editor (if the ACL update policy allows) of a shared
	// folder to remove another member. Apps must have full Dropbox access to use
	// this endpoint.
	RemoveFolderMember(arg *RemoveFolderMemberArg) (res *async.LaunchResultBase, err error)
	// Revoke a shared link. Note that even after revoking a shared link to a file,
	// the file may be accessible if there are shared links leading to any of the
	// file parent folders. To list all shared links that enable access to a
	// specific file, you can use the `ListSharedLinks` with the file as the
	// `ListSharedLinksArg.path` argument.
	RevokeSharedLink(arg *RevokeSharedLinkArg) (err error)
	// Share a folder with collaborators. Most sharing will be completed
	// synchronously. Large folders will be completed asynchronously. To make
	// testing the async case repeatable, set `ShareFolderArg.force_async`. If a
	// `ShareFolderLaunch.async_job_id` is returned, you'll need to call
	// `CheckShareJobStatus` until the action completes to get the metadata for the
	// folder. Apps must have full Dropbox access to use this endpoint.
	ShareFolder(arg *ShareFolderArg) (res *ShareFolderLaunch, err error)
	// Transfer ownership of a shared folder to a member of the shared folder. User
	// must have `AccessLevel.owner` access to the shared folder to perform a
	// transfer. Apps must have full Dropbox access to use this endpoint.
	TransferFolder(arg *TransferFolderArg) (err error)
	// The current user unmounts the designated folder. They can re-mount the
	// folder at a later time using `MountFolder`. Apps must have full Dropbox
	// access to use this endpoint.
	UnmountFolder(arg *UnmountFolderArg) (err error)
	// Remove all members from this file. Does not remove inherited members.
	UnshareFile(arg *UnshareFileArg) (err error)
	// Allows a shared folder owner to unshare the folder. You'll need to call
	// `CheckJobStatus` to determine if the action has completed successfully. Apps
	// must have full Dropbox access to use this endpoint.
	UnshareFolder(arg *UnshareFolderArg) (res *async.LaunchEmptyResult, err error)
	// Allows an owner or editor of a shared folder to update another member's
	// permissions. Apps must have full Dropbox access to use this endpoint.
	UpdateFolderMember(arg *UpdateFolderMemberArg) (res *MemberAccessLevelResult, err error)
	// Update the sharing policies for a shared folder. User must have
	// `AccessLevel.owner` access to the shared folder to update its policies. Apps
	// must have full Dropbox access to use this endpoint.
	UpdateFolderPolicy(arg *UpdateFolderPolicyArg) (res *SharedFolderMetadata, err error)
}
