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
	"io"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

type sharedLinkClient interface {
	CreateSharedLinkWithSettingsContext(context.Context, *sharing.CreateSharedLinkWithSettingsArg) (sharing.IsSharedLinkMetadata, error)
	CreateSharedLinkWithRawSettingsContext(context.Context, string, *rawSharedLinkSettings) (sharing.IsSharedLinkMetadata, error)
	GetSharedLinkFileContext(context.Context, *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, io.ReadCloser, error)
	GetSharedLinkMetadataContext(context.Context, *sharing.GetSharedLinkMetadataArg) (sharing.IsSharedLinkMetadata, error)
	ListSharedLinksContext(context.Context, *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error)
	ModifySharedLinkSettingsContext(context.Context, *sharing.ModifySharedLinkSettingsArgs) (sharing.IsSharedLinkMetadata, error)
	RevokeSharedLinkContext(context.Context, *sharing.RevokeSharedLinkArg) error
	ModifySharedLinkSettingsRawContext(context.Context, string, *rawSharedLinkSettings, bool) error
}

type sdkSharedLinkClient struct {
	sharing.ContextClient
	cfg dropbox.Config
}

var newSharedLinkClient = func(cfg dropbox.Config) sharedLinkClient {
	return &sdkSharedLinkClient{
		ContextClient: sharing.NewContext(cfg),
		cfg:           cfg,
	}
}

var shareLinkCmd = &cobra.Command{
	Use:   "share-link",
	Short: "Shared link commands",
	Long:  "Create, list, inspect, download, update, and revoke Dropbox shared links.",
}

func init() {
	RootCmd.AddCommand(shareLinkCmd)
}
