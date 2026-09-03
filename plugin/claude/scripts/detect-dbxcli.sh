#!/usr/bin/env bash
# Detect dbxcli and print installation guidance if missing.
# This script is informational only — it never installs software.
set -euo pipefail

if command -v dbxcli >/dev/null 2>&1; then
  version_json="$(dbxcli version --output=json 2>/dev/null)" || true

  if printf '%s' "$version_json" | grep -q '"ok":true'; then
    version="$(printf '%s' "$version_json" | grep -o '"version":"[^"]*"' | head -1 | cut -d'"' -f4)"
    echo "dbxcli ${version} is installed and responding."
    exit 0
  fi

  echo "dbxcli is on PATH but did not return valid JSON."
  echo "Try running:  dbxcli version --output=json"
  exit 0
fi

echo "dbxcli is not installed."
echo ""

case "$(uname -s)" in
  Darwin)
    echo "Install on macOS:"
    echo "  brew install dbxcli"
    ;;
  Linux)
    echo "Install on Linux (pick one):"
    echo "  • Download from https://github.com/dropbox/dbxcli/releases"
    echo "  • go install github.com/dropbox/dbxcli/v3@latest"
    ;;
  MINGW*|MSYS*|CYGWIN*)
    echo "Install on Windows:"
    echo "  winget install --exact --id Dropbox.dbxcli"
    ;;
  *)
    echo "Install from source or download a release:"
    echo "  https://github.com/dropbox/dbxcli/releases"
    ;;
esac

echo ""
echo "After installing, run:  dbxcli version --output=json"
