#!/usr/bin/env bash
# Validate the Claude Code dbxcli plugin structure and detection script.
# Does not require Dropbox authentication.
set -euo pipefail

root_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
plugin_dir="$root_dir/dist/claude/dbxcli-plugin"
source_dir="$root_dir/skills/dbxcli"
pass=0
fail=0

check() {
  local desc="$1"
  shift
  if "$@" >/dev/null 2>&1; then
    echo "  ok  $desc"
    pass=$((pass + 1))
  else
    echo "  FAIL  $desc"
    fail=$((fail + 1))
  fi
}

echo "=== Marketplace structure ==="

marketplace_file="$root_dir/.claude-plugin/marketplace.json"
check "marketplace.json exists" test -f "$marketplace_file"
check "marketplace.json is valid JSON" python3 -c "import json,sys; json.load(open(sys.argv[1]))" "$marketplace_file"
check "marketplace.json has owner field" python3 -c "
import json,sys
d = json.load(open(sys.argv[1]))
assert 'owner' in d, 'owner missing'
assert 'name' in d.get('owner', {}), 'owner.name missing'
" "$marketplace_file"
check "marketplace.json has plugins array" python3 -c "
import json,sys
d = json.load(open(sys.argv[1]))
assert 'plugins' in d, 'plugins missing'
assert isinstance(d['plugins'], list), 'plugins must be array'
" "$marketplace_file"

echo ""
echo "=== Plugin structure (plugin/claude) ==="

source_plugin_dir="$root_dir/plugin/claude"
check "plugin source directory exists" test -d "$source_plugin_dir"
check "plugin source plugin.json exists" test -f "$source_plugin_dir/.claude-plugin/plugin.json"

echo ""
echo "=== Plugin structure (dist) ==="

check "plugin dist directory exists" test -d "$plugin_dir"
check "plugin.json exists" test -f "$plugin_dir/.claude-plugin/plugin.json"
check "plugin.json is valid JSON" python3 -c "import json,sys; json.load(open(sys.argv[1]))" "$plugin_dir/.claude-plugin/plugin.json"
check "plugin.json has name field" python3 -c "
import json,sys
d = json.load(open(sys.argv[1]))
assert d.get('name') == 'dbxcli', 'name mismatch'
" "$plugin_dir/.claude-plugin/plugin.json"

check "hooks.json exists" test -f "$plugin_dir/hooks/hooks.json"
check "hooks.json is valid JSON" python3 -c "import json,sys; json.load(open(sys.argv[1]))" "$plugin_dir/hooks/hooks.json"
check "hooks.json has Setup hook" python3 -c "
import json,sys
d = json.load(open(sys.argv[1]))
assert 'Setup' in d.get('hooks', {}), 'no Setup hook'
" "$plugin_dir/hooks/hooks.json"

check "detect-dbxcli.sh exists" test -f "$plugin_dir/scripts/detect-dbxcli.sh"
check "detect-dbxcli.sh is executable" test -x "$plugin_dir/scripts/detect-dbxcli.sh"
check "README.md exists" test -f "$plugin_dir/README.md"

echo ""
echo "=== Skill content (plugin/claude) ==="

source_plugin_dir="$root_dir/plugin/claude"

check "plugin/claude/skills/dbxcli/ exists" test -d "$source_plugin_dir/skills/dbxcli"
check "plugin/claude skill SKILL.md exists" test -f "$source_plugin_dir/skills/dbxcli/SKILL.md"
check "plugin/claude skill references/ exists" test -d "$source_plugin_dir/skills/dbxcli/references"

echo ""
echo "=== Skill content (dist) ==="

check "dist skills/dbxcli/ directory exists" test -d "$plugin_dir/skills/dbxcli"
check "dist SKILL.md exists" test -f "$plugin_dir/skills/dbxcli/SKILL.md"
check "dist references/safety.md exists" test -f "$plugin_dir/skills/dbxcli/references/safety.md"
check "dist references/automation.md exists" test -f "$plugin_dir/skills/dbxcli/references/automation.md"
check "dist references/tool-integration.md exists" test -f "$plugin_dir/skills/dbxcli/references/tool-integration.md"

echo ""
echo "=== Skill sync validation ==="

check "plugin/claude/skills/dbxcli matches canonical source" diff -qr "$source_dir" "$source_plugin_dir/skills/dbxcli"
check "dist SKILL.md matches canonical source" diff -q "$source_dir/SKILL.md" "$plugin_dir/skills/dbxcli/SKILL.md"
check "dist safety.md matches canonical source" diff -q "$source_dir/references/safety.md" "$plugin_dir/skills/dbxcli/references/safety.md"

echo ""
echo "=== Host validation ==="

if command -v claude >/dev/null 2>&1; then
  check "claude plugin validate passes" claude plugin validate "$plugin_dir"
else
  echo "  skip  claude CLI not found — host validation skipped"
fi

echo ""
echo "=== Detection script ==="

output="$("$plugin_dir/scripts/detect-dbxcli.sh" 2>&1)" || true
if command -v dbxcli >/dev/null 2>&1; then
  check "detect script finds installed dbxcli" grep -q "installed" <<< "$output"
  check "dbxcli version --output=json returns ok" bash -c 'dbxcli version --output=json | python3 -c "import json,sys; assert json.load(sys.stdin)[\"ok\"]"'
else
  check "detect script reports dbxcli missing" grep -q "not installed" <<< "$output"
fi

echo ""
echo "=== Results: $pass passed, $fail failed ==="
[[ $fail -eq 0 ]]
