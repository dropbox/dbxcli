package cmd

import "github.com/spf13/cobra"

type jsonErrorResponse struct {
	OK       bool          `json:"ok"`
	Error    jsonError     `json:"error"`
	Warnings []jsonWarning `json:"warnings"`
}

type jsonError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

type jsonWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

const (
	jsonWarningCodeDeprecatedCommand = "deprecated_command"
	jsonWarningCodeSkippedSymlink    = "skipped_symlink"
)

type jsonOperationOutput struct {
	Input    any                   `json:"input"`
	Results  []jsonOperationResult `json:"results"`
	Warnings []jsonWarning         `json:"warnings"`
}

type jsonOperationResult struct {
	Status string `json:"status,omitempty"`
	Kind   string `json:"kind,omitempty"`
	Input  any    `json:"input,omitempty"`
	Result any    `json:"result,omitempty"`
}

func newJSONErrorResponse(err error) jsonErrorResponse {
	return jsonErrorResponse{
		OK: false,
		Error: jsonError{
			Message: err.Error(),
			Code:    jsonErrorCode(err),
		},
		Warnings: emptyJSONWarnings(),
	}
}

func newJSONOperationOutput(input any, results []jsonOperationResult, warnings []jsonWarning) jsonOperationOutput {
	return jsonOperationOutput{
		Input:    normalizeJSONInput(input),
		Results:  normalizeJSONOperationResults(results),
		Warnings: normalizeJSONWarnings(warnings),
	}
}

func renderJSONOperationOutput(cmd *cobra.Command, input any, results []jsonOperationResult) error {
	return renderJSONOperationOutputWithWarnings(cmd, input, results, nil)
}

func renderJSONOperationOutputWithWarnings(cmd *cobra.Command, input any, results []jsonOperationResult, warnings []jsonWarning) error {
	return commandOutput(cmd).Render(nil, newJSONOperationOutput(input, results, warnings))
}

func newJSONOperationResult(status, kind string, input any, result any) jsonOperationResult {
	return jsonOperationResult{
		Status: status,
		Kind:   kind,
		Input:  input,
		Result: result,
	}
}

func newJSONMetadataOperationResults(status string, entries []jsonMetadata) []jsonOperationResult {
	results := make([]jsonOperationResult, 0, len(entries))
	for _, entry := range entries {
		results = append(results, newJSONOperationResult(status, entry.Type, nil, entry))
	}
	return results
}

func normalizeJSONInput(input any) any {
	if input == nil {
		return struct{}{}
	}
	return input
}

func normalizeJSONOperationResults(results []jsonOperationResult) []jsonOperationResult {
	if results == nil {
		return []jsonOperationResult{}
	}
	return results
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
