#!/usr/bin/env bash
set -euo pipefail

dbxcli_bin="${1:-dbxcli}"
if ! command -v "$dbxcli_bin" >/dev/null 2>&1 && [[ ! -x "$dbxcli_bin" ]]; then
  echo "dbxcli executable not found: $dbxcli_bin" >&2
  exit 1
fi

assert_json() {
  local json="$1"
  shift
  jq -e "$@" >/dev/null <<<"$json" || { echo "JSON assertion failed" >&2; exit 1; }
}

version="$($dbxcli_bin version --output=json)"
assert_json "$version" '.ok == true and .command == "version"'

root_help="$($dbxcli_bin --help --output=json)"
assert_json "$root_help" '.ok == true and (.results | length > 0)'

put_help="$($dbxcli_bin put --help --output=json)"
assert_json "$put_help" '.ok == true and .results[0].result.input_schema.type == "object"'
assert_json "$put_help" '.results[0].result.supports_structured_output == true'
assert_json "$put_help" '.results[0].result.input_schema.properties.if_exists["x-cli-name"] == "if-exists"'

planned_put='{"source":"report.md","target":"/Reports/report.md","if_exists":"fail","dry_run":true}'
assert_json "$put_help" --argjson planned "$planned_put" '
  .results[0].result.input_schema as $schema |
  ($planned | keys | all(. as $key | $schema.properties[$key] != null)) and
  $schema.properties.if_exists["x-cli-name"] == "if-exists" and
  $schema.properties.dry_run["x-cli-name"] == "dry-run" and
  $schema.properties.dry_run.type == "boolean"
'

rm_help="$($dbxcli_bin rm --help --output=json)"
assert_json "$rm_help" '.ok == true and ([.results[0].result.flags[] | select(.name == "dry-run")] | length == 1)'
assert_json "$rm_help" '.results[0].result.destructive_level != "none"'

ls_help="$($dbxcli_bin ls --help --output=json)"
assert_json "$ls_help" '.ok == true and .results[0].result.input_schema.properties.limit["x-cli-name"] == "limit"'

search_help="$($dbxcli_bin search --help --output=json)"
assert_json "$search_help" '.ok == true and .results[0].result.input_schema.properties.content["x-cli-name"] == "content"'
assert_json "$search_help" '.results[0].result.input_schema.properties.path_scope["x-cli-name"] == "path-scope"'

login_help="$($dbxcli_bin login --help --output=json)"
assert_json "$login_help" '.ok == true and .results[0].result.supports_structured_output == false'

set +e
invalid="$($dbxcli_bin ls --output=json --not-a-flag / 2>/dev/null)"
exit_code=$?
set -e
[[ $exit_code -eq 7 ]] || { echo "expected unknown flag to exit 7, got $exit_code" >&2; exit 1; }
assert_json "$invalid" '.ok == false and .error.code == "unknown_flag"'

echo "dbxcli skill contract checks passed (13 assertions)"
