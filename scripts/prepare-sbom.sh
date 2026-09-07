#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

check_sbom() {
  jq --exit-status '
    .bomFormat == "CycloneDX" and
    .metadata.component.name == "github.com/stokaro/unswell" and
    (.components | type) == "array" and
    .metadata.component.licenses == [{license: {id: "MIT"}}] and
    (.metadata.component.evidence.licenses // []) == []
  ' "$1" >/dev/null
}

if [[ "${1:-}" == --self-test && $# == 1 ]]; then
  temporary=$(mktemp -d)
  trap 'rm -rf "$temporary"' EXIT
  jq -n '{bomFormat: "CycloneDX", metadata: {component: {
    name: "github.com/stokaro/unswell", evidence: {licenses: [{license: {id: "Apache-2.0"}}]}
  }}, components: [{name: "dependency", evidence: {licenses: [{license: {id: "Apache-2.0"}}]}}]}' >"$temporary/raw.json"
  bash "$0" "$temporary/raw.json" "$temporary/final.json"
  check_sbom "$temporary/final.json"
  jq '.components' "$temporary/raw.json" >"$temporary/dependencies-before.json"
  jq '.components' "$temporary/final.json" >"$temporary/dependencies-after.json"
  cmp "$temporary/dependencies-before.json" "$temporary/dependencies-after.json"
  for mutation in wrong missing; do
    if [[ "$mutation" == wrong ]]; then
      jq '.metadata.component.licenses = [{license: {id: "Apache-2.0"}}]' "$temporary/final.json" >"$temporary/bad.json"
    else
      jq 'del(.metadata.component.licenses)' "$temporary/final.json" >"$temporary/bad.json"
    fi
    if bash "$0" --check "$temporary/bad.json"; then
      printf 'SBOM gate accepted a %s project license\n' "$mutation" >&2
      exit 1
    fi
  done
  jq '.metadata.component.name = "another/project"' "$temporary/raw.json" >"$temporary/another.json"
  if bash "$0" "$temporary/another.json" "$temporary/invalid.json" 2>/dev/null; then
    printf 'SBOM preparation accepted another project\n' >&2
    exit 1
  fi
  cp "$temporary/final.json" "$temporary/preserved.json"
  if bash "$0" "$temporary/raw.json" "$temporary/final.json" 2>/dev/null; then
    printf 'SBOM preparation overwrote existing metadata\n' >&2
    exit 1
  fi
  cmp "$temporary/final.json" "$temporary/preserved.json"
  printf 'SBOM checks preserved dependency evidence and rejected missing or wrong licenses, another project, and overwrite.\n'
  exit 0
fi

if [[ "${1:-}" == --check && $# == 2 ]]; then
  check_sbom "$2"
  exit 0
fi
if [[ $# != 2 ]]; then
  printf 'Usage: %s INPUT OUTPUT | --check FILE | --self-test\n' "$0" >&2
  exit 2
fi
if [[ -e "$2" ]]; then
  printf 'SBOM output already exists: %s\n' "$2" >&2
  exit 1
fi
temporary=$(mktemp)
trap 'rm -f "$temporary"' EXIT
# The root LICENSE declares MIT. Bundled dependency notices are not its evidence.
jq --exit-status '
  if .bomFormat != "CycloneDX" or .metadata.component.name != "github.com/stokaro/unswell" or
    (.components | type) != "array" then error("Unexpected SBOM project or format") else . end |
  .metadata.component.licenses = [{license: {id: "MIT"}}] |
  del(.metadata.component.evidence.licenses) |
  if .metadata.component.evidence == {} then del(.metadata.component.evidence) else . end
' "$1" >"$temporary"
check_sbom "$temporary"
(
  set -o noclobber
  cat "$temporary" >"$2"
)
