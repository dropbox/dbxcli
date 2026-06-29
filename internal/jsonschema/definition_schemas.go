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

type definitionSchemaConfig struct {
	Required   []string
	Properties map[string]any
}

func commandDefinitionSchema(name string) definitionSchemaConfig {
	config, ok := commandDefinitionSchemas[name]
	if !ok {
		return definitionSchemaConfig{}
	}
	return config
}

var commandDefinitionSchemas = map[string]definitionSchemaConfig{
	"account": {
		Required: []string{"account_id", "disabled", "email", "email_verified", "type"},
		Properties: map[string]any{
			"auth": schemaRef("account_auth"),
			"name": schemaRef("account_name"),
			"team": schemaRef("account_team"),
			"type": stringEnum("basic", "full"),
		},
	},
	"account_auth": {
		Required: []string{"refreshable", "source"},
	},
	"account_name": {
		Properties: stringProperties("abbreviated_name", "display_name", "familiar_name", "given_name", "surname"),
	},
	"account_team": {
		Required: []string{"id", "name"},
	},
	"command_arg": {
		Required: []string{"description", "enum_values", "name", "placement", "required", "stream_dash", "value_kind", "variadic"},
		Properties: map[string]any{
			"enum_values": stringArraySchema(),
			"required":    booleanSchema(),
			"stream_dash": booleanSchema(),
			"variadic":    booleanSchema(),
		},
	},
	"command_example": {
		Required: []string{"command", "description"},
	},
	"command_flag": {
		Required: []string{"conflicts", "default", "enum_values", "inherited", "may_prompt", "name", "required", "sensitive", "shorthand", "type", "usage", "value_kind"},
		Properties: map[string]any{
			"conflicts":   stringArraySchema(),
			"enum_values": stringArraySchema(),
			"inherited":   booleanSchema(),
			"may_prompt":  booleanSchema(),
			"required":    booleanSchema(),
			"sensitive":   booleanSchema(),
		},
	},
	"command_input_property": {
		Required: []string{"type"},
		Properties: map[string]any{
			"default":       map[string]any{"type": []string{"array", "boolean", "integer", "string"}},
			"enum":          stringArraySchema(),
			"items":         schemaRef("command_input_property"),
			"minItems":      integerSchema(),
			"type":          stringEnum("array", "boolean", "integer", "string"),
			"writeOnly":     booleanSchema(),
			"x-conflicts":   stringArraySchema(),
			"x-inherited":   booleanSchema(),
			"x-may-prompt":  booleanSchema(),
			"x-sensitive":   booleanSchema(),
			"x-stream-dash": booleanSchema(),
		},
	},
	"command_input_schema": {
		Required: []string{"additionalProperties", "properties", "required", "type"},
		Properties: map[string]any{
			"additionalProperties": map[string]any{"const": false},
			"properties": map[string]any{
				"type":                 "object",
				"additionalProperties": schemaRef("command_input_property"),
			},
			"required": stringArraySchema(),
			"type":     map[string]any{"const": "object"},
		},
	},
	"command_manifest": {
		Required: []string{
			"aliases",
			"args",
			"auth_modes",
			"destructive_level",
			"dropbox_scopes",
			"examples",
			"flags",
			"input_schema",
			"manifest_version",
			"may_prompt",
			"path",
			"result_kinds",
			"result_statuses",
			"runnable",
			"schema_refs",
			"scope_accuracy",
			"short",
			"stdin_stdout",
			"supports_structured_output",
			"use",
			"warning_codes",
		},
		Properties: map[string]any{
			"aliases":                    stringArraySchema(),
			"args":                       arraySchema(schemaRef("command_arg")),
			"auth_modes":                 stringArraySchema(),
			"destructive_level":          stringEnum("admin", "delete", "none"),
			"dropbox_scopes":             stringArraySchema(),
			"examples":                   arraySchema(schemaRef("command_example")),
			"flags":                      arraySchema(schemaRef("command_flag")),
			"input_schema":               schemaRef("command_input_schema"),
			"manifest_version":           map[string]any{"const": "1"},
			"may_prompt":                 booleanSchema(),
			"result_kinds":               stringArraySchema(),
			"result_statuses":            stringArraySchema(),
			"runnable":                   booleanSchema(),
			"schema_refs":                schemaRef("command_schema_refs"),
			"stdin_stdout":               schemaRef("command_stdin_stdout"),
			"supports_structured_output": booleanSchema(),
			"warning_codes":              stringArraySchema(),
		},
	},
	"command_schema_refs": {
		Required: []string{"error_schema", "success_schema"},
	},
	"command_stdin_stdout": {
		Required: []string{"reads_stdin", "stderr", "stdout", "writes_binary_stdout"},
		Properties: map[string]any{
			"reads_stdin":          booleanSchema(),
			"stderr":               stringSchema(),
			"stdout":               stringSchema(),
			"writes_binary_stdout": booleanSchema(),
		},
	},
	"du_allocation": {
		Required: []string{"type"},
	},
	"du_output": {
		Required: []string{"allocation", "used"},
		Properties: map[string]any{
			"allocation": schemaRef("du_allocation"),
		},
	},
	"get_input": {
		Required: []string{"recursive", "source", "stdout", "target"},
	},
	"get_result_input": {
		Required: []string{"source", "target"},
	},
	"help_input": {
		Required: []string{"help", "path"},
	},
	"logout_result": {
		Required: []string{"remote_token_revoked", "removed_saved_credentials"},
	},
	"ls_input": {
		Required: []string{"include_deleted", "long", "only_deleted", "path", "recursive", "reverse"},
		Properties: map[string]any{
			"sort": stringEnum("name", "size", "time", "type"),
			"time": stringEnum("client", "server"),
		},
	},
	"metadata": {
		Required: []string{"type"},
		Properties: map[string]any{
			"type": stringEnum("deleted", "file", "folder"),
		},
	},
	"mkdir_input": {
		Required: []string{"parents", "path"},
	},
	"operation_output": {
		Required: []string{"command", "input", "ok", "results", "schema_version", "warnings"},
		Properties: map[string]any{
			"ok":             map[string]any{"const": true},
			"results":        arraySchema(schemaRef("operation_result")),
			"schema_version": map[string]any{"const": "1"},
			"warnings":       arraySchema(schemaRef("warning")),
		},
	},
	"operation_result": {
		Required: []string{"input", "kind", "result", "status"},
		Properties: map[string]any{
			"input":  map[string]any{"type": "object"},
			"result": map[string]any{},
		},
	},
	"put_input": {
		Required: []string{"if_exists", "recursive", "source", "stdin", "target"},
		Properties: map[string]any{
			"if_exists": stringEnum("fail", "overwrite", "skip"),
		},
	},
	"put_result_input": {
		Required: []string{"source", "target"},
	},
	"relocation_input": {
		Required: []string{"from_path", "to_path"},
	},
	"remove_input": {
		Required: []string{"force", "path", "permanent", "recursive"},
	},
	"restore_input": {
		Required: []string{"path", "revision"},
	},
	"revs_input": {
		Required: []string{"long", "path"},
		Properties: map[string]any{
			"time": stringEnum("client", "server"),
		},
	},
	"search_input": {
		Required: []string{"content", "long", "query", "reverse"},
		Properties: map[string]any{
			"order_by": stringEnum("modified", "relevance"),
			"sort":     stringEnum("name", "size", "time", "type"),
			"time":     stringEnum("client", "server"),
		},
	},
	"share_folder": {
		Required: []string{"type"},
		Properties: map[string]any{
			"owner_display_names": stringArraySchema(),
			"type":                stringEnum("shared_folder"),
		},
	},
	"share_link_create_input": {
		Required: []string{"path"},
		Properties: map[string]any{
			"access":   stringEnum("editor", "max", "viewer"),
			"audience": stringEnum("members", "no-one", "public", "team"),
		},
	},
	"share_link_download_input": {
		Required: []string{"url"},
	},
	"share_link_download_result": {
		Required: []string{"target"},
		Properties: map[string]any{
			"link": schemaRef("share_link_metadata"),
		},
	},
	"share_link_info_input": {
		Required: []string{"url"},
	},
	"share_link_list_input": {
		Required: []string{"direct_only"},
	},
	"share_link_metadata": {
		Required: []string{"type", "url"},
		Properties: map[string]any{
			"permissions": schemaRef("share_link_permissions"),
			"type":        stringEnum("file", "folder", "link"),
		},
	},
	"share_link_permissions": {
		Required: []string{
			"allow_comments",
			"allow_download",
			"can_allow_download",
			"can_disallow_download",
			"can_remove_expiry",
			"can_revoke",
			"can_set_expiry",
		},
	},
	"share_link_revoke_result": {
		Required: []string{"url"},
		Properties: map[string]any{
			"link": schemaRef("share_link_metadata"),
		},
	},
	"share_link_update_input": {
		Required: []string{"url"},
		Properties: map[string]any{
			"audience": stringEnum("members", "no-one", "public", "team"),
		},
	},
	"team_group": {
		Required: []string{"type"},
		Properties: map[string]any{
			"type": stringEnum("team_group"),
		},
	},
	"team_info": {
		Required: []string{"type"},
		Properties: map[string]any{
			"type": stringEnum("team"),
		},
	},
	"team_member": {
		Required: []string{"type"},
		Properties: map[string]any{
			"groups": stringArraySchema(),
			"name":   schemaRef("account_name"),
			"type":   stringEnum("team_member"),
		},
	},
	"team_member_add_input": {
		Required: []string{"email", "first_name", "last_name"},
	},
	"team_member_add_item": {
		Properties: map[string]any{
			"member": schemaRef("team_member"),
		},
	},
	"team_member_mutation": {
		Required: []string{"type"},
		Properties: map[string]any{
			"results": arraySchema(schemaRef("team_member_add_item")),
			"type":    stringEnum("team_member_add", "team_member_remove"),
		},
	},
	"team_member_remove_input": {
		Required: []string{"email"},
	},
	"version": {
		Required: []string{"sdk_version", "spec_version", "version"},
	},
}

func defaultPropertySchema(field string) map[string]any {
	switch field {
	case "aliases", "auth_modes", "conflicts", "dropbox_scopes", "enum", "enum_values", "groups", "owner_display_names", "required", "result_kinds", "result_statuses", "warning_codes", "x-conflicts":
		return stringArraySchema()
	case "allocated", "limit", "member_count", "num_licensed_users", "num_provisioned_users", "size", "used", "user_within_team_space_allocated", "user_within_team_space_used_cached":
		return integerSchema()
	case "additionalProperties", "allow_comments", "allow_download", "can_allow_download", "can_disallow_download", "can_remove_expiry", "can_remove_password", "can_revoke", "can_set_expiry", "can_set_password", "can_use_extended_sharing_controls", "content", "deleted", "direct_only", "disabled", "disallow_download", "email_verified", "force", "help", "include_deleted", "inherited", "is_directory_restricted", "is_inside_team_folder", "is_paired", "is_team_folder", "is_teammate", "long", "may_prompt", "only_deleted", "parents", "password", "permanent", "recursive", "refreshable", "remote_token_revoked", "remove_expiration", "remove_password", "removed_saved_credentials", "require_password", "reverse", "runnable", "sensitive", "stdin", "stdout", "stream_dash", "supports_structured_output", "variadic", "writeOnly", "writes_binary_stdout", "x-inherited", "x-may-prompt", "x-sensitive", "x-stream-dash":
		return booleanSchema()
	case "client_modified", "expires", "invited_on", "joined_on", "server_modified", "suspended_on", "time_invited":
		return dateTimeStringSchema()
	default:
		return stringSchema()
	}
}

func schemaRef(defName string) map[string]any {
	return map[string]any{"$ref": "#/$defs/" + defName}
}

func stringProperties(fields ...string) map[string]any {
	properties := make(map[string]any, len(fields))
	for _, field := range fields {
		properties[field] = stringSchema()
	}
	return properties
}

func stringSchema() map[string]any {
	return map[string]any{"type": "string"}
}

func booleanSchema() map[string]any {
	return map[string]any{"type": "boolean"}
}

func integerSchema() map[string]any {
	return map[string]any{
		"type":    "integer",
		"minimum": 0,
	}
}

func dateTimeStringSchema() map[string]any {
	return map[string]any{
		"type":   "string",
		"format": "date-time",
	}
}

func stringArraySchema() map[string]any {
	return arraySchema(stringSchema())
}

func arraySchema(items map[string]any) map[string]any {
	return map[string]any{
		"type":  "array",
		"items": items,
	}
}

func stringEnum(values ...string) map[string]any {
	return map[string]any{
		"type": "string",
		"enum": sortedCopy(values),
	}
}
