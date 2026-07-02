package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func TestJSONMetadataFromDropboxFile(t *testing.T) {
	clientModified := time.Date(2026, 6, 21, 10, 30, 0, 0, time.FixedZone("test", -7*60*60))
	serverModified := time.Date(2026, 6, 22, 11, 45, 0, 0, time.UTC)
	metadata := &files.FileMetadata{
		Metadata: files.Metadata{
			PathDisplay: "/Reports/File.txt",
			PathLower:   "/reports/file.txt",
		},
		Id:             "id:abc",
		Rev:            "rev123",
		Size:           0,
		ClientModified: dropbox.DBXTime(clientModified),
		ServerModified: dropbox.DBXTime(serverModified),
	}

	got, err := jsonMetadataFromDropbox(metadata)
	if err != nil {
		t.Fatal(err)
	}

	if got.Type != "file" {
		t.Fatalf("Type = %q, want file", got.Type)
	}
	if got.PathDisplay != "/Reports/File.txt" {
		t.Fatalf("PathDisplay = %q", got.PathDisplay)
	}
	if got.Size == nil || *got.Size != 0 {
		t.Fatalf("Size = %v, want pointer to 0", got.Size)
	}
	if got.ClientModified == nil || *got.ClientModified != "2026-06-21T17:30:00Z" {
		t.Fatalf("ClientModified = %v", got.ClientModified)
	}
	if got.ServerModified == nil || *got.ServerModified != "2026-06-22T11:45:00Z" {
		t.Fatalf("ServerModified = %v", got.ServerModified)
	}

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "is_downloadable") {
		t.Fatalf("encoded metadata exposes SDK-specific fields: %s", encoded)
	}
	if strings.Contains(string(encoded), "deleted") {
		t.Fatalf("encoded non-deleted metadata should omit deleted field: %s", encoded)
	}
}

func TestJSONMetadataFromDropboxFolder(t *testing.T) {
	metadata := &files.FolderMetadata{
		Metadata: files.Metadata{
			PathDisplay: "/Reports",
			PathLower:   "/reports",
		},
		Id: "id:folder",
	}

	got, err := jsonMetadataFromDropbox(metadata)
	if err != nil {
		t.Fatal(err)
	}

	if got.Type != "folder" {
		t.Fatalf("Type = %q, want folder", got.Type)
	}
	if got.ID != "id:folder" {
		t.Fatalf("ID = %q, want id:folder", got.ID)
	}
	if got.Size != nil {
		t.Fatalf("Size = %v, want nil for folder", got.Size)
	}
}

func TestJSONMetadataFromDropboxDeleted(t *testing.T) {
	metadata := &files.DeletedMetadata{
		Metadata: files.Metadata{
			PathDisplay: "/Reports/Old.txt",
			PathLower:   "/reports/old.txt",
		},
	}

	got, err := jsonMetadataFromDropbox(metadata)
	if err != nil {
		t.Fatal(err)
	}

	if got.Type != "deleted" {
		t.Fatalf("Type = %q, want deleted", got.Type)
	}
	if !got.Deleted {
		t.Fatal("Deleted = false, want true")
	}
}

func TestJSONMetadataFromDropboxRejectsUnknownMetadata(t *testing.T) {
	if _, err := jsonMetadataFromDropbox(nil); err == nil {
		t.Fatal("expected nil metadata to fail")
	}
}
