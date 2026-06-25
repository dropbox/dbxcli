# `dbxcli`: A command line tool for Dropbox users and team admins [UNOFFICIAL]

[![CI](https://github.com/dropbox/dbxcli/actions/workflows/ci.yml/badge.svg)](https://github.com/dropbox/dbxcli/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dropbox/dbxcli)](https://goreportcard.com/report/github.com/dropbox/dbxcli)

:warning: WARNING: This project is **NOT official**. What does this mean?

  * There is no formal Dropbox support for this project
  * Bugs may or may not get fixed
  * Not all SDK features may be implemented and implemented features may be buggy or incorrect

## Features

  * Supports basic file operations like ls, cp, mkdir, mv, rm (via the Files API)
  * Supports search with sorting and flexible time formatting
  * Supports file revisions and file restore
  * Chunked uploads for large files, paginated listing for large directories
  * Recursive directory uploads (`put -r`) and downloads (`get -r`)
  * Retry with exponential backoff for uploads and downloads
  * Supports a growing set of Team operations

## Installation

### Homebrew (macOS and Linux)

```sh
brew install dbxcli
```

### Linux

Download the archive for your architecture, verify its checksum, and install. Replace `X.Y.Z` with the latest version from the [Releases](https://github.com/dropbox/dbxcli/releases) page (without the leading `v`).

```sh
curl -LO https://github.com/dropbox/dbxcli/releases/download/vX.Y.Z/dbxcli_X.Y.Z_linux_amd64.tar.gz
curl -LO https://github.com/dropbox/dbxcli/releases/download/vX.Y.Z/SHA256SUMS
grep 'dbxcli_X.Y.Z_linux_amd64.tar.gz' SHA256SUMS | sha256sum -c -
tar -xzf dbxcli_X.Y.Z_linux_amd64.tar.gz
sudo mv dbxcli_X.Y.Z_linux_amd64/dbxcli /usr/local/bin/
```

For ARM systems, use `linux_arm64` or `linux_arm` instead of `linux_amd64`.

### macOS (manual)

If you prefer not to use Homebrew:

```sh
curl -LO https://github.com/dropbox/dbxcli/releases/download/vX.Y.Z/dbxcli_X.Y.Z_darwin_arm64.tar.gz
curl -LO https://github.com/dropbox/dbxcli/releases/download/vX.Y.Z/SHA256SUMS
grep 'dbxcli_X.Y.Z_darwin_arm64.tar.gz' SHA256SUMS | shasum -a 256 -c -
tar -xzf dbxcli_X.Y.Z_darwin_arm64.tar.gz
sudo mv dbxcli_X.Y.Z_darwin_arm64/dbxcli /usr/local/bin/
```

Use `darwin_amd64` for Intel Macs.

### Windows

Download `dbxcli_X.Y.Z_windows_amd64.zip` from the [Releases](https://github.com/dropbox/dbxcli/releases) page, extract it, and add the directory to your `PATH`.

### Release assets

All release archives are available at the [releases](https://github.com/dropbox/dbxcli/releases) page:

* `dbxcli_X.Y.Z_darwin_amd64.tar.gz`
* `dbxcli_X.Y.Z_darwin_arm64.tar.gz`
* `dbxcli_X.Y.Z_linux_amd64.tar.gz`
* `dbxcli_X.Y.Z_linux_arm64.tar.gz`
* `dbxcli_X.Y.Z_linux_arm.tar.gz`
* `dbxcli_X.Y.Z_openbsd_amd64.tar.gz`
* `dbxcli_X.Y.Z_windows_amd64.zip`
* `SHA256SUMS`

### Instructions for building yourself

1. Make sure `git` and `go` are installed.
2. Install the latest released version:
   ```sh
   $ go install github.com/dropbox/dbxcli@latest
   ```
3. Or build from source:
   ```sh
   $ git clone https://github.com/dropbox/dbxcli.git
   $ cd dbxcli
   $ go build .
   ```

To use your own Dropbox app while developing, provide its app key when logging in:

```sh
$ dbxcli login --app-key=your-app-key
```

## Usage

`dbxcli` is largely self documenting. Run `dbxcli -h` for a list of supported commands:

```sh
$ dbxcli --help
Use dbxcli to quickly interact with your Dropbox, upload/download files,
manage your team and more. It is easy, scriptable and works on all platforms!

Usage:
  dbxcli [command]

Available Commands:
  account     Display account information
  completion  Generate the autocompletion script for the specified shell
  cp          Copy a file or folder to a different location
  du          Display usage information
  get         Download a file or folder
  login       Log in and save Dropbox credentials
  logout      Log out of the current session
  ls          List files and folders
  mkdir       Create a new directory
  mv          Move files
  put         Upload files or directories
  restore     Restore files
  revs        List file revisions
  rm          Remove files
  search      Search
  share       Sharing commands
  team        Team management commands
  version     Print version information

Flags:
      --as-member string   Member ID to perform action as
      --output string      Output format: text, json (default "text")
  -v, --verbose            Enable verbose logging

Use "dbxcli [command] --help" for more information about a command.
```

### Output formats

Text output is the default. JSON output is available through the global `--output` flag as commands are migrated:

```sh
$ dbxcli <command> --output=json
$ dbxcli version --output=json
$ dbxcli account --output=json
$ dbxcli du --output=json
$ dbxcli ls --output=json /
$ dbxcli search --output=json report /Reports
$ dbxcli revs --output=json /Reports/old.pdf
$ dbxcli cp --output=json /Reports/old.pdf /Reports/copy.pdf
$ dbxcli mv --output=json /Reports/copy.pdf /Reports/archive/copy.pdf
$ dbxcli put --output=json README.md /README.md
$ dbxcli get --output=json /Reports/old.pdf ./old.pdf
$ dbxcli share-link create --output=json /Reports/old.pdf
$ dbxcli share-link list --output=json /Reports/old.pdf
$ dbxcli share-link info --output=json https://www.dropbox.com/s/example/old.pdf
$ dbxcli share-link update --output=json https://www.dropbox.com/s/example/old.pdf --expires 2026-07-01T00:00:00Z
$ dbxcli share-link revoke --output=json https://www.dropbox.com/s/example/old.pdf
$ dbxcli share-link download --output=json https://www.dropbox.com/s/example/old.pdf ./old.pdf
$ dbxcli share list folder --output=json
$ dbxcli team info --output=json
$ dbxcli team list-members --output=json
$ dbxcli team list-groups --output=json
$ dbxcli team add-member --output=json user@example.com User Name
$ dbxcli team remove-member --output=json user@example.com
$ dbxcli mkdir --output=json /new-folder
$ dbxcli rm --output=json /old-file.txt
$ dbxcli restore --output=json /Reports/old.pdf 015f...
```

Structured success output is rolling out command by command. Currently migrated commands are `version`, `account`, `du`, `ls`, `search`, `revs`, `cp`, `mv`, `put`, `get`, `share-link create`, `share-link list`, `share-link info`, `share-link update`, `share-link revoke`, `share-link download`, `share list folder`, `share list link`, `team info`, `team list-members`, `team list-groups`, `team add-member`, `team remove-member`, `mkdir`, `rm`, and `restore`. Commands that have not been migrated return a JSON error whose `error.message` is `structured output is not supported for this command yet` when used with `--output=json`.

Command results and JSON errors are written to stdout. Status, progress, human-facing warnings, diagnostics, and verbose logs are written to stderr. JSON errors include a `warnings` array for machine-actionable warnings; it is `[]` when no warnings are present. Successful JSON payloads use the same `warnings` field.
Current warning codes include `deprecated_command` for deprecated command paths and `skipped_symlink` for symlinks skipped by recursive upload.

Commands that intentionally do not support JSON output yet include `login`, `logout`, and `completion`. Cobra help output and shell-completion protocol commands are also text-only: `dbxcli --help --output=json`, `dbxcli --output=json` without a command, and command-specific help such as `dbxcli version --help --output=json` print text help.

JSON error responses use stable `error.code` values:

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

Successful JSON responses for migrated commands return `ok: true`, `schema_version: "1"`, `command`, an `input` object, a `results` array, and a `warnings` array. Result payloads are command-specific. Public top-level schemas and the command contract catalog live under [docs/json-schema/v1](docs/json-schema/v1/). If a multi-target or recursive command fails after some side effects have already happened, dbxcli returns a JSON error envelope and does not include partial success results. For commands such as `mkdir`, each result reports what happened to the requested path:

```json
{
  "ok": true,
  "schema_version": "1",
  "command": "mkdir",
  "input": {
    "path": "/new-folder",
    "parents": false
  },
  "results": [
    {
      "status": "created",
      "kind": "folder",
      "input": {
        "path": "/new-folder",
        "parents": false
      },
      "result": {
        "type": "folder",
        "path_display": "/new-folder",
        "path_lower": "/new-folder",
        "id": "id:..."
      }
    }
  ],
  "warnings": []
}
```

For `cp` and `mv`, each result input object uses `from_path` and `to_path`:

```json
{
  "ok": true,
  "schema_version": "1",
  "command": "cp",
  "input": {},
  "results": [
    {
      "status": "copied",
      "kind": "file",
      "input": {
        "from_path": "/Reports/old.pdf",
        "to_path": "/Reports/copy.pdf"
      },
      "result": {
        "type": "file",
        "path_display": "/Reports/copy.pdf",
        "path_lower": "/reports/copy.pdf",
        "id": "id:...",
        "rev": "...",
        "size": 123
      }
    }
  ],
  "warnings": []
}
```

For commands such as `rm`, `input` uses command-specific path and flag fields:

```json
{
  "ok": true,
  "schema_version": "1",
  "command": "rm",
  "input": {},
  "results": [
    {
      "status": "deleted",
      "kind": "file",
      "input": {
        "path": "/old-file.txt",
        "permanent": false,
        "recursive": false,
        "force": false
      },
      "result": {
        "type": "file",
        "path_display": "/old-file.txt",
        "path_lower": "/old-file.txt",
        "id": "id:...",
        "rev": "...",
        "size": 123
      }
    }
  ],
  "warnings": []
}
```

`put` always returns a `results` array, including single-file and stdin uploads. The top-level `input` describes the command request; each result reports whether a file was `uploaded`, `skipped`, or a directory was `created` or already `existing`:

```json
{
  "ok": true,
  "schema_version": "1",
  "command": "put",
  "input": {
    "source": "README.md",
    "target": "/README.md",
    "recursive": false,
    "if_exists": "overwrite",
    "stdin": false
  },
  "results": [
    {
      "status": "uploaded",
      "kind": "file",
      "input": {
        "source": "README.md",
        "target": "/README.md"
      },
      "result": {
        "type": "file",
        "path_display": "/README.md",
        "path_lower": "/readme.md",
        "id": "id:...",
        "rev": "...",
        "size": 123
      }
    }
  ],
  "warnings": []
}
```

`get` also returns a top-level `input`, a `results` array, and `warnings: []`. File downloads use `downloaded`; recursive folder downloads may also include local directory results with `created` or `existing`.

Entry-list commands such as `ls`, `search`, and `revs` use the operation-style wrapper. `ls` input includes the listed path; `search` input includes the query and optional path scope; `revs` input includes the file path. Results use `listed`, `found`, or `revision` status values and put Dropbox metadata under `result`:

```json
{
  "ok": true,
  "schema_version": "1",
  "command": "ls",
  "input": {
    "path": "/Reports",
    "recursive": false,
    "include_deleted": false,
    "only_deleted": false,
    "long": false,
    "reverse": false,
    "time": "server"
  },
  "results": [
    {
      "status": "listed",
      "kind": "file",
      "input": {},
      "result": {
        "type": "file",
        "path_display": "/Reports/q1.pdf",
        "path_lower": "/reports/q1.pdf",
        "id": "id:...",
        "rev": "...",
        "size": 123
      }
    }
  ],
  "warnings": []
}
```

Version, account, and usage commands use the operation-style wrapper with a single result:

```json
{
  "ok": true,
  "schema_version": "1",
  "command": "version",
  "input": {},
  "results": [
    {
      "status": "reported",
      "kind": "version",
      "input": {},
      "result": {
        "version": "3.4.0",
        "sdk_version": "6.0.5",
        "spec_version": "c36ba27"
      }
    }
  ],
  "warnings": []
}
```

```json
{
  "ok": true,
  "schema_version": "1",
  "command": "account",
  "input": {},
  "results": [
    {
      "status": "found",
      "kind": "account",
      "input": {},
      "result": {
        "type": "full",
        "account_id": "dbid:...",
        "email": "user@example.com",
        "email_verified": true,
        "disabled": false
      }
    }
  ],
  "warnings": []
}
```

```json
{
  "ok": true,
  "schema_version": "1",
  "command": "du",
  "input": {},
  "results": [
    {
      "status": "reported",
      "kind": "space_usage",
      "input": {},
      "result": {
        "used": 123,
        "allocation": {
          "type": "individual",
          "allocated": 100000
        }
      }
    }
  ],
  "warnings": []
}
```

Shared-link commands use the same operation-style wrapper. `share-link create`, `list`, `info`, and `update` put shared-link metadata directly under `result`; status values include `created`, `existing`, `listed`, `found`, and `updated`:

```json
{
  "ok": true,
  "schema_version": "1",
  "command": "share-link create",
  "input": {
    "path": "/Reports/old.pdf"
  },
  "results": [
    {
      "status": "created",
      "kind": "file",
      "input": {},
      "result": {
        "type": "file",
        "url": "https://www.dropbox.com/s/...",
        "name": "old.pdf",
        "path_lower": "/reports/old.pdf",
        "rev": "...",
        "size": 123
      }
    }
  ],
  "warnings": []
}
```

`share-link revoke` uses `revoked` results whose `result` contains the revoked URL and, when available, the shared-link metadata. `share-link download` uses `downloaded` results whose `result` contains the local `target` and `link` metadata.

The legacy `share list folder` command also supports operation-style JSON. It uses `listed` results with `shared_folder` metadata:

```json
{
  "ok": true,
  "schema_version": "1",
  "command": "share list folder",
  "input": {},
  "results": [
    {
      "status": "listed",
      "kind": "shared_folder",
      "input": {},
      "result": {
        "type": "shared_folder",
        "name": "Reports",
        "path_lower": "/reports",
        "shared_folder_id": "1234567890",
        "preview_url": "https://www.dropbox.com/scl/fo/...",
        "access_type": "owner",
        "is_inside_team_folder": false,
        "is_team_folder": false
      }
    }
  ],
  "warnings": []
}
```

Team commands use the same operation-style wrapper. `team info` returns a single `team` result, `team list-members` and `team list-groups` return `listed` results, and mutating member commands return the Dropbox launch status:

```json
{
  "ok": true,
  "schema_version": "1",
  "command": "team list-members",
  "input": {},
  "results": [
    {
      "status": "listed",
      "kind": "team_member",
      "input": {},
      "result": {
        "type": "team_member",
        "team_member_id": "dbmid:...",
        "email": "user@example.com",
        "email_verified": true,
        "status": "active",
        "role": "member_only"
      }
    }
  ],
  "warnings": []
}
```

`get --output=json <source> -` and `share-link download --output=json <url> -` are not supported because stdout is reserved for downloaded file bytes when the target is `-`.

In JSON mode, command errors are written to stdout as JSON, including errors from commands that do not yet support structured success output. The process still exits with a non-zero status. Detailed diagnostics may also be written to stderr:

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

Error `code` values are stable identifiers intended for scripts. The current stable codes are listed in the table above.

### Authentication

By default, `dbxcli` stores OAuth credentials in `~/.config/dbxcli/auth.json`.
Run `dbxcli login` to authorize dbxcli and save credentials:

```sh
$ dbxcli login
```

Commands require saved credentials. If no saved credentials are available, run
`dbxcli login` first or provide a token with `DBXCLI_ACCESS_TOKEN`.

Personal and team logins use bundled Dropbox app keys by default. You can pass
a custom app key as an option:

```sh
$ dbxcli login --app-key=your-app-key
```

You can also set custom app keys with environment variables:

```sh
$ DROPBOX_PERSONAL_APP_KEY=your-app-key dbxcli login
$ DROPBOX_TEAM_APP_KEY=your-app-key dbxcli login team-access
$ DROPBOX_MANAGE_APP_KEY=your-app-key dbxcli login team-manage
```

Saved login credentials include a Dropbox refresh token and are refreshed
automatically when the access token expires. If saved credentials are revoked or
need to be replaced, run `dbxcli login` again.

Set `DBXCLI_AUTH_FILE` to use a different credentials file:

```sh
$ DBXCLI_AUTH_FILE=/path/to/auth.json dbxcli login
```

For automation with short-lived Dropbox access tokens, set `DBXCLI_ACCESS_TOKEN`.
This token is used directly and is not saved or refreshed. If it expires, the
command fails and you must provide a fresh token:

```sh
$ DBXCLI_ACCESS_TOKEN=sl.xxxxxx dbxcli ls /
```

### Listing files

```sh
$ dbxcli ls -l /Photos
Revision              Size    Last modified Path
abc123                1.2 MiB 3 weeks ago   /Photos/vacation.jpg
def456                4.5 MiB 1 month ago   /Photos/family.png
```

#### Time format

By default, `ls -l`, `search -l`, and `revs -l` show relative timestamps ("3 weeks ago"). Use `--time-format` for absolute dates:

```sh
$ dbxcli ls -l --time-format=short /Photos
Revision              Size    Last modified    Path
abc123                1.2 MiB 2026-05-15 10:30 /Photos/vacation.jpg

$ dbxcli ls -l --time-format=rfc3339 /Photos
Revision              Size    Last modified        Path
abc123                1.2 MiB 2026-05-15T10:30:00Z /Photos/vacation.jpg
```

Use `--time=client` to display client-modified time instead of server-modified (default):

```sh
$ dbxcli ls -l --time=client --time-format=short /Photos
```

#### Sorting

Sort results with `--sort` and optionally `--reverse`:

```sh
$ dbxcli ls -l --sort=size /Documents          # smallest first
$ dbxcli ls -l --sort=size --reverse /Documents # largest first
$ dbxcli ls -l --sort=name /Documents           # alphabetical
$ dbxcli ls -l --sort=time /Documents           # oldest first
$ dbxcli ls -l --sort=type /Documents           # folders, files, deleted
```

### Searching

```sh
$ dbxcli search -l --time-format=short --sort=size "report"
```

All `--sort`, `--reverse`, `--time`, and `--time-format` flags work with both `ls` and `search`. The `--time` and `--time-format` flags also work with `revs -l`.

### Sharing

Create shared links:

```sh
$ dbxcli share-link create /file.txt # create or return an existing shared link
$ dbxcli share-link create /file.txt --access viewer # create a link with requested access
$ dbxcli share-link create /file.txt --audience team # create a link with requested audience
$ dbxcli share-link create /file.txt --allow-download # create a downloadable shared link
$ dbxcli share-link create /file.txt --disallow-download # create a shared link with downloads disabled
$ dbxcli share-link create /file.txt --expires 2026-07-01T00:00:00Z # create an expiring shared link
$ dbxcli share-link create /file.txt --password-prompt # create a password-protected shared link
$ dbxcli share-link create /file.txt --remove-expiration # remove expiration when returning an existing link
```

Inspect and list shared links:

```sh
$ dbxcli share-link info <url>       # display shared link information
$ dbxcli share-link info <url> --path /nested/file.txt # display information for a path inside the shared link
$ dbxcli share-link list             # list existing shared links
$ dbxcli share-link list /file.txt   # list direct shared links for a path
```

Download shared links:

```sh
$ dbxcli share-link download <url> [target] # download a shared-link file
$ dbxcli share-link download <url> --path /nested/file.txt # download a file inside a folder shared link
$ dbxcli share-link download <url> ./local.txt --path /nested/file.txt # download nested file to a local target
$ dbxcli share-link download <url> [target] --recursive # download a folder shared link
```

Update shared links:

```sh
$ dbxcli share-link update <url> --allow-download # update shared link settings
$ dbxcli share-link update <url> --disallow-download # disable downloads from a shared link
$ dbxcli share-link update <url> --audience public # update shared link audience
$ dbxcli share-link update <url> --expires 2026-07-01T00:00:00Z # update shared link expiration
$ dbxcli share-link update <url> --remove-expiration # remove shared link expiration
$ dbxcli share-link update <url> --password-prompt # set or change a shared link password
$ dbxcli share-link update <url> --remove-password # remove a shared link password
```

Revoke shared links:

```sh
$ dbxcli share-link revoke <url>     # revoke a shared link
$ dbxcli share-link revoke --path /file.txt # revoke direct shared links for a path
```

Compatibility and shared folders:

```sh
$ dbxcli share list link             # deprecated compatibility command
$ dbxcli share list folder           # list shared folders
```

`share-link create --access` supports `viewer`, `editor`, and `max`. Dropbox does not support changing access for an existing shared link, so `--access` fails clearly if the link already exists.

`share-link create --audience` and `share-link update --audience` support `public`, `team`, `members`, and `no-one`. Dropbox team and folder policies can still resolve the effective audience differently.

Dropbox account, team, and folder policies can reject shared-link settings such as passwords, expiration, audience, or disabled downloads. In that case, dbxcli returns the Dropbox API error, for example `settings_error/not_authorized/`.

`share-link create`, `share-link update`, `share-link info`, and `share-link download` support `--password <value>`, `--password-prompt`, and `--password-file <path>` for password-protected links. Use `--password-prompt` for interactive use so the password is not echoed.

`share-link download` writes to the metadata filename when `target` is omitted. Use `--path` to download a single file inside a folder shared link. Use `-` as the target to write file bytes to stdout. Folder shared links require `--recursive` and cannot be written to stdout.

New and changed commands should write command results to stdout. Status, progress, human-facing warnings, diagnostics, and verbose logs should go to stderr. Machine-actionable JSON warnings should use the `warnings` array.

### Team management

```sh
$ dbxcli team --help
Team management commands

Usage:
  dbxcli team [command]

Available Commands:
  add-member    Add a new member to a team
  info          Get team information
  list-groups   List groups
  list-members  List team members
  remove-member Remove member from a team

Global Flags:
      --as-member string   Member ID to perform action as
      --output string      Output format: text, json (default "text")
  -v, --verbose            Enable verbose logging

Use "dbxcli team [command] --help" for more information about a command.
```

The `--verbose` option will turn on verbose logging and is useful for debugging.

### Uploading files and directories

```sh
$ dbxcli put file.txt /destination/file.txt        # upload a single file
$ dbxcli put -r ./project /backup/project          # recursively upload a directory
$ dbxcli put -r -w 8 ./large-folder /backup/large  # use 8 workers per large file
$ dbxcli put --if-exists skip file.txt /dest.txt   # skip if the file already exists
```

By default, `put` overwrites existing destination files. Use `--if-exists overwrite|skip|fail` to choose whether existing files are overwritten, skipped, or treated as an error.

### Downloading files and directories

```sh
$ dbxcli get /remote/file.txt ./local-file.txt     # download a single file
$ dbxcli get -r /remote/folder ./local-folder      # recursively download a folder
```

### Piping with stdin/stdout

Use `-` as the local operand to stream through pipes:

```sh
$ printf 'hello' | dbxcli put - /hello.txt         # upload from stdin
$ tar cz ./src | dbxcli put - /backups/src.tgz     # pipe archive to Dropbox
$ dbxcli get /backups/src.tgz - | tar tz           # download to stdout and list
$ dbxcli get /file.txt - > local-copy.txt          # download to stdout, redirect to file
```

Stdin uploads are spooled to a temp file before uploading, so disk space up to the full input size is required. Stdout downloads are byte-clean: all progress and diagnostic output goes to stderr.

A bare `-` means stream only when it is the local operand. Dropbox paths named `-` are valid, for example `dbxcli put - /-` and `dbxcli get /- -`. To upload a local file literally named `-`, use `./-`.

### Removing files and folders

```sh
$ dbxcli rm /remote/file.txt                       # move a file to Dropbox trash
$ dbxcli rm -r /remote/folder                      # remove a non-empty folder
$ dbxcli rm --permanent /remote/file.txt           # permanently delete when Dropbox permits it
```

### Creating directories

```sh
$ dbxcli mkdir /projects/2026/reports   # creates all intermediate directories
$ dbxcli mkdir -p /projects/2026/reports # no error if directory already exists
```

## Contributing

 * Step 1: If you're submitting a non-trivial change, please fill out the [Dropbox Contributor License Agreement](https://opensource.dropbox.com/cla/) first.
 * Step 2: send a [pull request](https://help.github.com/articles/using-pull-requests/)
 * Step 3: Profit!
 
## Useful Resources

* [Go SDK documentation](https://godoc.org/github.com/dropbox/dropbox-sdk-go-unofficial)
* [API documentation](https://www.dropbox.com/developers/documentation/http/documentation)
