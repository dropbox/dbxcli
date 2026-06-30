# Automation and JSON output

`dbxcli` supports text output for humans, structured JSON command output for
scripts, and JSON help for machine-readable command discovery.

Command results and JSON errors are written to stdout. Status, progress,
human-facing warnings, diagnostics, and verbose logs are written to stderr.

## JSON command output

Text output is the default. JSON command output is available through the global
`--output` flag for commands that support structured execution output:

```sh
dbxcli ls --output=json /
dbxcli account --output=json
dbxcli logout --output=json
```

Use JSON help to discover whether a command supports structured command output:

```sh
dbxcli put --help --output=json
```

JSON help is also the Command Manifest v1 surface for tools and agents. Command
manifests expose machine-readable metadata such as structured positional
arguments, flag enum values and conflicts, prompt/sensitive-input metadata,
examples, auth modes, best-effort Dropbox scopes, stdin/stdout behavior, schema
refs, result statuses/kinds, and known warning codes when available.

Successful JSON responses use a stable envelope:

```json
{
  "ok": true,
  "schema_version": "1",
  "command": "ls",
  "input": {
    "path": "/Reports"
  },
  "results": [
    {
      "status": "listed",
      "kind": "file",
      "input": {},
      "result": {
        "type": "file",
        "path_display": "/Reports/q1.pdf",
        "path_lower": "/reports/q1.pdf"
      }
    }
  ],
  "warnings": []
}
```

In JSON mode, error responses are written to stdout and the process exits with
a non-zero status:

```json
{
  "ok": false,
  "schema_version": "1",
  "command": "rm",
  "error": {
    "message": "path exists and is not a folder: /old-file.txt",
    "code": "path_conflict"
  },
  "warnings": []
}
```

The full JSON command catalog, stable error codes, and schemas live in
[json-schema/v1](json-schema/v1/README.md).
Use `commands.schema.json` from that directory when a caller needs
command-specific success validation for `input`, `results`, primitive field
types, statuses, kinds, and warning codes.

## Schema-first automation

Automation should treat the CLI and schemas as the stable interface:

* Use `dbxcli --help --output=json` for command discovery.
* Use each manifest's `input_schema` to validate arguments and flags before
  building a CLI invocation.
* Use `commands.schema.json` to validate successful JSON responses.
* Use `error.schema.json` to validate JSON error responses.
* Prefer schema URLs from a pinned release tag when reproducibility matters.

dbxcli currently does not expose a separate machine protocol. Tools should
invoke the CLI, read stdout as JSON in `--output=json` mode, and treat stderr as
status, progress, warnings, diagnostics, and verbose logs.

## JSON help manifest

JSON help is the machine-readable command discovery surface. It is separate
from command execution output, does not require Dropbox auth, and works even for
commands that do not support structured command execution output.

```sh
dbxcli --help --output=json
dbxcli put --help --output=json
dbxcli share-link create --help --output=json
dbxcli --output=json help share-link create
```

Use JSON help to discover command paths, structured args, flags, generated
input schemas, aliases, known auth modes, known destructive levels,
stdin/stdout behavior, schema refs, and whether normal structured command
output is supported.

Each manifest result includes `input_schema`, a JSON Schema object for the
command's CLI inputs. It uses JSON-friendly names such as `if_exists`, includes
enum values for bounded arguments and flags, and preserves original CLI names in
`x-cli-name`, so tools can validate structured input and then build the correct
argument/flag invocation.

## Safe scripting patterns

Prefer flags that make automation outcomes explicit:

```sh
dbxcli put --if-exists fail report.md /Reports/report.md
dbxcli put --if-exists skip report.md /Reports/report.md --output=json
dbxcli share-link create /Reports/report.md --output=json
```

Use `--output=json` when the caller needs stable statuses, result kinds,
warnings, or error codes. Use text output when a command is part of a human
terminal workflow or when the command intentionally writes file bytes to stdout.

Check auth and identity before running a job:

```sh
dbxcli account --output=json
```

Use direct token auth for short-lived CI jobs:

```sh
DBXCLI_ACCESS_TOKEN="$DROPBOX_ACCESS_TOKEN" dbxcli ls --output=json /
```

Use a dedicated saved-auth file when a job needs refreshable credentials. Store
the file in a private temp or secret-backed path, not in the repository working
directory:

```sh
export DBXCLI_AUTH_FILE="${RUNNER_TEMP:-/tmp}/dbxcli-auth.json"
dbxcli login
dbxcli ls /
```

Do not commit, cache, or upload this file as an artifact; it can contain refresh
tokens.

Capture JSON output and handle failure by exit code:

```sh
if ! dbxcli put --if-exists fail --output=json report.md /Reports/report.md >result.json; then
  jq -r '.error.code + ": " + .error.message' result.json >&2
  exit 1
fi

jq -r '.results[] | [.status, .kind] | @tsv' result.json
```

GitHub Actions example:

This example assumes `dbxcli` is already installed on the runner.

```yaml
jobs:
  publish-report:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - name: Upload report to Dropbox
        env:
          DBXCLI_ACCESS_TOKEN: ${{ secrets.DROPBOX_ACCESS_TOKEN }}
        run: |
          dbxcli account --output=json
          dbxcli put --if-exists fail --output=json report.md /Reports/report.md
```

## Authentication for automation

By default, `dbxcli` stores OAuth credentials in:

```text
~/.config/dbxcli/auth.json
```

Use `dbxcli login` to save refreshable credentials:

```sh
dbxcli login
```

Use `dbxcli account --output=json` as an auth and identity check. The account
result includes `result.auth`:

`result.auth` example:

```json
{
  "auth": {
    "source": "saved",
    "refreshable": true,
    "auth_file": "default"
  }
}
```

Stable auth fields:

* `result.auth.source`: `saved` or `env`
* `result.auth.refreshable`: boolean
* `result.auth.auth_file`: `default`, `custom`, or `none`

Use `DBXCLI_ACCESS_TOKEN` for automation with short-lived Dropbox access tokens:

```sh
DBXCLI_ACCESS_TOKEN=sl.xxxxxx dbxcli ls --output=json /
```

This token is used directly and is not saved or refreshed. If it expires, the
command fails and you must provide a fresh token.

`result.auth` example:

```json
{
  "auth": {
    "source": "env",
    "refreshable": false,
    "auth_file": "none"
  }
}
```

Set `DBXCLI_AUTH_FILE` to use a different saved credentials file:

```sh
DBXCLI_AUTH_FILE=/path/to/auth.json dbxcli login
DBXCLI_AUTH_FILE=/path/to/auth.json dbxcli ls /
```

`result.auth` example:

```json
{
  "auth": {
    "source": "saved",
    "refreshable": true,
    "auth_file": "custom"
  }
}
```

`dbxcli logout` revokes saved Dropbox tokens and removes local saved
credentials. Environment-provided tokens are not saved locally, so `dbxcli`
cannot remove them. If `DBXCLI_ACCESS_TOKEN` is set, unset it before running
logout.

## Stdin and stdout

Use `-` as the local operand to stream through pipes:

```sh
printf 'hello' | dbxcli put - /hello.txt
tar cz ./src | dbxcli put - /backups/src.tgz
dbxcli get /backups/src.tgz - | tar tz
dbxcli get /file.txt - > local-copy.txt
```

Stdin uploads are spooled to a temp file before uploading, so disk space up to
the full input size is required. Stdout downloads are byte-clean: progress,
diagnostic output, human-facing warnings, and verbose logs go to stderr.

Commands that write file bytes to stdout cannot also write JSON results to
stdout. `dbxcli get <path> - --output=json` and
`dbxcli share-link download <url> - --output=json` return `invalid_arguments`.

A bare `-` means stdin/stdout only when it appears as the local operand. Dropbox
paths named `-` are valid, for example `dbxcli put - /-` and `dbxcli get /- -`.
To upload a local file literally named `-`, use `./-`.

## Exit status

`dbxcli` uses stable exit codes for shell scripts, CI jobs, and agents. Text
and JSON output modes use the same exit-code contract. Successful commands
exit `0`, including successful commands that return warnings.

| Exit code | Meaning | JSON error codes |
|-----------|---------|------------------|
| `0` | Success | none |
| `1` | Generic error | `command_failed`, `dropbox_api_error` |
| `2` | Auth failure | `auth_required`, `auth_refresh_failed`, `auth_exchange_failed`, `app_key_required`, `env_token_still_active` |
| `3` | Permission denied | `permission_denied` |
| `4` | Not found | `not_found` |
| `5` | Conflict | `path_conflict` |
| `6` | Rate limited | `rate_limited` |
| `7` | Validation or usage error | `invalid_arguments`, `unknown_command`, `unknown_flag`, `structured_output_unsupported` |
| `8` | Partial stdout transfer | `partial_transfer` |

In JSON mode, inspect both the process exit code and `error.code` for the most
specific machine-readable failure reason.

## Shell completion

Completion script/protocol output is text-only because shells expect completion
scripts or protocol output:

```sh
dbxcli completion bash
dbxcli completion zsh
dbxcli completion fish
dbxcli completion powershell
```

Passing `--output=json` to completion commands returns
`structured_output_unsupported`.
