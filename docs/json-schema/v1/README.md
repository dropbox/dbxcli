# dbxcli JSON schema v1

These schemas describe the stable top-level JSON envelopes emitted by
`dbxcli --output=json`.

- `success.schema.json` validates successful command responses.
- `error.schema.json` validates command error responses.
- `commands.json` documents command-specific input/result payload names,
  result statuses, result kinds, and warning codes.

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
existing stable error codes, and existing result status meanings should not be
removed or renamed within schema v1.

Error responses always include:

- `ok: false`
- `schema_version: "1"`
- `command`: command path when available, or `dbxcli` for root/pre-parse errors
- `error.message`: human-readable error text
- `error.code`: stable machine-readable error code
- `error.details`: optional machine-readable context, included only when
  dbxcli has reliable structured details such as `argument`, `arguments`,
  `flag`, `flags`, `value`, `path`, `token_type`, `login_command`, `env_var`,
  Dropbox `api_summary`, or Dropbox `api_endpoint`
- `warnings`: machine-actionable warnings, or `[]`

Command results and JSON errors are written to stdout. Status, progress,
human-facing warnings, diagnostics, and verbose logs are written to stderr.
In JSON mode, error responses are written to stdout and the process exits with
a non-zero status.

Commands that intentionally do not support JSON output yet include `login`,
`logout`, and `completion`. Cobra help output and shell-completion protocol
commands are also text-only: `dbxcli --help --output=json`, `dbxcli --output=json`
without a command, and command-specific help such as
`dbxcli version --help --output=json` print text help.

Current JSON-enabled command paths include `version`, `account`, `du`, `ls`,
`search`, `revs`, `cp`, `mv`, `put`, `get`, `share-link create`,
`share-link list`, `share-link info`, `share-link update`,
`share-link revoke`, `share-link download`, `share list folder`,
`share list link`, `team info`, `team list-members`, `team list-groups`,
`team add-member`, `team remove-member`, `mkdir`, `rm`, and `restore`.

Warnings are objects with a stable `code` and human-readable `message`; they
may include optional command-specific details. Current warning codes include
`deprecated_command` for deprecated command paths and `skipped_symlink` for
symlinks skipped by recursive upload.

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
| `permission_denied`             | Dropbox denied access because of permissions, scope, member selection, or state.  |
| `rate_limited`                  | Dropbox rate limited the request.                                                 |
| `dropbox_api_error`             | Dropbox returned an API error that does not map to a more specific code yet.      |
| `structured_output_unsupported` | The command does not support `--output=json` yet.                                 |
| `unknown_command`               | Cobra could not resolve the command.                                              |
| `unknown_flag`                  | Cobra could not resolve a flag.                                                   |
| `command_failed`                | Fallback for failures without a more specific stable code.                        |

Command-specific `input` and `result` payload contracts are listed in
`commands.json` and locked by the golden contract fixtures under
`cmd/testdata/json_contract/`.
