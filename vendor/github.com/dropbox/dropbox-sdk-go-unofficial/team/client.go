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
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	dropbox "github.com/dropbox/dropbox-sdk-go-unofficial"
	"github.com/dropbox/dropbox-sdk-go-unofficial/async"
	"github.com/dropbox/dropbox-sdk-go-unofficial/properties"
)

type Client interface {
	// Creates a new, empty group, with a requested name. Permission : Team
	// member management
	AlphaGroupsCreate(arg *AlphaGroupCreateArg) (res *AlphaGroupFullInfo, err error)
	// Retrieves information about one or more groups. Permission : Team
	// Information
	AlphaGroupsGetInfo(arg *GroupsSelector) (res []*AlphaGroupsGetInfoItem, err error)
	// Lists groups on a team. Permission : Team Information
	AlphaGroupsList(arg *GroupsListArg) (res *AlphaGroupsListResult, err error)
	// Once a cursor has been retrieved from `alphaGroupsList`, use this to
	// paginate through all groups. Permission : Team information
	AlphaGroupsListContinue(arg *GroupsListContinueArg) (res *AlphaGroupsListResult, err error)
	// Updates a group's name, external ID or management type. Permission : Team
	// member management
	AlphaGroupsUpdate(arg *AlphaGroupUpdateArgs) (res *AlphaGroupFullInfo, err error)
	// List all device sessions of a team's member.
	DevicesListMemberDevices(arg *ListMemberDevicesArg) (res *ListMemberDevicesResult, err error)
	// List all device sessions of a team.
	DevicesListMembersDevices(arg *ListMembersDevicesArg) (res *ListMembersDevicesResult, err error)
	// List all device sessions of a team.
	DevicesListTeamDevices(arg *ListTeamDevicesArg) (res *ListTeamDevicesResult, err error)
	// Revoke a device session of a team's member
	DevicesRevokeDeviceSession(arg *RevokeDeviceSessionArg) (err error)
	// Revoke a list of device sessions of team members
	DevicesRevokeDeviceSessionBatch(arg *RevokeDeviceSessionBatchArg) (res *RevokeDeviceSessionBatchResult, err error)
	// Retrieves information about a team.
	GetInfo() (res *TeamGetInfoResult, err error)
	// Creates a new, empty group, with a requested name. Permission : Team
	// member management
	GroupsCreate(arg *GroupCreateArg) (res *GroupFullInfo, err error)
	// Deletes a group. The group is deleted immediately. However the revoking
	// of group-owned resources may take additional time. Use the
	// `groupsJobStatusGet` to determine whether this process has completed.
	// Permission : Team member management
	GroupsDelete(arg *GroupSelector) (res *async.LaunchEmptyResult, err error)
	// Retrieves information about one or more groups. Permission : Team
	// Information
	GroupsGetInfo(arg *GroupsSelector) (res []*GroupsGetInfoItem, err error)
	// Once an async_job_id is returned from `groupsDelete`, `groupsMembersAdd`
	// , or `groupsMembersRemove` use this method to poll the status of
	// granting/revoking group members' access to group-owned resources.
	// Permission : Team member management
	GroupsJobStatusGet(arg *async.PollArg) (res *async.PollEmptyResult, err error)
	// Lists groups on a team. Permission : Team Information
	GroupsList(arg *GroupsListArg) (res *GroupsListResult, err error)
	// Once a cursor has been retrieved from `groupsList`, use this to paginate
	// through all groups. Permission : Team information
	GroupsListContinue(arg *GroupsListContinueArg) (res *GroupsListResult, err error)
	// Adds members to a group. The members are added immediately. However the
	// granting of group-owned resources may take additional time. Use the
	// `groupsJobStatusGet` to determine whether this process has completed.
	// Permission : Team member management
	GroupsMembersAdd(arg *GroupMembersAddArg) (res *GroupMembersChangeResult, err error)
	// Lists members of a group. Permission : Team Information
	GroupsMembersList(arg *GroupsMembersListArg) (res *GroupsMembersListResult, err error)
	// Once a cursor has been retrieved from `groupsMembersList`, use this to
	// paginate through all members of the group. Permission : Team information
	GroupsMembersListContinue(arg *GroupsMembersListContinueArg) (res *GroupsMembersListResult, err error)
	// Removes members from a group. The members are removed immediately.
	// However the revoking of group-owned resources may take additional time.
	// Use the `groupsJobStatusGet` to determine whether this process has
	// completed. Permission : Team member management
	GroupsMembersRemove(arg *GroupMembersRemoveArg) (res *GroupMembersChangeResult, err error)
	// Sets a member's access type in a group. Permission : Team member
	// management
	GroupsMembersSetAccessType(arg *GroupMembersSetAccessTypeArg) (res []*GroupsGetInfoItem, err error)
	// Updates a group's name and/or external ID. Permission : Team member
	// management
	GroupsUpdate(arg *GroupUpdateArgs) (res *GroupFullInfo, err error)
	// List all linked applications of the team member. Note, this endpoint does
	// not list any team-linked applications.
	LinkedAppsListMemberLinkedApps(arg *ListMemberAppsArg) (res *ListMemberAppsResult, err error)
	// List all applications linked to the team members' accounts. Note, this
	// endpoint does not list any team-linked applications.
	LinkedAppsListMembersLinkedApps(arg *ListMembersAppsArg) (res *ListMembersAppsResult, err error)
	// List all applications linked to the team members' accounts. Note, this
	// endpoint doesn't list any team-linked applications.
	LinkedAppsListTeamLinkedApps(arg *ListTeamAppsArg) (res *ListTeamAppsResult, err error)
	// Revoke a linked application of the team member
	LinkedAppsRevokeLinkedApp(arg *RevokeLinkedApiAppArg) (err error)
	// Revoke a list of linked applications of the team members
	LinkedAppsRevokeLinkedAppBatch(arg *RevokeLinkedApiAppBatchArg) (res *RevokeLinkedAppBatchResult, err error)
	// Adds members to a team. Permission : Team member management A maximum of
	// 20 members can be specified in a single call. If no Dropbox account
	// exists with the email address specified, a new Dropbox account will be
	// created with the given email address, and that account will be invited to
	// the team. If a personal Dropbox account exists with the email address
	// specified in the call, this call will create a placeholder Dropbox
	// account for the user on the team and send an email inviting the user to
	// migrate their existing personal account onto the team. Team member
	// management apps are required to set an initial given_name and surname for
	// a user to use in the team invitation and for 'Perform as team member'
	// actions taken on the user before they become 'active'.
	MembersAdd(arg *MembersAddArg) (res *MembersAddLaunch, err error)
	// Once an async_job_id is returned from `membersAdd` , use this to poll the
	// status of the asynchronous request. Permission : Team member management
	MembersAddJobStatusGet(arg *async.PollArg) (res *MembersAddJobStatus, err error)
	// Returns information about multiple team members. Permission : Team
	// information This endpoint will return `MembersGetInfoItem.id_not_found`,
	// for IDs (or emails) that cannot be matched to a valid team member.
	MembersGetInfo(arg *MembersGetInfoArgs) (res []*MembersGetInfoItem, err error)
	// Lists members of a team. Permission : Team information
	MembersList(arg *MembersListArg) (res *MembersListResult, err error)
	// Once a cursor has been retrieved from `membersList`, use this to paginate
	// through all team members. Permission : Team information
	MembersListContinue(arg *MembersListContinueArg) (res *MembersListResult, err error)
	// Removes a member from a team. Permission : Team member management Exactly
	// one of team_member_id, email, or external_id must be provided to identify
	// the user account. This is not a deactivation where the account can be
	// re-activated again. Calling `membersAdd` with the removed user's email
	// address will create a new account with a new team_member_id that will not
	// have access to any content that was shared with the initial account. This
	// endpoint may initiate an asynchronous job. To obtain the final result of
	// the job, the client should periodically poll `membersRemoveJobStatusGet`.
	MembersRemove(arg *MembersRemoveArg) (res *async.LaunchEmptyResult, err error)
	// Once an async_job_id is returned from `membersRemove` , use this to poll
	// the status of the asynchronous request. Permission : Team member
	// management
	MembersRemoveJobStatusGet(arg *async.PollArg) (res *async.PollEmptyResult, err error)
	// Sends welcome email to pending team member. Permission : Team member
	// management Exactly one of team_member_id, email, or external_id must be
	// provided to identify the user account. No-op if team member is not
	// pending.
	MembersSendWelcomeEmail(arg *UserSelectorArg) (err error)
	// Updates a team member's permissions. Permission : Team member management
	MembersSetAdminPermissions(arg *MembersSetPermissionsArg) (res *MembersSetPermissionsResult, err error)
	// Updates a team member's profile. Permission : Team member management
	MembersSetProfile(arg *MembersSetProfileArg) (res *TeamMemberInfo, err error)
	// Suspend a member from a team. Permission : Team member management Exactly
	// one of team_member_id, email, or external_id must be provided to identify
	// the user account.
	MembersSuspend(arg *MembersDeactivateArg) (err error)
	// Unsuspend a member from a team. Permission : Team member management
	// Exactly one of team_member_id, email, or external_id must be provided to
	// identify the user account.
	MembersUnsuspend(arg *MembersUnsuspendArg) (err error)
	// Add a property template. See route files/properties/add to add properties
	// to a file.
	PropertiesTemplateAdd(arg *AddPropertyTemplateArg) (res *AddPropertyTemplateResult, err error)
	// Get the schema for a specified template.
	PropertiesTemplateGet(arg *properties.GetPropertyTemplateArg) (res *properties.GetPropertyTemplateResult, err error)
	// Get the property template identifiers for a team. To get the schema of
	// each template use `propertiesTemplateGet`.
	PropertiesTemplateList() (res *properties.ListPropertyTemplateIds, err error)
	// Update a property template. This route can update the template name, the
	// template description and add optional properties to templates.
	PropertiesTemplateUpdate(arg *UpdatePropertyTemplateArg) (res *UpdatePropertyTemplateResult, err error)
	// Retrieves reporting data about a team's user activity.
	ReportsGetActivity(arg *DateRange) (res *GetActivityReport, err error)
	// Retrieves reporting data about a team's linked devices.
	ReportsGetDevices(arg *DateRange) (res *GetDevicesReport, err error)
	// Retrieves reporting data about a team's membership.
	ReportsGetMembership(arg *DateRange) (res *GetMembershipReport, err error)
	// Retrieves reporting data about a team's storage usage.
	ReportsGetStorage(arg *DateRange) (res *GetStorageReport, err error)
}

type apiImpl dropbox.Context
type AlphaGroupsCreateApiError struct {
	dropbox.ApiError
	EndpointError *GroupCreateError `json:"error"`
}

func (dbx *apiImpl) AlphaGroupsCreate(arg *AlphaGroupCreateArg) (res *AlphaGroupFullInfo, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "alpha/groups/create"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError AlphaGroupsCreateApiError
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

type AlphaGroupsGetInfoApiError struct {
	dropbox.ApiError
	EndpointError *GroupsGetInfoError `json:"error"`
}

func (dbx *apiImpl) AlphaGroupsGetInfo(arg *GroupsSelector) (res []*AlphaGroupsGetInfoItem, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "alpha/groups/get_info"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError AlphaGroupsGetInfoApiError
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

type AlphaGroupsListApiError struct {
	dropbox.ApiError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) AlphaGroupsList(arg *GroupsListArg) (res *AlphaGroupsListResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "alpha/groups/list"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError AlphaGroupsListApiError
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

type AlphaGroupsListContinueApiError struct {
	dropbox.ApiError
	EndpointError *GroupsListContinueError `json:"error"`
}

func (dbx *apiImpl) AlphaGroupsListContinue(arg *GroupsListContinueArg) (res *AlphaGroupsListResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "alpha/groups/list/continue"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError AlphaGroupsListContinueApiError
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

type AlphaGroupsUpdateApiError struct {
	dropbox.ApiError
	EndpointError *GroupUpdateError `json:"error"`
}

func (dbx *apiImpl) AlphaGroupsUpdate(arg *AlphaGroupUpdateArgs) (res *AlphaGroupFullInfo, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "alpha/groups/update"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError AlphaGroupsUpdateApiError
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

type DevicesListMemberDevicesApiError struct {
	dropbox.ApiError
	EndpointError *ListMemberDevicesError `json:"error"`
}

func (dbx *apiImpl) DevicesListMemberDevices(arg *ListMemberDevicesArg) (res *ListMemberDevicesResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "devices/list_member_devices"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError DevicesListMemberDevicesApiError
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

type DevicesListMembersDevicesApiError struct {
	dropbox.ApiError
	EndpointError *ListMembersDevicesError `json:"error"`
}

func (dbx *apiImpl) DevicesListMembersDevices(arg *ListMembersDevicesArg) (res *ListMembersDevicesResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "devices/list_members_devices"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError DevicesListMembersDevicesApiError
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

type DevicesListTeamDevicesApiError struct {
	dropbox.ApiError
	EndpointError *ListTeamDevicesError `json:"error"`
}

func (dbx *apiImpl) DevicesListTeamDevices(arg *ListTeamDevicesArg) (res *ListTeamDevicesResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "devices/list_team_devices"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError DevicesListTeamDevicesApiError
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

type DevicesRevokeDeviceSessionApiError struct {
	dropbox.ApiError
	EndpointError *RevokeDeviceSessionError `json:"error"`
}

func (dbx *apiImpl) DevicesRevokeDeviceSession(arg *RevokeDeviceSessionArg) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "devices/revoke_device_session"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError DevicesRevokeDeviceSessionApiError
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

type DevicesRevokeDeviceSessionBatchApiError struct {
	dropbox.ApiError
	EndpointError *RevokeDeviceSessionBatchError `json:"error"`
}

func (dbx *apiImpl) DevicesRevokeDeviceSessionBatch(arg *RevokeDeviceSessionBatchArg) (res *RevokeDeviceSessionBatchResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "devices/revoke_device_session_batch"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError DevicesRevokeDeviceSessionBatchApiError
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

type GetInfoApiError struct {
	dropbox.ApiError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) GetInfo() (res *TeamGetInfoResult, err error) {
	cli := dbx.Client

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "get_info"), nil)
	if err != nil {
		return
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
			var apiError GetInfoApiError
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

type GroupsCreateApiError struct {
	dropbox.ApiError
	EndpointError *GroupCreateError `json:"error"`
}

func (dbx *apiImpl) GroupsCreate(arg *GroupCreateArg) (res *GroupFullInfo, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "groups/create"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError GroupsCreateApiError
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

type GroupsDeleteApiError struct {
	dropbox.ApiError
	EndpointError *GroupDeleteError `json:"error"`
}

func (dbx *apiImpl) GroupsDelete(arg *GroupSelector) (res *async.LaunchEmptyResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "groups/delete"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError GroupsDeleteApiError
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

type GroupsGetInfoApiError struct {
	dropbox.ApiError
	EndpointError *GroupsGetInfoError `json:"error"`
}

func (dbx *apiImpl) GroupsGetInfo(arg *GroupsSelector) (res []*GroupsGetInfoItem, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "groups/get_info"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError GroupsGetInfoApiError
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

type GroupsJobStatusGetApiError struct {
	dropbox.ApiError
	EndpointError *GroupsPollError `json:"error"`
}

func (dbx *apiImpl) GroupsJobStatusGet(arg *async.PollArg) (res *async.PollEmptyResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "groups/job_status/get"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError GroupsJobStatusGetApiError
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

type GroupsListApiError struct {
	dropbox.ApiError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) GroupsList(arg *GroupsListArg) (res *GroupsListResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "groups/list"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError GroupsListApiError
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

type GroupsListContinueApiError struct {
	dropbox.ApiError
	EndpointError *GroupsListContinueError `json:"error"`
}

func (dbx *apiImpl) GroupsListContinue(arg *GroupsListContinueArg) (res *GroupsListResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "groups/list/continue"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError GroupsListContinueApiError
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

type GroupsMembersAddApiError struct {
	dropbox.ApiError
	EndpointError *GroupMembersAddError `json:"error"`
}

func (dbx *apiImpl) GroupsMembersAdd(arg *GroupMembersAddArg) (res *GroupMembersChangeResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "groups/members/add"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError GroupsMembersAddApiError
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

type GroupsMembersListApiError struct {
	dropbox.ApiError
	EndpointError *GroupSelectorError `json:"error"`
}

func (dbx *apiImpl) GroupsMembersList(arg *GroupsMembersListArg) (res *GroupsMembersListResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "groups/members/list"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError GroupsMembersListApiError
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

type GroupsMembersListContinueApiError struct {
	dropbox.ApiError
	EndpointError *GroupsMembersListContinueError `json:"error"`
}

func (dbx *apiImpl) GroupsMembersListContinue(arg *GroupsMembersListContinueArg) (res *GroupsMembersListResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "groups/members/list/continue"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError GroupsMembersListContinueApiError
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

type GroupsMembersRemoveApiError struct {
	dropbox.ApiError
	EndpointError *GroupMembersRemoveError `json:"error"`
}

func (dbx *apiImpl) GroupsMembersRemove(arg *GroupMembersRemoveArg) (res *GroupMembersChangeResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "groups/members/remove"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError GroupsMembersRemoveApiError
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

type GroupsMembersSetAccessTypeApiError struct {
	dropbox.ApiError
	EndpointError *GroupMemberSetAccessTypeError `json:"error"`
}

func (dbx *apiImpl) GroupsMembersSetAccessType(arg *GroupMembersSetAccessTypeArg) (res []*GroupsGetInfoItem, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "groups/members/set_access_type"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError GroupsMembersSetAccessTypeApiError
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

type GroupsUpdateApiError struct {
	dropbox.ApiError
	EndpointError *GroupUpdateError `json:"error"`
}

func (dbx *apiImpl) GroupsUpdate(arg *GroupUpdateArgs) (res *GroupFullInfo, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "groups/update"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError GroupsUpdateApiError
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

type LinkedAppsListMemberLinkedAppsApiError struct {
	dropbox.ApiError
	EndpointError *ListMemberAppsError `json:"error"`
}

func (dbx *apiImpl) LinkedAppsListMemberLinkedApps(arg *ListMemberAppsArg) (res *ListMemberAppsResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "linked_apps/list_member_linked_apps"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError LinkedAppsListMemberLinkedAppsApiError
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

type LinkedAppsListMembersLinkedAppsApiError struct {
	dropbox.ApiError
	EndpointError *ListMembersAppsError `json:"error"`
}

func (dbx *apiImpl) LinkedAppsListMembersLinkedApps(arg *ListMembersAppsArg) (res *ListMembersAppsResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "linked_apps/list_members_linked_apps"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError LinkedAppsListMembersLinkedAppsApiError
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

type LinkedAppsListTeamLinkedAppsApiError struct {
	dropbox.ApiError
	EndpointError *ListTeamAppsError `json:"error"`
}

func (dbx *apiImpl) LinkedAppsListTeamLinkedApps(arg *ListTeamAppsArg) (res *ListTeamAppsResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "linked_apps/list_team_linked_apps"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError LinkedAppsListTeamLinkedAppsApiError
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

type LinkedAppsRevokeLinkedAppApiError struct {
	dropbox.ApiError
	EndpointError *RevokeLinkedAppError `json:"error"`
}

func (dbx *apiImpl) LinkedAppsRevokeLinkedApp(arg *RevokeLinkedApiAppArg) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "linked_apps/revoke_linked_app"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError LinkedAppsRevokeLinkedAppApiError
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

type LinkedAppsRevokeLinkedAppBatchApiError struct {
	dropbox.ApiError
	EndpointError *RevokeLinkedAppBatchError `json:"error"`
}

func (dbx *apiImpl) LinkedAppsRevokeLinkedAppBatch(arg *RevokeLinkedApiAppBatchArg) (res *RevokeLinkedAppBatchResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "linked_apps/revoke_linked_app_batch"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError LinkedAppsRevokeLinkedAppBatchApiError
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

type MembersAddApiError struct {
	dropbox.ApiError
	EndpointError struct{} `json:"error"`
}

func (dbx *apiImpl) MembersAdd(arg *MembersAddArg) (res *MembersAddLaunch, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "members/add"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError MembersAddApiError
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

type MembersAddJobStatusGetApiError struct {
	dropbox.ApiError
	EndpointError *async.PollError `json:"error"`
}

func (dbx *apiImpl) MembersAddJobStatusGet(arg *async.PollArg) (res *MembersAddJobStatus, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "members/add/job_status/get"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError MembersAddJobStatusGetApiError
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

type MembersGetInfoApiError struct {
	dropbox.ApiError
	EndpointError *MembersGetInfoError `json:"error"`
}

func (dbx *apiImpl) MembersGetInfo(arg *MembersGetInfoArgs) (res []*MembersGetInfoItem, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "members/get_info"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError MembersGetInfoApiError
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

type MembersListApiError struct {
	dropbox.ApiError
	EndpointError *MembersListError `json:"error"`
}

func (dbx *apiImpl) MembersList(arg *MembersListArg) (res *MembersListResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "members/list"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError MembersListApiError
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

type MembersListContinueApiError struct {
	dropbox.ApiError
	EndpointError *MembersListContinueError `json:"error"`
}

func (dbx *apiImpl) MembersListContinue(arg *MembersListContinueArg) (res *MembersListResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "members/list/continue"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError MembersListContinueApiError
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

type MembersRemoveApiError struct {
	dropbox.ApiError
	EndpointError *MembersRemoveError `json:"error"`
}

func (dbx *apiImpl) MembersRemove(arg *MembersRemoveArg) (res *async.LaunchEmptyResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "members/remove"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError MembersRemoveApiError
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

type MembersRemoveJobStatusGetApiError struct {
	dropbox.ApiError
	EndpointError *async.PollError `json:"error"`
}

func (dbx *apiImpl) MembersRemoveJobStatusGet(arg *async.PollArg) (res *async.PollEmptyResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "members/remove/job_status/get"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError MembersRemoveJobStatusGetApiError
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

type MembersSendWelcomeEmailApiError struct {
	dropbox.ApiError
	EndpointError *MembersSendWelcomeError `json:"error"`
}

func (dbx *apiImpl) MembersSendWelcomeEmail(arg *UserSelectorArg) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "members/send_welcome_email"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError MembersSendWelcomeEmailApiError
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

type MembersSetAdminPermissionsApiError struct {
	dropbox.ApiError
	EndpointError *MembersSetPermissionsError `json:"error"`
}

func (dbx *apiImpl) MembersSetAdminPermissions(arg *MembersSetPermissionsArg) (res *MembersSetPermissionsResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "members/set_admin_permissions"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError MembersSetAdminPermissionsApiError
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

type MembersSetProfileApiError struct {
	dropbox.ApiError
	EndpointError *MembersSetProfileError `json:"error"`
}

func (dbx *apiImpl) MembersSetProfile(arg *MembersSetProfileArg) (res *TeamMemberInfo, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "members/set_profile"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError MembersSetProfileApiError
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

type MembersSuspendApiError struct {
	dropbox.ApiError
	EndpointError *MembersSuspendError `json:"error"`
}

func (dbx *apiImpl) MembersSuspend(arg *MembersDeactivateArg) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "members/suspend"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError MembersSuspendApiError
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

type MembersUnsuspendApiError struct {
	dropbox.ApiError
	EndpointError *MembersUnsuspendError `json:"error"`
}

func (dbx *apiImpl) MembersUnsuspend(arg *MembersUnsuspendArg) (err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "members/unsuspend"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError MembersUnsuspendApiError
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

type PropertiesTemplateAddApiError struct {
	dropbox.ApiError
	EndpointError *properties.ModifyPropertyTemplateError `json:"error"`
}

func (dbx *apiImpl) PropertiesTemplateAdd(arg *AddPropertyTemplateArg) (res *AddPropertyTemplateResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "properties/template/add"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError PropertiesTemplateAddApiError
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

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "properties/template/get"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "properties/template/list"), nil)
	if err != nil {
		return
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

type PropertiesTemplateUpdateApiError struct {
	dropbox.ApiError
	EndpointError *properties.ModifyPropertyTemplateError `json:"error"`
}

func (dbx *apiImpl) PropertiesTemplateUpdate(arg *UpdatePropertyTemplateArg) (res *UpdatePropertyTemplateResult, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "properties/template/update"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError PropertiesTemplateUpdateApiError
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

type ReportsGetActivityApiError struct {
	dropbox.ApiError
	EndpointError *DateRangeError `json:"error"`
}

func (dbx *apiImpl) ReportsGetActivity(arg *DateRange) (res *GetActivityReport, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "reports/get_activity"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError ReportsGetActivityApiError
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

type ReportsGetDevicesApiError struct {
	dropbox.ApiError
	EndpointError *DateRangeError `json:"error"`
}

func (dbx *apiImpl) ReportsGetDevices(arg *DateRange) (res *GetDevicesReport, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "reports/get_devices"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError ReportsGetDevicesApiError
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

type ReportsGetMembershipApiError struct {
	dropbox.ApiError
	EndpointError *DateRangeError `json:"error"`
}

func (dbx *apiImpl) ReportsGetMembership(arg *DateRange) (res *GetMembershipReport, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "reports/get_membership"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError ReportsGetMembershipApiError
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

type ReportsGetStorageApiError struct {
	dropbox.ApiError
	EndpointError *DateRangeError `json:"error"`
}

func (dbx *apiImpl) ReportsGetStorage(arg *DateRange) (res *GetStorageReport, err error) {
	cli := dbx.Client

	if dbx.Config.Verbose {
		log.Printf("arg: %v", arg)
	}
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", (*dropbox.Context)(dbx).GenerateURL("api", "team", "reports/get_storage"), bytes.NewReader(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
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
			var apiError ReportsGetStorageApiError
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
