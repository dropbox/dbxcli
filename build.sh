#!/bin/bash

set -e

RAW_VERSION="${VERSION:-${TRAVIS_TAG:-${GITHUB_REF_NAME:-dev}}}"
VERSION="${RAW_VERSION#v}"
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
    (
        if [ "$target" = "linux/amd64" ]; then
            # GOAMD64 does not constrain C code compiled by cgo.
            export GOAMD64=v1
            export CGO_CFLAGS="${CGO_CFLAGS:--O2 -g} -march=x86-64"
        fi
        GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="${LDFLAGS}" -o "$output" .
    )
done
