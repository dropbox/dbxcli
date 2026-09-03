# Machine-contract integration

The installed CLI, not this reference, owns command discovery. JSON help works
without Dropbox authentication:

```sh
dbxcli --help --output=json
dbxcli put --help --output=json
dbxcli share-link create --help --output=json
```

Each help result describes one command. Inspect
`results[].result.supports_structured_output` before normal execution. Its
`input_schema` is JSON Schema for positional arguments and flags. It uses
JSON-friendly names (for example `if_exists`) and retains the CLI spelling in
`x-cli-name`; validate an intended JSON input against it before constructing an
invocation.

Run supported operations with `--output=json`. Stdout is exactly one JSON
success or error envelope; stderr can contain progress, warnings, and
diagnostics. Never parse text output as a fallback.

```sh
dbxcli ls --output=json /
dbxcli put --if-exists fail --output=json report.md /Reports/report.md
```

Success has `ok: true`, `schema_version`, `command`, `input`, `results`, and
`warnings`. Use `results[].status` and `results[].kind` as the stable outcome.
An error has `ok: false` and an `error` object. Check the shell exit status as
well as `error.code`: the latter is the stable remediation key; message text is
human-facing and may change. Known error details are structured context, not a
license to expose sensitive values.

Common exit-code classes: auth (2), permission (3), not found (4), conflict
(5), rate limit (6), validation/unsupported structured output (7), and partial
stdout transfer (8). A rate-limit response may contain
`error.details.retry_after_seconds`; wait only when the user task remains safe
to retry. Never automatically retry a non-idempotent or destructive operation.

`ls` and `search` can span multiple Dropbox result pages, but dbxcli retrieves
those pages internally. Do not create a cursor loop in the agent. Use the
command's discovered `--limit` to bound tool output; a limit caps returned
items and does not establish that the full Dropbox result set has been seen.

For deep search, the discovered `search` manifest exposes `--content` when the
installed CLI supports it. It searches file contents in addition to filenames.
Use it only when the request is specifically about content, and use the
optional Dropbox path scope to avoid a broad account-wide search.

For schema-level validation of help and results, use the public
[JSON schema v1 documentation](https://github.com/dropbox/dbxcli/blob/master/docs/json-schema/v1/README.md):
`manifest.schema.json`, `commands.schema.json`, and `error.schema.json`. Pin a
release tag rather than `master` when a wrapper needs reproducible remote
schema URLs.
