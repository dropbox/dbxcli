# Agent instructions

Read `CONTRIBUTING.md` before making changes. Keep fixes focused and preserve
existing CLI behavior and JSON contracts unless the task requires changing them.
Inspect the working tree first and preserve unrelated changes.

## Required validation

For Go code or test changes, run all of the following before reporting the work
as ready, committing, pushing, or opening/updating a pull request:

```sh
golangci-lint run
govulncheck ./...
go test ./...
go vet ./...
go build ./...
git diff --check
```

`golangci-lint` is required even when tests, vet, and build pass. Those commands
do not cover all lint checks. Use `.golangci.yml`; do not disable checks merely
to make validation pass.

Use the Go toolchain and golangci-lint versions specified in
`.github/workflows/ci.yml` and check `.github/workflows/release.yml` when working
on releases. A local linter must support the Go toolchain used for analysis.
Do not substitute a different toolchain silently.

For concurrency changes, also run `go test -race ./...`. For workflow changes,
run `actionlint`. Check the CI workflows for other validation relevant to the
files changed.

If command structure, help, or JSON metadata changes, regenerate and inspect:

```sh
go run ./tools/gen-docs
go run ./tools/gen-json-schemas
```

Fix validation failures before publishing. If a check cannot run, report the
exact command and blocker; do not claim it passed. Distinguish local simulator
tests from live Dropbox API tests.

For documentation-only changes, inspect the diff and run `git diff --check`;
Go validation is unnecessary unless generated behavior is affected. Successful
checks need not be repeated if the code and toolchain have not changed.

## Final review

Inspect `git status --short`, `git diff --stat`, and the complete final diff.
Report the files changed, checks actually run, and remaining limitations.
Do not commit, push, or open a pull request unless the user requests it.
