package cmd

import (
	"errors"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

const (
	relocationIfExistsFail = "fail"
	relocationIfExistsSkip = "skip"
)

type relocationOptions struct {
	ifExists string
}

func parseRelocationOptions(cmd *cobra.Command) (relocationOptions, error) {
	ifExists, err := parseRelocationIfExists(cmd)
	if err != nil {
		return relocationOptions{}, err
	}
	return relocationOptions{ifExists: ifExists}, nil
}

func parseRelocationIfExists(cmd *cobra.Command) (string, error) {
	if cmd == nil {
		return relocationIfExistsFail, nil
	}
	ifExists, err := cmd.Flags().GetString("if-exists")
	if err != nil {
		return relocationIfExistsFail, nil
	}
	return normalizeRelocationIfExists(ifExists)
}

func normalizeRelocationIfExists(ifExists string) (string, error) {
	switch ifExists {
	case "", relocationIfExistsFail:
		return relocationIfExistsFail, nil
	case relocationIfExistsSkip:
		return relocationIfExistsSkip, nil
	default:
		return "", invalidArgumentsErrorfWithDetails("invalid --if-exists %q (use fail or skip)", flagValueErrorDetails("if-exists", ifExists), ifExists)
	}
}

func relocationSkipIfDestinationExists(dbx files.Client, arg *files.RelocationArg, opts relocationOptions) (relocationResult, bool, error) {
	if opts.ifExists != relocationIfExistsSkip {
		return relocationResult{}, false, nil
	}
	return relocationSkippedResult(dbx, arg)
}

func relocationSkipAfterDestinationConflict(dbx files.Client, arg *files.RelocationArg, err error, opts relocationOptions) (relocationResult, bool) {
	if opts.ifExists != relocationIfExistsSkip || !isRelocationDestinationConflict(err) {
		return relocationResult{}, false
	}

	result, skipped, skipErr := relocationSkippedResult(dbx, arg)
	if skipErr != nil || !skipped {
		return relocationResult{}, false
	}
	return result, true
}

func relocationSkippedResult(dbx files.Client, arg *files.RelocationArg) (relocationResult, bool, error) {
	metadata, exists, err := getDestinationMetadata(dbx, arg.ToPath)
	if err != nil || !exists {
		return relocationResult{}, false, err
	}
	result, err := newRelocationResultFromMetadata(arg, metadata)
	if err != nil {
		return relocationResult{}, false, err
	}
	return result, true, nil
}

func isRelocationDestinationConflict(err error) bool {
	var copyErr files.CopyV2APIError
	if errors.As(err, &copyErr) && relocationErrorHasDestinationConflict(copyErr.EndpointError) {
		return true
	}

	var copyErrPtr *files.CopyV2APIError
	if errors.As(err, &copyErrPtr) && copyErrPtr != nil && relocationErrorHasDestinationConflict(copyErrPtr.EndpointError) {
		return true
	}

	var moveErr files.MoveV2APIError
	if errors.As(err, &moveErr) && relocationErrorHasDestinationConflict(moveErr.EndpointError) {
		return true
	}

	var moveErrPtr *files.MoveV2APIError
	if errors.As(err, &moveErrPtr) && moveErrPtr != nil && relocationErrorHasDestinationConflict(moveErrPtr.EndpointError) {
		return true
	}

	return false
}

func relocationErrorHasDestinationConflict(err *files.RelocationError) bool {
	return err != nil &&
		err.Tag == files.RelocationErrorTo &&
		isRelocationWriteConflict(err.To)
}

func isRelocationWriteConflict(err *files.WriteError) bool {
	return err != nil &&
		err.Tag == files.WriteErrorConflict &&
		err.Conflict != nil &&
		(err.Conflict.Tag == files.WriteConflictErrorFile ||
			err.Conflict.Tag == files.WriteConflictErrorFolder)
}

func relocationFailureDetails(fromPath, toPath string) map[string]any {
	return relocationErrorDetails(fromPath, toPath)
}

func relocationAggregateError(commandName, operation string, count int, failures []map[string]any) error {
	details := operationErrorDetails(operation)
	if len(failures) == 1 {
		details = mergeJSONErrorDetails(details, failures[0])
	}
	return commandFailedErrorfWithDetails("%s: %d error(s)", details, commandName, count)
}
