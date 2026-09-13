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
# A plan built with pinned partitions can only be reopened with the same file.
partitions=research/methods/partitions-v1.json
[[ -f "$partitions" ]] || partitions=
findings=artifacts/measurement/findings
tasks=research/generation/runs/2026-09-11-pilot/tasks.json
generation=research/generation/runs/2026-09-11-pilot/records.json
protocol=research/methods/llm-patterns-v1.md
output=artifacts/origin
record=""
id=origin-pilot-v1
cap=15
max_source_bytes=102400
threshold=0.5
negative_roles=()
exclude_source_files=()
scored_partitions=()
extra_features=()
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
    --partitions)
      partitions=$2
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
    --negative-role)
      negative_roles+=("$2")
      shift 2
      ;;
    --findings)
      findings=$2
      shift 2
      ;;
    --max-source-bytes)
      max_source_bytes=$2
      shift 2
      ;;
    --exclude-source-file)
      exclude_source_files+=("$2")
      shift 2
      ;;
    --score-partition)
      scored_partitions+=("$2")
      shift 2
      ;;
    --extra-feature)
      extra_features+=("$2")
      shift 2
      ;;
    *)
      printf 'Usage: %s OPTIONS\n' "$0" >&2
      printf 'Options: --work DIR, --acquisition DIR, --dataset-plan FILE, --tasks FILE, --generation FILE, --protocol FILE, --output DIR, --record DIR, --id ID, --cap N, --max-source-bytes N, --exclude-source-file FILE (repeatable), --score-partition NAME (repeatable; default development and final_test), --extra-feature ID (repeatable), --threshold T, --negative-role ROLE (repeatable; default comment), --findings DIR\n' >&2
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

# The union takes every controlled source and the negative-role sources of
# the task repositories. The controlled sources enter whatever their role.
# The default negatives are comments, the role the tasks came from.
# Documentation roles give negatives in the format of the endpoints, which
# are Markdown documents. A cap per historical checkout
# keeps the corpus within one artifact. Files above the byte limit stay out:
# one bundle of thousands of comments would be most of the corpus and
# exhaust the measurement budget of its engine run. Only paragraphs are
# extracted.
if ((${#negative_roles[@]} == 0)); then
  negative_roles=(comment)
fi
role_flags=()
for role in "${negative_roles[@]}"; do
  role_flags+=(--role "$role")
done
repository_list=$(jq -r '[.tasks[].repository] | unique | .[]' "$tasks")
repositories=()
while IFS= read -r repository; do
  repositories+=(--repository "$repository")
done <<<"$repository_list"
# Sources the pattern measurement could not analyze, such as a block of
# non-Latin prose, stay out; the engine would refuse the whole run for one.
# A file of source IDs leaves out a whole arm of the controlled cohort, which
# is how a trial holds response length fixed across the corpus.
excluded_sources=$(
  jq -r '.units[] | select(.unmeasured == true) | .source_id' "$findings"/historical*.json
  if ((${#exclude_source_files[@]} > 0)); then
    cat "${exclude_source_files[@]}"
  fi
)
excluded_sources=$(printf '%s\n' "$excluded_sources" | sort -u)
exclusions=()
while IFS= read -r source_id; do
  [[ -n "$source_id" ]] && exclusions+=(--exclude-source "$source_id")
done <<<"$excluded_sources"
pin_flags=()
if [[ -n "$partitions" ]]; then
  pin_flags+=(--partitions "$partitions")
fi
"$corpus_tool" dataset union --root "$acquisition" --id "$id" --cohort controlled --cohort historical \
  ${pin_flags[@]+"${pin_flags[@]}"} \
  "${role_flags[@]}" --every-role-cohort controlled --unit-kind paragraph --max-per-checkout "$cap" \
  --uncapped-cohort controlled --max-source-bytes "$max_source_bytes" "${repositories[@]}" \
  ${exclusions[@]+"${exclusions[@]}"} <"$dataset_plan" >"$output/union.json"
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
# A declared ablation adds a feature family to the fourteen without touching
# the list the primary trial fixed.
features+=(${extra_features[@]+"${extra_features[@]}"})
feature_flags=()
for feature in "${features[@]}"; do
  feature_flags+=(--feature "$feature")
done
"$corpus_tool" train --root "$work" --labels provenance --kind paragraph "${feature_flags[@]}" \
  --missing-features exclude --calibration isotonic <"$output/candidates.json" >"$output/model.json"

# A stopping rule that opens the confirmation partition only on a positive
# development result needs a run that scores one partition and stops.
if ((${#scored_partitions[@]} == 0)); then
  scored_partitions=(development final_test)
fi
sha() { shasum -a 256 "$1" | cut -c1-64; }
model_sha=$(jq -r '.sha256' "$output/model.json")
corpus_sha=$(jq -r '.sha256' "$output/candidates.json")
protocol_sha=$(sha "$protocol")
for partition in "${scored_partitions[@]}"; do
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
  names=(union.json union-plan.json candidates.json verification.json model.json)
  for partition in "${scored_partitions[@]}"; do
    names+=("plan-$partition.json" "predictions-$partition.json" "evaluation-$partition.json")
  done
  for name in "${names[@]}"; do
    if [[ "$first" == true ]]; then first=false; else printf ',\n'; fi
    digest=$(sha "$output/$name")
    bytes=$(wc -c <"$output/$name")
    printf '    "%s": {"sha256": "%s", "bytes": %s}' "$name" "$digest" "${bytes// /}"
  done
  printf '\n  }\n}\n'
} >"$output/digests.json"

if [[ -n "$record" ]]; then
  mkdir -p "$record"
  kept=(union.json union-plan.json model.json digests.json)
  for partition in "${scored_partitions[@]}"; do
    kept+=("plan-$partition.json" "predictions-$partition.json" "evaluation-$partition.json")
  done
  for name in "${kept[@]}"; do
    cp "$output/$name" "$record/$name"
  done
fi
jq -c '{units: (.units | length), sources: (.sources | length)}' "$output/candidates.json"
jq -c '{task: .identity.task, basis, partitions: [.partitions[] | {name, rows: (.rows | length), classes, excluded}]}' "$output/model.json"
for partition in "${scored_partitions[@]}"; do
  jq -c --arg p "$partition" '{partition: $p, candidates, excluded: .excluded_labels, micro: .summary.micro}' "$output/evaluation-$partition.json"
done
