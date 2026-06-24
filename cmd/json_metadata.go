package cmd

import (
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

func jsonMetadataFromDropbox(metadata files.IsMetadata) jsonMetadata {
	switch m := metadata.(type) {
	case *files.FileMetadata:
		if m == nil {
			return jsonMetadata{Type: "unknown"}
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
		}
	case *files.FolderMetadata:
		if m == nil {
			return jsonMetadata{Type: "unknown"}
		}
		return jsonMetadata{
			Type:        "folder",
			PathDisplay: m.PathDisplay,
			PathLower:   m.PathLower,
			ID:          m.Id,
		}
	case *files.DeletedMetadata:
		if m == nil {
			return jsonMetadata{Type: "unknown"}
		}
		return jsonMetadata{
			Type:        "deleted",
			PathDisplay: m.PathDisplay,
			PathLower:   m.PathLower,
			Deleted:     true,
		}
	default:
		return jsonMetadata{Type: "unknown"}
	}
}

func jsonMetadataListFromDropbox(entries []files.IsMetadata) []jsonMetadata {
	result := make([]jsonMetadata, 0, len(entries))
	for _, entry := range entries {
		result = append(result, jsonMetadataFromDropbox(entry))
	}
	return result
}

func jsonTime(t time.Time) *string {
	if t.IsZero() {
		return nil
	}
	value := t.UTC().Format(time.RFC3339)
	return &value
}
