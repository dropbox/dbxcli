package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/async"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/team"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/team_common"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/users"
	"github.com/spf13/cobra"
)

type mockTeamClient struct {
	getInfoFn             func() (*team.TeamGetInfoResult, error)
	groupsListFn          func(arg *team.GroupsListArg) (*team.GroupsListResult, error)
	groupsListContinueFn  func(arg *team.GroupsListContinueArg) (*team.GroupsListResult, error)
	membersAddFn          func(arg *team.MembersAddArg) (*team.MembersAddLaunch, error)
	membersListFn         func(arg *team.MembersListArg) (*team.MembersListResult, error)
	membersListContinueFn func(arg *team.MembersListContinueArg) (*team.MembersListResult, error)
	membersRemoveFn       func(arg *team.MembersRemoveArg) (*async.LaunchEmptyResult, error)
}

func (m *mockTeamClient) GetInfo() (*team.TeamGetInfoResult, error) {
	if m.getInfoFn != nil {
		return m.getInfoFn()
	}
	return nil, nil
}

func (m *mockTeamClient) GroupsList(arg *team.GroupsListArg) (*team.GroupsListResult, error) {
	if m.groupsListFn != nil {
		return m.groupsListFn(arg)
	}
	return team.NewGroupsListResult(nil, "", false), nil
}

func (m *mockTeamClient) GroupsListContinue(arg *team.GroupsListContinueArg) (*team.GroupsListResult, error) {
	if m.groupsListContinueFn != nil {
		return m.groupsListContinueFn(arg)
	}
	return team.NewGroupsListResult(nil, "", false), nil
}

func (m *mockTeamClient) MembersAdd(arg *team.MembersAddArg) (*team.MembersAddLaunch, error) {
	if m.membersAddFn != nil {
		return m.membersAddFn(arg)
	}
	return nil, nil
}

func (m *mockTeamClient) MembersList(arg *team.MembersListArg) (*team.MembersListResult, error) {
	if m.membersListFn != nil {
		return m.membersListFn(arg)
	}
	return team.NewMembersListResult(nil, "", false), nil
}

func (m *mockTeamClient) MembersListContinue(arg *team.MembersListContinueArg) (*team.MembersListResult, error) {
	if m.membersListContinueFn != nil {
		return m.membersListContinueFn(arg)
	}
	return team.NewMembersListResult(nil, "", false), nil
}

func (m *mockTeamClient) MembersRemove(arg *team.MembersRemoveArg) (*async.LaunchEmptyResult, error) {
	if m.membersRemoveFn != nil {
		return m.membersRemoveFn(arg)
	}
	return nil, nil
}

type teamOperationOutputForTest[I, R any] struct {
	Input    I                                  `json:"input"`
	Results  []teamOperationResultForTest[I, R] `json:"results"`
	Warnings []jsonWarning                      `json:"warnings"`
}

type teamOperationResultForTest[I, R any] struct {
	Status string `json:"status"`
	Kind   string `json:"kind"`
	Input  I      `json:"input"`
	Result R      `json:"result"`
}

func TestTeamInfoJSONOutputsTeamInfo(t *testing.T) {
	stubTeamClient(t, &mockTeamClient{
		getInfoFn: func() (*team.TeamGetInfoResult, error) {
			return &team.TeamGetInfoResult{
				Name:                "Example Team",
				TeamId:              "tid:team",
				NumLicensedUsers:    10,
				NumProvisionedUsers: 7,
			}, nil
		},
	})

	cmd, stdout := teamTestCmd(t, true)
	if err := info(cmd, nil); err != nil {
		t.Fatalf("info error: %v", err)
	}

	got := decodeTeamOperationOutput[map[string]any, teamInfoJSON](t, stdout.Bytes())
	if len(got.Input) != 0 {
		t.Fatalf("input = %#v, want empty object", got.Input)
	}
	result := onlyTeamResult(t, got)
	if result.Status != teamJSONStatusFound || result.Kind != teamJSONKindTeam {
		t.Fatalf("operation result = %#v, want found team", result)
	}
	if result.Result.Name != "Example Team" || result.Result.TeamID != "tid:team" || result.Result.NumLicensedUsers != 10 || result.Result.NumProvisionedUsers != 7 {
		t.Fatalf("result = %#v, want team info", result.Result)
	}
}

func TestTeamInfoTextUsesCommandOutput(t *testing.T) {
	stubTeamClient(t, &mockTeamClient{
		getInfoFn: func() (*team.TeamGetInfoResult, error) {
			return &team.TeamGetInfoResult{Name: "Example Team", TeamId: "tid:team", NumLicensedUsers: 10, NumProvisionedUsers: 7}, nil
		},
	})

	cmd, stdout := teamTestCmd(t, false)
	if err := info(cmd, nil); err != nil {
		t.Fatalf("info error: %v", err)
	}
	if got := stdout.String(); !strings.Contains(got, "Name:") || !strings.Contains(got, "Example Team") {
		t.Fatalf("stdout = %q, want team info text", got)
	}
}

func TestTeamListMembersJSONPaginates(t *testing.T) {
	continueCalled := false
	stubTeamClient(t, &mockTeamClient{
		membersListFn: func(arg *team.MembersListArg) (*team.MembersListResult, error) {
			return team.NewMembersListResult([]*team.TeamMemberInfo{
				testTeamMember("dbmid:one", "one@example.com", "One User"),
			}, "member-cursor", true), nil
		},
		membersListContinueFn: func(arg *team.MembersListContinueArg) (*team.MembersListResult, error) {
			continueCalled = true
			if arg.Cursor != "member-cursor" {
				t.Fatalf("continue cursor = %q, want member-cursor", arg.Cursor)
			}
			return team.NewMembersListResult([]*team.TeamMemberInfo{
				testTeamMember("dbmid:two", "two@example.com", "Two User"),
			}, "", false), nil
		},
	})

	cmd, stdout := teamTestCmd(t, true)
	if err := listMembers(cmd, nil); err != nil {
		t.Fatalf("listMembers error: %v", err)
	}
	if !continueCalled {
		t.Fatal("MembersListContinue was not called")
	}
	got := decodeTeamOperationOutput[map[string]any, teamMemberJSON](t, stdout.Bytes())
	if len(got.Results) != 2 {
		t.Fatalf("results length = %d, want 2", len(got.Results))
	}
	first := got.Results[0]
	if first.Status != teamJSONStatusListed || first.Kind != teamJSONKindTeamMember {
		t.Fatalf("first result = %#v, want listed team_member", first)
	}
	if first.Result.Email != "one@example.com" || first.Result.TeamMemberID != "dbmid:one" || first.Result.Name.DisplayName != "One User" {
		t.Fatalf("first result = %#v, want first member", first.Result)
	}
	if got.Results[1].Result.Email != "two@example.com" {
		t.Fatalf("second result = %#v, want second member", got.Results[1].Result)
	}
}

func TestTeamListGroupsJSONPaginates(t *testing.T) {
	continueCalled := false
	stubTeamClient(t, &mockTeamClient{
		groupsListFn: func(arg *team.GroupsListArg) (*team.GroupsListResult, error) {
			return team.NewGroupsListResult([]*team_common.GroupSummary{
				testTeamGroup("Engineering", "gid:eng", 3),
			}, "group-cursor", true), nil
		},
		groupsListContinueFn: func(arg *team.GroupsListContinueArg) (*team.GroupsListResult, error) {
			continueCalled = true
			if arg.Cursor != "group-cursor" {
				t.Fatalf("continue cursor = %q, want group-cursor", arg.Cursor)
			}
			return team.NewGroupsListResult([]*team_common.GroupSummary{
				testTeamGroup("Support", "gid:support", 2),
			}, "", false), nil
		},
	})

	cmd, stdout := teamTestCmd(t, true)
	if err := listGroups(cmd, nil); err != nil {
		t.Fatalf("listGroups error: %v", err)
	}
	if !continueCalled {
		t.Fatal("GroupsListContinue was not called")
	}
	got := decodeTeamOperationOutput[map[string]any, teamGroupJSON](t, stdout.Bytes())
	if len(got.Results) != 2 {
		t.Fatalf("results length = %d, want 2", len(got.Results))
	}
	if got.Results[0].Status != teamJSONStatusListed || got.Results[0].Kind != teamJSONKindTeamGroup {
		t.Fatalf("first result = %#v, want listed team_group", got.Results[0])
	}
	if got.Results[0].Result.GroupName != "Engineering" || got.Results[0].Result.MemberCount != 3 {
		t.Fatalf("first result = %#v, want engineering group", got.Results[0].Result)
	}
	if got.Results[1].Result.GroupName != "Support" {
		t.Fatalf("second result = %#v, want support group", got.Results[1].Result)
	}
}

func TestTeamAddMemberJSONOutputsMutationResult(t *testing.T) {
	var gotArg *team.MembersAddArg
	member := testTeamMember("dbmid:new", "new@example.com", "New User")
	stubTeamClient(t, &mockTeamClient{
		membersAddFn: func(arg *team.MembersAddArg) (*team.MembersAddLaunch, error) {
			gotArg = arg
			return &team.MembersAddLaunch{
				Tagged: dropbox.Tagged{Tag: "complete"},
				Complete: []*team.MemberAddResult{
					{Tagged: dropbox.Tagged{Tag: "success"}, Success: member},
				},
			}, nil
		},
	})

	cmd, stdout := teamTestCmd(t, true)
	if err := addMember(cmd, []string{"new@example.com", "New", "User"}); err != nil {
		t.Fatalf("addMember error: %v", err)
	}
	if gotArg == nil || len(gotArg.NewMembers) != 1 || gotArg.NewMembers[0].MemberEmail != "new@example.com" || gotArg.NewMembers[0].MemberGivenName != "New" || gotArg.NewMembers[0].MemberSurname != "User" {
		t.Fatalf("MembersAdd arg = %#v, want new member", gotArg)
	}

	got := decodeTeamOperationOutput[teamMemberAddInput, teamMemberMutationJSON](t, stdout.Bytes())
	if got.Input.Email != "new@example.com" || got.Input.FirstName != "New" || got.Input.LastName != "User" {
		t.Fatalf("input = %#v, want add member input", got.Input)
	}
	result := onlyTeamResult(t, got)
	if result.Status != teamJSONStatusAdded || result.Kind != teamJSONKindTeamMember {
		t.Fatalf("operation result = %#v, want added team_member", result)
	}
	if result.Result.Type != teamJSONTypeMemberAdd || result.Result.Tag != "complete" || len(result.Result.Results) != 1 {
		t.Fatalf("mutation result = %#v, want complete add result", result.Result)
	}
	if result.Result.Results[0].Member == nil || result.Result.Results[0].Member.Email != "new@example.com" {
		t.Fatalf("add item = %#v, want new member metadata", result.Result.Results[0])
	}
}

func TestTeamRemoveMemberJSONOutputsMutationResult(t *testing.T) {
	var gotArg *team.MembersRemoveArg
	stubTeamClient(t, &mockTeamClient{
		membersRemoveFn: func(arg *team.MembersRemoveArg) (*async.LaunchEmptyResult, error) {
			gotArg = arg
			return &async.LaunchEmptyResult{Tagged: dropbox.Tagged{Tag: "complete"}}, nil
		},
	})

	cmd, stdout := teamTestCmd(t, true)
	if err := removeMember(cmd, []string{"old@example.com"}); err != nil {
		t.Fatalf("removeMember error: %v", err)
	}
	if gotArg == nil || gotArg.User == nil || gotArg.User.Tag != "email" || gotArg.User.Email != "old@example.com" {
		t.Fatalf("MembersRemove arg = %#v, want email selector", gotArg)
	}

	got := decodeTeamOperationOutput[teamMemberRemoveInput, teamMemberMutationJSON](t, stdout.Bytes())
	if got.Input.Email != "old@example.com" {
		t.Fatalf("input = %#v, want remove input", got.Input)
	}
	result := onlyTeamResult(t, got)
	if result.Status != teamJSONStatusRemoved || result.Kind != teamJSONKindTeamMember {
		t.Fatalf("operation result = %#v, want removed team_member", result)
	}
	if result.Result.Type != teamJSONTypeMemberRemove || result.Result.Tag != "complete" {
		t.Fatalf("mutation result = %#v, want complete remove result", result.Result)
	}
}

func TestTeamAddMemberErrorIncludesEmailDetails(t *testing.T) {
	stubTeamClient(t, &mockTeamClient{
		membersAddFn: func(arg *team.MembersAddArg) (*team.MembersAddLaunch, error) {
			return nil, errors.New("add failed")
		},
	})

	cmd, stdout := teamTestCmd(t, true)
	err := addMember(cmd, []string{"new@example.com", "New", "User"})
	if err == nil {
		t.Fatal("expected addMember error")
	}
	details := jsonErrorDetails(err)
	if details["operation"] != "team_add_member" || details["email"] != "new@example.com" {
		t.Fatalf("details = %#v, want team_add_member operation and email", details)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want no success output", got)
	}
}

func TestTeamRemoveMemberErrorIncludesEmailDetails(t *testing.T) {
	stubTeamClient(t, &mockTeamClient{
		membersRemoveFn: func(arg *team.MembersRemoveArg) (*async.LaunchEmptyResult, error) {
			return nil, errors.New("remove failed")
		},
	})

	cmd, stdout := teamTestCmd(t, true)
	err := removeMember(cmd, []string{"old@example.com"})
	if err == nil {
		t.Fatal("expected removeMember error")
	}
	details := jsonErrorDetails(err)
	if details["operation"] != "team_remove_member" || details["email"] != "old@example.com" {
		t.Fatalf("details = %#v, want team_remove_member operation and email", details)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want no success output", got)
	}
}

func TestTeamJSONErrorWritesNoSuccessOutput(t *testing.T) {
	stubTeamClient(t, &mockTeamClient{
		getInfoFn: func() (*team.TeamGetInfoResult, error) {
			return nil, errors.New("team failed")
		},
	})

	cmd, stdout := teamTestCmd(t, true)
	if err := info(cmd, nil); err == nil {
		t.Fatal("expected info error")
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want no success output", got)
	}
}

func TestTeamCommandsSupportStructuredOutput(t *testing.T) {
	for _, cmd := range []*cobra.Command{
		infoCmd,
		listMembersCmd,
		listGroupsCmd,
		addMemberCmd,
		removeMemberCmd,
	} {
		if !commandSupportsStructuredOutput(cmd) {
			t.Fatalf("%s should support structured output", cmd.CommandPath())
		}
	}
}

func teamTestCmd(t *testing.T, jsonOutput bool) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	if jsonOutput {
		cmd.Flags().String(outputFlag, "text", "")
		if err := cmd.Flags().Set(outputFlag, "json"); err != nil {
			t.Fatalf("set output: %v", err)
		}
	}
	return cmd, &stdout
}

func decodeTeamOperationOutput[I, R any](t *testing.T, data []byte) teamOperationOutputForTest[I, R] {
	t.Helper()
	var got teamOperationOutputForTest[I, R]
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("decode JSON output: %v\noutput: %s", err, string(data))
	}
	if got.Warnings == nil {
		t.Fatalf("warnings = nil, want empty array")
	}
	if len(got.Warnings) != 0 {
		t.Fatalf("warnings = %#v, want empty", got.Warnings)
	}
	return got
}

func onlyTeamResult[I, R any](t *testing.T, got teamOperationOutputForTest[I, R]) teamOperationResultForTest[I, R] {
	t.Helper()
	if len(got.Results) != 1 {
		t.Fatalf("results length = %d, want 1", len(got.Results))
	}
	return got.Results[0]
}

func stubTeamClient(t *testing.T, client teamClient) {
	t.Helper()
	orig := teamNewFunc
	teamNewFunc = func(dropbox.Config) teamClient { return client }
	t.Cleanup(func() {
		teamNewFunc = orig
	})
}

func testTeamMember(memberID, email, displayName string) *team.TeamMemberInfo {
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	return &team.TeamMemberInfo{
		Profile: &team.TeamMemberProfile{
			MemberProfile: team.MemberProfile{
				TeamMemberId:    memberID,
				AccountId:       "dbid:" + memberID,
				Email:           email,
				EmailVerified:   true,
				Status:          &team.TeamMemberStatus{Tagged: dropbox.Tagged{Tag: "active"}},
				Name:            users.NewName(strings.Split(displayName, " ")[0], "User", strings.Split(displayName, " ")[0], displayName, "TU"),
				MembershipType:  &team.TeamMembershipType{Tagged: dropbox.Tagged{Tag: "full"}},
				JoinedOn:        &now,
				ProfilePhotoUrl: "https://example.com/photo",
			},
			Groups:         []string{"gid:eng"},
			MemberFolderId: "ns:" + memberID,
		},
		Role: &team.AdminTier{Tagged: dropbox.Tagged{Tag: "member_only"}},
	}
}

func testTeamGroup(name, id string, memberCount uint32) *team_common.GroupSummary {
	group := team_common.NewGroupSummary(name, id, &team_common.GroupManagementType{Tagged: dropbox.Tagged{Tag: "user_managed"}})
	group.MemberCount = memberCount
	group.GroupExternalId = "external:" + id
	return group
}
