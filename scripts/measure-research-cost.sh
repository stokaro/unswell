#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
root=$PWD

# The research pipeline runs as separate stages, and a single total would hide
# which stage a method makes expensive. Each stage is measured on its own.
label=local
mode=run
output=

usage() {
  printf 'Usage: %s [--label NAME] [--output FILE] [--self-test]\n' "$0" >&2
}

# resource_flag selects the reporting form of the available time tool.
resource_flag() {
  if /usr/bin/time -v true >/dev/null 2>&1; then
    printf '%s' -v
    return 0
  fi
  if /usr/bin/time -l true >/dev/null 2>&1; then
    printf '%s' -l
    return 0
  fi
  printf 'No /usr/bin/time with resource reporting is available.\n' >&2
  return 2
}

# peak_bytes reads the maximum resident set from either time implementation.
# GNU reports kilobytes; the BSD tool reports bytes.
peak_bytes() {
  local stamp=$1 kilobytes bytes
  kilobytes=$(sed -n 's/.*Maximum resident set size (kbytes): *\([0-9][0-9]*\).*/\1/p' "$stamp" | tail -1)
  if [[ -n "$kilobytes" ]]; then
    printf '%s' "$((kilobytes * 1024))"
    return 0
  fi
  bytes=$(sed -n 's/^ *\([0-9][0-9]*\) *maximum resident set size.*/\1/p' "$stamp" | tail -1)
  printf '%s' "${bytes:-0}"
}

# stage runs one pipeline step under the time tool and appends its record. The
# expected exit code is explicit, so a stage that starts failing is a finding
# rather than a faster measurement.
stage() {
  local name=$1 expected=$2 input=$3 result=$4
  shift 4
  local stamp=$measurements/$name.time started ended seconds status=0 peak
  started=$(date +%s.%N)
  /usr/bin/time "$flag" -o "$stamp" "$corpus_tool" "$@" <"$input" >"$result" 2>"$measurements/$name.err" || status=$?
  ended=$(date +%s.%N)
  if [[ "$status" != "$expected" ]]; then
    printf 'Stage %s exited %s, expected %s:\n' "$name" "$status" "$expected" >&2
    tail -5 "$measurements/$name.err" >&2
    return 1
  fi
  seconds=$(awk -v a="$started" -v b="$ended" 'BEGIN { printf "%.3f", b - a }')
  peak=$(peak_bytes "$stamp")
  printf '%s %s %s %s\n' "$name" "$seconds" "$peak" "$status" >>"$measurements/records"
}

# prepare writes the frozen prediction plan the measured stages consume.
prepare() {
  python3 - "$workspace" "$fixture" <<'PYTHON'
import hashlib
import json
import sys

workspace, fixture = sys.argv[1:3]
corpus = json.load(open(f"{workspace}/corpus.json", encoding="utf-8"))
model = json.load(open(f"{workspace}/model.json", encoding="utf-8"))
protocol = b"Cost measurement of the research pipeline. No scientific qualification.\n"
with open(f"{workspace}/protocol.md", "wb") as handle:
    handle.write(protocol)
plan = {"version": "unswell-research-predictions-v2", "id": "cost-measurement-v1",
        "protocol_sha256": hashlib.sha256(protocol).hexdigest(), "model_sha256": model["sha256"],
        "corpus_sha256": corpus["sha256"], "partition": "final_test", "context": "prepared_piece",
        "response": model["options"]["estimator"], "threshold": 0.5}
with open(f"{workspace}/plan.json", "w", encoding="utf-8") as handle:
    json.dump(plan, handle)
PYTHON
}

measure_pipeline() {
  local sources=$fixture/sources round=$fixture/round.json
  : >"$measurements/records"
  # Startup isolates process creation and package loading from any real work.
  stage startup 2 /dev/null "$workspace/usage.json"
  stage plan 0 "$fixture/manifest.json" "$workspace/plan-out.json" plan
  cp "$workspace/plan-out.json" "$workspace/plan-in.json"
  stage extract 0 "$workspace/plan-in.json" "$workspace/corpus.json" extract --root "$sources"
  stage verify 0 "$workspace/corpus.json" "$workspace/verified.json" verify --root "$sources"
  stage train 0 "$workspace/corpus.json" "$workspace/model.json" train --root "$sources" --round "$round" \
    --kind paragraph --feature prose-words --calibration isotonic --allow-simulation
  prepare
  stage predict 0 "$workspace/corpus.json" "$workspace/predictions.json" predict --root "$sources" \
    --model "$workspace/model.json" --plan "$workspace/plan.json" --protocol "$workspace/protocol.md"
  stage evaluate 0 "$workspace/predictions.json" "$workspace/evaluation.json" evaluate \
    --corpus "$workspace/corpus.json" --round "$round" --allow-simulation
  stage figures 0 "$workspace/evaluation.json" "$workspace/reliability.svg" figures --plot reliability
}

write_report() {
  local destination=$1 commit version system machine cpus
  commit=$(git rev-parse HEAD)
  version=$(go env GOVERSION)
  system=$(uname -s)
  machine=$(uname -m)
  cpus=$(getconf _NPROCESSORS_ONLN 2>/dev/null || printf 0)
  python3 - "$destination" "$label" "$commit" "$version" "$system" "$machine" "$cpus" \
    "$measurements/records" "$workspace/corpus.json" <<'PYTHON'
import json
import sys

destination, label, commit, version, system, machine, cpus, records, corpus = sys.argv[1:10]
stages = []
with open(records, encoding="utf-8") as handle:
    for line in handle:
        name, seconds, peak, status = line.split()
        stages.append({"stage": name, "seconds": float(seconds),
                       "maximum_resident_bytes": int(peak), "exit_code": int(status)})
artifact = json.load(open(corpus, encoding="utf-8"))
report = {
    "format": "unswell-research-cost-observation-v1",
    "scope": "Tutorial fixture on this host. Measures pipeline cost, not model quality.",
    "label": label, "commit": commit, "go_version": version,
    "goos": system, "goarch": machine, "logical_cpus": int(cpus), "cgo_enabled": False,
    "units": len(artifact.get("units", [])), "sources": len(artifact.get("sources", [])),
    "stages": stages,
    "total_seconds": round(sum(entry["seconds"] for entry in stages), 3),
    "peak_resident_bytes": max(entry["maximum_resident_bytes"] for entry in stages),
}
text = json.dumps(report, indent=2, sort_keys=True) + "\n"
if destination == "-":
    sys.stdout.write(text)
else:
    with open(destination, "w", encoding="utf-8") as handle:
        handle.write(text)
PYTHON
}

self_test() {
  local status=0 records
  measure_pipeline
  records=$(wc -l <"$measurements/records")
  [[ "${records// /}" == "8" ]] || {
    printf 'Expected eight measured stages, recorded %s\n' "$records" >&2
    return 1
  }
  awk '{ if ($2 + 0 < 0 || $3 + 0 <= 0) exit 1 }' "$measurements/records" || {
    printf 'A stage reported no elapsed time or no peak memory\n' >&2
    return 1
  }
  # A stage that fails must fail the harness rather than record a fast run.
  # stage reports the mismatch through its exit status.
  # shellcheck disable=SC2310
  stage refused 0 "$workspace/predictions.json" "$workspace/rejected.json" evaluate \
    --corpus "$workspace/corpus.json" --round "$fixture/round.json" 2>/dev/null || status=$?
  [[ "$status" != 0 ]] || {
    printf 'A refused stage did not fail the harness\n' >&2
    return 1
  }
  printf 'Research cost harness measured every stage and rejects a failing one\n'
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --label)
      label=${2:?name required}
      shift 2
      ;;
    --output)
      output=${2:?file required}
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

flag=$(resource_flag)
workspace=$(mktemp -d)
trap 'rm -rf "$workspace"' EXIT
measurements=$workspace/measurements
mkdir -p "$measurements"
fixture=$root/research/annotation/training/testdata
corpus_tool=$workspace/corpus

(cd research/annotation && CGO_ENABLED=0 go build -trimpath -o "$corpus_tool" ./cmd/corpus)

case "$mode" in
  self-test) self_test ;;
  run)
    measure_pipeline
    write_report "${output:--}"
    ;;
esac
