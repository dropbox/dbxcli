# Sharing workflows

`dbxcli share-link` manages Dropbox shared links for files, folders, and nested
paths inside folder shared links.

## Output contract

Text output is the default. With `--output=json`, shared-link commands use the
standard JSON envelope with `ok`, `schema_version`, `command`, `input`,
`results`, and `warnings`. Result statuses include `created`, `existing`,
`listed`, `found`, `updated`, `revoked`, and `downloaded`. Result kinds use the
link or target type, such as `file`, `folder`, or `shared_link`.

In text mode, `share-link create` prints only the shared-link URL to stdout.
Status, warnings, diagnostics, and verbose logs go to stderr.

```sh
url="$(dbxcli share-link create /file.txt)"
```

For automation:

```sh
dbxcli share-link create /file.txt --output=json
dbxcli share-link list /file.txt --output=json
```

## Create shared links

```sh
dbxcli share-link create /file.txt
dbxcli share-link create /file.txt --access viewer
dbxcli share-link create /file.txt --audience team
dbxcli share-link create /file.txt --allow-download
dbxcli share-link create /file.txt --disallow-download
dbxcli share-link create /file.txt --expires 2026-07-01T00:00:00Z
dbxcli share-link create /file.txt --password-prompt
dbxcli share-link create /file.txt --remove-expiration
```

`share-link create` creates a new shared link or returns the existing direct
shared link for the path.

When `share-link create` returns an existing link, `--remove-expiration`
removes the expiration from that existing link.

`share-link create --access` supports `viewer`, `editor`, and `max`. Dropbox
does not support changing access for an existing shared link, so `--access`
fails clearly if the link already exists.

`share-link create --audience` supports `public`, `team`, `members`, and
`no-one`. Dropbox team and folder policies can still resolve the effective
audience differently.

## Inspect and list shared links

```sh
dbxcli share-link info <url>
dbxcli share-link info <url> --path nested/file.txt
dbxcli share-link list
dbxcli share-link list /file.txt
```

`share-link list /file.txt` lists direct shared links for that path, not
parent-folder links.

## Update shared links

```sh
dbxcli share-link update <url> --allow-download
dbxcli share-link update <url> --disallow-download
dbxcli share-link update <url> --audience public
dbxcli share-link update <url> --expires 2026-07-01T00:00:00Z
dbxcli share-link update <url> --remove-expiration
dbxcli share-link update <url> --password-prompt
dbxcli share-link update <url> --remove-password
```

Dropbox account, team, and folder policies can reject shared-link settings such
as passwords, expiration, audience, or disabled downloads. In that case, dbxcli
returns the Dropbox API error, for example `settings_error/not_authorized/`.

## Revoke shared links

```sh
dbxcli share-link revoke <url>
dbxcli share-link revoke --path /file.txt
```

Revoking a file's direct shared link does not revoke links to parent folders
that may still expose the file.

## Download shared links

```sh
dbxcli share-link download <url> [target]
dbxcli share-link download <url> --path nested/file.txt
dbxcli share-link download <url> ./local.txt --path nested/file.txt
dbxcli share-link download <url> [target] --recursive
```

`share-link download` writes to the metadata filename when `target` is omitted.
Use `--path` to download a single file inside a folder shared link. Use `-` as
the target to write file bytes to stdout. Folder shared links require
`--recursive` and cannot be written to stdout.

## Password-protected links

`share-link create`, `share-link update`, `share-link info`, and
`share-link download` support `--password <value>`, `--password-prompt`, and
`--password-file <path>` for password-protected links. Use `--password-prompt`
for interactive use so the password is not echoed.

Avoid `--password <value>` in shared shells because it may be saved in shell
history. Prefer `--password-prompt` for interactive use or `--password-file`
for automation.

## Compatibility commands

Older sharing commands remain available for compatibility:

```sh
dbxcli share list link
dbxcli share list folder
```

Prefer `share-link` commands for shared-link workflows.
