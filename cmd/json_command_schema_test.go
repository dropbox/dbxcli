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
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"

	schemagen "github.com/dropbox/dbxcli/v3/internal/jsonschema"
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
	schema := loadJSONValueFile(t, "../docs/json-schema/v1/commands.schema.json")
	validator := newSubsetJSONSchemaValidator(t, schema)

	for command, raw := range loadJSONGoldenSuccessOutputs(t) {
		t.Run(command, func(t *testing.T) {
			value := decodeJSONValue(t, raw)
			if err := validator.validate(schema, value, "$"); err != nil {
				t.Fatalf("golden success output does not validate: %v", err)
			}
		})
	}
}

func TestPublicCommandSuccessSchemaRejectsInvalidStatus(t *testing.T) {
	schema := loadJSONValueFile(t, "../docs/json-schema/v1/commands.schema.json")
	validator := newSubsetJSONSchemaValidator(t, schema)

	value := decodeJSONValue(t, loadJSONGoldenSuccessOutputs(t)["cp"])
	root := value.(map[string]any)
	results := root["results"].([]any)
	first := results[0].(map[string]any)
	first["status"] = "not-a-real-status"

	if err := validator.validate(schema, value, "$"); err == nil {
		t.Fatal("invalid status validated successfully")
	}
}

func TestPublicCommandSuccessSchemaRejectsInvalidPrimitiveType(t *testing.T) {
	schema := loadJSONValueFile(t, "../docs/json-schema/v1/commands.schema.json")
	validator := newSubsetJSONSchemaValidator(t, schema)

	value := decodeJSONValue(t, loadJSONGoldenSuccessOutputs(t)["put"])
	root := value.(map[string]any)
	input := root["input"].(map[string]any)
	input["recursive"] = "false"

	if err := validator.validate(schema, value, "$"); err == nil {
		t.Fatal("invalid primitive type validated successfully")
	}
}

func TestJSONHelpManifestsValidateAgainstPublicManifestSchema(t *testing.T) {
	schema := loadJSONValueFile(t, "../docs/json-schema/v1/manifest.schema.json")
	validator := newSubsetJSONSchemaValidator(t, schema)

	RootCmd.InitDefaultHelpCmd()
	for _, command := range publicCommandSubtree(RootCmd) {
		manifest := jsonCommandManifestFor(command)
		value := normalizeJSONValue(t, manifest)
		if err := validator.validate(schema, value, manifest.Path); err != nil {
			t.Fatalf("manifest %q does not validate: %v", manifest.Path, err)
		}
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

type subsetJSONSchemaValidator struct {
	root any
}

func newSubsetJSONSchemaValidator(t *testing.T, root any) subsetJSONSchemaValidator {
	t.Helper()
	if _, ok := root.(map[string]any); !ok {
		t.Fatalf("schema root is %T, want object", root)
	}
	return subsetJSONSchemaValidator{root: root}
}

func (v subsetJSONSchemaValidator) validate(schema any, value any, path string) error {
	switch schema := schema.(type) {
	case bool:
		if schema {
			return nil
		}
		return fmt.Errorf("%s is disallowed by false schema", path)
	case map[string]any:
		return v.validateObjectSchema(schema, value, path)
	default:
		return fmt.Errorf("%s schema has unsupported type %T", path, schema)
	}
}

func (v subsetJSONSchemaValidator) validateObjectSchema(schema map[string]any, value any, path string) error {
	if refValue, ok := schema["$ref"]; ok {
		ref, ok := refValue.(string)
		if !ok {
			return fmt.Errorf("%s $ref is %T, want string", path, refValue)
		}
		resolved, err := v.resolveRef(ref)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		return v.validate(resolved, value, path)
	}

	if allOf, ok := schema["allOf"]; ok {
		items, ok := allOf.([]any)
		if !ok {
			return fmt.Errorf("%s allOf is %T, want array", path, allOf)
		}
		for i, item := range items {
			if err := v.validate(item, value, fmt.Sprintf("%s allOf[%d]", path, i)); err != nil {
				return err
			}
		}
	}

	if oneOf, ok := schema["oneOf"]; ok {
		items, ok := oneOf.([]any)
		if !ok {
			return fmt.Errorf("%s oneOf is %T, want array", path, oneOf)
		}
		matches := 0
		for _, item := range items {
			if err := v.validate(item, value, path); err == nil {
				matches++
			}
		}
		if matches != 1 {
			return fmt.Errorf("%s matched %d oneOf schemas, want 1", path, matches)
		}
		return nil
	}

	if constValue, ok := schema["const"]; ok && !reflect.DeepEqual(value, constValue) {
		return fmt.Errorf("%s = %v, want const %v", path, value, constValue)
	}

	if enumValue, ok := schema["enum"]; ok {
		enum, ok := enumValue.([]any)
		if !ok {
			return fmt.Errorf("%s enum is %T, want array", path, enumValue)
		}
		found := false
		for _, allowed := range enum {
			if reflect.DeepEqual(value, allowed) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%s = %v, want one of %v", path, value, enum)
		}
	}

	if typeValue, ok := schema["type"]; ok {
		if err := validateJSONSchemaTypeValue(path, value, typeValue); err != nil {
			return err
		}
	}
	if minimumValue, ok := schema["minimum"]; ok {
		if err := validateJSONSchemaMinimum(path, value, minimumValue); err != nil {
			return err
		}
	}

	object, hasObject := value.(map[string]any)
	if requiredValue, ok := schema["required"]; ok {
		if !hasObject {
			return fmt.Errorf("%s required fields on non-object %T", path, value)
		}
		for _, field := range requiredValue.([]any) {
			name := field.(string)
			if _, ok := object[name]; !ok {
				return fmt.Errorf("%s missing required field %q", path, name)
			}
		}
	}

	if propertiesValue, ok := schema["properties"]; ok {
		if !hasObject {
			return fmt.Errorf("%s properties on non-object %T", path, value)
		}
		properties := propertiesValue.(map[string]any)
		if additionalProperties, ok := schema["additionalProperties"]; ok {
			switch additionalProperties := additionalProperties.(type) {
			case bool:
				if !additionalProperties {
					for name := range object {
						if _, ok := properties[name]; !ok {
							return fmt.Errorf("%s has unexpected property %q", path, name)
						}
					}
				}
			case map[string]any:
				for name, propertyValue := range object {
					if _, ok := properties[name]; ok {
						continue
					}
					if err := v.validate(additionalProperties, propertyValue, path+"."+name); err != nil {
						return err
					}
				}
			default:
				return fmt.Errorf("%s additionalProperties is %T, want bool or object", path, additionalProperties)
			}
		}
		for name, propertySchema := range properties {
			if propertyValue, ok := object[name]; ok {
				if err := v.validate(propertySchema, propertyValue, path+"."+name); err != nil {
					return err
				}
			}
		}
	}

	if itemsValue, ok := schema["items"]; ok {
		array, ok := value.([]any)
		if !ok {
			return fmt.Errorf("%s items on non-array %T", path, value)
		}
		for i, item := range array {
			if err := v.validate(itemsValue, item, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v subsetJSONSchemaValidator) resolveRef(ref string) (any, error) {
	if !strings.HasPrefix(ref, "#/") {
		return nil, fmt.Errorf("unsupported ref %q", ref)
	}
	current := v.root
	for _, token := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
		token = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
		object, ok := current.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("ref %q traversed into %T", ref, current)
		}
		next, ok := object[token]
		if !ok {
			return nil, fmt.Errorf("ref %q missing token %q", ref, token)
		}
		current = next
	}
	return current, nil
}

func validateJSONSchemaTypeValue(path string, value any, typeValue any) error {
	switch typeValue := typeValue.(type) {
	case string:
		return validateJSONSchemaType(path, value, typeValue)
	case []any:
		var errors []string
		for _, item := range typeValue {
			typeName, ok := item.(string)
			if !ok {
				return fmt.Errorf("%s type entry is %T, want string", path, item)
			}
			if err := validateJSONSchemaType(path, value, typeName); err == nil {
				return nil
			} else {
				errors = append(errors, err.Error())
			}
		}
		return fmt.Errorf("%s did not match any allowed type: %s", path, strings.Join(errors, "; "))
	default:
		return fmt.Errorf("%s type is %T, want string or array", path, typeValue)
	}
}

func validateJSONSchemaType(path string, value any, typeName string) error {
	switch typeName {
	case "object":
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("%s is %T, want object", path, value)
		}
	case "array":
		if _, ok := value.([]any); !ok {
			return fmt.Errorf("%s is %T, want array", path, value)
		}
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%s is %T, want string", path, value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("%s is %T, want boolean", path, value)
		}
	case "integer":
		number, ok := value.(float64)
		if !ok || number != math.Trunc(number) {
			return fmt.Errorf("%s is %T:%v, want integer", path, value, value)
		}
	case "number":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("%s is %T, want number", path, value)
		}
	default:
		return fmt.Errorf("%s has unsupported schema type %q", path, typeName)
	}
	return nil
}

func validateJSONSchemaMinimum(path string, value any, minimumValue any) error {
	minimum, ok := minimumValue.(float64)
	if !ok {
		return fmt.Errorf("%s minimum is %T, want number", path, minimumValue)
	}
	number, ok := value.(float64)
	if !ok {
		return fmt.Errorf("%s minimum on non-number %T", path, value)
	}
	if number < minimum {
		return fmt.Errorf("%s = %v, want >= %v", path, number, minimum)
	}
	return nil
}
