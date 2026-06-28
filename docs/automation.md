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

Use JSON help to discover command paths, flags, aliases, known auth modes, known
destructive levels, and whether normal structured command output is supported.

## Safe scripting patterns

Prefer flags that make automation outcomes explicit:

```sh
dbxcli put --if-exists fail report.md /Reports/report.md
dbxcli put --if-exists skip report.md /Reports/report.md --output=json
dbxcli rm /Reports/old-report.md --output=json
```

Use `--output=json` when the caller needs stable statuses, result kinds,
warnings, or error codes. Use text output when a command is part of a human
terminal workflow or when the command intentionally writes file bytes to stdout.

## Authentication for automation

By default, `dbxcli` stores OAuth credentials in:

```text
~/.config/dbxcli/auth.json
```

Use `dbxcli login` to save refreshable credentials:

```sh
dbxcli login
```

Use `DBXCLI_ACCESS_TOKEN` for automation with short-lived Dropbox access tokens:

```sh
DBXCLI_ACCESS_TOKEN=sl.xxxxxx dbxcli ls --output=json /
```

This token is used directly and is not saved or refreshed. If it expires, the
command fails and you must provide a fresh token.

Set `DBXCLI_AUTH_FILE` to use a different saved credentials file:

```sh
DBXCLI_AUTH_FILE=/path/to/auth.json dbxcli login
DBXCLI_AUTH_FILE=/path/to/auth.json dbxcli ls /
```

`dbxcli logout` revokes saved Dropbox tokens and removes local saved
credentials. If `DBXCLI_ACCESS_TOKEN` is set, unset it before running logout;
environment-provided tokens are not saved locally and cannot be removed by
dbxcli.

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

A bare `-` means stdin/stdout only when it appears as the local operand. Dropbox
paths named `-` are valid, for example `dbxcli put - /-` and `dbxcli get /- -`.
To upload a local file literally named `-`, use `./-`.

## Exit status

Current behavior:

* `0` means success.
* Non-zero means failure.
* In JSON mode, inspect `error.code` for stable machine handling.

Specific error-code-to-exit-code mapping is planned as a future automation
milestone.

## Shell completion

Shell completion commands intentionally remain text-only because shells expect
completion scripts or protocol output:

```sh
dbxcli completion bash
dbxcli completion zsh
dbxcli completion fish
dbxcli completion powershell
```
