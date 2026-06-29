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

package main

import (
	"fmt"
	"os"

	"github.com/dropbox/dbxcli/v3/internal/jsonschema"
)

const (
	commandCatalogPath       = "docs/json-schema/v1/commands.json"
	commandSuccessSchemaPath = "docs/json-schema/v1/commands.schema.json"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if _, err := os.Stat("go.mod"); err != nil {
		return fmt.Errorf("run from repository root: %w", err)
	}

	data, err := os.ReadFile(commandCatalogPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", commandCatalogPath, err)
	}

	catalog, err := jsonschema.DecodeCommandCatalog(data)
	if err != nil {
		return fmt.Errorf("decode %s: %w", commandCatalogPath, err)
	}

	schema, err := jsonschema.GenerateCommandSuccessSchema(catalog)
	if err != nil {
		return fmt.Errorf("generate command success schema: %w", err)
	}

	encoded, err := jsonschema.MarshalCanonical(schema)
	if err != nil {
		return fmt.Errorf("marshal command success schema: %w", err)
	}

	if err := os.WriteFile(commandSuccessSchemaPath, encoded, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", commandSuccessSchemaPath, err)
	}
	return nil
}
