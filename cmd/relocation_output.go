package cmd

import "github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"

const (
	relocationJSONStatusCopied  = "copied"
	relocationJSONStatusMoved   = "moved"
	relocationJSONStatusSkipped = "skipped"
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
	return newRelocationResultFromMetadata(arg, metadata)
}

func newRelocationResultFromMetadata(arg *files.RelocationArg, metadata files.IsMetadata) (relocationResult, error) {
	result, err := jsonMetadataFromDropbox(metadata)
	if err != nil {
		return relocationResult{}, err
	}
	result.PathDisplay = metadataDisplayPath(arg.ToPath, result.PathDisplay)

	return relocationResult{
		Input: relocationInput{
			FromPath: arg.FromPath,
			ToPath:   arg.ToPath,
		},
		Result: result,
	}, nil
}

func relocationOperationResult(status string, result relocationResult) jsonOperationResult {
	return newJSONOperationResult(status, result.Result.Type, result.Input, result.Result)
}
