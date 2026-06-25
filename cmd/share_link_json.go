package cmd

import (
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
)

type shareLinkJSONMetadata struct {
	Type           string                    `json:"type"`
	URL            string                    `json:"url"`
	Name           string                    `json:"name,omitempty"`
	PathLower      string                    `json:"path_lower,omitempty"`
	ID             string                    `json:"id,omitempty"`
	Expires        *string                   `json:"expires,omitempty"`
	Rev            string                    `json:"rev,omitempty"`
	Size           *uint64                   `json:"size,omitempty"`
	ServerModified *string                   `json:"server_modified,omitempty"`
	ClientModified *string                   `json:"client_modified,omitempty"`
	Permissions    *shareLinkJSONPermissions `json:"permissions,omitempty"`
}

type shareLinkJSONPermissions struct {
	ResolvedVisibility            string `json:"resolved_visibility,omitempty"`
	RequestedVisibility           string `json:"requested_visibility,omitempty"`
	EffectiveAudience             string `json:"effective_audience,omitempty"`
	AccessLevel                   string `json:"access_level,omitempty"`
	CanRevoke                     bool   `json:"can_revoke"`
	AllowDownload                 bool   `json:"allow_download"`
	CanSetExpiry                  bool   `json:"can_set_expiry"`
	CanRemoveExpiry               bool   `json:"can_remove_expiry"`
	CanAllowDownload              bool   `json:"can_allow_download"`
	CanDisallowDownload           bool   `json:"can_disallow_download"`
	AllowComments                 bool   `json:"allow_comments"`
	CanSetPassword                bool   `json:"can_set_password,omitempty"`
	CanRemovePassword             bool   `json:"can_remove_password,omitempty"`
	RequirePassword               bool   `json:"require_password,omitempty"`
	CanUseExtendedSharingControls bool   `json:"can_use_extended_sharing_controls,omitempty"`
}

const (
	shareLinkJSONStatusCreated    = "created"
	shareLinkJSONStatusDownloaded = "downloaded"
	shareLinkJSONStatusExisting   = "existing"
	shareLinkJSONStatusFound      = "found"
	shareLinkJSONStatusListed     = "listed"
	shareLinkJSONStatusRevoked    = "revoked"
	shareLinkJSONStatusUpdated    = "updated"

	shareLinkJSONKindSharedLink = "shared_link"
)

func shareLinkJSONMetadataFromDropbox(link sharing.IsSharedLinkMetadata) (shareLinkJSONMetadata, bool) {
	base, linkType, ok := sharedLinkBaseMetadata(link)
	if !ok {
		return shareLinkJSONMetadata{}, false
	}

	result := shareLinkJSONMetadata{
		Type:        linkType,
		URL:         base.Url,
		Name:        base.Name,
		PathLower:   base.PathLower,
		ID:          base.Id,
		Expires:     jsonTimePtr(base.Expires),
		Permissions: shareLinkJSONPermissionsFromDropbox(base.LinkPermissions),
	}

	if file, ok := link.(*sharing.FileLinkMetadata); ok {
		size := file.Size
		result.Rev = file.Rev
		result.Size = &size
		result.ServerModified = jsonTime(file.ServerModified)
		result.ClientModified = jsonTime(file.ClientModified)
	}

	return result, true
}

func shareLinkJSONOperationResult(status string, metadata shareLinkJSONMetadata) jsonOperationResult {
	return newJSONOperationResult(status, metadata.Type, nil, metadata)
}

func shareLinkJSONOperationResults(status string, entries []shareLinkJSONMetadata) []jsonOperationResult {
	results := make([]jsonOperationResult, 0, len(entries))
	for _, entry := range entries {
		results = append(results, shareLinkJSONOperationResult(status, entry))
	}
	return results
}

func shareLinkJSONMetadataListFromDropbox(links []sharing.IsSharedLinkMetadata) ([]shareLinkJSONMetadata, bool) {
	result := make([]shareLinkJSONMetadata, 0, len(links))
	for _, link := range links {
		metadata, ok := shareLinkJSONMetadataFromDropbox(link)
		if !ok {
			return nil, false
		}
		result = append(result, metadata)
	}
	return result, true
}

func shareLinkJSONPermissionsFromDropbox(permissions *sharing.LinkPermissions) *shareLinkJSONPermissions {
	if permissions == nil {
		return nil
	}

	result := &shareLinkJSONPermissions{
		CanRevoke:                     permissions.CanRevoke,
		AllowDownload:                 permissions.AllowDownload,
		CanSetExpiry:                  permissions.CanSetExpiry,
		CanRemoveExpiry:               permissions.CanRemoveExpiry,
		CanAllowDownload:              permissions.CanAllowDownload,
		CanDisallowDownload:           permissions.CanDisallowDownload,
		AllowComments:                 permissions.AllowComments,
		CanSetPassword:                permissions.CanSetPassword,
		CanRemovePassword:             permissions.CanRemovePassword,
		RequirePassword:               permissions.RequirePassword,
		CanUseExtendedSharingControls: permissions.CanUseExtendedSharingControls,
	}
	if permissions.ResolvedVisibility != nil {
		result.ResolvedVisibility = permissions.ResolvedVisibility.Tag
	}
	if permissions.RequestedVisibility != nil {
		result.RequestedVisibility = permissions.RequestedVisibility.Tag
	}
	if permissions.EffectiveAudience != nil {
		result.EffectiveAudience = permissions.EffectiveAudience.Tag
	}
	if permissions.LinkAccessLevel != nil {
		result.AccessLevel = permissions.LinkAccessLevel.Tag
	}
	return result
}

func jsonTimePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	return jsonTime(*value)
}
