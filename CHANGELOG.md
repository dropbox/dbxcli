# Change Log

## [Unreleased](https://github.com/dropbox/dbxcli/tree/HEAD)

[Full Changelog](https://github.com/dropbox/dbxcli/compare/v3.6.0...HEAD)

**Infrastructure:**

- Added scheduled/manual OSSF Scorecard scanning without public Scorecard API publishing.
- Added scheduled/manual non-blocking Markdown link checks for README and docs.
- Replaced standalone Staticcheck workflow steps with a narrow `golangci-lint` configuration.
- Added explicit `go.sum` cache dependency paths for `actions/setup-go`.

## [v3.6.0](https://github.com/dropbox/dbxcli/tree/v3.6.0) (2026-07-02)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v3.5.1...v3.6.0)

**Added:**

- Added `--if-exists fail|skip` to `cp` and `mv` commands.
- Added `--limit` flag to `ls` and `revs` commands.
- Added `--recursive` flag to `ls`.
- Added JSON command manifests with `input_schema` for machine-readable command discovery.
- Added `commands.schema.json` for per-command success and input validation.
- Added richer JSON `error.details` fields for operation context, paths, URLs, revisions, email addresses, member IDs, Dropbox API summaries, and retry-after values.
- Added typed error details schema and deprecated-command warnings in JSON errors.
- Added Dependabot configuration for Go modules and GitHub Actions dependencies.
- Added WinGet manifest templates and render script for Windows package manager submissions.
- Added `govulncheck` and `staticcheck` to CI and release workflows.
- Added version resolution from `debug.ReadBuildInfo` for `go install` users.

**Changed:**

- Upgraded Dropbox SDK from v6.0.5 to v6.2.0 (migrates `time.Time` fields to `dropbox.DBXTime`).
- Bumped `actions/checkout` from v6 to v7.
- Updated Homebrew formula to v3.5.1 with doc installation and test assertions.
- Updated root command help and generated command docs to reflect current workflows.
- Replaced subset schema validator with `santhosh-tekuri/jsonschema/v6`.

**Fixed:**

- Fixed `ls --only-deleted --limit` to count filtered entries correctly.
- Fixed trailing period in `logout` error message (staticcheck finding).

**Infrastructure:**

- Added golden JSON error output fixtures and live JSON error envelope tests.
- Hardened JSON error detail collection through wrapped and joined errors.
- Reduced false-positive Dropbox `api_summary` detection for local path errors.

## [v3.5.1](https://github.com/dropbox/dbxcli/tree/v3.5.1) (2026-06-28)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v3.5.0...v3.5.1)

**Added:**

- Added structured `logout --output=json` output with saved-credential removal, token-revoke status, and already-logged-out reporting.
- Added `result.auth` to `account --output=json` exposing credential source, refreshability, and auth file type.
- Added `--content`, `--limit`, and `--order-by` flags to `search` command.
- Added stable exit codes (0-8) mapped from JSON error codes for shell/CI scripting.
- Added `partial_transfer` error code for stdout download failures after partial output.

**Changed:**

- Bumped Go module path to `github.com/dropbox/dbxcli/v3`.
- Restructured README into focused topic docs under `docs/`.

**Fixed:**

- Fixed interface pointer comparison in share-link error parsers using reflection-based `samePointer` helper.
- Migrated error assertions to `errors.Is` for correct wrapped-error handling.
- Fixed reflection-based `ErrorSummary` extraction to avoid `CanInterface` pitfalls.

## [v3.5.0](https://github.com/dropbox/dbxcli/tree/v3.5.0) (2026-06-26)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v3.4.0...v3.5.0)

**Added:**

- Structured `--output=json` support across core file, share, account, team, usage, and version commands.
- Stable JSON success/error envelopes with schema v1, warning objects, stable error codes, and structured error details.
- JSON help manifests via `--help --output=json` for machine-readable command discovery.
- Root namespace auto-detection for team-folder access.
- Generated command reference docs with CI drift detection.
- `put` chunk-size validation: 4-128 MiB, multiple of 4 MiB.

**Changed:**

- Normalized JSON result shapes around `input`, `results`, and `warnings`.
- Refreshed README and moved detailed JSON schema docs to `docs/json-schema/v1/`.
- Improved `put --chunksize` and `--workers` help text.

**Infrastructure:**

- Added JSON contract tests with golden schema/output fixtures.

## [v3.4.0](https://github.com/dropbox/dbxcli/tree/v3.4.0) (2026-06-22)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v3.3.3...v3.4.0)

**Added:**

- Added the `share-link` command family for creating, listing, inspecting, updating, revoking, and downloading shared links.
- Added shared-link settings support for access, audience, expiration, password, allow-download, and disallow-download options.
- Added `--path` support for shared-link info/download/revoke workflows.
- Added recursive shared-link folder downloads.
- Added Unix pipe support with `put -` for stdin uploads and `get ... -` for stdout downloads.
- Added `put --if-exists overwrite|skip|fail` for explicit upload conflict behavior.

**Changed:**

- Migrated `search` to Dropbox SearchV2 with pagination support.
- Improved `revs` and `restore` help and time formatting.
- Updated README installation instructions for Homebrew and release archives.

## [v3.3.3](https://github.com/dropbox/dbxcli/tree/v3.3.3) (2026-06-17)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v3.3.2...v3.3.3)

**Changed:**

- Local commands such as help, version, and completion no longer require saved Dropbox credentials.

## [v3.3.2](https://github.com/dropbox/dbxcli/tree/v3.3.2) (2026-06-17)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v3.3.1...v3.3.2)

**Added:**

- Added GitHub Actions CI, release, and Pages workflows.
- Added versioned release archives, SHA256SUMS, release packaging, and multi-OS CI validation.
- Added bundled Dropbox team app keys.

**Fixed:**

- Fixed `ls` error handling.

## [v3.3.1](https://github.com/dropbox/dbxcli/tree/v3.3.1) (2026-06-15)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v3.3.0...v3.3.1)

**Added:**

- Added `dbxcli login` with OAuth authorization-code flow.
- Added PKCE authentication with offline refresh tokens and automatic access-token refresh.
- Added `DBXCLI_ACCESS_TOKEN` for short-lived direct token use.
- Added `DBXCLI_AUTH_FILE` for selecting an alternate credentials file.
- Added `rm --recursive`/`-r` and `rm --permanent`.

**Changed:**

- Saved OAuth credentials now use a refresh-token aware `auth.json` object format. Existing legacy token-string entries are still read, but any credential write rewrites the file in the new format.
- `rm --force` remains supported as an alias for recursive non-empty folder deletion.
- `rm --verbose` reports deleted paths.

## [v3.3.0](https://github.com/dropbox/dbxcli/tree/v3.3.0) (2026-06-12)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v3.2.1...v3.3.0)

**Added:**

- Added recursive folder downloads with `get -r`.
- Added recursive directory uploads with `put -r`.
- Added `mkdir -p`.
- Added `--sort`, `--reverse`, `--time`, and `--time-format` flags for `ls` and `search`.

**Changed:**

- Unified `cp` and `mv` destination handling for multiple sources, trailing slashes, and existing remote folders.
- Improved `cp` and `mv` error messages to show quoted paths and Dropbox API error text.
- Fixed local path handling to use platform-native filesystem paths where appropriate.

## [v3.2.1](https://github.com/dropbox/dbxcli/tree/v3.2.1) (2026-06-09)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v3.2.0...v3.2.1)

**Fixed:**

- Commands now return a non-zero exit code on errors.
- `mv` and `cp` now propagate errors correctly.
- Search output now prints one result per line and aligns long output with tabwriter columns.

## [v3.2.0](https://github.com/dropbox/dbxcli/tree/v3.2.0) (2026-06-08)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v3.1.0...v3.2.0)

**Added:**

- Added retry with exponential backoff for transient upload and download failures.
- Added atomic downloads through temp-file writes followed by rename.
- Added idempotent chunked upload recovery for accepted chunks.

**Fixed:**

- Retry download failures from `unexpected EOF`.
- Retry upload failures caused by transient server, rate-limit, network, and `too_many_write_operations` errors.
- Preserve symlinks on download.

## [v3.1.0](https://github.com/dropbox/dbxcli/tree/v3.1.0) (2026-06-08)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v3.0.0...v3.1.0)

**Changed:**

- Upgraded Go from 1.11 to 1.25.
- Updated dependencies, including Cobra, Dropbox SDK, OAuth2, and pflag.
- Replaced deprecated Go packages with standard-library equivalents.

**Fixed:**

- Fixed `ls /` root listing with the newer Dropbox SDK.
- Fixed `put` upload argument construction for Dropbox SDK v6.0.5.
- Added unit coverage for `ls` path validation and formatting helpers.

## [v3.0.0](https://github.com/dropbox/dbxcli/tree/v3.0.0) (2019-01-30)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v2.1.2...v3.0.0)

**Changed:**

- Updated dependencies and the underlying Dropbox SDK.
- Bumped the major version because of SDK-level changes.

## [v2.1.2](https://github.com/dropbox/dbxcli/tree/v2.1.2) (2018-12-05)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v2.1.1...v2.1.2)

**Implemented enhancements:**

- Provide credentials through environment variables [\#104](https://github.com/dropbox/dbxcli/issues/104)

**Fixed:**

- Fixed moving files between subfolders [\#105](https://github.com/dropbox/dbxcli/pull/105)

**Closed issues:**

- Move error when moving between subfolders [\#102](https://github.com/dropbox/dbxcli/issues/102)
- Using a SOCKS proxy? [\#97](https://github.com/dropbox/dbxcli/issues/97)
- dbxcli doesn't detect when OAuth2 token no longer works [\#94](https://github.com/dropbox/dbxcli/issues/94)
- Can't get the authorization code [\#92](https://github.com/dropbox/dbxcli/issues/92)
- Specify auth token as an argument [\#63](https://github.com/dropbox/dbxcli/issues/63)

## [v2.1.1](https://github.com/dropbox/dbxcli/tree/v2.1.1) (2018-01-03)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v2.1.0...v2.1.1)

**Fixed:**

- Fixed a segfault in `dbxcli account`.

## [v2.1.0](https://github.com/dropbox/dbxcli/tree/v2.1.0) (2017-12-13)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v2.0.9...v2.1.0)

**Fixed:**

- Intake fix for a Dropbox SDK issue: https://github.com/dropbox/dropbox-sdk-go-unofficial/issues/38

## [v2.0.9](https://github.com/dropbox/dbxcli/tree/v2.0.9) (2017-12-01)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v2.0.8...v2.0.9)

**Added:**

- Added OpenBSD binaries.

**Closed issues:**

- Invalid client\_id when trying to get authorization code [\#79](https://github.com/dropbox/dbxcli/issues/79)
- Build official binaries for more OS [\#76](https://github.com/dropbox/dbxcli/issues/76)

## [v2.0.8](https://github.com/dropbox/dbxcli/tree/v2.0.8) (2017-11-10)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v2.0.7...v2.0.8)

## [v2.0.7](https://github.com/dropbox/dbxcli/tree/v2.0.7) (2017-11-10)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v2.0.6...v2.0.7)

**Closed issues:**

- Better output for ls [\#78](https://github.com/dropbox/dbxcli/issues/78)

## [v2.0.6](https://github.com/dropbox/dbxcli/tree/v2.0.6) (2017-07-26)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v2.0.5...v2.0.6)

**Implemented enhancements:**

- Add `--recurse` option to `ls` [\#74](https://github.com/dropbox/dbxcli/issues/74)

**Merged pull requests:**

- Add `account` command [\#15](https://github.com/dropbox/dbxcli/pull/15) ([waits](https://github.com/waits))

## [v2.0.5](https://github.com/dropbox/dbxcli/tree/v2.0.5) (2017-07-26)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v2.0.4...v2.0.5)

## [v2.0.4](https://github.com/dropbox/dbxcli/tree/v2.0.4) (2017-07-26)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v2.0.3...v2.0.4)

**Closed issues:**

- Usage question with wildcards [\#72](https://github.com/dropbox/dbxcli/issues/72)
- Could not able to copy a tar.gz file in my dropbox folder [\#69](https://github.com/dropbox/dbxcli/issues/69)
- v2.0.2 in Windows ls not returning results unless verbose [\#68](https://github.com/dropbox/dbxcli/issues/68)
- Rename a file? [\#66](https://github.com/dropbox/dbxcli/issues/66)

## [v2.0.3](https://github.com/dropbox/dbxcli/tree/v2.0.3) (2017-07-24)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v2.0.2...v2.0.3)

**Closed issues:**

- How do I install this program? [\#67](https://github.com/dropbox/dbxcli/issues/67)
- Unable to Authorize App [\#64](https://github.com/dropbox/dbxcli/issues/64)

**Merged pull requests:**

- Switch to `dep` and update dependencies [\#73](https://github.com/dropbox/dbxcli/pull/73) ([diwakergupta](https://github.com/diwakergupta))

## [v2.0.2](https://github.com/dropbox/dbxcli/tree/v2.0.2) (2017-02-27)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v2.0.1...v2.0.2)

**Fixed bugs:**

- `get` does not work on JPG files [\#59](https://github.com/dropbox/dbxcli/issues/59)

**Closed issues:**

- Can't authorize team management  [\#58](https://github.com/dropbox/dbxcli/issues/58)

**Merged pull requests:**

- add build support for raspberry pi \(ARM\) [\#65](https://github.com/dropbox/dbxcli/pull/65) ([garyemerson](https://github.com/garyemerson))

## [v2.0.1](https://github.com/dropbox/dbxcli/tree/v2.0.1) (2017-02-14)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v2.0.0...v2.0.1)

**Closed issues:**

- Invalid client\_id when trying to get authorization code [\#62](https://github.com/dropbox/dbxcli/issues/62)
- Generating authorization code no longer works [\#61](https://github.com/dropbox/dbxcli/issues/61)

## [v2.0.0](https://github.com/dropbox/dbxcli/tree/v2.0.0) (2017-01-26)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v1.4.0...v2.0.0)

**Closed issues:**

- build instructions in readme [\#57](https://github.com/dropbox/dbxcli/issues/57)
- Authorization error - app has reached its team limit [\#47](https://github.com/dropbox/dbxcli/issues/47)
- `search` should show a full path to matching files [\#42](https://github.com/dropbox/dbxcli/issues/42)
- Recursive and force flags for rm [\#26](https://github.com/dropbox/dbxcli/issues/26)

**Merged pull requests:**

- Display full path [\#55](https://github.com/dropbox/dbxcli/pull/55) ([hut8](https://github.com/hut8))
- Add multiple args to rm [\#49](https://github.com/dropbox/dbxcli/pull/49) ([GrantSeltzer](https://github.com/GrantSeltzer))
- Update Golumns package [\#48](https://github.com/dropbox/dbxcli/pull/48) ([GrantSeltzer](https://github.com/GrantSeltzer))
- Add force flag for `rm`ing non-empty directories, remove `rmdir` [\#43](https://github.com/dropbox/dbxcli/pull/43) ([GrantSeltzer](https://github.com/GrantSeltzer))

## [v1.4.0](https://github.com/dropbox/dbxcli/tree/v1.4.0) (2016-08-01)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v1.3.0...v1.4.0)

**Merged pull requests:**

- Update golumns package to latest version - major bug fix [\#44](https://github.com/dropbox/dbxcli/pull/44) ([GrantSeltzer](https://github.com/GrantSeltzer))
- Adds another subcommand layer `share list`. [\#39](https://github.com/dropbox/dbxcli/pull/39) ([bonafidehan](https://github.com/bonafidehan))
- Adds `share list-links`. Paging for `share list-folders`. [\#38](https://github.com/dropbox/dbxcli/pull/38) ([bonafidehan](https://github.com/bonafidehan))
- Introduces `share` command and `list-folders` subcommand. [\#37](https://github.com/dropbox/dbxcli/pull/37) ([bonafidehan](https://github.com/bonafidehan))
- Introduces scoped search. A search can be scoped to the provided folder. [\#36](https://github.com/dropbox/dbxcli/pull/36) ([bonafidehan](https://github.com/bonafidehan))
- Replace strings with consts defined in root.go [\#33](https://github.com/dropbox/dbxcli/pull/33) ([GrantSeltzer](https://github.com/GrantSeltzer))
- Allow for multiple arguments to cp [\#32](https://github.com/dropbox/dbxcli/pull/32) ([GrantSeltzer](https://github.com/GrantSeltzer))
- Add `logout` command [\#23](https://github.com/dropbox/dbxcli/pull/23) ([waits](https://github.com/waits))

## [v1.3.0](https://github.com/dropbox/dbxcli/tree/v1.3.0) (2016-07-17)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v1.2.0...v1.3.0)

**Closed issues:**

- Have seperate commands for `rm` and `rmdir` [\#25](https://github.com/dropbox/dbxcli/issues/25)
- `put` command is sending wrong client\_modified timestamp [\#20](https://github.com/dropbox/dbxcli/issues/20)
- Make `ls` list files in multiple columns [\#17](https://github.com/dropbox/dbxcli/issues/17)
- Add `logout` or `revoke` command [\#16](https://github.com/dropbox/dbxcli/issues/16)

**Merged pull requests:**

- Allow for multiple arguments to mv [\#30](https://github.com/dropbox/dbxcli/pull/30) ([GrantSeltzer](https://github.com/GrantSeltzer))
- Split rm into rm/rmdir, added consts for dangling strings [\#28](https://github.com/dropbox/dbxcli/pull/28) ([GrantSeltzer](https://github.com/GrantSeltzer))
- Allow providing a directory as a destination for `get` [\#22](https://github.com/dropbox/dbxcli/pull/22) ([waits](https://github.com/waits))
- Set `client\_modified` parameter when uploading files [\#21](https://github.com/dropbox/dbxcli/pull/21) ([waits](https://github.com/waits))
- Display file sizes using multiples of 1024 for consistency with other Dropbox apps [\#19](https://github.com/dropbox/dbxcli/pull/19) ([waits](https://github.com/waits))

## [v1.2.0](https://github.com/dropbox/dbxcli/tree/v1.2.0) (2016-06-07)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v1.1.0...v1.2.0)

**Implemented enhancements:**

- Support `ls` on files [\#8](https://github.com/dropbox/dbxcli/issues/8)

**Closed issues:**

- "Usage" section of help text is missing arguments [\#13](https://github.com/dropbox/dbxcli/issues/13)
- `get` command panics without second argument [\#10](https://github.com/dropbox/dbxcli/issues/10)

**Merged pull requests:**

- Check `args` slice bounds in all commands [\#18](https://github.com/dropbox/dbxcli/pull/18) ([waits](https://github.com/waits))
- Add argument information to "usage" section of help text [\#14](https://github.com/dropbox/dbxcli/pull/14) ([waits](https://github.com/waits))
- Check `args` slice bounds in `get` and `put` functions [\#12](https://github.com/dropbox/dbxcli/pull/12) ([waits](https://github.com/waits))
- Support `ls` on files [\#11](https://github.com/dropbox/dbxcli/pull/11) ([waits](https://github.com/waits))

## [v1.1.0](https://github.com/dropbox/dbxcli/tree/v1.1.0) (2016-05-05)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v1.0.0...v1.1.0)

**Closed issues:**

- Bad authorization URL generated. [\#9](https://github.com/dropbox/dbxcli/issues/9)
- Fails on most uploads and downloads [\#7](https://github.com/dropbox/dbxcli/issues/7)

## [v1.0.0](https://github.com/dropbox/dbxcli/tree/v1.0.0) (2016-03-23)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v0.6.0...v1.0.0)

## [v0.6.0](https://github.com/dropbox/dbxcli/tree/v0.6.0) (2016-03-19)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v0.5.0...v0.6.0)

## [v0.5.0](https://github.com/dropbox/dbxcli/tree/v0.5.0) (2016-03-16)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v0.4.0...v0.5.0)

**Closed issues:**

- Improve \(or add?\) error handling [\#1](https://github.com/dropbox/dbxcli/issues/1)

## [v0.4.0](https://github.com/dropbox/dbxcli/tree/v0.4.0) (2016-03-15)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v0.3.0...v0.4.0)

## [v0.3.0](https://github.com/dropbox/dbxcli/tree/v0.3.0) (2016-03-15)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v0.2.0...v0.3.0)

## [v0.2.0](https://github.com/dropbox/dbxcli/tree/v0.2.0) (2016-03-14)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v0.1.1...v0.2.0)

**Closed issues:**

- Asks for authentication code on each run \[Linux\] [\#6](https://github.com/dropbox/dbxcli/issues/6)

**Merged pull requests:**

- Add zsh-completion [\#5](https://github.com/dropbox/dbxcli/pull/5) ([knakayama](https://github.com/knakayama))
- Run `go vet` in `before\_script` [\#4](https://github.com/dropbox/dbxcli/pull/4) ([diwakergupta](https://github.com/diwakergupta))
- Create directory [\#3](https://github.com/dropbox/dbxcli/pull/3) ([mattn](https://github.com/mattn))

## [v0.1.1](https://github.com/dropbox/dbxcli/tree/v0.1.1) (2016-03-11)
[Full Changelog](https://github.com/dropbox/dbxcli/compare/v0.1.0...v0.1.1)

**Merged pull requests:**

- Prepare to push releases through Travis [\#2](https://github.com/dropbox/dbxcli/pull/2) ([diwakergupta](https://github.com/diwakergupta))

## [v0.1.0](https://github.com/dropbox/dbxcli/tree/v0.1.0) (2016-03-10)


\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*
