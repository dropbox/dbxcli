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

// Package sharing : This namespace contains endpoints and data types for
// creating and managing shared links and shared folders.
package sharing

import (
	"encoding/json"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/team_common"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/users"
)

// AccessLevel : Defines the access levels for collaborators.
type AccessLevel struct {
	dropbox.Tagged
}

// Valid tag values for AccessLevel
const (
	AccessLevelOwner           = "owner"
	AccessLevelEditor          = "editor"
	AccessLevelViewer          = "viewer"
	AccessLevelViewerNoComment = "viewer_no_comment"
	AccessLevelOther           = "other"
)

// AclUpdatePolicy : Policy governing who can change a shared folder's access
// control list (ACL). In other words, who can add, remove, or change the
// privileges of members.
type AclUpdatePolicy struct {
	dropbox.Tagged
}

// Valid tag values for AclUpdatePolicy
const (
	AclUpdatePolicyOwner   = "owner"
	AclUpdatePolicyEditors = "editors"
	AclUpdatePolicyOther   = "other"
)

// AddFileMemberArgs : Arguments for `addFileMember`.
type AddFileMemberArgs struct {
	// File : File to which to add members.
	File string `json:"file"`
	// Members : Members to add. Note that even an email address is given, this
	// may result in a user being directy added to the membership if that email
	// is the user's main account email.
	Members []*MemberSelector `json:"members"`
	// CustomMessage : Message to send to added members in their invitation.
	CustomMessage string `json:"custom_message,omitempty"`
	// Quiet : Whether added members should be notified via device notifications
	// of their invitation.
	Quiet bool `json:"quiet"`
	// AccessLevel : AccessLevel union object, describing what access level we
	// want to give new members.
	AccessLevel *AccessLevel `json:"access_level"`
	// AddMessageAsComment : If the custom message should be added as a comment
	// on the file.
	AddMessageAsComment bool `json:"add_message_as_comment"`
}

// NewAddFileMemberArgs returns a new AddFileMemberArgs instance
func NewAddFileMemberArgs(File string, Members []*MemberSelector) *AddFileMemberArgs {
	s := new(AddFileMemberArgs)
	s.File = File
	s.Members = Members
	s.Quiet = false
	s.AccessLevel = &AccessLevel{Tagged: dropbox.Tagged{"viewer"}}
	s.AddMessageAsComment = false
	return s
}

// AddFileMemberError : Errors for `addFileMember`.
type AddFileMemberError struct {
	dropbox.Tagged
	// UserError : has no documentation (yet)
	UserError *SharingUserError `json:"user_error,omitempty"`
	// AccessError : has no documentation (yet)
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
}

// Valid tag values for AddFileMemberError
const (
	AddFileMemberErrorUserError      = "user_error"
	AddFileMemberErrorAccessError    = "access_error"
	AddFileMemberErrorRateLimit      = "rate_limit"
	AddFileMemberErrorInvalidComment = "invalid_comment"
	AddFileMemberErrorOther          = "other"
)

// UnmarshalJSON deserializes into a AddFileMemberError instance
func (u *AddFileMemberError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// UserError : has no documentation (yet)
		UserError json.RawMessage `json:"user_error,omitempty"`
		// AccessError : has no documentation (yet)
		AccessError json.RawMessage `json:"access_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "user_error":
		if err := json.Unmarshal(w.UserError, &u.UserError); err != nil {
			return err
		}

	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	}
	return nil
}

// AddFolderMemberArg : has no documentation (yet)
type AddFolderMemberArg struct {
	// SharedFolderId : The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// Members : The intended list of members to add.  Added members will
	// receive invites to join the shared folder.
	Members []*AddMember `json:"members"`
	// Quiet : Whether added members should be notified via email and device
	// notifications of their invite.
	Quiet bool `json:"quiet"`
	// CustomMessage : Optional message to display to added members in their
	// invitation.
	CustomMessage string `json:"custom_message,omitempty"`
}

// NewAddFolderMemberArg returns a new AddFolderMemberArg instance
func NewAddFolderMemberArg(SharedFolderId string, Members []*AddMember) *AddFolderMemberArg {
	s := new(AddFolderMemberArg)
	s.SharedFolderId = SharedFolderId
	s.Members = Members
	s.Quiet = false
	return s
}

// AddFolderMemberError : has no documentation (yet)
type AddFolderMemberError struct {
	dropbox.Tagged
	// AccessError : Unable to access shared folder.
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
	// BadMember : `AddFolderMemberArg.members` contains a bad invitation
	// recipient.
	BadMember *AddMemberSelectorError `json:"bad_member,omitempty"`
	// TooManyMembers : The value is the member limit that was reached.
	TooManyMembers uint64 `json:"too_many_members,omitempty"`
	// TooManyPendingInvites : The value is the pending invite limit that was
	// reached.
	TooManyPendingInvites uint64 `json:"too_many_pending_invites,omitempty"`
}

// Valid tag values for AddFolderMemberError
const (
	AddFolderMemberErrorAccessError           = "access_error"
	AddFolderMemberErrorEmailUnverified       = "email_unverified"
	AddFolderMemberErrorBadMember             = "bad_member"
	AddFolderMemberErrorCantShareOutsideTeam  = "cant_share_outside_team"
	AddFolderMemberErrorTooManyMembers        = "too_many_members"
	AddFolderMemberErrorTooManyPendingInvites = "too_many_pending_invites"
	AddFolderMemberErrorRateLimit             = "rate_limit"
	AddFolderMemberErrorTooManyInvitees       = "too_many_invitees"
	AddFolderMemberErrorInsufficientPlan      = "insufficient_plan"
	AddFolderMemberErrorTeamFolder            = "team_folder"
	AddFolderMemberErrorNoPermission          = "no_permission"
	AddFolderMemberErrorOther                 = "other"
)

// UnmarshalJSON deserializes into a AddFolderMemberError instance
func (u *AddFolderMemberError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// AccessError : Unable to access shared folder.
		AccessError json.RawMessage `json:"access_error,omitempty"`
		// BadMember : `AddFolderMemberArg.members` contains a bad invitation
		// recipient.
		BadMember json.RawMessage `json:"bad_member,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	case "bad_member":
		if err := json.Unmarshal(w.BadMember, &u.BadMember); err != nil {
			return err
		}

	case "too_many_members":
		if err := json.Unmarshal(body, &u.TooManyMembers); err != nil {
			return err
		}

	case "too_many_pending_invites":
		if err := json.Unmarshal(body, &u.TooManyPendingInvites); err != nil {
			return err
		}

	}
	return nil
}

// AddMember : The member and type of access the member should have when added
// to a shared folder.
type AddMember struct {
	// Member : The member to add to the shared folder.
	Member *MemberSelector `json:"member"`
	// AccessLevel : The access level to grant `member` to the shared folder.
	// `AccessLevel.owner` is disallowed.
	AccessLevel *AccessLevel `json:"access_level"`
}

// NewAddMember returns a new AddMember instance
func NewAddMember(Member *MemberSelector) *AddMember {
	s := new(AddMember)
	s.Member = Member
	s.AccessLevel = &AccessLevel{Tagged: dropbox.Tagged{"viewer"}}
	return s
}

// AddMemberSelectorError : has no documentation (yet)
type AddMemberSelectorError struct {
	dropbox.Tagged
	// InvalidDropboxId : The value is the ID that could not be identified.
	InvalidDropboxId string `json:"invalid_dropbox_id,omitempty"`
	// InvalidEmail : The value is the e-email address that is malformed.
	InvalidEmail string `json:"invalid_email,omitempty"`
	// UnverifiedDropboxId : The value is the ID of the Dropbox user with an
	// unverified e-mail address.  Invite unverified users by e-mail address
	// instead of by their Dropbox ID.
	UnverifiedDropboxId string `json:"unverified_dropbox_id,omitempty"`
}

// Valid tag values for AddMemberSelectorError
const (
	AddMemberSelectorErrorAutomaticGroup      = "automatic_group"
	AddMemberSelectorErrorInvalidDropboxId    = "invalid_dropbox_id"
	AddMemberSelectorErrorInvalidEmail        = "invalid_email"
	AddMemberSelectorErrorUnverifiedDropboxId = "unverified_dropbox_id"
	AddMemberSelectorErrorGroupDeleted        = "group_deleted"
	AddMemberSelectorErrorGroupNotOnTeam      = "group_not_on_team"
	AddMemberSelectorErrorOther               = "other"
)

// UnmarshalJSON deserializes into a AddMemberSelectorError instance
func (u *AddMemberSelectorError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "invalid_dropbox_id":
		if err := json.Unmarshal(body, &u.InvalidDropboxId); err != nil {
			return err
		}

	case "invalid_email":
		if err := json.Unmarshal(body, &u.InvalidEmail); err != nil {
			return err
		}

	case "unverified_dropbox_id":
		if err := json.Unmarshal(body, &u.UnverifiedDropboxId); err != nil {
			return err
		}

	}
	return nil
}

// ChangeFileMemberAccessArgs : Arguments for `changeFileMemberAccess`.
type ChangeFileMemberAccessArgs struct {
	// File : File for which we are changing a member's access.
	File string `json:"file"`
	// Member : The member whose access we are changing.
	Member *MemberSelector `json:"member"`
	// AccessLevel : The new access level for the member.
	AccessLevel *AccessLevel `json:"access_level"`
}

// NewChangeFileMemberAccessArgs returns a new ChangeFileMemberAccessArgs instance
func NewChangeFileMemberAccessArgs(File string, Member *MemberSelector, AccessLevel *AccessLevel) *ChangeFileMemberAccessArgs {
	s := new(ChangeFileMemberAccessArgs)
	s.File = File
	s.Member = Member
	s.AccessLevel = AccessLevel
	return s
}

// LinkMetadata : Metadata for a shared link. This can be either a
// `PathLinkMetadata` or `CollectionLinkMetadata`.
type LinkMetadata struct {
	// Url : URL of the shared link.
	Url string `json:"url"`
	// Visibility : Who can access the link.
	Visibility *Visibility `json:"visibility"`
	// Expires : Expiration time, if set. By default the link won't expire.
	Expires time.Time `json:"expires,omitempty"`
}

// NewLinkMetadata returns a new LinkMetadata instance
func NewLinkMetadata(Url string, Visibility *Visibility) *LinkMetadata {
	s := new(LinkMetadata)
	s.Url = Url
	s.Visibility = Visibility
	return s
}

// IsLinkMetadata is the interface type for LinkMetadata and its subtypes
type IsLinkMetadata interface {
	IsLinkMetadata()
}

// IsLinkMetadata implements the IsLinkMetadata interface
func (u *LinkMetadata) IsLinkMetadata() {}

type linkMetadataUnion struct {
	dropbox.Tagged
	// Path : has no documentation (yet)
	Path *PathLinkMetadata `json:"path,omitempty"`
	// Collection : has no documentation (yet)
	Collection *CollectionLinkMetadata `json:"collection,omitempty"`
}

// Valid tag values for LinkMetadata
const (
	LinkMetadataPath       = "path"
	LinkMetadataCollection = "collection"
)

// UnmarshalJSON deserializes into a linkMetadataUnion instance
func (u *linkMetadataUnion) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// Path : has no documentation (yet)
		Path json.RawMessage `json:"path,omitempty"`
		// Collection : has no documentation (yet)
		Collection json.RawMessage `json:"collection,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "path":
		if err := json.Unmarshal(body, &u.Path); err != nil {
			return err
		}

	case "collection":
		if err := json.Unmarshal(body, &u.Collection); err != nil {
			return err
		}

	}
	return nil
}

// CollectionLinkMetadata : Metadata for a collection-based shared link.
type CollectionLinkMetadata struct {
	LinkMetadata
}

// NewCollectionLinkMetadata returns a new CollectionLinkMetadata instance
func NewCollectionLinkMetadata(Url string, Visibility *Visibility) *CollectionLinkMetadata {
	s := new(CollectionLinkMetadata)
	s.Url = Url
	s.Visibility = Visibility
	return s
}

// CreateSharedLinkArg : has no documentation (yet)
type CreateSharedLinkArg struct {
	// Path : The path to share.
	Path string `json:"path"`
	// ShortUrl : Whether to return a shortened URL.
	ShortUrl bool `json:"short_url"`
	// PendingUpload : If it's okay to share a path that does not yet exist, set
	// this to either `PendingUploadMode.file` or `PendingUploadMode.folder` to
	// indicate whether to assume it's a file or folder.
	PendingUpload *PendingUploadMode `json:"pending_upload,omitempty"`
}

// NewCreateSharedLinkArg returns a new CreateSharedLinkArg instance
func NewCreateSharedLinkArg(Path string) *CreateSharedLinkArg {
	s := new(CreateSharedLinkArg)
	s.Path = Path
	s.ShortUrl = false
	return s
}

// CreateSharedLinkError : has no documentation (yet)
type CreateSharedLinkError struct {
	dropbox.Tagged
	// Path : has no documentation (yet)
	Path *files.LookupError `json:"path,omitempty"`
}

// Valid tag values for CreateSharedLinkError
const (
	CreateSharedLinkErrorPath  = "path"
	CreateSharedLinkErrorOther = "other"
)

// UnmarshalJSON deserializes into a CreateSharedLinkError instance
func (u *CreateSharedLinkError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// Path : has no documentation (yet)
		Path json.RawMessage `json:"path,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "path":
		if err := json.Unmarshal(w.Path, &u.Path); err != nil {
			return err
		}

	}
	return nil
}

// CreateSharedLinkWithSettingsArg : has no documentation (yet)
type CreateSharedLinkWithSettingsArg struct {
	// Path : The path to be shared by the shared link
	Path string `json:"path"`
	// Settings : The requested settings for the newly created shared link
	Settings *SharedLinkSettings `json:"settings,omitempty"`
}

// NewCreateSharedLinkWithSettingsArg returns a new CreateSharedLinkWithSettingsArg instance
func NewCreateSharedLinkWithSettingsArg(Path string) *CreateSharedLinkWithSettingsArg {
	s := new(CreateSharedLinkWithSettingsArg)
	s.Path = Path
	return s
}

// CreateSharedLinkWithSettingsError : has no documentation (yet)
type CreateSharedLinkWithSettingsError struct {
	dropbox.Tagged
	// Path : has no documentation (yet)
	Path *files.LookupError `json:"path,omitempty"`
	// SettingsError : There is an error with the given settings
	SettingsError *SharedLinkSettingsError `json:"settings_error,omitempty"`
}

// Valid tag values for CreateSharedLinkWithSettingsError
const (
	CreateSharedLinkWithSettingsErrorPath                    = "path"
	CreateSharedLinkWithSettingsErrorEmailNotVerified        = "email_not_verified"
	CreateSharedLinkWithSettingsErrorSharedLinkAlreadyExists = "shared_link_already_exists"
	CreateSharedLinkWithSettingsErrorSettingsError           = "settings_error"
	CreateSharedLinkWithSettingsErrorAccessDenied            = "access_denied"
)

// UnmarshalJSON deserializes into a CreateSharedLinkWithSettingsError instance
func (u *CreateSharedLinkWithSettingsError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// Path : has no documentation (yet)
		Path json.RawMessage `json:"path,omitempty"`
		// SettingsError : There is an error with the given settings
		SettingsError json.RawMessage `json:"settings_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "path":
		if err := json.Unmarshal(w.Path, &u.Path); err != nil {
			return err
		}

	case "settings_error":
		if err := json.Unmarshal(w.SettingsError, &u.SettingsError); err != nil {
			return err
		}

	}
	return nil
}

// FileAction : Sharing actions that may be taken on files.
type FileAction struct {
	dropbox.Tagged
}

// Valid tag values for FileAction
const (
	FileActionEditContents          = "edit_contents"
	FileActionInviteViewer          = "invite_viewer"
	FileActionInviteViewerNoComment = "invite_viewer_no_comment"
	FileActionUnshare               = "unshare"
	FileActionRelinquishMembership  = "relinquish_membership"
	FileActionShareLink             = "share_link"
	FileActionOther                 = "other"
)

// FileErrorResult : has no documentation (yet)
type FileErrorResult struct {
	dropbox.Tagged
	// FileNotFoundError : File specified by id was not found.
	FileNotFoundError string `json:"file_not_found_error,omitempty"`
	// InvalidFileActionError : User does not have permission to take the
	// specified action on the file.
	InvalidFileActionError string `json:"invalid_file_action_error,omitempty"`
	// PermissionDeniedError : User does not have permission to access file
	// specified by file.Id.
	PermissionDeniedError string `json:"permission_denied_error,omitempty"`
}

// Valid tag values for FileErrorResult
const (
	FileErrorResultFileNotFoundError      = "file_not_found_error"
	FileErrorResultInvalidFileActionError = "invalid_file_action_error"
	FileErrorResultPermissionDeniedError  = "permission_denied_error"
	FileErrorResultOther                  = "other"
)

// UnmarshalJSON deserializes into a FileErrorResult instance
func (u *FileErrorResult) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "file_not_found_error":
		if err := json.Unmarshal(body, &u.FileNotFoundError); err != nil {
			return err
		}

	case "invalid_file_action_error":
		if err := json.Unmarshal(body, &u.InvalidFileActionError); err != nil {
			return err
		}

	case "permission_denied_error":
		if err := json.Unmarshal(body, &u.PermissionDeniedError); err != nil {
			return err
		}

	}
	return nil
}

// SharedLinkMetadata : The metadata of a shared link
type SharedLinkMetadata struct {
	// Url : URL of the shared link.
	Url string `json:"url"`
	// Id : A unique identifier for the linked file.
	Id string `json:"id,omitempty"`
	// Name : The linked file name (including extension). This never contains a
	// slash.
	Name string `json:"name"`
	// Expires : Expiration time, if set. By default the link won't expire.
	Expires time.Time `json:"expires,omitempty"`
	// PathLower : The lowercased full path in the user's Dropbox. This always
	// starts with a slash. This field will only be present only if the linked
	// file is in the authenticated user's  dropbox.
	PathLower string `json:"path_lower,omitempty"`
	// LinkPermissions : The link's access permissions.
	LinkPermissions *LinkPermissions `json:"link_permissions"`
	// TeamMemberInfo : The team membership information of the link's owner.
	// This field will only be present  if the link's owner is a team member.
	TeamMemberInfo *TeamMemberInfo `json:"team_member_info,omitempty"`
	// ContentOwnerTeamInfo : The team information of the content's owner. This
	// field will only be present if the content's owner is a team member and
	// the content's owner team is different from the link's owner team.
	ContentOwnerTeamInfo *users.Team `json:"content_owner_team_info,omitempty"`
}

// NewSharedLinkMetadata returns a new SharedLinkMetadata instance
func NewSharedLinkMetadata(Url string, Name string, LinkPermissions *LinkPermissions) *SharedLinkMetadata {
	s := new(SharedLinkMetadata)
	s.Url = Url
	s.Name = Name
	s.LinkPermissions = LinkPermissions
	return s
}

// IsSharedLinkMetadata is the interface type for SharedLinkMetadata and its subtypes
type IsSharedLinkMetadata interface {
	IsSharedLinkMetadata()
}

// IsSharedLinkMetadata implements the IsSharedLinkMetadata interface
func (u *SharedLinkMetadata) IsSharedLinkMetadata() {}

type sharedLinkMetadataUnion struct {
	dropbox.Tagged
	// File : has no documentation (yet)
	File *FileLinkMetadata `json:"file,omitempty"`
	// Folder : has no documentation (yet)
	Folder *FolderLinkMetadata `json:"folder,omitempty"`
}

// Valid tag values for SharedLinkMetadata
const (
	SharedLinkMetadataFile   = "file"
	SharedLinkMetadataFolder = "folder"
)

// UnmarshalJSON deserializes into a sharedLinkMetadataUnion instance
func (u *sharedLinkMetadataUnion) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// File : has no documentation (yet)
		File json.RawMessage `json:"file,omitempty"`
		// Folder : has no documentation (yet)
		Folder json.RawMessage `json:"folder,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "file":
		if err := json.Unmarshal(body, &u.File); err != nil {
			return err
		}

	case "folder":
		if err := json.Unmarshal(body, &u.Folder); err != nil {
			return err
		}

	}
	return nil
}

// FileLinkMetadata : The metadata of a file shared link
type FileLinkMetadata struct {
	SharedLinkMetadata
	// ClientModified : The modification time set by the desktop client when the
	// file was added to Dropbox. Since this time is not verified (the Dropbox
	// server stores whatever the desktop client sends up), this should only be
	// used for display purposes (such as sorting) and not, for example, to
	// determine if a file has changed or not.
	ClientModified time.Time `json:"client_modified"`
	// ServerModified : The last time the file was modified on Dropbox.
	ServerModified time.Time `json:"server_modified"`
	// Rev : A unique identifier for the current revision of a file. This field
	// is the same rev as elsewhere in the API and can be used to detect changes
	// and avoid conflicts.
	Rev string `json:"rev"`
	// Size : The file size in bytes.
	Size uint64 `json:"size"`
}

// NewFileLinkMetadata returns a new FileLinkMetadata instance
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

// FileMemberActionError : has no documentation (yet)
type FileMemberActionError struct {
	dropbox.Tagged
}

// Valid tag values for FileMemberActionError
const (
	FileMemberActionErrorInvalidMember = "invalid_member"
	FileMemberActionErrorNoPermission  = "no_permission"
	FileMemberActionErrorOther         = "other"
)

// FileMemberActionIndividualResult : has no documentation (yet)
type FileMemberActionIndividualResult struct {
	dropbox.Tagged
	// Success : Member was successfully removed from this file. If AccessLevel
	// is given, the member still has access via a parent shared folder.
	Success *AccessLevel `json:"success,omitempty"`
	// MemberError : User was not able to perform this action.
	MemberError *FileMemberActionError `json:"member_error,omitempty"`
}

// Valid tag values for FileMemberActionIndividualResult
const (
	FileMemberActionIndividualResultSuccess     = "success"
	FileMemberActionIndividualResultMemberError = "member_error"
)

// UnmarshalJSON deserializes into a FileMemberActionIndividualResult instance
func (u *FileMemberActionIndividualResult) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// Success : Member was successfully removed from this file. If
		// AccessLevel is given, the member still has access via a parent shared
		// folder.
		Success json.RawMessage `json:"success,omitempty"`
		// MemberError : User was not able to perform this action.
		MemberError json.RawMessage `json:"member_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "success":
		if err := json.Unmarshal(body, &u.Success); err != nil {
			return err
		}

	case "member_error":
		if err := json.Unmarshal(w.MemberError, &u.MemberError); err != nil {
			return err
		}

	}
	return nil
}

// FileMemberActionResult : Per-member result for `removeFileMember2` or
// `addFileMember` or `changeFileMemberAccess`.
type FileMemberActionResult struct {
	// Member : One of specified input members.
	Member *MemberSelector `json:"member"`
	// Result : The outcome of the action on this member.
	Result *FileMemberActionIndividualResult `json:"result"`
}

// NewFileMemberActionResult returns a new FileMemberActionResult instance
func NewFileMemberActionResult(Member *MemberSelector, Result *FileMemberActionIndividualResult) *FileMemberActionResult {
	s := new(FileMemberActionResult)
	s.Member = Member
	s.Result = Result
	return s
}

// FileMemberRemoveActionResult : has no documentation (yet)
type FileMemberRemoveActionResult struct {
	dropbox.Tagged
	// Success : Member was successfully removed from this file.
	Success *MemberAccessLevelResult `json:"success,omitempty"`
	// MemberError : User was not able to remove this member.
	MemberError *FileMemberActionError `json:"member_error,omitempty"`
}

// Valid tag values for FileMemberRemoveActionResult
const (
	FileMemberRemoveActionResultSuccess     = "success"
	FileMemberRemoveActionResultMemberError = "member_error"
	FileMemberRemoveActionResultOther       = "other"
)

// UnmarshalJSON deserializes into a FileMemberRemoveActionResult instance
func (u *FileMemberRemoveActionResult) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// Success : Member was successfully removed from this file.
		Success json.RawMessage `json:"success,omitempty"`
		// MemberError : User was not able to remove this member.
		MemberError json.RawMessage `json:"member_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "success":
		if err := json.Unmarshal(body, &u.Success); err != nil {
			return err
		}

	case "member_error":
		if err := json.Unmarshal(w.MemberError, &u.MemberError); err != nil {
			return err
		}

	}
	return nil
}

// FilePermission : Whether the user is allowed to take the sharing action on
// the file.
type FilePermission struct {
	// Action : The action that the user may wish to take on the file.
	Action *FileAction `json:"action"`
	// Allow : True if the user is allowed to take the action.
	Allow bool `json:"allow"`
	// Reason : The reason why the user is denied the permission. Not present if
	// the action is allowed
	Reason *PermissionDeniedReason `json:"reason,omitempty"`
}

// NewFilePermission returns a new FilePermission instance
func NewFilePermission(Action *FileAction, Allow bool) *FilePermission {
	s := new(FilePermission)
	s.Action = Action
	s.Allow = Allow
	return s
}

// FolderAction : Actions that may be taken on shared folders.
type FolderAction struct {
	dropbox.Tagged
}

// Valid tag values for FolderAction
const (
	FolderActionChangeOptions         = "change_options"
	FolderActionEditContents          = "edit_contents"
	FolderActionInviteEditor          = "invite_editor"
	FolderActionInviteViewer          = "invite_viewer"
	FolderActionInviteViewerNoComment = "invite_viewer_no_comment"
	FolderActionRelinquishMembership  = "relinquish_membership"
	FolderActionUnmount               = "unmount"
	FolderActionUnshare               = "unshare"
	FolderActionLeaveACopy            = "leave_a_copy"
	FolderActionShareLink             = "share_link"
	FolderActionOther                 = "other"
)

// FolderLinkMetadata : The metadata of a folder shared link
type FolderLinkMetadata struct {
	SharedLinkMetadata
}

// NewFolderLinkMetadata returns a new FolderLinkMetadata instance
func NewFolderLinkMetadata(Url string, Name string, LinkPermissions *LinkPermissions) *FolderLinkMetadata {
	s := new(FolderLinkMetadata)
	s.Url = Url
	s.Name = Name
	s.LinkPermissions = LinkPermissions
	return s
}

// FolderPermission : Whether the user is allowed to take the action on the
// shared folder.
type FolderPermission struct {
	// Action : The action that the user may wish to take on the folder.
	Action *FolderAction `json:"action"`
	// Allow : True if the user is allowed to take the action.
	Allow bool `json:"allow"`
	// Reason : The reason why the user is denied the permission. Not present if
	// the action is allowed, or if no reason is available.
	Reason *PermissionDeniedReason `json:"reason,omitempty"`
}

// NewFolderPermission returns a new FolderPermission instance
func NewFolderPermission(Action *FolderAction, Allow bool) *FolderPermission {
	s := new(FolderPermission)
	s.Action = Action
	s.Allow = Allow
	return s
}

// FolderPolicy : A set of policies governing membership and privileges for a
// shared folder.
type FolderPolicy struct {
	// MemberPolicy : Who can be a member of this shared folder, as set on the
	// folder itself. The effective policy may differ from this value if the
	// team-wide policy is more restrictive. Present only if the folder is owned
	// by a team.
	MemberPolicy *MemberPolicy `json:"member_policy,omitempty"`
	// ResolvedMemberPolicy : Who can be a member of this shared folder, taking
	// into account both the folder and the team-wide policy. This value may
	// differ from that of member_policy if the team-wide policy is more
	// restrictive than the folder policy. Present only if the folder is owned
	// by a team.
	ResolvedMemberPolicy *MemberPolicy `json:"resolved_member_policy,omitempty"`
	// AclUpdatePolicy : Who can add and remove members from this shared folder.
	AclUpdatePolicy *AclUpdatePolicy `json:"acl_update_policy"`
	// SharedLinkPolicy : Who links can be shared with.
	SharedLinkPolicy *SharedLinkPolicy `json:"shared_link_policy"`
}

// NewFolderPolicy returns a new FolderPolicy instance
func NewFolderPolicy(AclUpdatePolicy *AclUpdatePolicy, SharedLinkPolicy *SharedLinkPolicy) *FolderPolicy {
	s := new(FolderPolicy)
	s.AclUpdatePolicy = AclUpdatePolicy
	s.SharedLinkPolicy = SharedLinkPolicy
	return s
}

// GetFileMetadataArg : Arguments of `getFileMetadata`
type GetFileMetadataArg struct {
	// File : The file to query.
	File string `json:"file"`
	// Actions : File actions to query.
	Actions []*FileAction `json:"actions,omitempty"`
}

// NewGetFileMetadataArg returns a new GetFileMetadataArg instance
func NewGetFileMetadataArg(File string) *GetFileMetadataArg {
	s := new(GetFileMetadataArg)
	s.File = File
	return s
}

// GetFileMetadataBatchArg : Arguments of `getFileMetadataBatch`
type GetFileMetadataBatchArg struct {
	// Files : The files to query.
	Files []string `json:"files"`
	// Actions : File actions to query.
	Actions []*FileAction `json:"actions,omitempty"`
}

// NewGetFileMetadataBatchArg returns a new GetFileMetadataBatchArg instance
func NewGetFileMetadataBatchArg(Files []string) *GetFileMetadataBatchArg {
	s := new(GetFileMetadataBatchArg)
	s.Files = Files
	return s
}

// GetFileMetadataBatchResult : Per file results of `getFileMetadataBatch`
type GetFileMetadataBatchResult struct {
	// File : This is the input file identifier corresponding to one of
	// `GetFileMetadataBatchArg.files`.
	File string `json:"file"`
	// Result : The result for this particular file
	Result *GetFileMetadataIndividualResult `json:"result"`
}

// NewGetFileMetadataBatchResult returns a new GetFileMetadataBatchResult instance
func NewGetFileMetadataBatchResult(File string, Result *GetFileMetadataIndividualResult) *GetFileMetadataBatchResult {
	s := new(GetFileMetadataBatchResult)
	s.File = File
	s.Result = Result
	return s
}

// GetFileMetadataError : Error result for `getFileMetadata`.
type GetFileMetadataError struct {
	dropbox.Tagged
	// UserError : has no documentation (yet)
	UserError *SharingUserError `json:"user_error,omitempty"`
	// AccessError : has no documentation (yet)
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
}

// Valid tag values for GetFileMetadataError
const (
	GetFileMetadataErrorUserError   = "user_error"
	GetFileMetadataErrorAccessError = "access_error"
	GetFileMetadataErrorOther       = "other"
)

// UnmarshalJSON deserializes into a GetFileMetadataError instance
func (u *GetFileMetadataError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// UserError : has no documentation (yet)
		UserError json.RawMessage `json:"user_error,omitempty"`
		// AccessError : has no documentation (yet)
		AccessError json.RawMessage `json:"access_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "user_error":
		if err := json.Unmarshal(w.UserError, &u.UserError); err != nil {
			return err
		}

	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	}
	return nil
}

// GetFileMetadataIndividualResult : has no documentation (yet)
type GetFileMetadataIndividualResult struct {
	dropbox.Tagged
	// Metadata : The result for this file if it was successful.
	Metadata *SharedFileMetadata `json:"metadata,omitempty"`
	// AccessError : The result for this file if it was an error.
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
}

// Valid tag values for GetFileMetadataIndividualResult
const (
	GetFileMetadataIndividualResultMetadata    = "metadata"
	GetFileMetadataIndividualResultAccessError = "access_error"
	GetFileMetadataIndividualResultOther       = "other"
)

// UnmarshalJSON deserializes into a GetFileMetadataIndividualResult instance
func (u *GetFileMetadataIndividualResult) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// Metadata : The result for this file if it was successful.
		Metadata json.RawMessage `json:"metadata,omitempty"`
		// AccessError : The result for this file if it was an error.
		AccessError json.RawMessage `json:"access_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "metadata":
		if err := json.Unmarshal(body, &u.Metadata); err != nil {
			return err
		}

	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	}
	return nil
}

// GetMetadataArgs : has no documentation (yet)
type GetMetadataArgs struct {
	// SharedFolderId : The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// Actions : This is a list indicating whether the returned folder data will
	// include a boolean value  `FolderPermission.allow` that describes whether
	// the current user can perform the  FolderAction on the folder.
	Actions []*FolderAction `json:"actions,omitempty"`
}

// NewGetMetadataArgs returns a new GetMetadataArgs instance
func NewGetMetadataArgs(SharedFolderId string) *GetMetadataArgs {
	s := new(GetMetadataArgs)
	s.SharedFolderId = SharedFolderId
	return s
}

// SharedLinkError : has no documentation (yet)
type SharedLinkError struct {
	dropbox.Tagged
}

// Valid tag values for SharedLinkError
const (
	SharedLinkErrorSharedLinkNotFound     = "shared_link_not_found"
	SharedLinkErrorSharedLinkAccessDenied = "shared_link_access_denied"
	SharedLinkErrorOther                  = "other"
)

// GetSharedLinkFileError : has no documentation (yet)
type GetSharedLinkFileError struct {
	dropbox.Tagged
}

// Valid tag values for GetSharedLinkFileError
const (
	GetSharedLinkFileErrorSharedLinkIsDirectory = "shared_link_is_directory"
)

// GetSharedLinkMetadataArg : has no documentation (yet)
type GetSharedLinkMetadataArg struct {
	// Url : URL of the shared link.
	Url string `json:"url"`
	// Path : If the shared link is to a folder, this parameter can be used to
	// retrieve the metadata for a specific file or sub-folder in this folder. A
	// relative path should be used.
	Path string `json:"path,omitempty"`
	// LinkPassword : If the shared link has a password, this parameter can be
	// used.
	LinkPassword string `json:"link_password,omitempty"`
}

// NewGetSharedLinkMetadataArg returns a new GetSharedLinkMetadataArg instance
func NewGetSharedLinkMetadataArg(Url string) *GetSharedLinkMetadataArg {
	s := new(GetSharedLinkMetadataArg)
	s.Url = Url
	return s
}

// GetSharedLinksArg : has no documentation (yet)
type GetSharedLinksArg struct {
	// Path : See `getSharedLinks` description.
	Path string `json:"path,omitempty"`
}

// NewGetSharedLinksArg returns a new GetSharedLinksArg instance
func NewGetSharedLinksArg() *GetSharedLinksArg {
	s := new(GetSharedLinksArg)
	return s
}

// GetSharedLinksError : has no documentation (yet)
type GetSharedLinksError struct {
	dropbox.Tagged
	// Path : has no documentation (yet)
	Path string `json:"path,omitempty"`
}

// Valid tag values for GetSharedLinksError
const (
	GetSharedLinksErrorPath  = "path"
	GetSharedLinksErrorOther = "other"
)

// UnmarshalJSON deserializes into a GetSharedLinksError instance
func (u *GetSharedLinksError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// Path : has no documentation (yet)
		Path json.RawMessage `json:"path,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "path":
		if err := json.Unmarshal(body, &u.Path); err != nil {
			return err
		}

	}
	return nil
}

// GetSharedLinksResult : has no documentation (yet)
type GetSharedLinksResult struct {
	// Links : Shared links applicable to the path argument.
	Links []IsLinkMetadata `json:"links"`
}

// NewGetSharedLinksResult returns a new GetSharedLinksResult instance
func NewGetSharedLinksResult(Links []IsLinkMetadata) *GetSharedLinksResult {
	s := new(GetSharedLinksResult)
	s.Links = Links
	return s
}

// GroupInfo : The information about a group. Groups is a way to manage a list
// of users  who need same access permission to the shared folder.
type GroupInfo struct {
	team_common.GroupSummary
	// GroupType : The type of group.
	GroupType *team_common.GroupType `json:"group_type"`
	// IsOwner : If the current user is an owner of the group.
	IsOwner bool `json:"is_owner"`
	// SameTeam : If the group is owned by the current user's team.
	SameTeam bool `json:"same_team"`
}

// NewGroupInfo returns a new GroupInfo instance
func NewGroupInfo(GroupName string, GroupId string, GroupManagementType *team_common.GroupManagementType, GroupType *team_common.GroupType, IsOwner bool, SameTeam bool) *GroupInfo {
	s := new(GroupInfo)
	s.GroupName = GroupName
	s.GroupId = GroupId
	s.GroupManagementType = GroupManagementType
	s.GroupType = GroupType
	s.IsOwner = IsOwner
	s.SameTeam = SameTeam
	return s
}

// MembershipInfo : The information about a member of the shared content.
type MembershipInfo struct {
	// AccessType : The access type for this member.
	AccessType *AccessLevel `json:"access_type"`
	// Permissions : The permissions that requesting user has on this member.
	// The set of permissions corresponds to the MemberActions in the request.
	Permissions []*MemberPermission `json:"permissions,omitempty"`
	// Initials : Suggested name initials for a member.
	Initials string `json:"initials,omitempty"`
	// IsInherited : True if the member has access from a parent folder.
	IsInherited bool `json:"is_inherited"`
}

// NewMembershipInfo returns a new MembershipInfo instance
func NewMembershipInfo(AccessType *AccessLevel) *MembershipInfo {
	s := new(MembershipInfo)
	s.AccessType = AccessType
	s.IsInherited = false
	return s
}

// GroupMembershipInfo : The information about a group member of the shared
// content.
type GroupMembershipInfo struct {
	MembershipInfo
	// Group : The information about the membership group.
	Group *GroupInfo `json:"group"`
}

// NewGroupMembershipInfo returns a new GroupMembershipInfo instance
func NewGroupMembershipInfo(AccessType *AccessLevel, Group *GroupInfo) *GroupMembershipInfo {
	s := new(GroupMembershipInfo)
	s.AccessType = AccessType
	s.Group = Group
	s.IsInherited = false
	return s
}

// InsufficientQuotaAmounts : has no documentation (yet)
type InsufficientQuotaAmounts struct {
	// SpaceNeeded : The amount of space needed to add the item (the size of the
	// item).
	SpaceNeeded uint64 `json:"space_needed"`
	// SpaceShortage : The amount of extra space needed to add the item.
	SpaceShortage uint64 `json:"space_shortage"`
	// SpaceLeft : The amount of space left in the user's Dropbox, less than
	// space_needed.
	SpaceLeft uint64 `json:"space_left"`
}

// NewInsufficientQuotaAmounts returns a new InsufficientQuotaAmounts instance
func NewInsufficientQuotaAmounts(SpaceNeeded uint64, SpaceShortage uint64, SpaceLeft uint64) *InsufficientQuotaAmounts {
	s := new(InsufficientQuotaAmounts)
	s.SpaceNeeded = SpaceNeeded
	s.SpaceShortage = SpaceShortage
	s.SpaceLeft = SpaceLeft
	return s
}

// InviteeInfo : Information about the recipient of a shared content invitation.
type InviteeInfo struct {
	dropbox.Tagged
	// Email : E-mail address of invited user.
	Email string `json:"email,omitempty"`
}

// Valid tag values for InviteeInfo
const (
	InviteeInfoEmail = "email"
	InviteeInfoOther = "other"
)

// UnmarshalJSON deserializes into a InviteeInfo instance
func (u *InviteeInfo) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "email":
		if err := json.Unmarshal(body, &u.Email); err != nil {
			return err
		}

	}
	return nil
}

// InviteeMembershipInfo : Information about an invited member of a shared
// content.
type InviteeMembershipInfo struct {
	MembershipInfo
	// Invitee : Recipient of the invitation.
	Invitee *InviteeInfo `json:"invitee"`
	// User : The user this invitation is tied to, if available.
	User *UserInfo `json:"user,omitempty"`
}

// NewInviteeMembershipInfo returns a new InviteeMembershipInfo instance
func NewInviteeMembershipInfo(AccessType *AccessLevel, Invitee *InviteeInfo) *InviteeMembershipInfo {
	s := new(InviteeMembershipInfo)
	s.AccessType = AccessType
	s.Invitee = Invitee
	s.IsInherited = false
	return s
}

// JobError : Error occurred while performing an asynchronous job from
// `unshareFolder` or `removeFolderMember`.
type JobError struct {
	dropbox.Tagged
	// UnshareFolderError : Error occurred while performing `unshareFolder`
	// action.
	UnshareFolderError *UnshareFolderError `json:"unshare_folder_error,omitempty"`
	// RemoveFolderMemberError : Error occurred while performing
	// `removeFolderMember` action.
	RemoveFolderMemberError *RemoveFolderMemberError `json:"remove_folder_member_error,omitempty"`
	// RelinquishFolderMembershipError : Error occurred while performing
	// `relinquishFolderMembership` action.
	RelinquishFolderMembershipError *RelinquishFolderMembershipError `json:"relinquish_folder_membership_error,omitempty"`
}

// Valid tag values for JobError
const (
	JobErrorUnshareFolderError              = "unshare_folder_error"
	JobErrorRemoveFolderMemberError         = "remove_folder_member_error"
	JobErrorRelinquishFolderMembershipError = "relinquish_folder_membership_error"
	JobErrorOther                           = "other"
)

// UnmarshalJSON deserializes into a JobError instance
func (u *JobError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// UnshareFolderError : Error occurred while performing `unshareFolder`
		// action.
		UnshareFolderError json.RawMessage `json:"unshare_folder_error,omitempty"`
		// RemoveFolderMemberError : Error occurred while performing
		// `removeFolderMember` action.
		RemoveFolderMemberError json.RawMessage `json:"remove_folder_member_error,omitempty"`
		// RelinquishFolderMembershipError : Error occurred while performing
		// `relinquishFolderMembership` action.
		RelinquishFolderMembershipError json.RawMessage `json:"relinquish_folder_membership_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "unshare_folder_error":
		if err := json.Unmarshal(w.UnshareFolderError, &u.UnshareFolderError); err != nil {
			return err
		}

	case "remove_folder_member_error":
		if err := json.Unmarshal(w.RemoveFolderMemberError, &u.RemoveFolderMemberError); err != nil {
			return err
		}

	case "relinquish_folder_membership_error":
		if err := json.Unmarshal(w.RelinquishFolderMembershipError, &u.RelinquishFolderMembershipError); err != nil {
			return err
		}

	}
	return nil
}

// JobStatus : has no documentation (yet)
type JobStatus struct {
	dropbox.Tagged
	// Failed : The asynchronous job returned an error.
	Failed *JobError `json:"failed,omitempty"`
}

// Valid tag values for JobStatus
const (
	JobStatusComplete = "complete"
	JobStatusFailed   = "failed"
)

// UnmarshalJSON deserializes into a JobStatus instance
func (u *JobStatus) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// Failed : The asynchronous job returned an error.
		Failed json.RawMessage `json:"failed,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "failed":
		if err := json.Unmarshal(w.Failed, &u.Failed); err != nil {
			return err
		}

	}
	return nil
}

// LinkPermissions : has no documentation (yet)
type LinkPermissions struct {
	// ResolvedVisibility : The current visibility of the link after considering
	// the shared links policies of the the team (in case the link's owner is
	// part of a team) and the shared folder (in case the linked file is part of
	// a shared folder). This field is shown only if the caller has access to
	// this info (the link's owner always has access to this data).
	ResolvedVisibility *ResolvedVisibility `json:"resolved_visibility,omitempty"`
	// RequestedVisibility : The shared link's requested visibility. This can be
	// overridden by the team and shared folder policies. The final visibility,
	// after considering these policies, can be found in `resolved_visibility`.
	// This is shown only if the caller is the link's owner.
	RequestedVisibility *RequestedVisibility `json:"requested_visibility,omitempty"`
	// CanRevoke : Whether the caller can revoke the shared link
	CanRevoke bool `json:"can_revoke"`
	// RevokeFailureReason : The failure reason for revoking the link. This
	// field will only be present if the `can_revoke` is false.
	RevokeFailureReason *SharedLinkAccessFailureReason `json:"revoke_failure_reason,omitempty"`
}

// NewLinkPermissions returns a new LinkPermissions instance
func NewLinkPermissions(CanRevoke bool) *LinkPermissions {
	s := new(LinkPermissions)
	s.CanRevoke = CanRevoke
	return s
}

// ListFileMembersArg : Arguments for `listFileMembers`.
type ListFileMembersArg struct {
	// File : The file for which you want to see members.
	File string `json:"file"`
	// Actions : The actions for which to return permissions on a member
	Actions []*MemberAction `json:"actions,omitempty"`
	// IncludeInherited : Whether to include members who only have access from a
	// parent shared folder.
	IncludeInherited bool `json:"include_inherited"`
	// Limit : Number of members to return max per query. Defaults to 100 if no
	// limit is specified.
	Limit uint32 `json:"limit"`
}

// NewListFileMembersArg returns a new ListFileMembersArg instance
func NewListFileMembersArg(File string) *ListFileMembersArg {
	s := new(ListFileMembersArg)
	s.File = File
	s.IncludeInherited = true
	s.Limit = 100
	return s
}

// ListFileMembersBatchArg : Arguments for `listFileMembersBatch`.
type ListFileMembersBatchArg struct {
	// Files : Files for which to return members.
	Files []string `json:"files"`
	// Limit : Number of members to return max per query. Defaults to 10 if no
	// limit is specified.
	Limit uint32 `json:"limit"`
}

// NewListFileMembersBatchArg returns a new ListFileMembersBatchArg instance
func NewListFileMembersBatchArg(Files []string) *ListFileMembersBatchArg {
	s := new(ListFileMembersBatchArg)
	s.Files = Files
	s.Limit = 10
	return s
}

// ListFileMembersBatchResult : Per-file result for `listFileMembersBatch`.
type ListFileMembersBatchResult struct {
	// File : This is the input file identifier, whether an ID or a path.
	File string `json:"file"`
	// Result : The result for this particular file
	Result *ListFileMembersIndividualResult `json:"result"`
}

// NewListFileMembersBatchResult returns a new ListFileMembersBatchResult instance
func NewListFileMembersBatchResult(File string, Result *ListFileMembersIndividualResult) *ListFileMembersBatchResult {
	s := new(ListFileMembersBatchResult)
	s.File = File
	s.Result = Result
	return s
}

// ListFileMembersContinueArg : Arguments for `listFileMembersContinue`.
type ListFileMembersContinueArg struct {
	// Cursor : The cursor returned by your last call to `listFileMembers`,
	// `listFileMembersContinue`, or `listFileMembersBatch`.
	Cursor string `json:"cursor"`
}

// NewListFileMembersContinueArg returns a new ListFileMembersContinueArg instance
func NewListFileMembersContinueArg(Cursor string) *ListFileMembersContinueArg {
	s := new(ListFileMembersContinueArg)
	s.Cursor = Cursor
	return s
}

// ListFileMembersContinueError : Error for `listFileMembersContinue`.
type ListFileMembersContinueError struct {
	dropbox.Tagged
	// UserError : has no documentation (yet)
	UserError *SharingUserError `json:"user_error,omitempty"`
	// AccessError : has no documentation (yet)
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
}

// Valid tag values for ListFileMembersContinueError
const (
	ListFileMembersContinueErrorUserError     = "user_error"
	ListFileMembersContinueErrorAccessError   = "access_error"
	ListFileMembersContinueErrorInvalidCursor = "invalid_cursor"
	ListFileMembersContinueErrorOther         = "other"
)

// UnmarshalJSON deserializes into a ListFileMembersContinueError instance
func (u *ListFileMembersContinueError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// UserError : has no documentation (yet)
		UserError json.RawMessage `json:"user_error,omitempty"`
		// AccessError : has no documentation (yet)
		AccessError json.RawMessage `json:"access_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "user_error":
		if err := json.Unmarshal(w.UserError, &u.UserError); err != nil {
			return err
		}

	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	}
	return nil
}

// ListFileMembersCountResult : has no documentation (yet)
type ListFileMembersCountResult struct {
	// Members : A list of members on this file.
	Members *SharedFileMembers `json:"members"`
	// MemberCount : The number of members on this file. This does not include
	// inherited members
	MemberCount uint32 `json:"member_count"`
}

// NewListFileMembersCountResult returns a new ListFileMembersCountResult instance
func NewListFileMembersCountResult(Members *SharedFileMembers, MemberCount uint32) *ListFileMembersCountResult {
	s := new(ListFileMembersCountResult)
	s.Members = Members
	s.MemberCount = MemberCount
	return s
}

// ListFileMembersError : Error for `listFileMembers`.
type ListFileMembersError struct {
	dropbox.Tagged
	// UserError : has no documentation (yet)
	UserError *SharingUserError `json:"user_error,omitempty"`
	// AccessError : has no documentation (yet)
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
}

// Valid tag values for ListFileMembersError
const (
	ListFileMembersErrorUserError   = "user_error"
	ListFileMembersErrorAccessError = "access_error"
	ListFileMembersErrorOther       = "other"
)

// UnmarshalJSON deserializes into a ListFileMembersError instance
func (u *ListFileMembersError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// UserError : has no documentation (yet)
		UserError json.RawMessage `json:"user_error,omitempty"`
		// AccessError : has no documentation (yet)
		AccessError json.RawMessage `json:"access_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "user_error":
		if err := json.Unmarshal(w.UserError, &u.UserError); err != nil {
			return err
		}

	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	}
	return nil
}

// ListFileMembersIndividualResult : has no documentation (yet)
type ListFileMembersIndividualResult struct {
	dropbox.Tagged
	// Result : The results of the query for this file if it was successful
	Result *ListFileMembersCountResult `json:"result,omitempty"`
	// AccessError : The result of the query for this file if it was an error.
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
}

// Valid tag values for ListFileMembersIndividualResult
const (
	ListFileMembersIndividualResultResult      = "result"
	ListFileMembersIndividualResultAccessError = "access_error"
	ListFileMembersIndividualResultOther       = "other"
)

// UnmarshalJSON deserializes into a ListFileMembersIndividualResult instance
func (u *ListFileMembersIndividualResult) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// Result : The results of the query for this file if it was successful
		Result json.RawMessage `json:"result,omitempty"`
		// AccessError : The result of the query for this file if it was an
		// error.
		AccessError json.RawMessage `json:"access_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "result":
		if err := json.Unmarshal(body, &u.Result); err != nil {
			return err
		}

	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	}
	return nil
}

// ListFilesArg : Arguments for `listReceivedFiles`.
type ListFilesArg struct {
	// Limit : Number of files to return max per query. Defaults to 100 if no
	// limit is specified.
	Limit uint32 `json:"limit"`
	// Actions : File actions to query.
	Actions []*FileAction `json:"actions,omitempty"`
}

// NewListFilesArg returns a new ListFilesArg instance
func NewListFilesArg() *ListFilesArg {
	s := new(ListFilesArg)
	s.Limit = 100
	return s
}

// ListFilesContinueArg : Arguments for `listReceivedFilesContinue`.
type ListFilesContinueArg struct {
	// Cursor : Cursor in `ListFilesResult.cursor`
	Cursor string `json:"cursor"`
}

// NewListFilesContinueArg returns a new ListFilesContinueArg instance
func NewListFilesContinueArg(Cursor string) *ListFilesContinueArg {
	s := new(ListFilesContinueArg)
	s.Cursor = Cursor
	return s
}

// ListFilesContinueError : Error results for `listReceivedFilesContinue`.
type ListFilesContinueError struct {
	dropbox.Tagged
	// UserError : User account had a problem.
	UserError *SharingUserError `json:"user_error,omitempty"`
}

// Valid tag values for ListFilesContinueError
const (
	ListFilesContinueErrorUserError     = "user_error"
	ListFilesContinueErrorInvalidCursor = "invalid_cursor"
	ListFilesContinueErrorOther         = "other"
)

// UnmarshalJSON deserializes into a ListFilesContinueError instance
func (u *ListFilesContinueError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// UserError : User account had a problem.
		UserError json.RawMessage `json:"user_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "user_error":
		if err := json.Unmarshal(w.UserError, &u.UserError); err != nil {
			return err
		}

	}
	return nil
}

// ListFilesResult : Success results for `listReceivedFiles`.
type ListFilesResult struct {
	// Entries : Information about the files shared with current user.
	Entries []*SharedFileMetadata `json:"entries"`
	// Cursor : Cursor used to obtain additional shared files.
	Cursor string `json:"cursor,omitempty"`
}

// NewListFilesResult returns a new ListFilesResult instance
func NewListFilesResult(Entries []*SharedFileMetadata) *ListFilesResult {
	s := new(ListFilesResult)
	s.Entries = Entries
	return s
}

// ListFolderMembersCursorArg : has no documentation (yet)
type ListFolderMembersCursorArg struct {
	// Actions : This is a list indicating whether each returned member will
	// include a boolean value `MemberPermission.allow` that describes whether
	// the current user can perform the MemberAction on the member.
	Actions []*MemberAction `json:"actions,omitempty"`
	// Limit : The maximum number of results that include members, groups and
	// invitees to return per request.
	Limit uint32 `json:"limit"`
}

// NewListFolderMembersCursorArg returns a new ListFolderMembersCursorArg instance
func NewListFolderMembersCursorArg() *ListFolderMembersCursorArg {
	s := new(ListFolderMembersCursorArg)
	s.Limit = 1000
	return s
}

// ListFolderMembersArgs : has no documentation (yet)
type ListFolderMembersArgs struct {
	ListFolderMembersCursorArg
	// SharedFolderId : The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
}

// NewListFolderMembersArgs returns a new ListFolderMembersArgs instance
func NewListFolderMembersArgs(SharedFolderId string) *ListFolderMembersArgs {
	s := new(ListFolderMembersArgs)
	s.SharedFolderId = SharedFolderId
	s.Limit = 1000
	return s
}

// ListFolderMembersContinueArg : has no documentation (yet)
type ListFolderMembersContinueArg struct {
	// Cursor : The cursor returned by your last call to `listFolderMembers` or
	// `listFolderMembersContinue`.
	Cursor string `json:"cursor"`
}

// NewListFolderMembersContinueArg returns a new ListFolderMembersContinueArg instance
func NewListFolderMembersContinueArg(Cursor string) *ListFolderMembersContinueArg {
	s := new(ListFolderMembersContinueArg)
	s.Cursor = Cursor
	return s
}

// ListFolderMembersContinueError : has no documentation (yet)
type ListFolderMembersContinueError struct {
	dropbox.Tagged
	// AccessError : has no documentation (yet)
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
}

// Valid tag values for ListFolderMembersContinueError
const (
	ListFolderMembersContinueErrorAccessError   = "access_error"
	ListFolderMembersContinueErrorInvalidCursor = "invalid_cursor"
	ListFolderMembersContinueErrorOther         = "other"
)

// UnmarshalJSON deserializes into a ListFolderMembersContinueError instance
func (u *ListFolderMembersContinueError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// AccessError : has no documentation (yet)
		AccessError json.RawMessage `json:"access_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	}
	return nil
}

// ListFoldersArgs : has no documentation (yet)
type ListFoldersArgs struct {
	// Limit : The maximum number of results to return per request.
	Limit uint32 `json:"limit"`
	// Actions : This is a list indicating whether each returned folder data
	// entry will include a boolean field `FolderPermission.allow` that
	// describes whether the current user can perform the `FolderAction` on the
	// folder.
	Actions []*FolderAction `json:"actions,omitempty"`
}

// NewListFoldersArgs returns a new ListFoldersArgs instance
func NewListFoldersArgs() *ListFoldersArgs {
	s := new(ListFoldersArgs)
	s.Limit = 1000
	return s
}

// ListFoldersContinueArg : has no documentation (yet)
type ListFoldersContinueArg struct {
	// Cursor : The cursor returned by the previous API call specified in the
	// endpoint description.
	Cursor string `json:"cursor"`
}

// NewListFoldersContinueArg returns a new ListFoldersContinueArg instance
func NewListFoldersContinueArg(Cursor string) *ListFoldersContinueArg {
	s := new(ListFoldersContinueArg)
	s.Cursor = Cursor
	return s
}

// ListFoldersContinueError : has no documentation (yet)
type ListFoldersContinueError struct {
	dropbox.Tagged
}

// Valid tag values for ListFoldersContinueError
const (
	ListFoldersContinueErrorInvalidCursor = "invalid_cursor"
	ListFoldersContinueErrorOther         = "other"
)

// ListFoldersResult : Result for `listFolders` or `listMountableFolders`,
// depending on which endpoint was requested. Unmounted shared folders can be
// identified by the absence of `SharedFolderMetadata.path_lower`.
type ListFoldersResult struct {
	// Entries : List of all shared folders the authenticated user has access
	// to.
	Entries []*SharedFolderMetadata `json:"entries"`
	// Cursor : Present if there are additional shared folders that have not
	// been returned yet. Pass the cursor into the corresponding continue
	// endpoint (either `listFoldersContinue` or `listMountableFoldersContinue`)
	// to list additional folders.
	Cursor string `json:"cursor,omitempty"`
}

// NewListFoldersResult returns a new ListFoldersResult instance
func NewListFoldersResult(Entries []*SharedFolderMetadata) *ListFoldersResult {
	s := new(ListFoldersResult)
	s.Entries = Entries
	return s
}

// ListSharedLinksArg : has no documentation (yet)
type ListSharedLinksArg struct {
	// Path : See `listSharedLinks` description.
	Path string `json:"path,omitempty"`
	// Cursor : The cursor returned by your last call to `listSharedLinks`.
	Cursor string `json:"cursor,omitempty"`
	// DirectOnly : See `listSharedLinks` description.
	DirectOnly bool `json:"direct_only,omitempty"`
}

// NewListSharedLinksArg returns a new ListSharedLinksArg instance
func NewListSharedLinksArg() *ListSharedLinksArg {
	s := new(ListSharedLinksArg)
	return s
}

// ListSharedLinksError : has no documentation (yet)
type ListSharedLinksError struct {
	dropbox.Tagged
	// Path : has no documentation (yet)
	Path *files.LookupError `json:"path,omitempty"`
}

// Valid tag values for ListSharedLinksError
const (
	ListSharedLinksErrorPath  = "path"
	ListSharedLinksErrorReset = "reset"
	ListSharedLinksErrorOther = "other"
)

// UnmarshalJSON deserializes into a ListSharedLinksError instance
func (u *ListSharedLinksError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// Path : has no documentation (yet)
		Path json.RawMessage `json:"path,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "path":
		if err := json.Unmarshal(w.Path, &u.Path); err != nil {
			return err
		}

	}
	return nil
}

// ListSharedLinksResult : has no documentation (yet)
type ListSharedLinksResult struct {
	// Links : Shared links applicable to the path argument.
	Links []IsSharedLinkMetadata `json:"links"`
	// HasMore : Is true if there are additional shared links that have not been
	// returned yet. Pass the cursor into `listSharedLinks` to retrieve them.
	HasMore bool `json:"has_more"`
	// Cursor : Pass the cursor into `listSharedLinks` to obtain the additional
	// links. Cursor is returned only if no path is given or the path is empty.
	Cursor string `json:"cursor,omitempty"`
}

// NewListSharedLinksResult returns a new ListSharedLinksResult instance
func NewListSharedLinksResult(Links []IsSharedLinkMetadata, HasMore bool) *ListSharedLinksResult {
	s := new(ListSharedLinksResult)
	s.Links = Links
	s.HasMore = HasMore
	return s
}

// MemberAccessLevelResult : Contains information about a member's access level
// to content after an operation.
type MemberAccessLevelResult struct {
	// AccessLevel : The member still has this level of access to the content
	// through a parent folder.
	AccessLevel *AccessLevel `json:"access_level,omitempty"`
	// Warning : A localized string with additional information about why the
	// user has this access level to the content.
	Warning string `json:"warning,omitempty"`
	// AccessDetails : The parent folders that a member has access to. The field
	// is present if the user has access to the first parent folder where the
	// member gains access.
	AccessDetails []*ParentFolderAccessInfo `json:"access_details,omitempty"`
}

// NewMemberAccessLevelResult returns a new MemberAccessLevelResult instance
func NewMemberAccessLevelResult() *MemberAccessLevelResult {
	s := new(MemberAccessLevelResult)
	return s
}

// MemberAction : Actions that may be taken on members of a shared folder.
type MemberAction struct {
	dropbox.Tagged
}

// Valid tag values for MemberAction
const (
	MemberActionLeaveACopy          = "leave_a_copy"
	MemberActionMakeEditor          = "make_editor"
	MemberActionMakeOwner           = "make_owner"
	MemberActionMakeViewer          = "make_viewer"
	MemberActionMakeViewerNoComment = "make_viewer_no_comment"
	MemberActionRemove              = "remove"
	MemberActionOther               = "other"
)

// MemberPermission : Whether the user is allowed to take the action on the
// associated member.
type MemberPermission struct {
	// Action : The action that the user may wish to take on the member.
	Action *MemberAction `json:"action"`
	// Allow : True if the user is allowed to take the action.
	Allow bool `json:"allow"`
	// Reason : The reason why the user is denied the permission. Not present if
	// the action is allowed
	Reason *PermissionDeniedReason `json:"reason,omitempty"`
}

// NewMemberPermission returns a new MemberPermission instance
func NewMemberPermission(Action *MemberAction, Allow bool) *MemberPermission {
	s := new(MemberPermission)
	s.Action = Action
	s.Allow = Allow
	return s
}

// MemberPolicy : Policy governing who can be a member of a shared folder. Only
// applicable to folders owned by a user on a team.
type MemberPolicy struct {
	dropbox.Tagged
}

// Valid tag values for MemberPolicy
const (
	MemberPolicyTeam   = "team"
	MemberPolicyAnyone = "anyone"
	MemberPolicyOther  = "other"
)

// MemberSelector : Includes different ways to identify a member of a shared
// folder.
type MemberSelector struct {
	dropbox.Tagged
	// DropboxId : Dropbox account, team member, or group ID of member.
	DropboxId string `json:"dropbox_id,omitempty"`
	// Email : E-mail address of member.
	Email string `json:"email,omitempty"`
}

// Valid tag values for MemberSelector
const (
	MemberSelectorDropboxId = "dropbox_id"
	MemberSelectorEmail     = "email"
	MemberSelectorOther     = "other"
)

// UnmarshalJSON deserializes into a MemberSelector instance
func (u *MemberSelector) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "dropbox_id":
		if err := json.Unmarshal(body, &u.DropboxId); err != nil {
			return err
		}

	case "email":
		if err := json.Unmarshal(body, &u.Email); err != nil {
			return err
		}

	}
	return nil
}

// ModifySharedLinkSettingsArgs : has no documentation (yet)
type ModifySharedLinkSettingsArgs struct {
	// Url : URL of the shared link to change its settings
	Url string `json:"url"`
	// Settings : Set of settings for the shared link.
	Settings *SharedLinkSettings `json:"settings"`
	// RemoveExpiration : If set to true, removes the expiration of the shared
	// link.
	RemoveExpiration bool `json:"remove_expiration"`
}

// NewModifySharedLinkSettingsArgs returns a new ModifySharedLinkSettingsArgs instance
func NewModifySharedLinkSettingsArgs(Url string, Settings *SharedLinkSettings) *ModifySharedLinkSettingsArgs {
	s := new(ModifySharedLinkSettingsArgs)
	s.Url = Url
	s.Settings = Settings
	s.RemoveExpiration = false
	return s
}

// ModifySharedLinkSettingsError : has no documentation (yet)
type ModifySharedLinkSettingsError struct {
	dropbox.Tagged
	// SettingsError : There is an error with the given settings
	SettingsError *SharedLinkSettingsError `json:"settings_error,omitempty"`
}

// Valid tag values for ModifySharedLinkSettingsError
const (
	ModifySharedLinkSettingsErrorSettingsError    = "settings_error"
	ModifySharedLinkSettingsErrorEmailNotVerified = "email_not_verified"
)

// UnmarshalJSON deserializes into a ModifySharedLinkSettingsError instance
func (u *ModifySharedLinkSettingsError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// SettingsError : There is an error with the given settings
		SettingsError json.RawMessage `json:"settings_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "settings_error":
		if err := json.Unmarshal(w.SettingsError, &u.SettingsError); err != nil {
			return err
		}

	}
	return nil
}

// MountFolderArg : has no documentation (yet)
type MountFolderArg struct {
	// SharedFolderId : The ID of the shared folder to mount.
	SharedFolderId string `json:"shared_folder_id"`
}

// NewMountFolderArg returns a new MountFolderArg instance
func NewMountFolderArg(SharedFolderId string) *MountFolderArg {
	s := new(MountFolderArg)
	s.SharedFolderId = SharedFolderId
	return s
}

// MountFolderError : has no documentation (yet)
type MountFolderError struct {
	dropbox.Tagged
	// AccessError : has no documentation (yet)
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
	// InsufficientQuota : The current user does not have enough space to mount
	// the shared folder.
	InsufficientQuota *InsufficientQuotaAmounts `json:"insufficient_quota,omitempty"`
}

// Valid tag values for MountFolderError
const (
	MountFolderErrorAccessError        = "access_error"
	MountFolderErrorInsideSharedFolder = "inside_shared_folder"
	MountFolderErrorInsufficientQuota  = "insufficient_quota"
	MountFolderErrorAlreadyMounted     = "already_mounted"
	MountFolderErrorNoPermission       = "no_permission"
	MountFolderErrorNotMountable       = "not_mountable"
	MountFolderErrorOther              = "other"
)

// UnmarshalJSON deserializes into a MountFolderError instance
func (u *MountFolderError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// AccessError : has no documentation (yet)
		AccessError json.RawMessage `json:"access_error,omitempty"`
		// InsufficientQuota : The current user does not have enough space to
		// mount the shared folder.
		InsufficientQuota json.RawMessage `json:"insufficient_quota,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	case "insufficient_quota":
		if err := json.Unmarshal(body, &u.InsufficientQuota); err != nil {
			return err
		}

	}
	return nil
}

// ParentFolderAccessInfo : Contains information about a parent folder that a
// member has access to.
type ParentFolderAccessInfo struct {
	// FolderName : Display name for the folder.
	FolderName string `json:"folder_name"`
	// SharedFolderId : The identifier of the parent shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// Permissions : The user's permissions for the parent shared folder.
	Permissions []*MemberPermission `json:"permissions"`
}

// NewParentFolderAccessInfo returns a new ParentFolderAccessInfo instance
func NewParentFolderAccessInfo(FolderName string, SharedFolderId string, Permissions []*MemberPermission) *ParentFolderAccessInfo {
	s := new(ParentFolderAccessInfo)
	s.FolderName = FolderName
	s.SharedFolderId = SharedFolderId
	s.Permissions = Permissions
	return s
}

// PathLinkMetadata : Metadata for a path-based shared link.
type PathLinkMetadata struct {
	LinkMetadata
	// Path : Path in user's Dropbox.
	Path string `json:"path"`
}

// NewPathLinkMetadata returns a new PathLinkMetadata instance
func NewPathLinkMetadata(Url string, Visibility *Visibility, Path string) *PathLinkMetadata {
	s := new(PathLinkMetadata)
	s.Url = Url
	s.Visibility = Visibility
	s.Path = Path
	return s
}

// PendingUploadMode : Flag to indicate pending upload default (for linking to
// not-yet-existing paths).
type PendingUploadMode struct {
	dropbox.Tagged
}

// Valid tag values for PendingUploadMode
const (
	PendingUploadModeFile   = "file"
	PendingUploadModeFolder = "folder"
)

// PermissionDeniedReason : Possible reasons the user is denied a permission.
type PermissionDeniedReason struct {
	dropbox.Tagged
}

// Valid tag values for PermissionDeniedReason
const (
	PermissionDeniedReasonUserNotSameTeamAsOwner    = "user_not_same_team_as_owner"
	PermissionDeniedReasonUserNotAllowedByOwner     = "user_not_allowed_by_owner"
	PermissionDeniedReasonTargetIsIndirectMember    = "target_is_indirect_member"
	PermissionDeniedReasonTargetIsOwner             = "target_is_owner"
	PermissionDeniedReasonTargetIsSelf              = "target_is_self"
	PermissionDeniedReasonTargetNotActive           = "target_not_active"
	PermissionDeniedReasonFolderIsLimitedTeamFolder = "folder_is_limited_team_folder"
	PermissionDeniedReasonOther                     = "other"
)

// RelinquishFileMembershipArg : has no documentation (yet)
type RelinquishFileMembershipArg struct {
	// File : The path or id for the file.
	File string `json:"file"`
}

// NewRelinquishFileMembershipArg returns a new RelinquishFileMembershipArg instance
func NewRelinquishFileMembershipArg(File string) *RelinquishFileMembershipArg {
	s := new(RelinquishFileMembershipArg)
	s.File = File
	return s
}

// RelinquishFileMembershipError : has no documentation (yet)
type RelinquishFileMembershipError struct {
	dropbox.Tagged
	// AccessError : has no documentation (yet)
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
}

// Valid tag values for RelinquishFileMembershipError
const (
	RelinquishFileMembershipErrorAccessError  = "access_error"
	RelinquishFileMembershipErrorGroupAccess  = "group_access"
	RelinquishFileMembershipErrorNoPermission = "no_permission"
	RelinquishFileMembershipErrorOther        = "other"
)

// UnmarshalJSON deserializes into a RelinquishFileMembershipError instance
func (u *RelinquishFileMembershipError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// AccessError : has no documentation (yet)
		AccessError json.RawMessage `json:"access_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	}
	return nil
}

// RelinquishFolderMembershipArg : has no documentation (yet)
type RelinquishFolderMembershipArg struct {
	// SharedFolderId : The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// LeaveACopy : Keep a copy of the folder's contents upon relinquishing
	// membership.
	LeaveACopy bool `json:"leave_a_copy"`
}

// NewRelinquishFolderMembershipArg returns a new RelinquishFolderMembershipArg instance
func NewRelinquishFolderMembershipArg(SharedFolderId string) *RelinquishFolderMembershipArg {
	s := new(RelinquishFolderMembershipArg)
	s.SharedFolderId = SharedFolderId
	s.LeaveACopy = false
	return s
}

// RelinquishFolderMembershipError : has no documentation (yet)
type RelinquishFolderMembershipError struct {
	dropbox.Tagged
	// AccessError : has no documentation (yet)
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
}

// Valid tag values for RelinquishFolderMembershipError
const (
	RelinquishFolderMembershipErrorAccessError  = "access_error"
	RelinquishFolderMembershipErrorFolderOwner  = "folder_owner"
	RelinquishFolderMembershipErrorMounted      = "mounted"
	RelinquishFolderMembershipErrorGroupAccess  = "group_access"
	RelinquishFolderMembershipErrorTeamFolder   = "team_folder"
	RelinquishFolderMembershipErrorNoPermission = "no_permission"
	RelinquishFolderMembershipErrorOther        = "other"
)

// UnmarshalJSON deserializes into a RelinquishFolderMembershipError instance
func (u *RelinquishFolderMembershipError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// AccessError : has no documentation (yet)
		AccessError json.RawMessage `json:"access_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	}
	return nil
}

// RemoveFileMemberArg : Arguments for `removeFileMember2`.
type RemoveFileMemberArg struct {
	// File : File from which to remove members.
	File string `json:"file"`
	// Member : Member to remove from this file. Note that even if an email is
	// specified, it may result in the removal of a user (not an invitee) if the
	// user's main account corresponds to that email address.
	Member *MemberSelector `json:"member"`
}

// NewRemoveFileMemberArg returns a new RemoveFileMemberArg instance
func NewRemoveFileMemberArg(File string, Member *MemberSelector) *RemoveFileMemberArg {
	s := new(RemoveFileMemberArg)
	s.File = File
	s.Member = Member
	return s
}

// RemoveFileMemberError : Errors for `removeFileMember2`.
type RemoveFileMemberError struct {
	dropbox.Tagged
	// UserError : has no documentation (yet)
	UserError *SharingUserError `json:"user_error,omitempty"`
	// AccessError : has no documentation (yet)
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
	// NoExplicitAccess : This member does not have explicit access to the file
	// and therefore cannot be removed. The return value is the access that a
	// user might have to the file from a parent folder.
	NoExplicitAccess *MemberAccessLevelResult `json:"no_explicit_access,omitempty"`
}

// Valid tag values for RemoveFileMemberError
const (
	RemoveFileMemberErrorUserError        = "user_error"
	RemoveFileMemberErrorAccessError      = "access_error"
	RemoveFileMemberErrorNoExplicitAccess = "no_explicit_access"
	RemoveFileMemberErrorOther            = "other"
)

// UnmarshalJSON deserializes into a RemoveFileMemberError instance
func (u *RemoveFileMemberError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// UserError : has no documentation (yet)
		UserError json.RawMessage `json:"user_error,omitempty"`
		// AccessError : has no documentation (yet)
		AccessError json.RawMessage `json:"access_error,omitempty"`
		// NoExplicitAccess : This member does not have explicit access to the
		// file and therefore cannot be removed. The return value is the access
		// that a user might have to the file from a parent folder.
		NoExplicitAccess json.RawMessage `json:"no_explicit_access,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "user_error":
		if err := json.Unmarshal(w.UserError, &u.UserError); err != nil {
			return err
		}

	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	case "no_explicit_access":
		if err := json.Unmarshal(body, &u.NoExplicitAccess); err != nil {
			return err
		}

	}
	return nil
}

// RemoveFolderMemberArg : has no documentation (yet)
type RemoveFolderMemberArg struct {
	// SharedFolderId : The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// Member : The member to remove from the folder.
	Member *MemberSelector `json:"member"`
	// LeaveACopy : If true, the removed user will keep their copy of the folder
	// after it's unshared, assuming it was mounted. Otherwise, it will be
	// removed from their Dropbox. Also, this must be set to false when kicking
	// a group.
	LeaveACopy bool `json:"leave_a_copy"`
}

// NewRemoveFolderMemberArg returns a new RemoveFolderMemberArg instance
func NewRemoveFolderMemberArg(SharedFolderId string, Member *MemberSelector, LeaveACopy bool) *RemoveFolderMemberArg {
	s := new(RemoveFolderMemberArg)
	s.SharedFolderId = SharedFolderId
	s.Member = Member
	s.LeaveACopy = LeaveACopy
	return s
}

// RemoveFolderMemberError : has no documentation (yet)
type RemoveFolderMemberError struct {
	dropbox.Tagged
	// AccessError : has no documentation (yet)
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
	// MemberError : has no documentation (yet)
	MemberError *SharedFolderMemberError `json:"member_error,omitempty"`
}

// Valid tag values for RemoveFolderMemberError
const (
	RemoveFolderMemberErrorAccessError  = "access_error"
	RemoveFolderMemberErrorMemberError  = "member_error"
	RemoveFolderMemberErrorFolderOwner  = "folder_owner"
	RemoveFolderMemberErrorGroupAccess  = "group_access"
	RemoveFolderMemberErrorTeamFolder   = "team_folder"
	RemoveFolderMemberErrorNoPermission = "no_permission"
	RemoveFolderMemberErrorOther        = "other"
)

// UnmarshalJSON deserializes into a RemoveFolderMemberError instance
func (u *RemoveFolderMemberError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// AccessError : has no documentation (yet)
		AccessError json.RawMessage `json:"access_error,omitempty"`
		// MemberError : has no documentation (yet)
		MemberError json.RawMessage `json:"member_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	case "member_error":
		if err := json.Unmarshal(w.MemberError, &u.MemberError); err != nil {
			return err
		}

	}
	return nil
}

// RemoveMemberJobStatus : has no documentation (yet)
type RemoveMemberJobStatus struct {
	dropbox.Tagged
	// Complete : Removing the folder member has finished. The value is
	// information about whether the member has another form of access.
	Complete *MemberAccessLevelResult `json:"complete,omitempty"`
	// Failed : has no documentation (yet)
	Failed *RemoveFolderMemberError `json:"failed,omitempty"`
}

// Valid tag values for RemoveMemberJobStatus
const (
	RemoveMemberJobStatusComplete = "complete"
	RemoveMemberJobStatusFailed   = "failed"
)

// UnmarshalJSON deserializes into a RemoveMemberJobStatus instance
func (u *RemoveMemberJobStatus) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// Complete : Removing the folder member has finished. The value is
		// information about whether the member has another form of access.
		Complete json.RawMessage `json:"complete,omitempty"`
		// Failed : has no documentation (yet)
		Failed json.RawMessage `json:"failed,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "complete":
		if err := json.Unmarshal(body, &u.Complete); err != nil {
			return err
		}

	case "failed":
		if err := json.Unmarshal(w.Failed, &u.Failed); err != nil {
			return err
		}

	}
	return nil
}

// RequestedVisibility : The access permission that can be requested by the
// caller for the shared link. Note that the final resolved visibility of the
// shared link takes into account other aspects, such as team and shared folder
// settings. Check the `ResolvedVisibility` for more info on the possible
// resolved visibility values of shared links.
type RequestedVisibility struct {
	dropbox.Tagged
}

// Valid tag values for RequestedVisibility
const (
	RequestedVisibilityPublic   = "public"
	RequestedVisibilityTeamOnly = "team_only"
	RequestedVisibilityPassword = "password"
)

// ResolvedVisibility : The actual access permissions values of shared links
// after taking into account user preferences and the team and shared folder
// settings. Check the `RequestedVisibility` for more info on the possible
// visibility values that can be set by the shared link's owner.
type ResolvedVisibility struct {
	dropbox.Tagged
}

// Valid tag values for ResolvedVisibility
const (
	ResolvedVisibilityTeamAndPassword  = "team_and_password"
	ResolvedVisibilitySharedFolderOnly = "shared_folder_only"
	ResolvedVisibilityOther            = "other"
)

// RevokeSharedLinkArg : has no documentation (yet)
type RevokeSharedLinkArg struct {
	// Url : URL of the shared link.
	Url string `json:"url"`
}

// NewRevokeSharedLinkArg returns a new RevokeSharedLinkArg instance
func NewRevokeSharedLinkArg(Url string) *RevokeSharedLinkArg {
	s := new(RevokeSharedLinkArg)
	s.Url = Url
	return s
}

// RevokeSharedLinkError : has no documentation (yet)
type RevokeSharedLinkError struct {
	dropbox.Tagged
}

// Valid tag values for RevokeSharedLinkError
const (
	RevokeSharedLinkErrorSharedLinkMalformed = "shared_link_malformed"
)

// ShareFolderArg : has no documentation (yet)
type ShareFolderArg struct {
	// Path : The path to the folder to share. If it does not exist, then a new
	// one is created.
	Path string `json:"path"`
	// MemberPolicy : Who can be a member of this shared folder. Only applicable
	// if the current user is on a team.
	MemberPolicy *MemberPolicy `json:"member_policy"`
	// AclUpdatePolicy : Who can add and remove members of this shared folder.
	AclUpdatePolicy *AclUpdatePolicy `json:"acl_update_policy"`
	// SharedLinkPolicy : The policy to apply to shared links created for
	// content inside this shared folder.  The current user must be on a team to
	// set this policy to `SharedLinkPolicy.members`.
	SharedLinkPolicy *SharedLinkPolicy `json:"shared_link_policy"`
	// ForceAsync : Whether to force the share to happen asynchronously.
	ForceAsync bool `json:"force_async"`
}

// NewShareFolderArg returns a new ShareFolderArg instance
func NewShareFolderArg(Path string) *ShareFolderArg {
	s := new(ShareFolderArg)
	s.Path = Path
	s.MemberPolicy = &MemberPolicy{Tagged: dropbox.Tagged{"anyone"}}
	s.AclUpdatePolicy = &AclUpdatePolicy{Tagged: dropbox.Tagged{"owner"}}
	s.SharedLinkPolicy = &SharedLinkPolicy{Tagged: dropbox.Tagged{"anyone"}}
	s.ForceAsync = false
	return s
}

// ShareFolderErrorBase : has no documentation (yet)
type ShareFolderErrorBase struct {
	dropbox.Tagged
	// BadPath : `ShareFolderArg.path` is invalid.
	BadPath *SharePathError `json:"bad_path,omitempty"`
}

// Valid tag values for ShareFolderErrorBase
const (
	ShareFolderErrorBaseEmailUnverified                 = "email_unverified"
	ShareFolderErrorBaseBadPath                         = "bad_path"
	ShareFolderErrorBaseTeamPolicyDisallowsMemberPolicy = "team_policy_disallows_member_policy"
	ShareFolderErrorBaseDisallowedSharedLinkPolicy      = "disallowed_shared_link_policy"
	ShareFolderErrorBaseOther                           = "other"
)

// UnmarshalJSON deserializes into a ShareFolderErrorBase instance
func (u *ShareFolderErrorBase) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// BadPath : `ShareFolderArg.path` is invalid.
		BadPath json.RawMessage `json:"bad_path,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "bad_path":
		if err := json.Unmarshal(w.BadPath, &u.BadPath); err != nil {
			return err
		}

	}
	return nil
}

// ShareFolderError : has no documentation (yet)
type ShareFolderError struct {
	dropbox.Tagged
}

// Valid tag values for ShareFolderError
const (
	ShareFolderErrorNoPermission = "no_permission"
)

// ShareFolderJobStatus : has no documentation (yet)
type ShareFolderJobStatus struct {
	dropbox.Tagged
	// Complete : The share job has finished. The value is the metadata for the
	// folder.
	Complete *SharedFolderMetadata `json:"complete,omitempty"`
	// Failed : has no documentation (yet)
	Failed *ShareFolderError `json:"failed,omitempty"`
}

// Valid tag values for ShareFolderJobStatus
const (
	ShareFolderJobStatusComplete = "complete"
	ShareFolderJobStatusFailed   = "failed"
)

// UnmarshalJSON deserializes into a ShareFolderJobStatus instance
func (u *ShareFolderJobStatus) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// Complete : The share job has finished. The value is the metadata for
		// the folder.
		Complete json.RawMessage `json:"complete,omitempty"`
		// Failed : has no documentation (yet)
		Failed json.RawMessage `json:"failed,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "complete":
		if err := json.Unmarshal(body, &u.Complete); err != nil {
			return err
		}

	case "failed":
		if err := json.Unmarshal(w.Failed, &u.Failed); err != nil {
			return err
		}

	}
	return nil
}

// ShareFolderLaunch : has no documentation (yet)
type ShareFolderLaunch struct {
	dropbox.Tagged
	// Complete : has no documentation (yet)
	Complete *SharedFolderMetadata `json:"complete,omitempty"`
}

// Valid tag values for ShareFolderLaunch
const (
	ShareFolderLaunchComplete = "complete"
)

// UnmarshalJSON deserializes into a ShareFolderLaunch instance
func (u *ShareFolderLaunch) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// Complete : has no documentation (yet)
		Complete json.RawMessage `json:"complete,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "complete":
		if err := json.Unmarshal(body, &u.Complete); err != nil {
			return err
		}

	}
	return nil
}

// SharePathError : has no documentation (yet)
type SharePathError struct {
	dropbox.Tagged
	// AlreadyShared : Folder is already shared. Contains metadata about the
	// existing shared folder.
	AlreadyShared *SharedFolderMetadata `json:"already_shared,omitempty"`
}

// Valid tag values for SharePathError
const (
	SharePathErrorIsFile               = "is_file"
	SharePathErrorInsideSharedFolder   = "inside_shared_folder"
	SharePathErrorContainsSharedFolder = "contains_shared_folder"
	SharePathErrorIsAppFolder          = "is_app_folder"
	SharePathErrorInsideAppFolder      = "inside_app_folder"
	SharePathErrorIsPublicFolder       = "is_public_folder"
	SharePathErrorInsidePublicFolder   = "inside_public_folder"
	SharePathErrorAlreadyShared        = "already_shared"
	SharePathErrorInvalidPath          = "invalid_path"
	SharePathErrorIsOsxPackage         = "is_osx_package"
	SharePathErrorInsideOsxPackage     = "inside_osx_package"
	SharePathErrorOther                = "other"
)

// UnmarshalJSON deserializes into a SharePathError instance
func (u *SharePathError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// AlreadyShared : Folder is already shared. Contains metadata about the
		// existing shared folder.
		AlreadyShared json.RawMessage `json:"already_shared,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "already_shared":
		if err := json.Unmarshal(body, &u.AlreadyShared); err != nil {
			return err
		}

	}
	return nil
}

// SharedFileMembers : Shared file user, group, and invitee membership. Used for
// the results of `listFileMembers` and `listFileMembersContinue`, and used as
// part of the results for `listFileMembersBatch`.
type SharedFileMembers struct {
	// Users : The list of user members of the shared file.
	Users []*UserMembershipInfo `json:"users"`
	// Groups : The list of group members of the shared file.
	Groups []*GroupMembershipInfo `json:"groups"`
	// Invitees : The list of invited members of a file, but have not logged in
	// and claimed this.
	Invitees []*InviteeMembershipInfo `json:"invitees"`
	// Cursor : Present if there are additional shared file members that have
	// not been returned yet. Pass the cursor into `listFileMembersContinue` to
	// list additional members.
	Cursor string `json:"cursor,omitempty"`
}

// NewSharedFileMembers returns a new SharedFileMembers instance
func NewSharedFileMembers(Users []*UserMembershipInfo, Groups []*GroupMembershipInfo, Invitees []*InviteeMembershipInfo) *SharedFileMembers {
	s := new(SharedFileMembers)
	s.Users = Users
	s.Groups = Groups
	s.Invitees = Invitees
	return s
}

// SharedFileMetadata : Properties of the shared file.
type SharedFileMetadata struct {
	// Policy : Policies governing this shared file.
	Policy *FolderPolicy `json:"policy"`
	// Permissions : The sharing permissions that requesting user has on this
	// file. This corresponds to the entries given in
	// `GetFileMetadataBatchArg.actions` or `GetFileMetadataArg.actions`.
	Permissions []*FilePermission `json:"permissions,omitempty"`
	// OwnerTeam : The team that owns the file. This field is not present if the
	// file is not owned by a team.
	OwnerTeam *users.Team `json:"owner_team,omitempty"`
	// ParentSharedFolderId : The ID of the parent shared folder. This field is
	// present only if the file is contained within a shared folder.
	ParentSharedFolderId string `json:"parent_shared_folder_id,omitempty"`
	// PreviewUrl : URL for displaying a web preview of the shared file.
	PreviewUrl string `json:"preview_url"`
	// PathLower : The lower-case full path of this file. Absent for unmounted
	// files.
	PathLower string `json:"path_lower,omitempty"`
	// PathDisplay : The cased path to be used for display purposes only. In
	// rare instances the casing will not correctly match the user's filesystem,
	// but this behavior will match the path provided in the Core API v1. Absent
	// for unmounted files.
	PathDisplay string `json:"path_display,omitempty"`
	// Name : The name of this file.
	Name string `json:"name"`
	// Id : The ID of the file.
	Id string `json:"id"`
}

// NewSharedFileMetadata returns a new SharedFileMetadata instance
func NewSharedFileMetadata(Policy *FolderPolicy, PreviewUrl string, Name string, Id string) *SharedFileMetadata {
	s := new(SharedFileMetadata)
	s.Policy = Policy
	s.PreviewUrl = PreviewUrl
	s.Name = Name
	s.Id = Id
	return s
}

// SharedFolderAccessError : There is an error accessing the shared folder.
type SharedFolderAccessError struct {
	dropbox.Tagged
}

// Valid tag values for SharedFolderAccessError
const (
	SharedFolderAccessErrorInvalidId       = "invalid_id"
	SharedFolderAccessErrorNotAMember      = "not_a_member"
	SharedFolderAccessErrorEmailUnverified = "email_unverified"
	SharedFolderAccessErrorUnmounted       = "unmounted"
	SharedFolderAccessErrorOther           = "other"
)

// SharedFolderMemberError : has no documentation (yet)
type SharedFolderMemberError struct {
	dropbox.Tagged
	// NoExplicitAccess : The target member only has inherited access to the
	// shared folder.
	NoExplicitAccess *MemberAccessLevelResult `json:"no_explicit_access,omitempty"`
}

// Valid tag values for SharedFolderMemberError
const (
	SharedFolderMemberErrorInvalidDropboxId = "invalid_dropbox_id"
	SharedFolderMemberErrorNotAMember       = "not_a_member"
	SharedFolderMemberErrorNoExplicitAccess = "no_explicit_access"
	SharedFolderMemberErrorOther            = "other"
)

// UnmarshalJSON deserializes into a SharedFolderMemberError instance
func (u *SharedFolderMemberError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// NoExplicitAccess : The target member only has inherited access to the
		// shared folder.
		NoExplicitAccess json.RawMessage `json:"no_explicit_access,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "no_explicit_access":
		if err := json.Unmarshal(body, &u.NoExplicitAccess); err != nil {
			return err
		}

	}
	return nil
}

// SharedFolderMembers : Shared folder user and group membership.
type SharedFolderMembers struct {
	// Users : The list of user members of the shared folder.
	Users []*UserMembershipInfo `json:"users"`
	// Groups : The list of group members of the shared folder.
	Groups []*GroupMembershipInfo `json:"groups"`
	// Invitees : The list of invitees to the shared folder.
	Invitees []*InviteeMembershipInfo `json:"invitees"`
	// Cursor : Present if there are additional shared folder members that have
	// not been returned yet. Pass the cursor into `listFolderMembersContinue`
	// to list additional members.
	Cursor string `json:"cursor,omitempty"`
}

// NewSharedFolderMembers returns a new SharedFolderMembers instance
func NewSharedFolderMembers(Users []*UserMembershipInfo, Groups []*GroupMembershipInfo, Invitees []*InviteeMembershipInfo) *SharedFolderMembers {
	s := new(SharedFolderMembers)
	s.Users = Users
	s.Groups = Groups
	s.Invitees = Invitees
	return s
}

// SharedFolderMetadataBase : Properties of the shared folder.
type SharedFolderMetadataBase struct {
	// AccessType : The current user's access level for this shared folder.
	AccessType *AccessLevel `json:"access_type"`
	// IsTeamFolder : Whether this folder is a `team folder`
	// <https://www.dropbox.com/en/help/986>.
	IsTeamFolder bool `json:"is_team_folder"`
	// Policy : Policies governing this shared folder.
	Policy *FolderPolicy `json:"policy"`
	// OwnerTeam : The team that owns the folder. This field is not present if
	// the folder is not owned by a team.
	OwnerTeam *users.Team `json:"owner_team,omitempty"`
	// ParentSharedFolderId : The ID of the parent shared folder. This field is
	// present only if the folder is contained within another shared folder.
	ParentSharedFolderId string `json:"parent_shared_folder_id,omitempty"`
}

// NewSharedFolderMetadataBase returns a new SharedFolderMetadataBase instance
func NewSharedFolderMetadataBase(AccessType *AccessLevel, IsTeamFolder bool, Policy *FolderPolicy) *SharedFolderMetadataBase {
	s := new(SharedFolderMetadataBase)
	s.AccessType = AccessType
	s.IsTeamFolder = IsTeamFolder
	s.Policy = Policy
	return s
}

// SharedFolderMetadata : The metadata which includes basic information about
// the shared folder.
type SharedFolderMetadata struct {
	SharedFolderMetadataBase
	// PathLower : The lower-cased full path of this shared folder. Absent for
	// unmounted folders.
	PathLower string `json:"path_lower,omitempty"`
	// Name : The name of the this shared folder.
	Name string `json:"name"`
	// SharedFolderId : The ID of the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// Permissions : Actions the current user may perform on the folder and its
	// contents. The set of permissions corresponds to the FolderActions in the
	// request.
	Permissions []*FolderPermission `json:"permissions,omitempty"`
	// TimeInvited : Timestamp indicating when the current user was invited to
	// this shared folder.
	TimeInvited time.Time `json:"time_invited"`
	// PreviewUrl : URL for displaying a web preview of the shared folder.
	PreviewUrl string `json:"preview_url"`
}

// NewSharedFolderMetadata returns a new SharedFolderMetadata instance
func NewSharedFolderMetadata(AccessType *AccessLevel, IsTeamFolder bool, Policy *FolderPolicy, Name string, SharedFolderId string, TimeInvited time.Time, PreviewUrl string) *SharedFolderMetadata {
	s := new(SharedFolderMetadata)
	s.AccessType = AccessType
	s.IsTeamFolder = IsTeamFolder
	s.Policy = Policy
	s.Name = Name
	s.SharedFolderId = SharedFolderId
	s.TimeInvited = TimeInvited
	s.PreviewUrl = PreviewUrl
	return s
}

// SharedLinkAccessFailureReason : has no documentation (yet)
type SharedLinkAccessFailureReason struct {
	dropbox.Tagged
}

// Valid tag values for SharedLinkAccessFailureReason
const (
	SharedLinkAccessFailureReasonLoginRequired       = "login_required"
	SharedLinkAccessFailureReasonEmailVerifyRequired = "email_verify_required"
	SharedLinkAccessFailureReasonPasswordRequired    = "password_required"
	SharedLinkAccessFailureReasonTeamOnly            = "team_only"
	SharedLinkAccessFailureReasonOwnerOnly           = "owner_only"
	SharedLinkAccessFailureReasonOther               = "other"
)

// SharedLinkPolicy : Policy governing who can view shared links.
type SharedLinkPolicy struct {
	dropbox.Tagged
}

// Valid tag values for SharedLinkPolicy
const (
	SharedLinkPolicyAnyone  = "anyone"
	SharedLinkPolicyMembers = "members"
	SharedLinkPolicyOther   = "other"
)

// SharedLinkSettings : has no documentation (yet)
type SharedLinkSettings struct {
	// RequestedVisibility : The requested access for this shared link.
	RequestedVisibility *RequestedVisibility `json:"requested_visibility,omitempty"`
	// LinkPassword : If `requested_visibility` is
	// `RequestedVisibility.password` this is needed to specify the password to
	// access the link.
	LinkPassword string `json:"link_password,omitempty"`
	// Expires : Expiration time of the shared link. By default the link won't
	// expire.
	Expires time.Time `json:"expires,omitempty"`
}

// NewSharedLinkSettings returns a new SharedLinkSettings instance
func NewSharedLinkSettings() *SharedLinkSettings {
	s := new(SharedLinkSettings)
	return s
}

// SharedLinkSettingsError : has no documentation (yet)
type SharedLinkSettingsError struct {
	dropbox.Tagged
}

// Valid tag values for SharedLinkSettingsError
const (
	SharedLinkSettingsErrorInvalidSettings = "invalid_settings"
	SharedLinkSettingsErrorNotAuthorized   = "not_authorized"
)

// SharingFileAccessError : User could not access this file.
type SharingFileAccessError struct {
	dropbox.Tagged
}

// Valid tag values for SharingFileAccessError
const (
	SharingFileAccessErrorNoPermission       = "no_permission"
	SharingFileAccessErrorInvalidFile        = "invalid_file"
	SharingFileAccessErrorIsFolder           = "is_folder"
	SharingFileAccessErrorInsidePublicFolder = "inside_public_folder"
	SharingFileAccessErrorInsideOsxPackage   = "inside_osx_package"
	SharingFileAccessErrorOther              = "other"
)

// SharingUserError : User account had a problem preventing this action.
type SharingUserError struct {
	dropbox.Tagged
}

// Valid tag values for SharingUserError
const (
	SharingUserErrorEmailUnverified = "email_unverified"
	SharingUserErrorOther           = "other"
)

// TeamMemberInfo : Information about a team member.
type TeamMemberInfo struct {
	// TeamInfo : Information about the member's team
	TeamInfo *users.Team `json:"team_info"`
	// DisplayName : The display name of the user.
	DisplayName string `json:"display_name"`
	// MemberId : ID of user as a member of a team. This field will only be
	// present if the member is in the same team as current user.
	MemberId string `json:"member_id,omitempty"`
}

// NewTeamMemberInfo returns a new TeamMemberInfo instance
func NewTeamMemberInfo(TeamInfo *users.Team, DisplayName string) *TeamMemberInfo {
	s := new(TeamMemberInfo)
	s.TeamInfo = TeamInfo
	s.DisplayName = DisplayName
	return s
}

// TransferFolderArg : has no documentation (yet)
type TransferFolderArg struct {
	// SharedFolderId : The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// ToDropboxId : A account or team member ID to transfer ownership to.
	ToDropboxId string `json:"to_dropbox_id"`
}

// NewTransferFolderArg returns a new TransferFolderArg instance
func NewTransferFolderArg(SharedFolderId string, ToDropboxId string) *TransferFolderArg {
	s := new(TransferFolderArg)
	s.SharedFolderId = SharedFolderId
	s.ToDropboxId = ToDropboxId
	return s
}

// TransferFolderError : has no documentation (yet)
type TransferFolderError struct {
	dropbox.Tagged
	// AccessError : has no documentation (yet)
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
}

// Valid tag values for TransferFolderError
const (
	TransferFolderErrorAccessError             = "access_error"
	TransferFolderErrorInvalidDropboxId        = "invalid_dropbox_id"
	TransferFolderErrorNewOwnerNotAMember      = "new_owner_not_a_member"
	TransferFolderErrorNewOwnerUnmounted       = "new_owner_unmounted"
	TransferFolderErrorNewOwnerEmailUnverified = "new_owner_email_unverified"
	TransferFolderErrorTeamFolder              = "team_folder"
	TransferFolderErrorNoPermission            = "no_permission"
	TransferFolderErrorOther                   = "other"
)

// UnmarshalJSON deserializes into a TransferFolderError instance
func (u *TransferFolderError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// AccessError : has no documentation (yet)
		AccessError json.RawMessage `json:"access_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	}
	return nil
}

// UnmountFolderArg : has no documentation (yet)
type UnmountFolderArg struct {
	// SharedFolderId : The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
}

// NewUnmountFolderArg returns a new UnmountFolderArg instance
func NewUnmountFolderArg(SharedFolderId string) *UnmountFolderArg {
	s := new(UnmountFolderArg)
	s.SharedFolderId = SharedFolderId
	return s
}

// UnmountFolderError : has no documentation (yet)
type UnmountFolderError struct {
	dropbox.Tagged
	// AccessError : has no documentation (yet)
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
}

// Valid tag values for UnmountFolderError
const (
	UnmountFolderErrorAccessError    = "access_error"
	UnmountFolderErrorNoPermission   = "no_permission"
	UnmountFolderErrorNotUnmountable = "not_unmountable"
	UnmountFolderErrorOther          = "other"
)

// UnmarshalJSON deserializes into a UnmountFolderError instance
func (u *UnmountFolderError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// AccessError : has no documentation (yet)
		AccessError json.RawMessage `json:"access_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	}
	return nil
}

// UnshareFileArg : Arguments for `unshareFile`.
type UnshareFileArg struct {
	// File : The file to unshare.
	File string `json:"file"`
}

// NewUnshareFileArg returns a new UnshareFileArg instance
func NewUnshareFileArg(File string) *UnshareFileArg {
	s := new(UnshareFileArg)
	s.File = File
	return s
}

// UnshareFileError : Error result for `unshareFile`.
type UnshareFileError struct {
	dropbox.Tagged
	// UserError : has no documentation (yet)
	UserError *SharingUserError `json:"user_error,omitempty"`
	// AccessError : has no documentation (yet)
	AccessError *SharingFileAccessError `json:"access_error,omitempty"`
}

// Valid tag values for UnshareFileError
const (
	UnshareFileErrorUserError   = "user_error"
	UnshareFileErrorAccessError = "access_error"
	UnshareFileErrorOther       = "other"
)

// UnmarshalJSON deserializes into a UnshareFileError instance
func (u *UnshareFileError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// UserError : has no documentation (yet)
		UserError json.RawMessage `json:"user_error,omitempty"`
		// AccessError : has no documentation (yet)
		AccessError json.RawMessage `json:"access_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "user_error":
		if err := json.Unmarshal(w.UserError, &u.UserError); err != nil {
			return err
		}

	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	}
	return nil
}

// UnshareFolderArg : has no documentation (yet)
type UnshareFolderArg struct {
	// SharedFolderId : The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// LeaveACopy : If true, members of this shared folder will get a copy of
	// this folder after it's unshared. Otherwise, it will be removed from their
	// Dropbox. The current user, who is an owner, will always retain their
	// copy.
	LeaveACopy bool `json:"leave_a_copy"`
}

// NewUnshareFolderArg returns a new UnshareFolderArg instance
func NewUnshareFolderArg(SharedFolderId string) *UnshareFolderArg {
	s := new(UnshareFolderArg)
	s.SharedFolderId = SharedFolderId
	s.LeaveACopy = false
	return s
}

// UnshareFolderError : has no documentation (yet)
type UnshareFolderError struct {
	dropbox.Tagged
	// AccessError : has no documentation (yet)
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
}

// Valid tag values for UnshareFolderError
const (
	UnshareFolderErrorAccessError  = "access_error"
	UnshareFolderErrorTeamFolder   = "team_folder"
	UnshareFolderErrorNoPermission = "no_permission"
	UnshareFolderErrorOther        = "other"
)

// UnmarshalJSON deserializes into a UnshareFolderError instance
func (u *UnshareFolderError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// AccessError : has no documentation (yet)
		AccessError json.RawMessage `json:"access_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	}
	return nil
}

// UpdateFolderMemberArg : has no documentation (yet)
type UpdateFolderMemberArg struct {
	// SharedFolderId : The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// Member : The member of the shared folder to update.  Only the
	// `MemberSelector.dropbox_id` may be set at this time.
	Member *MemberSelector `json:"member"`
	// AccessLevel : The new access level for `member`. `AccessLevel.owner` is
	// disallowed.
	AccessLevel *AccessLevel `json:"access_level"`
}

// NewUpdateFolderMemberArg returns a new UpdateFolderMemberArg instance
func NewUpdateFolderMemberArg(SharedFolderId string, Member *MemberSelector, AccessLevel *AccessLevel) *UpdateFolderMemberArg {
	s := new(UpdateFolderMemberArg)
	s.SharedFolderId = SharedFolderId
	s.Member = Member
	s.AccessLevel = AccessLevel
	return s
}

// UpdateFolderMemberError : has no documentation (yet)
type UpdateFolderMemberError struct {
	dropbox.Tagged
	// AccessError : has no documentation (yet)
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
	// MemberError : has no documentation (yet)
	MemberError *SharedFolderMemberError `json:"member_error,omitempty"`
	// NoExplicitAccess : If updating the access type required the member to be
	// added to the shared folder and there was an error when adding the member.
	NoExplicitAccess *AddFolderMemberError `json:"no_explicit_access,omitempty"`
}

// Valid tag values for UpdateFolderMemberError
const (
	UpdateFolderMemberErrorAccessError      = "access_error"
	UpdateFolderMemberErrorMemberError      = "member_error"
	UpdateFolderMemberErrorNoExplicitAccess = "no_explicit_access"
	UpdateFolderMemberErrorInsufficientPlan = "insufficient_plan"
	UpdateFolderMemberErrorNoPermission     = "no_permission"
	UpdateFolderMemberErrorOther            = "other"
)

// UnmarshalJSON deserializes into a UpdateFolderMemberError instance
func (u *UpdateFolderMemberError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// AccessError : has no documentation (yet)
		AccessError json.RawMessage `json:"access_error,omitempty"`
		// MemberError : has no documentation (yet)
		MemberError json.RawMessage `json:"member_error,omitempty"`
		// NoExplicitAccess : If updating the access type required the member to
		// be added to the shared folder and there was an error when adding the
		// member.
		NoExplicitAccess json.RawMessage `json:"no_explicit_access,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	case "member_error":
		if err := json.Unmarshal(w.MemberError, &u.MemberError); err != nil {
			return err
		}

	case "no_explicit_access":
		if err := json.Unmarshal(w.NoExplicitAccess, &u.NoExplicitAccess); err != nil {
			return err
		}

	}
	return nil
}

// UpdateFolderPolicyArg : If any of the policy's are unset, then they retain
// their current setting.
type UpdateFolderPolicyArg struct {
	// SharedFolderId : The ID for the shared folder.
	SharedFolderId string `json:"shared_folder_id"`
	// MemberPolicy : Who can be a member of this shared folder. Only applicable
	// if the current user is on a team.
	MemberPolicy *MemberPolicy `json:"member_policy,omitempty"`
	// AclUpdatePolicy : Who can add and remove members of this shared folder.
	AclUpdatePolicy *AclUpdatePolicy `json:"acl_update_policy,omitempty"`
	// SharedLinkPolicy : The policy to apply to shared links created for
	// content inside this shared folder. The current user must be on a team to
	// set this policy to `SharedLinkPolicy.members`.
	SharedLinkPolicy *SharedLinkPolicy `json:"shared_link_policy,omitempty"`
}

// NewUpdateFolderPolicyArg returns a new UpdateFolderPolicyArg instance
func NewUpdateFolderPolicyArg(SharedFolderId string) *UpdateFolderPolicyArg {
	s := new(UpdateFolderPolicyArg)
	s.SharedFolderId = SharedFolderId
	return s
}

// UpdateFolderPolicyError : has no documentation (yet)
type UpdateFolderPolicyError struct {
	dropbox.Tagged
	// AccessError : has no documentation (yet)
	AccessError *SharedFolderAccessError `json:"access_error,omitempty"`
}

// Valid tag values for UpdateFolderPolicyError
const (
	UpdateFolderPolicyErrorAccessError                     = "access_error"
	UpdateFolderPolicyErrorNotOnTeam                       = "not_on_team"
	UpdateFolderPolicyErrorTeamPolicyDisallowsMemberPolicy = "team_policy_disallows_member_policy"
	UpdateFolderPolicyErrorDisallowedSharedLinkPolicy      = "disallowed_shared_link_policy"
	UpdateFolderPolicyErrorNoPermission                    = "no_permission"
	UpdateFolderPolicyErrorOther                           = "other"
)

// UnmarshalJSON deserializes into a UpdateFolderPolicyError instance
func (u *UpdateFolderPolicyError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// AccessError : has no documentation (yet)
		AccessError json.RawMessage `json:"access_error,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "access_error":
		if err := json.Unmarshal(w.AccessError, &u.AccessError); err != nil {
			return err
		}

	}
	return nil
}

// UserInfo : Basic information about a user. Use `usersAccount` and
// `usersAccountBatch` to obtain more detailed information.
type UserInfo struct {
	// AccountId : The account ID of the user.
	AccountId string `json:"account_id"`
	// SameTeam : If the user is in the same team as current user.
	SameTeam bool `json:"same_team"`
	// TeamMemberId : The team member ID of the shared folder member. Only
	// present if `same_team` is true.
	TeamMemberId string `json:"team_member_id,omitempty"`
}

// NewUserInfo returns a new UserInfo instance
func NewUserInfo(AccountId string, SameTeam bool) *UserInfo {
	s := new(UserInfo)
	s.AccountId = AccountId
	s.SameTeam = SameTeam
	return s
}

// UserMembershipInfo : The information about a user member of the shared
// content.
type UserMembershipInfo struct {
	MembershipInfo
	// User : The account information for the membership user.
	User *UserInfo `json:"user"`
}

// NewUserMembershipInfo returns a new UserMembershipInfo instance
func NewUserMembershipInfo(AccessType *AccessLevel, User *UserInfo) *UserMembershipInfo {
	s := new(UserMembershipInfo)
	s.AccessType = AccessType
	s.User = User
	s.IsInherited = false
	return s
}

// Visibility : Who can access a shared link. The most open visibility is
// `public`. The default depends on many aspects, such as team and user
// preferences and shared folder settings.
type Visibility struct {
	dropbox.Tagged
}

// Valid tag values for Visibility
const (
	VisibilityPublic           = "public"
	VisibilityTeamOnly         = "team_only"
	VisibilityPassword         = "password"
	VisibilityTeamAndPassword  = "team_and_password"
	VisibilitySharedFolderOnly = "shared_folder_only"
	VisibilityOther            = "other"
)
