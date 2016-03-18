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

package team

import (
	"encoding/json"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/async"
	"github.com/dropbox/dropbox-sdk-go-unofficial/users"
)

type DeviceSession struct {
	// The session id
	SessionId string `json:"session_id"`
	// The IP address of the last activity from this session
	IpAddress string `json:"ip_address,omitempty"`
	// The country from which the last activity from this session was made
	Country string `json:"country,omitempty"`
	// The time this session was created
	Created time.Time `json:"created,omitempty"`
	// The time of the last activity from this session
	Updated time.Time `json:"updated,omitempty"`
}

func NewDeviceSession(SessionId string) *DeviceSession {
	s := new(DeviceSession)
	s.SessionId = SessionId
	return s
}

// Information on active web sessions
type ActiveWebSession struct {
	// The session id
	SessionId string `json:"session_id"`
	// Information on the hosting device
	UserAgent string `json:"user_agent"`
	// Information on the hosting operating system
	Os string `json:"os"`
	// Information on the browser used for this web session
	Browser string `json:"browser"`
	// The IP address of the last activity from this session
	IpAddress string `json:"ip_address,omitempty"`
	// The country from which the last activity from this session was made
	Country string `json:"country,omitempty"`
	// The time this session was created
	Created time.Time `json:"created,omitempty"`
	// The time of the last activity from this session
	Updated time.Time `json:"updated,omitempty"`
}

func NewActiveWebSession(SessionId string, UserAgent string, Os string, Browser string) *ActiveWebSession {
	s := new(ActiveWebSession)
	s.SessionId = SessionId
	s.UserAgent = UserAgent
	s.Os = Os
	s.Browser = Browser
	return s
}

// Describes which team-related admin permissions a user has.
type AdminTier struct {
	Tag string `json:".tag"`
}

// Information on linked third party applications
type ApiApp struct {
	// The application unique id
	AppId string `json:"app_id"`
	// The application name
	AppName string `json:"app_name"`
	// Whether the linked application uses a dedicated folder
	IsAppFolder bool `json:"is_app_folder"`
	// The application publisher name
	Publisher string `json:"publisher,omitempty"`
	// The publisher's URL
	PublisherUrl string `json:"publisher_url,omitempty"`
	// The time this application was linked
	Linked time.Time `json:"linked,omitempty"`
}

func NewApiApp(AppId string, AppName string, IsAppFolder bool) *ApiApp {
	s := new(ApiApp)
	s.AppId = AppId
	s.AppName = AppName
	s.IsAppFolder = IsAppFolder
	return s
}

// Base report structure.
type BaseDfbReport struct {
	// First date present in the results as 'YYYY-MM-DD' or None.
	StartDate string `json:"start_date"`
}

func NewBaseDfbReport(StartDate string) *BaseDfbReport {
	s := new(BaseDfbReport)
	s.StartDate = StartDate
	return s
}

// Input arguments that can be provided for most reports.
type DateRange struct {
	// Optional starting date (inclusive)
	StartDate time.Time `json:"start_date,omitempty"`
	// Optional ending date (exclusive)
	EndDate time.Time `json:"end_date,omitempty"`
}

func NewDateRange() *DateRange {
	s := new(DateRange)
	return s
}

// Errors that can originate from problems in input arguments to reports.
type DateRangeError struct {
	Tag string `json:".tag"`
}

// Information about linked Dropbox desktop client sessions
type DesktopClientSession struct {
	// The session id
	SessionId string `json:"session_id"`
	// Name of the hosting desktop
	HostName string `json:"host_name"`
	// The Dropbox desktop client type
	ClientType *DesktopPlatform `json:"client_type"`
	// The Dropbox client version
	ClientVersion string `json:"client_version"`
	// Information on the hosting platform
	Platform string `json:"platform"`
	// Whether it's possible to delete all of the account files upon unlinking
	IsDeleteOnUnlinkSupported bool `json:"is_delete_on_unlink_supported"`
	// The IP address of the last activity from this session
	IpAddress string `json:"ip_address,omitempty"`
	// The country from which the last activity from this session was made
	Country string `json:"country,omitempty"`
	// The time this session was created
	Created time.Time `json:"created,omitempty"`
	// The time of the last activity from this session
	Updated time.Time `json:"updated,omitempty"`
}

func NewDesktopClientSession(SessionId string, HostName string, ClientType *DesktopPlatform, ClientVersion string, Platform string, IsDeleteOnUnlinkSupported bool) *DesktopClientSession {
	s := new(DesktopClientSession)
	s.SessionId = SessionId
	s.HostName = HostName
	s.ClientType = ClientType
	s.ClientVersion = ClientVersion
	s.Platform = Platform
	s.IsDeleteOnUnlinkSupported = IsDeleteOnUnlinkSupported
	return s
}

type DesktopPlatform struct {
	Tag string `json:".tag"`
}

type DeviceSessionArg struct {
	// The session id
	SessionId string `json:"session_id"`
	// The unique id of the member owning the device
	TeamMemberId string `json:"team_member_id"`
}

func NewDeviceSessionArg(SessionId string, TeamMemberId string) *DeviceSessionArg {
	s := new(DeviceSessionArg)
	s.SessionId = SessionId
	s.TeamMemberId = TeamMemberId
	return s
}

// Each of the items is an array of values, one value per day. The value is the
// number of devices active within a time window, ending with that day. If there
// is no data for a day, then the value will be None.
type DevicesActive struct {
	// Array of number of linked windows (desktop) clients with activity.
	Windows []uint64 `json:"windows"`
	// Array of number of linked mac (desktop) clients with activity.
	Macos []uint64 `json:"macos"`
	// Array of number of linked linus (desktop) clients with activity.
	Linux []uint64 `json:"linux"`
	// Array of number of linked ios devices with activity.
	Ios []uint64 `json:"ios"`
	// Array of number of linked android devices with activity.
	Android []uint64 `json:"android"`
	// Array of number of other linked devices (blackberry, windows phone, etc)
	// with activity.
	Other []uint64 `json:"other"`
	// Array of total number of linked clients with activity.
	Total []uint64 `json:"total"`
}

func NewDevicesActive(Windows []uint64, Macos []uint64, Linux []uint64, Ios []uint64, Android []uint64, Other []uint64, Total []uint64) *DevicesActive {
	s := new(DevicesActive)
	s.Windows = Windows
	s.Macos = Macos
	s.Linux = Linux
	s.Ios = Ios
	s.Android = Android
	s.Other = Other
	s.Total = Total
	return s
}

type EmmState struct {
	Tag string `json:".tag"`
}

// Activity Report Result. Each of the items in the storage report is an array
// of values, one value per day. If there is no data for a day, then the value
// will be None.
type GetActivityReport struct {
	// First date present in the results as 'YYYY-MM-DD' or None.
	StartDate string `json:"start_date"`
	// Array of total number of adds by team members.
	Adds []uint64 `json:"adds"`
	// Array of number of edits by team members. If the same user edits the same
	// file multiple times this is counted as a single edit.
	Edits []uint64 `json:"edits"`
	// Array of total number of deletes by team members.
	Deletes []uint64 `json:"deletes"`
	// Array of the number of users who have been active in the last 28 days.
	ActiveUsers28Day []uint64 `json:"active_users_28_day"`
	// Array of the number of users who have been active in the last week.
	ActiveUsers7Day []uint64 `json:"active_users_7_day"`
	// Array of the number of users who have been active in the last day.
	ActiveUsers1Day []uint64 `json:"active_users_1_day"`
	// Array of the number of shared folders with some activity in the last 28
	// days.
	ActiveSharedFolders28Day []uint64 `json:"active_shared_folders_28_day"`
	// Array of the number of shared folders with some activity in the last week.
	ActiveSharedFolders7Day []uint64 `json:"active_shared_folders_7_day"`
	// Array of the number of shared folders with some activity in the last day.
	ActiveSharedFolders1Day []uint64 `json:"active_shared_folders_1_day"`
	// Array of the number of shared links created.
	SharedLinksCreated []uint64 `json:"shared_links_created"`
	// Array of the number of views by team users to shared links created by the
	// team.
	SharedLinksViewedByTeam []uint64 `json:"shared_links_viewed_by_team"`
	// Array of the number of views by users outside of the team to shared links
	// created by the team.
	SharedLinksViewedByOutsideUser []uint64 `json:"shared_links_viewed_by_outside_user"`
	// Array of the number of views by non-logged-in users to shared links created
	// by the team.
	SharedLinksViewedByNotLoggedIn []uint64 `json:"shared_links_viewed_by_not_logged_in"`
	// Array of the total number of views to shared links created by the team.
	SharedLinksViewedTotal []uint64 `json:"shared_links_viewed_total"`
}

func NewGetActivityReport(StartDate string, Adds []uint64, Edits []uint64, Deletes []uint64, ActiveUsers28Day []uint64, ActiveUsers7Day []uint64, ActiveUsers1Day []uint64, ActiveSharedFolders28Day []uint64, ActiveSharedFolders7Day []uint64, ActiveSharedFolders1Day []uint64, SharedLinksCreated []uint64, SharedLinksViewedByTeam []uint64, SharedLinksViewedByOutsideUser []uint64, SharedLinksViewedByNotLoggedIn []uint64, SharedLinksViewedTotal []uint64) *GetActivityReport {
	s := new(GetActivityReport)
	s.StartDate = StartDate
	s.Adds = Adds
	s.Edits = Edits
	s.Deletes = Deletes
	s.ActiveUsers28Day = ActiveUsers28Day
	s.ActiveUsers7Day = ActiveUsers7Day
	s.ActiveUsers1Day = ActiveUsers1Day
	s.ActiveSharedFolders28Day = ActiveSharedFolders28Day
	s.ActiveSharedFolders7Day = ActiveSharedFolders7Day
	s.ActiveSharedFolders1Day = ActiveSharedFolders1Day
	s.SharedLinksCreated = SharedLinksCreated
	s.SharedLinksViewedByTeam = SharedLinksViewedByTeam
	s.SharedLinksViewedByOutsideUser = SharedLinksViewedByOutsideUser
	s.SharedLinksViewedByNotLoggedIn = SharedLinksViewedByNotLoggedIn
	s.SharedLinksViewedTotal = SharedLinksViewedTotal
	return s
}

// Devices Report Result. Contains subsections for different time ranges of
// activity. Each of the items in each subsection of the storage report is an
// array of values, one value per day. If there is no data for a day, then the
// value will be None.
type GetDevicesReport struct {
	// First date present in the results as 'YYYY-MM-DD' or None.
	StartDate string `json:"start_date"`
	// Report of the number of devices active in the last day.
	Active1Day *DevicesActive `json:"active_1_day"`
	// Report of the number of devices active in the last 7 days.
	Active7Day *DevicesActive `json:"active_7_day"`
	// Report of the number of devices active in the last 28 days.
	Active28Day *DevicesActive `json:"active_28_day"`
}

func NewGetDevicesReport(StartDate string, Active1Day *DevicesActive, Active7Day *DevicesActive, Active28Day *DevicesActive) *GetDevicesReport {
	s := new(GetDevicesReport)
	s.StartDate = StartDate
	s.Active1Day = Active1Day
	s.Active7Day = Active7Day
	s.Active28Day = Active28Day
	return s
}

// Membership Report Result. Each of the items in the storage report is an array
// of values, one value per day. If there is no data for a day, then the value
// will be None.
type GetMembershipReport struct {
	// First date present in the results as 'YYYY-MM-DD' or None.
	StartDate string `json:"start_date"`
	// Team size, for each day.
	TeamSize []uint64 `json:"team_size"`
	// The number of pending invites to the team, for each day.
	PendingInvites []uint64 `json:"pending_invites"`
	// The number of members that joined the team, for each day.
	MembersJoined []uint64 `json:"members_joined"`
	// The number of suspended team members, for each day.
	SuspendedMembers []uint64 `json:"suspended_members"`
	// The total number of licenses the team has, for each day.
	Licenses []uint64 `json:"licenses"`
}

func NewGetMembershipReport(StartDate string, TeamSize []uint64, PendingInvites []uint64, MembersJoined []uint64, SuspendedMembers []uint64, Licenses []uint64) *GetMembershipReport {
	s := new(GetMembershipReport)
	s.StartDate = StartDate
	s.TeamSize = TeamSize
	s.PendingInvites = PendingInvites
	s.MembersJoined = MembersJoined
	s.SuspendedMembers = SuspendedMembers
	s.Licenses = Licenses
	return s
}

// Storage Report Result. Each of the items in the storage report is an array of
// values, one value per day. If there is no data for a day, then the value will
// be None.
type GetStorageReport struct {
	// First date present in the results as 'YYYY-MM-DD' or None.
	StartDate string `json:"start_date"`
	// Sum of the shared, unshared, and datastore usages, for each day.
	TotalUsage []uint64 `json:"total_usage"`
	// Array of the combined size (bytes) of team members' shared folders, for each
	// day.
	SharedUsage []uint64 `json:"shared_usage"`
	// Array of the combined size (bytes) of team members' root namespaces, for
	// each day.
	UnsharedUsage []uint64 `json:"unshared_usage"`
	// Array of the number of shared folders owned by team members, for each day.
	SharedFolders []uint64 `json:"shared_folders"`
	// Array of storage summaries of team members' account sizes. Each storage
	// summary is an array of key, value pairs, where each pair describes a storage
	// bucket. The key indicates the upper bound of the bucket and the value is the
	// number of users in that bucket. There is one such summary per day. If there
	// is no data for a day, the storage summary will be empty.
	MemberStorageMap [][]*StorageBucket `json:"member_storage_map"`
}

func NewGetStorageReport(StartDate string, TotalUsage []uint64, SharedUsage []uint64, UnsharedUsage []uint64, SharedFolders []uint64, MemberStorageMap [][]*StorageBucket) *GetStorageReport {
	s := new(GetStorageReport)
	s.StartDate = StartDate
	s.TotalUsage = TotalUsage
	s.SharedUsage = SharedUsage
	s.UnsharedUsage = UnsharedUsage
	s.SharedFolders = SharedFolders
	s.MemberStorageMap = MemberStorageMap
	return s
}

// Role of a user in group.
type GroupAccessType struct {
	Tag string `json:".tag"`
}

type GroupCreateArg struct {
	// Group name.
	GroupName string `json:"group_name"`
	// Optional argument. The creator of a team can associate an arbitrary external
	// ID to the group.
	GroupExternalId string `json:"group_external_id,omitempty"`
}

func NewGroupCreateArg(GroupName string) *GroupCreateArg {
	s := new(GroupCreateArg)
	s.GroupName = GroupName
	return s
}

type GroupCreateError struct {
	Tag string `json:".tag"`
}

// Error that can be raised when `GroupSelector`is used.
type GroupSelectorError struct {
	Tag string `json:".tag"`
}

type GroupDeleteError struct {
	Tag string `json:".tag"`
}

// Information about a group.
type GroupSummary struct {
	GroupName string `json:"group_name"`
	GroupId   string `json:"group_id"`
	// The number of members in the group.
	MemberCount uint32 `json:"member_count"`
	// External ID of group. This is an arbitrary ID that an admin can attach to a
	// group.
	GroupExternalId string `json:"group_external_id,omitempty"`
}

func NewGroupSummary(GroupName string, GroupId string, MemberCount uint32) *GroupSummary {
	s := new(GroupSummary)
	s.GroupName = GroupName
	s.GroupId = GroupId
	s.MemberCount = MemberCount
	return s
}

// Full description of a group.
type GroupFullInfo struct {
	GroupName string `json:"group_name"`
	GroupId   string `json:"group_id"`
	// The number of members in the group.
	MemberCount uint32 `json:"member_count"`
	// List of group members.
	Members []*GroupMemberInfo `json:"members"`
	// The group creation time as a UTC timestamp in milliseconds since the Unix
	// epoch.
	Created uint64 `json:"created"`
	// External ID of group. This is an arbitrary ID that an admin can attach to a
	// group.
	GroupExternalId string `json:"group_external_id,omitempty"`
}

func NewGroupFullInfo(GroupName string, GroupId string, MemberCount uint32, Members []*GroupMemberInfo, Created uint64) *GroupFullInfo {
	s := new(GroupFullInfo)
	s.GroupName = GroupName
	s.GroupId = GroupId
	s.MemberCount = MemberCount
	s.Members = Members
	s.Created = Created
	return s
}

// Profile of group member, and role in group.
type GroupMemberInfo struct {
	// Profile of group member.
	Profile *MemberProfile `json:"profile"`
	// The role that the user has in the group.
	AccessType *GroupAccessType `json:"access_type"`
}

func NewGroupMemberInfo(Profile *MemberProfile, AccessType *GroupAccessType) *GroupMemberInfo {
	s := new(GroupMemberInfo)
	s.Profile = Profile
	s.AccessType = AccessType
	return s
}

// Argument for selecting a group and a single user.
type GroupMemberSelector struct {
	// Specify a group.
	Group *GroupSelector `json:"group"`
	// Identity of a user that is a member of `group`.
	User *UserSelectorArg `json:"user"`
}

func NewGroupMemberSelector(Group *GroupSelector, User *UserSelectorArg) *GroupMemberSelector {
	s := new(GroupMemberSelector)
	s.Group = Group
	s.User = User
	return s
}

// Error that can be raised when `GroupMemberSelector` is used, and the user is
// required to be a member of the specified group.
type GroupMemberSelectorError struct {
	Tag string `json:".tag"`
}

type GroupMembersAddArg struct {
	// Group to which users will be added.
	Group *GroupSelector `json:"group"`
	// List of users to be added to the group.
	Members []*MemberAccess `json:"members"`
}

func NewGroupMembersAddArg(Group *GroupSelector, Members []*MemberAccess) *GroupMembersAddArg {
	s := new(GroupMembersAddArg)
	s.Group = Group
	s.Members = Members
	return s
}

type GroupMembersAddError struct {
	Tag string `json:".tag"`
	// These members are not part of your team. Currently, you cannot add members
	// to a group if they are not part of your team, though this may change in a
	// subsequent version. To add new members to your Dropbox Business team, use
	// the `MembersAdd` endpoint.
	MembersNotInTeam []string `json:"members_not_in_team,omitempty"`
	// These users were not found in Dropbox.
	UsersNotFound []string `json:"users_not_found,omitempty"`
}

func (u *GroupMembersAddError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// These members are not part of your team. Currently, you cannot add members
		// to a group if they are not part of your team, though this may change in a
		// subsequent version. To add new members to your Dropbox Business team, use
		// the `MembersAdd` endpoint.
		MembersNotInTeam json.RawMessage `json:"members_not_in_team"`
		// These users were not found in Dropbox.
		UsersNotFound json.RawMessage `json:"users_not_found"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "members_not_in_team":
		{
			if len(w.MembersNotInTeam) == 0 {
				break
			}
			if err := json.Unmarshal(w.MembersNotInTeam, &u.MembersNotInTeam); err != nil {
				return err
			}
		}
	case "users_not_found":
		{
			if len(w.UsersNotFound) == 0 {
				break
			}
			if err := json.Unmarshal(w.UsersNotFound, &u.UsersNotFound); err != nil {
				return err
			}
		}
	}
	return nil
}

// Result returned by `GroupsMembersAdd` and `GroupsMembersRemove`.
type GroupMembersChangeResult struct {
	// Lists the group members after the member change operation has been
	// performed.
	GroupInfo *GroupFullInfo `json:"group_info"`
	// An ID that can be used to obtain the status of granting/revoking group-owned
	// resources.
	AsyncJobId string `json:"async_job_id"`
}

func NewGroupMembersChangeResult(GroupInfo *GroupFullInfo, AsyncJobId string) *GroupMembersChangeResult {
	s := new(GroupMembersChangeResult)
	s.GroupInfo = GroupInfo
	s.AsyncJobId = AsyncJobId
	return s
}

type GroupMembersRemoveArg struct {
	Group *GroupSelector     `json:"group"`
	Users []*UserSelectorArg `json:"users"`
}

func NewGroupMembersRemoveArg(Group *GroupSelector, Users []*UserSelectorArg) *GroupMembersRemoveArg {
	s := new(GroupMembersRemoveArg)
	s.Group = Group
	s.Users = Users
	return s
}

// Error that can be raised when `GroupMembersSelector` is used, and the users
// are required to be members of the specified group.
type GroupMembersSelectorError struct {
	Tag string `json:".tag"`
}

type GroupMembersRemoveError struct {
	Tag string `json:".tag"`
}

// Argument for selecting a group and a list of users.
type GroupMembersSelector struct {
	// Specify a group.
	Group *GroupSelector `json:"group"`
	// A list of users that are members of `group`.
	Users *UsersSelectorArg `json:"users"`
}

func NewGroupMembersSelector(Group *GroupSelector, Users *UsersSelectorArg) *GroupMembersSelector {
	s := new(GroupMembersSelector)
	s.Group = Group
	s.Users = Users
	return s
}

type GroupMembersSetAccessTypeArg struct {
	// Specify a group.
	Group *GroupSelector `json:"group"`
	// Identity of a user that is a member of `group`.
	User *UserSelectorArg `json:"user"`
	// New group access type the user will have.
	AccessType *GroupAccessType `json:"access_type"`
}

func NewGroupMembersSetAccessTypeArg(Group *GroupSelector, User *UserSelectorArg, AccessType *GroupAccessType) *GroupMembersSetAccessTypeArg {
	s := new(GroupMembersSetAccessTypeArg)
	s.Group = Group
	s.User = User
	s.AccessType = AccessType
	return s
}

// Argument for selecting a single group, either by group_id or by external
// group ID.
type GroupSelector struct {
	Tag string `json:".tag"`
	// Group ID.
	GroupId string `json:"group_id,omitempty"`
	// External ID of the group.
	GroupExternalId string `json:"group_external_id,omitempty"`
}

func (u *GroupSelector) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// Group ID.
		GroupId json.RawMessage `json:"group_id"`
		// External ID of the group.
		GroupExternalId json.RawMessage `json:"group_external_id"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "group_id":
		{
			if len(w.GroupId) == 0 {
				break
			}
			if err := json.Unmarshal(w.GroupId, &u.GroupId); err != nil {
				return err
			}
		}
	case "group_external_id":
		{
			if len(w.GroupExternalId) == 0 {
				break
			}
			if err := json.Unmarshal(w.GroupExternalId, &u.GroupExternalId); err != nil {
				return err
			}
		}
	}
	return nil
}

// The group type determines how a group is created and managed.
type GroupType struct {
	Tag string `json:".tag"`
}

type GroupUpdateArgs struct {
	// Specify a group.
	Group *GroupSelector `json:"group"`
	// Optional argument. Set group name to this if provided.
	NewGroupName string `json:"new_group_name,omitempty"`
	// Optional argument. New group external ID. If the argument is None, the
	// group's external_id won't be updated. If the argument is empty string, the
	// group's external id will be cleared.
	NewGroupExternalId string `json:"new_group_external_id,omitempty"`
}

func NewGroupUpdateArgs(Group *GroupSelector) *GroupUpdateArgs {
	s := new(GroupUpdateArgs)
	s.Group = Group
	return s
}

type GroupUpdateError struct {
	Tag string `json:".tag"`
}

type GroupsGetInfoError struct {
	Tag string `json:".tag"`
}

type GroupsGetInfoItem struct {
	Tag string `json:".tag"`
	// An ID that was provided as a parameter to `GroupsGetInfo`, and did not match
	// a corresponding group. The ID can be a group ID, or an external ID,
	// depending on how the method was called.
	IdNotFound string `json:"id_not_found,omitempty"`
	// Info about a group.
	GroupInfo *GroupFullInfo `json:"group_info,omitempty"`
}

func (u *GroupsGetInfoItem) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// An ID that was provided as a parameter to `GroupsGetInfo`, and did not
		// match a corresponding group. The ID can be a group ID, or an external ID,
		// depending on how the method was called.
		IdNotFound json.RawMessage `json:"id_not_found"`
		// Info about a group.
		GroupInfo json.RawMessage `json:"group_info"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "id_not_found":
		{
			if len(w.IdNotFound) == 0 {
				break
			}
			if err := json.Unmarshal(w.IdNotFound, &u.IdNotFound); err != nil {
				return err
			}
		}
	case "group_info":
		{
			if err := json.Unmarshal(body, &u.GroupInfo); err != nil {
				return err
			}
		}
	}
	return nil
}

type GroupsListArg struct {
	// Number of results to return per call.
	Limit uint32 `json:"limit"`
}

func NewGroupsListArg() *GroupsListArg {
	s := new(GroupsListArg)
	s.Limit = 1000
	return s
}

type GroupsListContinueArg struct {
	// Indicates from what point to get the next set of groups.
	Cursor string `json:"cursor"`
}

func NewGroupsListContinueArg(Cursor string) *GroupsListContinueArg {
	s := new(GroupsListContinueArg)
	s.Cursor = Cursor
	return s
}

type GroupsListContinueError struct {
	Tag string `json:".tag"`
}

type GroupsListResult struct {
	Groups []*GroupSummary `json:"groups"`
	// Pass the cursor into `MembersListContinue` to obtain the additional members.
	Cursor string `json:"cursor"`
	// Is true if there are additional team members that have not been returned
	// yet. An additional call to `MembersListContinue` can retrieve them.
	HasMore bool `json:"has_more"`
}

func NewGroupsListResult(Groups []*GroupSummary, Cursor string, HasMore bool) *GroupsListResult {
	s := new(GroupsListResult)
	s.Groups = Groups
	s.Cursor = Cursor
	s.HasMore = HasMore
	return s
}

type GroupsPollError struct {
	Tag string `json:".tag"`
}

// Argument for selecting a list of groups, either by group_ids, or external
// group IDs.
type GroupsSelector struct {
	Tag string `json:".tag"`
	// List of group IDs.
	GroupIds []string `json:"group_ids,omitempty"`
	// List of external IDs of groups.
	GroupExternalIds []string `json:"group_external_ids,omitempty"`
}

func (u *GroupsSelector) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// List of group IDs.
		GroupIds json.RawMessage `json:"group_ids"`
		// List of external IDs of groups.
		GroupExternalIds json.RawMessage `json:"group_external_ids"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "group_ids":
		{
			if len(w.GroupIds) == 0 {
				break
			}
			if err := json.Unmarshal(w.GroupIds, &u.GroupIds); err != nil {
				return err
			}
		}
	case "group_external_ids":
		{
			if len(w.GroupExternalIds) == 0 {
				break
			}
			if err := json.Unmarshal(w.GroupExternalIds, &u.GroupExternalIds); err != nil {
				return err
			}
		}
	}
	return nil
}

type ListMemberAppsArg struct {
	// The team member id
	TeamMemberId string `json:"team_member_id"`
}

func NewListMemberAppsArg(TeamMemberId string) *ListMemberAppsArg {
	s := new(ListMemberAppsArg)
	s.TeamMemberId = TeamMemberId
	return s
}

// Error returned by `LinkedAppsListMemberLinkedApps`.
type ListMemberAppsError struct {
	Tag string `json:".tag"`
}

type ListMemberAppsResult struct {
	// List of third party applications linked by this team member
	LinkedApiApps []*ApiApp `json:"linked_api_apps"`
}

func NewListMemberAppsResult(LinkedApiApps []*ApiApp) *ListMemberAppsResult {
	s := new(ListMemberAppsResult)
	s.LinkedApiApps = LinkedApiApps
	return s
}

type ListMemberDevicesArg struct {
	// The team's member id
	TeamMemberId string `json:"team_member_id"`
	// Whether to list web sessions of the team's member
	IncludeWebSessions bool `json:"include_web_sessions"`
	// Whether to list linked desktop devices of the team's member
	IncludeDesktopClients bool `json:"include_desktop_clients"`
	// Whether to list linked mobile devices of the team's member
	IncludeMobileClients bool `json:"include_mobile_clients"`
}

func NewListMemberDevicesArg(TeamMemberId string) *ListMemberDevicesArg {
	s := new(ListMemberDevicesArg)
	s.TeamMemberId = TeamMemberId
	s.IncludeWebSessions = true
	s.IncludeDesktopClients = true
	s.IncludeMobileClients = true
	return s
}

type ListMemberDevicesError struct {
	Tag string `json:".tag"`
}

type ListMemberDevicesResult struct {
	// List of web sessions made by this team member
	ActiveWebSessions []*ActiveWebSession `json:"active_web_sessions,omitempty"`
	// List of desktop clients used by this team member
	DesktopClientSessions []*DesktopClientSession `json:"desktop_client_sessions,omitempty"`
	// List of mobile client used by this team member
	MobileClientSessions []*MobileClientSession `json:"mobile_client_sessions,omitempty"`
}

func NewListMemberDevicesResult() *ListMemberDevicesResult {
	s := new(ListMemberDevicesResult)
	return s
}

// Arguments for `LinkedAppsListTeamLinkedApps`.
type ListTeamAppsArg struct {
	// At the first call to the `LinkedAppsListTeamLinkedApps` the cursor shouldn't
	// be passed. Then, if the result of the call includes a cursor, the following
	// requests should include the received cursors in order to receive the next
	// sub list of the team applications
	Cursor string `json:"cursor,omitempty"`
}

func NewListTeamAppsArg() *ListTeamAppsArg {
	s := new(ListTeamAppsArg)
	return s
}

// Error returned by `LinkedAppsListTeamLinkedApps`
type ListTeamAppsError struct {
	Tag string `json:".tag"`
}

// Information returned by `LinkedAppsListTeamLinkedApps`.
type ListTeamAppsResult struct {
	// The linked applications of each member of the team
	Apps []*MemberLinkedApps `json:"apps"`
	// If true, then there are more apps available. Pass the cursor to
	// `LinkedAppsListTeamLinkedApps` to retrieve the rest.
	HasMore bool `json:"has_more"`
	// Pass the cursor into `LinkedAppsListTeamLinkedApps` to receive the next sub
	// list of team's applications.
	Cursor string `json:"cursor,omitempty"`
}

func NewListTeamAppsResult(Apps []*MemberLinkedApps, HasMore bool) *ListTeamAppsResult {
	s := new(ListTeamAppsResult)
	s.Apps = Apps
	s.HasMore = HasMore
	return s
}

type ListTeamDevicesArg struct {
	// At the first call to the `DevicesListTeamDevices` the cursor shouldn't be
	// passed. Then, if the result of the call includes a cursor, the following
	// requests should include the received cursors in order to receive the next
	// sub list of team devices
	Cursor string `json:"cursor,omitempty"`
	// Whether to list web sessions of the team members
	IncludeWebSessions bool `json:"include_web_sessions"`
	// Whether to list desktop clients of the team members
	IncludeDesktopClients bool `json:"include_desktop_clients"`
	// Whether to list mobile clients of the team members
	IncludeMobileClients bool `json:"include_mobile_clients"`
}

func NewListTeamDevicesArg() *ListTeamDevicesArg {
	s := new(ListTeamDevicesArg)
	s.IncludeWebSessions = true
	s.IncludeDesktopClients = true
	s.IncludeMobileClients = true
	return s
}

type ListTeamDevicesError struct {
	Tag string `json:".tag"`
}

type ListTeamDevicesResult struct {
	// The devices of each member of the team
	Devices []*MemberDevices `json:"devices"`
	// If true, then there are more devices available. Pass the cursor to
	// `DevicesListTeamDevices` to retrieve the rest.
	HasMore bool `json:"has_more"`
	// Pass the cursor into `DevicesListTeamDevices` to receive the next sub list
	// of team's devices.
	Cursor string `json:"cursor,omitempty"`
}

func NewListTeamDevicesResult(Devices []*MemberDevices, HasMore bool) *ListTeamDevicesResult {
	s := new(ListTeamDevicesResult)
	s.Devices = Devices
	s.HasMore = HasMore
	return s
}

// Specify access type a member should have when joined to a group.
type MemberAccess struct {
	// Identity of a user.
	User *UserSelectorArg `json:"user"`
	// Access type.
	AccessType *GroupAccessType `json:"access_type"`
}

func NewMemberAccess(User *UserSelectorArg, AccessType *GroupAccessType) *MemberAccess {
	s := new(MemberAccess)
	s.User = User
	s.AccessType = AccessType
	return s
}

type MemberAddArg struct {
	MemberEmail string `json:"member_email"`
	// Member's first name.
	MemberGivenName string `json:"member_given_name"`
	// Member's last name.
	MemberSurname string `json:"member_surname"`
	// External ID for member.
	MemberExternalId string `json:"member_external_id,omitempty"`
	// Whether to send a welcome email to the member. If send_welcome_email is
	// false, no email invitation will be sent to the user. This may be useful for
	// apps using single sign-on (SSO) flows for onboarding that want to handle
	// announcements themselves.
	SendWelcomeEmail bool       `json:"send_welcome_email"`
	Role             *AdminTier `json:"role"`
}

func NewMemberAddArg(MemberEmail string, MemberGivenName string, MemberSurname string) *MemberAddArg {
	s := new(MemberAddArg)
	s.MemberEmail = MemberEmail
	s.MemberGivenName = MemberGivenName
	s.MemberSurname = MemberSurname
	s.SendWelcomeEmail = true
	s.Role = &AdminTier{Tag: "member_only"}
	return s
}

// Describes the result of attempting to add a single user to the team.
// 'success' is the only value indicating that a user was indeed added to the
// team - the other values explain the type of failure that occurred, and
// include the email of the user for which the operation has failed.
type MemberAddResult struct {
	Tag string `json:".tag"`
	// Describes a user that was successfully added to the team.
	Success *TeamMemberInfo `json:"success,omitempty"`
	// Team is already full. The organization has no available licenses.
	TeamLicenseLimit string `json:"team_license_limit,omitempty"`
	// Team is already full. The free team member limit has been reached.
	FreeTeamMemberLimitReached string `json:"free_team_member_limit_reached,omitempty"`
	// User is already on this team. The provided email address is associated with
	// a user who is already a member of or invited to the team.
	UserAlreadyOnTeam string `json:"user_already_on_team,omitempty"`
	// User is already on another team. The provided email address is associated
	// with a user that is already a member or invited to another team.
	UserOnAnotherTeam string `json:"user_on_another_team,omitempty"`
	// User is already paired.
	UserAlreadyPaired string `json:"user_already_paired,omitempty"`
	// User migration has failed.
	UserMigrationFailed string `json:"user_migration_failed,omitempty"`
	// A user with the given external member ID already exists on the team.
	DuplicateExternalMemberId string `json:"duplicate_external_member_id,omitempty"`
	// User creation has failed.
	UserCreationFailed string `json:"user_creation_failed,omitempty"`
}

func (u *MemberAddResult) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// Describes a user that was successfully added to the team.
		Success json.RawMessage `json:"success"`
		// Team is already full. The organization has no available licenses.
		TeamLicenseLimit json.RawMessage `json:"team_license_limit"`
		// Team is already full. The free team member limit has been reached.
		FreeTeamMemberLimitReached json.RawMessage `json:"free_team_member_limit_reached"`
		// User is already on this team. The provided email address is associated with
		// a user who is already a member of or invited to the team.
		UserAlreadyOnTeam json.RawMessage `json:"user_already_on_team"`
		// User is already on another team. The provided email address is associated
		// with a user that is already a member or invited to another team.
		UserOnAnotherTeam json.RawMessage `json:"user_on_another_team"`
		// User is already paired.
		UserAlreadyPaired json.RawMessage `json:"user_already_paired"`
		// User migration has failed.
		UserMigrationFailed json.RawMessage `json:"user_migration_failed"`
		// A user with the given external member ID already exists on the team.
		DuplicateExternalMemberId json.RawMessage `json:"duplicate_external_member_id"`
		// User creation has failed.
		UserCreationFailed json.RawMessage `json:"user_creation_failed"`
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
	case "team_license_limit":
		{
			if len(w.TeamLicenseLimit) == 0 {
				break
			}
			if err := json.Unmarshal(w.TeamLicenseLimit, &u.TeamLicenseLimit); err != nil {
				return err
			}
		}
	case "free_team_member_limit_reached":
		{
			if len(w.FreeTeamMemberLimitReached) == 0 {
				break
			}
			if err := json.Unmarshal(w.FreeTeamMemberLimitReached, &u.FreeTeamMemberLimitReached); err != nil {
				return err
			}
		}
	case "user_already_on_team":
		{
			if len(w.UserAlreadyOnTeam) == 0 {
				break
			}
			if err := json.Unmarshal(w.UserAlreadyOnTeam, &u.UserAlreadyOnTeam); err != nil {
				return err
			}
		}
	case "user_on_another_team":
		{
			if len(w.UserOnAnotherTeam) == 0 {
				break
			}
			if err := json.Unmarshal(w.UserOnAnotherTeam, &u.UserOnAnotherTeam); err != nil {
				return err
			}
		}
	case "user_already_paired":
		{
			if len(w.UserAlreadyPaired) == 0 {
				break
			}
			if err := json.Unmarshal(w.UserAlreadyPaired, &u.UserAlreadyPaired); err != nil {
				return err
			}
		}
	case "user_migration_failed":
		{
			if len(w.UserMigrationFailed) == 0 {
				break
			}
			if err := json.Unmarshal(w.UserMigrationFailed, &u.UserMigrationFailed); err != nil {
				return err
			}
		}
	case "duplicate_external_member_id":
		{
			if len(w.DuplicateExternalMemberId) == 0 {
				break
			}
			if err := json.Unmarshal(w.DuplicateExternalMemberId, &u.DuplicateExternalMemberId); err != nil {
				return err
			}
		}
	case "user_creation_failed":
		{
			if len(w.UserCreationFailed) == 0 {
				break
			}
			if err := json.Unmarshal(w.UserCreationFailed, &u.UserCreationFailed); err != nil {
				return err
			}
		}
	}
	return nil
}

// Information on devices of a team's member.
type MemberDevices struct {
	// The member unique Id
	TeamMemberId string `json:"team_member_id"`
	// List of web sessions made by this team member
	WebSessions []*ActiveWebSession `json:"web_sessions,omitempty"`
	// List of desktop clients by this team member
	DesktopClients []*DesktopClientSession `json:"desktop_clients,omitempty"`
	// List of mobile clients by this team member
	MobileClients []*MobileClientSession `json:"mobile_clients,omitempty"`
}

func NewMemberDevices(TeamMemberId string) *MemberDevices {
	s := new(MemberDevices)
	s.TeamMemberId = TeamMemberId
	return s
}

// Information on linked applications of a team member.
type MemberLinkedApps struct {
	// The member unique Id
	TeamMemberId string `json:"team_member_id"`
	// List of third party applications linked by this team member
	LinkedApiApps []*ApiApp `json:"linked_api_apps"`
}

func NewMemberLinkedApps(TeamMemberId string, LinkedApiApps []*ApiApp) *MemberLinkedApps {
	s := new(MemberLinkedApps)
	s.TeamMemberId = TeamMemberId
	s.LinkedApiApps = LinkedApiApps
	return s
}

// Basic member profile.
type MemberProfile struct {
	// ID of user as a member of a team.
	TeamMemberId string `json:"team_member_id"`
	// Email address of user.
	Email string `json:"email"`
	// Is true if the user's email is verified to be owned by the user.
	EmailVerified bool `json:"email_verified"`
	// The user's status as a member of a specific team.
	Status *TeamMemberStatus `json:"status"`
	// Representations for a person's name.
	Name *users.Name `json:"name"`
	// External ID that a team can attach to the user. An application using the API
	// may find it easier to use their own IDs instead of Dropbox IDs like
	// account_id or team_member_id.
	ExternalId string `json:"external_id,omitempty"`
}

func NewMemberProfile(TeamMemberId string, Email string, EmailVerified bool, Status *TeamMemberStatus, Name *users.Name) *MemberProfile {
	s := new(MemberProfile)
	s.TeamMemberId = TeamMemberId
	s.Email = Email
	s.EmailVerified = EmailVerified
	s.Status = Status
	s.Name = Name
	return s
}

// Error that can be returned whenever a struct derived from `UserSelectorArg`
// is used.
type UserSelectorError struct {
	Tag string `json:".tag"`
}

type MemberSelectorError struct {
	Tag string `json:".tag"`
}

type MembersAddArg struct {
	// Details of new members to be added to the team.
	NewMembers []*MemberAddArg `json:"new_members"`
	// Whether to force the add to happen asynchronously.
	ForceAsync bool `json:"force_async"`
}

func NewMembersAddArg(NewMembers []*MemberAddArg) *MembersAddArg {
	s := new(MembersAddArg)
	s.NewMembers = NewMembers
	s.ForceAsync = false
	return s
}

type MembersAddJobStatus struct {
	Tag string `json:".tag"`
	// The asynchronous job has finished. For each member that was specified in the
	// parameter `MembersAddArg` that was provided to `MembersAdd`, a corresponding
	// item is returned in this list.
	Complete []*MemberAddResult `json:"complete,omitempty"`
	// The asynchronous job returned an error. The string contains an error
	// message.
	Failed string `json:"failed,omitempty"`
}

func (u *MembersAddJobStatus) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// The asynchronous job has finished. For each member that was specified in
		// the parameter `MembersAddArg` that was provided to `MembersAdd`, a
		// corresponding item is returned in this list.
		Complete json.RawMessage `json:"complete"`
		// The asynchronous job returned an error. The string contains an error
		// message.
		Failed json.RawMessage `json:"failed"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "complete":
		{
			if len(w.Complete) == 0 {
				break
			}
			if err := json.Unmarshal(w.Complete, &u.Complete); err != nil {
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

type MembersAddLaunch struct {
	Tag      string             `json:".tag"`
	Complete []*MemberAddResult `json:"complete,omitempty"`
}

func (u *MembersAddLaunch) UnmarshalJSON(body []byte) error {
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
			if len(w.Complete) == 0 {
				break
			}
			if err := json.Unmarshal(w.Complete, &u.Complete); err != nil {
				return err
			}
		}
	}
	return nil
}

// Exactly one of team_member_id, email, or external_id must be provided to
// identify the user account.
type MembersDeactivateArg struct {
	// Identity of user to remove/suspend.
	User *UserSelectorArg `json:"user"`
	// If provided, controls if the user's data will be deleted on their linked
	// devices.
	WipeData bool `json:"wipe_data"`
}

func NewMembersDeactivateArg(User *UserSelectorArg) *MembersDeactivateArg {
	s := new(MembersDeactivateArg)
	s.User = User
	s.WipeData = true
	return s
}

type MembersDeactivateError struct {
	Tag string `json:".tag"`
}

type MembersGetInfoArgs struct {
	// List of team members.
	Members []*UserSelectorArg `json:"members"`
}

func NewMembersGetInfoArgs(Members []*UserSelectorArg) *MembersGetInfoArgs {
	s := new(MembersGetInfoArgs)
	s.Members = Members
	return s
}

type MembersGetInfoError struct {
	Tag string `json:".tag"`
}

// Describes a result obtained for a single user whose id was specified in the
// parameter of `MembersGetInfo`.
type MembersGetInfoItem struct {
	Tag string `json:".tag"`
	// An ID that was provided as a parameter to `MembersGetInfo`, and did not
	// match a corresponding user. This might be a team_member_id, an email, or an
	// external ID, depending on how the method was called.
	IdNotFound string `json:"id_not_found,omitempty"`
	// Info about a team member.
	MemberInfo *TeamMemberInfo `json:"member_info,omitempty"`
}

func (u *MembersGetInfoItem) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// An ID that was provided as a parameter to `MembersGetInfo`, and did not
		// match a corresponding user. This might be a team_member_id, an email, or an
		// external ID, depending on how the method was called.
		IdNotFound json.RawMessage `json:"id_not_found"`
		// Info about a team member.
		MemberInfo json.RawMessage `json:"member_info"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "id_not_found":
		{
			if len(w.IdNotFound) == 0 {
				break
			}
			if err := json.Unmarshal(w.IdNotFound, &u.IdNotFound); err != nil {
				return err
			}
		}
	case "member_info":
		{
			if err := json.Unmarshal(body, &u.MemberInfo); err != nil {
				return err
			}
		}
	}
	return nil
}

type MembersListArg struct {
	// Number of results to return per call.
	Limit uint32 `json:"limit"`
}

func NewMembersListArg() *MembersListArg {
	s := new(MembersListArg)
	s.Limit = 1000
	return s
}

type MembersListContinueArg struct {
	// Indicates from what point to get the next set of members.
	Cursor string `json:"cursor"`
}

func NewMembersListContinueArg(Cursor string) *MembersListContinueArg {
	s := new(MembersListContinueArg)
	s.Cursor = Cursor
	return s
}

type MembersListContinueError struct {
	Tag string `json:".tag"`
}

type MembersListError struct {
	Tag string `json:".tag"`
}

type MembersListResult struct {
	// List of team members.
	Members []*TeamMemberInfo `json:"members"`
	// Pass the cursor into `MembersListContinue` to obtain the additional members.
	Cursor string `json:"cursor"`
	// Is true if there are additional team members that have not been returned
	// yet. An additional call to `MembersListContinue` can retrieve them.
	HasMore bool `json:"has_more"`
}

func NewMembersListResult(Members []*TeamMemberInfo, Cursor string, HasMore bool) *MembersListResult {
	s := new(MembersListResult)
	s.Members = Members
	s.Cursor = Cursor
	s.HasMore = HasMore
	return s
}

type MembersRemoveArg struct {
	// Identity of user to remove/suspend.
	User *UserSelectorArg `json:"user"`
	// If provided, controls if the user's data will be deleted on their linked
	// devices.
	WipeData bool `json:"wipe_data"`
	// If provided, files from the deleted member account will be transferred to
	// this user.
	TransferDestId *UserSelectorArg `json:"transfer_dest_id,omitempty"`
	// If provided, errors during the transfer process will be sent via email to
	// this user. If the transfer_dest_id argument was provided, then this argument
	// must be provided as well.
	TransferAdminId *UserSelectorArg `json:"transfer_admin_id,omitempty"`
}

func NewMembersRemoveArg(User *UserSelectorArg) *MembersRemoveArg {
	s := new(MembersRemoveArg)
	s.User = User
	s.WipeData = true
	return s
}

type MembersRemoveError struct {
	Tag string `json:".tag"`
}

type MembersSendWelcomeError struct {
	Tag string `json:".tag"`
}

// Exactly one of team_member_id, email, or external_id must be provided to
// identify the user account.
type MembersSetPermissionsArg struct {
	// Identity of user whose role will be set.
	User *UserSelectorArg `json:"user"`
	// The new role of the member.
	NewRole *AdminTier `json:"new_role"`
}

func NewMembersSetPermissionsArg(User *UserSelectorArg, NewRole *AdminTier) *MembersSetPermissionsArg {
	s := new(MembersSetPermissionsArg)
	s.User = User
	s.NewRole = NewRole
	return s
}

type MembersSetPermissionsError struct {
	Tag string `json:".tag"`
}

type MembersSetPermissionsResult struct {
	// The member ID of the user to which the change was applied.
	TeamMemberId string `json:"team_member_id"`
	// The role after the change.
	Role *AdminTier `json:"role"`
}

func NewMembersSetPermissionsResult(TeamMemberId string, Role *AdminTier) *MembersSetPermissionsResult {
	s := new(MembersSetPermissionsResult)
	s.TeamMemberId = TeamMemberId
	s.Role = Role
	return s
}

// Exactly one of team_member_id, email, or external_id must be provided to
// identify the user account. At least one of new_email, new_external_id,
// new_given_name, and/or new_surname must be provided.
type MembersSetProfileArg struct {
	// Identity of user whose profile will be set.
	User *UserSelectorArg `json:"user"`
	// New email for member.
	NewEmail string `json:"new_email,omitempty"`
	// New external ID for member.
	NewExternalId string `json:"new_external_id,omitempty"`
	// New given name for member.
	NewGivenName string `json:"new_given_name,omitempty"`
	// New surname for member.
	NewSurname string `json:"new_surname,omitempty"`
}

func NewMembersSetProfileArg(User *UserSelectorArg) *MembersSetProfileArg {
	s := new(MembersSetProfileArg)
	s.User = User
	return s
}

type MembersSetProfileError struct {
	Tag string `json:".tag"`
}

type MembersSuspendError struct {
	Tag string `json:".tag"`
}

// Exactly one of team_member_id, email, or external_id must be provided to
// identify the user account.
type MembersUnsuspendArg struct {
	// Identity of user to unsuspend.
	User *UserSelectorArg `json:"user"`
}

func NewMembersUnsuspendArg(User *UserSelectorArg) *MembersUnsuspendArg {
	s := new(MembersUnsuspendArg)
	s.User = User
	return s
}

type MembersUnsuspendError struct {
	Tag string `json:".tag"`
}

type MobileClientPlatform struct {
	Tag string `json:".tag"`
}

// Information about linked Dropbox mobile client sessions
type MobileClientSession struct {
	// The session id
	SessionId string `json:"session_id"`
	// The device name
	DeviceName string `json:"device_name"`
	// The mobile application type
	ClientType *MobileClientPlatform `json:"client_type"`
	// The IP address of the last activity from this session
	IpAddress string `json:"ip_address,omitempty"`
	// The country from which the last activity from this session was made
	Country string `json:"country,omitempty"`
	// The time this session was created
	Created time.Time `json:"created,omitempty"`
	// The time of the last activity from this session
	Updated time.Time `json:"updated,omitempty"`
	// The dropbox client version
	ClientVersion string `json:"client_version,omitempty"`
	// The hosting OS version
	OsVersion string `json:"os_version,omitempty"`
	// last carrier used by the device
	LastCarrier string `json:"last_carrier,omitempty"`
}

func NewMobileClientSession(SessionId string, DeviceName string, ClientType *MobileClientPlatform) *MobileClientSession {
	s := new(MobileClientSession)
	s.SessionId = SessionId
	s.DeviceName = DeviceName
	s.ClientType = ClientType
	return s
}

type RevokeDesktopClientArg struct {
	// The session id
	SessionId string `json:"session_id"`
	// The unique id of the member owning the device
	TeamMemberId string `json:"team_member_id"`
	// Whether to delete all files of the account (this is possible only if
	// supported by the desktop client and  will be made the next time the client
	// access the account)
	DeleteOnUnlink bool `json:"delete_on_unlink"`
}

func NewRevokeDesktopClientArg(SessionId string, TeamMemberId string) *RevokeDesktopClientArg {
	s := new(RevokeDesktopClientArg)
	s.SessionId = SessionId
	s.TeamMemberId = TeamMemberId
	s.DeleteOnUnlink = false
	return s
}

type RevokeDeviceSessionArg struct {
	Tag string `json:".tag"`
	// End an active session
	WebSession *DeviceSessionArg `json:"web_session,omitempty"`
	// Unlink a linked desktop device
	DesktopClient *RevokeDesktopClientArg `json:"desktop_client,omitempty"`
	// Unlink a linked mobile device
	MobileClient *DeviceSessionArg `json:"mobile_client,omitempty"`
}

func (u *RevokeDeviceSessionArg) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// End an active session
		WebSession json.RawMessage `json:"web_session"`
		// Unlink a linked desktop device
		DesktopClient json.RawMessage `json:"desktop_client"`
		// Unlink a linked mobile device
		MobileClient json.RawMessage `json:"mobile_client"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "web_session":
		{
			if err := json.Unmarshal(body, &u.WebSession); err != nil {
				return err
			}
		}
	case "desktop_client":
		{
			if err := json.Unmarshal(body, &u.DesktopClient); err != nil {
				return err
			}
		}
	case "mobile_client":
		{
			if err := json.Unmarshal(body, &u.MobileClient); err != nil {
				return err
			}
		}
	}
	return nil
}

type RevokeDeviceSessionBatchArg struct {
	RevokeDevices []*RevokeDeviceSessionArg `json:"revoke_devices"`
}

func NewRevokeDeviceSessionBatchArg(RevokeDevices []*RevokeDeviceSessionArg) *RevokeDeviceSessionBatchArg {
	s := new(RevokeDeviceSessionBatchArg)
	s.RevokeDevices = RevokeDevices
	return s
}

type RevokeDeviceSessionBatchError struct {
	Tag string `json:".tag"`
}

type RevokeDeviceSessionBatchResult struct {
	RevokeDevicesStatus []*RevokeDeviceSessionStatus `json:"revoke_devices_status"`
}

func NewRevokeDeviceSessionBatchResult(RevokeDevicesStatus []*RevokeDeviceSessionStatus) *RevokeDeviceSessionBatchResult {
	s := new(RevokeDeviceSessionBatchResult)
	s.RevokeDevicesStatus = RevokeDevicesStatus
	return s
}

type RevokeDeviceSessionError struct {
	Tag string `json:".tag"`
}

type RevokeDeviceSessionStatus struct {
	// Result of the revoking request
	Success bool `json:"success"`
	// The error cause in case of a failure
	ErrorType *RevokeDeviceSessionError `json:"error_type,omitempty"`
}

func NewRevokeDeviceSessionStatus(Success bool) *RevokeDeviceSessionStatus {
	s := new(RevokeDeviceSessionStatus)
	s.Success = Success
	return s
}

type RevokeLinkedApiAppArg struct {
	// The application's unique id
	AppId string `json:"app_id"`
	// The unique id of the member owning the device
	TeamMemberId string `json:"team_member_id"`
	// Whether to keep the application dedicated folder (in case the application
	// uses  one)
	KeepAppFolder bool `json:"keep_app_folder"`
}

func NewRevokeLinkedApiAppArg(AppId string, TeamMemberId string) *RevokeLinkedApiAppArg {
	s := new(RevokeLinkedApiAppArg)
	s.AppId = AppId
	s.TeamMemberId = TeamMemberId
	s.KeepAppFolder = true
	return s
}

type RevokeLinkedApiAppBatchArg struct {
	RevokeLinkedApp []*RevokeLinkedApiAppArg `json:"revoke_linked_app"`
}

func NewRevokeLinkedApiAppBatchArg(RevokeLinkedApp []*RevokeLinkedApiAppArg) *RevokeLinkedApiAppBatchArg {
	s := new(RevokeLinkedApiAppBatchArg)
	s.RevokeLinkedApp = RevokeLinkedApp
	return s
}

// Error returned by `LinkedAppsRevokeLinkedAppBatch`.
type RevokeLinkedAppBatchError struct {
	Tag string `json:".tag"`
}

type RevokeLinkedAppBatchResult struct {
	RevokeLinkedAppStatus []*RevokeLinkedAppStatus `json:"revoke_linked_app_status"`
}

func NewRevokeLinkedAppBatchResult(RevokeLinkedAppStatus []*RevokeLinkedAppStatus) *RevokeLinkedAppBatchResult {
	s := new(RevokeLinkedAppBatchResult)
	s.RevokeLinkedAppStatus = RevokeLinkedAppStatus
	return s
}

// Error returned by `LinkedAppsRevokeLinkedApp`.
type RevokeLinkedAppError struct {
	Tag string `json:".tag"`
}

type RevokeLinkedAppStatus struct {
	// Result of the revoking request
	Success bool `json:"success"`
	// The error cause in case of a failure
	ErrorType *RevokeLinkedAppError `json:"error_type,omitempty"`
}

func NewRevokeLinkedAppStatus(Success bool) *RevokeLinkedAppStatus {
	s := new(RevokeLinkedAppStatus)
	s.Success = Success
	return s
}

// Policy governing which shared folders a team member can join.
type SharedFolderJoinPolicy struct {
	Tag string `json:".tag"`
}

// Policy governing who can be a member of a folder shared by a team member.
type SharedFolderMemberPolicy struct {
	Tag string `json:".tag"`
}

// Policy governing the visibility of newly created shared links.
type SharedLinkCreatePolicy struct {
	Tag string `json:".tag"`
}

// Describes the number of users in a specific storage bucket.
type StorageBucket struct {
	// The name of the storage bucket. For example, '1G' is a bucket of users with
	// storage size up to 1 Giga.
	Bucket string `json:"bucket"`
	// The number of people whose storage is in the range of this storage bucket.
	Users uint64 `json:"users"`
}

func NewStorageBucket(Bucket string, Users uint64) *StorageBucket {
	s := new(StorageBucket)
	s.Bucket = Bucket
	s.Users = Users
	return s
}

type TeamGetInfoResult struct {
	// The name of the team.
	Name string `json:"name"`
	// The ID of the team.
	TeamId string `json:"team_id"`
	// The number of licenses available to the team.
	NumLicensedUsers uint32 `json:"num_licensed_users"`
	// The number of accounts that have been invited or are already active members
	// of the team.
	NumProvisionedUsers uint32        `json:"num_provisioned_users"`
	Policies            *TeamPolicies `json:"policies"`
}

func NewTeamGetInfoResult(Name string, TeamId string, NumLicensedUsers uint32, NumProvisionedUsers uint32, Policies *TeamPolicies) *TeamGetInfoResult {
	s := new(TeamGetInfoResult)
	s.Name = Name
	s.TeamId = TeamId
	s.NumLicensedUsers = NumLicensedUsers
	s.NumProvisionedUsers = NumProvisionedUsers
	s.Policies = Policies
	return s
}

// Information about a team member.
type TeamMemberInfo struct {
	// Profile of a user as a member of a team.
	Profile *TeamMemberProfile `json:"profile"`
	// The user's role in the team.
	Role *AdminTier `json:"role"`
}

func NewTeamMemberInfo(Profile *TeamMemberProfile, Role *AdminTier) *TeamMemberInfo {
	s := new(TeamMemberInfo)
	s.Profile = Profile
	s.Role = Role
	return s
}

// Profile of a user as a member of a team.
type TeamMemberProfile struct {
	// ID of user as a member of a team.
	TeamMemberId string `json:"team_member_id"`
	// Email address of user.
	Email string `json:"email"`
	// Is true if the user's email is verified to be owned by the user.
	EmailVerified bool `json:"email_verified"`
	// The user's status as a member of a specific team.
	Status *TeamMemberStatus `json:"status"`
	// Representations for a person's name.
	Name *users.Name `json:"name"`
	// List of group IDs of groups that the user belongs to.
	Groups []string `json:"groups"`
	// External ID that a team can attach to the user. An application using the API
	// may find it easier to use their own IDs instead of Dropbox IDs like
	// account_id or team_member_id.
	ExternalId string `json:"external_id,omitempty"`
}

func NewTeamMemberProfile(TeamMemberId string, Email string, EmailVerified bool, Status *TeamMemberStatus, Name *users.Name, Groups []string) *TeamMemberProfile {
	s := new(TeamMemberProfile)
	s.TeamMemberId = TeamMemberId
	s.Email = Email
	s.EmailVerified = EmailVerified
	s.Status = Status
	s.Name = Name
	s.Groups = Groups
	return s
}

// The user's status as a member of a specific team.
type TeamMemberStatus struct {
	Tag string `json:".tag"`
}

// Policies governing team members.
type TeamPolicies struct {
	// Policies governing sharing.
	Sharing *TeamSharingPolicies `json:"sharing"`
	// This describes the Enterprise Mobility Management (EMM) state for this team.
	// This information can be used to understand if an organization is integrating
	// with a third-party EMM vendor to further manage and apply restrictions upon
	// the team's Dropbox usage on mobile devices. This is a new feature and in the
	// future we'll be adding more new fields and additional documentation.
	EmmState *EmmState `json:"emm_state"`
}

func NewTeamPolicies(Sharing *TeamSharingPolicies, EmmState *EmmState) *TeamPolicies {
	s := new(TeamPolicies)
	s.Sharing = Sharing
	s.EmmState = EmmState
	return s
}

// Policies governing sharing within and outside of the team.
type TeamSharingPolicies struct {
	// Who can join folders shared by team members.
	SharedFolderMemberPolicy *SharedFolderMemberPolicy `json:"shared_folder_member_policy"`
	// Which shared folders team members can join.
	SharedFolderJoinPolicy *SharedFolderJoinPolicy `json:"shared_folder_join_policy"`
	// What is the visibility of newly created shared links.
	SharedLinkCreatePolicy *SharedLinkCreatePolicy `json:"shared_link_create_policy"`
}

func NewTeamSharingPolicies(SharedFolderMemberPolicy *SharedFolderMemberPolicy, SharedFolderJoinPolicy *SharedFolderJoinPolicy, SharedLinkCreatePolicy *SharedLinkCreatePolicy) *TeamSharingPolicies {
	s := new(TeamSharingPolicies)
	s.SharedFolderMemberPolicy = SharedFolderMemberPolicy
	s.SharedFolderJoinPolicy = SharedFolderJoinPolicy
	s.SharedLinkCreatePolicy = SharedLinkCreatePolicy
	return s
}

// Argument for selecting a single user, either by team_member_id, external_id
// or email.
type UserSelectorArg struct {
	Tag          string `json:".tag"`
	TeamMemberId string `json:"team_member_id,omitempty"`
	ExternalId   string `json:"external_id,omitempty"`
	Email        string `json:"email,omitempty"`
}

func (u *UserSelectorArg) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag          string          `json:".tag"`
		TeamMemberId json.RawMessage `json:"team_member_id"`
		ExternalId   json.RawMessage `json:"external_id"`
		Email        json.RawMessage `json:"email"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "team_member_id":
		{
			if len(w.TeamMemberId) == 0 {
				break
			}
			if err := json.Unmarshal(w.TeamMemberId, &u.TeamMemberId); err != nil {
				return err
			}
		}
	case "external_id":
		{
			if len(w.ExternalId) == 0 {
				break
			}
			if err := json.Unmarshal(w.ExternalId, &u.ExternalId); err != nil {
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

// Argument for selecting a list of users, either by team_member_ids,
// external_ids or emails.
type UsersSelectorArg struct {
	Tag string `json:".tag"`
	// List of member IDs.
	TeamMemberIds []string `json:"team_member_ids,omitempty"`
	// List of external user IDs.
	ExternalIds []string `json:"external_ids,omitempty"`
	// List of email addresses.
	Emails []string `json:"emails,omitempty"`
}

func (u *UsersSelectorArg) UnmarshalJSON(body []byte) error {
	type wrap struct {
		Tag string `json:".tag"`
		// List of member IDs.
		TeamMemberIds json.RawMessage `json:"team_member_ids"`
		// List of external user IDs.
		ExternalIds json.RawMessage `json:"external_ids"`
		// List of email addresses.
		Emails json.RawMessage `json:"emails"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch w.Tag {
	case "team_member_ids":
		{
			if len(w.TeamMemberIds) == 0 {
				break
			}
			if err := json.Unmarshal(w.TeamMemberIds, &u.TeamMemberIds); err != nil {
				return err
			}
		}
	case "external_ids":
		{
			if len(w.ExternalIds) == 0 {
				break
			}
			if err := json.Unmarshal(w.ExternalIds, &u.ExternalIds); err != nil {
				return err
			}
		}
	case "emails":
		{
			if len(w.Emails) == 0 {
				break
			}
			if err := json.Unmarshal(w.Emails, &u.Emails); err != nil {
				return err
			}
		}
	}
	return nil
}

type Team interface {
	// List all device sessions of a team's member.
	DevicesListMemberDevices(arg *ListMemberDevicesArg) (res *ListMemberDevicesResult, err error)
	// List all device sessions of a team.
	DevicesListTeamDevices(arg *ListTeamDevicesArg) (res *ListTeamDevicesResult, err error)
	// Revoke a device session of a team's member
	DevicesRevokeDeviceSession(arg *RevokeDeviceSessionArg) (err error)
	// Revoke a list of device sessions of team members
	DevicesRevokeDeviceSessionBatch(arg *RevokeDeviceSessionBatchArg) (res *RevokeDeviceSessionBatchResult, err error)
	// Retrieves information about a team.
	GetInfo() (res *TeamGetInfoResult, err error)
	// Creates a new, empty group, with a requested name. Permission : Team member
	// management
	GroupsCreate(arg *GroupCreateArg) (res *GroupFullInfo, err error)
	// Deletes a group. The group is deleted immediately. However the revoking of
	// group-owned resources may take additional time. Use the `GroupsJobStatusGet`
	// to determine whether this process has completed. Permission : Team member
	// management
	GroupsDelete(arg *GroupSelector) (res *async.LaunchEmptyResult, err error)
	// Retrieves information about one or more groups. Permission : Team
	// Information
	GroupsGetInfo(arg *GroupsSelector) (res []*GroupsGetInfoItem, err error)
	// Once an async_job_id is returned from `GroupsDelete`, `GroupsMembersAdd` ,
	// or `GroupsMembersRemove` use this method to poll the status of
	// granting/revoking group members' access to group-owned resources. Permission
	// : Team member management
	GroupsJobStatusGet(arg *async.PollArg) (res *async.PollEmptyResult, err error)
	// Lists groups on a team. Permission : Team Information
	GroupsList(arg *GroupsListArg) (res *GroupsListResult, err error)
	// Once a cursor has been retrieved from `GroupsList`, use this to paginate
	// through all groups. Permission : Team information
	GroupsListContinue(arg *GroupsListContinueArg) (res *GroupsListResult, err error)
	// Adds members to a group. The members are added immediately. However the
	// granting of group-owned resources may take additional time. Use the
	// `GroupsJobStatusGet` to determine whether this process has completed.
	// Permission : Team member management
	GroupsMembersAdd(arg *GroupMembersAddArg) (res *GroupMembersChangeResult, err error)
	// Removes members from a group. The members are removed immediately. However
	// the revoking of group-owned resources may take additional time. Use the
	// `GroupsJobStatusGet` to determine whether this process has completed.
	// Permission : Team member management
	GroupsMembersRemove(arg *GroupMembersRemoveArg) (res *GroupMembersChangeResult, err error)
	// Sets a member's access type in a group. Permission : Team member management
	GroupsMembersSetAccessType(arg *GroupMembersSetAccessTypeArg) (res []*GroupsGetInfoItem, err error)
	// Updates a group's name and/or external ID. Permission : Team member
	// management
	GroupsUpdate(arg *GroupUpdateArgs) (res *GroupFullInfo, err error)
	// List all linked applications of the team member. Note, this endpoint doesn't
	// list any team-linked applications.
	LinkedAppsListMemberLinkedApps(arg *ListMemberAppsArg) (res *ListMemberAppsResult, err error)
	// List all applications linked to the team members' accounts. Note, this
	// endpoint doesn't list any team-linked applications.
	LinkedAppsListTeamLinkedApps(arg *ListTeamAppsArg) (res *ListTeamAppsResult, err error)
	// Revoke a linked application of the team member
	LinkedAppsRevokeLinkedApp(arg *RevokeLinkedApiAppArg) (err error)
	// Revoke a list of linked applications of the team members
	LinkedAppsRevokeLinkedAppBatch(arg *RevokeLinkedApiAppBatchArg) (res *RevokeLinkedAppBatchResult, err error)
	// Adds members to a team. Permission : Team member management A maximum of 20
	// members can be specified in a single call. If no Dropbox account exists with
	// the email address specified, a new Dropbox account will be created with the
	// given email address, and that account will be invited to the team. If a
	// personal Dropbox account exists with the email address specified in the
	// call, this call will create a placeholder Dropbox account for the user on
	// the team and send an email inviting the user to migrate their existing
	// personal account onto the team. Team member management apps are required to
	// set an initial given_name and surname for a user to use in the team
	// invitation and for 'Perform as team member' actions taken on the user before
	// they become 'active'.
	MembersAdd(arg *MembersAddArg) (res *MembersAddLaunch, err error)
	// Once an async_job_id is returned from `MembersAdd` , use this to poll the
	// status of the asynchronous request. Permission : Team member management
	MembersAddJobStatusGet(arg *async.PollArg) (res *MembersAddJobStatus, err error)
	// Returns information about multiple team members. Permission : Team
	// information This endpoint will return an empty member_info item, for IDs (or
	// emails) that cannot be matched to a valid team member.
	MembersGetInfo(arg *MembersGetInfoArgs) (res []*MembersGetInfoItem, err error)
	// Lists members of a team. Permission : Team information
	MembersList(arg *MembersListArg) (res *MembersListResult, err error)
	// Once a cursor has been retrieved from `MembersList`, use this to paginate
	// through all team members. Permission : Team information
	MembersListContinue(arg *MembersListContinueArg) (res *MembersListResult, err error)
	// Removes a member from a team. Permission : Team member management Exactly
	// one of team_member_id, email, or external_id must be provided to identify
	// the user account. This is not a deactivation where the account can be
	// re-activated again. Calling `MembersAdd` with the removed user's email
	// address will create a new account with a new team_member_id that will not
	// have access to any content that was shared with the initial account. This
	// endpoint may initiate an asynchronous job. To obtain the final result of the
	// job, the client should periodically poll `MembersRemoveJobStatusGet`.
	MembersRemove(arg *MembersRemoveArg) (res *async.LaunchEmptyResult, err error)
	// Once an async_job_id is returned from `MembersRemove` , use this to poll the
	// status of the asynchronous request. Permission : Team member management
	MembersRemoveJobStatusGet(arg *async.PollArg) (res *async.PollEmptyResult, err error)
	// Sends welcome email to pending team member. Permission : Team member
	// management Exactly one of team_member_id, email, or external_id must be
	// provided to identify the user account. No-op if team member is not pending.
	MembersSendWelcomeEmail(arg *UserSelectorArg) (err error)
	// Updates a team member's permissions. Permission : Team member management
	MembersSetAdminPermissions(arg *MembersSetPermissionsArg) (res *MembersSetPermissionsResult, err error)
	// Updates a team member's profile. Permission : Team member management
	MembersSetProfile(arg *MembersSetProfileArg) (res *TeamMemberInfo, err error)
	// Suspend a member from a team. Permission : Team member management Exactly
	// one of team_member_id, email, or external_id must be provided to identify
	// the user account.
	MembersSuspend(arg *MembersDeactivateArg) (err error)
	// Unsuspend a member from a team. Permission : Team member management Exactly
	// one of team_member_id, email, or external_id must be provided to identify
	// the user account.
	MembersUnsuspend(arg *MembersUnsuspendArg) (err error)
	// Retrieves reporting data about a team's user activity.
	ReportsGetActivity(arg *DateRange) (res *GetActivityReport, err error)
	// Retrieves reporting data about a team's linked devices.
	ReportsGetDevices(arg *DateRange) (res *GetDevicesReport, err error)
	// Retrieves reporting data about a team's membership.
	ReportsGetMembership(arg *DateRange) (res *GetMembershipReport, err error)
	// Retrieves reporting data about a team's storage usage.
	ReportsGetStorage(arg *DateRange) (res *GetStorageReport, err error)
}
