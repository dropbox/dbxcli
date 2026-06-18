package cmd

import (
	"io"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/async"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/file_properties"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

type mockFilesClient struct {
	downloadFn              func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error)
	uploadFn                func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error)
	uploadSessionStartFn    func(arg *files.UploadSessionStartArg, content io.Reader) (*files.UploadSessionStartResult, error)
	uploadSessionAppendV2Fn func(arg *files.UploadSessionAppendArg, content io.Reader) error
	uploadSessionFinishFn   func(arg *files.UploadSessionFinishArg, content io.Reader) (*files.FileMetadata, error)
	copyV2Fn                func(arg *files.RelocationArg) (*files.RelocationResult, error)
	createFolderV2Fn        func(arg *files.CreateFolderArg) (*files.CreateFolderResult, error)
	deleteV2Fn              func(arg *files.DeleteArg) (*files.DeleteResult, error)
	getMetadataFn           func(arg *files.GetMetadataArg) (files.IsMetadata, error)
	listFolderFn            func(arg *files.ListFolderArg) (*files.ListFolderResult, error)
	listFolderContinueFn    func(arg *files.ListFolderContinueArg) (*files.ListFolderResult, error)
	moveV2Fn                func(arg *files.RelocationArg) (*files.RelocationResult, error)
	permanentlyDeleteFn     func(arg *files.DeleteArg) error
	searchV2Fn              func(arg *files.SearchV2Arg) (*files.SearchV2Result, error)
	searchContinueV2Fn      func(arg *files.SearchV2ContinueArg) (*files.SearchV2Result, error)
}

func (m *mockFilesClient) Download(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
	if m.downloadFn != nil {
		return m.downloadFn(arg)
	}
	return nil, nil, nil
}

func (m *mockFilesClient) Upload(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
	if m.uploadFn != nil {
		return m.uploadFn(arg, content)
	}
	return nil, nil
}

func (m *mockFilesClient) UploadSessionStart(arg *files.UploadSessionStartArg, content io.Reader) (*files.UploadSessionStartResult, error) {
	if m.uploadSessionStartFn != nil {
		return m.uploadSessionStartFn(arg, content)
	}
	return nil, nil
}

func (m *mockFilesClient) UploadSessionAppendV2(arg *files.UploadSessionAppendArg, content io.Reader) error {
	if m.uploadSessionAppendV2Fn != nil {
		return m.uploadSessionAppendV2Fn(arg, content)
	}
	return nil
}

func (m *mockFilesClient) UploadSessionFinish(arg *files.UploadSessionFinishArg, content io.Reader) (*files.FileMetadata, error) {
	if m.uploadSessionFinishFn != nil {
		return m.uploadSessionFinishFn(arg, content)
	}
	return nil, nil
}

// Stubs for the rest of the interface
func (m *mockFilesClient) AlphaGetMetadata(arg *files.AlphaGetMetadataArg) (files.IsMetadata, error) {
	return nil, nil
}
func (m *mockFilesClient) AlphaUpload(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
	return nil, nil
}
func (m *mockFilesClient) CopyV2(arg *files.RelocationArg) (*files.RelocationResult, error) {
	if m.copyV2Fn != nil {
		return m.copyV2Fn(arg)
	}
	return nil, nil
}
func (m *mockFilesClient) Copy(arg *files.RelocationArg) (files.IsMetadata, error) {
	return nil, nil
}
func (m *mockFilesClient) CopyBatchV2(arg *files.RelocationBatchArgBase) (*files.RelocationBatchV2Launch, error) {
	return nil, nil
}
func (m *mockFilesClient) CopyBatch(arg *files.RelocationBatchArg) (*files.RelocationBatchLaunch, error) {
	return nil, nil
}
func (m *mockFilesClient) CopyBatchCheckV2(arg *async.PollArg) (*files.RelocationBatchV2JobStatus, error) {
	return nil, nil
}
func (m *mockFilesClient) CopyBatchCheck(arg *async.PollArg) (*files.RelocationBatchJobStatus, error) {
	return nil, nil
}
func (m *mockFilesClient) CopyReferenceGet(arg *files.GetCopyReferenceArg) (*files.GetCopyReferenceResult, error) {
	return nil, nil
}
func (m *mockFilesClient) CopyReferenceSave(arg *files.SaveCopyReferenceArg) (*files.SaveCopyReferenceResult, error) {
	return nil, nil
}
func (m *mockFilesClient) CreateFolderV2(arg *files.CreateFolderArg) (*files.CreateFolderResult, error) {
	if m.createFolderV2Fn != nil {
		return m.createFolderV2Fn(arg)
	}
	return nil, nil
}
func (m *mockFilesClient) CreateFolder(arg *files.CreateFolderArg) (*files.FolderMetadata, error) {
	return nil, nil
}
func (m *mockFilesClient) CreateFolderBatch(arg *files.CreateFolderBatchArg) (*files.CreateFolderBatchLaunch, error) {
	return nil, nil
}
func (m *mockFilesClient) CreateFolderBatchCheck(arg *async.PollArg) (*files.CreateFolderBatchJobStatus, error) {
	return nil, nil
}
func (m *mockFilesClient) DeleteV2(arg *files.DeleteArg) (*files.DeleteResult, error) {
	if m.deleteV2Fn != nil {
		return m.deleteV2Fn(arg)
	}
	return nil, nil
}
func (m *mockFilesClient) Delete(arg *files.DeleteArg) (files.IsMetadata, error) { return nil, nil }
func (m *mockFilesClient) DeleteBatch(arg *files.DeleteBatchArg) (*files.DeleteBatchLaunch, error) {
	return nil, nil
}
func (m *mockFilesClient) DeleteBatchCheck(arg *async.PollArg) (*files.DeleteBatchJobStatus, error) {
	return nil, nil
}
func (m *mockFilesClient) DownloadZip(arg *files.DownloadZipArg) (*files.DownloadZipResult, io.ReadCloser, error) {
	return nil, nil, nil
}
func (m *mockFilesClient) Export(arg *files.ExportArg) (*files.ExportResult, io.ReadCloser, error) {
	return nil, nil, nil
}
func (m *mockFilesClient) GetFileLockBatch(arg *files.LockFileBatchArg) (*files.LockFileBatchResult, error) {
	return nil, nil
}
func (m *mockFilesClient) GetMetadata(arg *files.GetMetadataArg) (files.IsMetadata, error) {
	if m.getMetadataFn != nil {
		return m.getMetadataFn(arg)
	}
	return nil, nil
}
func (m *mockFilesClient) GetPreview(arg *files.PreviewArg) (*files.FileMetadata, io.ReadCloser, error) {
	return nil, nil, nil
}
func (m *mockFilesClient) GetTemporaryLink(arg *files.GetTemporaryLinkArg) (*files.GetTemporaryLinkResult, error) {
	return nil, nil
}
func (m *mockFilesClient) GetTemporaryUploadLink(arg *files.GetTemporaryUploadLinkArg) (*files.GetTemporaryUploadLinkResult, error) {
	return nil, nil
}
func (m *mockFilesClient) GetThumbnail(arg *files.ThumbnailArg) (*files.FileMetadata, io.ReadCloser, error) {
	return nil, nil, nil
}
func (m *mockFilesClient) GetThumbnailV2(arg *files.ThumbnailV2Arg) (*files.PreviewResult, io.ReadCloser, error) {
	return nil, nil, nil
}
func (m *mockFilesClient) GetThumbnailBatch(arg *files.GetThumbnailBatchArg) (*files.GetThumbnailBatchResult, error) {
	return nil, nil
}
func (m *mockFilesClient) ListFolder(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
	if m.listFolderFn != nil {
		return m.listFolderFn(arg)
	}
	return nil, nil
}
func (m *mockFilesClient) ListFolderContinue(arg *files.ListFolderContinueArg) (*files.ListFolderResult, error) {
	if m.listFolderContinueFn != nil {
		return m.listFolderContinueFn(arg)
	}
	return nil, nil
}
func (m *mockFilesClient) ListFolderGetLatestCursor(arg *files.ListFolderArg) (*files.ListFolderGetLatestCursorResult, error) {
	return nil, nil
}
func (m *mockFilesClient) ListFolderLongpoll(arg *files.ListFolderLongpollArg) (*files.ListFolderLongpollResult, error) {
	return nil, nil
}
func (m *mockFilesClient) ListRevisions(arg *files.ListRevisionsArg) (*files.ListRevisionsResult, error) {
	return nil, nil
}
func (m *mockFilesClient) LockFileBatch(arg *files.LockFileBatchArg) (*files.LockFileBatchResult, error) {
	return nil, nil
}
func (m *mockFilesClient) MoveV2(arg *files.RelocationArg) (*files.RelocationResult, error) {
	if m.moveV2Fn != nil {
		return m.moveV2Fn(arg)
	}
	return nil, nil
}
func (m *mockFilesClient) Move(arg *files.RelocationArg) (files.IsMetadata, error) {
	return nil, nil
}
func (m *mockFilesClient) MoveBatchV2(arg *files.MoveBatchArg) (*files.RelocationBatchV2Launch, error) {
	return nil, nil
}
func (m *mockFilesClient) MoveBatch(arg *files.RelocationBatchArg) (*files.RelocationBatchLaunch, error) {
	return nil, nil
}
func (m *mockFilesClient) MoveBatchCheckV2(arg *async.PollArg) (*files.RelocationBatchV2JobStatus, error) {
	return nil, nil
}
func (m *mockFilesClient) MoveBatchCheck(arg *async.PollArg) (*files.RelocationBatchJobStatus, error) {
	return nil, nil
}
func (m *mockFilesClient) PaperCreate(arg *files.PaperCreateArg, content io.Reader) (*files.PaperCreateResult, error) {
	return nil, nil
}
func (m *mockFilesClient) PaperUpdate(arg *files.PaperUpdateArg, content io.Reader) (*files.PaperUpdateResult, error) {
	return nil, nil
}
func (m *mockFilesClient) PermanentlyDelete(arg *files.DeleteArg) error {
	if m.permanentlyDeleteFn != nil {
		return m.permanentlyDeleteFn(arg)
	}
	return nil
}
func (m *mockFilesClient) PropertiesAdd(arg *file_properties.AddPropertiesArg) error {
	return nil
}
func (m *mockFilesClient) PropertiesOverwrite(arg *file_properties.OverwritePropertyGroupArg) error {
	return nil
}
func (m *mockFilesClient) PropertiesRemove(arg *file_properties.RemovePropertiesArg) error {
	return nil
}
func (m *mockFilesClient) PropertiesTemplateGet(arg *file_properties.GetTemplateArg) (*file_properties.GetTemplateResult, error) {
	return nil, nil
}
func (m *mockFilesClient) PropertiesTemplateList() (*file_properties.ListTemplateResult, error) {
	return nil, nil
}
func (m *mockFilesClient) PropertiesUpdate(arg *file_properties.UpdatePropertiesArg) error {
	return nil
}
func (m *mockFilesClient) Restore(arg *files.RestoreArg) (*files.FileMetadata, error) {
	return nil, nil
}
func (m *mockFilesClient) SaveUrl(arg *files.SaveUrlArg) (*files.SaveUrlResult, error) {
	return nil, nil
}
func (m *mockFilesClient) SaveUrlCheckJobStatus(arg *async.PollArg) (*files.SaveUrlJobStatus, error) {
	return nil, nil
}
func (m *mockFilesClient) Search(arg *files.SearchArg) (*files.SearchResult, error) {
	return nil, nil
}
func (m *mockFilesClient) SearchV2(arg *files.SearchV2Arg) (*files.SearchV2Result, error) {
	if m.searchV2Fn != nil {
		return m.searchV2Fn(arg)
	}
	return nil, nil
}
func (m *mockFilesClient) SearchContinueV2(arg *files.SearchV2ContinueArg) (*files.SearchV2Result, error) {
	if m.searchContinueV2Fn != nil {
		return m.searchContinueV2Fn(arg)
	}
	return nil, nil
}
func (m *mockFilesClient) TagsAdd(arg *files.AddTagArg) error { return nil }
func (m *mockFilesClient) TagsGet(arg *files.GetTagsArg) (*files.GetTagsResult, error) {
	return nil, nil
}
func (m *mockFilesClient) TagsRemove(arg *files.RemoveTagArg) error { return nil }
func (m *mockFilesClient) UnlockFileBatch(arg *files.UnlockFileBatchArg) (*files.LockFileBatchResult, error) {
	return nil, nil
}
func (m *mockFilesClient) UploadSessionAppend(arg *files.UploadSessionCursor, content io.Reader) error {
	return nil
}
func (m *mockFilesClient) UploadSessionStartBatch(arg *files.UploadSessionStartBatchArg) (*files.UploadSessionStartBatchResult, error) {
	return nil, nil
}
func (m *mockFilesClient) UploadSessionFinishBatchV2(arg *files.UploadSessionFinishBatchArg) (*files.UploadSessionFinishBatchResult, error) {
	return nil, nil
}
func (m *mockFilesClient) UploadSessionFinishBatch(arg *files.UploadSessionFinishBatchArg) (*files.UploadSessionFinishBatchLaunch, error) {
	return nil, nil
}
func (m *mockFilesClient) UploadSessionFinishBatchCheck(arg *async.PollArg) (*files.UploadSessionFinishBatchJobStatus, error) {
	return nil, nil
}
