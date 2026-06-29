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

package jsonschema

import (
	"encoding/json"
	"fmt"
	"sort"
)

const (
	Draft202012            = "https://json-schema.org/draft/2020-12/schema"
	CommandSuccessSchemaID = "https://raw.githubusercontent.com/dropbox/dbxcli/master/docs/json-schema/v1/commands.schema.json"
)

type CommandCatalog struct {
	Definitions map[string][]string        `json:"definitions"`
	Commands    map[string]CommandContract `json:"commands"`
}

type CommandContract struct {
	TopLevel      string   `json:"top_level"`
	ResultWrapper string   `json:"result_wrapper"`
	Input         string   `json:"input"`
	ResultInput   *string  `json:"result_input"`
	Result        string   `json:"result"`
	Statuses      []string `json:"statuses"`
	Kinds         []string `json:"kinds"`
	Warnings      []string `json:"warnings"`
}

func DecodeCommandCatalog(data []byte) (CommandCatalog, error) {
	var catalog CommandCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return CommandCatalog{}, err
	}
	if len(catalog.Definitions) == 0 {
		return CommandCatalog{}, fmt.Errorf("command catalog has no definitions")
	}
	if len(catalog.Commands) == 0 {
		return CommandCatalog{}, fmt.Errorf("command catalog has no commands")
	}
	return catalog, nil
}

func GenerateCommandSuccessSchema(catalog CommandCatalog) (map[string]any, error) {
	defs := map[string]any{
		"warning": warningSchema(),
	}

	for _, name := range sortedMapKeys(catalog.Definitions) {
		defs[name] = objectSchema(catalog.Definitions[name])
	}

	commandRefs := make([]any, 0, len(catalog.Commands))
	for _, command := range sortedMapKeys(catalog.Commands) {
		contract := catalog.Commands[command]
		if err := validateCommandContract(command, contract, catalog.Definitions); err != nil {
			return nil, err
		}

		commandDef := commandSchemaDefinitionName(command)
		resultDef := commandResultDefinitionName(command)
		warningDef := commandWarningDefinitionName(command)

		defs[resultDef] = commandResultSchema(contract)
		defs[warningDef] = commandWarningsSchema(contract.Warnings)
		defs[commandDef] = commandSuccessSchema(command, contract, resultDef, warningDef)
		commandRefs = append(commandRefs, ref(commandDef))
	}

	return map[string]any{
		"$schema": Draft202012,
		"$id":     CommandSuccessSchemaID,
		"title":   "dbxcli command-specific JSON success responses",
		"type":    "object",
		"oneOf":   commandRefs,
		"$defs":   defs,
	}, nil
}

func MarshalCanonical(value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func validateCommandContract(command string, contract CommandContract, definitions map[string][]string) error {
	if contract.TopLevel != "operation_output" {
		return fmt.Errorf("%s: unsupported top_level %q", command, contract.TopLevel)
	}
	if contract.ResultWrapper != "operation_result" {
		return fmt.Errorf("%s: unsupported result_wrapper %q", command, contract.ResultWrapper)
	}
	for _, refName := range []string{contract.Input, contract.Result} {
		if _, ok := definitions[refName]; !ok {
			return fmt.Errorf("%s: unknown definition %q", command, refName)
		}
	}
	if contract.ResultInput == nil {
		return fmt.Errorf("%s: missing result_input", command)
	}
	if _, ok := definitions[*contract.ResultInput]; !ok {
		return fmt.Errorf("%s: unknown result_input definition %q", command, *contract.ResultInput)
	}
	if len(contract.Statuses) == 0 {
		return fmt.Errorf("%s: missing statuses", command)
	}
	if len(contract.Kinds) == 0 {
		return fmt.Errorf("%s: missing kinds", command)
	}
	return nil
}

func commandSuccessSchema(command string, contract CommandContract, resultDef string, warningDef string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"ok", "schema_version", "command", "input", "results", "warnings"},
		"properties": map[string]any{
			"ok":             map[string]any{"const": true},
			"schema_version": map[string]any{"const": "1"},
			"command":        map[string]any{"const": command},
			"input":          ref(contract.Input),
			"results": map[string]any{
				"type":  "array",
				"items": ref(resultDef),
			},
			"warnings": ref(warningDef),
		},
	}
}

func commandResultSchema(contract CommandContract) map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"status", "kind", "input", "result"},
		"properties": map[string]any{
			"status": map[string]any{"enum": sortedCopy(contract.Statuses)},
			"kind":   map[string]any{"enum": sortedCopy(contract.Kinds)},
			"input":  ref(*contract.ResultInput),
			"result": ref(contract.Result),
		},
	}
}

func commandWarningsSchema(codes []string) map[string]any {
	if len(codes) == 0 {
		return map[string]any{
			"type":  "array",
			"items": false,
		}
	}
	return map[string]any{
		"type": "array",
		"items": map[string]any{
			"allOf": []any{
				ref("warning"),
				map[string]any{
					"type": "object",
					"properties": map[string]any{
						"code": map[string]any{"enum": sortedCopy(codes)},
					},
				},
			},
		},
	}
}

func objectSchema(fields []string) map[string]any {
	properties := make(map[string]any, len(fields))
	for _, field := range fields {
		properties[field] = map[string]any{}
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties":           properties,
	}
}

func warningSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"code", "message"},
		"properties": map[string]any{
			"code":    map[string]any{"type": "string"},
			"message": map[string]any{"type": "string"},
			"path":    map[string]any{"type": "string"},
		},
	}
}

func ref(defName string) map[string]any {
	return map[string]any{"$ref": "#/$defs/" + defName}
}

func commandSchemaDefinitionName(command string) string {
	return CommandDefinitionName(command)
}

func commandResultDefinitionName(command string) string {
	return schemaDefinitionName("result", command)
}

func commandWarningDefinitionName(command string) string {
	return schemaDefinitionName("warnings", command)
}

func CommandDefinitionName(command string) string {
	return schemaDefinitionName("command", command)
}

func schemaDefinitionName(prefix string, value string) string {
	result := prefix + "_"
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			result += string(r)
		case r >= 'A' && r <= 'Z':
			result += string(r)
		case r >= '0' && r <= '9':
			result += string(r)
		case r == '_':
			result += string(r)
		default:
			result += fmt.Sprintf("_%x", r)
		}
	}
	return result
}

func sortedMapKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedCopy(values []string) []string {
	copied := append([]string{}, values...)
	sort.Strings(copied)
	return copied
}
