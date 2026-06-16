# Chocolatey package validation

The public Chocolatey package uses the Windows release archive and its SHA256 checksum.

Before submitting a new version:

1. Publish a GitHub Release with `dbxcli_<version>_windows_amd64.zip` and `SHA256SUMS`.
2. Copy the Windows archive checksum into `tools/chocolateyInstall.ps1`.
3. Pack and test locally on Windows:
   ```powershell
   choco pack packaging\chocolatey\dbxcli\dbxcli.nuspec --version <version> --outputdirectory dist
   choco install dbxcli --source dist --version <version> -y
   dbxcli --help
   choco uninstall dbxcli -y
   ```
4. Push the resulting `.nupkg` to the Chocolatey community repository.
