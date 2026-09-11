#!/usr/bin/env bash
# Run the origin-channel experiment of the pattern protocol. The corpus is a
# union of the controlled cohort and the historical shards of the task
# repositories. The fit uses provenance labels. Predictions are frozen on the
# selection and confirmation partitions and scored against the same labels.
# Every step is a corpus command; this script orders them and records
# digests. Nothing here qualifies an origin model.
set -euo pipefail
script_directory=$(dirname "$0")
cd "$script_directory/.."

work=${UNSWELL_ACQUISITION_WORK:-$HOME/.cache/unswell/acquisition/work}
acquisition=artifacts/acquisition
dataset_plan=artifacts/measurement/dataset-plan.json
tasks=research/generation/runs/2026-09-11-pilot/tasks.json
generation=research/generation/runs/2026-09-11-pilot/records.json
protocol=research/methods/llm-patterns-v1.md
output=artifacts/origin
record=""
id=origin-pilot-v1
cap=15
max_source_bytes=102400
threshold=0.5
while [[ $# -gt 0 ]]; do
  case "$1" in
    --work)
      work=$2
      shift 2
      ;;
    --acquisition)
      acquisition=$2
      shift 2
      ;;
    --dataset-plan)
      dataset_plan=$2
      shift 2
      ;;
    --tasks)
      tasks=$2
      shift 2
      ;;
    --generation)
      generation=$2
      shift 2
      ;;
    --protocol)
      protocol=$2
      shift 2
      ;;
    --output)
      output=$2
      shift 2
      ;;
    --record)
      record=$2
      shift 2
      ;;
    --id)
      id=$2
      shift 2
      ;;
    --cap)
      cap=$2
      shift 2
      ;;
    --threshold)
      threshold=$2
      shift 2
      ;;
    --max-source-bytes)
      max_source_bytes=$2
      shift 2
      ;;
    *)
      printf 'Usage: %s OPTIONS\n' "$0" >&2
      printf 'Options: --work DIR, --acquisition DIR, --dataset-plan FILE, --tasks FILE, --generation FILE, --protocol FILE, --output DIR, --record DIR, --id ID, --cap N, --max-source-bytes N, --threshold T\n' >&2
      exit 2
      ;;
  esac
done
if ! [[ "$cap" =~ ^[0-9]+$ ]] || ! [[ "$max_source_bytes" =~ ^[0-9]+$ ]]; then
  printf 'The per-checkout cap and the source byte limit must be non-negative integers.\n' >&2
  exit 2
fi

mkdir -p "$output"
output=$(cd "$output" && pwd)
corpus_tool=$output/corpus
(cd research/annotation && go build -o "$corpus_tool" ./cmd/corpus)
tool_commit=$(git rev-parse HEAD)

# The union takes every controlled source and the comment-role sources of the
# task repositories. A cap per historical checkout keeps the corpus within one
# artifact. Files above the byte limit stay out: one bundle of thousands of
# comments would be most of the corpus and exhaust the measurement budget of
# its engine run. Only paragraphs are extracted.
repository_list=$(jq -r '[.tasks[].repository] | unique | .[]' "$tasks")
repositories=()
while IFS= read -r repository; do
  repositories+=(--repository "$repository")
done <<<"$repository_list"
"$corpus_tool" dataset union --root "$acquisition" --id "$id" --cohort controlled --cohort historical \
  --role comment --unit-kind paragraph --max-per-checkout "$cap" --uncapped-cohort controlled \
  --max-source-bytes "$max_source_bytes" "${repositories[@]}" <"$dataset_plan" >"$output/union.json"
"$corpus_tool" plan <"$output/union.json" >"$output/union-plan.json"
"$corpus_tool" extract --root "$work" <"$output/union-plan.json" >"$output/candidates.json"
"$corpus_tool" verify --root "$work" <"$output/candidates.json" >"$output/verification.json"

# The structure, lexical, part-of-speech, and readability features of the
# shared contract enter the fit. The repetition family stays out: it abstains
# on a paragraph of one sentence, which most comments are, and excluding those
# units would fit the model on the long paragraphs only. A unit that still
# lacks a feature is excluded and counted rather than failing the run.
features=(prose-words counted-characters prose-sentences mean-sentence-words sentence-word-stddev
  shortest-sentence longest-sentence type-token-ratio hapax-token-ratio noun-token-ratio verb-token-ratio
  adjective-token-ratio adverb-token-ratio automated-readability-index)
feature_flags=()
for feature in "${features[@]}"; do
  feature_flags+=(--feature "$feature")
done
"$corpus_tool" train --root "$work" --labels provenance --kind paragraph "${feature_flags[@]}" \
  --missing-features exclude --calibration isotonic <"$output/candidates.json" >"$output/model.json"

sha() { shasum -a 256 "$1" | cut -c1-64; }
model_sha=$(jq -r '.sha256' "$output/model.json")
corpus_sha=$(jq -r '.sha256' "$output/candidates.json")
protocol_sha=$(sha "$protocol")
for partition in development final_test; do
  cat >"$output/plan-$partition.json" <<PLAN
{
  "version": "unswell-research-predictions-v2",
  "id": "$id-$partition",
  "protocol_sha256": "$protocol_sha",
  "model_sha256": "$model_sha",
  "corpus_sha256": "$corpus_sha",
  "partition": "$partition",
  "context": "prepared_piece",
  "response": "isotonic",
  "threshold": $threshold
}
PLAN
  "$corpus_tool" predict --root "$work" --model "$output/model.json" --plan "$output/plan-$partition.json" \
    --protocol "$protocol" <"$output/candidates.json" >"$output/predictions-$partition.json"
  "$corpus_tool" evaluate --corpus "$output/candidates.json" --labels provenance --generation "$generation" \
    <"$output/predictions-$partition.json" >"$output/evaluation-$partition.json"
done

# Digests bind the record to the exact artifacts, including the candidate
# artifact that stays outside the repository.
{
  printf '{\n  "format": "unswell-origin-experiment-digests-v1",\n  "tool_commit": "%s",\n  "files": {\n' "$tool_commit"
  first=true
  for name in union.json union-plan.json candidates.json verification.json model.json \
    plan-development.json predictions-development.json evaluation-development.json \
    plan-final_test.json predictions-final_test.json evaluation-final_test.json; do
    if [[ "$first" == true ]]; then first=false; else printf ',\n'; fi
    digest=$(sha "$output/$name")
    bytes=$(wc -c <"$output/$name")
    printf '    "%s": {"sha256": "%s", "bytes": %s}' "$name" "$digest" "${bytes// /}"
  done
  printf '\n  }\n}\n'
} >"$output/digests.json"

if [[ -n "$record" ]]; then
  mkdir -p "$record"
  for name in union.json union-plan.json model.json plan-development.json predictions-development.json \
    evaluation-development.json plan-final_test.json predictions-final_test.json evaluation-final_test.json digests.json; do
    cp "$output/$name" "$record/$name"
  done
fi
jq -c '{units: (.units | length), sources: (.sources | length)}' "$output/candidates.json"
jq -c '{task: .identity.task, basis, partitions: [.partitions[] | {name, rows: (.rows | length), classes, excluded}]}' "$output/model.json"
for partition in development final_test; do
  jq -c --arg p "$partition" '{partition: $p, candidates, excluded: .excluded_labels, micro: .summary.micro}' "$output/evaluation-$partition.json"
done
