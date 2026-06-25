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
	"fmt"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	noDesc := false
	programName := RootCmd.Name()

	completionCmd := &cobra.Command{
		Use:               "completion [bash|zsh|fish|powershell]",
		Short:             "Generate the autocompletion script for the specified shell",
		Long:              completionLong(programName),
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	addCompletionNoDescFlag := func(cmd *cobra.Command) {
		cmd.Flags().BoolVar(&noDesc, "no-descriptions", false, "disable completion descriptions")
	}

	bash := &cobra.Command{
		Use:                   "bash",
		Short:                 "Generate the autocompletion script for bash",
		Long:                  bashCompletionLong(programName),
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		ValidArgsFunction:     cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenBashCompletionV2(cmd.Root().OutOrStdout(), !noDesc)
		},
	}
	addCompletionNoDescFlag(bash)

	zsh := &cobra.Command{
		Use:               "zsh",
		Short:             "Generate the autocompletion script for zsh",
		Long:              zshCompletionLong(programName),
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			if noDesc {
				return cmd.Root().GenZshCompletionNoDesc(cmd.Root().OutOrStdout())
			}
			return cmd.Root().GenZshCompletion(cmd.Root().OutOrStdout())
		},
	}
	addCompletionNoDescFlag(zsh)

	fish := &cobra.Command{
		Use:               "fish",
		Short:             "Generate the autocompletion script for fish",
		Long:              fishCompletionLong(programName),
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenFishCompletion(cmd.Root().OutOrStdout(), !noDesc)
		},
	}
	addCompletionNoDescFlag(fish)

	powershell := &cobra.Command{
		Use:               "powershell",
		Short:             "Generate the autocompletion script for powershell",
		Long:              powershellCompletionLong(programName),
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			if noDesc {
				return cmd.Root().GenPowerShellCompletion(cmd.Root().OutOrStdout())
			}
			return cmd.Root().GenPowerShellCompletionWithDesc(cmd.Root().OutOrStdout())
		},
	}
	addCompletionNoDescFlag(powershell)

	completionCmd.AddCommand(bash, zsh, fish, powershell)
	return completionCmd
}

func completionLong(programName string) string {
	return fmt.Sprintf(`Generate the autocompletion script for %s for the specified shell.
See each sub-command's help for details on how to use the generated script.
`, programName)
}

func bashCompletionLong(programName string) string {
	return fmt.Sprintf(`Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(%[1]s completion bash)

To load completions for every new session, execute once:

#### Linux:

	%[1]s completion bash > /etc/bash_completion.d/%[1]s

#### macOS:

	%[1]s completion bash > $(brew --prefix)/etc/bash_completion.d/%[1]s

You will need to start a new shell for this setup to take effect.
`, programName)
}

func zshCompletionLong(programName string) string {
	return fmt.Sprintf(`Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(%[1]s completion zsh)

To load completions for every new session, execute once:

#### Linux:

	%[1]s completion zsh > "${fpath[1]}/_%[1]s"

#### macOS:

	%[1]s completion zsh > $(brew --prefix)/share/zsh/site-functions/_%[1]s

You will need to start a new shell for this setup to take effect.
`, programName)
}

func fishCompletionLong(programName string) string {
	return fmt.Sprintf(`Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	%[1]s completion fish | source

To load completions for every new session, execute once:

	%[1]s completion fish > ~/.config/fish/completions/%[1]s.fish

You will need to start a new shell for this setup to take effect.
`, programName)
}

func powershellCompletionLong(programName string) string {
	return fmt.Sprintf(`Generate the autocompletion script for powershell.

To load completions in your current shell session:

	%[1]s completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.
`, programName)
}

func init() {
	RootCmd.AddCommand(newCompletionCmd())
}
