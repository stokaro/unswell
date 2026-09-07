#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
apidiff=$(cd tools && go tool -n apidiff)
temporary=$(mktemp -d)
trap 'rm -rf "$temporary"' EXIT

compatible() {
  "$apidiff" -m -incompatible "$1" "$2" >"$temporary/diff.txt" || return 2
  if [[ -s "$temporary/diff.txt" ]]; then
    cat "$temporary/diff.txt" >&2 || return 2
    return 1
  fi
  return 0
}

if [[ "${1:-}" == --self-test ]]; then
  mkdir "$temporary/fixture"
  minimum_go=$(awk '/^go / {print $2}' go.mod)
  printf 'module example.com/fixture\n\ngo %s\n' "$minimum_go" >"$temporary/fixture/go.mod"
  printf 'package fixture\nfunc Present() {}\n' >"$temporary/fixture/api.go"
  (cd "$temporary/fixture" && "$apidiff" -m -w "$temporary/before.api" example.com/fixture)
  printf 'package fixture\nfunc Present(value string) {}\n' >"$temporary/fixture/api.go"
  (cd "$temporary/fixture" && "$apidiff" -m -w "$temporary/after.api" example.com/fixture)
  compatible "$temporary/before.api" "$temporary/before.api"
  status=0
  # compatible distinguishes tool failure (2) from incompatibility (1).
  # shellcheck disable=SC2310
  compatible "$temporary/before.api" "$temporary/after.api" 2>/dev/null || status=$?
  if [[ "$status" != 1 ]]; then
    printf 'API gate returned %s for an incompatible signature; expected 1.\n' "$status" >&2
    exit 1
  fi
  status=0
  # A missing snapshot must propagate the tool failure, not look compatible.
  # shellcheck disable=SC2310
  compatible "$temporary/missing.api" "$temporary/after.api" 2>/dev/null || status=$?
  if [[ "$status" != 2 ]]; then
    printf 'API gate returned %s for a missing snapshot; expected 2.\n' "$status" >&2
    exit 1
  fi
  printf 'API gate rejected an incompatible signature change.\n'
  exit 0
fi

shallow=$(git rev-parse --is-shallow-repository)
if [[ "$shallow" == true ]]; then
  printf 'API comparison requires complete tag history\n' >&2
  exit 1
fi
"$apidiff" -m -w "$temporary/current.api" github.com/stokaro/unswell
# The first release has a reviewed bootstrap snapshot. Subsequent CI runs compare
# with the immutable snapshot from the most recent reachable release tag.
baseline=docs/api/alpha.api
previous=$(git describe --tags --abbrev=0 --match 'v*' HEAD^ 2>/dev/null || true)
if [[ -n "$previous" ]]; then
  git show "$previous:docs/api/alpha.api" >"$temporary/released.api"
  baseline="$temporary/released.api"
fi
# compatible checks each command failure explicitly.
# shellcheck disable=SC2310
if ! compatible "$baseline" "$temporary/current.api"; then
  printf 'API comparison failed; resolve the reported error before release.\n' >&2
  exit 1
fi
# Ensure each release carries a current snapshot for the following release.
"$apidiff" -m docs/api/alpha.api "$temporary/current.api" >"$temporary/freshness.txt"
if [[ -s "$temporary/freshness.txt" ]]; then
  cat "$temporary/freshness.txt" >&2
  printf 'The committed API snapshot is stale\n' >&2
  exit 1
fi
