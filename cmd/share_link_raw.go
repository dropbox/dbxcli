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
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/auth"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
)

type rawCreateSharedLinkWithSettingsArg struct {
	Path     string                 `json:"path"`
	Settings *rawSharedLinkSettings `json:"settings,omitempty"`
}

type rawModifySharedLinkSettingsArgs struct {
	URL              string                 `json:"url"`
	Settings         *rawSharedLinkSettings `json:"settings"`
	RemoveExpiration bool                   `json:"remove_expiration"`
}

type rawSharedLinkSettings struct {
	RequirePassword *bool                             `json:"require_password,omitempty"`
	LinkPassword    string                            `json:"link_password,omitempty"`
	Expires         *dropbox.DBXTime                  `json:"expires,omitempty"`
	Audience        *sharing.LinkAudience             `json:"audience,omitempty"`
	Access          *sharing.RequestedLinkAccessLevel `json:"access,omitempty"`
	AllowDownload   *bool                             `json:"allow_download,omitempty"`
}

func rawSharedLinkExpires(value *time.Time) *dropbox.DBXTime {
	if value == nil {
		return nil
	}
	t := dropbox.DBXTime(*value)
	return &t
}

func (dbx *sdkSharedLinkClient) CreateSharedLinkWithRawSettings(path string, settings *rawSharedLinkSettings) (sharing.IsSharedLinkMetadata, error) {
	arg := &rawCreateSharedLinkWithSettingsArg{
		Path:     path,
		Settings: settings,
	}
	req := dropbox.Request{
		Host:      "api",
		Namespace: "sharing",
		Route:     "create_shared_link_with_settings",
		Auth:      "user",
		Style:     "rpc",
		Arg:       arg,
	}

	resp, respBody, err := executeSharingRawRequest(dbx.cfg, req, parseCreateSharedLinkWithSettingsError)
	if err != nil {
		return nil, err
	}
	if respBody != nil {
		_ = respBody.Close()
	}

	return parseSharedLinkMetadata(resp)
}

func (dbx *sdkSharedLinkClient) ModifySharedLinkSettingsRaw(url string, settings *rawSharedLinkSettings, removeExpiration bool) error {
	arg := &rawModifySharedLinkSettingsArgs{
		URL:              url,
		Settings:         settings,
		RemoveExpiration: removeExpiration,
	}
	req := dropbox.Request{
		Host:      "api",
		Namespace: "sharing",
		Route:     "modify_shared_link_settings",
		Auth:      "user",
		Style:     "rpc",
		Arg:       arg,
	}

	_, respBody, err := executeSharingRawRequest(dbx.cfg, req, parseModifySharedLinkSettingsError)
	if err != nil {
		return err
	}
	if respBody != nil {
		_ = respBody.Close()
	}

	return nil
}

func executeSharingRawRequest(cfg dropbox.Config, req dropbox.Request, parseError func(error) error) ([]byte, io.ReadCloser, error) {
	ctx := dropbox.NewContext(cfg)
	resp, respBody, err := (&ctx).Execute(req, nil)
	if err != nil {
		return nil, nil, parseError(err)
	}
	return resp, respBody, nil
}

func parseCreateSharedLinkWithSettingsError(err error) error {
	var appErr sharing.CreateSharedLinkWithSettingsAPIError
	parsed := auth.ParseError(err, &appErr)
	if samePointer(parsed, &appErr) {
		return appErr
	}
	return parsed
}

func parseModifySharedLinkSettingsError(err error) error {
	var appErr sharing.ModifySharedLinkSettingsAPIError
	parsed := auth.ParseError(err, &appErr)
	if samePointer(parsed, &appErr) {
		return appErr
	}
	return parsed
}

func samePointer(value, target any) bool {
	valueRef := reflect.ValueOf(value)
	targetRef := reflect.ValueOf(target)
	if !valueRef.IsValid() || !targetRef.IsValid() {
		return false
	}
	if valueRef.Kind() != reflect.Ptr || targetRef.Kind() != reflect.Ptr {
		return false
	}
	if valueRef.Type() != targetRef.Type() {
		return false
	}
	return valueRef.Pointer() == targetRef.Pointer()
}

func parseSharedLinkMetadata(resp []byte) (sharing.IsSharedLinkMetadata, error) {
	var tagged struct {
		Tag string `json:".tag"`
	}
	if err := json.Unmarshal(resp, &tagged); err != nil {
		return nil, err
	}

	switch tagged.Tag {
	case "file":
		var file sharing.FileLinkMetadata
		if err := json.Unmarshal(resp, &file); err != nil {
			return nil, err
		}
		return &file, nil
	case "folder":
		var folder sharing.FolderLinkMetadata
		if err := json.Unmarshal(resp, &folder); err != nil {
			return nil, err
		}
		return &folder, nil
	default:
		return nil, fmt.Errorf("shared link response has unknown type %q", tagged.Tag)
	}
}
