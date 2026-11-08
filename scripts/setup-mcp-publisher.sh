#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
version=$(cat .mcp-publisher-version)
platform=$(uname -s)
architecture=$(uname -m)
# Asset IDs belong to the pinned release; the SHA-256 check remains authoritative.
case "$version:$platform:$architecture" in
  1.8.1:Linux:x86_64)
    target=linux_amd64
    asset_id=504323513
    digest=a06c9096dcb9727c13555b6be26c7effa707b01f06a4c561ba7a3635443cf2cc
    ;;
  1.8.1:Linux:aarch64)
    target=linux_arm64
    asset_id=504323512
    digest=8dd75a6cf6845688b5d4e46df58d3ca26d5c8d233bb0626606e1db82c5e883e4
    ;;
  1.8.1:Darwin:arm64)
    target=darwin_arm64
    asset_id=504323523
    digest=e45e520892460732a4bdf37255576415d4a53ec171f8b913faf15bb1aef7cb77
    ;;
  1.8.1:Darwin:x86_64)
    target=darwin_amd64
    asset_id=504323524
    digest=88126981225e7714fcc6b7a10cdba4a80ae5901e9740a8c06d0d5195c8bc294c
    ;;
  *)
    printf 'No pinned MCP publisher archive for %s/%s/%s\n' "$version" "$platform" "$architecture" >&2
    exit 1
    ;;
esac
temporary=$(mktemp -d)
trap 'rm -rf "$temporary"' EXIT
if ! curl --fail --location --silent --show-error --retry 3 --connect-timeout 10 --max-time 120 \
  "https://github.com/modelcontextprotocol/registry/releases/download/v$version/mcp-publisher_$target.tar.gz" \
  --output "$temporary/publisher.tar.gz"; then
  printf 'Direct release download failed; trying pinned GitHub asset %s\n' "$asset_id" >&2
  rm -f "$temporary/publisher.tar.gz"
  curl --fail --location --silent --show-error --retry 3 --connect-timeout 10 --max-time 120 \
    --header 'Accept: application/octet-stream' \
    --header 'X-GitHub-Api-Version: 2022-11-28' \
    "https://api.github.com/repos/modelcontextprotocol/registry/releases/assets/$asset_id" \
    --output "$temporary/publisher.tar.gz"
fi
printf '%s  %s\n' "$digest" "$temporary/publisher.tar.gz" | shasum -a 256 --check --status
tar -xzf "$temporary/publisher.tar.gz" -C "$temporary" mcp-publisher
mkdir -p bin
install -m 755 "$temporary/mcp-publisher" bin/mcp-publisher
