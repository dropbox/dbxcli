# Contributing to `dbxcli`

Thanks for your interest in improving `dbxcli`.

## Before you start

For non-trivial contributions, complete the Dropbox Contributor License Agreement:

https://opensource.dropbox.com/cla/

If you're planning a larger feature or behavior change, consider opening an issue first.

## Development

Requirements:
- Go 1.25+
- Git

Build:
```
go build .
```

Run:
```
go run . --help
```

## Repository layout
- cmd/            Cobra commands
- docs/           Generated documentation
- docs/json-schema/v1/  Generated JSON schemas
- packaging/      Release packaging
- tools/          Documentation and schema generators

## Design principles

`dbxcli` is designed for humans, scripts, CI pipelines, and AI agents.

When contributing, prefer changes that preserve:
- backward compatibility where practical
- stable automation interfaces
- predictable JSON output and JSON help
- composable command-line behavior
- cross-platform support

## Making changes

Keep pull requests focused.

When adding or modifying commands:
- update help text
- add or update tests
- preserve JSON compatibility when possible
- regenerate documentation
- regenerate JSON schemas if applicable

See:
- docs/automation.md
- docs/json-schema/v1/README.md

## Validation

Run before submitting:
```
go test ./...
go vet ./...
go build ./...
```

For concurrency-sensitive changes:
```
go test -race ./...
```

Recommended:
```
golangci-lint run
govulncheck ./...
```

## Generated files

If command structure, help, or JSON metadata changes:
```
go run ./tools/gen-docs
go run ./tools/gen-json-schemas
```

Ensure the generated files are committed.

## Pull requests

Please include:
- what changed
- why
- validation performed
- related issue (if any)

Small, focused pull requests are easier to review.

## Reporting bugs

Include:
- `dbxcli` version
- operating system
- command executed
- expected behavior
- actual behavior
- reproduction steps

Remove access tokens and other sensitive information before posting logs.
