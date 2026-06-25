package cmd

import "github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"

const (
	relocationJSONStatusCopied = "copied"
	relocationJSONStatusMoved  = "moved"
)

type relocationInput struct {
	FromPath string `json:"from_path"`
	ToPath   string `json:"to_path"`
}

type relocationResult struct {
	Input  relocationInput `json:"input"`
	Result jsonMetadata    `json:"result"`
}

func newRelocationResult(arg *files.RelocationArg, res *files.RelocationResult) (relocationResult, error) {
	var metadata files.IsMetadata
	if res != nil {
		metadata = res.Metadata
	}
	result, err := jsonMetadataFromDropbox(metadata)
	if err != nil {
		return relocationResult{}, err
	}

	return relocationResult{
		Input: relocationInput{
			FromPath: arg.FromPath,
			ToPath:   arg.ToPath,
		},
		Result: result,
	}, nil
}

func relocationOperationResults(status string, results []relocationResult) []jsonOperationResult {
	operationResults := make([]jsonOperationResult, 0, len(results))
	for _, result := range results {
		operationResults = append(operationResults, newJSONOperationResult(status, result.Result.Type, result.Input, result.Result))
	}
	return operationResults
}
