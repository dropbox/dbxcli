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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestModifySharedLinkSettingsRawSendsRequirePasswordFalse(t *testing.T) {
	var body map[string]any
	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != "https://api.dropboxapi.com/2/sharing/modify_shared_link_settings" {
				t.Fatalf("url = %q, want modify_shared_link_settings route", req.URL.String())
			}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{".tag":"file"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}
	dbx := &sdkSharedLinkClient{
		cfg: dropbox.Config{
			Token:  "token",
			Client: httpClient,
		},
	}

	requirePassword := false
	if err := dbx.ModifySharedLinkSettingsRaw("https://example.com/link", &rawSharedLinkSettings{
		RequirePassword: &requirePassword,
	}, false); err != nil {
		t.Fatalf("ModifySharedLinkSettingsRaw error: %v", err)
	}

	if body["url"] != "https://example.com/link" {
		t.Fatalf("url = %q, want https://example.com/link", body["url"])
	}
	settings, ok := body["settings"].(map[string]any)
	if !ok {
		t.Fatalf("settings = %#v, want object", body["settings"])
	}
	requirePasswordValue, ok := settings["require_password"].(bool)
	if !ok {
		t.Fatalf("require_password = %#v, want bool", settings["require_password"])
	}
	if requirePasswordValue {
		t.Fatal("require_password = true, want false")
	}
}

func TestCreateSharedLinkWithRawSettingsSendsAllowDownloadFalse(t *testing.T) {
	var body map[string]any
	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != "https://api.dropboxapi.com/2/sharing/create_shared_link_with_settings" {
				t.Fatalf("url = %q, want create_shared_link_with_settings route", req.URL.String())
			}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{".tag":"file","url":"https://example.com/link","name":"file.txt"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}
	dbx := &sdkSharedLinkClient{
		cfg: dropbox.Config{
			Token:  "token",
			Client: httpClient,
		},
	}

	allowDownload := false
	link, err := dbx.CreateSharedLinkWithRawSettings("/file.txt", &rawSharedLinkSettings{
		AllowDownload: &allowDownload,
	})
	if err != nil {
		t.Fatalf("CreateSharedLinkWithRawSettings error: %v", err)
	}

	if body["path"] != "/file.txt" {
		t.Fatalf("path = %q, want /file.txt", body["path"])
	}
	settings, ok := body["settings"].(map[string]any)
	if !ok {
		t.Fatalf("settings = %#v, want object", body["settings"])
	}
	allowDownloadValue, ok := settings["allow_download"].(bool)
	if !ok {
		t.Fatalf("allow_download = %#v, want bool", settings["allow_download"])
	}
	if allowDownloadValue {
		t.Fatal("allow_download = true, want false")
	}
	url, ok := sharedLinkURL(link)
	if !ok || url != "https://example.com/link" {
		t.Fatalf("url = %q, %t, want https://example.com/link", url, ok)
	}
}

func TestModifySharedLinkSettingsRawSendsAllowDownloadFalse(t *testing.T) {
	var body map[string]any
	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != "https://api.dropboxapi.com/2/sharing/modify_shared_link_settings" {
				t.Fatalf("url = %q, want modify_shared_link_settings route", req.URL.String())
			}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{".tag":"file"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}
	dbx := &sdkSharedLinkClient{
		cfg: dropbox.Config{
			Token:  "token",
			Client: httpClient,
		},
	}

	allowDownload := false
	if err := dbx.ModifySharedLinkSettingsRaw("https://example.com/link", &rawSharedLinkSettings{
		AllowDownload: &allowDownload,
	}, false); err != nil {
		t.Fatalf("ModifySharedLinkSettingsRaw error: %v", err)
	}

	if body["url"] != "https://example.com/link" {
		t.Fatalf("url = %q, want https://example.com/link", body["url"])
	}
	settings, ok := body["settings"].(map[string]any)
	if !ok {
		t.Fatalf("settings = %#v, want object", body["settings"])
	}
	allowDownloadValue, ok := settings["allow_download"].(bool)
	if !ok {
		t.Fatalf("allow_download = %#v, want bool", settings["allow_download"])
	}
	if allowDownloadValue {
		t.Fatal("allow_download = true, want false")
	}
}

func TestRawCreateSharedLinkErrorParserReturnsValueError(t *testing.T) {
	err := parseCreateSharedLinkWithSettingsError(dropbox.SDKInternalError{
		StatusCode: http.StatusConflict,
		Content:    `{ "error_summary": "shared_link_already_exists/", "error": { ".tag": "shared_link_already_exists" } }`,
	})

	if _, ok := err.(sharing.CreateSharedLinkWithSettingsAPIError); !ok {
		t.Fatalf("error type = %T, want value CreateSharedLinkWithSettingsAPIError", err)
	}
	if _, ok := err.(*sharing.CreateSharedLinkWithSettingsAPIError); ok {
		t.Fatalf("error type = %T, want non-pointer CreateSharedLinkWithSettingsAPIError", err)
	}
}

func TestRawModifySharedLinkErrorParserReturnsValueError(t *testing.T) {
	err := parseModifySharedLinkSettingsError(dropbox.SDKInternalError{
		StatusCode: http.StatusConflict,
		Content:    `{ "error_summary": "settings_error/", "error": { ".tag": "settings_error" } }`,
	})

	if _, ok := err.(sharing.ModifySharedLinkSettingsAPIError); !ok {
		t.Fatalf("error type = %T, want value ModifySharedLinkSettingsAPIError", err)
	}
	if _, ok := err.(*sharing.ModifySharedLinkSettingsAPIError); ok {
		t.Fatalf("error type = %T, want non-pointer ModifySharedLinkSettingsAPIError", err)
	}
}
