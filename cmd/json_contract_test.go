package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/dropbox/dbxcli/internal/output"
	"github.com/spf13/cobra"
)

func TestStructuredOutputCommandAudit(t *testing.T) {
	got := structuredOutputCommandPathsWithVersion()
	sort.Strings(got)

	want := expectedStructuredOutputCommands()
	sort.Strings(want)

	if len(got) != len(want) {
		t.Fatalf("structured commands = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("structured commands = %v, want %v", got, want)
		}
	}
}

func TestStructuredOutputSuccessFixtureAudit(t *testing.T) {
	contractCommands := jsonContractCommandPathsWithVersion()
	contractSet := make(map[string]bool, len(contractCommands))
	for _, command := range contractCommands {
		contractSet[command] = true
	}

	fixtures := jsonSuccessFixtureCoverage()
	for _, command := range contractCommands {
		if fixtures[command].file == "" {
			t.Errorf("JSON contract command %q has no success JSON fixture coverage entry", command)
		}
	}
	for command, fixture := range fixtures {
		if fixture.file == "" {
			t.Errorf("fixture coverage for %q has empty file", command)
		}
		if !contractSet[command] {
			t.Errorf("fixture coverage includes non-contract command %q", command)
		}
		if len(fixture.tests) == 0 {
			t.Errorf("fixture coverage for %q has no test functions", command)
		}
		source, err := os.ReadFile(fixture.file)
		if err != nil {
			t.Errorf("read fixture file for %q: %v", command, err)
			continue
		}
		for _, testName := range fixture.tests {
			if !strings.Contains(string(source), "func "+testName+"(") {
				t.Errorf("fixture coverage for %q references missing %s in %s", command, testName, fixture.file)
			}
		}
	}
}

func TestStructuredOutputGoldenSchemaAudit(t *testing.T) {
	contract := loadJSONGoldenContract(t)

	assertStringSliceMapEqual(t, "schema definitions", contract.Definitions, jsonContractDefinitions())

	contractCommands := jsonContractCommandPathsWithVersion()
	contractSet := make(map[string]bool, len(contractCommands))
	for _, command := range contractCommands {
		contractSet[command] = true
	}

	want := jsonCommandSchemas()
	for _, command := range contractCommands {
		gotSchema, ok := contract.Commands[command]
		if !ok {
			t.Errorf("JSON contract command %q has no golden schema", command)
			continue
		}
		wantSchema, ok := want[command]
		if !ok {
			t.Errorf("JSON contract command %q has no code-derived schema", command)
			continue
		}
		assertGoldenCommandSchemaEqual(t, command, gotSchema, wantSchema)
		assertGoldenCommandSchemaReferences(t, command, gotSchema, contract.Definitions)
		assertGoldenCommandStatuses(t, command, gotSchema)
	}
	for command, schema := range contract.Commands {
		if !contractSet[command] {
			t.Errorf("golden schema includes non-contract command %q", command)
		}
		assertGoldenCommandSchemaReferences(t, command, schema, contract.Definitions)
		assertGoldenCommandStatuses(t, command, schema)
	}
	for command := range want {
		if !contractSet[command] {
			t.Errorf("code-derived schema includes non-contract command %q", command)
		}
	}
}

func TestStructuredOutputGoldenSuccessOutputAudit(t *testing.T) {
	fixtures := loadJSONGoldenSuccessOutputs(t)
	examples := jsonGoldenSuccessOutputExamples()

	contractCommands := jsonContractCommandPathsWithVersion()
	contractSet := make(map[string]bool, len(contractCommands))
	for _, command := range contractCommands {
		contractSet[command] = true
	}

	for _, command := range contractCommands {
		fixture, ok := fixtures[command]
		if !ok {
			t.Errorf("JSON contract command %q has no golden success output", command)
			continue
		}
		example, ok := examples[command]
		if !ok {
			t.Errorf("JSON contract command %q has no code-derived success output example", command)
			continue
		}
		assertGoldenJSONEqual(t, command, fixture, example)
		assertGoldenSuccessOutputStatuses(t, command, fixture)
	}
	for command := range fixtures {
		if !contractSet[command] {
			t.Errorf("golden success output includes non-contract command %q", command)
		}
		assertGoldenSuccessOutputStatuses(t, command, fixtures[command])
	}
	for command := range examples {
		if !contractSet[command] {
			t.Errorf("code-derived success output example includes non-contract command %q", command)
		}
	}
}

func TestPublicJSONSchemaFiles(t *testing.T) {
	tests := []struct {
		file       string
		ok         bool
		required   []string
		properties []string
	}{
		{
			file:       "../docs/json-schema/v1/success.schema.json",
			ok:         true,
			required:   []string{"ok", "schema_version", "command", "input", "results", "warnings"},
			properties: []string{"ok", "schema_version", "command", "input", "results", "warnings"},
		},
		{
			file:       "../docs/json-schema/v1/error.schema.json",
			ok:         false,
			required:   []string{"ok", "schema_version", "command", "error", "warnings"},
			properties: []string{"ok", "schema_version", "command", "error", "warnings"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			schema := loadPublicJSONSchema(t, tt.file)
			if schema.Schema == "" {
				t.Fatalf("%s has no $schema", tt.file)
			}
			if schema.ID == "" {
				t.Fatalf("%s has no $id", tt.file)
			}
			if schema.Type != "object" {
				t.Fatalf("%s type = %q, want object", tt.file, schema.Type)
			}
			if got, want := schema.Properties["ok"].Const, tt.ok; got != want {
				t.Fatalf("%s ok const = %v, want %v", tt.file, got, want)
			}
			if got, want := schema.Properties["schema_version"].Const, jsonSchemaVersion; got != want {
				t.Fatalf("%s schema_version const = %v, want %v", tt.file, got, want)
			}
			assertStringSliceEqual(t, tt.file+" required", schema.Required, tt.required)
			assertStringSliceEqual(t, tt.file+" properties", mapKeys(schema.Properties), tt.properties)
			if tt.ok {
				assertStringSliceEqual(t, tt.file+" result required", schema.Defs["result"].Required, []string{"status", "kind", "input", "result"})
			} else {
				errorSchema := schema.Properties["error"]
				codeSchema := errorSchema.Properties["code"]
				assertStringSliceEqual(t, tt.file+" error code enum", codeSchema.Enum, expectedJSONErrorCodes())
				if _, ok := errorSchema.Properties["details"]; !ok {
					t.Fatalf("%s error schema missing details property", tt.file)
				}
			}
		})
	}
}

func TestPublicJSONCommandCatalogMatchesGoldenContract(t *testing.T) {
	got := loadJSONContractFile(t, "../docs/json-schema/v1/commands.json", "public command catalog")
	want := loadJSONGoldenContract(t)
	if reflect.DeepEqual(got, want) {
		return
	}

	gotJSON, _ := json.MarshalIndent(got, "", "  ")
	wantJSON, _ := json.MarshalIndent(want, "", "  ")
	t.Fatalf("public command catalog = %s, want %s", gotJSON, wantJSON)
}

func structuredOutputCommandPathsWithVersion() []string {
	paths := structuredOutputCommandPaths(RootCmd)
	return append(paths, NewVersionCommand("test").Name())
}

func jsonContractCommandPathsWithVersion() []string {
	paths := structuredOutputCommandPathsWithVersion()
	return append(paths, "help")
}

func expectedStructuredOutputCommands() []string {
	return []string{
		"account",
		"cp",
		"du",
		"get",
		"logout",
		"ls",
		"mkdir",
		"mv",
		"put",
		"restore",
		"revs",
		"rm",
		"search",
		"share list folder",
		"share list link",
		"share-link create",
		"share-link download",
		"share-link info",
		"share-link list",
		"share-link revoke",
		"share-link update",
		"team add-member",
		"team info",
		"team list-groups",
		"team list-members",
		"team remove-member",
		"version",
	}
}

type jsonSuccessFixture struct {
	file  string
	tests []string
}

type jsonGoldenContract struct {
	Definitions map[string][]string                `json:"definitions"`
	Commands    map[string]jsonGoldenCommandSchema `json:"commands"`
}

type jsonGoldenCommandSchema struct {
	TopLevel      string   `json:"top_level"`
	ResultWrapper string   `json:"result_wrapper"`
	Input         string   `json:"input"`
	ResultInput   *string  `json:"result_input"`
	Result        string   `json:"result"`
	Statuses      []string `json:"statuses"`
	Kinds         []string `json:"kinds"`
	Warnings      []string `json:"warnings"`
}

type publicJSONSchema struct {
	Schema     string                              `json:"$schema"`
	ID         string                              `json:"$id"`
	Type       string                              `json:"type"`
	Required   []string                            `json:"required"`
	Properties map[string]publicJSONSchemaProperty `json:"properties"`
	Defs       map[string]publicJSONSchema         `json:"$defs"`
}

type publicJSONSchemaProperty struct {
	Const      any                                 `json:"const"`
	Enum       []string                            `json:"enum"`
	Properties map[string]publicJSONSchemaProperty `json:"properties"`
}

func jsonSuccessFixtureCoverage() map[string]jsonSuccessFixture {
	return map[string]jsonSuccessFixture{
		"account": {
			file:  "account_test.go",
			tests: []string{"TestAccountCurrentJSONOutputsAccount", "TestAccountLookupJSONUsesAccountID"},
		},
		"cp": {
			file:  "cp_test.go",
			tests: []string{"TestCpJSONOutputsRelocationResults", "TestCpJSONMultipleSourcesOutputsMultipleResults"},
		},
		"du": {
			file:  "du_test.go",
			tests: []string{"TestDuJSONIndividualAllocation", "TestDuJSONTeamAllocation"},
		},
		"get": {
			file:  "get_test.go",
			tests: []string{"TestGetJSONFileOutputsDownloadedResult", "TestGetJSONRecursiveOutputsDirectoryAndFileResults"},
		},
		"help": {
			file:  "help_json_test.go",
			tests: []string{"TestJSONHelpSupportedForms", "TestJSONHelpManifestFields"},
		},
		"ls": {
			file:  "ls_test.go",
			tests: []string{"TestLsJSONListsResultsAndInput", "TestLsJSONDeletedEntryIsStructured"},
		},
		"logout": {
			file:  "logout_test.go",
			tests: []string{"TestLogoutJSONReturnsLoggedOut", "TestLogoutJSONReturnsAlreadyLoggedOut", "TestLogoutJSONWarnsOnRemoteRevokeFailureAfterRemovingCredentials"},
		},
		"mkdir": {
			file:  "mkdir_test.go",
			tests: []string{"TestMkdirJSONOutputsCreatedFolder", "TestMkdirJSONParentsReturnsExistingFolderMetadata"},
		},
		"mv": {
			file:  "mv_test.go",
			tests: []string{"TestMvJSONOutputsRelocationResults", "TestMvJSONMultipleSourcesOutputsMultipleResults"},
		},
		"put": {
			file:  "put_test.go",
			tests: []string{"TestPutJSONSingleFileOutputsUploadedResult", "TestPutJSONRecursiveOutputsDirectoryAndFileResults"},
		},
		"restore": {
			file:  "restore_test.go",
			tests: []string{"TestRestoreJSONOutputsInputAndMetadata"},
		},
		"revs": {
			file:  "revs_test.go",
			tests: []string{"TestRevsJSONOutputsInputAndResults"},
		},
		"rm": {
			file:  "rm_test.go",
			tests: []string{"TestRmJSONDeletesFile", "TestRmJSONMultipleTargets"},
		},
		"search": {
			file:  "search_test.go",
			tests: []string{"TestSearchJSONOutputsInputAndResults", "TestSearchJSONOmitsPathWithoutScope"},
		},
		"share list folder": {
			file:  "share_list_folders_test.go",
			tests: []string{"TestShareListFoldersJSONOutputsSharedFolders", "TestShareListFoldersJSONPaginates"},
		},
		"share list link": {
			file:  "share_link_json_test.go",
			tests: []string{"TestDeprecatedShareListLinkJSONIncludesWarning"},
		},
		"share-link create": {
			file:  "share_link_json_test.go",
			tests: []string{"TestShareLinkCreateJSONOutputsLinkMetadata"},
		},
		"share-link download": {
			file:  "share_link_json_test.go",
			tests: []string{"TestShareLinkDownloadJSONOutputsTargetAndMetadata"},
		},
		"share-link info": {
			file:  "share_link_json_test.go",
			tests: []string{"TestShareLinkInfoJSONOutputsPermissions"},
		},
		"share-link list": {
			file:  "share_link_json_test.go",
			tests: []string{"TestShareLinkListJSONOutputsResultsAndInput"},
		},
		"share-link revoke": {
			file:  "share_link_json_test.go",
			tests: []string{"TestShareLinkRevokeJSONOutputsRevokedURL"},
		},
		"share-link update": {
			file:  "share_link_json_test.go",
			tests: []string{"TestShareLinkUpdateJSONOutputsUpdatedMetadata"},
		},
		"team add-member": {
			file:  "team_json_test.go",
			tests: []string{"TestTeamAddMemberJSONOutputsMutationResult"},
		},
		"team info": {
			file:  "team_json_test.go",
			tests: []string{"TestTeamInfoJSONOutputsTeamInfo"},
		},
		"team list-groups": {
			file:  "team_json_test.go",
			tests: []string{"TestTeamListGroupsJSONPaginates"},
		},
		"team list-members": {
			file:  "team_json_test.go",
			tests: []string{"TestTeamListMembersJSONPaginates"},
		},
		"team remove-member": {
			file:  "team_json_test.go",
			tests: []string{"TestTeamRemoveMemberJSONOutputsMutationResult"},
		},
		"version": {
			file:  "version_test.go",
			tests: []string{"TestVersionJSONOutputsVersionInfo"},
		},
	}
}

func loadJSONGoldenContract(t *testing.T) jsonGoldenContract {
	t.Helper()

	return loadJSONContractFile(t, "testdata/json_contract/success_schemas.json", "golden schema fixture")
}

func loadJSONContractFile(t *testing.T, file string, label string) jsonGoldenContract {
	t.Helper()

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read %s %s: %v", label, file, err)
	}

	var raw struct {
		Definitions map[string]json.RawMessage            `json:"definitions"`
		Commands    map[string]map[string]json.RawMessage `json:"commands"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("decode raw %s %s: %v", label, file, err)
	}
	if len(raw.Definitions) == 0 {
		t.Fatalf("%s %s has no definitions", label, file)
	}
	if len(raw.Commands) == 0 {
		t.Fatalf("%s %s has no commands", label, file)
	}

	requiredCommandFields := []string{
		"top_level",
		"result_wrapper",
		"input",
		"result_input",
		"result",
		"statuses",
		"kinds",
		"warnings",
	}
	for command, fields := range raw.Commands {
		for _, field := range requiredCommandFields {
			if _, ok := fields[field]; !ok {
				t.Errorf("%s for %q missing %q", label, command, field)
			}
		}
	}

	var contract jsonGoldenContract
	if err := json.Unmarshal(data, &contract); err != nil {
		t.Fatalf("decode %s %s: %v", label, file, err)
	}
	return normalizeGoldenContract(contract)
}

func loadJSONGoldenSuccessOutputs(t *testing.T) map[string]json.RawMessage {
	t.Helper()

	data, err := os.ReadFile("testdata/json_contract/success_outputs.json")
	if err != nil {
		t.Fatalf("read golden success output fixture: %v", err)
	}

	var fixtures map[string]json.RawMessage
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatalf("decode golden success output fixture: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatalf("golden success output fixture has no commands")
	}
	return fixtures
}

func loadPublicJSONSchema(t *testing.T, file string) publicJSONSchema {
	t.Helper()

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read public JSON schema %s: %v", file, err)
	}

	var schema publicJSONSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("decode public JSON schema %s: %v", file, err)
	}
	return schema
}

func expectedJSONErrorCodes() []string {
	return []string{
		jsonErrorCodeAppKeyRequired,
		jsonErrorCodeAuthExchangeFailed,
		jsonErrorCodeAuthRefreshFailed,
		jsonErrorCodeAuthRequired,
		jsonErrorCodeCommandFailed,
		jsonErrorCodeDropboxAPIError,
		jsonErrorCodeEnvTokenStillActive,
		jsonErrorCodeInvalidArguments,
		jsonErrorCodeNotFound,
		jsonErrorCodePathConflict,
		jsonErrorCodePermissionDenied,
		jsonErrorCodeRateLimited,
		jsonErrorCodeStructuredOutputUnsupported,
		jsonErrorCodeUnknownCommand,
		jsonErrorCodeUnknownFlag,
	}
}

func assertGoldenJSONEqual(t *testing.T, command string, fixture json.RawMessage, actual any) {
	t.Helper()

	actualJSON, err := json.Marshal(actual)
	if err != nil {
		t.Fatalf("marshal JSON example for %q: %v", command, err)
	}

	var got any
	if err := json.Unmarshal(fixture, &got); err != nil {
		t.Fatalf("decode golden output for %q: %v", command, err)
	}
	var want any
	if err := json.Unmarshal(actualJSON, &want); err != nil {
		t.Fatalf("decode generated output for %q: %v", command, err)
	}
	if reflect.DeepEqual(got, want) {
		return
	}

	gotJSON, _ := json.MarshalIndent(got, "", "  ")
	wantJSON, _ := json.MarshalIndent(want, "", "  ")
	t.Errorf("golden output for %q = %s, want %s", command, gotJSON, wantJSON)
}

func assertGoldenSuccessOutputStatuses(t *testing.T, command string, fixture json.RawMessage) {
	t.Helper()

	var got jsonOperationOutput
	if err := json.Unmarshal(fixture, &got); err != nil {
		t.Fatalf("decode golden output for %q: %v", command, err)
	}
	for i, result := range got.Results {
		if result.Status == "" {
			t.Errorf("golden output for %q result %d has empty status", command, i)
		}
		if result.Status == "unknown" {
			t.Errorf("golden output for %q result %d must not use unknown status", command, i)
		}
		if result.Kind == "" {
			t.Errorf("golden output for %q result %d has empty kind", command, i)
		}
		if result.Kind == "unknown" {
			t.Errorf("golden output for %q result %d must not use unknown kind", command, i)
		}
	}
}

func jsonGoldenSuccessOutputExamples() map[string]jsonOperationOutput {
	file := sampleJSONFileMetadata("/Reports/old.pdf")
	copyFile := sampleJSONFileMetadata("/Reports/copy.pdf")
	folder := sampleJSONFolderMetadata("/Reports")
	sharedLink := sampleShareLinkJSONMetadata()
	teamMember := sampleTeamMemberJSON()

	examples := map[string]jsonOperationOutput{
		"account": newJSONOperationOutput(accountInput{AccountID: "dbid:lookup"}, []jsonOperationResult{
			newJSONOperationResult(accountJSONStatusFound, accountKindAccount, accountInput{AccountID: "dbid:lookup"}, sampleJSONAccount()),
		}, nil),
		"cp": newJSONOperationOutput(nil, []jsonOperationResult{
			newJSONOperationResult(relocationJSONStatusCopied, copyFile.Type, relocationInput{FromPath: "/Reports/old.pdf", ToPath: "/Reports/copy.pdf"}, copyFile),
		}, nil),
		"du": newJSONOperationOutput(duInput{}, []jsonOperationResult{
			newJSONOperationResult(duJSONStatusReported, duKindSpaceUsage, duInput{}, duOutput{
				Used: 2048,
				Allocation: duAllocation{
					Type:                          "team",
					Allocated:                     uint64Ptr(1000000),
					Used:                          uint64Ptr(2048),
					UserWithinTeamSpaceAllocated:  uint64Ptr(500000),
					UserWithinTeamSpaceUsedCached: uint64Ptr(1024),
					UserWithinTeamSpaceLimitType:  "fixed",
				},
			}),
		}, nil),
		"get": newJSONOperationOutput(getCommandInput{Source: "/Reports/old.pdf", Target: "old.pdf", Recursive: false, Stdout: false}, []jsonOperationResult{
			newJSONOperationResult(getStatusDownloaded, getKindFile, getResultInput{Source: "/Reports/old.pdf", Target: "old.pdf"}, file),
		}, nil),
		"help": newJSONOperationOutput(jsonHelpInput{Help: true, Path: "ls"}, []jsonOperationResult{
			newJSONOperationResult(jsonHelpStatusDescribed, jsonHelpKindCommand, nil, jsonCommandManifest{
				Path:     "ls",
				Use:      "dbxcli ls [flags] [<path>]",
				Short:    "List files and folders",
				Aliases:  []string{},
				Runnable: true,
				Flags: []jsonCommandFlag{{
					Name:      "output",
					Type:      "string",
					Default:   "text",
					Usage:     "Output format: text, json",
					Inherited: true,
				}},
				SupportsStructuredOutput: true,
				AuthModes:                []string{"personal", "team-access"},
				DestructiveLevel:         destructiveLevelNone,
			}),
		}, nil),
		"ls": newJSONOperationOutput(lsInput{Path: "/Reports", Recursive: false, IncludeDeleted: true, OnlyDeleted: false, Long: true, Sort: "name", Reverse: false, Time: "server", TimeFormat: "2006-01-02"}, []jsonOperationResult{
			newJSONOperationResult(lsJSONStatusListed, file.Type, nil, file),
		}, nil),
		"logout": newJSONOperationOutput(nil, []jsonOperationResult{
			newJSONOperationResult(logoutStatusLoggedOut, logoutKindAuth, nil, logoutResult{RemovedSavedCredentials: true, RemoteTokenRevoked: true}),
		}, nil),
		"mkdir": newJSONOperationOutput(mkdirInput{Path: "/Reports/new", Parents: true}, []jsonOperationResult{
			newJSONOperationResult(mkdirStatusCreated, mkdirKindFolder, mkdirInput{Path: "/Reports/new", Parents: true}, sampleJSONFolderMetadata("/Reports/new")),
		}, nil),
		"mv": newJSONOperationOutput(nil, []jsonOperationResult{
			newJSONOperationResult(relocationJSONStatusMoved, "file", relocationInput{FromPath: "/Reports/copy.pdf", ToPath: "/Reports/moved.pdf"}, sampleJSONFileMetadata("/Reports/moved.pdf")),
		}, nil),
		"put": newJSONOperationOutput(putCommandInput{Source: "README.md", Target: "/README.md", Recursive: true, IfExists: putIfExistsOverwrite, Stdin: false}, []jsonOperationResult{
			newJSONOperationResult(putStatusUploaded, putKindFile, putResultInput{Source: "README.md", Target: "/README.md"}, sampleJSONFileMetadata("/README.md")),
		}, []jsonWarning{{Code: jsonWarningCodeSkippedSymlink, Message: "skipped symlink", Path: "docs/link"}}),
		"restore": newJSONOperationOutput(restoreInput{Path: "/Reports/old.pdf", Revision: "015f"}, []jsonOperationResult{
			newJSONOperationResult(restoreStatusRestored, restoreKindFile, restoreInput{Path: "/Reports/old.pdf", Revision: "015f"}, file),
		}, nil),
		"revs": newJSONOperationOutput(revsInput{Path: "/Reports/old.pdf", Long: true, Time: "server", TimeFormat: "2006-01-02"}, []jsonOperationResult{
			newJSONOperationResult(revsJSONStatusRevision, file.Type, nil, file),
		}, nil),
		"rm": newJSONOperationOutput(nil, []jsonOperationResult{
			newJSONOperationResult(removeJSONStatusDeleted, file.Type, removeInput{Path: "/Reports/old.pdf", Permanent: false, Recursive: false, Force: false}, file),
		}, nil),
		"search": newJSONOperationOutput(searchInput{Query: "report", Path: "/Reports", Long: true, Sort: "name", Reverse: false, Time: "server", TimeFormat: "2006-01-02"}, []jsonOperationResult{
			newJSONOperationResult(searchJSONStatusFound, folder.Type, nil, folder),
		}, nil),
		"share list folder": newJSONOperationOutput(shareFolderListInput{}, []jsonOperationResult{
			newJSONOperationResult(shareFolderJSONStatusListed, shareFolderJSONKindFolder, nil, sampleShareFolderJSONMetadata()),
		}, nil),
		"share list link": newJSONOperationOutput(shareLinkListInput{Path: "/Reports/old.pdf", DirectOnly: true}, []jsonOperationResult{
			shareLinkJSONOperationResult(shareLinkJSONStatusListed, sharedLink),
		}, []jsonWarning{{Code: jsonWarningCodeDeprecatedCommand, Message: "use `dbxcli share-link list` instead"}}),
		"share-link create": newJSONOperationOutput(shareLinkCreateInput{Path: "/Reports/old.pdf", Access: "viewer", Audience: "public", Expires: "2026-07-01T00:00:00Z", RemoveExpiration: false, AllowDownload: true, DisallowDownload: false, Password: true}, []jsonOperationResult{
			shareLinkJSONOperationResult(shareLinkJSONStatusCreated, sharedLink),
		}, nil),
		"share-link download": newJSONOperationOutput(shareLinkDownloadInput{URL: sharedLink.URL, Target: "old.pdf", Path: "/old.pdf", Recursive: false, Password: true}, []jsonOperationResult{
			newJSONOperationResult(shareLinkJSONStatusDownloaded, sharedLink.Type, nil, shareLinkDownloadResult{Target: "old.pdf", Link: sharedLink}),
		}, nil),
		"share-link info": newJSONOperationOutput(shareLinkInfoInput{URL: sharedLink.URL, Path: "/old.pdf", Password: true}, []jsonOperationResult{
			shareLinkJSONOperationResult(shareLinkJSONStatusFound, sharedLink),
		}, nil),
		"share-link list": newJSONOperationOutput(shareLinkListInput{Path: "/Reports/old.pdf", DirectOnly: true}, []jsonOperationResult{
			shareLinkJSONOperationResult(shareLinkJSONStatusListed, sharedLink),
		}, nil),
		"share-link revoke": newJSONOperationOutput(shareLinkRevokeInput{Path: "/Reports/old.pdf"}, []jsonOperationResult{
			newJSONOperationResult(shareLinkJSONStatusRevoked, sharedLink.Type, nil, shareLinkRevokeResult{URL: sharedLink.URL, Link: &sharedLink}),
		}, nil),
		"share-link update": newJSONOperationOutput(shareLinkUpdateInput{URL: sharedLink.URL, Audience: "public", Expires: "2026-07-01T00:00:00Z", RemoveExpiration: false, AllowDownload: true, DisallowDownload: false, Password: true, RemovePassword: false}, []jsonOperationResult{
			shareLinkJSONOperationResult(shareLinkJSONStatusUpdated, sharedLink),
		}, nil),
		"team add-member": newJSONOperationOutput(teamMemberAddInput{Email: "ada@example.com", FirstName: "Ada", LastName: "Lovelace"}, []jsonOperationResult{
			newJSONOperationResult(teamJSONStatusAdded, teamJSONKindTeamMember, teamMemberAddInput{Email: "ada@example.com", FirstName: "Ada", LastName: "Lovelace"}, teamMemberMutationJSON{
				Type: teamJSONTypeMemberAdd,
				Tag:  "complete",
				Results: []teamMemberAddItemJSON{{
					Tag:    "success",
					Email:  "ada@example.com",
					Member: &teamMember,
				}},
			}),
		}, nil),
		"team info": newJSONOperationOutput(teamInfoInput{}, []jsonOperationResult{
			newJSONOperationResult(teamJSONStatusFound, teamJSONKindTeam, teamInfoInput{}, teamInfoJSON{Type: teamJSONKindTeam, Name: "Engineering", TeamID: "team-id", NumLicensedUsers: 10, NumProvisionedUsers: 8}),
		}, nil),
		"team list-groups": newJSONOperationOutput(teamInfoInput{}, []jsonOperationResult{
			newJSONOperationResult(teamJSONStatusListed, teamJSONKindTeamGroup, nil, teamGroupJSON{Type: teamJSONKindTeamGroup, GroupName: "Developers", GroupID: "g:dev", GroupExternalID: "external-dev", MemberCount: 3, GroupManagementType: "company_managed"}),
		}, nil),
		"team list-members": newJSONOperationOutput(teamInfoInput{}, []jsonOperationResult{
			newJSONOperationResult(teamJSONStatusListed, teamJSONKindTeamMember, nil, teamMember),
		}, nil),
		"team remove-member": newJSONOperationOutput(teamMemberRemoveInput{Email: "ada@example.com"}, []jsonOperationResult{
			newJSONOperationResult(teamJSONStatusRemoved, teamJSONKindTeamMember, teamMemberRemoveInput{Email: "ada@example.com"}, teamMemberMutationJSON{Type: teamJSONTypeMemberRemove, Tag: "complete", AsyncJobID: "async-job-id"}),
		}, nil),
		"version": newJSONOperationOutput(versionInput{}, []jsonOperationResult{
			newJSONOperationResult(versionJSONStatusReported, versionKindVersion, versionInput{}, versionOutput{Version: "1.2.3", SDKVersion: "sdk-version", SpecVersion: "spec-version"}),
		}, nil),
	}
	for command, example := range examples {
		example.OK = true
		example.SchemaVersion = jsonSchemaVersion
		example.Command = command
		examples[command] = example
	}
	return examples
}

func sampleJSONAccount() jsonAccount {
	return jsonAccount{
		Type:            "full",
		AccountID:       "dbid:account",
		Name:            &jsonAccountName{GivenName: "Ada", Surname: "Lovelace", FamiliarName: "Ada", DisplayName: "Ada Lovelace", AbbreviatedName: "AL"},
		Email:           "ada@example.com",
		EmailVerified:   true,
		Disabled:        false,
		ProfilePhotoURL: "https://example.com/profile.jpg",
		Locale:          "en",
		ReferralLink:    "https://example.com/referral",
		IsPaired:        boolPtr(false),
		AccountType:     "basic",
		IsTeammate:      boolPtr(true),
		TeamMemberID:    "dbmid:team-member",
		Team:            &jsonAccountTeam{ID: "team-id", Name: "Engineering", MemberID: "dbmid:team-member"},
	}
}

func sampleJSONFileMetadata(path string) jsonMetadata {
	return jsonMetadata{
		Type:           "file",
		PathDisplay:    path,
		PathLower:      strings.ToLower(path),
		ID:             "id:file",
		Rev:            "015f",
		Size:           uint64Ptr(123),
		ServerModified: jsonContractStringPtr("2026-06-25T12:00:00Z"),
		ClientModified: jsonContractStringPtr("2026-06-25T11:00:00Z"),
	}
}

func sampleJSONFolderMetadata(path string) jsonMetadata {
	return jsonMetadata{
		Type:        "folder",
		PathDisplay: path,
		PathLower:   strings.ToLower(path),
		ID:          "id:folder",
	}
}

func sampleShareFolderJSONMetadata() shareFolderJSONMetadata {
	return shareFolderJSONMetadata{
		Type:                 shareFolderJSONKindFolder,
		Name:                 "Reports",
		PathLower:            "/reports",
		SharedFolderID:       "sfid:reports",
		PreviewURL:           "https://www.dropbox.com/preview",
		AccessType:           "owner",
		IsInsideTeamFolder:   false,
		IsTeamFolder:         true,
		OwnerDisplayNames:    []string{"Ada Lovelace"},
		ParentSharedFolderID: "sfid:parent",
		ParentFolderName:     "Parent",
		TimeInvited:          jsonContractStringPtr("2026-06-25T10:00:00Z"),
		AccessInheritance:    "inherit",
	}
}

func sampleShareLinkJSONMetadata() shareLinkJSONMetadata {
	return shareLinkJSONMetadata{
		Type:           "file",
		URL:            "https://www.dropbox.com/s/example/old.pdf",
		Name:           "old.pdf",
		PathLower:      "/reports/old.pdf",
		ID:             "id:shared-file",
		Expires:        jsonContractStringPtr("2026-07-01T00:00:00Z"),
		Rev:            "015f",
		Size:           uint64Ptr(123),
		ServerModified: jsonContractStringPtr("2026-06-25T12:00:00Z"),
		ClientModified: jsonContractStringPtr("2026-06-25T11:00:00Z"),
		Permissions: &shareLinkJSONPermissions{
			ResolvedVisibility:            "public",
			RequestedVisibility:           "public",
			EffectiveAudience:             "public",
			AccessLevel:                   "viewer",
			CanRevoke:                     true,
			AllowDownload:                 true,
			CanSetExpiry:                  true,
			CanRemoveExpiry:               true,
			CanAllowDownload:              true,
			CanDisallowDownload:           true,
			AllowComments:                 false,
			CanSetPassword:                true,
			CanRemovePassword:             true,
			RequirePassword:               true,
			CanUseExtendedSharingControls: true,
		},
	}
}

func sampleTeamMemberJSON() teamMemberJSON {
	return teamMemberJSON{
		Type:                  teamJSONKindTeamMember,
		TeamMemberID:          "dbmid:team-member",
		ExternalID:            "external-member",
		AccountID:             "dbid:account",
		Email:                 "ada@example.com",
		EmailVerified:         true,
		Status:                "active",
		Name:                  &jsonAccountName{GivenName: "Ada", Surname: "Lovelace", FamiliarName: "Ada", DisplayName: "Ada Lovelace", AbbreviatedName: "AL"},
		Role:                  "member_only",
		Groups:                []string{"g:dev"},
		MemberFolderID:        "ns:member-folder",
		MembershipType:        "full",
		InvitedOn:             jsonContractStringPtr("2026-06-24T12:00:00Z"),
		JoinedOn:              jsonContractStringPtr("2026-06-25T12:00:00Z"),
		SuspendedOn:           jsonContractStringPtr("2026-06-26T12:00:00Z"),
		PersistentID:          "persistent-id",
		IsDirectoryRestricted: true,
		ProfilePhotoURL:       "https://example.com/member.jpg",
	}
}

func jsonContractStringPtr(value string) *string {
	return &value
}

func jsonContractDefinitions() map[string][]string {
	return normalizeStringSliceMap(map[string][]string{
		"account":                    jsonFieldNames[jsonAccount](),
		"account_input":              jsonFieldNames[accountInput](),
		"account_name":               jsonFieldNames[jsonAccountName](),
		"account_team":               jsonFieldNames[jsonAccountTeam](),
		"du_allocation":              jsonFieldNames[duAllocation](),
		"du_output":                  jsonFieldNames[duOutput](),
		"empty":                      {},
		"command_flag":               jsonFieldNames[jsonCommandFlag](),
		"command_manifest":           jsonFieldNames[jsonCommandManifest](),
		"get_input":                  jsonFieldNames[getCommandInput](),
		"get_result_input":           jsonFieldNames[getResultInput](),
		"help_input":                 jsonFieldNames[jsonHelpInput](),
		"ls_input":                   jsonFieldNames[lsInput](),
		"logout_result":              jsonFieldNames[logoutResult](),
		"metadata":                   jsonFieldNames[jsonMetadata](),
		"mkdir_input":                jsonFieldNames[mkdirInput](),
		"operation_output":           jsonFieldNames[jsonOperationOutput](),
		"operation_result":           jsonFieldNames[jsonOperationResult](),
		"put_input":                  jsonFieldNames[putCommandInput](),
		"put_result_input":           jsonFieldNames[putResultInput](),
		"relocation_input":           jsonFieldNames[relocationInput](),
		"remove_input":               jsonFieldNames[removeInput](),
		"restore_input":              jsonFieldNames[restoreInput](),
		"revs_input":                 jsonFieldNames[revsInput](),
		"search_input":               jsonFieldNames[searchInput](),
		"share_folder":               jsonFieldNames[shareFolderJSONMetadata](),
		"share_link_create_input":    jsonFieldNames[shareLinkCreateInput](),
		"share_link_download_input":  jsonFieldNames[shareLinkDownloadInput](),
		"share_link_download_result": jsonFieldNames[shareLinkDownloadResult](),
		"share_link_info_input":      jsonFieldNames[shareLinkInfoInput](),
		"share_link_list_input":      jsonFieldNames[shareLinkListInput](),
		"share_link_metadata":        jsonFieldNames[shareLinkJSONMetadata](),
		"share_link_permissions":     jsonFieldNames[shareLinkJSONPermissions](),
		"share_link_revoke_input":    jsonFieldNames[shareLinkRevokeInput](),
		"share_link_revoke_result":   jsonFieldNames[shareLinkRevokeResult](),
		"share_link_update_input":    jsonFieldNames[shareLinkUpdateInput](),
		"team_group":                 jsonFieldNames[teamGroupJSON](),
		"team_info":                  jsonFieldNames[teamInfoJSON](),
		"team_member":                jsonFieldNames[teamMemberJSON](),
		"team_member_add_input":      jsonFieldNames[teamMemberAddInput](),
		"team_member_add_item":       jsonFieldNames[teamMemberAddItemJSON](),
		"team_member_mutation":       jsonFieldNames[teamMemberMutationJSON](),
		"team_member_remove_input":   jsonFieldNames[teamMemberRemoveInput](),
		"version":                    jsonFieldNames[versionOutput](),
	})
}

func jsonCommandSchemas() map[string]jsonGoldenCommandSchema {
	return map[string]jsonGoldenCommandSchema{
		"account":           operationSchema("account_input", schemaRef("account_input"), "account", []string{accountJSONStatusFound}, []string{accountKindAccount}, nil),
		"cp":                operationSchema("empty", schemaRef("relocation_input"), "metadata", []string{relocationJSONStatusCopied}, metadataKinds(), nil),
		"du":                operationSchema("empty", schemaRef("empty"), "du_output", []string{duJSONStatusReported}, []string{duKindSpaceUsage}, nil),
		"get":               operationSchema("get_input", schemaRef("get_result_input"), "metadata", []string{getStatusCreated, getStatusDownloaded, getStatusExisting}, []string{getKindFile, getKindFolder}, nil),
		"help":              operationSchema("help_input", schemaRef("empty"), "command_manifest", []string{jsonHelpStatusDescribed}, []string{jsonHelpKindCommand}, nil),
		"ls":                operationSchema("ls_input", schemaRef("empty"), "metadata", []string{lsJSONStatusListed}, metadataKinds(), nil),
		"logout":            operationSchema("empty", schemaRef("empty"), "logout_result", []string{logoutStatusAlreadyLoggedOut, logoutStatusLoggedOut}, []string{logoutKindAuth}, []string{jsonWarningCodeTokenRevokeFailed}),
		"mkdir":             operationSchema("mkdir_input", schemaRef("mkdir_input"), "metadata", []string{mkdirStatusCreated, mkdirStatusExisting}, []string{mkdirKindFolder}, nil),
		"mv":                operationSchema("empty", schemaRef("relocation_input"), "metadata", []string{relocationJSONStatusMoved}, metadataKinds(), nil),
		"put":               operationSchema("put_input", schemaRef("put_result_input"), "metadata", []string{putStatusCreated, putStatusExisting, putStatusSkipped, putStatusUploaded}, []string{putKindFile, putKindFolder}, []string{jsonWarningCodeSkippedSymlink}),
		"restore":           operationSchema("restore_input", schemaRef("restore_input"), "metadata", []string{restoreStatusRestored}, []string{restoreKindFile}, nil),
		"revs":              operationSchema("revs_input", schemaRef("empty"), "metadata", []string{revsJSONStatusRevision}, []string{"file"}, nil),
		"rm":                operationSchema("empty", schemaRef("remove_input"), "metadata", []string{removeJSONStatusDeleted, removeJSONStatusPermanentlyDeleted}, metadataKinds(), nil),
		"search":            operationSchema("search_input", schemaRef("empty"), "metadata", []string{searchJSONStatusFound}, metadataKinds(), nil),
		"share list folder": operationSchema("empty", schemaRef("empty"), "share_folder", []string{shareFolderJSONStatusListed}, []string{shareFolderJSONKindFolder}, nil),
		"share list link":   operationSchema("share_link_list_input", schemaRef("empty"), "share_link_metadata", []string{shareLinkJSONStatusListed}, shareLinkKinds(), []string{jsonWarningCodeDeprecatedCommand}),
		"share-link create": operationSchema("share_link_create_input", schemaRef("empty"), "share_link_metadata", []string{shareLinkJSONStatusCreated, shareLinkJSONStatusExisting}, shareLinkKinds(), nil),
		"share-link download": operationSchema(
			"share_link_download_input",
			schemaRef("empty"),
			"share_link_download_result",
			[]string{shareLinkJSONStatusDownloaded},
			shareLinkKinds(),
			nil,
		),
		"share-link info":    operationSchema("share_link_info_input", schemaRef("empty"), "share_link_metadata", []string{shareLinkJSONStatusFound}, shareLinkKinds(), nil),
		"share-link list":    operationSchema("share_link_list_input", schemaRef("empty"), "share_link_metadata", []string{shareLinkJSONStatusListed}, shareLinkKinds(), nil),
		"share-link revoke":  operationSchema("share_link_revoke_input", schemaRef("empty"), "share_link_revoke_result", []string{shareLinkJSONStatusRevoked}, append(shareLinkKinds(), shareLinkJSONKindSharedLink), nil),
		"share-link update":  operationSchema("share_link_update_input", schemaRef("empty"), "share_link_metadata", []string{shareLinkJSONStatusUpdated}, shareLinkKinds(), nil),
		"team add-member":    operationSchema("team_member_add_input", schemaRef("team_member_add_input"), "team_member_mutation", []string{teamJSONStatusAdded, teamJSONStatusCompleted, teamJSONStatusStarted}, []string{teamJSONKindTeamMember}, nil),
		"team info":          operationSchema("empty", schemaRef("empty"), "team_info", []string{teamJSONStatusFound}, []string{teamJSONKindTeam}, nil),
		"team list-groups":   operationSchema("empty", schemaRef("empty"), "team_group", []string{teamJSONStatusListed}, []string{teamJSONKindTeamGroup}, nil),
		"team list-members":  operationSchema("empty", schemaRef("empty"), "team_member", []string{teamJSONStatusListed}, []string{teamJSONKindTeamMember}, nil),
		"team remove-member": operationSchema("team_member_remove_input", schemaRef("team_member_remove_input"), "team_member_mutation", []string{teamJSONStatusCompleted, teamJSONStatusRemoved, teamJSONStatusStarted}, []string{teamJSONKindTeamMember}, nil),
		"version":            operationSchema("empty", schemaRef("empty"), "version", []string{versionJSONStatusReported}, []string{versionKindVersion}, nil),
	}
}

func operationSchema(input string, resultInput *string, result string, statuses, kinds, warnings []string) jsonGoldenCommandSchema {
	return normalizeGoldenCommandSchema(jsonGoldenCommandSchema{
		TopLevel:      "operation_output",
		ResultWrapper: "operation_result",
		Input:         input,
		ResultInput:   resultInput,
		Result:        result,
		Statuses:      statuses,
		Kinds:         kinds,
		Warnings:      warnings,
	})
}

func schemaRef(name string) *string {
	return &name
}

func metadataKinds() []string {
	return []string{"deleted", "file", "folder"}
}

func shareLinkKinds() []string {
	return []string{"file", "folder", "link"}
}

func jsonFieldNames[T any]() []string {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil
	}

	var names []string
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "-" {
			continue
		}
		if name == "" {
			name = field.Name
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func assertStringSliceMapEqual(t *testing.T, label string, got, want map[string][]string) {
	t.Helper()

	got = normalizeStringSliceMap(got)
	want = normalizeStringSliceMap(want)
	if reflect.DeepEqual(got, want) {
		return
	}

	gotJSON, _ := json.MarshalIndent(got, "", "  ")
	wantJSON, _ := json.MarshalIndent(want, "", "  ")
	t.Fatalf("%s = %s, want %s", label, gotJSON, wantJSON)
}

func assertStringSliceEqual(t *testing.T, label string, got, want []string) {
	t.Helper()

	got = sortedCopy(got)
	want = sortedCopy(want)
	if reflect.DeepEqual(got, want) {
		return
	}
	t.Fatalf("%s = %v, want %v", label, got, want)
}

func assertGoldenCommandSchemaEqual(t *testing.T, command string, got, want jsonGoldenCommandSchema) {
	t.Helper()

	got = normalizeGoldenCommandSchema(got)
	want = normalizeGoldenCommandSchema(want)
	if reflect.DeepEqual(got, want) {
		return
	}

	gotJSON, _ := json.MarshalIndent(got, "", "  ")
	wantJSON, _ := json.MarshalIndent(want, "", "  ")
	t.Errorf("golden schema for %q = %s, want %s", command, gotJSON, wantJSON)
}

func assertGoldenCommandSchemaReferences(t *testing.T, command string, schema jsonGoldenCommandSchema, definitions map[string][]string) {
	t.Helper()

	refs := []string{schema.TopLevel, schema.ResultWrapper, schema.Input, schema.Result}
	if schema.ResultInput != nil {
		refs = append(refs, *schema.ResultInput)
	}
	for _, ref := range refs {
		if _, ok := definitions[ref]; !ok {
			t.Errorf("golden schema for %q references unknown definition %q", command, ref)
		}
	}
}

func assertGoldenCommandStatuses(t *testing.T, command string, schema jsonGoldenCommandSchema) {
	t.Helper()

	if len(schema.Statuses) == 0 {
		t.Errorf("golden schema for %q has no result statuses", command)
	}
	for _, status := range schema.Statuses {
		if status == "unknown" {
			t.Errorf("golden schema for %q must not allow unknown result status", command)
		}
	}
	for _, kind := range schema.Kinds {
		if kind == "unknown" {
			t.Errorf("golden schema for %q must not allow unknown result kind", command)
		}
	}
}

func normalizeGoldenContract(contract jsonGoldenContract) jsonGoldenContract {
	contract.Definitions = normalizeStringSliceMap(contract.Definitions)
	commands := make(map[string]jsonGoldenCommandSchema, len(contract.Commands))
	for command, schema := range contract.Commands {
		commands[command] = normalizeGoldenCommandSchema(schema)
	}
	contract.Commands = commands
	return contract
}

func normalizeGoldenCommandSchema(schema jsonGoldenCommandSchema) jsonGoldenCommandSchema {
	schema.Statuses = sortedCopy(schema.Statuses)
	schema.Kinds = sortedCopy(schema.Kinds)
	schema.Warnings = sortedCopy(schema.Warnings)
	return schema
}

func normalizeStringSliceMap(values map[string][]string) map[string][]string {
	normalized := make(map[string][]string, len(values))
	for key, value := range values {
		normalized[key] = sortedCopy(value)
	}
	return normalized
}

func sortedCopy(values []string) []string {
	if values == nil {
		return []string{}
	}
	copied := append([]string{}, values...)
	sort.Strings(copied)
	return copied
}

func mapKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func structuredOutputCommandPaths(root *cobra.Command) []string {
	var paths []string
	var walk func(*cobra.Command, []string)
	walk = func(cmd *cobra.Command, parents []string) {
		parts := parents
		if cmd != root {
			parts = append(append([]string{}, parents...), cmd.Name())
			if commandSupportsStructuredOutput(cmd) {
				paths = append(paths, strings.Join(parts, " "))
			}
		}
		for _, child := range cmd.Commands() {
			walk(child, parts)
		}
	}
	walk(root, nil)
	return paths
}

func TestJSONOperationOutputContractShape(t *testing.T) {
	encoded, err := json.Marshal(newJSONOperationOutput(nil, nil, nil))
	if err != nil {
		t.Fatalf("marshal operation output: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatalf("decode operation output: %v", err)
	}
	for _, key := range []string{"ok", "schema_version", "command", "input", "results", "warnings"} {
		if _, ok := raw[key]; !ok {
			t.Fatalf("operation output = %s, missing %q", encoded, key)
		}
	}
	if len(raw) != 6 {
		t.Fatalf("operation output = %s, want only ok/schema_version/command/input/results/warnings", encoded)
	}
}

func TestUnsupportedCommandsReturnJSONErrorEnvelope(t *testing.T) {
	for _, name := range []string{"login", "completion"} {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := &cobra.Command{Use: name}
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.Flags().String(outputFlag, "json", "")

			err := validateOutputFormat(cmd)
			if !errors.Is(err, output.ErrStructuredOutputUnsupported) {
				t.Fatalf("validateOutputFormat error = %v, want structured output unsupported", err)
			}

			renderCommandError(cmd, err)
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}

			got := decodeJSONErrorResponse(t, stdout.String())
			if got.OK {
				t.Fatalf("ok = true, want false")
			}
			if got.Error.Code != jsonErrorCodeStructuredOutputUnsupported {
				t.Fatalf("code = %q, want %q", got.Error.Code, jsonErrorCodeStructuredOutputUnsupported)
			}
			if got.Warnings == nil || len(got.Warnings) != 0 {
				t.Fatalf("warnings = %+v, want empty array", got.Warnings)
			}
		})
	}
}
