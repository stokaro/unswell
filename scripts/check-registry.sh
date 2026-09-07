#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
version=$(sed -n 's/^const Version = "\(.*\)"/\1/p' result.go)
jq --exit-status --arg version "$version" '
  .name == "io.github.stokaro/unswell" and .version == $version and
  .packages == [{registryType: "oci", identifier: ("ghcr.io/stokaro/unswell-mcp:" + $version), transport: {type: "stdio"}}]
' mcp/server.json >/dev/null
bin/mcp-publisher validate mcp/server.json
temporary=$(mktemp -d)
trap 'rm -rf "$temporary"' EXIT
jq '.name = "invalid-name"' mcp/server.json >"$temporary/invalid.json"
if bin/mcp-publisher validate "$temporary/invalid.json" >"$temporary/validation.log" 2>&1; then
  printf 'MCP registry validation accepted an invalid server name.\n' >&2
  exit 1
fi
printf 'MCP registry metadata passed; the validator rejected an invalid server name.\n'

# Exercise the response gate locally; only verify-registry.sh proves publication.
jq --null-input --slurpfile server mcp/server.json \
  '{server: $server[0], _meta: {"io.modelcontextprotocol.registry/official": {status: "active"}}}' \
  >"$temporary/response.json"
response_check=(jq --exit-status --slurpfile expected mcp/server.json --from-file scripts/registry-response.jq)
"${response_check[@]}" "$temporary/response.json" >/dev/null
for mutation in \
  '.server.name = "io.github.other/unswell"' \
  '.server.version = "0.0.0"' \
  '.server.packages = []' \
  '.server.repository.url = "https://github.com/example/other"' \
  '._meta["io.modelcontextprotocol.registry/official"].status = "deleted"'; do
  jq "$mutation" "$temporary/response.json" >"$temporary/changed.json"
  if "${response_check[@]}" "$temporary/changed.json" >/dev/null; then
    printf 'Registry response verification accepted a mismatched or inactive entry.\n' >&2
    exit 1
  fi
done
printf 'The response gate rejected every mismatched or inactive fixture.\n'
