package cmd

import (
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

const (
	relocationJSONStatusCopied      = "copied"
	relocationJSONStatusMoved       = "moved"
	relocationJSONStatusSkipped     = "skipped"
	relocationJSONStatusAutorenamed = "autorenamed"
)

type relocationInput struct {
	FromPath string `json:"from_path"`
	ToPath   string `json:"to_path"`
	DryRun   bool   `json:"dry_run,omitempty"`
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

func newPlannedRelocationResult(arg *files.RelocationArg, metadata files.IsMetadata) (relocationResult, error) {
	result, err := jsonMetadataFromDropbox(metadata)
	if err != nil {
		return relocationResult{}, err
	}
	result.PathDisplay = arg.ToPath
	result.PathLower = strings.ToLower(arg.ToPath)

	return relocationResult{
		Input: relocationInput{
			FromPath: arg.FromPath,
			ToPath:   arg.ToPath,
			DryRun:   true,
		},
		Result: result,
	}, nil
}

func relocationOperationResult(status string, result relocationResult) jsonOperationResult {
	return newJSONOperationResult(plannedStatus(result.Input.DryRun, status), result.Result.Type, result.Input, result.Result)
}

// relocationSuccessStatus returns the JSON status for a successful copy/move.
// When --if-exists=autorename caused the server to pick a different path than
// requested, it reports "autorenamed"; otherwise it keeps baseStatus.
func relocationSuccessStatus(baseStatus string, arg *files.RelocationArg, result relocationResult, opts relocationOptions) string {
	if opts.ifExists == relocationIfExistsAutorename && !sameDropboxMetadataPath(result.Result.PathDisplay, result.Result.PathLower, arg.ToPath) {
		return relocationJSONStatusAutorenamed
	}
	return baseStatus
}
