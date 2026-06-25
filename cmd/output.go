package cmd

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/dropbox/dbxcli/internal/output"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	dropboxauth "github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/auth"
	"github.com/spf13/cobra"
)

const (
	outputFlag                          = "output"
	structuredOutputSupportedAnnotation = "dbxcli.supportsStructuredOutput"

	jsonErrorCodeCommandFailed               = "command_failed"
	jsonErrorCodeAppKeyRequired              = "app_key_required"
	jsonErrorCodeAuthExchangeFailed          = "auth_exchange_failed"
	jsonErrorCodeAuthRefreshFailed           = "auth_refresh_failed"
	jsonErrorCodeAuthRequired                = "auth_required"
	jsonErrorCodeDropboxAPIError             = "dropbox_api_error"
	jsonErrorCodeInvalidArguments            = "invalid_arguments"
	jsonErrorCodeNotFound                    = "not_found"
	jsonErrorCodePathConflict                = "path_conflict"
	jsonErrorCodePermissionDenied            = "permission_denied"
	jsonErrorCodeRateLimited                 = "rate_limited"
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

func authRequiredErrorf(format string, args ...any) error {
	return newCodedError(jsonErrorCodeAuthRequired, fmt.Errorf(format, args...))
}

func appKeyRequiredError(message string) error {
	return newCodedError(jsonErrorCodeAppKeyRequired, errors.New(message))
}

func appKeyRequiredErrorf(format string, args ...any) error {
	return newCodedError(jsonErrorCodeAppKeyRequired, fmt.Errorf(format, args...))
}

func authExchangeFailedError(message string) error {
	return newCodedError(jsonErrorCodeAuthExchangeFailed, errors.New(message))
}

func authExchangeFailedErrorf(format string, args ...any) error {
	return newCodedError(jsonErrorCodeAuthExchangeFailed, fmt.Errorf(format, args...))
}

func authRefreshFailedErrorf(format string, args ...any) error {
	return newCodedError(jsonErrorCodeAuthRefreshFailed, fmt.Errorf(format, args...))
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
		renderErr := output.New(cmd.OutOrStdout(), cmd.ErrOrStderr(), output.FormatJSON).Render(nil, newJSONErrorResponse(cmd, err))
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
	if code := dropboxAPIJSONErrorCode(err); code != "" {
		return code
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

func dropboxAPIJSONErrorCode(err error) string {
	var rateLimitErr dropboxauth.RateLimitAPIError
	var rateLimitErrPtr *dropboxauth.RateLimitAPIError
	if errors.As(err, &rateLimitErr) || errors.As(err, &rateLimitErrPtr) {
		return jsonErrorCodeRateLimited
	}

	var authErr dropboxauth.AuthAPIError
	var authErrPtr *dropboxauth.AuthAPIError
	if errors.As(err, &authErr) {
		return dropboxAuthAPIErrorCode(authErr.AuthError)
	}
	if errors.As(err, &authErrPtr) {
		if authErrPtr == nil {
			return jsonErrorCodeDropboxAPIError
		}
		return dropboxAuthAPIErrorCode(authErrPtr.AuthError)
	}

	var accessErr dropboxauth.AccessAPIError
	var accessErrPtr *dropboxauth.AccessAPIError
	if errors.As(err, &accessErr) || errors.As(err, &accessErrPtr) {
		return jsonErrorCodePermissionDenied
	}

	if summary, ok := dropboxAPIErrorSummary(err); ok {
		return dropboxAPIMessageErrorCode(summary)
	}
	if summary, ok := dropboxAPISummaryFromMessage(err.Error()); ok {
		return dropboxAPIMessageErrorCode(summary)
	}
	return ""
}

func dropboxAuthAPIErrorCode(authErr *dropboxauth.AuthError) string {
	if authErr == nil {
		return jsonErrorCodeDropboxAPIError
	}
	switch authErr.Tag {
	case dropboxauth.AuthErrorInvalidAccessToken, dropboxauth.AuthErrorExpiredAccessToken:
		return jsonErrorCodeAuthRequired
	case dropboxauth.AuthErrorInvalidSelectUser,
		dropboxauth.AuthErrorInvalidSelectAdmin,
		dropboxauth.AuthErrorUserSuspended,
		dropboxauth.AuthErrorMissingScope,
		dropboxauth.AuthErrorRouteAccessDenied:
		return jsonErrorCodePermissionDenied
	default:
		return jsonErrorCodeDropboxAPIError
	}
}

func dropboxAPIErrorSummary(err error) (string, bool) {
	for err != nil {
		if summary, ok := dropboxAPIErrorSummaryValue(err); ok {
			return summary, true
		}
		err = errors.Unwrap(err)
	}
	return "", false
}

func dropboxAPIErrorSummaryValue(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	value := reflect.ValueOf(err)
	if !value.IsValid() {
		return "", false
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return "", false
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return "", false
	}

	typ := value.Type()
	if typ == reflect.TypeOf(dropbox.APIError{}) {
		return err.Error(), true
	}
	if !strings.HasPrefix(typ.PkgPath(), "github.com/dropbox/dropbox-sdk-go-unofficial/") {
		return "", false
	}

	field := value.FieldByName("APIError")
	if field.IsValid() && field.CanInterface() {
		if apiErr, ok := field.Interface().(dropbox.APIError); ok {
			return apiErr.Error(), true
		}
	}
	if strings.HasSuffix(typ.Name(), "APIError") {
		return err.Error(), true
	}
	return "", false
}

func dropboxAPISummaryFromMessage(message string) (string, bool) {
	lower := strings.ToLower(message)
	if strings.Contains(lower, "error in call to api function") {
		return message, true
	}

	trimmed := strings.TrimSpace(message)
	if isDropboxAPISummary(trimmed) {
		return trimmed, true
	}
	if idx := strings.LastIndex(trimmed, ": "); idx >= 0 {
		tail := strings.TrimSpace(trimmed[idx+2:])
		if isDropboxAPISummary(tail) {
			return tail, true
		}
	}
	return "", false
}

func isDropboxAPISummary(message string) bool {
	if message == "" || strings.ContainsAny(message, " \t\r\n\"") || !strings.Contains(message, "/") {
		return false
	}
	segments := strings.Split(message, "/")
	validSegments := 0
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." || strings.HasPrefix(segment, "...") {
			continue
		}
		if !isDropboxAPISummarySegment(segment) {
			return false
		}
		validSegments++
	}
	return validSegments >= 1
}

func isDropboxAPISummarySegment(segment string) bool {
	for _, r := range segment {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return true
}

func dropboxAPIMessageErrorCode(message string) string {
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "invalid_access_token") ||
		strings.Contains(lower, "expired_access_token"):
		return jsonErrorCodeAuthRequired
	case strings.Contains(lower, "too_many_requests") ||
		strings.Contains(lower, "rate_limit") ||
		strings.Contains(lower, "rate_limited"):
		return jsonErrorCodeRateLimited
	case strings.Contains(lower, "path/conflict") ||
		strings.Contains(lower, "to/conflict") ||
		strings.Contains(lower, "from/conflict"):
		return jsonErrorCodePathConflict
	case strings.Contains(lower, "not_found") ||
		strings.Contains(lower, "not found"):
		return jsonErrorCodeNotFound
	case strings.Contains(lower, "no_permission") ||
		strings.Contains(lower, "access_denied") ||
		strings.Contains(lower, "insufficient_permissions") ||
		strings.Contains(lower, "missing_scope") ||
		strings.Contains(lower, "route_access_denied") ||
		strings.Contains(lower, "user_suspended") ||
		strings.Contains(lower, "invalid_select_user") ||
		strings.Contains(lower, "invalid_select_admin"):
		return jsonErrorCodePermissionDenied
	case strings.Contains(lower, "error in call to api function"):
		return jsonErrorCodeDropboxAPIError
	default:
		return ""
	}
}
