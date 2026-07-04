package cmd

import (
	"context"
	"io"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

type filesClient interface {
	CopyV2Context(context.Context, *files.RelocationArg) (*files.RelocationResult, error)
	CreateFolderV2Context(context.Context, *files.CreateFolderArg) (*files.CreateFolderResult, error)
	DeleteV2Context(context.Context, *files.DeleteArg) (*files.DeleteResult, error)
	DownloadContext(context.Context, *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error)
	GetMetadataContext(context.Context, *files.GetMetadataArg) (files.IsMetadata, error)
	ListFolderContext(context.Context, *files.ListFolderArg) (*files.ListFolderResult, error)
	ListFolderContinueContext(context.Context, *files.ListFolderContinueArg) (*files.ListFolderResult, error)
	ListRevisionsContext(context.Context, *files.ListRevisionsArg) (*files.ListRevisionsResult, error)
	MoveV2Context(context.Context, *files.RelocationArg) (*files.RelocationResult, error)
	PermanentlyDeleteContext(context.Context, *files.DeleteArg) error
	RestoreContext(context.Context, *files.RestoreArg) (*files.FileMetadata, error)
	SearchV2Context(context.Context, *files.SearchV2Arg) (*files.SearchV2Result, error)
	SearchContinueV2Context(context.Context, *files.SearchV2ContinueArg) (*files.SearchV2Result, error)
	UploadContext(context.Context, *files.UploadArg, io.Reader) (*files.FileMetadata, error)
	UploadSessionAppendV2Context(context.Context, *files.UploadSessionAppendArg, io.Reader) error
	UploadSessionFinishContext(context.Context, *files.UploadSessionFinishArg, io.Reader) (*files.FileMetadata, error)
	UploadSessionStartContext(context.Context, *files.UploadSessionStartArg, io.Reader) (*files.UploadSessionStartResult, error)
}

var filesNewFunc = func(cfg dropbox.Config) filesClient {
	return files.NewContext(cfg)
}
