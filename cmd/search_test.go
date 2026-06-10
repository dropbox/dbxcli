package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func TestSearchArgValidation(t *testing.T) {
	err := search(searchCmd, []string{})
	if err == nil {
		t.Error("expected error for no args")
	}
}

func TestSearchPathScopeValidation(t *testing.T) {
	err := search(searchCmd, []string{"query", "no-slash"})
	if err == nil {
		t.Error("expected error for path-scope without leading slash")
	}
}

func TestRenderSearchResultsSeparatesMatchesWithNewlines(t *testing.T) {
	res := files.NewSearchResult([]*files.SearchMatch{
		files.NewSearchMatch(nil, &files.FileMetadata{
			Metadata: files.Metadata{PathDisplay: "/first.txt"},
		}),
		files.NewSearchMatch(nil, &files.FolderMetadata{
			Metadata: files.Metadata{PathDisplay: "/second"},
		}),
	}, false, 0)

	var out bytes.Buffer
	if err := renderSearchResults(&out, res, listOptions{long: false}); err != nil {
		t.Fatalf("renderSearchResults returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected one output line per match, got %d lines in %q", len(lines), out.String())
	}

	if got, want := strings.TrimSpace(lines[0]), "/first.txt"; got != want {
		t.Errorf("first rendered match = %q, want %q", got, want)
	}
	if got, want := strings.TrimSpace(lines[1]), "/second"; got != want {
		t.Errorf("second rendered match = %q, want %q", got, want)
	}
}
