# Change Log

## [Unreleased](https://github.com/dropbox/dbxcli/tree/HEAD)

[Full Changelog](https://github.com/dropbox/dbxcli/compare/v2.0.8...HEAD)

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