package cmd

import (
	"fmt"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

type jsonMetadata struct {
	Type           string  `json:"type"`
	PathDisplay    string  `json:"path_display,omitempty"`
	PathLower      string  `json:"path_lower,omitempty"`
	ID             string  `json:"id,omitempty"`
	Rev            string  `json:"rev,omitempty"`
	Size           *uint64 `json:"size,omitempty"`
	ServerModified *string `json:"server_modified,omitempty"`
	ClientModified *string `json:"client_modified,omitempty"`
	Deleted        bool    `json:"deleted,omitempty"`
}

func jsonMetadataFromDropbox(metadata files.IsMetadata) (jsonMetadata, error) {
	switch m := metadata.(type) {
	case *files.FileMetadata:
		if m == nil {
			return jsonMetadata{}, fmt.Errorf("unexpected nil Dropbox file metadata")
		}
		size := m.Size
		return jsonMetadata{
			Type:           "file",
			PathDisplay:    m.PathDisplay,
			PathLower:      m.PathLower,
			ID:             m.Id,
			Rev:            m.Rev,
			Size:           &size,
			ServerModified: jsonTime(m.ServerModified),
			ClientModified: jsonTime(m.ClientModified),
		}, nil
	case *files.FolderMetadata:
		if m == nil {
			return jsonMetadata{}, fmt.Errorf("unexpected nil Dropbox folder metadata")
		}
		return jsonMetadata{
			Type:        "folder",
			PathDisplay: m.PathDisplay,
			PathLower:   m.PathLower,
			ID:          m.Id,
		}, nil
	case *files.DeletedMetadata:
		if m == nil {
			return jsonMetadata{}, fmt.Errorf("unexpected nil Dropbox deleted metadata")
		}
		return jsonMetadata{
			Type:        "deleted",
			PathDisplay: m.PathDisplay,
			PathLower:   m.PathLower,
			Deleted:     true,
		}, nil
	default:
		return jsonMetadata{}, fmt.Errorf("unexpected Dropbox metadata type %T", metadata)
	}
}

func jsonMetadataListFromDropbox(entries []files.IsMetadata) ([]jsonMetadata, error) {
	result := make([]jsonMetadata, 0, len(entries))
	for _, entry := range entries {
		metadata, err := jsonMetadataFromDropbox(entry)
		if err != nil {
			return nil, err
		}
		result = append(result, metadata)
	}
	return result, nil
}

func jsonTime(t time.Time) *string {
	if t.IsZero() {
		return nil
	}
	value := t.UTC().Format(time.RFC3339)
	return &value
}
