# Homebrew formula validation

This formula is a draft for submission to `homebrew/core`.
Homebrew should build `dbxcli` from the tagged source tarball, not from release binaries.

Before submitting a new version:

1. Update `url` and `sha256` in `dbxcli.rb` to the new tag source tarball.
2. Run:
   ```sh
   brew audit --new --formula ./packaging/homebrew/dbxcli.rb
   brew install --build-from-source ./packaging/homebrew/dbxcli.rb
   brew test ./packaging/homebrew/dbxcli.rb
   brew uninstall dbxcli
   ```
3. Copy the formula into a `homebrew/core` pull request after the release is published.
