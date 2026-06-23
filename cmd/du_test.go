package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/team_common"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/users"
	"github.com/spf13/cobra"
)

func TestDuTextUsesCommandOutput(t *testing.T) {
	cmd, stdout := testDuCmd()
	stubUsersClient(t, &mockUsersClient{
		getSpaceUsageFn: func() (*users.SpaceUsage, error) {
			return individualSpaceUsage(), nil
		},
	})

	if err := du(cmd, nil); err != nil {
		t.Fatalf("du returned error: %v", err)
	}
	output := stdout.String()
	for _, want := range []string{
		"Used: 1.0 KiB",
		"Type: individual",
		"Allocated: 2.0 KiB",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout = %q, want %q", output, want)
		}
	}
}

func TestDuJSONIndividualAllocation(t *testing.T) {
	cmd, stdout := testDuCmd()
	setDuOutputJSON(t, cmd)
	stubUsersClient(t, &mockUsersClient{
		getSpaceUsageFn: func() (*users.SpaceUsage, error) {
			return individualSpaceUsage(), nil
		},
	})

	if err := du(cmd, nil); err != nil {
		t.Fatalf("du returned error: %v", err)
	}

	got := decodeDuOutput(t, stdout)
	if got.Used != 1024 {
		t.Fatalf("used = %d, want 1024", got.Used)
	}
	if got.Allocation.Type != "individual" || got.Allocation.Allocated == nil || *got.Allocation.Allocated != 2048 {
		t.Fatalf("allocation = %#v, want individual allocation", got.Allocation)
	}
	if got.Allocation.Used != nil {
		t.Fatalf("allocation.used = %#v, want omitted for individual allocation", got.Allocation.Used)
	}
}

func TestDuJSONTeamAllocation(t *testing.T) {
	cmd, stdout := testDuCmd()
	setDuOutputJSON(t, cmd)
	stubUsersClient(t, &mockUsersClient{
		getSpaceUsageFn: func() (*users.SpaceUsage, error) {
			return teamSpaceUsage(), nil
		},
	})

	if err := du(cmd, nil); err != nil {
		t.Fatalf("du returned error: %v", err)
	}

	got := decodeDuOutput(t, stdout)
	if got.Used != 4096 {
		t.Fatalf("used = %d, want 4096", got.Used)
	}
	if got.Allocation.Type != "team" {
		t.Fatalf("allocation.type = %q, want team", got.Allocation.Type)
	}
	if got.Allocation.Allocated == nil || *got.Allocation.Allocated != 8192 {
		t.Fatalf("allocation.allocated = %#v, want 8192", got.Allocation.Allocated)
	}
	if got.Allocation.Used == nil || *got.Allocation.Used != 2048 {
		t.Fatalf("allocation.used = %#v, want 2048", got.Allocation.Used)
	}
	if got.Allocation.UserWithinTeamSpaceLimitType != "alert_only" {
		t.Fatalf("user_within_team_space_limit_type = %q, want alert_only", got.Allocation.UserWithinTeamSpaceLimitType)
	}
}

func TestDuJSONErrorWritesNoOutput(t *testing.T) {
	cmd, stdout := testDuCmd()
	setDuOutputJSON(t, cmd)
	stubUsersClient(t, &mockUsersClient{
		getSpaceUsageFn: func() (*users.SpaceUsage, error) {
			return nil, errors.New("du failed")
		},
	})

	if err := du(cmd, nil); err == nil {
		t.Fatal("expected du error")
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty output on error", got)
	}
}

func TestDuCommandSupportsStructuredOutput(t *testing.T) {
	if !commandSupportsStructuredOutput(duCmd) {
		t.Fatal("du command should support structured output")
	}
}

func testDuCmd() (*cobra.Command, *bytes.Buffer) {
	var stdout bytes.Buffer
	cmd := &cobra.Command{Use: "du"}
	cmd.SetOut(&stdout)
	cmd.Flags().String(outputFlag, "text", "")
	return cmd, &stdout
}

func setDuOutputJSON(t *testing.T, cmd *cobra.Command) {
	t.Helper()
	if err := cmd.Flags().Set(outputFlag, "json"); err != nil {
		t.Fatalf("set output: %v", err)
	}
}

func decodeDuOutput(t *testing.T, out *bytes.Buffer) duOutput {
	t.Helper()

	var got duOutput
	if err := json.NewDecoder(out).Decode(&got); err != nil {
		t.Fatalf("decode JSON output: %v\noutput: %s", err, out.String())
	}
	return got
}

func individualSpaceUsage() *users.SpaceUsage {
	return users.NewSpaceUsage(1024, &users.SpaceAllocation{
		Tagged: dropbox.Tagged{Tag: users.SpaceAllocationIndividual},
		Individual: &users.IndividualSpaceAllocation{
			Allocated: 2048,
		},
	})
}

func teamSpaceUsage() *users.SpaceUsage {
	return users.NewSpaceUsage(4096, &users.SpaceAllocation{
		Tagged: dropbox.Tagged{Tag: users.SpaceAllocationTeam},
		Team: &users.TeamSpaceAllocation{
			Used:                          2048,
			Allocated:                     8192,
			UserWithinTeamSpaceAllocated:  4096,
			UserWithinTeamSpaceUsedCached: 1024,
			UserWithinTeamSpaceLimitType: &team_common.MemberSpaceLimitType{
				Tagged: dropbox.Tagged{Tag: "alert_only"},
			},
		},
	})
}
