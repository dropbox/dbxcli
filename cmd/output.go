package cmd

import (
	"github.com/dropbox/dbxcli/internal/output"
	"github.com/spf13/cobra"
)

const jsonOutputFlag = "json"

func commandOutput(cmd *cobra.Command) *output.Renderer {
	if cmd == nil {
		return output.New(nil, nil, output.FormatText)
	}

	return output.New(cmd.OutOrStdout(), cmd.ErrOrStderr(), commandOutputFormat(cmd))
}

func commandOutputFormat(cmd *cobra.Command) output.Format {
	jsonOutput, err := cmd.Flags().GetBool(jsonOutputFlag)
	if err != nil {
		jsonOutput, err = cmd.InheritedFlags().GetBool(jsonOutputFlag)
	}
	if err != nil {
		jsonOutput, err = cmd.PersistentFlags().GetBool(jsonOutputFlag)
	}
	if err == nil && jsonOutput {
		return output.FormatJSON
	}
	return output.FormatText
}

func commandVerbose(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err == nil {
		return verbose
	}
	verbose, err = cmd.InheritedFlags().GetBool("verbose")
	if err == nil {
		return verbose
	}
	verbose, err = cmd.PersistentFlags().GetBool("verbose")
	return err == nil && verbose
}

func commandVerboseStatus(cmd *cobra.Command, format string, args ...any) {
	if commandVerbose(cmd) {
		commandOutput(cmd).Status(format, args...)
	}
}
