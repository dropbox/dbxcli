# Safety and data handling

Treat tokens, refresh tokens, authorization codes, app secrets, and auth files
as secrets. Never ask users to paste them into a chat or command line; do not
read or display an auth file. Avoid environment dumps and shell tracing. Redact
any secret accidentally present in command output before reporting it.

`dbxcli get <remote> -` and `dbxcli share-link download <url> -` write raw
bytes to stdout. They cannot use `--output=json`. Do not use these forms when a
tool captures stdout, because binary data can corrupt a tool result or consume
context. Use a named local destination instead, then report only safe metadata
such as the destination path, size, and checksum if necessary.

Likewise, do not upload binary data into a chat transcript. For a local source,
pass its path to the CLI. For a generated stream, use a pipe only when the
execution environment will not return those bytes as a tool result.

Treat delete, overwrite, move, restore, permission changes, team member
changes, and creation or sharing of public links as meaningful external
effects. Scope them to the user's request, preview when available, and confirm
before executing the real mutation.

## Scope verification

Before executing a mutation, verify that the operation matches the exact scope
the user authorized. Check:

- Exact Dropbox paths or shared-link targets
- File vs folder type (if a path resolves to a folder when a file was expected, confirm)
- Single vs multiple targets (if dry-run shows more than expected, stop)
- Recursive behavior (do not add `--recursive` unless requested)
- Permanent vs recoverable deletion (do not add `--permanent` unless explicitly requested)
- Overwrite/conflict policy (use explicit `--if-exists`, do not silently choose `overwrite`)

Do not silently expand scope beyond the user's request. If the requested target,
operation mode, or affected path count differs from what the user described,
stop and confirm the actual scope before proceeding.

The public [security policy](https://github.com/dropbox/dbxcli/blob/master/SECURITY.md)
and [automation contract](https://github.com/dropbox/dbxcli/blob/master/docs/automation.md)
contain the authoritative protocol and credential details.
