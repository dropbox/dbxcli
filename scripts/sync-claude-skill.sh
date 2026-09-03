#!/usr/bin/env bash
# Sync the canonical skill source to the Claude Code plugin directory.
# The canonical source is skills/dbxcli/.
# The Claude plugin must be self-contained, so it includes a copy.
set -euo pipefail

root_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
source_dir="$root_dir/skills/dbxcli"
target_dir="$root_dir/plugin/claude/skills/dbxcli"

if [[ ! -d "$source_dir" ]]; then
  echo "error: canonical skill source not found: $source_dir" >&2
  exit 1
fi

echo "Syncing canonical skill to Claude plugin..."
echo "  source: $source_dir"
echo "  target: $target_dir"

rm -rf "$target_dir"
mkdir -p "$(dirname "$target_dir")"
cp -R "$source_dir" "$target_dir"

echo "✓ Sync complete"
echo ""
echo "The plugin is now self-contained. Verify with:"
echo "  ./scripts/test-plugin.sh"
