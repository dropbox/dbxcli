// Copyright © 2026 Dropbox, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	schemagen "github.com/dropbox/dbxcli/v3/internal/jsonschema"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/spf13/cobra"
)

func TestPublicJSONCommandSuccessSchemaMatchesGeneratedCatalog(t *testing.T) {
	catalog := loadCommandSchemaCatalog(t)

	generated, err := schemagen.GenerateCommandSuccessSchema(catalog)
	if err != nil {
		t.Fatalf("generate command success schema: %v", err)
	}

	got := loadJSONValueFile(t, "../docs/json-schema/v1/commands.schema.json")
	want := normalizeJSONValue(t, generated)
	if reflect.DeepEqual(got, want) {
		return
	}

	gotJSON, _ := json.MarshalIndent(got, "", "  ")
	wantJSON, _ := json.MarshalIndent(want, "", "  ")
	t.Fatalf("public command success schema = %s, want %s", gotJSON, wantJSON)
}

func TestGoldenSuccessOutputsValidateAgainstPublicCommandSuccessSchema(t *testing.T) {
	schema := compileJSONSchemaFile(t, "../docs/json-schema/v1/commands.schema.json")

	for command, raw := range loadJSONGoldenSuccessOutputs(t) {
		t.Run(command, func(t *testing.T) {
			value := decodeJSONValueForSchema(t, raw)
			if err := schema.Validate(value); err != nil {
				t.Fatalf("golden success output does not validate: %v", err)
			}
		})
	}
}

func TestPublicCommandSuccessSchemaRejectsInvalidStatus(t *testing.T) {
	schema := compileJSONSchemaFile(t, "../docs/json-schema/v1/commands.schema.json")

	value := decodeJSONValue(t, loadJSONGoldenSuccessOutputs(t)["cp"])
	root := value.(map[string]any)
	results := root["results"].([]any)
	first := results[0].(map[string]any)
	first["status"] = "not-a-real-status"
	value = normalizeJSONValueForSchema(t, value)

	if err := schema.Validate(value); err == nil {
		t.Fatal("invalid status validated successfully")
	}
}

func TestPublicCommandSuccessSchemaRejectsInvalidPrimitiveType(t *testing.T) {
	schema := compileJSONSchemaFile(t, "../docs/json-schema/v1/commands.schema.json")

	value := decodeJSONValue(t, loadJSONGoldenSuccessOutputs(t)["put"])
	root := value.(map[string]any)
	input := root["input"].(map[string]any)
	input["recursive"] = "false"
	value = normalizeJSONValueForSchema(t, value)

	if err := schema.Validate(value); err == nil {
		t.Fatal("invalid primitive type validated successfully")
	}
}

func TestJSONHelpManifestsValidateAgainstPublicManifestSchema(t *testing.T) {
	schema := compileJSONSchemaFile(t, "../docs/json-schema/v1/manifest.schema.json")

	RootCmd.InitDefaultHelpCmd()
	for _, command := range publicCommandSubtree(RootCmd) {
		manifest := jsonCommandManifestFor(command)
		value := normalizeJSONValueForSchema(t, manifest)
		if err := schema.Validate(value); err != nil {
			t.Fatalf("manifest %q does not validate: %v", manifest.Path, err)
		}
	}
}

func TestJSONErrorExamplesValidateAgainstPublicErrorSchema(t *testing.T) {
	schema := compileJSONSchemaFile(t, "../docs/json-schema/v1/error.schema.json")
	examples := []struct {
		name string
		err  error
	}{
		{name: "generic", err: errors.New("failed")},
		{name: "invalid arguments", err: invalidArgumentsErrorfWithDetails("invalid --if-exists %q", flagValueErrorDetails("if-exists", "replace"), "replace")},
		{name: "path conflict", err: pathConflictErrorWithPath("/file", "path exists: %s", "/file")},
		{name: "auth required", err: missingAccessTokenError(tokenPersonal)},
		{name: "partial transfer", err: partialStdoutError(12)},
	}

	for _, example := range examples {
		t.Run(example.name, func(t *testing.T) {
			value := normalizeJSONValueForSchema(t, newJSONErrorResponse(RootCmd, example.err))
			if err := schema.Validate(value); err != nil {
				t.Fatalf("JSON error response does not validate: %v", err)
			}
		})
	}
}

func TestJSONErrorSchemaRejectsUnknownDetailsKey(t *testing.T) {
	schema := compileJSONSchemaFile(t, "../docs/json-schema/v1/error.schema.json")

	value := normalizeJSONValueForSchema(t, newJSONErrorResponse(RootCmd, pathConflictErrorWithPath("/file", "path exists: %s", "/file")))
	details := jsonErrorDetailsFromSchemaValue(t, value)
	details["unexpected"] = "value"

	if err := schema.Validate(value); err == nil {
		t.Fatal("JSON error response with unknown details key validated successfully")
	}
}

func TestJSONErrorSchemaRejectsInvalidDetailsType(t *testing.T) {
	schema := compileJSONSchemaFile(t, "../docs/json-schema/v1/error.schema.json")

	value := normalizeJSONValueForSchema(t, newJSONErrorResponse(RootCmd, partialStdoutError(12)))
	details := jsonErrorDetailsFromSchemaValue(t, value)
	details["bytes_written"] = "12"

	if err := schema.Validate(value); err == nil {
		t.Fatal("JSON error response with invalid details type validated successfully")
	}
}

func TestLiveJSONSuccessOutputsValidateAgainstPublicCommandSuccessSchema(t *testing.T) {
	schema := compileJSONSchemaFile(t, "../docs/json-schema/v1/commands.schema.json")

	t.Run("version", func(t *testing.T) {
		var stdout bytes.Buffer
		cmd := NewVersionCommand("test-version")
		cmd.SetOut(&stdout)
		cmd.Flags().String(outputFlag, "json", "")

		if err := versionCommand(cmd, "test-version"); err != nil {
			t.Fatalf("versionCommand returned error: %v", err)
		}
		assertJSONBytesValidateAgainstSchema(t, schema, stdout.Bytes())
	})

	t.Run("ls", func(t *testing.T) {
		cmd, stdout := testLsCmd(t)
		setLsOutputJSON(t, cmd)
		stubFilesClient(t, &mockFilesClient{
			listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
				return &files.ListFolderResult{
					Entries: []files.IsMetadata{
						&files.FileMetadata{
							Metadata: files.Metadata{
								Name:        "file.txt",
								PathDisplay: "/file.txt",
								PathLower:   "/file.txt",
							},
							Id:   "id:file",
							Rev:  "rev-file",
							Size: 42,
						},
					},
				}, nil
			},
		})

		if err := ls(cmd, []string{"/"}); err != nil {
			t.Fatalf("ls returned error: %v", err)
		}
		assertJSONBytesValidateAgainstSchema(t, schema, stdout.Bytes())
	})

	t.Run("put", func(t *testing.T) {
		tmpFile := writeJSONSchemaTempFile(t, "schema-live-put.txt", "data")
		stubFilesClient(t, &mockFilesClient{
			uploadFn: func(arg *files.UploadArg, content io.Reader) (*files.FileMetadata, error) {
				if _, err := io.ReadAll(content); err != nil {
					t.Fatal(err)
				}
				return putFileMetadata(arg.Path, 4), nil
			},
		})

		var stdout bytes.Buffer
		cmd := testPutJSONCmd(&stdout, nil)
		cmd.Use = "put"
		if err := put(cmd, []string{tmpFile, "/schema-live-put.txt"}); err != nil {
			t.Fatalf("put returned error: %v", err)
		}
		assertJSONBytesValidateAgainstSchema(t, schema, stdout.Bytes())
	})

	t.Run("share-link list", func(t *testing.T) {
		stubSharedLinkClient(t, &mockSharedLinkClient{
			listSharedLinksFn: func(arg *sharing.ListSharedLinksArg) (*sharing.ListSharedLinksResult, error) {
				return sharing.NewListSharedLinksResult([]sharing.IsSharedLinkMetadata{
					sharedLinkFile("/docs/report.txt", "https://example.com/report"),
				}, false), nil
			},
		})

		var stdout bytes.Buffer
		cmd := &cobra.Command{Use: "list"}
		parent := &cobra.Command{Use: "share-link"}
		root := &cobra.Command{Use: "dbxcli"}
		root.AddCommand(parent)
		parent.AddCommand(cmd)
		cmd.SetOut(&stdout)
		setShareLinkOutputJSON(t, cmd)

		if err := shareLinkList(cmd, []string{"/docs/report.txt"}); err != nil {
			t.Fatalf("shareLinkList returned error: %v", err)
		}
		assertJSONBytesValidateAgainstSchema(t, schema, stdout.Bytes())
	})
}

func TestLiveJSONErrorOutputsValidateAgainstPublicErrorSchema(t *testing.T) {
	schema := compileJSONSchemaFile(t, "../docs/json-schema/v1/error.schema.json")

	t.Run("invalid arguments", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd := &cobra.Command{Use: "cp"}
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.Flags().String(outputFlag, "json", "")

		err := cp(cmd, []string{"/source"})
		if err == nil {
			t.Fatal("cp returned nil error, want invalid arguments")
		}
		renderCommandError(cmd, err)

		if stderr.Len() != 0 {
			t.Fatalf("stderr = %q, want empty", stderr.String())
		}
		assertJSONBytesValidateAgainstSchema(t, schema, stdout.Bytes())
	})

	t.Run("deprecated command error", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd := &cobra.Command{
			Use:        "link",
			Deprecated: shareListLinksDeprecatedMessage,
		}
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.Flags().String(outputFlag, "json", "")

		err := shareListLinks(cmd, []string{"/one", "/two"})
		if err == nil {
			t.Fatal("shareListLinks returned nil error, want invalid arguments")
		}
		renderCommandError(cmd, err)

		if stderr.Len() != 0 {
			t.Fatalf("stderr = %q, want empty", stderr.String())
		}
		assertJSONBytesValidateAgainstSchema(t, schema, stdout.Bytes())

		var got jsonErrorResponse
		if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
			t.Fatalf("decode JSON error response: %v", err)
		}
		if len(got.Warnings) != 1 || got.Warnings[0].Code != jsonWarningCodeDeprecatedCommand {
			t.Fatalf("warnings = %+v, want deprecated command warning", got.Warnings)
		}
	})
}

func loadCommandSchemaCatalog(t *testing.T) schemagen.CommandCatalog {
	t.Helper()

	data := readJSONFile(t, "../docs/json-schema/v1/commands.json")
	catalog, err := schemagen.DecodeCommandCatalog(data)
	if err != nil {
		t.Fatalf("decode command catalog: %v", err)
	}
	return catalog
}

func loadJSONValueFile(t *testing.T, file string) any {
	t.Helper()
	return decodeJSONValue(t, readJSONFile(t, file))
}

func readJSONFile(t *testing.T, file string) []byte {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	return data
}

func decodeJSONValue(t *testing.T, data []byte) any {
	t.Helper()

	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("decode JSON value: %v", err)
	}
	return value
}

func normalizeJSONValue(t *testing.T, value any) any {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal JSON value: %v", err)
	}
	return decodeJSONValue(t, data)
}

func compileJSONSchemaFile(t *testing.T, file string) *jsonschema.Schema {
	t.Helper()

	data := readJSONFile(t, file)
	doc := decodeJSONValueForSchema(t, data)
	object, ok := doc.(map[string]any)
	if !ok {
		t.Fatalf("%s schema root is %T, want object", file, doc)
	}
	id, ok := object["$id"].(string)
	if !ok || id == "" {
		t.Fatalf("%s schema has no $id", file)
	}

	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	if err := compiler.AddResource(id, doc); err != nil {
		t.Fatalf("add schema resource %s: %v", file, err)
	}
	schema, err := compiler.Compile(id)
	if err != nil {
		t.Fatalf("compile schema %s: %v", file, err)
	}
	return schema
}

func assertJSONBytesValidateAgainstSchema(t *testing.T, schema *jsonschema.Schema, data []byte) {
	t.Helper()

	value := decodeJSONValueForSchema(t, data)
	if err := schema.Validate(value); err != nil {
		t.Fatalf("JSON output does not validate against public schema: %v\noutput: %s", err, string(data))
	}
}

func writeJSONSchemaTempFile(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

func jsonErrorDetailsFromSchemaValue(t *testing.T, value any) map[string]any {
	t.Helper()

	root, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("value is %T, want object", value)
	}
	errorObject, ok := root["error"].(map[string]any)
	if !ok {
		t.Fatalf("error is %T, want object", root["error"])
	}
	details, ok := errorObject["details"].(map[string]any)
	if !ok {
		t.Fatalf("details is %T, want object", errorObject["details"])
	}
	return details
}

func decodeJSONValueForSchema(t *testing.T, data []byte) any {
	t.Helper()

	value, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode JSON value for schema validation: %v", err)
	}
	return value
}

func normalizeJSONValueForSchema(t *testing.T, value any) any {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal JSON value for schema validation: %v", err)
	}
	return decodeJSONValueForSchema(t, data)
}
