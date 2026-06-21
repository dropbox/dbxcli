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
	"io"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

type sharedLinkClient interface {
	CreateSharedLinkWithSettings(arg *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error)
	GetSharedLinkFile(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error)
	GetSharedLinkMetadata(arg *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error)
	ListSharedLinks(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error)
	ModifySharedLinkSettings(arg *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error)
	RemoveSharedLinkPassword(url string) error
	RevokeSharedLink(arg *sharing.RevokeSharedLinkArg) error
}

type sdkSharedLinkClient struct {
	sharing.Client
	cfg dropbox.Config
}

var newSharedLinkClient = func(cfg dropbox.Config) sharedLinkClient {
	return &sdkSharedLinkClient{
		Client: sharing.New(cfg),
		cfg:    cfg,
	}
}

var shareLinkCmd = &cobra.Command{
	Use:   "share-link",
	Short: "Shared link commands",
}

func init() {
	RootCmd.AddCommand(shareLinkCmd)
}
