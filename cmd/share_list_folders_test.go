package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/spf13/cobra"
)

type mockSharedFolderClient struct {
	listFoldersFn         func(arg *sharing.ListFoldersArgs) (*sharing.ListFoldersResult, error)
	listFoldersContinueFn func(arg *sharing.ListFoldersContinueArg) (*sharing.ListFoldersResult, error)
}

func (m *mockSharedFolderClient) ListFolders(arg *sharing.ListFoldersArgs) (*sharing.ListFoldersResult, error) {
	if m.listFoldersFn != nil {
		return m.listFoldersFn(arg)
	}
	return sharing.NewListFoldersResult(nil), nil
}

func (m *mockSharedFolderClient) ListFoldersContinue(arg *sharing.ListFoldersContinueArg) (*sharing.ListFoldersResult, error) {
	if m.listFoldersContinueFn != nil {
		return m.listFoldersContinueFn(arg)
	}
	return sharing.NewListFoldersResult(nil), nil
}

func TestShareListFoldersTextUsesCommandOutput(t *testing.T) {
	stubSharedFolderClient(t, &mockSharedFolderClient{
		listFoldersFn: func(arg *sharing.ListFoldersArgs) (*sharing.ListFoldersResult, error) {
			return sharing.NewListFoldersResult([]*sharing.SharedFolderMetadata{
				testSharedFolder("/docs", "https://example.com/docs"),
			}), nil
		},
	})

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	if err := shareListFolders(cmd, nil); err != nil {
		t.Fatalf("shareListFolders error: %v", err)
	}

	if got, want := stdout.String(), "/docs\thttps://example.com/docs\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestShareListFoldersJSONOutputsSharedFolders(t *testing.T) {
	invited := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	stubSharedFolderClient(t, &mockSharedFolderClient{
		listFoldersFn: func(arg *sharing.ListFoldersArgs) (*sharing.ListFoldersResult, error) {
			folder := testSharedFolder("/docs", "https://example.com/docs")
			folder.Name = "Docs"
			folder.SharedFolderId = "sfid:docs"
			folder.AccessType = &sharing.AccessLevel{Tagged: dropbox.Tagged{Tag: sharing.AccessLevelOwner}}
			folder.IsInsideTeamFolder = true
			folder.IsTeamFolder = true
			folder.OwnerDisplayNames = []string{"Owner One"}
			folder.ParentSharedFolderId = "sfid:parent"
			folder.ParentFolderName = "Parent"
			folder.TimeInvited = dropbox.DBXTime(invited)
			folder.AccessInheritance = &sharing.AccessInheritance{Tagged: dropbox.Tagged{Tag: sharing.AccessInheritanceInherit}}
			return sharing.NewListFoldersResult([]*sharing.SharedFolderMetadata{folder}), nil
		},
	})

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	setShareLinkOutputJSON(t, cmd)

	if err := shareListFolders(cmd, nil); err != nil {
		t.Fatalf("shareListFolders error: %v", err)
	}

	got := decodeShareLinkOperationOutput[map[string]any, shareFolderJSONMetadata](t, stdout.Bytes())
	if len(got.Input) != 0 {
		t.Fatalf("input = %#v, want empty object", got.Input)
	}
	if len(got.Results) != 1 {
		t.Fatalf("results length = %d, want 1", len(got.Results))
	}
	result := got.Results[0]
	if result.Status != shareFolderJSONStatusListed || result.Kind != shareFolderJSONKindFolder {
		t.Fatalf("result = %#v, want listed shared_folder", result)
	}
	folder := result.Result
	if folder.Type != shareFolderJSONKindFolder || folder.Name != "Docs" || folder.PathLower != "/docs" {
		t.Fatalf("folder = %#v, want docs shared folder metadata", folder)
	}
	if folder.SharedFolderID != "sfid:docs" || folder.PreviewURL != "https://example.com/docs" {
		t.Fatalf("folder = %#v, want id and preview URL", folder)
	}
	if folder.AccessType != sharing.AccessLevelOwner || folder.AccessInheritance != sharing.AccessInheritanceInherit {
		t.Fatalf("folder = %#v, want access tags", folder)
	}
	if !folder.IsInsideTeamFolder || !folder.IsTeamFolder {
		t.Fatalf("folder = %#v, want team folder flags", folder)
	}
	if len(folder.OwnerDisplayNames) != 1 || folder.OwnerDisplayNames[0] != "Owner One" {
		t.Fatalf("owners = %#v, want owner display name", folder.OwnerDisplayNames)
	}
	if folder.ParentSharedFolderID != "sfid:parent" || folder.ParentFolderName != "Parent" {
		t.Fatalf("folder = %#v, want parent fields", folder)
	}
	if folder.TimeInvited == nil || *folder.TimeInvited != "2026-06-24T12:00:00Z" {
		t.Fatalf("time_invited = %#v, want RFC3339 time", folder.TimeInvited)
	}
	if strings.Contains(stdout.String(), `"entries"`) {
		t.Fatalf("JSON output = %s, want operation results and no entries key", stdout.String())
	}
}

func TestShareListFoldersJSONPaginates(t *testing.T) {
	continueCalled := false
	stubSharedFolderClient(t, &mockSharedFolderClient{
		listFoldersFn: func(arg *sharing.ListFoldersArgs) (*sharing.ListFoldersResult, error) {
			result := sharing.NewListFoldersResult([]*sharing.SharedFolderMetadata{
				testSharedFolder("/one", "https://example.com/one"),
			})
			result.Cursor = "cursor-1"
			return result, nil
		},
		listFoldersContinueFn: func(arg *sharing.ListFoldersContinueArg) (*sharing.ListFoldersResult, error) {
			continueCalled = true
			if arg.Cursor != "cursor-1" {
				t.Fatalf("continue cursor = %q, want cursor-1", arg.Cursor)
			}
			return sharing.NewListFoldersResult([]*sharing.SharedFolderMetadata{
				testSharedFolder("/two", "https://example.com/two"),
			}), nil
		},
	})

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	setShareLinkOutputJSON(t, cmd)

	if err := shareListFolders(cmd, nil); err != nil {
		t.Fatalf("shareListFolders error: %v", err)
	}
	if !continueCalled {
		t.Fatal("ListFoldersContinue was not called")
	}
	got := decodeShareLinkOperationOutput[shareFolderListInput, shareFolderJSONMetadata](t, stdout.Bytes())
	if len(got.Results) != 2 {
		t.Fatalf("results length = %d, want 2", len(got.Results))
	}
	if got.Results[0].Result.PathLower != "/one" || got.Results[1].Result.PathLower != "/two" {
		t.Fatalf("results = %#v, want paginated shared folders", got.Results)
	}
}

func TestShareListFoldersJSONErrorWritesNoSuccessOutput(t *testing.T) {
	stubSharedFolderClient(t, &mockSharedFolderClient{
		listFoldersFn: func(arg *sharing.ListFoldersArgs) (*sharing.ListFoldersResult, error) {
			return nil, errors.New("list failed")
		},
	})

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	setShareLinkOutputJSON(t, cmd)

	if err := shareListFolders(cmd, nil); err == nil {
		t.Fatal("expected shareListFolders error")
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty output on error", got)
	}
}

func TestShareListFoldersCommandSupportsStructuredOutput(t *testing.T) {
	if !commandSupportsStructuredOutput(shareListFoldersCmd) {
		t.Fatal("share list folder command should support structured output")
	}
}

func stubSharedFolderClient(t *testing.T, client sharedFolderClient) {
	t.Helper()

	orig := newSharedFolderClient
	newSharedFolderClient = func(_ dropbox.Config) sharedFolderClient { return client }
	t.Cleanup(func() { newSharedFolderClient = orig })
}

func testSharedFolder(pathLower, previewURL string) *sharing.SharedFolderMetadata {
	return &sharing.SharedFolderMetadata{
		SharedFolderMetadataBase: sharing.SharedFolderMetadataBase{
			PathLower: pathLower,
		},
		Name:           strings.TrimPrefix(pathLower, "/"),
		PreviewUrl:     previewURL,
		SharedFolderId: "sfid:" + strings.TrimPrefix(pathLower, "/"),
	}
}
