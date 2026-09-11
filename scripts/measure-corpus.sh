#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
root=$PWD

# The historical measurement runs the offline pipeline over acquired shards:
# one dataset plan, pinned shards, per-shard extraction and verification,
# label-free findings under one policy, and the E1 tables. Every step writes a
# file that the next step reads, so a reviewer can rerun any of them.
# Shards live under one directory per cohort: acquisition/<cohort>/shards.
# The dataset root is the acquisition directory, so shard paths stay relative
# and every cohort enters one global plan.
acquisition=artifacts/acquisition
work=${UNSWELL_ACQUISITION_WORK:-$HOME/.cache/unswell/acquisition/work}
output=artifacts/measurement
policy=research/acquisition/policy-e1.yaml
classes=research/methods/rule-classes-v1.json
seed=
only=()
resume=0

usage() {
  printf 'Usage: bash scripts/measure-corpus.sh [--acquisition DIR] [--work DIR] [--output DIR] [--policy FILE] [--classes FILE] [--only SHARD]... [--resume]\n' >&2
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --acquisition)
      acquisition=$2
      shift 2
      ;;
    --work)
      work=$2
      shift 2
      ;;
    --output)
      output=$2
      shift 2
      ;;
    --policy)
      policy=$2
      shift 2
      ;;
    --classes)
      classes=$2
      shift 2
      ;;
    --only)
      only+=("$2")
      shift 2
      ;;
    --resume)
      resume=1
      shift
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

command -v jq >/dev/null || {
  printf 'jq is required.\n' >&2
  exit 1
}

digest() {
  if command -v sha256sum >/dev/null; then
    sha256sum "$1" | cut -d' ' -f1
  else
    shasum -a 256 "$1" | cut -d' ' -f1
  fi
}

corpus_tool=$root/artifacts/corpus/corpus
# A full run starts from empty per-shard outputs so that a stale artifact of
# an earlier shard layout cannot enter the tables; a limited rerun or a
# resumed run keeps them and skips shards that already have findings.
if [[ ${#only[@]} -eq 0 && "$resume" == 0 ]]; then
  rm -rf "$output/plans" "$output/candidates" "$output/verifications" "$output/findings"
fi
mkdir -p "$root/artifacts/corpus" "$output/plans" "$output/candidates" "$output/verifications" "$output/findings"
(cd research/annotation && CGO_ENABLED=0 go build -o "$corpus_tool" ./cmd/corpus)

# The dataset manifest lists every shard by digest and repeats the header of
# the first shard; the plan then checks that every shard repeats it.
shard_root=$acquisition
shard_files=("$acquisition"/*/shards/*.json)
if [[ ! -f "${shard_files[0]}" ]]; then
  printf 'no shard manifests under %s\n' "$acquisition" >&2
  exit 1
fi
first=${shard_files[0]}
seed=$(jq -r '.seed' "$first")
classes_digest=$(digest "$classes")
entries='[]'
for shard in "${shard_files[@]}"; do
  relative=${shard#"$acquisition"/}
  sum=$(digest "$shard")
  size=$(wc -c <"$shard")
  size=${size// /}
  entries=$(jq --arg p "$relative" --arg s "$sum" --argjson b "$size" '. + [{path: $p, sha256: $s, bytes: $b}]' <<<"$entries")
done
jq --arg seed "$seed" --arg classes "$classes_digest" --argjson shards "$entries" \
  '{version: "unswell-corpus-dataset-v1", id: "historical-pilot", seed: $seed, weights: .weights,
    extraction_policy: .extraction_policy, unit_kinds: .unit_kinds, rule_classes_sha256: $classes, shards: $shards}' \
  "$first" >"$output/dataset.json"
"$corpus_tool" dataset plan --root "$shard_root" <"$output/dataset.json" >"$output/dataset-plan.json"
rm -rf "$output/pinned"
mkdir -p "$output/pinned"
"$corpus_tool" dataset pin --root "$shard_root" --output "$output/pinned" <"$output/dataset-plan.json" >"$output/pinned.json"
"$corpus_tool" dataset verify --root "$shard_root" --pinned "$output/pinned" <"$output/dataset-plan.json" >"$output/dataset-verification.json"
groups=$(jq '.groups | length' "$output/dataset-plan.json")
sources=$(jq '.sources | length' "$output/dataset-plan.json")
printf 'dataset: %s groups, %s sources\n' "$groups" "$sources"

# A shard whose extraction or measurement fails twice is skipped and written
# to the failed-shards log with the error, so the tables say what they omit.
# A parse timeout under load is the usual cause; the retry covers it.
failed_log=$output/failed-shards.json
if [[ (${#only[@]} -eq 0 && "$resume" == 0) || ! -f "$failed_log" ]]; then
  printf '[]' >"$failed_log"
fi
record_failure() {
  local name=$1 step=$2 message=$3
  jq --arg name "$name" --arg step "$step" --arg message "$message" \
    '[.[] | select(.shard != $name)] + [{"shard": $name, "step": $step, "error": $message}]' "$failed_log" >"$failed_log.tmp"
  mv "$failed_log.tmp" "$failed_log"
}
# A limited rerun that succeeds clears the shard's earlier failure entry.
clear_failure() {
  jq --arg name "$1" '[.[] | select(.shard != $name)]' "$failed_log" >"$failed_log.tmp"
  mv "$failed_log.tmp" "$failed_log"
}

# run_step reads the input file and writes the output file afresh on each of
# two attempts, sets step_ok to 1 on success, and records the first failure
# message otherwise; it never fails the script itself.
run_step() {
  local name=$1 step=$2 input=$3 result=$4 log message
  shift 4
  step_ok=0
  log=$output/results-$name-$step.log
  for _ in 1 2; do
    if "$@" <"$input" >"$result" 2>"$log.attempt"; then
      step_ok=1
      rm -f "$log.attempt"
      return 0
    fi
    if [[ ! -f "$log" ]]; then
      mv "$log.attempt" "$log"
    fi
  done
  rm -f "$log.attempt" "$result"
  message=$(tail -n 1 "$log")
  record_failure "$name" "$step" "$message"
  printf 'skipped %s at %s: %s\n' "$name" "$step" "$message"
}

step_ok=0
for pinned in "$output/pinned"/*/shards/*.json; do
  name=$(basename "$pinned" .json)
  if [[ ${#only[@]} -gt 0 ]]; then
    selected=0
    for wanted in "${only[@]}"; do
      if [[ "$wanted" == "$name" ]]; then selected=1; fi
    done
    if [[ "$selected" == 0 ]]; then continue; fi
  fi
  if [[ "$resume" == 1 && -s "$output/findings/$name.json" ]]; then
    continue
  fi
  rm -f "$output/results-$name-"*.log
  # Every source of a shard comes from one repository, whose checkout lives
  # under the work directory by cohort and slug.
  repository=$(jq -r '.sources[0].repository' "$pinned")
  shard_cohort=$(jq -r '.sources[0].snapshot.cohort' "$pinned")
  checkout=$work/$shard_cohort/${repository//\//__}
  if [[ ! -d "$checkout" ]]; then
    printf 'missing checkout for %s\n' "$name" >&2
    exit 1
  fi
  "$corpus_tool" plan <"$pinned" >"$output/plans/$name.json"
  rm -f "$output/findings/$name.json"
  run_step "$name" extract "$output/plans/$name.json" "$output/candidates/$name.json" "$corpus_tool" extract --root "$checkout"
  if [[ "$step_ok" == 0 ]]; then continue; fi
  run_step "$name" verify "$output/candidates/$name.json" "$output/verifications/$name.json" "$corpus_tool" verify --root "$checkout"
  if [[ "$step_ok" == 0 ]]; then continue; fi
  run_step "$name" measure "$output/candidates/$name.json" "$output/findings/$name.json" "$corpus_tool" measure --root "$checkout" --policy "$policy"
  if [[ "$step_ok" == 0 ]]; then continue; fi
  clear_failure "$name"
  units=$(jq '.units | length' "$output/candidates/$name.json")
  found=$(jq '[.documents[].findings] | add // 0' "$output/findings/$name.json")
  printf 'measured %s: %s units, %s findings\n' "$name" "$units" "$found"
done
skipped=$(jq 'length' "$failed_log")
printf 'skipped shards: %s (see %s)\n' "$skipped" "$failed_log"
# The tables read every finding artifact present, including those of earlier
# runs that a limited rerun did not repeat.
findings_args=()
for findings in "$output/findings"/*.json; do
  findings_args+=(--findings "$findings")
done
"$corpus_tool" analyze --classes "$classes" "${findings_args[@]}" >"$output/tables.json"
printf 'tables: %s\n' "$output/tables.json"

# The unit analyses count paragraphs and sentences, not documents. The
# period order is: the placebo boundaries of 2012 and 2016, the H1 boundary
# of 2018, the H0 baseline, the contemporary snapshots, then natural and
# controlled text. Each consecutive pair gets a first-appearance filter. It
# drops the later cohort's units whose text the earlier snapshot of the same
# repository already holds. The filtered tables then contrast the pair. Only
# shards with findings enter a filter or selection: the analysis checks that
# every listed unit was measured.
present=()
for cohort in historical-2012 historical-2016 historical-2018 historical contemporary natural controlled; do
  if [[ -d "$output/pinned/$cohort" ]]; then present+=("$cohort"); fi
done
candidate_args() {
  local side=$1 cohort=$2 pinned name
  for pinned in "$output/pinned/$cohort"/shards/*.json; do
    name=$(basename "$pinned" .json)
    if [[ -s "$output/candidates/$name.json" && -s "$output/findings/$name.json" ]]; then
      printf -- '--%s\n%s\n' "$side" "$output/candidates/$name.json"
    fi
  done
}
for kind in paragraph sentence; do
  "$corpus_tool" analyze --classes "$classes" "${findings_args[@]}" --unit-kind "$kind" >"$output/tables-$kind.json"
  printf 'tables: %s\n' "$output/tables-$kind.json"
  for ((i = 1; i < ${#present[@]}; i++)); do
    earlier=${present[i - 1]}
    later=${present[i]}
    appearance_args=()
    candidate_args earlier "$earlier" >"$output/appearance-args.txt"
    candidate_args later "$later" >>"$output/appearance-args.txt"
    while IFS= read -r line; do appearance_args+=("$line"); done <"$output/appearance-args.txt"
    rm -f "$output/appearance-args.txt"
    filter=$output/first-appearance-$later-vs-$earlier-$kind.json
    "$corpus_tool" first-appearance --unit-kind "$kind" "${appearance_args[@]}" >"$filter"
    "$corpus_tool" analyze --classes "$classes" "${findings_args[@]}" --unit-kind "$kind" --baseline "$earlier" \
      --first-appearance "$filter" >"$output/tables-first-appearance-$later-vs-$earlier-$kind.json"
    new_units=$(jq '.new_units' "$filter")
    later_units=$(jq '.later_units' "$filter")
    printf 'first appearance: %s new of %s %s units in %s against %s; tables: %s\n' "$new_units" "$later_units" "$kind" \
      "$later" "$earlier" "$output/tables-first-appearance-$later-vs-$earlier-$kind.json"
  done
  # The unit selection counts each unit text once across every cohort in
  # period order, whatever repository a copy sits in.
  order=$(
    IFS=,
    printf '%s' "${present[*]}"
  )
  selection_args=()
  for cohort in "${present[@]}"; do
    candidate_args candidates "$cohort" >>"$output/selection-args.txt"
  done
  while IFS= read -r line; do selection_args+=("$line"); done <"$output/selection-args.txt"
  rm -f "$output/selection-args.txt"
  selection=$output/unit-selection-$kind.json
  "$corpus_tool" dedupe --unit-kind "$kind" --order "$order" "${selection_args[@]}" >"$selection"
  "$corpus_tool" analyze --classes "$classes" "${findings_args[@]}" --unit-kind "$kind" --selection "$selection" \
    >"$output/tables-unique-$kind.json"
  kept_units=$(jq '[.cohorts[].kept] | add' "$selection")
  all_units=$(jq '[.cohorts[].units] | add' "$selection")
  printf 'unique: %s kept of %s %s units; tables: %s\n' "$kept_units" "$all_units" "$kind" "$output/tables-unique-$kind.json"
done
