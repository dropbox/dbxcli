# Homebrew formula maintenance

The public Homebrew formula is available at:

https://formulae.brew.sh/formula/dbxcli

The `homebrew/core` formula is the source of truth for `brew install dbxcli`.
The local formula at `packaging/homebrew/dbxcli.rb` is only a validation and
fallback copy.

Homebrew may pick up new tagged releases through livecheck/autobump, so a manual
`homebrew/core` PR is not always required.

After publishing a release, verify Homebrew state:

```sh
brew livecheck dbxcli
brew info dbxcli
```

Validate the local fallback formula when formula logic changes:

```sh
brew audit --new --formula ./packaging/homebrew/dbxcli.rb
brew install --build-from-source ./packaging/homebrew/dbxcli.rb
brew test ./packaging/homebrew/dbxcli.rb
brew uninstall dbxcli
```

Submit a manual `homebrew/core` update only if Homebrew does not pick up the
release automatically or if formula logic needs to change.
