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
	"encoding/json"
	"io"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/async"
	"github.com/dropbox/dropbox-sdk-go-unofficial/properties"
)

type PropertiesError struct {
	Tag  string       `json:".tag"`
	Path *LookupError `json:"path,omitempty"`
}

func (u *PropertiesError) UnmarshalJSON(body []byte) error {
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

type InvalidPropertyGroupError struct {
	Tag string `json:".tag"`
}

type AddPropertiesError struct {
	Tag string `json:".tag"`
}

type GetMetadataArg struct {
	// The path of a file or folder on Dropbox.
	Path string `json:"path"`
	// If true, `FileMetadata.media_info` is set for photo and video.
	IncludeMediaInfo bool `json:"include_media_info"`
	// If true, `DeletedMetadata` will be returned for deleted file or folder,
	// otherwise `LookupError.not_found` will be returned.
	IncludeDeleted bool `json:"include_deleted"`
	// If true, the results will include a flag for each file indicating whether or
	// not  that file has any explicit members.
	IncludeHasExplicitSharedMembers bool `json:"include_has_explicit_shared_members"`
}

func NewGetMetadataArg(Path string) *GetMetadataArg {
	s := new(GetMetadataArg)
	s.Path = Path
	s.IncludeMediaInfo = false
	s.IncludeDeleted = false
	s.IncludeHasExplicitSharedMembers = false
	return s
}

type AlphaGetMetadataArg struct {
	// The path of a file or folder on Dropbox.
	Path string `json:"path"`
	// If true, `FileMetadata.media_info` is set for photo and video.
	IncludeMediaInfo bool `json:"include_media_info"`
	// If true, `DeletedMetadata` will be returned for deleted file or folder,
	// otherwise `LookupError.not_found` will be returned.
	IncludeDeleted bool `json:"include_deleted"`
	// If true, the results will include a flag for each file indicating whether or
	// not  that file has any explicit members.
	IncludeHasExplicitSharedMembers bool `json:"include_has_explicit_shared_members"`
	// If true, `FileMetadata.property_groups` is set for files with custom
	// properties.
	IncludePropertyTemplates []string `json:"include_property_templates,omitempty"`
}

func NewAlphaGetMetadataArg(Path string) *AlphaGetMetadataArg {
	s := new(AlphaGetMetadataArg)
	s.Path = Path
	s.IncludeMediaInfo = false
	s.IncludeDeleted = false
	s.IncludeHasExplicitSharedMembers = false
	return s
}

type GetMetadataError struct {
	Tag  string       `json:".tag"`
	Path *LookupError `json:"path,omitempty"`
}

func (u *GetMetadataError) UnmarshalJSON(body []byte) error {
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

type AlphaGetMetadataError struct {
	Tag             string                 `json:".tag"`
	PropertiesError *LookUpPropertiesError `json:"properties_error,omitempty"`
}

func (u *AlphaGetMetadataError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag             string          `json:".tag"`
		PropertiesError json.RawMessage `json:"properties_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "properties_error":
		{
			if len(w.PropertiesError) == 0 {
				break
			}
			if err := json.Unmarshal(w.PropertiesError, &u.PropertiesError); err != nil {
				return err
			}
		}
	}
	return nil
}

type CommitInfo struct {
	// Path in the user's Dropbox to save the file.
	Path string `json:"path"`
	// Selects what to do if the file already exists.
	Mode *WriteMode `json:"mode"`
	// If there's a conflict, as determined by `mode`, have the Dropbox server try
	// to autorename the file to avoid conflict.
	Autorename bool `json:"autorename"`
	// The value to store as the `client_modified` timestamp. Dropbox automatically
	// records the time at which the file was written to the Dropbox servers. It
	// can also record an additional timestamp, provided by Dropbox desktop
	// clients, mobile clients, and API apps of when the file was actually created
	// or modified.
	ClientModified time.Time `json:"client_modified,omitempty"`
	// Normally, users are made aware of any file modifications in their Dropbox
	// account via notifications in the client software. If `True`, this tells the
	// clients that this modification shouldn't result in a user notification.
	Mute bool `json:"mute"`
}

func NewCommitInfo(Path string) *CommitInfo {
	s := new(CommitInfo)
	s.Path = Path
	s.Mode = &WriteMode{Tag: "add"}
	s.Autorename = false
	s.Mute = false
	return s
}

type CommitInfoWithProperties struct {
	// Path in the user's Dropbox to save the file.
	Path string `json:"path"`
	// Selects what to do if the file already exists.
	Mode *WriteMode `json:"mode"`
	// If there's a conflict, as determined by `mode`, have the Dropbox server try
	// to autorename the file to avoid conflict.
	Autorename bool `json:"autorename"`
	// The value to store as the `client_modified` timestamp. Dropbox automatically
	// records the time at which the file was written to the Dropbox servers. It
	// can also record an additional timestamp, provided by Dropbox desktop
	// clients, mobile clients, and API apps of when the file was actually created
	// or modified.
	ClientModified time.Time `json:"client_modified,omitempty"`
	// Normally, users are made aware of any file modifications in their Dropbox
	// account via notifications in the client software. If `True`, this tells the
	// clients that this modification shouldn't result in a user notification.
	Mute bool `json:"mute"`
	// List of custom properties to add to file.
	PropertyGroups []*properties.PropertyGroup `json:"property_groups,omitempty"`
}

func NewCommitInfoWithProperties(Path string) *CommitInfoWithProperties {
	s := new(CommitInfoWithProperties)
	s.Path = Path
	s.Mode = &WriteMode{Tag: "add"}
	s.Autorename = false
	s.Mute = false
	return s
}

type CreateFolderArg struct {
	// Path in the user's Dropbox to create.
	Path string `json:"path"`
}

func NewCreateFolderArg(Path string) *CreateFolderArg {
	s := new(CreateFolderArg)
	s.Path = Path
	return s
}

type CreateFolderError struct {
	Tag  string      `json:".tag"`
	Path *WriteError `json:"path,omitempty"`
}

func (u *CreateFolderError) UnmarshalJSON(body []byte) error {
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

type DeleteArg struct {
	// Path in the user's Dropbox to delete.
	Path string `json:"path"`
}

func NewDeleteArg(Path string) *DeleteArg {
	s := new(DeleteArg)
	s.Path = Path
	return s
}

type DeleteError struct {
	Tag        string       `json:".tag"`
	PathLookup *LookupError `json:"path_lookup,omitempty"`
	PathWrite  *WriteError  `json:"path_write,omitempty"`
}

func (u *DeleteError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag        string          `json:".tag"`
		PathLookup json.RawMessage `json:"path_lookup"`
		PathWrite  json.RawMessage `json:"path_write"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "path_lookup":
		{
			if len(w.PathLookup) == 0 {
				break
			}
			if err := json.Unmarshal(w.PathLookup, &u.PathLookup); err != nil {
				return err
			}
		}
	case "path_write":
		{
			if len(w.PathWrite) == 0 {
				break
			}
			if err := json.Unmarshal(w.PathWrite, &u.PathWrite); err != nil {
				return err
			}
		}
	}
	return nil
}

// Metadata for a file or folder.
type Metadata struct {
	Tag     string           `json:".tag"`
	File    *FileMetadata    `json:"file,omitempty"`
	Folder  *FolderMetadata  `json:"folder,omitempty"`
	Deleted *DeletedMetadata `json:"deleted,omitempty"`
}

func (u *Metadata) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag     string          `json:".tag"`
		File    json.RawMessage `json:"file"`
		Folder  json.RawMessage `json:"folder"`
		Deleted json.RawMessage `json:"deleted"`
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
	case "deleted":
		{
			if err := json.Unmarshal(body, &u.Deleted); err != nil {
				return err
			}
		}
	}
	return nil
}

// Indicates that there used to be a file or folder at this path, but it no
// longer exists.
type DeletedMetadata struct {
	// The last component of the path (including extension). This never contains a
	// slash.
	Name string `json:"name"`
	// The lowercased full path in the user's Dropbox. This always starts with a
	// slash. This field will be null if the file or folder is not mounted.
	PathLower string `json:"path_lower,omitempty"`
	// The cased path to be used for display purposes only. In rare instances the
	// casing will not correctly match the user's filesystem, but this behavior
	// will match the path provided in the Core API v1. Changes to the casing of
	// paths won't be returned by `ListFolderContinue`. This field will be null if
	// the file or folder is not mounted.
	PathDisplay string `json:"path_display,omitempty"`
	// Deprecated. Please use `FileSharingInfo.parent_shared_folder_id` or
	// `FolderSharingInfo.parent_shared_folder_id` instead.
	ParentSharedFolderId string `json:"parent_shared_folder_id,omitempty"`
}

func NewDeletedMetadata(Name string) *DeletedMetadata {
	s := new(DeletedMetadata)
	s.Name = Name
	return s
}

// Dimensions for a photo or video.
type Dimensions struct {
	// Height of the photo/video.
	Height uint64 `json:"height"`
	// Width of the photo/video.
	Width uint64 `json:"width"`
}

func NewDimensions(Height uint64, Width uint64) *Dimensions {
	s := new(Dimensions)
	s.Height = Height
	s.Width = Width
	return s
}

type DownloadArg struct {
	// The path of the file to download.
	Path string `json:"path"`
	// Deprecated. Please specify revision in `path` instead
	Rev string `json:"rev,omitempty"`
}

func NewDownloadArg(Path string) *DownloadArg {
	s := new(DownloadArg)
	s.Path = Path
	return s
}

type DownloadError struct {
	Tag  string       `json:".tag"`
	Path *LookupError `json:"path,omitempty"`
}

func (u *DownloadError) UnmarshalJSON(body []byte) error {
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

type FileMetadata struct {
	// The last component of the path (including extension). This never contains a
	// slash.
	Name string `json:"name"`
	// A unique identifier for the file.
	Id string `json:"id"`
	// For files, this is the modification time set by the desktop client when the
	// file was added to Dropbox. Since this time is not verified (the Dropbox
	// server stores whatever the desktop client sends up), this should only be
	// used for display purposes (such as sorting) and not, for example, to
	// determine if a file has changed or not.
	ClientModified time.Time `json:"client_modified"`
	// The last time the file was modified on Dropbox.
	ServerModified time.Time `json:"server_modified"`
	// A unique identifier for the current revision of a file. This field is the
	// same rev as elsewhere in the API and can be used to detect changes and avoid
	// conflicts.
	Rev string `json:"rev"`
	// The file size in bytes.
	Size uint64 `json:"size"`
	// The lowercased full path in the user's Dropbox. This always starts with a
	// slash. This field will be null if the file or folder is not mounted.
	PathLower string `json:"path_lower,omitempty"`
	// The cased path to be used for display purposes only. In rare instances the
	// casing will not correctly match the user's filesystem, but this behavior
	// will match the path provided in the Core API v1. Changes to the casing of
	// paths won't be returned by `ListFolderContinue`. This field will be null if
	// the file or folder is not mounted.
	PathDisplay string `json:"path_display,omitempty"`
	// Deprecated. Please use `FileSharingInfo.parent_shared_folder_id` or
	// `FolderSharingInfo.parent_shared_folder_id` instead.
	ParentSharedFolderId string `json:"parent_shared_folder_id,omitempty"`
	// Additional information if the file is a photo or video.
	MediaInfo *MediaInfo `json:"media_info,omitempty"`
	// Set if this file is contained in a shared folder.
	SharingInfo *FileSharingInfo `json:"sharing_info,omitempty"`
	// Additional information if the file has custom properties with the property
	// template specified.
	PropertyGroups []*properties.PropertyGroup `json:"property_groups,omitempty"`
	// This flag will only be present if include_has_explicit_shared_members  is
	// true in `ListFolder` or `GetMetadata`. If this  flag is present, it will be
	// true if this file has any explicit shared  members. This is different from
	// sharing_info in that this could be true  in the case where a file has
	// explicit members but is not contained within  a shared folder.
	HasExplicitSharedMembers bool `json:"has_explicit_shared_members,omitempty"`
}

func NewFileMetadata(Name string, Id string, ClientModified time.Time, ServerModified time.Time, Rev string, Size uint64) *FileMetadata {
	s := new(FileMetadata)
	s.Name = Name
	s.Id = Id
	s.ClientModified = ClientModified
	s.ServerModified = ServerModified
	s.Rev = Rev
	s.Size = Size
	return s
}

// Sharing info for a file or folder.
type SharingInfo struct {
	// True if the file or folder is inside a read-only shared folder.
	ReadOnly bool `json:"read_only"`
}

func NewSharingInfo(ReadOnly bool) *SharingInfo {
	s := new(SharingInfo)
	s.ReadOnly = ReadOnly
	return s
}

// Sharing info for a file which is contained by a shared folder.
type FileSharingInfo struct {
	// True if the file or folder is inside a read-only shared folder.
	ReadOnly bool `json:"read_only"`
	// ID of shared folder that holds this file.
	ParentSharedFolderId string `json:"parent_shared_folder_id"`
	// The last user who modified the file. This field will be null if the user's
	// account has been deleted.
	ModifiedBy string `json:"modified_by,omitempty"`
}

func NewFileSharingInfo(ReadOnly bool, ParentSharedFolderId string) *FileSharingInfo {
	s := new(FileSharingInfo)
	s.ReadOnly = ReadOnly
	s.ParentSharedFolderId = ParentSharedFolderId
	return s
}

type FolderMetadata struct {
	// The last component of the path (including extension). This never contains a
	// slash.
	Name string `json:"name"`
	// A unique identifier for the folder.
	Id string `json:"id"`
	// The lowercased full path in the user's Dropbox. This always starts with a
	// slash. This field will be null if the file or folder is not mounted.
	PathLower string `json:"path_lower,omitempty"`
	// The cased path to be used for display purposes only. In rare instances the
	// casing will not correctly match the user's filesystem, but this behavior
	// will match the path provided in the Core API v1. Changes to the casing of
	// paths won't be returned by `ListFolderContinue`. This field will be null if
	// the file or folder is not mounted.
	PathDisplay string `json:"path_display,omitempty"`
	// Deprecated. Please use `FileSharingInfo.parent_shared_folder_id` or
	// `FolderSharingInfo.parent_shared_folder_id` instead.
	ParentSharedFolderId string `json:"parent_shared_folder_id,omitempty"`
	// Deprecated. Please use `sharing_info` instead.
	SharedFolderId string `json:"shared_folder_id,omitempty"`
	// Set if the folder is contained in a shared folder or is a shared folder
	// mount point.
	SharingInfo *FolderSharingInfo `json:"sharing_info,omitempty"`
	// Additional information if the file has custom properties with the property
	// template specified.
	PropertyGroups []*properties.PropertyGroup `json:"property_groups,omitempty"`
}

func NewFolderMetadata(Name string, Id string) *FolderMetadata {
	s := new(FolderMetadata)
	s.Name = Name
	s.Id = Id
	return s
}

// Sharing info for a folder which is contained in a shared folder or is a
// shared folder mount point.
type FolderSharingInfo struct {
	// True if the file or folder is inside a read-only shared folder.
	ReadOnly bool `json:"read_only"`
	// Set if the folder is contained by a shared folder.
	ParentSharedFolderId string `json:"parent_shared_folder_id,omitempty"`
	// If this folder is a shared folder mount point, the ID of the shared folder
	// mounted at this location.
	SharedFolderId string `json:"shared_folder_id,omitempty"`
}

func NewFolderSharingInfo(ReadOnly bool) *FolderSharingInfo {
	s := new(FolderSharingInfo)
	s.ReadOnly = ReadOnly
	return s
}

type GetCopyReferenceArg struct {
	// The path to the file or folder you want to get a copy reference to.
	Path string `json:"path"`
}

func NewGetCopyReferenceArg(Path string) *GetCopyReferenceArg {
	s := new(GetCopyReferenceArg)
	s.Path = Path
	return s
}

type GetCopyReferenceError struct {
	Tag  string       `json:".tag"`
	Path *LookupError `json:"path,omitempty"`
}

func (u *GetCopyReferenceError) UnmarshalJSON(body []byte) error {
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

type GetCopyReferenceResult struct {
	// Metadata of the file or folder.
	Metadata *Metadata `json:"metadata"`
	// A copy reference to the file or folder.
	CopyReference string `json:"copy_reference"`
	// The expiration date of the copy reference. This value is currently set to be
	// far enough in the future so that expiration is effectively not an issue.
	Expires time.Time `json:"expires"`
}

func NewGetCopyReferenceResult(Metadata *Metadata, CopyReference string, Expires time.Time) *GetCopyReferenceResult {
	s := new(GetCopyReferenceResult)
	s.Metadata = Metadata
	s.CopyReference = CopyReference
	s.Expires = Expires
	return s
}

type GetTemporaryLinkArg struct {
	// The path to the file you want a temporary link to.
	Path string `json:"path"`
}

func NewGetTemporaryLinkArg(Path string) *GetTemporaryLinkArg {
	s := new(GetTemporaryLinkArg)
	s.Path = Path
	return s
}

type GetTemporaryLinkError struct {
	Tag  string       `json:".tag"`
	Path *LookupError `json:"path,omitempty"`
}

func (u *GetTemporaryLinkError) UnmarshalJSON(body []byte) error {
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

type GetTemporaryLinkResult struct {
	// Metadata of the file.
	Metadata *FileMetadata `json:"metadata"`
	// The temporary link which can be used to stream content the file.
	Link string `json:"link"`
}

func NewGetTemporaryLinkResult(Metadata *FileMetadata, Link string) *GetTemporaryLinkResult {
	s := new(GetTemporaryLinkResult)
	s.Metadata = Metadata
	s.Link = Link
	return s
}

// GPS coordinates for a photo or video.
type GpsCoordinates struct {
	// Latitude of the GPS coordinates.
	Latitude float64 `json:"latitude"`
	// Longitude of the GPS coordinates.
	Longitude float64 `json:"longitude"`
}

func NewGpsCoordinates(Latitude float64, Longitude float64) *GpsCoordinates {
	s := new(GpsCoordinates)
	s.Latitude = Latitude
	s.Longitude = Longitude
	return s
}

type ListFolderArg struct {
	// The path to the folder you want to see the contents of.
	Path string `json:"path"`
	// If true, the list folder operation will be applied recursively to all
	// subfolders and the response will contain contents of all subfolders.
	Recursive bool `json:"recursive"`
	// If true, `FileMetadata.media_info` is set for photo and video.
	IncludeMediaInfo bool `json:"include_media_info"`
	// If true, the results will include entries for files and folders that used to
	// exist but were deleted.
	IncludeDeleted bool `json:"include_deleted"`
	// If true, the results will include a flag for each file indicating whether or
	// not  that file has any explicit members.
	IncludeHasExplicitSharedMembers bool `json:"include_has_explicit_shared_members"`
}

func NewListFolderArg(Path string) *ListFolderArg {
	s := new(ListFolderArg)
	s.Path = Path
	s.Recursive = false
	s.IncludeMediaInfo = false
	s.IncludeDeleted = false
	s.IncludeHasExplicitSharedMembers = false
	return s
}

type ListFolderContinueArg struct {
	// The cursor returned by your last call to `ListFolder` or
	// `ListFolderContinue`.
	Cursor string `json:"cursor"`
}

func NewListFolderContinueArg(Cursor string) *ListFolderContinueArg {
	s := new(ListFolderContinueArg)
	s.Cursor = Cursor
	return s
}

type ListFolderContinueError struct {
	Tag  string       `json:".tag"`
	Path *LookupError `json:"path,omitempty"`
}

func (u *ListFolderContinueError) UnmarshalJSON(body []byte) error {
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

type ListFolderError struct {
	Tag  string       `json:".tag"`
	Path *LookupError `json:"path,omitempty"`
}

func (u *ListFolderError) UnmarshalJSON(body []byte) error {
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

type ListFolderGetLatestCursorResult struct {
	// Pass the cursor into `ListFolderContinue` to see what's changed in the
	// folder since your previous query.
	Cursor string `json:"cursor"`
}

func NewListFolderGetLatestCursorResult(Cursor string) *ListFolderGetLatestCursorResult {
	s := new(ListFolderGetLatestCursorResult)
	s.Cursor = Cursor
	return s
}

type ListFolderLongpollArg struct {
	// A cursor as returned by `ListFolder` or `ListFolderContinue`. Cursors
	// retrieved by setting `ListFolderArg.include_media_info` to `True` are not
	// supported.
	Cursor string `json:"cursor"`
	// A timeout in seconds. The request will block for at most this length of
	// time, plus up to 90 seconds of random jitter added to avoid the thundering
	// herd problem. Care should be taken when using this parameter, as some
	// network infrastructure does not support long timeouts.
	Timeout uint64 `json:"timeout"`
}

func NewListFolderLongpollArg(Cursor string) *ListFolderLongpollArg {
	s := new(ListFolderLongpollArg)
	s.Cursor = Cursor
	s.Timeout = 30
	return s
}

type ListFolderLongpollError struct {
	Tag string `json:".tag"`
}

type ListFolderLongpollResult struct {
	// Indicates whether new changes are available. If true, call
	// `ListFolderContinue` to retrieve the changes.
	Changes bool `json:"changes"`
	// If present, backoff for at least this many seconds before calling
	// `ListFolderLongpoll` again.
	Backoff uint64 `json:"backoff,omitempty"`
}

func NewListFolderLongpollResult(Changes bool) *ListFolderLongpollResult {
	s := new(ListFolderLongpollResult)
	s.Changes = Changes
	return s
}

type ListFolderResult struct {
	// The files and (direct) subfolders in the folder.
	Entries []*Metadata `json:"entries"`
	// Pass the cursor into `ListFolderContinue` to see what's changed in the
	// folder since your previous query.
	Cursor string `json:"cursor"`
	// If true, then there are more entries available. Pass the cursor to
	// `ListFolderContinue` to retrieve the rest.
	HasMore bool `json:"has_more"`
}

func NewListFolderResult(Entries []*Metadata, Cursor string, HasMore bool) *ListFolderResult {
	s := new(ListFolderResult)
	s.Entries = Entries
	s.Cursor = Cursor
	s.HasMore = HasMore
	return s
}

type ListRevisionsArg struct {
	// The path to the file you want to see the revisions of.
	Path string `json:"path"`
	// The maximum number of revision entries returned.
	Limit uint64 `json:"limit"`
}

func NewListRevisionsArg(Path string) *ListRevisionsArg {
	s := new(ListRevisionsArg)
	s.Path = Path
	s.Limit = 10
	return s
}

type ListRevisionsError struct {
	Tag  string       `json:".tag"`
	Path *LookupError `json:"path,omitempty"`
}

func (u *ListRevisionsError) UnmarshalJSON(body []byte) error {
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

type ListRevisionsResult struct {
	// If the file is deleted.
	IsDeleted bool `json:"is_deleted"`
	// The revisions for the file. Only non-delete revisions will show up here.
	Entries []*FileMetadata `json:"entries"`
}

func NewListRevisionsResult(IsDeleted bool, Entries []*FileMetadata) *ListRevisionsResult {
	s := new(ListRevisionsResult)
	s.IsDeleted = IsDeleted
	s.Entries = Entries
	return s
}

type LookUpPropertiesError struct {
	Tag string `json:".tag"`
}

type LookupError struct {
	Tag           string `json:".tag"`
	MalformedPath string `json:"malformed_path,omitempty"`
}

func (u *LookupError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag           string          `json:".tag"`
		MalformedPath json.RawMessage `json:"malformed_path,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "malformed_path":
		{
			if len(w.MalformedPath) == 0 {
				break
			}
			if err := json.Unmarshal(w.MalformedPath, &u.MalformedPath); err != nil {
				return err
			}
		}
	}
	return nil
}

type MediaInfo struct {
	Tag string `json:".tag"`
	// The metadata for the photo/video.
	Metadata *MediaMetadata `json:"metadata,omitempty"`
}

func (u *MediaInfo) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// The metadata for the photo/video.
		Metadata json.RawMessage `json:"metadata"`
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
	}
	return nil
}

// Metadata for a photo or video.
type MediaMetadata struct {
	Tag   string         `json:".tag"`
	Photo *PhotoMetadata `json:"photo,omitempty"`
	Video *VideoMetadata `json:"video,omitempty"`
}

func (u *MediaMetadata) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag   string          `json:".tag"`
		Photo json.RawMessage `json:"photo"`
		Video json.RawMessage `json:"video"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "photo":
		{
			if err := json.Unmarshal(body, &u.Photo); err != nil {
				return err
			}
		}
	case "video":
		{
			if err := json.Unmarshal(body, &u.Video); err != nil {
				return err
			}
		}
	}
	return nil
}

// Metadata for a photo.
type PhotoMetadata struct {
	// Dimension of the photo/video.
	Dimensions *Dimensions `json:"dimensions,omitempty"`
	// The GPS coordinate of the photo/video.
	Location *GpsCoordinates `json:"location,omitempty"`
	// The timestamp when the photo/video is taken.
	TimeTaken time.Time `json:"time_taken,omitempty"`
}

func NewPhotoMetadata() *PhotoMetadata {
	s := new(PhotoMetadata)
	return s
}

type PreviewArg struct {
	// The path of the file to preview.
	Path string `json:"path"`
	// Deprecated. Please specify revision in `path` instead
	Rev string `json:"rev,omitempty"`
}

func NewPreviewArg(Path string) *PreviewArg {
	s := new(PreviewArg)
	s.Path = Path
	return s
}

type PreviewError struct {
	Tag string `json:".tag"`
	// An error occurs when downloading metadata for the file.
	Path *LookupError `json:"path,omitempty"`
}

func (u *PreviewError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// An error occurs when downloading metadata for the file.
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

type PropertyGroupUpdate struct {
	// A unique identifier for a property template.
	TemplateId string `json:"template_id"`
	// List of property fields to update if the field already exists. If the field
	// doesn't exist, add the field to the property group.
	AddOrUpdateFields []*properties.PropertyField `json:"add_or_update_fields,omitempty"`
	// List of property field names to remove from property group if the field
	// exists.
	RemoveFields []string `json:"remove_fields,omitempty"`
}

func NewPropertyGroupUpdate(TemplateId string) *PropertyGroupUpdate {
	s := new(PropertyGroupUpdate)
	s.TemplateId = TemplateId
	return s
}

type PropertyGroupWithPath struct {
	// A unique identifier for the file.
	Path string `json:"path"`
	// Filled custom property templates associated with a file.
	PropertyGroups []*properties.PropertyGroup `json:"property_groups"`
}

func NewPropertyGroupWithPath(Path string, PropertyGroups []*properties.PropertyGroup) *PropertyGroupWithPath {
	s := new(PropertyGroupWithPath)
	s.Path = Path
	s.PropertyGroups = PropertyGroups
	return s
}

type RelocationArg struct {
	// Path in the user's Dropbox to be copied or moved.
	FromPath string `json:"from_path"`
	// Path in the user's Dropbox that is the destination.
	ToPath string `json:"to_path"`
}

func NewRelocationArg(FromPath string, ToPath string) *RelocationArg {
	s := new(RelocationArg)
	s.FromPath = FromPath
	s.ToPath = ToPath
	return s
}

type RelocationError struct {
	Tag        string       `json:".tag"`
	FromLookup *LookupError `json:"from_lookup,omitempty"`
	FromWrite  *WriteError  `json:"from_write,omitempty"`
	To         *WriteError  `json:"to,omitempty"`
}

func (u *RelocationError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag        string          `json:".tag"`
		FromLookup json.RawMessage `json:"from_lookup"`
		FromWrite  json.RawMessage `json:"from_write"`
		To         json.RawMessage `json:"to"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "from_lookup":
		{
			if len(w.FromLookup) == 0 {
				break
			}
			if err := json.Unmarshal(w.FromLookup, &u.FromLookup); err != nil {
				return err
			}
		}
	case "from_write":
		{
			if len(w.FromWrite) == 0 {
				break
			}
			if err := json.Unmarshal(w.FromWrite, &u.FromWrite); err != nil {
				return err
			}
		}
	case "to":
		{
			if len(w.To) == 0 {
				break
			}
			if err := json.Unmarshal(w.To, &u.To); err != nil {
				return err
			}
		}
	}
	return nil
}

type RemovePropertiesArg struct {
	// A unique identifier for the file.
	Path string `json:"path"`
	// A list of identifiers for a property template created by route
	// properties/template/add.
	PropertyTemplateIds []string `json:"property_template_ids"`
}

func NewRemovePropertiesArg(Path string, PropertyTemplateIds []string) *RemovePropertiesArg {
	s := new(RemovePropertiesArg)
	s.Path = Path
	s.PropertyTemplateIds = PropertyTemplateIds
	return s
}

type RemovePropertiesError struct {
	Tag                 string                 `json:".tag"`
	PropertyGroupLookup *LookUpPropertiesError `json:"property_group_lookup,omitempty"`
}

func (u *RemovePropertiesError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag                 string          `json:".tag"`
		PropertyGroupLookup json.RawMessage `json:"property_group_lookup"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "property_group_lookup":
		{
			if len(w.PropertyGroupLookup) == 0 {
				break
			}
			if err := json.Unmarshal(w.PropertyGroupLookup, &u.PropertyGroupLookup); err != nil {
				return err
			}
		}
	}
	return nil
}

type RestoreArg struct {
	// The path to the file you want to restore.
	Path string `json:"path"`
	// The revision to restore for the file.
	Rev string `json:"rev"`
}

func NewRestoreArg(Path string, Rev string) *RestoreArg {
	s := new(RestoreArg)
	s.Path = Path
	s.Rev = Rev
	return s
}

type RestoreError struct {
	Tag string `json:".tag"`
	// An error occurs when downloading metadata for the file.
	PathLookup *LookupError `json:"path_lookup,omitempty"`
	// An error occurs when trying to restore the file to that path.
	PathWrite *WriteError `json:"path_write,omitempty"`
}

func (u *RestoreError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// An error occurs when downloading metadata for the file.
		PathLookup json.RawMessage `json:"path_lookup"`
		// An error occurs when trying to restore the file to that path.
		PathWrite json.RawMessage `json:"path_write"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "path_lookup":
		{
			if len(w.PathLookup) == 0 {
				break
			}
			if err := json.Unmarshal(w.PathLookup, &u.PathLookup); err != nil {
				return err
			}
		}
	case "path_write":
		{
			if len(w.PathWrite) == 0 {
				break
			}
			if err := json.Unmarshal(w.PathWrite, &u.PathWrite); err != nil {
				return err
			}
		}
	}
	return nil
}

type SaveCopyReferenceArg struct {
	// A copy reference returned by `CopyReferenceGet`.
	CopyReference string `json:"copy_reference"`
	// Path in the user's Dropbox that is the destination.
	Path string `json:"path"`
}

func NewSaveCopyReferenceArg(CopyReference string, Path string) *SaveCopyReferenceArg {
	s := new(SaveCopyReferenceArg)
	s.CopyReference = CopyReference
	s.Path = Path
	return s
}

type SaveCopyReferenceError struct {
	Tag  string      `json:".tag"`
	Path *WriteError `json:"path,omitempty"`
}

func (u *SaveCopyReferenceError) UnmarshalJSON(body []byte) error {
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

type SaveCopyReferenceResult struct {
	// The metadata of the saved file or folder in the user's Dropbox.
	Metadata *Metadata `json:"metadata"`
}

func NewSaveCopyReferenceResult(Metadata *Metadata) *SaveCopyReferenceResult {
	s := new(SaveCopyReferenceResult)
	s.Metadata = Metadata
	return s
}

type SaveUrlArg struct {
	// The path in Dropbox where the URL will be saved to.
	Path string `json:"path"`
	// The URL to be saved.
	Url string `json:"url"`
}

func NewSaveUrlArg(Path string, Url string) *SaveUrlArg {
	s := new(SaveUrlArg)
	s.Path = Path
	s.Url = Url
	return s
}

type SaveUrlError struct {
	Tag  string      `json:".tag"`
	Path *WriteError `json:"path,omitempty"`
}

func (u *SaveUrlError) UnmarshalJSON(body []byte) error {
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

type SaveUrlJobStatus struct {
	Tag string `json:".tag"`
	// Metadata of the file where the URL is saved to.
	Complete *FileMetadata `json:"complete,omitempty"`
	Failed   *SaveUrlError `json:"failed,omitempty"`
}

func (u *SaveUrlJobStatus) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// Metadata of the file where the URL is saved to.
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

type SaveUrlResult struct {
	Tag string `json:".tag"`
	// Metadata of the file where the URL is saved to.
	Complete *FileMetadata `json:"complete,omitempty"`
}

func (u *SaveUrlResult) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// Metadata of the file where the URL is saved to.
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

type SearchArg struct {
	// The path in the user's Dropbox to search. Should probably be a folder.
	Path string `json:"path"`
	// The string to search for. The search string is split on spaces into multiple
	// tokens. For file name searching, the last token is used for prefix matching
	// (i.e. "bat c" matches "bat cave" but not "batman car").
	Query string `json:"query"`
	// The starting index within the search results (used for paging).
	Start uint64 `json:"start"`
	// The maximum number of search results to return.
	MaxResults uint64 `json:"max_results"`
	// The search mode (filename, filename_and_content, or deleted_filename). Note
	// that searching file content is only available for Dropbox Business accounts.
	Mode *SearchMode `json:"mode"`
}

func NewSearchArg(Path string, Query string) *SearchArg {
	s := new(SearchArg)
	s.Path = Path
	s.Query = Query
	s.Start = 0
	s.MaxResults = 100
	s.Mode = &SearchMode{Tag: "filename"}
	return s
}

type SearchError struct {
	Tag  string       `json:".tag"`
	Path *LookupError `json:"path,omitempty"`
}

func (u *SearchError) UnmarshalJSON(body []byte) error {
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

type SearchMatch struct {
	// The type of the match.
	MatchType *SearchMatchType `json:"match_type"`
	// The metadata for the matched file or folder.
	Metadata *Metadata `json:"metadata"`
}

func NewSearchMatch(MatchType *SearchMatchType, Metadata *Metadata) *SearchMatch {
	s := new(SearchMatch)
	s.MatchType = MatchType
	s.Metadata = Metadata
	return s
}

// Indicates what type of match was found for a given item.
type SearchMatchType struct {
	Tag string `json:".tag"`
}

type SearchMode struct {
	Tag string `json:".tag"`
}

type SearchResult struct {
	// A list (possibly empty) of matches for the query.
	Matches []*SearchMatch `json:"matches"`
	// Used for paging. If true, indicates there is another page of results
	// available that can be fetched by calling `Search` again.
	More bool `json:"more"`
	// Used for paging. Value to set the start argument to when calling `Search` to
	// fetch the next page of results.
	Start uint64 `json:"start"`
}

func NewSearchResult(Matches []*SearchMatch, More bool, Start uint64) *SearchResult {
	s := new(SearchResult)
	s.Matches = Matches
	s.More = More
	s.Start = Start
	return s
}

type ThumbnailArg struct {
	// The path to the image file you want to thumbnail.
	Path string `json:"path"`
	// The format for the thumbnail image, jpeg (default) or png. For  images that
	// are photos, jpeg should be preferred, while png is  better for screenshots
	// and digital arts.
	Format *ThumbnailFormat `json:"format"`
	// The size for the thumbnail image.
	Size *ThumbnailSize `json:"size"`
}

func NewThumbnailArg(Path string) *ThumbnailArg {
	s := new(ThumbnailArg)
	s.Path = Path
	s.Format = &ThumbnailFormat{Tag: "jpeg"}
	s.Size = &ThumbnailSize{Tag: "w64h64"}
	return s
}

type ThumbnailError struct {
	Tag string `json:".tag"`
	// An error occurs when downloading metadata for the image.
	Path *LookupError `json:"path,omitempty"`
}

func (u *ThumbnailError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// An error occurs when downloading metadata for the image.
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

type ThumbnailFormat struct {
	Tag string `json:".tag"`
}

type ThumbnailSize struct {
	Tag string `json:".tag"`
}

type UpdatePropertiesError struct {
	Tag                 string                 `json:".tag"`
	PropertyGroupLookup *LookUpPropertiesError `json:"property_group_lookup,omitempty"`
}

func (u *UpdatePropertiesError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag                 string          `json:".tag"`
		PropertyGroupLookup json.RawMessage `json:"property_group_lookup"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "property_group_lookup":
		{
			if len(w.PropertyGroupLookup) == 0 {
				break
			}
			if err := json.Unmarshal(w.PropertyGroupLookup, &u.PropertyGroupLookup); err != nil {
				return err
			}
		}
	}
	return nil
}

type UpdatePropertyGroupArg struct {
	// A unique identifier for the file.
	Path string `json:"path"`
	// Filled custom property templates associated with a file.
	UpdatePropertyGroups []*PropertyGroupUpdate `json:"update_property_groups"`
}

func NewUpdatePropertyGroupArg(Path string, UpdatePropertyGroups []*PropertyGroupUpdate) *UpdatePropertyGroupArg {
	s := new(UpdatePropertyGroupArg)
	s.Path = Path
	s.UpdatePropertyGroups = UpdatePropertyGroups
	return s
}

type UploadError struct {
	Tag string `json:".tag"`
	// Unable to save the uploaded contents to a file.
	Path *UploadWriteFailed `json:"path,omitempty"`
}

func (u *UploadError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// Unable to save the uploaded contents to a file.
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
			if err := json.Unmarshal(body, &u.Path); err != nil {
				return err
			}
		}
	}
	return nil
}

type UploadErrorWithProperties struct {
	Tag             string                     `json:".tag"`
	PropertiesError *InvalidPropertyGroupError `json:"properties_error,omitempty"`
}

func (u *UploadErrorWithProperties) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag             string          `json:".tag"`
		PropertiesError json.RawMessage `json:"properties_error"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "properties_error":
		{
			if len(w.PropertiesError) == 0 {
				break
			}
			if err := json.Unmarshal(w.PropertiesError, &u.PropertiesError); err != nil {
				return err
			}
		}
	}
	return nil
}

type UploadSessionAppendArg struct {
	// Contains the upload session ID and the offset.
	Cursor *UploadSessionCursor `json:"cursor"`
	// If true, the current session will be closed, at which point you won't be
	// able to call `UploadSessionAppendV2` anymore with the current session.
	Close bool `json:"close"`
}

func NewUploadSessionAppendArg(Cursor *UploadSessionCursor) *UploadSessionAppendArg {
	s := new(UploadSessionAppendArg)
	s.Cursor = Cursor
	s.Close = false
	return s
}

type UploadSessionCursor struct {
	// The upload session ID (returned by `UploadSessionStart`).
	SessionId string `json:"session_id"`
	// The amount of data that has been uploaded so far. We use this to make sure
	// upload data isn't lost or duplicated in the event of a network error.
	Offset uint64 `json:"offset"`
}

func NewUploadSessionCursor(SessionId string, Offset uint64) *UploadSessionCursor {
	s := new(UploadSessionCursor)
	s.SessionId = SessionId
	s.Offset = Offset
	return s
}

type UploadSessionFinishArg struct {
	// Contains the upload session ID and the offset.
	Cursor *UploadSessionCursor `json:"cursor"`
	// Contains the path and other optional modifiers for the commit.
	Commit *CommitInfo `json:"commit"`
}

func NewUploadSessionFinishArg(Cursor *UploadSessionCursor, Commit *CommitInfo) *UploadSessionFinishArg {
	s := new(UploadSessionFinishArg)
	s.Cursor = Cursor
	s.Commit = Commit
	return s
}

type UploadSessionFinishError struct {
	Tag string `json:".tag"`
	// The session arguments are incorrect; the value explains the reason.
	LookupFailed *UploadSessionLookupError `json:"lookup_failed,omitempty"`
	// Unable to save the uploaded contents to a file.
	Path *WriteError `json:"path,omitempty"`
}

func (u *UploadSessionFinishError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// The session arguments are incorrect; the value explains the reason.
		LookupFailed json.RawMessage `json:"lookup_failed"`
		// Unable to save the uploaded contents to a file.
		Path json.RawMessage `json:"path"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "lookup_failed":
		{
			if len(w.LookupFailed) == 0 {
				break
			}
			if err := json.Unmarshal(w.LookupFailed, &u.LookupFailed); err != nil {
				return err
			}
		}
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

type UploadSessionLookupError struct {
	Tag string `json:".tag"`
	// The specified offset was incorrect. See the value for the correct offset.
	// (This error may occur when a previous request was received and processed
	// successfully but the client did not receive the response, e.g. due to a
	// network error.)
	IncorrectOffset *UploadSessionOffsetError `json:"incorrect_offset,omitempty"`
}

func (u *UploadSessionLookupError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// The specified offset was incorrect. See the value for the correct offset.
		// (This error may occur when a previous request was received and processed
		// successfully but the client did not receive the response, e.g. due to a
		// network error.)
		IncorrectOffset json.RawMessage `json:"incorrect_offset"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "incorrect_offset":
		{
			if err := json.Unmarshal(body, &u.IncorrectOffset); err != nil {
				return err
			}
		}
	}
	return nil
}

type UploadSessionOffsetError struct {
	// The offset up to which data has been collected.
	CorrectOffset uint64 `json:"correct_offset"`
}

func NewUploadSessionOffsetError(CorrectOffset uint64) *UploadSessionOffsetError {
	s := new(UploadSessionOffsetError)
	s.CorrectOffset = CorrectOffset
	return s
}

type UploadSessionStartArg struct {
	// If true, the current session will be closed, at which point you won't be
	// able to call `UploadSessionAppendV2` anymore with the current session.
	Close bool `json:"close"`
}

func NewUploadSessionStartArg() *UploadSessionStartArg {
	s := new(UploadSessionStartArg)
	s.Close = false
	return s
}

type UploadSessionStartResult struct {
	// A unique identifier for the upload session. Pass this to
	// `UploadSessionAppendV2` and `UploadSessionFinish`.
	SessionId string `json:"session_id"`
}

func NewUploadSessionStartResult(SessionId string) *UploadSessionStartResult {
	s := new(UploadSessionStartResult)
	s.SessionId = SessionId
	return s
}

type UploadWriteFailed struct {
	// The reason why the file couldn't be saved.
	Reason *WriteError `json:"reason"`
	// The upload session ID; this may be used to retry the commit.
	UploadSessionId string `json:"upload_session_id"`
}

func NewUploadWriteFailed(Reason *WriteError, UploadSessionId string) *UploadWriteFailed {
	s := new(UploadWriteFailed)
	s.Reason = Reason
	s.UploadSessionId = UploadSessionId
	return s
}

// Metadata for a video.
type VideoMetadata struct {
	// Dimension of the photo/video.
	Dimensions *Dimensions `json:"dimensions,omitempty"`
	// The GPS coordinate of the photo/video.
	Location *GpsCoordinates `json:"location,omitempty"`
	// The timestamp when the photo/video is taken.
	TimeTaken time.Time `json:"time_taken,omitempty"`
	// The duration of the video in milliseconds.
	Duration uint64 `json:"duration,omitempty"`
}

func NewVideoMetadata() *VideoMetadata {
	s := new(VideoMetadata)
	return s
}

type WriteConflictError struct {
	Tag string `json:".tag"`
}

type WriteError struct {
	Tag           string `json:".tag"`
	MalformedPath string `json:"malformed_path,omitempty"`
	// Couldn't write to the target path because there was something in the way.
	Conflict *WriteConflictError `json:"conflict,omitempty"`
}

func (u *WriteError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag           string          `json:".tag"`
		MalformedPath json.RawMessage `json:"malformed_path,omitempty"`
		// Couldn't write to the target path because there was something in the way.
		Conflict json.RawMessage `json:"conflict"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "malformed_path":
		{
			if len(w.MalformedPath) == 0 {
				break
			}
			if err := json.Unmarshal(w.MalformedPath, &u.MalformedPath); err != nil {
				return err
			}
		}
	case "conflict":
		{
			if len(w.Conflict) == 0 {
				break
			}
			if err := json.Unmarshal(w.Conflict, &u.Conflict); err != nil {
				return err
			}
		}
	}
	return nil
}

// Your intent when writing a file to some path. This is used to determine what
// constitutes a conflict and what the autorename strategy is. In some
// situations, the conflict behavior is identical: (a) If the target path
// doesn't contain anything, the file is always written; no conflict. (b) If the
// target path contains a folder, it's always a conflict. (c) If the target path
// contains a file with identical contents, nothing gets written; no conflict.
// The conflict checking differs in the case where there's a file at the target
// path with contents different from the contents you're trying to write.
type WriteMode struct {
	Tag string `json:".tag"`
	// Overwrite if the given "rev" matches the existing file's "rev". The
	// autorename strategy is to append the string "conflicted copy" to the file
	// name. For example, "document.txt" might become "document (conflicted
	// copy).txt" or "document (Panda's conflicted copy).txt".
	Update string `json:"update,omitempty"`
}

func (u *WriteMode) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// Overwrite if the given "rev" matches the existing file's "rev". The
		// autorename strategy is to append the string "conflicted copy" to the file
		// name. For example, "document.txt" might become "document (conflicted
		// copy).txt" or "document (Panda's conflicted copy).txt".
		Update json.RawMessage `json:"update"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "update":
		{
			if len(w.Update) == 0 {
				break
			}
			if err := json.Unmarshal(w.Update, &u.Update); err != nil {
				return err
			}
		}
	}
	return nil
}

type Files interface {
	// Returns the metadata for a file or folder. This is an alpha endpoint
	// compatible with the properties API. Note: Metadata for the root folder is
	// unsupported.
	AlphaGetMetadata(arg *AlphaGetMetadataArg) (res *Metadata, err error)
	// Create a new file with the contents provided in the request. Note that this
	// endpoint is part of the properties API alpha and is slightly different from
	// `Upload`. Do not use this to upload a file larger than 150 MB. Instead,
	// create an upload session with `UploadSessionStart`.
	AlphaUpload(arg *CommitInfoWithProperties, content io.Reader) (res *FileMetadata, err error)
	// Copy a file or folder to a different location in the user's Dropbox. If the
	// source path is a folder all its contents will be copied.
	Copy(arg *RelocationArg) (res *Metadata, err error)
	// Get a copy reference to a file or folder. This reference string can be used
	// to save that file or folder to another user's Dropbox by passing it to
	// `CopyReferenceSave`.
	CopyReferenceGet(arg *GetCopyReferenceArg) (res *GetCopyReferenceResult, err error)
	// Save a copy reference returned by `CopyReferenceGet` to the user's Dropbox.
	CopyReferenceSave(arg *SaveCopyReferenceArg) (res *SaveCopyReferenceResult, err error)
	// Create a folder at a given path.
	CreateFolder(arg *CreateFolderArg) (res *FolderMetadata, err error)
	// Delete the file or folder at a given path. If the path is a folder, all its
	// contents will be deleted too. A successful response indicates that the file
	// or folder was deleted. The returned metadata will be the corresponding
	// `FileMetadata` or `FolderMetadata` for the item at time of deletion, and not
	// a `DeletedMetadata` object.
	Delete(arg *DeleteArg) (res *Metadata, err error)
	// Download a file from a user's Dropbox.
	Download(arg *DownloadArg) (res *FileMetadata, content io.ReadCloser, err error)
	// Returns the metadata for a file or folder. Note: Metadata for the root
	// folder is unsupported.
	GetMetadata(arg *GetMetadataArg) (res *Metadata, err error)
	// Get a preview for a file. Currently previews are only generated for the
	// files with  the following extensions: .doc, .docx, .docm, .ppt, .pps, .ppsx,
	// .ppsm, .pptx, .pptm,  .xls, .xlsx, .xlsm, .rtf
	GetPreview(arg *PreviewArg) (res *FileMetadata, content io.ReadCloser, err error)
	// Get a temporary link to stream content of a file. This link will expire in
	// four hours and afterwards you will get 410 Gone. Content-Type of the link is
	// determined automatically by the file's mime type.
	GetTemporaryLink(arg *GetTemporaryLinkArg) (res *GetTemporaryLinkResult, err error)
	// Get a thumbnail for an image. This method currently supports files with the
	// following file extensions: jpg, jpeg, png, tiff, tif, gif and bmp. Photos
	// that are larger than 20MB in size won't be converted to a thumbnail.
	GetThumbnail(arg *ThumbnailArg) (res *FileMetadata, content io.ReadCloser, err error)
	// Returns the contents of a folder.
	ListFolder(arg *ListFolderArg) (res *ListFolderResult, err error)
	// Once a cursor has been retrieved from `ListFolder`, use this to paginate
	// through all files and retrieve updates to the folder.
	ListFolderContinue(arg *ListFolderContinueArg) (res *ListFolderResult, err error)
	// A way to quickly get a cursor for the folder's state. Unlike `ListFolder`,
	// `ListFolderGetLatestCursor` doesn't return any entries. This endpoint is for
	// app which only needs to know about new files and modifications and doesn't
	// need to know about files that already exist in Dropbox.
	ListFolderGetLatestCursor(arg *ListFolderArg) (res *ListFolderGetLatestCursorResult, err error)
	// A longpoll endpoint to wait for changes on an account. In conjunction with
	// `ListFolderContinue`, this call gives you a low-latency way to monitor an
	// account for file changes. The connection will block until there are changes
	// available or a timeout occurs. This endpoint is useful mostly for
	// client-side apps. If you're looking for server-side notifications, check out
	// our `webhooks documentation`
	// <https://www.dropbox.com/developers/reference/webhooks>.
	ListFolderLongpoll(arg *ListFolderLongpollArg) (res *ListFolderLongpollResult, err error)
	// Return revisions of a file
	ListRevisions(arg *ListRevisionsArg) (res *ListRevisionsResult, err error)
	// Move a file or folder to a different location in the user's Dropbox. If the
	// source path is a folder all its contents will be moved.
	Move(arg *RelocationArg) (res *Metadata, err error)
	// Permanently delete the file or folder at a given path (see
	// https://www.dropbox.com/en/help/40). Note: This endpoint is only available
	// for Dropbox Business apps.
	PermanentlyDelete(arg *DeleteArg) (err error)
	// Add custom properties to a file using a filled property template. See
	// properties/template/add to create new property templates.
	PropertiesAdd(arg *PropertyGroupWithPath) (err error)
	// Overwrite custom properties from a specified template associated with a
	// file.
	PropertiesOverwrite(arg *PropertyGroupWithPath) (err error)
	// Remove all custom properties from a specified template associated with a
	// file. To remove specific property key value pairs, see `PropertiesUpdate`.
	// To update a property template, see properties/template/update. Property
	// templates can't be removed once created.
	PropertiesRemove(arg *RemovePropertiesArg) (err error)
	// Get the schema for a specified template.
	PropertiesTemplateGet(arg *properties.GetPropertyTemplateArg) (res *properties.GetPropertyTemplateResult, err error)
	// Get the property template identifiers for a user. To get the schema of each
	// template use `PropertiesTemplateGet`.
	PropertiesTemplateList() (res *properties.ListPropertyTemplateIds, err error)
	// Add, update or remove custom properties from a specified template associated
	// with a file. Fields that already exist and not described in the request will
	// not be modified.
	PropertiesUpdate(arg *UpdatePropertyGroupArg) (err error)
	// Restore a file to a specific revision
	Restore(arg *RestoreArg) (res *FileMetadata, err error)
	// Save a specified URL into a file in user's Dropbox. If the given path
	// already exists, the file will be renamed to avoid the conflict (e.g. myfile
	// (1).txt).
	SaveUrl(arg *SaveUrlArg) (res *SaveUrlResult, err error)
	// Check the status of a `SaveUrl` job.
	SaveUrlCheckJobStatus(arg *async.PollArg) (res *SaveUrlJobStatus, err error)
	// Searches for files and folders. Note: Recent changes may not immediately be
	// reflected in search results due to a short delay in indexing.
	Search(arg *SearchArg) (res *SearchResult, err error)
	// Create a new file with the contents provided in the request. Do not use this
	// to upload a file larger than 150 MB. Instead, create an upload session with
	// `UploadSessionStart`.
	Upload(arg *CommitInfo, content io.Reader) (res *FileMetadata, err error)
	// Append more data to an upload session. A single request should not upload
	// more than 150 MB of file contents.
	UploadSessionAppend(arg *UploadSessionCursor, content io.Reader) (err error)
	// Append more data to an upload session. When the parameter close is set, this
	// call will close the session. A single request should not upload more than
	// 150 MB of file contents.
	UploadSessionAppendV2(arg *UploadSessionAppendArg, content io.Reader) (err error)
	// Finish an upload session and save the uploaded data to the given file path.
	// A single request should not upload more than 150 MB of file contents.
	UploadSessionFinish(arg *UploadSessionFinishArg, content io.Reader) (res *FileMetadata, err error)
	// Upload sessions allow you to upload a single file using multiple requests.
	// This call starts a new upload session with the given data.  You can then use
	// `UploadSessionAppendV2` to add more data and `UploadSessionFinish` to save
	// all the data to a file in Dropbox. A single request should not upload more
	// than 150 MB of file contents.
	UploadSessionStart(arg *UploadSessionStartArg, content io.Reader) (res *UploadSessionStartResult, err error)
}
