package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dropbox/dbxcli/internal/output"
	"github.com/spf13/cobra"
)

const (
	outputFlag                          = "output"
	structuredOutputSupportedAnnotation = "dbxcli.supportsStructuredOutput"
)

type jsonErrorResponse struct {
	OK    bool      `json:"ok"`
	Error jsonError `json:"error"`
}

type jsonError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

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
	value, err = cmd.PersistentFlags().GetString(outputFlag)
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

func renderCommandError(cmd *cobra.Command, err error) {
	renderCommandErrorWithJSON(cmd, err, false)
}

func renderCommandErrorWithJSON(cmd *cobra.Command, err error, forceJSON bool) {
	if err == nil {
		return
	}
	if cmd == nil {
		cmd = RootCmd
	}

	if forceJSON || commandOutputFormat(cmd) == output.FormatJSON {
		renderErr := output.New(cmd.OutOrStdout(), cmd.ErrOrStderr(), output.FormatJSON).Render(nil, jsonErrorResponse{
			OK: false,
			Error: jsonError{
				Message: err.Error(),
				Code:    jsonErrorCode(err),
			},
		})
		if renderErr == nil {
			return
		}
		err = renderErr
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
	if jsonErrorCode(err) == "unknown_command" {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Run '%s --help' for usage.\n", cmd.CommandPath())
	}
}

// outputJSONRequested is a narrow pre-parse fallback for errors that happen
// before Cobra has resolved a command/flag context, such as unknown commands.
func outputJSONRequested(args []string) bool {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--":
			return false
		case "--output=json":
			return true
		case "--output":
			return i+1 < len(args) && args[i+1] == "json"
		}
	}
	return false
}

// jsonErrorCode derives stable script-facing codes from existing CLI errors.
// If a matched error message changes, update this mapping with it.
func jsonErrorCode(err error) string {
	if errors.Is(err, output.ErrStructuredOutputUnsupported) {
		return "structured_output_unsupported"
	}

	message := err.Error()
	switch {
	case strings.Contains(message, "unsupported output format"):
		return "unsupported_output_format"
	case strings.Contains(message, "unknown command"):
		return "unknown_command"
	case strings.Contains(message, "unknown flag"):
		return "unknown_flag"
	case strings.Contains(message, "path exists and is not a folder"):
		return "path_conflict"
	case strings.Contains(message, "requires a"):
		return "invalid_arguments"
	default:
		return "command_failed"
	}
}
