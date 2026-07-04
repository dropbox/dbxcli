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

type shareLinkRevokeInput struct {
	URL  string `json:"url,omitempty"`
	Path string `json:"path,omitempty"`
}

type shareLinkRevokeResult struct {
	URL  string                 `json:"url"`
	Link *shareLinkJSONMetadata `json:"link,omitempty"`
}

func shareLinkRevoke(cmd *cobra.Command, args []string) error {
	opts, err := parseShareLinkRevokeOptions(cmd, args)
	if err != nil {
		return err
	}

	if opts.path != "" {
		revoked, err := revokeSharedLinksForPath(cmd, opts.path)
		if err != nil {
			return err
		}
		return renderJSONOperationOutput(cmd, shareLinkRevokeInput{Path: opts.path}, shareLinkRevokeOperationResults(revoked))
	}

	if len(args) != 1 {
		return invalidArgumentsErrorWithDetails("`share-link revoke` requires a `url` argument", argumentErrorDetails("url"))
	}

	url := args[0]
	if url == "" {
		return invalidArgumentsErrorWithDetails("`share-link revoke` requires a non-empty URL", mergeJSONErrorDetails(argumentErrorDetails("url"), urlErrorDetails(url)))
	}

	dbx := newSharedLinkClient(config)
	arg := sharing.NewRevokeSharedLinkArg(url)
	if err := dbx.RevokeSharedLinkContext(currentContext(), arg); err != nil {
		return withJSONErrorDetails(err, urlErrorDetails(url), operationErrorDetails("share_link_revoke"))
	}

	commandVerboseStatus(cmd, "Revoked shared link %s", url)
	return renderJSONOperationOutput(
		cmd,
		shareLinkRevokeInput{URL: url},
		shareLinkRevokeOperationResults([]shareLinkRevokeResult{{URL: url}}),
	)
}

func shareLinkRevokeOperationResults(revoked []shareLinkRevokeResult) []jsonOperationResult {
	results := make([]jsonOperationResult, 0, len(revoked))
	for _, result := range revoked {
		kind := shareLinkJSONKindSharedLink
		if result.Link != nil {
			kind = result.Link.Type
		}
		results = append(results, newJSONOperationResult(shareLinkJSONStatusRevoked, kind, nil, result))
	}
	return results
}

func parseShareLinkRevokeOptions(cmd *cobra.Command, args []string) (shareLinkRevokeOptions, error) {
	var opts shareLinkRevokeOptions

	if !localFlagChanged(cmd, "path") {
		return opts, nil
	}
	pathArg, err := localStringFlag(cmd, "path")
	if err != nil {
		return opts, err
	}
	if len(args) != 0 {
		details := mergeJSONErrorDetails(operationErrorDetails("share_link_revoke"), flagErrorDetails("path"), argumentErrorDetails("url"))
		if pathArg != "" {
			details = mergeJSONErrorDetails(details, pathErrorDetails(pathArg))
		}
		return opts, invalidArgumentsErrorWithDetails("`--path` cannot be used with a shared link URL", details)
	}
	if pathArg == "" {
		return opts, invalidArgumentsErrorWithDetails("`--path` requires a non-empty path", flagErrorDetails("path"))
	}

	dropboxPath, err := validatePath(pathArg)
	if err != nil {
		return opts, err
	}
	if dropboxPath == "" {
		return opts, invalidArgumentsErrorWithDetails("cannot revoke shared links for Dropbox root", mergeJSONErrorDetails(operationErrorDetails("share_link_revoke"), pathErrorDetails("/")))
	}

	opts.path = dropboxPath
	return opts, nil
}

func revokeSharedLinksForPath(cmd *cobra.Command, path string) ([]shareLinkRevokeResult, error) {
	arg := sharing.NewListSharedLinksArg()
	arg.Path = path
	arg.DirectOnly = true

	dbx := newSharedLinkClient(config)
	links, err := listSharedLinks(dbx, arg)
	if err != nil {
		return nil, withJSONErrorDetails(err, operationErrorDetails("share_link_revoke"), pathErrorDetails(path))
	}
	if len(links) == 0 {
		return nil, withJSONErrorDetails(fmt.Errorf("no direct shared links found for %q", path), operationErrorDetails("share_link_revoke"), pathErrorDetails(path))
	}

	revoked := make([]shareLinkRevokeResult, 0, len(links))
	for _, link := range links {
		url, ok := sharedLinkURL(link)
		if !ok {
			return nil, withJSONErrorDetails(errors.New("shared link response did not include a URL"), operationErrorDetails("share_link_revoke"), pathErrorDetails(path))
		}
		metadata, ok := shareLinkJSONMetadataFromDropbox(link)
		if !ok {
			return nil, withJSONErrorDetails(errors.New("found unknown shared link type"), operationErrorDetails("share_link_revoke"), pathErrorDetails(path))
		}
		if err := dbx.RevokeSharedLinkContext(currentContext(), sharing.NewRevokeSharedLinkArg(url)); err != nil {
			return nil, withJSONErrorDetails(fmt.Errorf("revoke shared link %s: %w", url, err), urlErrorDetails(url), operationErrorDetails("share_link_revoke"))
		}
		revoked = append(revoked, shareLinkRevokeResult{
			URL:  url,
			Link: &metadata,
		})
	}

	commandVerboseStatus(cmd, "Revoked %d shared links for %s", len(links), path)
	return revoked, nil
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
	enableStructuredOutput(shareLinkRevokeCmd)
}
