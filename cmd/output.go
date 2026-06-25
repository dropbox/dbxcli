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

	jsonErrorCodeCommandFailed               = "command_failed"
	jsonErrorCodeInvalidArguments            = "invalid_arguments"
	jsonErrorCodePathConflict                = "path_conflict"
	jsonErrorCodeStructuredOutputUnsupported = "structured_output_unsupported"
	jsonErrorCodeUnknownCommand              = "unknown_command"
	jsonErrorCodeUnknownFlag                 = "unknown_flag"
	jsonErrorCodeUnsupportedOutputFormat     = "unsupported_output_format"
)

type jsonCodedError interface {
	error
	JSONErrorCode() string
}

type codedError struct {
	code string
	err  error
}

func (e codedError) Error() string {
	return e.err.Error()
}

func (e codedError) Unwrap() error {
	return e.err
}

func (e codedError) JSONErrorCode() string {
	return e.code
}

func newCodedError(code string, err error) error {
	if err == nil {
		return nil
	}
	return codedError{
		code: code,
		err:  err,
	}
}

func invalidArgumentsError(message string) error {
	return newCodedError(jsonErrorCodeInvalidArguments, errors.New(message))
}

func invalidArgumentsErrorf(format string, args ...any) error {
	return newCodedError(jsonErrorCodeInvalidArguments, fmt.Errorf(format, args...))
}

func pathConflictErrorf(format string, args ...any) error {
	return newCodedError(jsonErrorCodePathConflict, fmt.Errorf(format, args...))
}

func unsupportedOutputFormatErrorf(format string, args ...any) error {
	return newCodedError(jsonErrorCodeUnsupportedOutputFormat, fmt.Errorf(format, args...))
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
		return "", unsupportedOutputFormatErrorf("unsupported output format %q: use text or json", value)
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
		renderErr := output.New(cmd.OutOrStdout(), cmd.ErrOrStderr(), output.FormatJSON).Render(nil, newJSONErrorResponse(err))
		if renderErr == nil {
			return
		}
		err = renderErr
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
	if jsonErrorCode(err) == jsonErrorCodeUnknownCommand {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Run '%s --help' for usage.\n", cmd.CommandPath())
	}
}

// outputJSONRequested is a narrow pre-parse fallback for errors that happen
// before Cobra has resolved a command/flag context, such as unknown commands.
func outputJSONRequested(args []string) bool {
	jsonRequested := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--":
			return jsonRequested
		case "--output=json":
			jsonRequested = true
		case "--output=text":
			jsonRequested = false
		case "--output":
			if i+1 >= len(args) {
				return false
			}
			jsonRequested = args[i+1] == "json"
			i++
		default:
			if strings.HasPrefix(arg, "--output=") {
				jsonRequested = strings.TrimPrefix(arg, "--output=") == "json"
			}
		}
	}
	return jsonRequested
}

// jsonErrorCode derives stable script-facing codes from existing CLI errors.
// Prefer coded errors for dbxcli-owned validation failures. String matching is
// kept only for Cobra-generated unknown command/flag errors and legacy fallback.
func jsonErrorCode(err error) string {
	var coded jsonCodedError
	if errors.As(err, &coded) {
		return coded.JSONErrorCode()
	}
	if errors.Is(err, output.ErrStructuredOutputUnsupported) {
		return jsonErrorCodeStructuredOutputUnsupported
	}

	message := err.Error()
	switch {
	case strings.Contains(message, "unknown command"):
		return jsonErrorCodeUnknownCommand
	case strings.Contains(message, "unknown flag"):
		return jsonErrorCodeUnknownFlag
	default:
		return jsonErrorCodeCommandFailed
	}
}
