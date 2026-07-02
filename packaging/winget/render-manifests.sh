#!/usr/bin/env bash

set -euo pipefail

package_id="Dropbox.dbxcli"
version="${1:-${VERSION:-${GITHUB_REF_NAME:-}}}"

if [[ -z "${version}" ]]; then
  echo "usage: $0 vX.Y.Z" >&2
  echo "or set VERSION/GITHUB_REF_NAME" >&2
  exit 1
fi

version="${version#v}"
dist_dir="${DIST_DIR:-dist}"
checksum_file="${dist_dir}/SHA256SUMS"
archive_name="dbxcli_${version}_windows_amd64.zip"

if [[ ! -f "${checksum_file}" ]]; then
  echo "missing checksum file: ${checksum_file}" >&2
  exit 1
fi

if ! checksum="$(
  awk -v archive="${archive_name}" '
    $2 == archive {
      print toupper($1)
      found = 1
    }
    END {
      if (!found) {
        exit 1
      }
    }
  ' "${checksum_file}"
)"; then
  echo "missing checksum for ${archive_name} in ${checksum_file}" >&2
  exit 1
fi

if [[ -z "${checksum}" ]]; then
  echo "missing checksum for ${archive_name} in ${checksum_file}" >&2
  exit 1
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
template_dir="${script_dir}/templates"
output_root="${OUTPUT_DIR:-${dist_dir}/winget}"
output_dir="${output_root}/manifests/d/Dropbox/dbxcli/${version}"

mkdir -p "${output_dir}"

render_template() {
  local input="$1"
  local output="$2"

  sed \
    -e "s/{{VERSION}}/${version}/g" \
    -e "s/{{WINDOWS_AMD64_SHA256}}/${checksum}/g" \
    "${input}" > "${output}"
}

render_template "${template_dir}/${package_id}.yaml.template" "${output_dir}/${package_id}.yaml"
render_template "${template_dir}/${package_id}.installer.yaml.template" "${output_dir}/${package_id}.installer.yaml"
render_template "${template_dir}/${package_id}.locale.en-US.yaml.template" "${output_dir}/${package_id}.locale.en-US.yaml"

echo "Created WinGet manifests in ${output_dir}:"
echo "  ${package_id}.yaml"
echo "  ${package_id}.installer.yaml"
echo "  ${package_id}.locale.en-US.yaml"
