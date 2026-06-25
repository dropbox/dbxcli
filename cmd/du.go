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
	"fmt"
	"io"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/users"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

type duInput struct{}

type duOutput struct {
	Used       uint64       `json:"used"`
	Allocation duAllocation `json:"allocation"`
}

type duAllocation struct {
	Type                          string  `json:"type"`
	Allocated                     *uint64 `json:"allocated,omitempty"`
	Used                          *uint64 `json:"used,omitempty"`
	UserWithinTeamSpaceAllocated  *uint64 `json:"user_within_team_space_allocated,omitempty"`
	UserWithinTeamSpaceUsedCached *uint64 `json:"user_within_team_space_used_cached,omitempty"`
	UserWithinTeamSpaceLimitType  string  `json:"user_within_team_space_limit_type,omitempty"`
}

const (
	duJSONStatusReported = "reported"
	duKindSpaceUsage     = "space_usage"
)

func du(cmd *cobra.Command, args []string) (err error) {
	dbx := usersNewFunc(config)
	usage, err := dbx.GetSpaceUsage()
	if err != nil {
		return
	}

	return commandOutput(cmd).Render(func(w io.Writer) error {
		return renderUsage(w, usage)
	}, withJSONCommand(cmd, newDuOperationOutput(usage)))
}

func renderUsage(out io.Writer, usage *users.SpaceUsage) error {
	fmt.Fprintf(out, "Used: %s\n", humanize.IBytes(usage.Used))
	fmt.Fprintf(out, "Type: %s\n", usage.Allocation.Tag)

	allocation := usage.Allocation

	switch allocation.Tag {
	case "individual":
		fmt.Fprintf(out, "Allocated: %s\n", humanize.IBytes(allocation.Individual.Allocated))
	case "team":
		fmt.Fprintf(out, "Allocated: %s (Used: %s)\n",
			humanize.IBytes(allocation.Team.Allocated),
			humanize.IBytes(allocation.Team.Used))
	}

	return nil
}

func newDuOutput(usage *users.SpaceUsage) duOutput {
	return duOutput{
		Used:       usage.Used,
		Allocation: newDuAllocation(usage.Allocation),
	}
}

func newDuOperationOutput(usage *users.SpaceUsage) jsonOperationOutput {
	input := duInput{}
	return newJSONOperationOutput(input, []jsonOperationResult{
		newJSONOperationResult(duJSONStatusReported, duKindSpaceUsage, input, newDuOutput(usage)),
	}, nil)
}

func newDuAllocation(allocation *users.SpaceAllocation) duAllocation {
	result := duAllocation{
		Type: allocation.Tag,
	}

	switch allocation.Tag {
	case "individual":
		result.Allocated = uint64Ptr(allocation.Individual.Allocated)
	case "team":
		result.Allocated = uint64Ptr(allocation.Team.Allocated)
		result.Used = uint64Ptr(allocation.Team.Used)
		result.UserWithinTeamSpaceAllocated = uint64Ptr(allocation.Team.UserWithinTeamSpaceAllocated)
		result.UserWithinTeamSpaceUsedCached = uint64Ptr(allocation.Team.UserWithinTeamSpaceUsedCached)
		if allocation.Team.UserWithinTeamSpaceLimitType != nil {
			result.UserWithinTeamSpaceLimitType = allocation.Team.UserWithinTeamSpaceLimitType.Tag
		}
	}

	return result
}

func uint64Ptr(value uint64) *uint64 {
	return &value
}

// duCmd represents the du command
var duCmd = &cobra.Command{
	Use:   "du",
	Short: "Display usage information",
	RunE:  du,
}

func init() {
	RootCmd.AddCommand(duCmd)
	enableStructuredOutput(duCmd)
}
