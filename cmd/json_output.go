package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

type jsonErrorResponse struct {
	OK            bool          `json:"ok"`
	SchemaVersion string        `json:"schema_version"`
	Command       string        `json:"command"`
	Error         jsonError     `json:"error"`
	Warnings      []jsonWarning `json:"warnings"`
}

type jsonError struct {
	Message string         `json:"message"`
	Code    string         `json:"code"`
	Details map[string]any `json:"details,omitempty"`
}

type jsonWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

const (
	jsonWarningCodeDeprecatedCommand = "deprecated_command"
	jsonWarningCodeSkippedSymlink    = "skipped_symlink"
	jsonWarningCodeTokenRevokeFailed = "token_revoke_failed"
)

type jsonOperationOutput struct {
	OK            bool                  `json:"ok"`
	SchemaVersion string                `json:"schema_version"`
	Command       string                `json:"command"`
	Input         any                   `json:"input"`
	Results       []jsonOperationResult `json:"results"`
	Warnings      []jsonWarning         `json:"warnings"`
}

type jsonOperationResult struct {
	Status string `json:"status"`
	Kind   string `json:"kind"`
	Input  any    `json:"input"`
	Result any    `json:"result"`
}

const jsonSchemaVersion = "1"

func newJSONErrorResponse(cmd *cobra.Command, err error) jsonErrorResponse {
	return jsonErrorResponse{
		OK:            false,
		SchemaVersion: jsonSchemaVersion,
		Command:       jsonCommandPath(cmd),
		Error: jsonError{
			Message: err.Error(),
			Code:    jsonErrorCode(err),
			Details: jsonErrorDetails(err),
		},
		Warnings: emptyJSONWarnings(),
	}
}

func newJSONOperationOutput(input any, results []jsonOperationResult, warnings []jsonWarning) jsonOperationOutput {
	return jsonOperationOutput{
		OK:            true,
		SchemaVersion: jsonSchemaVersion,
		Input:         normalizeJSONInput(input),
		Results:       normalizeJSONOperationResults(results),
		Warnings:      normalizeJSONWarnings(warnings),
	}
}

func newJSONCommandOperationOutput(cmd *cobra.Command, input any, results []jsonOperationResult, warnings []jsonWarning) jsonOperationOutput {
	return withJSONCommand(cmd, newJSONOperationOutput(input, results, warnings))
}

func withJSONCommand(cmd *cobra.Command, out jsonOperationOutput) jsonOperationOutput {
	out.OK = true
	out.SchemaVersion = jsonSchemaVersion
	out.Command = jsonCommandPath(cmd)
	return out
}

func renderJSONOperationOutput(cmd *cobra.Command, input any, results []jsonOperationResult) error {
	return renderJSONOperationOutputWithWarnings(cmd, input, results, nil)
}

func renderJSONOperationOutputWithWarnings(cmd *cobra.Command, input any, results []jsonOperationResult, warnings []jsonWarning) error {
	return commandOutput(cmd).Render(nil, newJSONCommandOperationOutput(cmd, input, results, warnings))
}

func newJSONOperationResult(status, kind string, input any, result any) jsonOperationResult {
	return normalizeJSONOperationResult(jsonOperationResult{
		Status: status,
		Kind:   kind,
		Input:  input,
		Result: result,
	})
}

func newJSONMetadataOperationResults(status string, entries []jsonMetadata) []jsonOperationResult {
	results := make([]jsonOperationResult, 0, len(entries))
	for _, entry := range entries {
		results = append(results, newJSONOperationResult(status, entry.Type, nil, entry))
	}
	return results
}

func normalizeJSONInput(input any) any {
	return normalizeJSONObject(input)
}

func normalizeJSONOperationResults(results []jsonOperationResult) []jsonOperationResult {
	if results == nil {
		return []jsonOperationResult{}
	}
	for i := range results {
		results[i] = normalizeJSONOperationResult(results[i])
	}
	return results
}

func normalizeJSONOperationResult(result jsonOperationResult) jsonOperationResult {
	result.Input = normalizeJSONObject(result.Input)
	result.Result = normalizeJSONObject(result.Result)
	return result
}

func normalizeJSONObject(value any) any {
	if value == nil {
		return struct{}{}
	}
	return value
}

func emptyJSONWarnings() []jsonWarning {
	return []jsonWarning{}
}

func normalizeJSONWarnings(warnings []jsonWarning) []jsonWarning {
	if warnings == nil {
		return emptyJSONWarnings()
	}
	return warnings
}

func jsonCommandPath(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	if cmd.Parent() == nil {
		return cmd.Name()
	}

	var parts []string
	for c := cmd; c != nil && c.Parent() != nil; c = c.Parent() {
		if name := c.Name(); name != "" {
			parts = append([]string{name}, parts...)
		}
	}
	return strings.Join(parts, " ")
}
