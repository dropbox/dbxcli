#!/bin/bash

set -e

dbxcli=$(realpath $1)
echo "Testing binary at ${dbxcli}"

echo "Testing du"
${dbxcli} du > /dev/null

echo "Testing ls"
${dbxcli} ls > /dev/null

echo "Testing ls -l"
${dbxcli} ls -l > /dev/null

echo "Testing search"
${dbxcli} search dropbox > /dev/null

echo "All tests passed!"
