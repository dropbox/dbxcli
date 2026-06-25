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
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type sharedLinkPasswordOptions struct {
	password string
	set      bool
}

var readSharedLinkPassword = defaultReadSharedLinkPassword

func addSharedLinkPasswordFlags(cmd *cobra.Command) {
	cmd.Flags().String("password", "", "Password for password-protected shared links")
	cmd.Flags().Bool("password-prompt", false, "Prompt for the shared link password")
	cmd.Flags().String("password-file", "", "Read the shared link password from a file")
}

func sharedLinkPasswordFromFlags(cmd *cobra.Command) (sharedLinkPasswordOptions, error) {
	var sourceCount int
	passwordChanged := localFlagChanged(cmd, "password")
	if passwordChanged {
		sourceCount++
	}

	passwordPrompt, err := localBoolFlag(cmd, "password-prompt")
	if err != nil {
		return sharedLinkPasswordOptions{}, err
	}
	if passwordPrompt {
		sourceCount++
	}

	passwordFile, err := localStringFlag(cmd, "password-file")
	if err != nil {
		return sharedLinkPasswordOptions{}, err
	}
	if passwordFile != "" {
		sourceCount++
	}

	if sourceCount == 0 {
		return sharedLinkPasswordOptions{}, nil
	}
	if sourceCount > 1 {
		return sharedLinkPasswordOptions{}, invalidArgumentsError("use only one of `--password`, `--password-prompt`, or `--password-file`")
	}

	var password string
	switch {
	case passwordChanged:
		password, err = cmd.Flags().GetString("password")
	case passwordPrompt:
		password, err = readSharedLinkPassword("Shared link password: ", cmd.InOrStdin(), cmd.ErrOrStderr())
	case passwordFile != "":
		password, err = sharedLinkPasswordFromFile(passwordFile)
	}
	if err != nil {
		return sharedLinkPasswordOptions{}, err
	}
	if password == "" {
		return sharedLinkPasswordOptions{}, invalidArgumentsError("shared link password cannot be empty")
	}

	return sharedLinkPasswordOptions{
		password: password,
		set:      true,
	}, nil
}

func defaultReadSharedLinkPassword(prompt string, in io.Reader, errOut io.Writer) (string, error) {
	if errOut == nil {
		errOut = io.Discard
	}
	if _, err := fmt.Fprint(errOut, prompt); err != nil {
		return "", err
	}

	if f, ok := in.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		password, err := term.ReadPassword(int(f.Fd()))
		_, _ = fmt.Fprintln(errOut)
		if err != nil {
			return "", err
		}
		return string(password), nil
	}

	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func sharedLinkPasswordFromFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(data), "\r\n"), nil
}

func localFlagChanged(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil && cmd.Flags().Changed(name)
}

func localBoolFlag(cmd *cobra.Command, name string) (bool, error) {
	if cmd.Flags().Lookup(name) == nil {
		return false, nil
	}
	return cmd.Flags().GetBool(name)
}

func localStringFlag(cmd *cobra.Command, name string) (string, error) {
	if cmd.Flags().Lookup(name) == nil {
		return "", nil
	}
	return cmd.Flags().GetString(name)
}
