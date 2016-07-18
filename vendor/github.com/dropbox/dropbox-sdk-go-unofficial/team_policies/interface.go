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

package team_policies

type EmmState struct {
	Tag string `json:".tag"`
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

// Policies governing team members.
type TeamMemberPolicies struct {
	// Policies governing sharing.
	Sharing *TeamSharingPolicies `json:"sharing"`
	// This describes the Enterprise Mobility Management (EMM) state for this team.
	// This information can be used to understand if an organization is integrating
	// with a third-party EMM vendor to further manage and apply restrictions upon
	// the team's Dropbox usage on mobile devices. This is a new feature and in the
	// future we'll be adding more new fields and additional documentation.
	EmmState *EmmState `json:"emm_state"`
}

func NewTeamMemberPolicies(Sharing *TeamSharingPolicies, EmmState *EmmState) *TeamMemberPolicies {
	s := new(TeamMemberPolicies)
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
