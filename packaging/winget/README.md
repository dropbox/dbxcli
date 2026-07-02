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

## Validate and submit

From the `winget-pkgs` checkout:

```powershell
winget validate .\manifests\d\Dropbox\dbxcli\X.Y.Z
winget install --manifest .\manifests\d\Dropbox\dbxcli\X.Y.Z
```

Open a pull request against `microsoft/winget-pkgs` after local validation
passes.

Do not add `winget install Dropbox.dbxcli` to the main README until the package
has been accepted into the public WinGet source.
