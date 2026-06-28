# `dbxcli`: Dropbox from the command line

[![CI](https://github.com/dropbox/dbxcli/actions/workflows/ci.yml/badge.svg)](https://github.com/dropbox/dbxcli/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dropbox/dbxcli/v3)](https://goreportcard.com/report/github.com/dropbox/dbxcli/v3)

`dbxcli` is a scriptable Dropbox CLI for files, shared links, teams, and
automation workflows. It is built for humans at the terminal, scripts, CI jobs,
and agent-style workflows.

## Why use dbxcli?

* Work with Dropbox from the terminal
* Upload and download files, folders, and streams
* Create, inspect, update, revoke, and download shared links
* Use JSON output for scripts and automation
* Pipe generated content directly into Dropbox
* Manage team workflows with member-scoped access

## Quickstart

```sh
dbxcli login
dbxcli ls /
dbxcli put local.txt /remote.txt
dbxcli get /remote.txt ./remote.txt
dbxcli share-link create /remote.txt
```

For automation, use structured command output and JSON help discovery:

```sh
dbxcli ls --output=json /
dbxcli --help --output=json
dbxcli put --help --output=json
```

## Common workflows

Upload a file:

```sh
dbxcli put report.pdf /Reports/report.pdf
```

Upload without overwriting:

```sh
dbxcli put --if-exists fail report.md /Reports/report.md
```

Upload from a pipe:

```sh
tar cz ./project | dbxcli put - /Backups/project.tgz
```

Download to stdout:

```sh
dbxcli get /Backups/project.tgz - | tar tz
```

Create a shared link:

```sh
dbxcli share-link create /Reports/report.pdf
```

Use JSON output:

```sh
dbxcli ls --output=json /
```

## Features

* File operations: `ls`, `cp`, `mkdir`, `mv`, `rm`, `put`, and `get`
* Recursive upload and download with `put -r` and `get -r`
* Pipe-friendly transfers with stdin upload and stdout download
* Upload conflict control with `put --if-exists overwrite|skip|fail`
* Shared-link creation, listing, inspection, update, revoke, and download
* Search, file revisions, restore, flexible sorting, and time formatting
* Chunked uploads for large files and paginated listing for large directories
* OAuth login with refreshable saved credentials
* Direct token automation with `DBXCLI_ACCESS_TOKEN`
* Alternate saved-credential files with `DBXCLI_AUTH_FILE`
* Structured JSON success and error envelopes for migrated commands
* JSON help manifests for machine-readable command discovery
* Team administration commands and member-scoped access with `--as-member`

## Installation

### Homebrew

```sh
brew install dbxcli
```

### Release archives

Download the archive for your platform from the
[releases](https://github.com/dropbox/dbxcli/releases) page, verify its
checksum, and install the `dbxcli` binary somewhere on your `PATH`.

Linux example:

```sh
curl -LO https://github.com/dropbox/dbxcli/releases/download/vX.Y.Z/dbxcli_X.Y.Z_linux_amd64.tar.gz
curl -LO https://github.com/dropbox/dbxcli/releases/download/vX.Y.Z/SHA256SUMS
grep 'dbxcli_X.Y.Z_linux_amd64.tar.gz' SHA256SUMS | sha256sum -c -
tar -xzf dbxcli_X.Y.Z_linux_amd64.tar.gz
sudo mv dbxcli_X.Y.Z_linux_amd64/dbxcli /usr/local/bin/
```

Release assets include:

* `dbxcli_X.Y.Z_darwin_amd64.tar.gz`
* `dbxcli_X.Y.Z_darwin_arm64.tar.gz`
* `dbxcli_X.Y.Z_linux_amd64.tar.gz`
* `dbxcli_X.Y.Z_linux_arm64.tar.gz`
* `dbxcli_X.Y.Z_linux_arm.tar.gz`
* `dbxcli_X.Y.Z_openbsd_amd64.tar.gz`
* `dbxcli_X.Y.Z_windows_amd64.zip`
* `SHA256SUMS`

### Build from source

```sh
go install github.com/dropbox/dbxcli/v3@latest
```

Or build from a clone:

```sh
git clone https://github.com/dropbox/dbxcli.git
cd dbxcli
go build .
```

## Support posture

`dbxcli` is maintained in the Dropbox GitHub organization by Dropbox engineers,
but it is not a formally supported Dropbox product. Use GitHub issues and pull
requests for bugs and contributions; Dropbox Support does not provide support
for this CLI. The CLI implements a practical subset of Dropbox API features,
not the full API surface.

## Command reference

The complete generated command reference is available here:

* [dbxcli command reference](https://github.com/dropbox/dbxcli/blob/master/docs/commands/dbxcli.md)

For command-specific help, run:

```sh
dbxcli --help
dbxcli put --help
dbxcli share-link --help
dbxcli share-link create --help
```

For machine-readable command discovery, use JSON help:

```sh
dbxcli --help --output=json
dbxcli put --help --output=json
```

## Deeper documentation

* [Automation and JSON output](https://github.com/dropbox/dbxcli/blob/master/docs/automation.md)
* [Sharing workflows](https://github.com/dropbox/dbxcli/blob/master/docs/sharing.md)
* [JSON schema v1](https://github.com/dropbox/dbxcli/blob/master/docs/json-schema/v1/README.md)
* [Release history](https://github.com/dropbox/dbxcli/blob/master/CHANGELOG.md)

Generated Cobra command docs live under `docs/commands/` and are kept close to
the actual CLI by `go run ./tools/gen-docs`.

## Contributing

* If you are submitting a non-trivial change, please fill out the
  [Dropbox Contributor License Agreement](https://opensource.dropbox.com/cla/)
  first.
* Open a [pull request](https://help.github.com/articles/using-pull-requests/)
  with a clear description of the change.
* Include tests or manual validation details when relevant.

## Useful resources

* [Go SDK documentation](https://pkg.go.dev/github.com/dropbox/dropbox-sdk-go-unofficial)
* [Dropbox API documentation](https://www.dropbox.com/developers/documentation/http/documentation)
