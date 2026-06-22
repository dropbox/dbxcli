package cmd

import (
	"fmt"

	"github.com/dropbox/dbxcli/internal/output"
	"github.com/spf13/cobra"
)

const (
	outputFlag                          = "output"
	structuredOutputSupportedAnnotation = "dbxcli.supportsStructuredOutput"
)

func commandOutput(cmd *cobra.Command) *output.Renderer {
	if cmd == nil {
		return output.New(nil, nil, output.FormatText)
	}

	return output.New(cmd.OutOrStdout(), cmd.ErrOrStderr(), commandOutputFormat(cmd))
}

func commandOutputFormat(cmd *cobra.Command) output.Format {
	format, err := commandOutputFormatE(cmd)
	if err != nil {
		return output.FormatText
	}
	return format
}

func commandOutputFormatE(cmd *cobra.Command) (output.Format, error) {
	value := string(output.FormatText)
	if cmd != nil {
		value = commandOutputFlagValue(cmd)
	}
	return parseOutputFormat(value)
}

func commandOutputFlagValue(cmd *cobra.Command) string {
	value, err := cmd.Flags().GetString(outputFlag)
	if err == nil {
		return value
	}
	value, err = cmd.InheritedFlags().GetString(outputFlag)
	if err == nil {
		return value
	}
	return string(output.FormatText)
}

func parseOutputFormat(value string) (output.Format, error) {
	switch output.Format(value) {
	case output.FormatText:
		return output.FormatText, nil
	case output.FormatJSON:
		return output.FormatJSON, nil
	default:
		return "", fmt.Errorf("unsupported output format %q: use text or json", value)
	}
}

func validateOutputFormat(cmd *cobra.Command) error {
	format, err := commandOutputFormatE(cmd)
	if err != nil {
		return err
	}
	if format == output.FormatJSON && !commandSupportsStructuredOutput(cmd) {
		return output.ErrStructuredOutputUnsupported
	}
	return nil
}

func commandSupportsStructuredOutput(cmd *cobra.Command) bool {
	return cmd != nil && cmd.Annotations[structuredOutputSupportedAnnotation] == "true"
}

func enableStructuredOutput(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}
	cmd.Annotations[structuredOutputSupportedAnnotation] = "true"
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
