#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
name=$(jq --raw-output --exit-status '.name | @uri' mcp/server.json)
version=$(jq --raw-output --exit-status '.version | @uri' mcp/server.json)
mkdir -p artifacts/registry
curl --fail --silent --show-error --retry 3 --retry-all-errors --connect-timeout 10 --max-time 30 \
  "https://registry.modelcontextprotocol.io/v0.1/servers/$name/versions/$version" \
  --output artifacts/registry/server.json
jq --exit-status --slurpfile expected mcp/server.json \
  --from-file scripts/registry-response.jq artifacts/registry/server.json >/dev/null
printf 'The public MCP Registry contains the expected active version and OCI package.\n'
