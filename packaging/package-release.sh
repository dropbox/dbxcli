#!/usr/bin/env bash

set -euo pipefail

version="${1:-${VERSION:-${TRAVIS_TAG:-${GITHUB_REF_NAME:-dev}}}}"
asset_version="${version#v}"
dist_dir="${DIST_DIR:-dist}"
work_dir="${dist_dir}/package"

default_targets=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm64"
  "linux/arm"
  "openbsd/amd64"
  "windows/amd64"
)

if [[ -n "${TARGETS:-}" ]]; then
  read -r -a targets <<< "${TARGETS}"
else
  targets=("${default_targets[@]}")
fi

if [[ ! -f README.md || ! -f LICENSE ]]; then
  echo "README.md and LICENSE must exist in the repository root" >&2
  exit 1
fi

mkdir -p "${dist_dir}"
rm -rf "${work_dir}"
rm -f "${dist_dir}"/dbxcli_"${asset_version}"_*.tar.gz
rm -f "${dist_dir}"/dbxcli_"${asset_version}"_*.zip
rm -f "${dist_dir}/SHA256SUMS"
mkdir -p "${work_dir}"

archive_names=()

for target in "${targets[@]}"; do
  goos="${target%/*}"
  goarch="${target#*/}"
  bin_name="dbxcli-${goos}-${goarch}"
  exe_name="dbxcli"

  if [[ "${goos}" == "windows" ]]; then
    bin_name="${bin_name}.exe"
    exe_name="dbxcli.exe"
  fi

  bin_path="${dist_dir}/${bin_name}"
  if [[ ! -f "${bin_path}" ]]; then
    echo "missing built binary: ${bin_path}" >&2
    exit 1
  fi

  archive_base="dbxcli_${asset_version}_${goos}_${goarch}"
  stage_dir="${work_dir}/${archive_base}"
  mkdir -p "${stage_dir}"
  cp "${bin_path}" "${stage_dir}/${exe_name}"
  cp README.md LICENSE "${stage_dir}/"

  if [[ "${goos}" == "windows" ]]; then
    archive_name="${archive_base}.zip"
    if command -v zip >/dev/null 2>&1; then
      (cd "${work_dir}" && zip -qr "../${archive_name}" "${archive_base}")
    elif command -v pwsh >/dev/null 2>&1; then
      (cd "${work_dir}" && pwsh -NoLogo -NoProfile -NonInteractive -Command "Compress-Archive -LiteralPath '${archive_base}' -DestinationPath '../${archive_name}' -Force")
    elif command -v powershell.exe >/dev/null 2>&1; then
      (cd "${work_dir}" && powershell.exe -NoLogo -NoProfile -NonInteractive -Command "Compress-Archive -LiteralPath '${archive_base}' -DestinationPath '../${archive_name}' -Force")
    else
      echo "zip, pwsh, or powershell.exe is required to create Windows archives" >&2
      exit 1
    fi
  else
    archive_name="${archive_base}.tar.gz"
    (cd "${work_dir}" && tar -czf "../${archive_name}" "${archive_base}")
  fi

  archive_names+=("${archive_name}")
done

if command -v sha256sum >/dev/null 2>&1; then
  checksum_cmd=(sha256sum)
else
  checksum_cmd=(shasum -a 256)
fi

checksum_names=("${archive_names[@]}")
sbom_name="dbxcli_${asset_version}_sbom.spdx.json"
if [[ -f "${dist_dir}/${sbom_name}" ]]; then
  checksum_names+=("${sbom_name}")
fi

(cd "${dist_dir}" && "${checksum_cmd[@]}" "${checksum_names[@]}" > SHA256SUMS)

echo "Created release archives in ${dist_dir}:"
for archive_name in "${archive_names[@]}"; do
  echo "  ${archive_name}"
done
for checksum_name in "${checksum_names[@]:${#archive_names[@]}}"; do
  echo "  ${checksum_name}"
done
echo "  SHA256SUMS"
