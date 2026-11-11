package cmd

import (
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

const listingOutputE2EEnv = "DBXCLI_TEST_LISTING_OUTPUT_FIXTURE"

func installListingOutputFixture(t *testing.T) {
	t.Helper()
	stubUsersClient(t, &mockUsersClient{})
	stubFilesClient(t, &mockFilesClient{
		listFolderFn: func(*files.ListFolderArg) (*files.ListFolderResult, error) {
			return &files.ListFolderResult{Entries: []files.IsMetadata{
				&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/one"}},
				&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/two words"}},
				&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/three.zip"}},
				&files.FolderMetadata{Metadata: files.Metadata{PathDisplay: "/four"}},
				&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/five"}},
			}}, nil
		},
		searchV2Fn: func(*files.SearchV2Arg) (*files.SearchV2Result, error) {
			return files.NewSearchV2Result([]*files.SearchMatchV2{
				searchMatch(&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/zeta"}}),
				searchMatch(&files.FileMetadata{Metadata: files.Metadata{PathDisplay: "/alpha two"}}),
			}, false), nil
		},
	})
}

func TestListingTextOutputEndToEnd(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"ls", []string{"ls"}, "/one /two words /three.zip /four\n/five\n"},
		{"search", []string{"search", "query"}, "/zeta\n/alpha two\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr := executeExitTestSubprocess(t, tt.args, map[string]string{
				envAccessToken:      "test-token",
				listingOutputE2EEnv: "1",
			})
			if code != 0 || stderr != "" {
				t.Fatalf("exit code = %d, stderr = %q", code, stderr)
			}
			if stdout != tt.want {
				t.Fatalf("output = %q, want %q", stdout, tt.want)
			}
		})
	}
}

func TestListingLongOutputHasNoTrailingPaddingEndToEnd(t *testing.T) {
	for _, args := range [][]string{{"ls", "-l"}, {"search", "-l", "query"}} {
		t.Run(args[0], func(t *testing.T) {
			code, stdout, stderr := executeExitTestSubprocess(t, args, map[string]string{
				envAccessToken:      "test-token",
				listingOutputE2EEnv: "1",
			})
			if code != 0 || stderr != "" {
				t.Fatalf("exit code = %d, stderr = %q", code, stderr)
			}
			for _, line := range strings.Split(strings.TrimSuffix(stdout, "\n"), "\n") {
				if strings.TrimRight(line, " \t") != line {
					t.Fatalf("line has trailing padding: %q", line)
				}
			}
		})
	}
}
