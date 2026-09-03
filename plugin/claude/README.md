# dbxcli plugin for Claude Code

A thin packaging layer that exposes the canonical `dbxcli` skill to Claude Code
as an installable plugin.

## What this plugin does

- Packages the portable `dbxcli` skill so Claude Code discovers it as
  `/dbxcli:dbxcli`.
- Provides a Setup hook that detects whether `dbxcli` is installed and prints
  platform-specific installation guidance if it is missing.

The plugin does not implement Dropbox API calls, maintain a command catalog, or
store credentials. The local `dbxcli` executable is the runtime, and its JSON
help manifest is the source of truth for available commands, flags, and schemas.

## Prerequisites

- **Claude Code** — the plugin host.
- **dbxcli** — installed and on `PATH`. The plugin detects its absence but does
  not install it automatically.

Install dbxcli:

| Platform | Command |
|----------|---------|
| macOS    | `brew install dbxcli` |
| Windows  | `winget install --exact --id Dropbox.dbxcli` |
| Linux    | Download from [releases](https://github.com/dropbox/dbxcli/releases) or `go install github.com/dropbox/dbxcli/v3@latest` |

For interactive use, authenticate with `dbxcli login`; automation may use
supported non-interactive authentication.

## How the canonical skill is included

The skill source of truth is `skills/dbxcli/` in the dbxcli repository. The
packaging script (`scripts/package-skills.sh`) copies that portable skill into
each platform artifact. Claude Code packaging adds only its manifest, hook,
script, and README; it does not own the skill's behavior.

## Structure

```
dist/claude/dbxcli-plugin/
├── .claude-plugin/plugin.json      # Plugin manifest
├── skills/dbxcli/                  # Copied from skills/dbxcli/
│   ├── SKILL.md
│   ├── agents/
│   └── references/
├── hooks/hooks.json                # Setup hook
├── scripts/detect-dbxcli.sh        # Detection script
└── README.md                       # This file
```

## Build

```bash
./scripts/package-skills.sh
```

## Test locally

Load the plugin without installing:

```bash
claude --plugin-dir dist/claude/dbxcli-plugin
```

Run the structural validation (no Dropbox auth required):

```bash
./scripts/test-plugin.sh
```

If dbxcli is installed, the test also verifies that the Setup hook detection
script runs correctly and that `dbxcli version --output=json` returns a valid
response.
