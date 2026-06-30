# dbxcli JSON schema v1

These schemas describe the stable top-level JSON envelopes emitted by
`dbxcli --output=json` and JSON help responses emitted by
`dbxcli --help --output=json`.

- `success.schema.json` validates successful command responses.
- `error.schema.json` validates command error responses.
- `manifest.schema.json` validates each command manifest object emitted by
  JSON help.
- `commands.json` documents command-specific input/result payload names,
  result statuses, result kinds, and warning codes.
- `commands.schema.json` validates command-specific success responses using
  `commands.json`: exact command names, top-level input fields, per-result
  input/result fields, result statuses/kinds, and warning codes.

Successful responses always include:

- `ok: true`
- `schema_version: "1"`
- `command`: command path without the binary name, such as `ls` or
  `share-link create`
- `input`: command-specific request fields
- `results`: command-specific result objects; every result includes `status`,
  `kind`, `input`, and `result`
- `warnings`: machine-actionable warnings, or `[]`

Schema v1 is intended to be stable. New fields, commands, warning codes, and
error details may be added in minor releases. Existing top-level fields,
existing stable error codes, and existing result status meanings will not be
removed or renamed within schema v1.

JSON responses must not include access tokens, refresh tokens, authorization
codes, or app secrets.

Error responses always include:

- `ok: false`
- `schema_version: "1"`
- `command`: command path when available, or `dbxcli` for root/pre-parse errors
- `error.message`: human-readable error text
- `error.code`: stable machine-readable error code
- `error.details`: optional machine-readable context, included only when
  dbxcli has reliable structured details such as `argument`, `arguments`,
  `flag`, `flags`, `value`, `path`, `token_type`, `login_command`, `env_var`,
  Dropbox `api_summary`, Dropbox `api_endpoint`, or `bytes_written`
- `warnings`: machine-actionable warnings, or `[]`

Command results and JSON errors are written to stdout. Status, progress,
human-facing warnings, diagnostics, and verbose logs are written to stderr.
In JSON mode, error responses are written to stdout and the process exits with
a non-zero status.

Commands that intentionally do not support structured command-result JSON yet
include `login` and `completion`. Their help output is still
available as a JSON command manifest with `--help --output=json`; for example,
`dbxcli --help --output=json`, `dbxcli version --help --output=json`, and
`dbxcli --output=json help version`. `dbxcli --output=json` without `--help`
continues to print text help, and shell-completion protocol commands remain
text-only.

The current JSON-enabled command paths are listed in `commands.json`.
`commands.schema.json` is generated from that catalog plus schema metadata in
the generator:

```sh
go run ./tools/gen-json-schemas
```

The command-specific schema locks which fields may appear, which fields are
required when stable, primitive field types, stable nested objects, result
statuses/kinds, and warning codes. It intentionally avoids over-modeling
Dropbox SDK enum values that dbxcli does not own.

## Command Manifest v1

JSON help is the canonical command manifest surface:

```sh
dbxcli --help --output=json
dbxcli put --help --output=json
dbxcli --output=json help put
```

Each manifest result has `status: "described"`, `kind: "command"`, and a
`result` object that validates against `manifest.schema.json`.

Manifest v1 keeps the original JSON help fields and adds machine-contract
metadata:

- `manifest_version: "1"`
- structured `args`
- enriched `flags`, including enum values, conflicts, prompt behavior, and
  sensitive inputs
- generated `input_schema` for command arguments and flags, including enum
  values when an argument or flag has a bounded value set
- `examples`
- `schema_refs`
- best-effort audited `dropbox_scopes`
- `stdin_stdout`
- `result_statuses`, `result_kinds`, and `warning_codes`

`input_schema` is a JSON Schema object generated from the command manifest's
structured positional arguments and flags. It is intended for tool callers,
MCP-style integrations, and automation planners that need to validate command
inputs before building a CLI invocation. It excludes `--help` and `--output`,
uses JSON-friendly property names such as `if_exists`, and preserves the
original CLI names in `x-cli-name`. Flags that are accepted globally but do not
affect a no-auth command may be omitted from that command's `input_schema`.
Sensitive inputs are marked with `writeOnly` and `x-sensitive`; flag conflicts
are listed in `x-conflicts`.

For commands with structured JSON output, `schema_refs.command_success_schema`
points to the command-specific definition inside `commands.schema.json`.
`schema_refs.command_contract` points to the source catalog entry in
`commands.json`.

`scope_accuracy` is currently `audited_best_effort` for commands with audited
manifest metadata. Scope metadata is intended for planning and diagnostics;
Dropbox API errors remain the source of truth at runtime.

`account` results include auth context under `result.auth`:
`result.auth.source` is `saved` or `env`; `result.auth.refreshable` is a
boolean; and `result.auth.auth_file` is `default`, `custom`, or `none`.
dbxcli does not include the full auth file path by default.

Warnings are objects with a stable `code` and human-readable `message`; they
may include optional command-specific details. JSON responses from deprecated
command paths include `deprecated_command`. Current warning codes include
`deprecated_command` for deprecated command paths and `skipped_symlink` for
symlinks skipped by recursive upload. `logout` may return `token_revoke_failed`
when saved credentials were removed locally but one or more Dropbox tokens could
not be revoked remotely.

Stable error codes:

| Code                            | Meaning                                                                           |
|---------------------------------|-----------------------------------------------------------------------------------|
| `invalid_arguments`             | The command arguments or flags are invalid.                                       |
| `path_conflict`                 | A local or Dropbox path conflicts with the requested operation.                   |
| `auth_required`                 | No usable saved credentials were found, or Dropbox rejected the saved token.      |
| `auth_refresh_failed`           | Saved refreshable credentials could not be refreshed.                             |
| `app_key_required`              | Login or token refresh needs a Dropbox app key.                                   |
| `auth_exchange_failed`          | The OAuth authorization-code exchange failed or returned unusable tokens.         |
| `not_found`                     | Dropbox reported that the requested object was not found.                         |
| `partial_transfer`              | A download-to-stdout stream failed after partial output was already written.      |
| `permission_denied`             | Dropbox denied access because of permissions, scope, member selection, or state.  |
| `rate_limited`                  | Dropbox rate limited the request.                                                 |
| `dropbox_api_error`             | Dropbox returned an API error that does not map to a more specific code yet.      |
| `env_token_still_active`        | `DBXCLI_ACCESS_TOKEN` is set and must be unset before logout can complete.        |
| `structured_output_unsupported` | The command does not support `--output=json` yet.                                 |
| `unknown_command`               | Cobra could not resolve the command.                                              |
| `unknown_flag`                  | Cobra could not resolve a flag.                                                   |
| `command_failed`                | Fallback for failures without a more specific stable code.                        |

Command-specific `input` and `result` payload contracts are listed in
`commands.json`, validated through `commands.schema.json`, and locked by the
golden contract fixtures under `cmd/testdata/json_contract/`.
