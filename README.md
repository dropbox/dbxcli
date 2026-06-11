# `dbxcli`: A command line tool for Dropbox users and team admins [UNOFFICIAL]

[![Build Status](https://travis-ci.org/dropbox/dbxcli.svg?branch=master)](https://travis-ci.org/dropbox/dbxcli)
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
  * Recursive directory uploads (`put -r`)
  * Retry with exponential backoff for uploads and downloads
  * Supports a growing set of Team operations

## Installation

Download pre-compiled binaries for Mac, Windows and Linux from the [releases](https://github.com/dropbox/dbxcli/releases) page.

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
4. Download the `dbxcli` binary for OSX and rename it.  *IMPORTANT:* Check that the tag `v3.2.1` on the URL below is the latest release tag on the [Releases](https://github.com/dropbox/dbxcli/releases) page.
```sh
$ wget https://github.com/dropbox/dbxcli/releases/download/v3.2.1/dbxcli-darwin-amd64 
$ mv dbxcli-darwin-amd64 dbxcli
```
5. Finally, make the binary an executable file and you are good to go!
```
$ chmod +x dbxcli
```

### Instructions for building yourself
For newcomers the go build process can be a bit arcane, these steps can be followed to build `dbxcli` yourself.

1. Make sure `git`, `go`, and `gox` are installed. 
2. Create a Go folder. For example, `mkdir $HOME/go` or `mkdir $HOME/.go`. Navigate to it.
3. `go get github.com/dropbox/dbxcli`. That's right, you don't manually clone it, this does it for you.
4. `cd ~/go/src/github.com/dropbox/dbxcli` (adapt accordingly based on step 2).

Now we need to pause for a second to get development keys. 

5. Head to `https://www.dropbox.com/developers/apps` (sign in if necessary) and choose "Create app". Use the Dropbox API and give it Full Dropbox access. Name and create the app.
6. You'll be presented with a dashboard with an "App key" and an "App secret".
7. Replace the value for `personalAppKey` in  `root.go` with the key from the webpage.
8. Replace the value for `personalAppSecret` with the secret from the webpage.

Finally we're ready to build. Run `go build`, and you'll see a `dbxcli` binary has been created in the current directory. Congrats, we're done!

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
  get         Download a file
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
  -v, --verbose            Enable verbose logging

Use "dbxcli [command] --help" for more information about a command.
```

### Listing files

```sh
$ dbxcli ls -l /Photos
Revision              Size    Last modified Path
abc123                1.2 MiB 3 weeks ago   /Photos/vacation.jpg
def456                4.5 MiB 1 month ago   /Photos/family.png
```

#### Time format

By default, `ls -l` and `search -l` show relative timestamps ("3 weeks ago"). Use `--time-format` for absolute dates:

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

All `--sort`, `--reverse`, `--time`, and `--time-format` flags work with both `ls` and `search`.

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
  -v, --verbose            Enable verbose logging

Use "dbxcli team [command] --help" for more information about a command.
```

The `--verbose` option will turn on verbose logging and is useful for debugging.

### Uploading files and directories

```sh
$ dbxcli put file.txt /destination/file.txt        # upload a single file
$ dbxcli put -r ./project /backup/project          # recursively upload a directory
$ dbxcli put -r -w 8 ./large-folder /backup/large  # use 8 workers per large file
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
