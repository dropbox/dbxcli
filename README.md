# `dbxcli`: A command line tool for Dropbox users and team admins

[![Go Report Card](https://goreportcard.com/badge/github.com/dropbox/dbxcli)](https://goreportcard.com/report/github.com/dropbox/dbxcli)

## Features

  * Supports basic file operations like ls, cp, mkdir, mv (via the Files API)
  * Supports search
  * Supports file revisions and file restore
  * Chunked uploads for large files, paginated listing for large directories
  * Supports a growing set of Team operations

## Installation

Download pre-compiled binaries for Mac, Windows and Linux from the [releases](https://github.com/dropbox/dbxcli/releases) page.

### Linux Installation of pre-compiled binaries
1. Check whether your CPU is compatible with AMD/Intel or ARM
```sh
uname -p
```
If the result reads `x86_64` use the `AMD64` binary.  We will continue assuming you have x86_64 architecture (most common).

2. Download the `dbxcli` binary into your binary folder and rename it to `dbxcli`
```sh
$ cd /bin
$ sudo wget https://github.com/dropbox/dbxcli/releases/download/v2.0.9/dbxcli-linux-amd64
$ sudo mv dbxcli-linux-amd64 dbxcli 
```
3. Finally, make the binary an executable file and you are good to go!
```sh
$ chmod +x dbxcli
```

### Mac OSX Installation of pre-compiled binaries
These instructions aim to help both experts and novice `dbxcli` users. Please submit an issue if they don't work for you.  

1. Make sure you download and place the binary in a folder that's on your `$PATH`.  If you are unsure what this means, go to *step 2*. Otherwise, skip to *step 3*
2. Create a `bin` directory under your home directory.
```
$ mkdir ~/bin
$ cd ~/bin
```
3. Add the following line at the end of your `~/.bash_profile` file.  [Link with instructions](https://natelandau.com/my-mac-osx-bash_profile/) on how to find this file
```sh
export PATH=$PATH:$HOME/bin
```
4. Download the `dbxcli` binary for OSX and rename it.  *IMPORTANT:* Check that the tag `v2.0.9` on the URL below is the latest release tag on the [Releases](https://github.com/dropbox/dbxcli/releases) page.
```sh
$ wget https://github.com/dropbox/dbxcli/releases/download/v2.0.9/dbxcli-darwin-amd64 
$ mv dbxcli-darwin-amd64 dbxcli
```
5. Finally, make the binary an executable file and you are good to go!
```
$ chmod +x dbxcli
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
