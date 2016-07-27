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

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/properties"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/team_common"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/team_policies"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/users"
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
	DeviceSession
	// Information on the hosting device
	UserAgent string `json:"user_agent"`
	// Information on the hosting operating system
	Os string `json:"os"`
	// Information on the browser used for this web session
	Browser string `json:"browser"`
}

func NewActiveWebSession(SessionId string, UserAgent string, Os string, Browser string) *ActiveWebSession {
	s := new(ActiveWebSession)
	s.SessionId = SessionId
	s.UserAgent = UserAgent
	s.Os = Os
	s.Browser = Browser
	return s
}

// Arguments for adding property templates.
type AddPropertyTemplateArg struct {
	properties.PropertyGroupTemplate
}

func NewAddPropertyTemplateArg(Name string, Description string, Fields []*properties.PropertyFieldTemplate) *AddPropertyTemplateArg {
	s := new(AddPropertyTemplateArg)
	s.Name = Name
	s.Description = Description
	s.Fields = Fields
	return s
}

type AddPropertyTemplateResult struct {
	// An identifier for property template added by `propertiesTemplateAdd`.
	TemplateId string `json:"template_id"`
}

func NewAddPropertyTemplateResult(TemplateId string) *AddPropertyTemplateResult {
	s := new(AddPropertyTemplateResult)
	s.TemplateId = TemplateId
	return s
}

// Describes which team-related admin permissions a user has.
type AdminTier struct {
	dropbox.Tagged
}

type GroupCreateArg struct {
	// Group name.
	GroupName string `json:"group_name"`
	// The creator of a team can associate an arbitrary external ID to the
	// group.
	GroupExternalId string `json:"group_external_id,omitempty"`
}

func NewGroupCreateArg(GroupName string) *GroupCreateArg {
	s := new(GroupCreateArg)
	s.GroupName = GroupName
	return s
}

type AlphaGroupCreateArg struct {
	GroupCreateArg
	// Whether the team can be managed by selected users, or only by team admins
	GroupManagementType *team_common.GroupManagementType `json:"group_management_type"`
}

func NewAlphaGroupCreateArg(GroupName string) *AlphaGroupCreateArg {
	s := new(AlphaGroupCreateArg)
	s.GroupName = GroupName
	s.GroupManagementType = &team_common.GroupManagementType{Tagged: dropbox.Tagged{"company_managed"}}
	return s
}

// Full description of a group.
type AlphaGroupFullInfo struct {
	team_common.AlphaGroupSummary
	// List of group members.
	Members []*GroupMemberInfo `json:"members,omitempty"`
	// The group creation time as a UTC timestamp in milliseconds since the Unix
	// epoch.
	Created uint64 `json:"created"`
}

func NewAlphaGroupFullInfo(GroupName string, GroupId string, GroupManagementType *team_common.GroupManagementType, Created uint64) *AlphaGroupFullInfo {
	s := new(AlphaGroupFullInfo)
	s.GroupName = GroupName
	s.GroupId = GroupId
	s.GroupManagementType = GroupManagementType
	s.Created = Created
	return s
}

type IncludeMembersArg struct {
	// Whether to return the list of members in the group.  Note that the
	// default value will cause all the group members  to be returned in the
	// response. This may take a long time for large groups.
	ReturnMembers bool `json:"return_members"`
}

func NewIncludeMembersArg() *IncludeMembersArg {
	s := new(IncludeMembersArg)
	s.ReturnMembers = true
	return s
}

type GroupUpdateArgs struct {
	IncludeMembersArg
	// Specify a group.
	Group *GroupSelector `json:"group"`
	// Optional argument. Set group name to this if provided.
	NewGroupName string `json:"new_group_name,omitempty"`
	// Optional argument. New group external ID. If the argument is None, the
	// group's external_id won't be updated. If the argument is empty string,
	// the group's external id will be cleared.
	NewGroupExternalId string `json:"new_group_external_id,omitempty"`
}

func NewGroupUpdateArgs(Group *GroupSelector) *GroupUpdateArgs {
	s := new(GroupUpdateArgs)
	s.Group = Group
	s.ReturnMembers = true
	return s
}

type AlphaGroupUpdateArgs struct {
	GroupUpdateArgs
	// Set new group management type, if provided.
	NewGroupManagementType *team_common.GroupManagementType `json:"new_group_management_type,omitempty"`
}

func NewAlphaGroupUpdateArgs(Group *GroupSelector) *AlphaGroupUpdateArgs {
	s := new(AlphaGroupUpdateArgs)
	s.Group = Group
	s.ReturnMembers = true
	return s
}

type AlphaGroupsGetInfoItem struct {
	dropbox.Tagged
	// An ID that was provided as a parameter to `alphaGroupsGetInfo`, and did
	// not match a corresponding group. The ID can be a group ID, or an external
	// ID, depending on how the method was called.
	IdNotFound string `json:"id_not_found,omitempty"`
	// Info about a group.
	GroupInfo *AlphaGroupFullInfo `json:"group_info,omitempty"`
}

func (u *AlphaGroupsGetInfoItem) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// Info about a group.
		GroupInfo json.RawMessage `json:"group_info,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "id_not_found":
		if err := json.Unmarshal(body, &u.IdNotFound); err != nil {
			return err
		}

	case "group_info":
		if err := json.Unmarshal(body, &u.GroupInfo); err != nil {
			return err
		}

	}
	return nil
}

type AlphaGroupsListResult struct {
	Groups []*team_common.AlphaGroupSummary `json:"groups"`
	// Pass the cursor into `alphaGroupsListContinue` to obtain the additional
	// groups.
	Cursor string `json:"cursor"`
	// Is true if there are additional groups that have not been returned yet.
	// An additional call to `alphaGroupsListContinue` can retrieve them.
	HasMore bool `json:"has_more"`
}

func NewAlphaGroupsListResult(Groups []*team_common.AlphaGroupSummary, Cursor string, HasMore bool) *AlphaGroupsListResult {
	s := new(AlphaGroupsListResult)
	s.Groups = Groups
	s.Cursor = Cursor
	s.HasMore = HasMore
	return s
}

// Information on linked third party applications
type ApiApp struct {
	// The application unique id
	AppId string `json:"app_id"`
	// The application name
	AppName string `json:"app_name"`
	// The application publisher name
	Publisher string `json:"publisher,omitempty"`
	// The publisher's URL
	PublisherUrl string `json:"publisher_url,omitempty"`
	// The time this application was linked
	Linked time.Time `json:"linked,omitempty"`
	// Whether the linked application uses a dedicated folder
	IsAppFolder bool `json:"is_app_folder"`
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
	dropbox.Tagged
}

// Information about linked Dropbox desktop client sessions
type DesktopClientSession struct {
	DeviceSession
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
	dropbox.Tagged
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

// Activity Report Result. Each of the items in the storage report is an array
// of values, one value per day. If there is no data for a day, then the value
// will be None.
type GetActivityReport struct {
	BaseDfbReport
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
	// Array of the number of shared folders with some activity in the last
	// week.
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
	// Array of the number of views by non-logged-in users to shared links
	// created by the team.
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
	BaseDfbReport
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
	BaseDfbReport
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
	BaseDfbReport
	// Sum of the shared, unshared, and datastore usages, for each day.
	TotalUsage []uint64 `json:"total_usage"`
	// Array of the combined size (bytes) of team members' shared folders, for
	// each day.
	SharedUsage []uint64 `json:"shared_usage"`
	// Array of the combined size (bytes) of team members' root namespaces, for
	// each day.
	UnsharedUsage []uint64 `json:"unshared_usage"`
	// Array of the number of shared folders owned by team members, for each
	// day.
	SharedFolders []uint64 `json:"shared_folders"`
	// Array of storage summaries of team members' account sizes. Each storage
	// summary is an array of key, value pairs, where each pair describes a
	// storage bucket. The key indicates the upper bound of the bucket and the
	// value is the number of users in that bucket. There is one such summary
	// per day. If there is no data for a day, the storage summary will be
	// empty.
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
	dropbox.Tagged
}

type GroupCreateError struct {
	dropbox.Tagged
}

// Error that can be raised when `GroupSelector` is used.
type GroupSelectorError struct {
	dropbox.Tagged
}

type GroupDeleteError struct {
	dropbox.Tagged
}

// Full description of a group.
type GroupFullInfo struct {
	team_common.GroupSummary
	// List of group members.
	Members []*GroupMemberInfo `json:"members,omitempty"`
	// The group creation time as a UTC timestamp in milliseconds since the Unix
	// epoch.
	Created uint64 `json:"created"`
}

func NewGroupFullInfo(GroupName string, GroupId string, Created uint64) *GroupFullInfo {
	s := new(GroupFullInfo)
	s.GroupName = GroupName
	s.GroupId = GroupId
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
	dropbox.Tagged
}

type GroupMemberSetAccessTypeError struct {
	dropbox.Tagged
}

type GroupMembersAddArg struct {
	IncludeMembersArg
	// Group to which users will be added.
	Group *GroupSelector `json:"group"`
	// List of users to be added to the group.
	Members []*MemberAccess `json:"members"`
}

func NewGroupMembersAddArg(Group *GroupSelector, Members []*MemberAccess) *GroupMembersAddArg {
	s := new(GroupMembersAddArg)
	s.Group = Group
	s.Members = Members
	s.ReturnMembers = true
	return s
}

type GroupMembersAddError struct {
	dropbox.Tagged
	// These members are not part of your team. Currently, you cannot add
	// members to a group if they are not part of your team, though this may
	// change in a subsequent version. To add new members to your Dropbox
	// Business team, use the `membersAdd` endpoint.
	MembersNotInTeam []string `json:"members_not_in_team,omitempty"`
	// These users were not found in Dropbox.
	UsersNotFound []string `json:"users_not_found,omitempty"`
	// A company-managed group cannot be managed by a user.
	UserCannotBeManagerOfCompanyManagedGroup []string `json:"user_cannot_be_manager_of_company_managed_group,omitempty"`
}

func (u *GroupMembersAddError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// These members are not part of your team. Currently, you cannot add
		// members to a group if they are not part of your team, though this may
		// change in a subsequent version. To add new members to your Dropbox
		// Business team, use the `membersAdd` endpoint.
		MembersNotInTeam json.RawMessage `json:"members_not_in_team,omitempty"`
		// These users were not found in Dropbox.
		UsersNotFound json.RawMessage `json:"users_not_found,omitempty"`
		// A company-managed group cannot be managed by a user.
		UserCannotBeManagerOfCompanyManagedGroup json.RawMessage `json:"user_cannot_be_manager_of_company_managed_group,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "members_not_in_team":
		if err := json.Unmarshal(body, &u.MembersNotInTeam); err != nil {
			return err
		}

	case "users_not_found":
		if err := json.Unmarshal(body, &u.UsersNotFound); err != nil {
			return err
		}

	case "user_cannot_be_manager_of_company_managed_group":
		if err := json.Unmarshal(body, &u.UserCannotBeManagerOfCompanyManagedGroup); err != nil {
			return err
		}

	}
	return nil
}

// Result returned by `groupsMembersAdd` and `groupsMembersRemove`.
type GroupMembersChangeResult struct {
	// The group info after member change operation has been performed.
	GroupInfo *GroupFullInfo `json:"group_info"`
	// An ID that can be used to obtain the status of granting/revoking
	// group-owned resources.
	AsyncJobId string `json:"async_job_id"`
}

func NewGroupMembersChangeResult(GroupInfo *GroupFullInfo, AsyncJobId string) *GroupMembersChangeResult {
	s := new(GroupMembersChangeResult)
	s.GroupInfo = GroupInfo
	s.AsyncJobId = AsyncJobId
	return s
}

type GroupMembersRemoveArg struct {
	IncludeMembersArg
	// Group from which users will be removed.
	Group *GroupSelector `json:"group"`
	// List of users to be removed from the group.
	Users []*UserSelectorArg `json:"users"`
}

func NewGroupMembersRemoveArg(Group *GroupSelector, Users []*UserSelectorArg) *GroupMembersRemoveArg {
	s := new(GroupMembersRemoveArg)
	s.Group = Group
	s.Users = Users
	s.ReturnMembers = true
	return s
}

// Error that can be raised when `GroupMembersSelector` is used, and the users
// are required to be members of the specified group.
type GroupMembersSelectorError struct {
	dropbox.Tagged
}

type GroupMembersRemoveError struct {
	dropbox.Tagged
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
	GroupMemberSelector
	// New group access type the user will have.
	AccessType *GroupAccessType `json:"access_type"`
	// Whether to return the list of members in the group.  Note that the
	// default value will cause all the group members  to be returned in the
	// response. This may take a long time for large groups.
	ReturnMembers bool `json:"return_members"`
}

func NewGroupMembersSetAccessTypeArg(Group *GroupSelector, User *UserSelectorArg, AccessType *GroupAccessType) *GroupMembersSetAccessTypeArg {
	s := new(GroupMembersSetAccessTypeArg)
	s.Group = Group
	s.User = User
	s.AccessType = AccessType
	s.ReturnMembers = true
	return s
}

// Argument for selecting a single group, either by group_id or by external
// group ID.
type GroupSelector struct {
	dropbox.Tagged
	// Group ID.
	GroupId string `json:"group_id,omitempty"`
	// External ID of the group.
	GroupExternalId string `json:"group_external_id,omitempty"`
}

func (u *GroupSelector) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "group_id":
		if err := json.Unmarshal(body, &u.GroupId); err != nil {
			return err
		}

	case "group_external_id":
		if err := json.Unmarshal(body, &u.GroupExternalId); err != nil {
			return err
		}

	}
	return nil
}

type GroupUpdateError struct {
	dropbox.Tagged
}

type GroupsGetInfoError struct {
	dropbox.Tagged
}

type GroupsGetInfoItem struct {
	dropbox.Tagged
	// An ID that was provided as a parameter to `groupsGetInfo`, and did not
	// match a corresponding group. The ID can be a group ID, or an external ID,
	// depending on how the method was called.
	IdNotFound string `json:"id_not_found,omitempty"`
	// Info about a group.
	GroupInfo *GroupFullInfo `json:"group_info,omitempty"`
}

func (u *GroupsGetInfoItem) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// Info about a group.
		GroupInfo json.RawMessage `json:"group_info,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "id_not_found":
		if err := json.Unmarshal(body, &u.IdNotFound); err != nil {
			return err
		}

	case "group_info":
		if err := json.Unmarshal(body, &u.GroupInfo); err != nil {
			return err
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
	dropbox.Tagged
}

type GroupsListResult struct {
	Groups []*team_common.GroupSummary `json:"groups"`
	// Pass the cursor into `groupsListContinue` to obtain the additional
	// groups.
	Cursor string `json:"cursor"`
	// Is true if there are additional groups that have not been returned yet.
	// An additional call to `groupsListContinue` can retrieve them.
	HasMore bool `json:"has_more"`
}

func NewGroupsListResult(Groups []*team_common.GroupSummary, Cursor string, HasMore bool) *GroupsListResult {
	s := new(GroupsListResult)
	s.Groups = Groups
	s.Cursor = Cursor
	s.HasMore = HasMore
	return s
}

type GroupsMembersListArg struct {
	// The group whose members are to be listed.
	Group *GroupSelector `json:"group"`
	// Number of results to return per call.
	Limit uint32 `json:"limit"`
}

func NewGroupsMembersListArg(Group *GroupSelector) *GroupsMembersListArg {
	s := new(GroupsMembersListArg)
	s.Group = Group
	s.Limit = 1000
	return s
}

type GroupsMembersListContinueArg struct {
	// Indicates from what point to get the next set of groups.
	Cursor string `json:"cursor"`
}

func NewGroupsMembersListContinueArg(Cursor string) *GroupsMembersListContinueArg {
	s := new(GroupsMembersListContinueArg)
	s.Cursor = Cursor
	return s
}

type GroupsMembersListContinueError struct {
	dropbox.Tagged
}

type GroupsMembersListResult struct {
	Members []*GroupMemberInfo `json:"members"`
	// Pass the cursor into `groupsMembersListContinue` to obtain additional
	// group members.
	Cursor string `json:"cursor"`
	// Is true if there are additional group members that have not been returned
	// yet. An additional call to `groupsMembersListContinue` can retrieve them.
	HasMore bool `json:"has_more"`
}

func NewGroupsMembersListResult(Members []*GroupMemberInfo, Cursor string, HasMore bool) *GroupsMembersListResult {
	s := new(GroupsMembersListResult)
	s.Members = Members
	s.Cursor = Cursor
	s.HasMore = HasMore
	return s
}

type GroupsPollError struct {
	dropbox.Tagged
}

// Argument for selecting a list of groups, either by group_ids, or external
// group IDs.
type GroupsSelector struct {
	dropbox.Tagged
	// List of group IDs.
	GroupIds []string `json:"group_ids,omitempty"`
	// List of external IDs of groups.
	GroupExternalIds []string `json:"group_external_ids,omitempty"`
}

func (u *GroupsSelector) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// List of group IDs.
		GroupIds json.RawMessage `json:"group_ids,omitempty"`
		// List of external IDs of groups.
		GroupExternalIds json.RawMessage `json:"group_external_ids,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "group_ids":
		if err := json.Unmarshal(body, &u.GroupIds); err != nil {
			return err
		}

	case "group_external_ids":
		if err := json.Unmarshal(body, &u.GroupExternalIds); err != nil {
			return err
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

// Error returned by `linkedAppsListMemberLinkedApps`.
type ListMemberAppsError struct {
	dropbox.Tagged
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
	dropbox.Tagged
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

// Arguments for `linkedAppsListMembersLinkedApps`.
type ListMembersAppsArg struct {
	// At the first call to the `linkedAppsListMembersLinkedApps` the cursor
	// shouldn't be passed. Then, if the result of the call includes a cursor,
	// the following requests should include the received cursors in order to
	// receive the next sub list of the team applications
	Cursor string `json:"cursor,omitempty"`
}

func NewListMembersAppsArg() *ListMembersAppsArg {
	s := new(ListMembersAppsArg)
	return s
}

// Error returned by `linkedAppsListMembersLinkedApps`
type ListMembersAppsError struct {
	dropbox.Tagged
}

// Information returned by `linkedAppsListMembersLinkedApps`.
type ListMembersAppsResult struct {
	// The linked applications of each member of the team
	Apps []*MemberLinkedApps `json:"apps"`
	// If true, then there are more apps available. Pass the cursor to
	// `linkedAppsListMembersLinkedApps` to retrieve the rest.
	HasMore bool `json:"has_more"`
	// Pass the cursor into `linkedAppsListMembersLinkedApps` to receive the
	// next sub list of team's applications.
	Cursor string `json:"cursor,omitempty"`
}

func NewListMembersAppsResult(Apps []*MemberLinkedApps, HasMore bool) *ListMembersAppsResult {
	s := new(ListMembersAppsResult)
	s.Apps = Apps
	s.HasMore = HasMore
	return s
}

type ListMembersDevicesArg struct {
	// At the first call to the `devicesListMembersDevices` the cursor shouldn't
	// be passed. Then, if the result of the call includes a cursor, the
	// following requests should include the received cursors in order to
	// receive the next sub list of team devices
	Cursor string `json:"cursor,omitempty"`
	// Whether to list web sessions of the team members
	IncludeWebSessions bool `json:"include_web_sessions"`
	// Whether to list desktop clients of the team members
	IncludeDesktopClients bool `json:"include_desktop_clients"`
	// Whether to list mobile clients of the team members
	IncludeMobileClients bool `json:"include_mobile_clients"`
}

func NewListMembersDevicesArg() *ListMembersDevicesArg {
	s := new(ListMembersDevicesArg)
	s.IncludeWebSessions = true
	s.IncludeDesktopClients = true
	s.IncludeMobileClients = true
	return s
}

type ListMembersDevicesError struct {
	dropbox.Tagged
}

type ListMembersDevicesResult struct {
	// The devices of each member of the team
	Devices []*MemberDevices `json:"devices"`
	// If true, then there are more devices available. Pass the cursor to
	// `devicesListMembersDevices` to retrieve the rest.
	HasMore bool `json:"has_more"`
	// Pass the cursor into `devicesListMembersDevices` to receive the next sub
	// list of team's devices.
	Cursor string `json:"cursor,omitempty"`
}

func NewListMembersDevicesResult(Devices []*MemberDevices, HasMore bool) *ListMembersDevicesResult {
	s := new(ListMembersDevicesResult)
	s.Devices = Devices
	s.HasMore = HasMore
	return s
}

// Arguments for `linkedAppsListTeamLinkedApps`.
type ListTeamAppsArg struct {
	// At the first call to the `linkedAppsListTeamLinkedApps` the cursor
	// shouldn't be passed. Then, if the result of the call includes a cursor,
	// the following requests should include the received cursors in order to
	// receive the next sub list of the team applications
	Cursor string `json:"cursor,omitempty"`
}

func NewListTeamAppsArg() *ListTeamAppsArg {
	s := new(ListTeamAppsArg)
	return s
}

// Error returned by `linkedAppsListTeamLinkedApps`
type ListTeamAppsError struct {
	dropbox.Tagged
}

// Information returned by `linkedAppsListTeamLinkedApps`.
type ListTeamAppsResult struct {
	// The linked applications of each member of the team
	Apps []*MemberLinkedApps `json:"apps"`
	// If true, then there are more apps available. Pass the cursor to
	// `linkedAppsListTeamLinkedApps` to retrieve the rest.
	HasMore bool `json:"has_more"`
	// Pass the cursor into `linkedAppsListTeamLinkedApps` to receive the next
	// sub list of team's applications.
	Cursor string `json:"cursor,omitempty"`
}

func NewListTeamAppsResult(Apps []*MemberLinkedApps, HasMore bool) *ListTeamAppsResult {
	s := new(ListTeamAppsResult)
	s.Apps = Apps
	s.HasMore = HasMore
	return s
}

type ListTeamDevicesArg struct {
	// At the first call to the `devicesListTeamDevices` the cursor shouldn't be
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
	dropbox.Tagged
}

type ListTeamDevicesResult struct {
	// The devices of each member of the team
	Devices []*MemberDevices `json:"devices"`
	// If true, then there are more devices available. Pass the cursor to
	// `devicesListTeamDevices` to retrieve the rest.
	HasMore bool `json:"has_more"`
	// Pass the cursor into `devicesListTeamDevices` to receive the next sub
	// list of team's devices.
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
	// false, no email invitation will be sent to the user. This may be useful
	// for apps using single sign-on (SSO) flows for onboarding that want to
	// handle announcements themselves.
	SendWelcomeEmail bool       `json:"send_welcome_email"`
	Role             *AdminTier `json:"role"`
}

func NewMemberAddArg(MemberEmail string, MemberGivenName string, MemberSurname string) *MemberAddArg {
	s := new(MemberAddArg)
	s.MemberEmail = MemberEmail
	s.MemberGivenName = MemberGivenName
	s.MemberSurname = MemberSurname
	s.SendWelcomeEmail = true
	s.Role = &AdminTier{Tagged: dropbox.Tagged{"member_only"}}
	return s
}

// Describes the result of attempting to add a single user to the team.
// 'success' is the only value indicating that a user was indeed added to the
// team - the other values explain the type of failure that occurred, and
// include the email of the user for which the operation has failed.
type MemberAddResult struct {
	dropbox.Tagged
	// Describes a user that was successfully added to the team.
	Success *TeamMemberInfo `json:"success,omitempty"`
	// Team is already full. The organization has no available licenses.
	TeamLicenseLimit string `json:"team_license_limit,omitempty"`
	// Team is already full. The free team member limit has been reached.
	FreeTeamMemberLimitReached string `json:"free_team_member_limit_reached,omitempty"`
	// User is already on this team. The provided email address is associated
	// with a user who is already a member of or invited to the team.
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
		dropbox.Tagged
		// Describes a user that was successfully added to the team.
		Success json.RawMessage `json:"success,omitempty"`
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

	case "team_license_limit":
		if err := json.Unmarshal(body, &u.TeamLicenseLimit); err != nil {
			return err
		}

	case "free_team_member_limit_reached":
		if err := json.Unmarshal(body, &u.FreeTeamMemberLimitReached); err != nil {
			return err
		}

	case "user_already_on_team":
		if err := json.Unmarshal(body, &u.UserAlreadyOnTeam); err != nil {
			return err
		}

	case "user_on_another_team":
		if err := json.Unmarshal(body, &u.UserOnAnotherTeam); err != nil {
			return err
		}

	case "user_already_paired":
		if err := json.Unmarshal(body, &u.UserAlreadyPaired); err != nil {
			return err
		}

	case "user_migration_failed":
		if err := json.Unmarshal(body, &u.UserMigrationFailed); err != nil {
			return err
		}

	case "duplicate_external_member_id":
		if err := json.Unmarshal(body, &u.DuplicateExternalMemberId); err != nil {
			return err
		}

	case "user_creation_failed":
		if err := json.Unmarshal(body, &u.UserCreationFailed); err != nil {
			return err
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
	// External ID that a team can attach to the user. An application using the
	// API may find it easier to use their own IDs instead of Dropbox IDs like
	// account_id or team_member_id.
	ExternalId string `json:"external_id,omitempty"`
	// A user's account identifier.
	AccountId string `json:"account_id,omitempty"`
	// Email address of user.
	Email string `json:"email"`
	// Is true if the user's email is verified to be owned by the user.
	EmailVerified bool `json:"email_verified"`
	// The user's status as a member of a specific team.
	Status *TeamMemberStatus `json:"status"`
	// Representations for a person's name.
	Name *users.Name `json:"name"`
	// The user's membership type: full (normal team member) vs limited (does
	// not use a license; no access to the team's shared quota).
	MembershipType *TeamMembershipType `json:"membership_type"`
}

func NewMemberProfile(TeamMemberId string, Email string, EmailVerified bool, Status *TeamMemberStatus, Name *users.Name, MembershipType *TeamMembershipType) *MemberProfile {
	s := new(MemberProfile)
	s.TeamMemberId = TeamMemberId
	s.Email = Email
	s.EmailVerified = EmailVerified
	s.Status = Status
	s.Name = Name
	s.MembershipType = MembershipType
	return s
}

// Error that can be returned whenever a struct derived from `UserSelectorArg`
// is used.
type UserSelectorError struct {
	dropbox.Tagged
}

type MemberSelectorError struct {
	dropbox.Tagged
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
	dropbox.Tagged
	// The asynchronous job has finished. For each member that was specified in
	// the parameter `MembersAddArg` that was provided to `membersAdd`, a
	// corresponding item is returned in this list.
	Complete []*MemberAddResult `json:"complete,omitempty"`
	// The asynchronous job returned an error. The string contains an error
	// message.
	Failed string `json:"failed,omitempty"`
}

func (u *MembersAddJobStatus) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// The asynchronous job has finished. For each member that was specified
		// in the parameter `MembersAddArg` that was provided to `membersAdd`, a
		// corresponding item is returned in this list.
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

	case "failed":
		if err := json.Unmarshal(body, &u.Failed); err != nil {
			return err
		}

	}
	return nil
}

type MembersAddLaunch struct {
	dropbox.Tagged
	Complete []*MemberAddResult `json:"complete,omitempty"`
}

func (u *MembersAddLaunch) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
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
	dropbox.Tagged
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
	dropbox.Tagged
}

// Describes a result obtained for a single user whose id was specified in the
// parameter of `membersGetInfo`.
type MembersGetInfoItem struct {
	dropbox.Tagged
	// An ID that was provided as a parameter to `membersGetInfo`, and did not
	// match a corresponding user. This might be a team_member_id, an email, or
	// an external ID, depending on how the method was called.
	IdNotFound string `json:"id_not_found,omitempty"`
	// Info about a team member.
	MemberInfo *TeamMemberInfo `json:"member_info,omitempty"`
}

func (u *MembersGetInfoItem) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// Info about a team member.
		MemberInfo json.RawMessage `json:"member_info,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "id_not_found":
		if err := json.Unmarshal(body, &u.IdNotFound); err != nil {
			return err
		}

	case "member_info":
		if err := json.Unmarshal(body, &u.MemberInfo); err != nil {
			return err
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
	dropbox.Tagged
}

type MembersListError struct {
	dropbox.Tagged
}

type MembersListResult struct {
	// List of team members.
	Members []*TeamMemberInfo `json:"members"`
	// Pass the cursor into `membersListContinue` to obtain the additional
	// members.
	Cursor string `json:"cursor"`
	// Is true if there are additional team members that have not been returned
	// yet. An additional call to `membersListContinue` can retrieve them.
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
	MembersDeactivateArg
	// If provided, files from the deleted member account will be transferred to
	// this user.
	TransferDestId *UserSelectorArg `json:"transfer_dest_id,omitempty"`
	// If provided, errors during the transfer process will be sent via email to
	// this user. If the transfer_dest_id argument was provided, then this
	// argument must be provided as well.
	TransferAdminId *UserSelectorArg `json:"transfer_admin_id,omitempty"`
	// Downgrade the member to a Basic account. The user will retain the email
	// address associated with their Dropbox  account and data in their account
	// that is not restricted to team members.
	KeepAccount bool `json:"keep_account"`
}

func NewMembersRemoveArg(User *UserSelectorArg) *MembersRemoveArg {
	s := new(MembersRemoveArg)
	s.User = User
	s.WipeData = true
	s.KeepAccount = false
	return s
}

type MembersRemoveError struct {
	dropbox.Tagged
}

type MembersSendWelcomeError struct {
	dropbox.Tagged
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
	dropbox.Tagged
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
	dropbox.Tagged
}

type MembersSuspendError struct {
	dropbox.Tagged
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
	dropbox.Tagged
}

type MobileClientPlatform struct {
	dropbox.Tagged
}

// Information about linked Dropbox mobile client sessions
type MobileClientSession struct {
	DeviceSession
	// The device name
	DeviceName string `json:"device_name"`
	// The mobile application type
	ClientType *MobileClientPlatform `json:"client_type"`
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
	DeviceSessionArg
	// Whether to delete all files of the account (this is possible only if
	// supported by the desktop client and  will be made the next time the
	// client access the account)
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
	dropbox.Tagged
	// End an active session
	WebSession *DeviceSessionArg `json:"web_session,omitempty"`
	// Unlink a linked desktop device
	DesktopClient *RevokeDesktopClientArg `json:"desktop_client,omitempty"`
	// Unlink a linked mobile device
	MobileClient *DeviceSessionArg `json:"mobile_client,omitempty"`
}

func (u *RevokeDeviceSessionArg) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// End an active session
		WebSession json.RawMessage `json:"web_session,omitempty"`
		// Unlink a linked desktop device
		DesktopClient json.RawMessage `json:"desktop_client,omitempty"`
		// Unlink a linked mobile device
		MobileClient json.RawMessage `json:"mobile_client,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "web_session":
		if err := json.Unmarshal(body, &u.WebSession); err != nil {
			return err
		}

	case "desktop_client":
		if err := json.Unmarshal(body, &u.DesktopClient); err != nil {
			return err
		}

	case "mobile_client":
		if err := json.Unmarshal(body, &u.MobileClient); err != nil {
			return err
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
	dropbox.Tagged
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
	dropbox.Tagged
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

// Error returned by `linkedAppsRevokeLinkedAppBatch`.
type RevokeLinkedAppBatchError struct {
	dropbox.Tagged
}

type RevokeLinkedAppBatchResult struct {
	RevokeLinkedAppStatus []*RevokeLinkedAppStatus `json:"revoke_linked_app_status"`
}

func NewRevokeLinkedAppBatchResult(RevokeLinkedAppStatus []*RevokeLinkedAppStatus) *RevokeLinkedAppBatchResult {
	s := new(RevokeLinkedAppBatchResult)
	s.RevokeLinkedAppStatus = RevokeLinkedAppStatus
	return s
}

// Error returned by `linkedAppsRevokeLinkedApp`.
type RevokeLinkedAppError struct {
	dropbox.Tagged
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

// Describes the number of users in a specific storage bucket.
type StorageBucket struct {
	// The name of the storage bucket. For example, '1G' is a bucket of users
	// with storage size up to 1 Giga.
	Bucket string `json:"bucket"`
	// The number of people whose storage is in the range of this storage
	// bucket.
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
	// The number of accounts that have been invited or are already active
	// members of the team.
	NumProvisionedUsers uint32                            `json:"num_provisioned_users"`
	Policies            *team_policies.TeamMemberPolicies `json:"policies"`
}

func NewTeamGetInfoResult(Name string, TeamId string, NumLicensedUsers uint32, NumProvisionedUsers uint32, Policies *team_policies.TeamMemberPolicies) *TeamGetInfoResult {
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
	MemberProfile
	// List of group IDs of groups that the user belongs to.
	Groups []string `json:"groups"`
}

func NewTeamMemberProfile(TeamMemberId string, Email string, EmailVerified bool, Status *TeamMemberStatus, Name *users.Name, MembershipType *TeamMembershipType, Groups []string) *TeamMemberProfile {
	s := new(TeamMemberProfile)
	s.TeamMemberId = TeamMemberId
	s.Email = Email
	s.EmailVerified = EmailVerified
	s.Status = Status
	s.Name = Name
	s.MembershipType = MembershipType
	s.Groups = Groups
	return s
}

// The user's status as a member of a specific team.
type TeamMemberStatus struct {
	dropbox.Tagged
}

type TeamMembershipType struct {
	dropbox.Tagged
}

type UpdatePropertyTemplateArg struct {
	// An identifier for property template added by `propertiesTemplateAdd`.
	TemplateId string `json:"template_id"`
	// A display name for the property template. Property template names can be
	// up to 256 bytes.
	Name string `json:"name,omitempty"`
	// Description for new property template. Property template descriptions can
	// be up to 1024 bytes.
	Description string `json:"description,omitempty"`
	// This is a list of custom properties to add to the property template.
	// There can be up to 64 properties in a single property template.
	AddFields []*properties.PropertyFieldTemplate `json:"add_fields,omitempty"`
}

func NewUpdatePropertyTemplateArg(TemplateId string) *UpdatePropertyTemplateArg {
	s := new(UpdatePropertyTemplateArg)
	s.TemplateId = TemplateId
	return s
}

type UpdatePropertyTemplateResult struct {
	// An identifier for property template added by `propertiesTemplateAdd`.
	TemplateId string `json:"template_id"`
}

func NewUpdatePropertyTemplateResult(TemplateId string) *UpdatePropertyTemplateResult {
	s := new(UpdatePropertyTemplateResult)
	s.TemplateId = TemplateId
	return s
}

// Argument for selecting a single user, either by team_member_id, external_id
// or email.
type UserSelectorArg struct {
	dropbox.Tagged
	TeamMemberId string `json:"team_member_id,omitempty"`
	ExternalId   string `json:"external_id,omitempty"`
	Email        string `json:"email,omitempty"`
}

func (u *UserSelectorArg) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "team_member_id":
		if err := json.Unmarshal(body, &u.TeamMemberId); err != nil {
			return err
		}

	case "external_id":
		if err := json.Unmarshal(body, &u.ExternalId); err != nil {
			return err
		}

	case "email":
		if err := json.Unmarshal(body, &u.Email); err != nil {
			return err
		}

	}
	return nil
}

// Argument for selecting a list of users, either by team_member_ids,
// external_ids or emails.
type UsersSelectorArg struct {
	dropbox.Tagged
	// List of member IDs.
	TeamMemberIds []string `json:"team_member_ids,omitempty"`
	// List of external user IDs.
	ExternalIds []string `json:"external_ids,omitempty"`
	// List of email addresses.
	Emails []string `json:"emails,omitempty"`
}

func (u *UsersSelectorArg) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// List of member IDs.
		TeamMemberIds json.RawMessage `json:"team_member_ids,omitempty"`
		// List of external user IDs.
		ExternalIds json.RawMessage `json:"external_ids,omitempty"`
		// List of email addresses.
		Emails json.RawMessage `json:"emails,omitempty"`
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "team_member_ids":
		if err := json.Unmarshal(body, &u.TeamMemberIds); err != nil {
			return err
		}

	case "external_ids":
		if err := json.Unmarshal(body, &u.ExternalIds); err != nil {
			return err
		}

	case "emails":
		if err := json.Unmarshal(body, &u.Emails); err != nil {
			return err
		}

	}
	return nil
}
