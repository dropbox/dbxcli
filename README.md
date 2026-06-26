# `dbxcli`: A command line tool for Dropbox users and team admins

[![CI](https://github.com/dropbox/dbxcli/actions/workflows/ci.yml/badge.svg)](https://github.com/dropbox/dbxcli/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dropbox/dbxcli)](https://goreportcard.com/report/github.com/dropbox/dbxcli)

## Support posture

`dbxcli` is maintained in the Dropbox GitHub organization by Dropbox engineers, but it is not a formally supported Dropbox product. Use GitHub issues and pull requests for bugs and contributions; Dropbox Support does not provide support for this CLI. The CLI implements a practical subset of Dropbox API features, not the full API surface.

## Features

  * File operations: `ls`, `cp`, `mkdir`, `mv`, `rm`, `put`, and `get`
  * Recursive upload and download with `put -r` and `get -r`
  * Pipe-friendly transfers: upload from stdin with `put -` and download to stdout with `get ... -`
  * Upload conflict control with `put --if-exists overwrite|skip|fail`
  * Shared-link management with `share-link create`, `list`, `info`, `update`, `revoke`, and `download`
  * Structured JSON output with stable success and error envelopes for migrated commands
  * Search, file revisions, restore, flexible sorting, and time formatting
  * Chunked uploads for large files and paginated listing for large directories
  * OAuth login with refreshable saved credentials, `DBXCLI_ACCESS_TOKEN`, and `DBXCLI_AUTH_FILE`
  * Team administration commands and member-scoped access with `--as-member`

## Quickstart

```sh
dbxcli login
dbxcli ls /
dbxcli put local.txt /remote.txt
dbxcli get /remote.txt ./remote.txt
dbxcli share-link create /remote.txt
```

See the [GitHub Releases](https://github.com/dropbox/dbxcli/releases) page for version-by-version changes.

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

`dbxcli` is largely self-documenting. Run `dbxcli -h` for a list of supported commands:

```sh
$ dbxcli --help
Use dbxcli to quickly interact with your Dropbox, upload/download files,
manage your team and more. It is easy, scriptable and works on all platforms!

Usage:
  dbxcli [command]

Available Commands:
  account     Display account information
  completion  Generate the autocompletion script for the specified shell
  cp          Copy a file or folder to a different location in the user's Dropbox. If the source path is a folder all its contents will be copied.
  du          Display usage information
  get         Download a file or folder
  help        Help about any command
  login       Log in and save Dropbox credentials
  logout      Log out of the current session
  ls          List files and folders
  mkdir       Create a new directory
  mv          Move files
  put         Upload files or directories
  restore     Restore a file revision
  revs        List file revisions
  rm          Remove files or folders
  search      Search
  share       Sharing commands
  share-link  Shared link commands
  team        Team management commands
  version     Print version information

Flags:
      --as-member string   Member ID to perform action as
  -h, --help               help for dbxcli
      --output string      Output format: text, json (default "text")
  -v, --verbose            Enable verbose logging

Use "dbxcli [command] --help" for more information about a command.
```

The complete generated command reference is available in [docs/commands/dbxcli.md](https://github.com/dropbox/dbxcli/blob/master/docs/commands/dbxcli.md).

### Output formats

Text output is the default. JSON output is available through the global `--output` flag for migrated commands:

```sh
$ dbxcli ls --output=json /
```

Command results are written to stdout. Status, progress, human-facing warnings, diagnostics, and verbose logs are written to stderr.

Successful JSON responses return `ok: true`, `schema_version: "1"`, `command`, an `input` object, a `results` array, and a `warnings` array. Each result includes `status`, `kind`, `input`, and `result`. Machine-actionable JSON warnings use the `warnings` array.

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

In JSON mode, error responses are written to stdout and the process exits with a non-zero status. JSON errors return `ok: false`, a stable `error.code`, a human-readable `error.message`, optional `error.details`, and `warnings`:

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

The full JSON command catalog, stable error codes, and schemas live in [docs/json-schema/v1](https://github.com/dropbox/dbxcli/tree/master/docs/json-schema/v1). Commands that intentionally do not support JSON output yet include `login`, `logout`, and `completion`. Help output and shell-completion protocol commands are text-only.

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

A bare `-` means stdin/stdout only when it appears as the local operand. Dropbox paths named `-` are valid, for example `dbxcli put - /-` and `dbxcli get /- -`. To upload a local file literally named `-`, use `./-`.

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

 * If you're submitting a non-trivial change, please fill out the [Dropbox Contributor License Agreement](https://opensource.dropbox.com/cla/) first.
 * Open a [pull request](https://help.github.com/articles/using-pull-requests/) with a clear description of the change.
 * Include tests or manual validation details when relevant.
 
## Useful Resources

* [Go SDK documentation](https://pkg.go.dev/github.com/dropbox/dropbox-sdk-go-unofficial)
* [API documentation](https://www.dropbox.com/developers/documentation/http/documentation)
