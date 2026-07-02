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

package main

import (
	"log"
	"runtime/debug"
	"strings"

	"github.com/dropbox/dbxcli/v3/cmd"
)

const defaultVersion = "dev"

var (
	version       = defaultVersion
	readBuildInfo = debug.ReadBuildInfo
)

func init() {
	// Log date, time and file information by default
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	cmd.RootCmd.AddCommand(cmd.NewVersionCommand(resolvedVersion()))
}

func main() {
	cmd.Execute()
}

func resolvedVersion() string {
	if version != "" && version != defaultVersion {
		return version
	}

	info, ok := readBuildInfo()
	if !ok {
		return defaultVersion
	}

	moduleVersion := info.Main.Version
	if moduleVersion == "" || moduleVersion == "(devel)" {
		return defaultVersion
	}

	return strings.TrimPrefix(moduleVersion, "v")
}
