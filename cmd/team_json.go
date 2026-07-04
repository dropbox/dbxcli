// Copyright © 2016 Dropbox, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/async"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/team"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/team_common"
)

type teamClient interface {
	GetInfoContext(context.Context) (*team.TeamGetInfoResult, error)
	GroupsListContext(context.Context, *team.GroupsListArg) (*team.GroupsListResult, error)
	GroupsListContinueContext(context.Context, *team.GroupsListContinueArg) (*team.GroupsListResult, error)
	MembersAddContext(context.Context, *team.MembersAddArg) (*team.MembersAddLaunch, error)
	MembersListContext(context.Context, *team.MembersListArg) (*team.MembersListResult, error)
	MembersListContinueContext(context.Context, *team.MembersListContinueArg) (*team.MembersListResult, error)
	MembersRemoveContext(context.Context, *team.MembersRemoveArg) (*async.LaunchEmptyResult, error)
}

type teamInfoInput struct{}

type teamInfoJSON struct {
	Type                string `json:"type"`
	Name                string `json:"name"`
	TeamID              string `json:"team_id"`
	NumLicensedUsers    uint32 `json:"num_licensed_users"`
	NumProvisionedUsers uint32 `json:"num_provisioned_users"`
}

type teamMemberJSON struct {
	Type                  string           `json:"type"`
	TeamMemberID          string           `json:"team_member_id"`
	ExternalID            string           `json:"external_id,omitempty"`
	AccountID             string           `json:"account_id,omitempty"`
	Email                 string           `json:"email"`
	EmailVerified         bool             `json:"email_verified"`
	Status                string           `json:"status,omitempty"`
	Name                  *jsonAccountName `json:"name,omitempty"`
	Role                  string           `json:"role,omitempty"`
	Groups                []string         `json:"groups,omitempty"`
	MemberFolderID        string           `json:"member_folder_id,omitempty"`
	MembershipType        string           `json:"membership_type,omitempty"`
	InvitedOn             *string          `json:"invited_on,omitempty"`
	JoinedOn              *string          `json:"joined_on,omitempty"`
	SuspendedOn           *string          `json:"suspended_on,omitempty"`
	PersistentID          string           `json:"persistent_id,omitempty"`
	IsDirectoryRestricted bool             `json:"is_directory_restricted,omitempty"`
	ProfilePhotoURL       string           `json:"profile_photo_url,omitempty"`
}

type teamGroupJSON struct {
	Type                string `json:"type"`
	GroupName           string `json:"group_name"`
	GroupID             string `json:"group_id"`
	GroupExternalID     string `json:"group_external_id,omitempty"`
	MemberCount         uint32 `json:"member_count,omitempty"`
	GroupManagementType string `json:"group_management_type,omitempty"`
}

type teamMemberAddInput struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type teamMemberRemoveInput struct {
	Email string `json:"email"`
}

type teamMemberMutationJSON struct {
	Type       string                  `json:"type"`
	Tag        string                  `json:"tag"`
	AsyncJobID string                  `json:"async_job_id,omitempty"`
	Results    []teamMemberAddItemJSON `json:"results,omitempty"`
}

type teamMemberAddItemJSON struct {
	Tag    string          `json:"tag"`
	Email  string          `json:"email,omitempty"`
	Member *teamMemberJSON `json:"member,omitempty"`
}

const (
	teamJSONKindTeam         = "team"
	teamJSONKindTeamGroup    = "team_group"
	teamJSONKindTeamMember   = "team_member"
	teamJSONStatusAdded      = "added"
	teamJSONStatusCompleted  = "completed"
	teamJSONStatusListed     = "listed"
	teamJSONStatusFound      = "found"
	teamJSONStatusRemoved    = "removed"
	teamJSONStatusStarted    = "started"
	teamJSONTypeMemberAdd    = "team_member_add"
	teamJSONTypeMemberRemove = "team_member_remove"
)

var teamNewFunc = func(cfg dropbox.Config) teamClient {
	return team.NewContext(cfg)
}

func teamInfoOperationOutput(info *team.TeamGetInfoResult) jsonOperationOutput {
	input := teamInfoInput{}
	return newJSONOperationOutput(input, []jsonOperationResult{
		newJSONOperationResult(teamJSONStatusFound, teamJSONKindTeam, input, teamInfoJSONFromDropbox(info)),
	}, nil)
}

func teamMemberOperationResults(members []*team.TeamMemberInfo) []jsonOperationResult {
	results := make([]jsonOperationResult, 0, len(members))
	for _, member := range members {
		results = append(results, newJSONOperationResult(teamJSONStatusListed, teamJSONKindTeamMember, nil, teamMemberJSONFromDropbox(member)))
	}
	return results
}

func teamGroupOperationResults(groups []*team_common.GroupSummary) []jsonOperationResult {
	results := make([]jsonOperationResult, 0, len(groups))
	for _, group := range groups {
		results = append(results, newJSONOperationResult(teamJSONStatusListed, teamJSONKindTeamGroup, nil, teamGroupJSONFromDropbox(group)))
	}
	return results
}

func teamMemberAddOperationOutput(input teamMemberAddInput, res *team.MembersAddLaunch) jsonOperationOutput {
	return newJSONOperationOutput(input, []jsonOperationResult{
		newJSONOperationResult(teamMemberAddStatus(res), teamJSONKindTeamMember, input, teamMemberAddJSONFromDropbox(res)),
	}, nil)
}

func teamMemberRemoveOperationOutput(input teamMemberRemoveInput, res *async.LaunchEmptyResult) jsonOperationOutput {
	return newJSONOperationOutput(input, []jsonOperationResult{
		newJSONOperationResult(teamMemberRemoveStatus(res), teamJSONKindTeamMember, input, teamMemberRemoveJSONFromDropbox(res)),
	}, nil)
}

func teamInfoJSONFromDropbox(info *team.TeamGetInfoResult) teamInfoJSON {
	if info == nil {
		return teamInfoJSON{Type: teamJSONKindTeam}
	}
	return teamInfoJSON{
		Type:                teamJSONKindTeam,
		Name:                info.Name,
		TeamID:              info.TeamId,
		NumLicensedUsers:    info.NumLicensedUsers,
		NumProvisionedUsers: info.NumProvisionedUsers,
	}
}

func teamMemberJSONFromDropbox(member *team.TeamMemberInfo) teamMemberJSON {
	result := teamMemberJSON{Type: teamJSONKindTeamMember}
	if member == nil {
		return result
	}
	if member.Role != nil {
		result.Role = member.Role.Tag
	}
	profile := member.Profile
	if profile == nil {
		return result
	}
	result.TeamMemberID = profile.TeamMemberId
	result.ExternalID = profile.ExternalId
	result.AccountID = profile.AccountId
	result.Email = profile.Email
	result.EmailVerified = profile.EmailVerified
	if profile.Status != nil {
		result.Status = profile.Status.Tag
	}
	result.Name = jsonAccountNameFromDropbox(profile.Name)
	result.Groups = profile.Groups
	result.MemberFolderID = profile.MemberFolderId
	if profile.MembershipType != nil {
		result.MembershipType = profile.MembershipType.Tag
	}
	result.InvitedOn = jsonDBXTimePtr(profile.InvitedOn)
	result.JoinedOn = jsonDBXTimePtr(profile.JoinedOn)
	result.SuspendedOn = jsonDBXTimePtr(profile.SuspendedOn)
	result.PersistentID = profile.PersistentId
	result.IsDirectoryRestricted = profile.IsDirectoryRestricted
	result.ProfilePhotoURL = profile.ProfilePhotoUrl
	return result
}

func teamGroupJSONFromDropbox(group *team_common.GroupSummary) teamGroupJSON {
	if group == nil {
		return teamGroupJSON{Type: teamJSONKindTeamGroup}
	}
	result := teamGroupJSON{
		Type:            teamJSONKindTeamGroup,
		GroupName:       group.GroupName,
		GroupID:         group.GroupId,
		GroupExternalID: group.GroupExternalId,
		MemberCount:     group.MemberCount,
	}
	if group.GroupManagementType != nil {
		result.GroupManagementType = group.GroupManagementType.Tag
	}
	return result
}

func teamMemberAddJSONFromDropbox(res *team.MembersAddLaunch) teamMemberMutationJSON {
	result := teamMemberMutationJSON{Type: teamJSONTypeMemberAdd}
	if res == nil {
		return result
	}
	result.Tag = res.Tag
	result.AsyncJobID = res.AsyncJobId
	result.Results = make([]teamMemberAddItemJSON, 0, len(res.Complete))
	for _, item := range res.Complete {
		result.Results = append(result.Results, teamMemberAddItemJSONFromDropbox(item))
	}
	return result
}

func teamMemberAddItemJSONFromDropbox(item *team.MemberAddResult) teamMemberAddItemJSON {
	if item == nil {
		return teamMemberAddItemJSON{}
	}
	result := teamMemberAddItemJSON{
		Tag:   item.Tag,
		Email: teamMemberAddResultEmail(item),
	}
	if item.Success != nil {
		member := teamMemberJSONFromDropbox(item.Success)
		result.Member = &member
		if item.Success.Profile != nil {
			result.Email = item.Success.Profile.Email
		}
	}
	return result
}

func teamMemberAddResultEmail(item *team.MemberAddResult) string {
	switch item.Tag {
	case "team_license_limit":
		return item.TeamLicenseLimit
	case "free_team_member_limit_reached":
		return item.FreeTeamMemberLimitReached
	case "user_already_on_team":
		return item.UserAlreadyOnTeam
	case "user_on_another_team":
		return item.UserOnAnotherTeam
	case "user_already_paired":
		return item.UserAlreadyPaired
	case "user_migration_failed":
		return item.UserMigrationFailed
	case "duplicate_external_member_id":
		return item.DuplicateExternalMemberId
	case "duplicate_member_persistent_id":
		return item.DuplicateMemberPersistentId
	case "persistent_id_disabled":
		return item.PersistentIdDisabled
	case "user_creation_failed":
		return item.UserCreationFailed
	default:
		return ""
	}
}

func teamMemberRemoveJSONFromDropbox(res *async.LaunchEmptyResult) teamMemberMutationJSON {
	result := teamMemberMutationJSON{Type: teamJSONTypeMemberRemove}
	if res == nil {
		return result
	}
	result.Tag = res.Tag
	result.AsyncJobID = res.AsyncJobId
	return result
}

func teamMemberAddStatus(res *team.MembersAddLaunch) string {
	if res == nil {
		return teamJSONStatusCompleted
	}
	if res.Tag != "complete" {
		return teamJSONStatusStarted
	}
	for _, item := range res.Complete {
		if item != nil && item.Tag != "success" {
			return teamJSONStatusCompleted
		}
	}
	return teamJSONStatusAdded
}

func teamMemberRemoveStatus(res *async.LaunchEmptyResult) string {
	if res == nil {
		return teamJSONStatusCompleted
	}
	if res.Tag == "complete" {
		return teamJSONStatusRemoved
	}
	return teamJSONStatusStarted
}
