$ErrorActionPreference = 'Stop'

$toolsDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$packageName = $env:ChocolateyPackageName
$packageVersion = $env:ChocolateyPackageVersion

$url64 = "https://github.com/dropbox/dbxcli/releases/download/v$packageVersion/dbxcli_${packageVersion}_windows_amd64.zip"
$checksum64 = "__WINDOWS_AMD64_SHA256__"

function Install-DbxcliExe {
  $rootExePath = Join-Path $toolsDir 'dbxcli.exe'
  if (Test-Path -LiteralPath $rootExePath) {
    return
  }

  $archiveExePath = Join-Path $toolsDir "dbxcli_${packageVersion}_windows_amd64\dbxcli.exe"
  if (!(Test-Path -LiteralPath $archiveExePath)) {
    throw "Could not find dbxcli.exe after extracting the release archive"
  }

  Copy-Item -LiteralPath $archiveExePath -Destination $rootExePath -Force
}

function Get-Sha256Checksum($path) {
  $sha256 = [System.Security.Cryptography.SHA256]::Create()
  $stream = [System.IO.File]::OpenRead($path)

  try {
    return [System.BitConverter]::ToString($sha256.ComputeHash($stream)).Replace('-', '').ToLowerInvariant()
  }
  finally {
    $stream.Dispose()
    $sha256.Dispose()
  }
}

if ($env:DBXCLI_CHOCOLATEY_URL64) {
  $url64 = $env:DBXCLI_CHOCOLATEY_URL64
}

if ($env:DBXCLI_CHOCOLATEY_CHECKSUM64) {
  $checksum64 = $env:DBXCLI_CHOCOLATEY_CHECKSUM64
}

if ($checksum64 -eq "__WINDOWS_AMD64_SHA256__") {
  throw "dbxcli Chocolatey package checksum was not set for version $packageVersion"
}

$isRemoteUrl = $url64 -match '^https?://'

if (!$isRemoteUrl) {
  if (!(Test-Path -LiteralPath $url64)) {
    throw "Local dbxcli archive not found: $url64"
  }

  $actualChecksum = Get-Sha256Checksum $url64
  if ($actualChecksum -ne $checksum64.ToLowerInvariant()) {
    throw "Checksum mismatch for $url64. Expected $checksum64, got $actualChecksum"
  }

  Get-ChocolateyUnzip -FileFullPath $url64 -Destination $toolsDir
  Install-DbxcliExe
  return
}

$packageArgs = @{
  packageName    = $packageName
  unzipLocation  = $toolsDir
  url64bit       = $url64
  checksum64     = $checksum64
  checksumType64 = 'sha256'
}

Install-ChocolateyZipPackage @packageArgs
Install-DbxcliExe
