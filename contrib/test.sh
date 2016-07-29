#!/bin/bash

set -e

dbxcli=$(realpath $1)
echo "Testing binary at ${dbxcli}"

echo "Testing version"
${dbxcli} version > /dev/null

echo "Testing du"
${dbxcli} du > /dev/null

echo "Testing ls"
${dbxcli} ls > /dev/null

echo "Testing ls -l"
${dbxcli} ls -l > /dev/null

echo "Testing search"
${dbxcli} search dropbox > /dev/null

echo "Testing mkdir"
d=dbxcli-$(date +%s)
${dbxcli} mkdir ${d} > /dev/null

echo "Testing rmdir"
${dbxcli} rmdir ${d} > /dev/null

#echo "Testing cp"
#${dbxcli} cp ${dbxcli} ${d}/dbxcli > /dev/null

echo "All tests passed!"
