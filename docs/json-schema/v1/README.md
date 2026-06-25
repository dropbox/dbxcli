# dbxcli JSON schema v1

These schemas describe the stable top-level JSON envelopes emitted by
`dbxcli --output=json`.

- `success.schema.json` validates successful command responses.
- `error.schema.json` validates command error responses.
- `commands.json` documents command-specific input/result payload names,
  result statuses, result kinds, and warning codes.

Successful responses always include:

- `ok: true`
- `schema_version: "1"`
- `command`: command path without the binary name, such as `ls` or
  `share-link create`
- `input`: command-specific request fields
- `results`: command-specific result objects; every result includes `status`,
  `kind`, `input`, and `result`
- `warnings`: machine-actionable warnings, or `[]`

Error responses always include:

- `ok: false`
- `schema_version: "1"`
- `command`: command path when available, or `dbxcli` for root/pre-parse errors
- `error.message`: human-readable error text
- `error.code`: stable machine-readable error code
- `warnings`: machine-actionable warnings, or `[]`

Command-specific `input` and `result` payload contracts are listed in
`commands.json` and locked by the golden contract fixtures under
`cmd/testdata/json_contract/`.
