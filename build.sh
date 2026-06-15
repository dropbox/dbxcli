#!/bin/bash

set -e

VERSION="${TRAVIS_TAG:-${GITHUB_REF_NAME:-dev}}"
LDFLAGS="-s -w -X main.version=${VERSION}"

TARGETS="darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 linux/arm openbsd/amd64 windows/amd64"

mkdir -p dist

for target in $TARGETS; do
    GOOS="${target%/*}"
    GOARCH="${target#*/}"
    output="dist/dbxcli-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output="${output}.exe"
    fi
    echo "Building ${target}..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="${LDFLAGS}" -o "$output" .
done
