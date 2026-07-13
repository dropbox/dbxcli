# `dbxcli`: Dropbox from the command line

[![CI](https://github.com/dropbox/dbxcli/actions/workflows/ci.yml/badge.svg)](https://github.com/dropbox/dbxcli/actions/workflows/ci.yml)
[![Scorecard](https://github.com/dropbox/dbxcli/actions/workflows/scorecard.yml/badge.svg)](https://github.com/dropbox/dbxcli/actions/workflows/scorecard.yml)
[![CodeQL](https://github.com/dropbox/dbxcli/actions/workflows/codeql.yml/badge.svg)](https://github.com/dropbox/dbxcli/actions/workflows/codeql.yml)

`dbxcli` is a scriptable Dropbox CLI for files, shared links, teams, and
automation workflows. It is built for humans in the terminal, scripts, CI jobs,
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

On team accounts where `/` is not writable, run `dbxcli ls /` and use a
writable personal or team folder instead.

For automation, use structured command output and JSON help discovery:

```sh
dbxcli ls --output=json /
dbxcli --help --output=json
dbxcli put --help --output=json
```

Stable JSON envelopes, error codes, and process exit codes are documented in
[Automation and JSON output](https://github.com/dropbox/dbxcli/blob/master/docs/automation.md).

## JSON output

For commands that support structured execution output, `--output=json` runs the
command and emits stable schema v1 success and error envelopes for automation.

JSON help is the machine-readable command-discovery surface. Use it to discover
command paths, arguments, flags, aliases, input schemas, auth behavior,
stdin/stdout behavior, schema references, and whether structured command
execution output is supported:

```sh
dbxcli --help --output=json
dbxcli put --help --output=json
```

See the
[JSON schema v1 docs](https://github.com/dropbox/dbxcli/blob/master/docs/json-schema/v1/README.md)
for schemas, stability policy, command contracts, and examples.

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

In text mode, `share-link create` prints only the shared-link URL to stdout:

```sh
url="$(dbxcli share-link create /Reports/report.pdf)"
```

## Troubleshooting

### Why can uploading to `/remote.txt` fail on team accounts?

Some team accounts may not have a writable Dropbox root namespace. Run
`dbxcli ls /` first, then upload under a writable folder, such as your personal
folder or a team folder.

### Proxy configuration

`dbxcli` uses Go's standard HTTP proxy behavior, so `HTTPS_PROXY`,
`HTTP_PROXY`, and `NO_PROXY` apply to Dropbox API requests and OAuth token
exchange/refresh requests made by the CLI.

For Dropbox API and OAuth requests, set `HTTPS_PROXY`:

```sh
HTTPS_PROXY=http://127.0.0.1:8080 dbxcli ls /
```

For a shell session:

```sh
export HTTPS_PROXY=http://proxy.company.example:8080
export NO_PROXY=localhost,127.0.0.1,.company.example
dbxcli login
```

On Windows PowerShell:

```powershell
$env:HTTPS_PROXY = "http://127.0.0.1:8080"
dbxcli login
```

On Windows cmd:

```bat
set HTTPS_PROXY=http://127.0.0.1:8080
dbxcli login
```

`HTTP_PROXY` is also honored for plain HTTP requests. Use `NO_PROXY` to bypass
the proxy for local or internal hosts. Lowercase forms such as `https_proxy`
and `no_proxy` are also supported by Go's HTTP stack.

If your proxy requires basic authentication, include credentials in the proxy
URL:

```sh
HTTPS_PROXY=http://user:password@proxy.company.example:8080 dbxcli ls /
```

URL-encode special characters in proxy usernames or passwords. Be careful with
proxy credentials in environment variables, shell history, CI logs, and process
listings.

The browser authorization step in `dbxcli login` is outside `dbxcli`; configure
your browser or operating-system proxy separately if that page also needs a
proxy.

## Features

* File operations: `ls`, `cp`, `mkdir`, `mv`, `rm`, `put`, and `get`
* Recursive upload and download with `put -r` and `get -r`
* Pipe-friendly transfers with stdin upload and stdout download
* Conflict control with `put --if-exists overwrite|skip|autorename|fail` and `cp`/`mv --if-exists fail|skip|autorename`
* Shared-link creation, listing, inspection, update, revoke, and download
* Search, file revisions, restore, flexible sorting, and time formatting
* Chunked uploads for large files and paginated listing for large directories
* OAuth login with refreshable saved credentials
* Direct token automation with `DBXCLI_ACCESS_TOKEN`
* Alternate saved-credential files with `DBXCLI_AUTH_FILE`
* Structured JSON success and error envelopes for supported commands
* JSON help manifests for machine-readable command discovery
* Team administration commands and member-scoped access with `--as-member`

## Installation

### Homebrew

```sh
brew install dbxcli
```

Homebrew formula: [formulae.brew.sh/formula/dbxcli](https://formulae.brew.sh/formula/dbxcli)

### WinGet

```powershell
winget install --exact --id Dropbox.dbxcli
```

WinGet manifest: [microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs/tree/master/manifests/d/Dropbox/dbxcli)

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

Generated Cobra command docs live under `docs/commands/`, and CI verifies they
stay in sync with the CLI.

## Contributing

* If you are submitting a non-trivial change, please fill out the
  [Dropbox Contributor License Agreement](https://opensource.dropbox.com/cla/)
  first.
* Open a [pull request](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request)
  with a clear description of the change.
* Include tests or manual validation details when relevant.

## Useful resources

* [Go SDK documentation](https://pkg.go.dev/github.com/dropbox/dropbox-sdk-go-unofficial)
* [Dropbox API documentation](https://www.dropbox.com/developers/documentation/http/documentation)
