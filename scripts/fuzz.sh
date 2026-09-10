#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

# Active fuzzing budget. Every discovered target runs for this duration with this
# many workers; docs/validation.md records the campaign used for acceptance.
fuzztime=10s
workers=2
mode=run
report=

usage() {
  printf 'Usage: %s [--fuzztime DURATION] [--parallel WORKERS] [--report FILE] [--list] [--self-test]\n' "$0" >&2
}

# Pairs every discovered fuzz target with its package. `go test -list` is the
# authority, so a new Fuzz function cannot be missed by a hand-written list.
discover_targets() {
  local directory=$1 packages=${2:-./...}
  (cd "$directory" && go test -list '^Fuzz' "$packages") | awk '
    /^Fuzz[A-Za-z0-9_]*$/ { names[++count] = $0; next }
    $1 == "ok" || $1 == "?" || $1 == "FAIL" {
      for (position = 1; position <= count; position++) print $2, names[position]
      count = 0
    }
  '
}

fuzz_module() {
  local directory=$1 targets package name started
  targets=$(discover_targets "$directory")
  [[ -n "$targets" ]] || return 0
  while read -r package name; do
    printf '== %s %s (fuzztime=%s workers=%s)\n' "$package" "$name" "$fuzztime" "$workers"
    started=$SECONDS
    (cd "$directory" && go test "$package" -run '^$' -fuzz "^${name}\$" -fuzztime "$fuzztime" -parallel "$workers") || return $?
    printf '%s %s %s\n' "$package" "$name" "$((SECONDS - started))" >>"$measurements"
  done <<<"$targets"
}

write_report() {
  local destination=$1
  python3 - "$destination" "$fuzztime" "$workers" "$measurements" <<'PYTHON'
import json
import sys

destination, fuzztime, workers, measurements = sys.argv[1:5]
targets = []
with open(measurements, encoding="utf-8") as handle:
    for line in handle:
        package, target, seconds = line.split()
        targets.append({"package": package, "target": target, "seconds": int(seconds), "outcome": "pass"})
with open(destination, "w", encoding="utf-8") as handle:
    json.dump({"fuzztime": fuzztime, "workers": int(workers), "targets": targets}, handle, indent=2, sort_keys=True)
    handle.write("\n")
PYTHON
}

modules() {
  local directory role
  while read -r directory role; do
    [[ "$role" == runtime || "$role" == consumer ]] || continue
    printf '%s\n' "$directory"
  done <.gomodules
}

self_test() {
  local status discovered
  fuzz_self_test_directory=$(mktemp -d)
  local temporary="$fuzz_self_test_directory"
  trap 'rm -rf "$fuzz_self_test_directory"' EXIT
  cat >"$temporary/go.mod" <<'MODULE'
module fuzzselftest

go 1.25.0
MODULE
  cat >"$temporary/subject.go" <<'SOURCE'
package fuzzselftest

// Accept reports whether the input avoids the rejected marker.
func Accept(input string) bool { return input != "reject" }
SOURCE
  cat >"$temporary/subject_test.go" <<'SOURCE'
package fuzzselftest_test

import (
	"testing"

	"fuzzselftest"
)

func FuzzAccepted(f *testing.F) {
	f.Add("value")
	f.Fuzz(func(t *testing.T, input string) {
		_ = fuzzselftest.Accept(input)
	})
}

func FuzzRejected(f *testing.F) {
	f.Add("reject")
	f.Fuzz(func(t *testing.T, input string) {
		if !fuzzselftest.Accept(input) {
			t.Fatalf("rejected input reached the target: %q", input)
		}
	})
}
SOURCE
  discovered=$(GOWORK=off discover_targets "$temporary" ./... | awk '{ print $2 }' | sort | tr '\n' ' ')
  [[ "$discovered" == "FuzzAccepted FuzzRejected " ]] || {
    printf 'Discovery missed a fuzz target: %s\n' "$discovered" >&2
    return 1
  }
  measurements="$temporary/measurements"
  status=0
  fuzztime=5s
  # fuzz_module reports the first failing target through its exit status.
  # shellcheck disable=SC2310
  GOWORK=off fuzz_module "$temporary" >/dev/null 2>&1 || status=$?
  [[ "$status" != 0 ]] || {
    printf 'A failing fuzz target did not fail the run\n' >&2
    return 1
  }
  printf 'Fuzz discovery and failure propagation verified\n'
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --fuzztime)
      fuzztime=${2:?duration required}
      shift 2
      ;;
    --parallel)
      workers=${2:?worker count required}
      shift 2
      ;;
    --list)
      mode=list
      shift
      ;;
    --report)
      report=${2:?file required}
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
  list)
    directories=$(modules)
    while read -r directory; do
      discover_targets "$directory"
    done <<<"$directories"
    ;;
  run)
    directories=$(modules)
    measurements=$(mktemp)
    trap 'rm -f "$measurements"' EXIT
    while read -r directory; do
      fuzz_module "$directory"
    done <<<"$directories"
    [[ -z "$report" ]] || write_report "$report"
    ;;
esac
