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
	"os"
	"reflect"
	"testing"

	schemagen "github.com/dropbox/dbxcli/v3/internal/jsonschema"
	"github.com/santhosh-tekuri/jsonschema/v6"
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
