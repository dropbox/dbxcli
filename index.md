# `dbxcli`: A command line tool for Dropbox users and team admins

This is *BETA* software -- use with caution!

## Features

  * Supports basic file operations like ls, cp, mkdir, mv (via the Files API)
  * Supports search
  * Supports file revisions and file restore
  * Chunked uploads for large files, paginated listing for large directories
  * Supports a growing set of Team operations

## Installation

Download pre-compiled binaries for Mac, Windows and Linux from the [releases](https://github.com/dropbox/dbxcli/releases) page.

## Usage

`dbxcli` is largely self documenting. Run `dbxcli -h` for a list of supported commands:

```sh
$ dbxcli --help
Use dbxcli to quickly interact with your Dropbox, upload/download files,
manage your team and more. It is easy, scriptable and works on all platforms!

Usage:
  dbxcli [command]

Available Commands:
  cp          Copy files
  du          Display usage information
  get         Download a file
  ls          List files
  mkdir       Create a new directory
  mv          Move files
  put         Upload files
  restore     Restore files
  revs        List file revisions
  rm          Remove files
  search      Search
  team        Team management commands
  version     Print version information

Flags:
      --as-member string   Member ID to perform action as
  -v, --verbose            Enable verbose logging

Use "dbxcli [command] --help" for more information about a command.

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
  -v, --verbose            Enable verbose logging

Use "dbxcli team [command] --help" for more information about a command.
```

The `--verbose` option will turn on verbose logging and is useful for debugging.

## We need your help!

`dbxcli` is under active development! As you can see from the [API docs](https://www.dropbox.com/developers/documentation/http/documentation), we only support a small number of features today and have only scratched the surface of what's possible. We would love feedback from you, our users, to guide what to build next and how to improve the tool.

So please, file feature requests, report bugs or better yet, send us pull requests! More on contributing below.

## Contributing

 * Step 1: If you're submitting a non-trivial change, please fill out the [Dropbox Contributor License Agreement](https://opensource.dropbox.com/cla/) first.
 * Step 2: send a [pull request](https://help.github.com/articles/using-pull-requests/)
 * Step 3: Profit!
 
## Useful Resources

* [Go SDK documentation](https://godoc.org/github.com/dropbox/dropbox-sdk-go-unofficial)
* [API documentation](https://www.dropbox.com/developers/documentation/http/documentation)
