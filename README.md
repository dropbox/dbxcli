# Dropbox CLI

## Installation

## Commands

List files in Dropbox folder.

	dbxcli ls [dropbox://FOLDER]

Put file into Dropbox

	dbxcli put FILE dropbox://DEST

Get file from Dropbox

	dbxcli get dropbox://SRC FILE

Remove file from Dropbox

	dbxcli rm dropbox://PATH

Copy file

	dbxcli cp dropbox://SRC dropbox://DEST

Move file

	dbxcli mv dropbox://SRC dropbox://DEST

List file revisions

	dbxcli revs dropbox://PATH

Restore file

	dbxcli restore dropbox://PATH REVISION
