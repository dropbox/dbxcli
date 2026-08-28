# Write operations and interaction

Discover the specific command first. The manifest tells you whether it supports
`--dry-run`, `--if-exists`, `--yes`, structured output, and prompts. These flags
are command-specific; never attach one speculatively.

Use `--dry-run` to preview a user-authorized mutation when it is offered. A
successful preview is not permission to perform the real action: obtain or use
the user's explicit confirmation for the real scope.

When `--if-exists` is available, pass an explicit value. Typical policies are
`fail`, `skip`, and `autorename`; some commands also offer `overwrite`. Select
only a policy compatible with the user's stated intent. In particular, do not
silently choose `overwrite`.

Use `--yes` only after confirmation has been established and only if the
discovered manifest exposes it. It acknowledges an operation; it does not
replace authorization.

Before an automated job that needs Dropbox access, use
`dbxcli account --output=json` as an auth and identity check. Prefer a
short-lived, pre-provisioned `DBXCLI_ACCESS_TOKEN` in the execution environment
for CI. When saved credentials are required, set `DBXCLI_AUTH_FILE` to a
private secret-backed or temporary location outside the repository. Do not
commit, cache, upload, or print that file.

The full public source is
[Automation and JSON output](https://github.com/dropbox/dbxcli/blob/master/docs/automation.md);
this reference intentionally does not duplicate its command catalog or schema.
