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
Status, human-facing warnings, diagnostics, and verbose logs go to stderr.

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

## List folder contents

```sh
dbxcli share-link ls <url>
dbxcli share-link ls <url> -l
dbxcli share-link ls <url> --path /documents
dbxcli share-link ls <url> --recursive
dbxcli share-link ls <url> --path /documents --recursive --output=json
```

`share-link ls` lists the contents of a folder shared link, including folders
not mounted in your Dropbox. `share-link list` lists shared links themselves.
Without `--recursive`, only immediate children are listed. With `--path`, all
returned paths still start at the shared-link root, so they can be passed to
`share-link download --path` or `share-link info --path`. Use `/` for the root.
Repeated slashes, trailing slashes, and `.` components are normalized like
other Dropbox paths. `..` components, backslashes, and NUL bytes are rejected
before normalization.

Remote names are preserved independently of the local OS, including Unicode,
colons, and Windows device names such as `CON`. Dropbox paths always use `/`;
drive prefixes have no special meaning. Local filesystem restrictions are
checked separately when downloading entries to disk.

By default, text output shows only paths in columns, like `dbxcli ls`.
Use `--long` (`-l`) to show type, human-readable file size (`-` for folders),
and path. Empty folders produce no text in the default format.
Entries are sorted by path. JSON results use status `listed`, kind `file` or
`folder`, and `result` fields `name`, `path`, and `size` (files only, in bytes).
Input metadata includes the URL, optional path, recursive flag, and password
presence; `long: true` is included when `--long` is requested. The password is
never included. JSON entries retain the same metadata in both text formats.
A failed listing returns an error without emitting successful partial results.

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
dbxcli share-link download <url> ./out --path /2027/Logoset --recursive
```

`share-link download` writes to the metadata filename when `target` is omitted.
Use `--path` to select a file or folder inside a folder shared link. Selected
files download with or without `--recursive`. Folders require `--recursive`.
Use `-` as the target to write file bytes to stdout; stdout cannot be combined
with `--recursive` or `--output=json`. Folders cannot be written to stdout.

For a selected folder, only its descendants are downloaded, including empty
folders. If `./out` already exists as a directory, the example above writes
`./out/Logoset/logo.png` and `./out/Logoset/icons/icon.svg`. If `./out` does not
exist, it becomes the selected folder's destination directly, producing
`./out/logo.png` and `./out/icons/icon.svg`. If target is omitted, the selected
folder's metadata name is used. This also works for shared folders not mounted
in your Dropbox and with the existing password flags.

Paths use `/` relative to the original shared-link root and can be copied from
`share-link ls`. Repeated slashes and `.` components are normalized; `..`,
backslashes, and NUL bytes are rejected before normalization. Use the command
without `--path` to download the root. Remote entry names are preserved,
including Unicode, and checked against local filesystem path rules before
writing files.

JSON v1 output retains the URL, selected path, recursive flag, and password
presence in `input`. The `downloaded` result has kind `file` or `folder`, the
actual local `target`, and metadata for the selected item in `link`. A traversal
or download failure returns an error rather than successful partial results;
files already downloaded remain on disk.

## Password-protected links

`share-link create`, `share-link update`, `share-link info`, `share-link ls`, and
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
