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
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

const (
	configFileName = "auth.json"
	envAccessToken = "DBXCLI_ACCESS_TOKEN"
	envAuthFile    = "DBXCLI_AUTH_FILE"
)

// TokenMap maps domains to a map of commands to tokens.
// For each domain, we want to save different tokens depending on the
// command type: personal, team access and team manage.
type TokenMap map[string]map[string]string

func authFilePath() (string, error) {
	if filePath := os.Getenv(envAuthFile); filePath != "" {
		return homedir.Expand(filePath)
	}

	dir, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".config", "dbxcli", configFileName), nil
}

func readTokens(filePath string) (TokenMap, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var tokens TokenMap
	if err := json.Unmarshal(b, &tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

func writeTokens(filePath string, tokens TokenMap) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(filePath), 0700); err != nil {
			return err
		}
	}

	b, err := json.Marshal(tokens)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, b, 0600)
}
