#!/bin/bash

set -e

dbxcli=$(realpath $1)
echo "Testing binary at ${dbxcli}"

echo "Testing version"
${dbxcli} version > /dev/null

echo "Testing du"
${dbxcli} du > /dev/null

echo "Testing mkdir"
d=dbxcli-$(date +%s)
${dbxcli} mkdir ${d} > /dev/null

echo "Testing put"
${dbxcli} put ${dbxcli} ${d}/dbxcli > /dev/null

echo "Testing get"
${dbxcli} get ${d}/dbxcli /tmp/dbxcli > /dev/null
# Make sure files are the same
cmp --silent ${dbxcli} /tmp/dbxcli

echo "Testing ls -l"
${dbxcli} ls -l ${d} > /dev/null

echo "Testing search"
${dbxcli} search dropbox /${d} > /dev/null

echo "Testing cp"
${dbxcli} cp ${d}/dbxcli ${d}/dbxcli-new

echo "Testing rmdir"
${dbxcli} rmdir ${d}

echo "All tests passed!"
