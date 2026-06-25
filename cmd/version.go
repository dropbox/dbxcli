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

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/spf13/cobra"
)

type versionInput struct{}

type versionOutput struct {
	Version     string `json:"version"`
	SDKVersion  string `json:"sdk_version"`
	SpecVersion string `json:"spec_version"`
}

const versionKindVersion = "version"

var dropboxVersionFunc = dropbox.Version

// NewVersionCommand creates the version command. The version value is supplied
// by main so release builds can continue setting it with ldflags.
func NewVersionCommand(version string) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return versionCommand(cmd, version)
		},
	}
	enableStructuredOutput(versionCmd)
	return versionCmd
}

func versionCommand(cmd *cobra.Command, version string) error {
	info := newVersionOutput(version)
	return commandOutput(cmd).Render(func(w io.Writer) error {
		return renderVersion(w, info)
	}, newVersionOperationOutput(info))
}

func renderVersion(out io.Writer, info versionOutput) error {
	fmt.Fprintln(out, "dbxcli version:", info.Version)
	fmt.Fprintln(out, "SDK version:", info.SDKVersion)
	fmt.Fprintln(out, "Spec version:", info.SpecVersion)
	return nil
}

func newVersionOutput(version string) versionOutput {
	sdkVersion, specVersion := dropboxVersionFunc()
	return versionOutput{
		Version:     version,
		SDKVersion:  sdkVersion,
		SpecVersion: specVersion,
	}
}

func newVersionOperationOutput(info versionOutput) jsonOperationOutput {
	input := versionInput{}
	return newJSONOperationOutput(input, []jsonOperationResult{
		newJSONOperationResult("", versionKindVersion, input, info),
	}, nil)
}
