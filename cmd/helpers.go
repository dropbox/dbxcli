package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

// configFile returns the path to the config file.
func configFile() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user homedir: %w", err)
	}

	return path.Join(dir, configBase, appName, configFileName), nil
}

// readTokens returns a token map read from either the
// DROPBOX_TOKENS environment variable or the filePath, in that
// order.
func readTokens(filePath string) (TokenMap, error) {
	var data []byte
	if envTokens := os.Getenv(tokensEnv); envTokens != "" {
		data = []byte(envTokens)
	} else {
		var err error
		if data, err = os.ReadFile(filePath); err != nil {
			if os.IsNotExist(err) {
				return make(TokenMap), nil
			}
			return nil, fmt.Errorf("read tokens: %w", err)
		}
	}

	var tokens TokenMap
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("decode tokens: %w", err)
	}

	return tokens, nil
}
