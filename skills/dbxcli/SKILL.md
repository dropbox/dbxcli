---
name: dbxcli
description: Safely operate Dropbox through a locally installed dbxcli command, using its JSON manifest and schema-backed machine contract. Use for Dropbox file, shared-link, team, or account work; do not call the Dropbox API directly.
---

# dbxcli

Use the local `dbxcli` executable as the only Dropbox integration. Do not
reimplement Dropbox API calls, scrape text help, or maintain a command catalog
in this skill. The CLI's JSON help manifest is authoritative for the installed
version.

## Start safely

1. Check availability with `command -v dbxcli`, then run
   `dbxcli version --output=json`. If it is unavailable, say so and give
   installation guidance for the user's operating system, using the
   [dbxcli releases](https://github.com/dropbox/dbxcli/releases) page. Do not
   download or install it unless the user authorizes that action.
2. Before a command you have not already discovered in the current task, run
   `dbxcli [command path] --help --output=json`. Begin with
   `dbxcli --help --output=json` when the command path is unknown.
3. Read the manifest's `supports_structured_output`, `input_schema`,
   `stdin_stdout`, `destructive_level`, `flags`, and `args`. Do not infer a
   command or flag from memory.
4. Represent the intended arguments and flags as JSON-shaped input and validate
   it against that command's `input_schema` before building the shell command.
   Map fields to command-line names using each field's `x-cli-name`.

Read [tool-integration.md](references/tool-integration.md) for the discovery,
validation, result, and error protocol. Read [automation.md](references/automation.md)
for writes and confirmation behavior. Read [safety.md](references/safety.md)
before handling credentials, transfers, deletion, replacement, or sharing.

## Execution contract

For normal command execution, always pass `--output=json` and parse stdout as a
single JSON envelope. Treat stderr as diagnostics only. Check both the process
exit code and `.ok`:

- If `.ok` is `true`, use documented `results[].status`, `results[].kind`, and
  `warnings`; do not rely on prose or undocumented fields.
- If `.ok` is `false`, branch on stable `.error.code`, not `.error.message`.
  Surface a concise, redacted explanation and use structured `.error.details`
  only when relevant. Do not blindly retry writes or auth errors.
- If the manifest says `supports_structured_output: false`, do not run that
  command as a machine-action. Explain the limitation or use a safe supported
  alternative. JSON help itself remains available.

For destructive or externally visible actions, first discover the command and
validate inputs, then prefer `--dry-run` if the manifest exposes it. Use an
explicit `--if-exists` policy whenever it is available; never assume that a
default overwrite or conflict policy matches the user's intent. Require clear
user confirmation before the real destructive action unless the user has
already explicitly requested the exact action. When a command exposes `--yes`,
use it only after that confirmation to prevent an interactive prompt from
blocking automation.

## Large listings, search, and multi-step work

The CLI follows Dropbox pagination internally; agents must not invent or pass
cursors. For a broad `ls` or `search`, discover the command and use its
`--limit` flag to bound the result delivered to the tool. Start with the
narrowest sensible folder or search path; do not recursively enumerate a whole
Dropbox when a scoped query will answer the request. A limited result is a
selection, not proof that no additional matches exist.

For search requests about text *inside* files, inspect the `search` manifest
and pass `--content` only when it is available and the user requested a
content search. Otherwise search is filename-oriented. Scope the search path
and limit whenever practical.

For a search → get → process task: discover and validate `search`, select the
exact result path from its JSON metadata, then discover and validate `get`.
Download to a named local file (never stdout), process that local file, and
report the output path or a concise result. A later upload, share-link, or
replacement is a separate externally visible action and needs its own
discovery, safety policy, and authorization.

## Boundaries

- Never put tokens, auth codes, app secrets, environment dumps, or auth-file
  contents in prompts, commands, logs, JSON fixtures, tool results, commits,
  or artifacts. Refer to secret names and paths only when needed.
- Do not use `DBXCLI_ACCESS_TOKEN=value` inline. Pass an already-provisioned
  secret through the execution environment. Keep `DBXCLI_AUTH_FILE` outside
  the repository and do not read, upload, or commit it.
- Never send binary file data through a tool result. For `get` or
  `share-link download`, download to a local file and report its path and
  metadata. `local operand -` is a byte stream and cannot be combined with
  JSON output.
- Do not use this skill to expose shared links, alter permissions, overwrite,
  move, restore, or delete without user-authorized scope.
