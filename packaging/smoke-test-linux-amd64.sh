#!/usr/bin/env bash

set -euo pipefail

archive="${1:?Usage: $0 ARCHIVE}"
work_dir="$(mktemp -d)"
trap 'rm -rf "$work_dir"' EXIT

tar -xzf "$archive" --strip-components=1 -C "$work_dir"

# Exercise the packaged binary's cgo startup on a baseline CPU without AVX.
timeout 30s qemu-x86_64 -cpu qemu64 "$work_dir/dbxcli" version
