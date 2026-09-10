#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
root=$PWD

output=artifacts/coverage
mode=run

usage() {
  printf 'Usage: %s [--output DIRECTORY] [--self-test]\n' "$0" >&2
}

# One instrumented run per module replaces the plain test run; nothing is
# executed twice to collect coverage.
measure() {
  local destination=$1 directory role slug
  rm -rf "$destination"
  mkdir -p "$destination"
  while read -r directory role; do
    [[ "$role" == runtime || "$role" == consumer ]] || continue
    slug=${directory//\//-}
    [[ "$slug" != "." ]] || slug=root
    (cd "$directory" && go test -count=1 -coverprofile="$root/$destination/${slug}.out" -coverpkg=./... ./...)
  done <.gomodules
}

report() {
  local destination=$1 commit version platform
  commit=$(git rev-parse HEAD)
  version=$(go env GOVERSION)
  platform=$(go env GOOS GOARCH)
  UNSWELL_COVERAGE_COMMIT=$commit UNSWELL_COVERAGE_GO=$version \
    UNSWELL_COVERAGE_GOOS=${platform%%$'\n'*} UNSWELL_COVERAGE_GOARCH=${platform##*$'\n'} \
    python3 scripts/coverage-report.py "$destination"/*.out
}

self_test() {
  local status
  coverage_self_test_directory=$(mktemp -d)
  local temporary="$coverage_self_test_directory"
  trap 'rm -rf "$coverage_self_test_directory"' EXIT
  python3 - "$temporary" <<'PYTHON'
import sys
from pathlib import Path

module = "github.com/stokaro/unswell"
areas = [
    module,
    f"{module}/rule",
    f"{module}/ruleset",
    f"{module}/builtin",
    f"{module}/config",
    f"{module}/internal/appconfig",
    f"{module}/extract",
    f"{module}/document",
    f"{module}/internal/mapping",
]


def profile(covered_blocks):
    lines = ["mode: set"]
    for package in areas:
        for block in range(10):
            count = 1 if block < covered_blocks else 0
            lines.append(f"{package}/file.go:{block + 1}.1,{block + 1}.2 1 {count}")
    return "\n".join(lines) + "\n"


directory = Path(sys.argv[1])
(directory / "passing.out").write_text(profile(10))
(directory / "failing.out").write_text(profile(8))
PYTHON
  python3 scripts/coverage-report.py "$temporary/passing.out" >/dev/null
  status=0
  python3 scripts/coverage-report.py "$temporary/failing.out" >/dev/null || status=$?
  [[ "$status" != 0 ]] || {
    printf 'The coverage gate accepted an area below its target\n' >&2
    return 1
  }
  printf 'Coverage gate rejects areas below their targets\n'
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --output)
      output=${2:?directory required}
      shift 2
      ;;
    --self-test)
      mode=self-test
      shift
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

case "$mode" in
  self-test) self_test ;;
  run)
    measure "$output"
    report "$output" | tee "$output/summary.json"
    ;;
esac
