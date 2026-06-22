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
	"errors"
	"fmt"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

type shareLinkRevokeOptions struct {
	path string
}

func shareLinkRevoke(cmd *cobra.Command, args []string) error {
	opts, err := parseShareLinkRevokeOptions(cmd, args)
	if err != nil {
		return err
	}

	if opts.path != "" {
		return revokeSharedLinksForPath(cmd, opts.path)
	}

	if len(args) != 1 {
		return errors.New("`share-link revoke` requires a `url` argument")
	}

	url := args[0]
	if url == "" {
		return errors.New("`share-link revoke` requires a non-empty URL")
	}

	dbx := newSharedLinkClient(config)
	arg := sharing.NewRevokeSharedLinkArg(url)
	if err := dbx.RevokeSharedLink(arg); err != nil {
		return err
	}

	commandVerboseStatus(cmd, "Revoked shared link %s", url)
	return nil
}

func parseShareLinkRevokeOptions(cmd *cobra.Command, args []string) (shareLinkRevokeOptions, error) {
	var opts shareLinkRevokeOptions

	if !localFlagChanged(cmd, "path") {
		return opts, nil
	}
	if len(args) != 0 {
		return opts, errors.New("`--path` cannot be used with a shared link URL")
	}

	pathArg, err := localStringFlag(cmd, "path")
	if err != nil {
		return opts, err
	}
	if pathArg == "" {
		return opts, errors.New("`--path` requires a non-empty path")
	}

	path, err := validatePath(pathArg)
	if err != nil {
		return opts, err
	}
	if path == "" {
		return opts, errors.New("cannot revoke shared links for Dropbox root")
	}

	opts.path = path
	return opts, nil
}

func revokeSharedLinksForPath(cmd *cobra.Command, path string) error {
	arg := sharing.NewListSharedLinksArg()
	arg.Path = path
	arg.DirectOnly = true

	dbx := newSharedLinkClient(config)
	links, err := listSharedLinks(dbx, arg)
	if err != nil {
		return err
	}
	if len(links) == 0 {
		return fmt.Errorf("no direct shared links found for %q", path)
	}

	for _, link := range links {
		url, ok := sharedLinkURL(link)
		if !ok {
			return errors.New("shared link response did not include a URL")
		}
		if err := dbx.RevokeSharedLink(sharing.NewRevokeSharedLinkArg(url)); err != nil {
			return fmt.Errorf("revoke shared link %s: %w", url, err)
		}
	}

	commandVerboseStatus(cmd, "Revoked %d shared links for %s", len(links), path)
	return nil
}

var shareLinkRevokeCmd = &cobra.Command{
	Use:   "revoke [url]",
	Short: "Revoke shared links",
	Long:  "Revoke a shared link by URL, or revoke all direct shared links for a Dropbox path with --path.",
	Example: `  dbxcli share-link revoke https://www.dropbox.com/s/example/file.txt
  dbxcli share-link revoke --path /file.txt`,
	RunE: shareLinkRevoke,
}

func init() {
	shareLinkRevokeCmd.Flags().String("path", "", "Revoke direct shared links for a Dropbox path")
	shareLinkCmd.AddCommand(shareLinkRevokeCmd)
}
