package cmd

import "github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"

type relocationInput struct {
	FromPath string `json:"from_path"`
	ToPath   string `json:"to_path"`
}

type relocationResult struct {
	Input  relocationInput `json:"input"`
	Result jsonMetadata    `json:"result"`
}

type relocationOutput struct {
	Results []relocationResult `json:"results"`
}

func newRelocationResult(arg *files.RelocationArg, res *files.RelocationResult) relocationResult {
	var metadata files.IsMetadata
	if res != nil {
		metadata = res.Metadata
	}

	return relocationResult{
		Input: relocationInput{
			FromPath: arg.FromPath,
			ToPath:   arg.ToPath,
		},
		Result: jsonMetadataFromDropbox(metadata),
	}
}
