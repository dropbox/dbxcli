#!/bin/bash

set -e

dbxcli=$(realpath $1)
echo "Testing binary at ${dbxcli}"

echo "Testing version"
${dbxcli} version

echo "Testing du"
${dbxcli} du

echo "Testing mkdir"
d=dbxcli-$(date +%s)
${dbxcli} mkdir ${d}

echo "Testing put"
${dbxcli} put ${dbxcli} ${d}/dbxcli

echo "Testing get"
${dbxcli} get ${d}/dbxcli /tmp/dbxcli
# Make sure files are the same
cmp --silent ${dbxcli} /tmp/dbxcli

echo "Testing ls -l"
${dbxcli} ls -l ${d}

echo "Testing search"
${dbxcli} search dropbox /${d}

echo "Testing cp"
${dbxcli} cp ${d}/dbxcli ${d}/dbxcli-new

echo "Testing revs"
rev=$(${dbxcli} revs ${d}/dbxcli)

echo "Testing mv"
${dbxcli} mv ${d}/dbxcli ${d}/dbxcli-old

echo "Testing mv"
${dbxcli} mv ${d}/dbxcli ${d}/dbxcli-old/

echo "Testing restore"
${dbxcli} restore ${d}/dbxcli ${rev}

echo "Testing rm -f"
${dbxcli} rm -f ${d}

echo "Testing share commands"

echo "Testing share list folder"
${dbxcli} share list folder
echo "Testing share list link"
${dbxcli} share list link

echo "Testing team commands"
echo "Testing team info"
${dbxcli} team list-groups

echo "Testing team list-members"
${dbxcli} team list-members

echo "Testing ls as"
id=$(${dbxcli} team list-members | grep active | cut -d' ' -f3)
${dbxcli} ls --as-member ${id}

echo "All tests passed!"
