#!/usr/bin/env bash
set -euo pipefail

root_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
source_dir="$root_dir/skills/dbxcli"
plugin_dir="$root_dir/plugin/claude"
dist_dir="$root_dir/dist"
chatgpt_dir="$dist_dir/chatgpt"
claude_dir="$dist_dir/claude/dbxcli-plugin"
openclaw_dir="$dist_dir/openclaw/dbxcli-plugin"

for required in "$source_dir/SKILL.md" "$source_dir/agents/openai.yaml"; do
  [[ -f "$required" ]] || { echo "missing source file: $required" >&2; exit 1; }
done

for required in "$plugin_dir/.claude-plugin/plugin.json" \
                "$plugin_dir/hooks/hooks.json" \
                "$plugin_dir/scripts/detect-dbxcli.sh" \
                "$plugin_dir/README.md"; do
  [[ -f "$required" ]] || { echo "missing plugin file: $required" >&2; exit 1; }
done

rm -rf "$chatgpt_dir" "$claude_dir" "$openclaw_dir"
mkdir -p "$chatgpt_dir" "$claude_dir/skills" "$openclaw_dir/skills"

# ChatGPT: skill zip
cp -R "$source_dir" "$chatgpt_dir/dbxcli"
(
  cd "$chatgpt_dir"
  rm -f skill.zip
  zip -qr skill.zip dbxcli
)

# Claude Code: plugin with skill, hooks, and scripts
cp -R "$plugin_dir/.claude-plugin" "$claude_dir/.claude-plugin"
cp -R "$plugin_dir/hooks"          "$claude_dir/hooks"
cp -R "$plugin_dir/scripts"        "$claude_dir/scripts"
cp    "$plugin_dir/README.md"      "$claude_dir/README.md"
cp -R "$source_dir"                "$claude_dir/skills/dbxcli"
chmod +x "$claude_dir/scripts/detect-dbxcli.sh"

# OpenClaw: skill directory
cp -R "$source_dir" "$openclaw_dir/skills/dbxcli"
cat > "$openclaw_dir/openclaw.plugin.json" <<'EOF'
{
  "id": "dbxcli",
  "name": "dbxcli",
  "version": "0.1.0",
  "description": "Safely operate Dropbox through a locally installed dbxcli CLI",
  "skills": ["skills"]
}
EOF

echo "built $chatgpt_dir/skill.zip"
echo "built $claude_dir"
echo "built $openclaw_dir"
