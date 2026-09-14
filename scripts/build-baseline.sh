#!/usr/bin/env bash
# Build one construction baseline table: the key rates of a cohort by role,
# the artifact "corpus propose" reads when it measures an outside tree. The
# input is a measurement directory, so the table is a function of the pinned
# shards and the extraction that produced their candidates. Rerunning this
# after an extraction change is what keeps the baseline and the engine in
# step.
set -euo pipefail
script_directory=$(dirname "$0")
cd "$script_directory/.."
root=$PWD

measurement=artifacts/measurement
cohort=historical
measure=closed-3gram
min_count=2
source_note="the historical cohort of the pattern protocol's corpus, by role"
output=

usage() {
  printf 'Usage: bash scripts/build-baseline.sh --output FILE [--measurement DIR] [--cohort NAME] [--measure NAME] [--min-count N] [--source TEXT]\n' >&2
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --measurement)
      measurement=$2
      shift 2
      ;;
    --cohort)
      cohort=$2
      shift 2
      ;;
    --measure)
      measure=$2
      shift 2
      ;;
    --min-count)
      min_count=$2
      shift 2
      ;;
    --source)
      source_note=$2
      shift 2
      ;;
    --output)
      output=$2
      shift 2
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

if [[ -z "$output" ]]; then
  usage
  exit 2
fi
command -v jq >/dev/null || {
  printf 'jq is required.\n' >&2
  exit 1
}

corpus_tool=$root/artifacts/corpus/corpus
mkdir -p "$root/artifacts/corpus"
(cd research/annotation && CGO_ENABLED=0 go build -o "$corpus_tool" ./cmd/corpus)

# The cohort's own pinned shards name the candidate artifacts to count. A
# shard that failed to measure has no candidate file and is left out; the
# count of what entered is printed so a partial measurement cannot pass for a
# whole one.
candidate_args=()
shards=0
for pinned in "$measurement/pinned/$cohort"/shards/*.json; do
  shards=$((shards + 1))
  candidate=$measurement/candidates/$(basename "$pinned")
  if [[ -s "$candidate" ]]; then candidate_args+=(--candidates "$candidate"); fi
done
counted=$((${#candidate_args[@]} / 2))
if [[ "$counted" -eq 0 ]]; then
  printf 'no candidate artifact under %s for cohort %s\n' "$measurement" "$cohort" >&2
  exit 1
fi
printf 'counting %s of %s %s shards\n' "$counted" "$shards" "$cohort"

work_file=$(mktemp)
trap 'rm -f "$work_file"' EXIT
"$corpus_tool" frequencies --baselines --baseline "$cohort" --min-count "$min_count" \
  "${candidate_args[@]}" >"$work_file"

# The tables carry every measure and every contrast; a baseline artifact
# carries one measure and its rates. The reduction lives beside this script.
jq -S --arg source "$source_note" --arg measure "$measure" \
  -f "$root/scripts/build-baseline.jq" "$work_file" >"$output"

roles=$(jq '.baselines | length' "$output")
terms=$(jq '[.baselines[].terms | length] | add' "$output")
words=$(jq '[.baselines[].words] | add' "$output")
printf 'baseline: %s roles, %s terms, %s words; %s\n' "$roles" "$terms" "$words" "$output"
