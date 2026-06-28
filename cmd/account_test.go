package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/users"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/users_common"
	"github.com/spf13/cobra"
)

func TestAccountCurrentTextUsesCommandOutput(t *testing.T) {
	cmd, stdout := testAccountCmd()
	setCurrentAuthContextForTest(t, &authContext{Source: authSourceSaved, Refreshable: true, AuthFile: authFileDefault})
	stubUsersClient(t, &mockUsersClient{
		getCurrentAccountFn: func() (*users.FullAccount, error) {
			return testFullAccount(), nil
		},
	})

	if err := account(cmd, nil); err != nil {
		t.Fatalf("account returned error: %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "Logged in as Test User <test@example.com>") {
		t.Fatalf("stdout = %q, want current account text output", output)
	}
	if !strings.Contains(output, "Account Type:") {
		t.Fatalf("stdout = %q, want account type", output)
	}
	if !strings.Contains(output, "Auth:") || !strings.Contains(output, "Source:      saved credentials") || !strings.Contains(output, "Refreshable: true") || !strings.Contains(output, "Auth File:   default") {
		t.Fatalf("stdout = %q, want auth context", output)
	}
	if !strings.Contains(output, "Profile Photo URL:") {
		t.Fatalf("stdout = %q, want Profile Photo URL label", output)
	}
	if strings.Contains(output, "Profile Photo Url:") {
		t.Fatalf("stdout = %q, want URL acronym casing", output)
	}
	if strings.Contains(output, "Acting As:") {
		t.Fatalf("stdout = %q, want no Acting As section", output)
	}
}

func TestAccountCurrentJSONOutputsAccount(t *testing.T) {
	cmd, stdout := testAccountCmd()
	setAccountOutputJSON(t, cmd)
	setCurrentAuthContextForTest(t, &authContext{Source: authSourceSaved, Refreshable: true, AuthFile: authFileDefault})
	stubUsersClient(t, &mockUsersClient{
		getCurrentAccountFn: func() (*users.FullAccount, error) {
			return testFullAccount(), nil
		},
	})

	if err := account(cmd, nil); err != nil {
		t.Fatalf("account returned error: %v", err)
	}

	got := decodeAccountOutput(t, stdout)
	if got.Input.AccountID != "" {
		t.Fatalf("input.account_id = %q, want empty for current account", got.Input.AccountID)
	}
	result := got.Results[0]
	if result.Kind != accountKindAccount {
		t.Fatalf("kind = %q, want account", result.Kind)
	}
	if result.Input.AccountID != "" {
		t.Fatalf("result input.account_id = %q, want empty for current account", result.Input.AccountID)
	}
	account := result.Result
	if account.Type != "full" || account.AccountID != "dbid:current" || account.Email != "test@example.com" {
		t.Fatalf("account = %#v, want current full account", account)
	}
	if account.Name == nil || account.Name.DisplayName != "Test User" {
		t.Fatalf("name = %#v, want display name", account.Name)
	}
	if account.AccountType != "business" {
		t.Fatalf("account_type = %q, want business", account.AccountType)
	}
	if account.IsPaired == nil || *account.IsPaired {
		t.Fatalf("is_paired = %#v, want false pointer", account.IsPaired)
	}
	if account.Team == nil || account.Team.ID != "team-id" || account.Team.MemberID != "dbmid:member" {
		t.Fatalf("team = %#v, want team metadata", account.Team)
	}
	assertAccountAuth(t, account.Auth, authSourceSaved, true, authFileDefault)
}

func TestAccountLookupJSONUsesAccountID(t *testing.T) {
	cmd, stdout := testAccountCmd()
	setAccountOutputJSON(t, cmd)
	setCurrentAuthContextForTest(t, &authContext{Source: authSourceEnv, Refreshable: false, AuthFile: authFileNone})
	var gotArg *users.GetAccountArg
	stubUsersClient(t, &mockUsersClient{
		getAccountFn: func(arg *users.GetAccountArg) (*users.BasicAccount, error) {
			gotArg = arg
			return testBasicAccount(), nil
		},
	})

	if err := account(cmd, []string{"dbid:lookup"}); err != nil {
		t.Fatalf("account returned error: %v", err)
	}
	if gotArg == nil || gotArg.AccountId != "dbid:lookup" {
		t.Fatalf("GetAccount arg = %#v, want dbid:lookup", gotArg)
	}

	got := decodeAccountOutput(t, stdout)
	if got.Input.AccountID != "dbid:lookup" {
		t.Fatalf("input.account_id = %q, want dbid:lookup", got.Input.AccountID)
	}
	result := got.Results[0]
	if result.Kind != accountKindAccount {
		t.Fatalf("kind = %q, want account", result.Kind)
	}
	if result.Input.AccountID != "dbid:lookup" {
		t.Fatalf("result input.account_id = %q, want dbid:lookup", result.Input.AccountID)
	}
	account := result.Result
	if account.Type != "basic" || account.AccountID != "dbid:lookup" || account.Email != "lookup@example.com" {
		t.Fatalf("account = %#v, want lookup basic account", account)
	}
	if account.IsTeammate == nil || *account.IsTeammate {
		t.Fatalf("is_teammate = %#v, want false pointer", account.IsTeammate)
	}
	assertAccountAuth(t, account.Auth, authSourceEnv, false, authFileNone)
}

func TestAccountLookupTextIncludesAuth(t *testing.T) {
	cmd, stdout := testAccountCmd()
	setCurrentAuthContextForTest(t, &authContext{Source: authSourceEnv, Refreshable: false, AuthFile: authFileNone})
	stubUsersClient(t, &mockUsersClient{
		getAccountFn: func(arg *users.GetAccountArg) (*users.BasicAccount, error) {
			return testBasicAccount(), nil
		},
	})

	if err := account(cmd, []string{"dbid:lookup"}); err != nil {
		t.Fatalf("account returned error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Auth:") || !strings.Contains(output, "Source:      DBXCLI_ACCESS_TOKEN") || !strings.Contains(output, "Refreshable: false") || !strings.Contains(output, "Auth File:   none") {
		t.Fatalf("stdout = %q, want env auth context", output)
	}
	if strings.Contains(output, "Profile Photo URL:") {
		t.Fatalf("stdout = %q, want empty profile photo omitted", output)
	}
	if strings.Contains(output, "Acting As:") {
		t.Fatalf("stdout = %q, want no Acting As section", output)
	}
}

func TestAccountJSONErrorWritesNoOutput(t *testing.T) {
	cmd, stdout := testAccountCmd()
	setAccountOutputJSON(t, cmd)
	stubUsersClient(t, &mockUsersClient{
		getCurrentAccountFn: func() (*users.FullAccount, error) {
			return nil, errors.New("account failed")
		},
	})

	if err := account(cmd, nil); err == nil {
		t.Fatal("expected account error")
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty output on error", got)
	}
}

func TestAccountCommandSupportsStructuredOutput(t *testing.T) {
	if !commandSupportsStructuredOutput(accountCmd) {
		t.Fatal("account command should support structured output")
	}
}

func testAccountCmd() (*cobra.Command, *bytes.Buffer) {
	var stdout bytes.Buffer
	cmd := &cobra.Command{Use: "account"}
	cmd.SetOut(&stdout)
	cmd.Flags().String(outputFlag, "text", "")
	return cmd, &stdout
}

func setAccountOutputJSON(t *testing.T, cmd *cobra.Command) {
	t.Helper()
	if err := cmd.Flags().Set(outputFlag, "json"); err != nil {
		t.Fatalf("set output: %v", err)
	}
}

func setCurrentAuthContextForTest(t *testing.T, ctx *authContext) {
	t.Helper()

	orig := currentAuthContext
	currentAuthContext = ctx
	t.Cleanup(func() {
		currentAuthContext = orig
	})
}

type accountJSONOutput struct {
	Input    accountInput        `json:"input"`
	Results  []accountJSONResult `json:"results"`
	Warnings []jsonWarning       `json:"warnings"`
}

type accountJSONResult struct {
	Kind   string       `json:"kind"`
	Input  accountInput `json:"input"`
	Result jsonAccount  `json:"result"`
}

func decodeAccountOutput(t *testing.T, out *bytes.Buffer) accountJSONOutput {
	t.Helper()

	var got accountJSONOutput
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("decode JSON output: %v\noutput: %s", err, out.String())
	}
	if got.Warnings == nil {
		t.Fatalf("warnings = nil, want empty array")
	}
	if len(got.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want empty", got.Warnings)
	}
	if len(got.Results) != 1 {
		t.Fatalf("results len = %d, want 1", len(got.Results))
	}
	return got
}

func assertAccountAuth(t *testing.T, got *accountAuth, source string, refreshable bool, authFile string) {
	t.Helper()

	if got == nil {
		t.Fatal("auth = nil, want auth context")
	}
	if got.Source != source || got.Refreshable != refreshable || got.AuthFile != authFile {
		t.Fatalf("auth = %#v, want source=%q refreshable=%t auth_file=%q", got, source, refreshable, authFile)
	}
}

func testFullAccount() *users.FullAccount {
	account := users.NewFullAccount(
		"dbid:current",
		users.NewName("Test", "User", "Test", "Test User", "TU"),
		"test@example.com",
		true,
		false,
		"en",
		"https://dropbox.example/ref",
		false,
		accountType(users_common.AccountTypeBusiness),
		nil,
	)
	account.ProfilePhotoUrl = "https://dropbox.example/photo"
	account.Team = users.NewFullTeam("team-id", "Team Name", nil, nil)
	account.TeamMemberId = "dbmid:member"
	return account
}

func testBasicAccount() *users.BasicAccount {
	return users.NewBasicAccount(
		"dbid:lookup",
		users.NewName("Lookup", "User", "Lookup", "Lookup User", "LU"),
		"lookup@example.com",
		false,
		false,
		false,
	)
}

func accountType(tag string) *users_common.AccountType {
	return &users_common.AccountType{
		Tagged: dropbox.Tagged{Tag: tag},
	}
}
