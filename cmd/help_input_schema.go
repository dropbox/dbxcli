package cmd

import (
	"strconv"
	"strings"
)

func commandInputSchemaFor(args []jsonCommandArg, flags []jsonCommandFlag) jsonCommandInputSchema {
	schema := jsonCommandInputSchema{
		Type:                 "object",
		AdditionalProperties: false,
		Required:             []string{},
		Properties:           map[string]jsonCommandInputProperty{},
	}

	for _, arg := range args {
		name := commandInputPropertyName(arg.Name)
		property := jsonCommandInputProperty{
			Type:        commandInputPropertyType(arg.ValueKind, ""),
			Description: arg.Description,
			XCLIKind:    "arg",
			XCLIName:    arg.Name,
			XValueKind:  arg.ValueKind,
			XStreamDash: arg.StreamDash,
		}
		if arg.Variadic {
			property.Type = "array"
			property.Items = &jsonCommandInputProperty{Type: commandInputPropertyType(arg.ValueKind, "")}
			if arg.Required {
				property.MinItems = 1
			}
		}
		if format := commandInputPropertyFormat(arg.ValueKind); format != "" && !arg.Variadic {
			property.Format = format
		}
		schema.Properties[name] = property
		if arg.Required {
			schema.Required = append(schema.Required, name)
		}
	}

	for _, flag := range flags {
		if !commandInputSchemaIncludesFlag(flag) {
			continue
		}
		name := commandInputPropertyName(flag.Name)
		propertyType := commandInputPropertyType(flag.ValueKind, flag.Type)
		property := jsonCommandInputProperty{
			Type:        propertyType,
			Description: flag.Usage,
			Enum:        sortedCopyStringSlice(flag.EnumValues),
			XCLIKind:    "flag",
			XCLIName:    flag.Name,
			XValueKind:  flag.ValueKind,
			XSensitive:  flag.Sensitive,
			XConflicts:  commandInputPropertyNames(flag.Conflicts),
			XInherited:  flag.Inherited,
			XShorthand:  flag.Shorthand,
			XMayPrompt:  flag.MayPrompt,
		}
		if format := commandInputPropertyFormat(flag.ValueKind); format != "" {
			property.Format = format
		}
		if flag.Sensitive {
			property.WriteOnly = true
		}
		if value, ok := commandInputFlagDefault(flag, propertyType); ok {
			property.Default = value
		}
		schema.Properties[name] = property
		if flag.Required {
			schema.Required = append(schema.Required, name)
		}
	}

	return schema
}

func commandInputSchemaIncludesFlag(flag jsonCommandFlag) bool {
	switch flag.Name {
	case "help", outputFlag:
		return false
	default:
		return true
	}
}

func commandInputPropertyNames(names []string) []string {
	result := make([]string, 0, len(names))
	for _, name := range names {
		result = append(result, commandInputPropertyName(name))
	}
	return sortedCopyStringSlice(result)
}

func commandInputPropertyName(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

func commandInputPropertyType(valueKind string, flagType string) string {
	switch valueKind {
	case "boolean":
		return "boolean"
	case "bytes", "integer":
		return "integer"
	case "enum", "string", "dropbox_path", "local_path", "dropbox_member_id",
		"dropbox_app_key", "local_file", "secret", "rfc3339_timestamp",
		"url", "email", "account_id", "auth_type", "command_path", "revision":
		return "string"
	}

	switch flagType {
	case "bool":
		return "boolean"
	case "int", "int64", "uint64":
		return "integer"
	default:
		return "string"
	}
}

func commandInputPropertyFormat(valueKind string) string {
	switch valueKind {
	case "email":
		return "email"
	case "rfc3339_timestamp":
		return "date-time"
	case "url":
		return "uri"
	default:
		return ""
	}
}

func commandInputFlagDefault(flag jsonCommandFlag, propertyType string) (any, bool) {
	if flag.Default == "" {
		return nil, false
	}

	switch propertyType {
	case "boolean":
		value, err := strconv.ParseBool(flag.Default)
		if err != nil {
			return nil, false
		}
		return value, true
	case "integer":
		value, err := strconv.ParseInt(flag.Default, 10, 64)
		if err != nil {
			return nil, false
		}
		return value, true
	default:
		return flag.Default, true
	}
}
