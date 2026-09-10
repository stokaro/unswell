#!/usr/bin/env bash
# Build and test the public API consumer against a published module version.
# The consumer sources come from the same tag by default, because a later
# consumer may use packages a published module does not contain yet.
set -euo pipefail
script_directory=$(dirname "$0")
cd "$script_directory/.."
root=$PWD

version=""
consumer_ref=""
output=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      version=$2
      shift 2
      ;;
    --consumer-ref)
      consumer_ref=$2
      shift 2
      ;;
    --output)
      output=$2
      shift 2
      ;;
    *)
      printf 'Usage: %s [--version TAG] [--consumer-ref REF] [--output FILE]\n' "$0" >&2
      exit 2
      ;;
  esac
done
if [[ -z "$version" ]]; then
  version=$(gh release view --repo stokaro/unswell --json tagName --jq .tagName)
fi
if [[ -z "$consumer_ref" ]]; then
  consumer_ref=$version
fi

workspace=$(mktemp -d)
trap 'rm -rf "$workspace"' EXIT
consumer=$workspace/consumer
mkdir -p "$consumer"
git archive "$consumer_ref" examples/consumer | tar -x -C "$workspace"
cp "$workspace"/examples/consumer/* "$consumer/"

# The published module must satisfy the consumer with no local replacement.
# The go/analysis adapter is a separate module and keeps its own resolution.
(
  cd "$consumer"
  sed -i.bak '/^replace github.com\/stokaro\/unswell =>/d' go.mod
  rm -f go.mod.bak
  go mod edit -require="github.com/stokaro/unswell@$version"
  go mod edit -replace="github.com/stokaro/unswell/goanalysis=$root/goanalysis"
  GOFLAGS=-mod=mod go mod tidy >"$workspace/tidy.log" 2>&1
  go test ./... >"$workspace/test.log" 2>&1
)

module_version=$(cd "$consumer" && go list -m github.com/stokaro/unswell)
module_sum=$(cd "$consumer" && go list -m -f '{{.Sum}}' github.com/stokaro/unswell)
passed=$(grep -c '^ok' "$workspace/test.log" || true)

# Report which packages the current working tree's consumer needs that the
# published module does not contain. That is a fact about the next release.
newer=$(python3 scripts/published-module-gap.py "$root" "$consumer")
compiler=$(go version)

report=$(
  cat <<JSON
{
  "format": "unswell-published-module-check-v1",
  "requested_version": "$version",
  "consumer_ref": "$consumer_ref",
  "resolved_module": "$module_version",
  "module_sum": "$module_sum",
  "compiler": "$compiler",
  "gotoolchain": "${GOTOOLCHAIN:-runtime default}",
  "passing_packages": $passed,
  "packages_added_since_release": $newer,
  "scope": "One consumer module built against a published version. Unreleased changes are not covered."
}
JSON
)
if [[ -n "$output" ]]; then
  printf '%s\n' "$report" >"$output"
fi
printf '%s\n' "$report"
printf 'The published module %s satisfied the public API consumer from %s.\n' "$module_version" "$consumer_ref" >&2
