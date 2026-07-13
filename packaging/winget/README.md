# WinGet package maintenance

WinGet packages are published through the Microsoft community package
repository:

https://github.com/microsoft/winget-pkgs

Use these templates after a GitHub Release is published and the Windows archive
is available at:

```text
https://github.com/dropbox/dbxcli/releases/download/vX.Y.Z/dbxcli_X.Y.Z_windows_amd64.zip
```

The generated manifest is for a portable zip package. It installs the nested
`dbxcli.exe` binary and exposes `dbxcli` as the portable command alias.

## Automatic release updates

The release workflow submits a WinGet update after publishing the GitHub
Release assets. It uses Microsoft's WingetCreate tool to open a pull request in
`microsoft/winget-pkgs`; Microsoft validation and review still apply.

Configure an Actions repository secret named `WINGET_CREATE_GITHUB_TOKEN`
before publishing a release. The secret must contain a GitHub personal access
token (classic) with the `public_repo` scope. WingetCreate does not support
fine-grained personal access tokens.

WingetCreate is pinned and checksum-verified in the release workflow. Update
both its release URL and expected SHA-256 digest when upgrading the tool.

## Generate manifests

Run the normal release packaging first so `dist/SHA256SUMS` exists:

```sh
VERSION=vX.Y.Z ./build.sh
VERSION=vX.Y.Z ./packaging/package-release.sh
```

Then render the WinGet manifests:

```sh
./packaging/winget/render-manifests.sh vX.Y.Z
```

By default this writes:

```text
dist/winget/manifests/d/Dropbox/dbxcli/X.Y.Z/
```

To write directly into a local `winget-pkgs` checkout:

```sh
OUTPUT_DIR=/path/to/winget-pkgs ./packaging/winget/render-manifests.sh vX.Y.Z
```

This creates files under:

```text
/path/to/winget-pkgs/manifests/d/Dropbox/dbxcli/X.Y.Z/
```

## Manual validation and submission

From the `winget-pkgs` checkout:

```powershell
winget validate .\manifests\d\Dropbox\dbxcli\X.Y.Z
winget install --manifest .\manifests\d\Dropbox\dbxcli\X.Y.Z
```

Open a pull request against `microsoft/winget-pkgs` after local validation
passes.

The package is available from the public WinGet source as `Dropbox.dbxcli`:

```powershell
winget install --exact --id Dropbox.dbxcli
```
