#!/usr/bin/env bash
# Run the LLMDet part of experiment E3. The corpus is the union of the
# historical and contemporary shards, as in the baseline comparison. A table
# pack names the local dictionary tables and tokenizer files of the ported
# models. The arms compare the proxy columns alone and joined with the
# structural features against the structural features and the n-grams.
# Every step is a corpus command; this script orders them and records
# digests and costs. The tables stay outside the repository.
set -euo pipefail
script_directory=$(dirname "$0")
cd "$script_directory/.."

work=${UNSWELL_ACQUISITION_WORK:-$HOME/.cache/unswell/acquisition/work}
acquisition=artifacts/acquisition
dataset_plan=artifacts/measurement/dataset-plan.json
findings=artifacts/measurement/findings
protocol=research/methods/llm-patterns-v1.md
output=artifacts/llmdet
record=""
id=e3-llmdet-v1
cap=4
max_source_bytes=102400
threshold=0.5
pack=${UNSWELL_LLMDET_PACK:-$HOME/.cache/unswell/llmdet/pack.json}
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
    --findings)
      findings=$2
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
    --max-source-bytes)
      max_source_bytes=$2
      shift 2
      ;;
    --threshold)
      threshold=$2
      shift 2
      ;;
    --pack)
      pack=$2
      shift 2
      ;;
    *)
      printf 'Usage: %s OPTIONS\n' "$0" >&2
      printf 'Options: --work DIR, --acquisition DIR, --dataset-plan FILE, --findings DIR, --protocol FILE, --output DIR, --record DIR, --id ID, --cap N, --max-source-bytes N, --threshold T, --pack FILE\n' >&2
      exit 2
      ;;
  esac
done
for value in "$cap" "$max_source_bytes"; do
  if ! [[ "$value" =~ ^[0-9]+$ ]]; then
    printf 'The cap and the byte limit must be non-negative integers.\n' >&2
    exit 2
  fi
done
if [[ ! -f "$pack" ]]; then
  printf 'The table pack %s does not exist; convert the dictionaries first.\n' "$pack" >&2
  exit 2
fi

mkdir -p "$output"
output=$(cd "$output" && pwd)
corpus_tool=$output/corpus
(cd research/annotation && go build -o "$corpus_tool" ./cmd/corpus)
tool_commit=$(git rev-parse HEAD)

resource_flag() {
  if /usr/bin/time -v true >/dev/null 2>&1; then
    printf '%s' -v
  elif /usr/bin/time -l true >/dev/null 2>&1; then
    printf '%s' -l
  else
    printf 'No /usr/bin/time with resource reporting is available.\n' >&2
    return 2
  fi
}
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
flag=$(resource_flag)
costs=$output/costs.records
: >"$costs"

# stage runs one corpus command under the time tool and appends its wall time,
# peak resident set, and exit code. A stage that fails stops the run.
stage() {
  local name=$1 input=$2 result=$3
  shift 3
  local stamp=$output/$name.time started ended status=0 seconds peak
  started=$(date +%s.%N)
  /usr/bin/time "$flag" -o "$stamp" "$corpus_tool" "$@" <"$input" >"$result" 2>"$output/$name.err" || status=$?
  ended=$(date +%s.%N)
  seconds=$(awk -v a="$started" -v b="$ended" 'BEGIN { printf "%.3f", b - a }')
  peak=$(peak_bytes "$stamp")
  printf '%s %s %s %s\n' "$name" "$seconds" "$peak" "$status" >>"$costs"
  if ((status != 0)); then
    tail -5 "$output/$name.err" >&2
    return "$status"
  fi
}

# The union is the one of the baseline comparison, so the corpus and its
# partitions are the same. Sources the pattern measurement could not analyze
# stay out for the same reason.
excluded_sources=$(jq -r '.units[] | select(.unmeasured == true) | .source_id' "$findings"/historical*.json "$findings"/contemporary*.json | sort -u)
exclusions=()
while IFS= read -r source_id; do
  [[ -n "$source_id" ]] && exclusions+=(--exclude-source "$source_id")
done <<<"$excluded_sources"
printf '%s\n' "$excluded_sources" | jq -R . | jq -s '{format: "unswell-baseline-experiment-exclusions-v1", reason: "unmeasured under the pattern measurement", sources: .}' >"$output/excluded-sources.json"
"$corpus_tool" dataset union --root "$acquisition" --id "$id" --cohort historical --cohort contemporary \
  --role documentation --role readme --role release_note \
  --unit-kind paragraph --max-per-checkout "$cap" --max-source-bytes "$max_source_bytes" \
  ${exclusions[@]+"${exclusions[@]}"} <"$dataset_plan" >"$output/union.json"
stage plan "$output/union.json" "$output/union-plan.json" plan
stage extract "$output/union-plan.json" "$output/candidates.json" extract --root "$work"
stage verify "$output/candidates.json" "$output/verification.json" verify --root "$work"

# The pack manifest is copied beside the run, so the digests name the exact
# tables and tokenizers. The commands read the original, whose relative
# paths resolve against its own directory.
cp "$pack" "$output/pack.json"

# Arms. The structural arm and the n-gram arm are the baselines. The LLMDet
# arm takes two columns per model, a proxy perplexity and a context
# coverage, and the joint arm adds them to the structural features. Every
# logistic fit stops at a gradient norm of 1e-6.
tolerance=1e-6
prepared_features=(prose-words counted-characters prose-sentences mean-sentence-words sentence-word-stddev
  shortest-sentence longest-sentence type-token-ratio hapax-token-ratio noun-token-ratio verb-token-ratio
  adjective-token-ratio adverb-token-ratio automated-readability-index)
feature_flags() {
  local feature
  for feature in "$@"; do
    printf -- '--feature\n%s\n' "$feature"
  done
}
arms=(prepared lexical llmdet prepared-llmdet)
train_flags() {
  case "$1" in
    prepared)
      feature_flags "${prepared_features[@]}"
      printf -- '--missing-features\nexclude\n--tolerance\n%s\n' "$tolerance"
      ;;
    lexical)
      printf -- '--lexical\n--tolerance\n%s\n' "$tolerance"
      ;;
    llmdet)
      printf -- '--llmdet-pack\n%s\n--missing-features\nexclude\n--tolerance\n%s\n' "$pack" "$tolerance"
      ;;
    prepared-llmdet)
      feature_flags "${prepared_features[@]}"
      printf -- '--llmdet-pack\n%s\n--missing-features\nexclude\n--tolerance\n%s\n' "$pack" "$tolerance"
      ;;
  esac
}
predict_flags() {
  case "$1" in
    *llmdet) printf -- '--llmdet-pack\n%s\n' "$pack" ;;
    *) : ;;
  esac
}

sha() { shasum -a 256 "$1" | cut -c1-64; }
corpus_sha=$(jq -r '.sha256' "$output/candidates.json")
protocol_sha=$(sha "$protocol")
for arm in "${arms[@]}"; do
  mkdir -p "$output/$arm"
  flag_lines=$(train_flags "$arm")
  flags=()
  while IFS= read -r line; do
    [[ -n "$line" ]] && flags+=("$line")
  done <<<"$flag_lines"
  stage "train-$arm" "$output/candidates.json" "$output/$arm/model.json" train --root "$work" --labels cohort \
    --kind paragraph --calibration isotonic "${flags[@]}"
  model_sha=$(jq -r '.sha256' "$output/$arm/model.json")
  extra_lines=$(predict_flags "$arm")
  extra=()
  while IFS= read -r line; do
    [[ -n "$line" ]] && extra+=("$line")
  done <<<"$extra_lines"
  for partition in development final_test; do
    cat >"$output/$arm/plan-$partition.json" <<PLAN
{
  "version": "unswell-research-predictions-v2",
  "id": "$id-$arm-$partition",
  "protocol_sha256": "$protocol_sha",
  "model_sha256": "$model_sha",
  "corpus_sha256": "$corpus_sha",
  "partition": "$partition",
  "context": "prepared_piece",
  "response": "isotonic",
  "threshold": $threshold
}
PLAN
    stage "predict-$arm-$partition" "$output/candidates.json" "$output/$arm/predictions-$partition.json" \
      predict --root "$work" --model "$output/$arm/model.json" --plan "$output/$arm/plan-$partition.json" \
      --protocol "$protocol" ${extra[@]+"${extra[@]}"}
    stage "evaluate-$arm-$partition" "$output/$arm/predictions-$partition.json" "$output/$arm/evaluation-$partition.json" \
      evaluate --corpus "$output/candidates.json" --labels cohort
  done
done

# Every arm reads the prepared piece, so every pair is comparable. The proxy
# columns compare with the structural features alone and joined with them,
# and the joint arm compares with the n-grams.
compare_pair() {
  local candidate=$1 comparator=$2 name=$3
  local a=$output/$candidate/predictions-final_test.json b=$output/$comparator/predictions-final_test.json
  local candidate_sha comparator_sha
  candidate_sha=$(jq -r '.sha256' "$a")
  comparator_sha=$(jq -r '.sha256' "$b")
  cat >"$output/compare-$name-plan.json" <<PLAN
{
  "version": "unswell-research-comparison-v1",
  "id": "$id-$name",
  "protocol_sha256": "$protocol_sha",
  "candidate_predictions_sha256": "$candidate_sha",
  "comparator_predictions_sha256": "$comparator_sha"
}
PLAN
  stage "compare-$name" "$a" "$output/compare-$name.json" compare --plan "$output/compare-$name-plan.json" \
    --protocol "$protocol" --comparator "$b" --corpus "$output/candidates.json" --labels cohort
}
compare_pair llmdet prepared llmdet-vs-prepared
compare_pair prepared-llmdet prepared joint-vs-prepared
compare_pair prepared-llmdet lexical joint-vs-lexical

# Costs and digests bind the record to the exact run. The pack manifest
# carries the digest of every table and tokenizer file, so the record names
# them without holding them.
system_name=$(uname -s)
machine_name=$(uname -m)
{
  printf '{\n  "format": "unswell-baseline-experiment-costs-v1",\n  "tool_commit": "%s",\n  "host": "%s %s",\n  "stages": [\n' "$tool_commit" "$system_name" "$machine_name"
  first=true
  while read -r name seconds peak status; do
    if [[ "$first" == true ]]; then first=false; else printf ',\n'; fi
    printf '    {"stage": "%s", "seconds": %s, "maximum_resident_bytes": %s, "exit_code": %s}' "$name" "$seconds" "$peak" "$status"
  done <"$costs"
  printf '\n  ]\n}\n'
} >"$output/costs.json"
artifact_files=$(find "$output" -name '*.json' ! -name digests.json ! -name costs.json | sort)
{
  printf '{\n  "format": "unswell-baseline-experiment-digests-v1",\n  "tool_commit": "%s",\n  "files": {\n' "$tool_commit"
  first=true
  while IFS= read -r file; do
    name=${file#"$output/"}
    if [[ "$first" == true ]]; then first=false; else printf ',\n'; fi
    digest=$(sha "$file")
    bytes=$(wc -c <"$file")
    printf '    "%s": {"sha256": "%s", "bytes": %s}' "$name" "$digest" "${bytes// /}"
  done <<<"$artifact_files"
  printf '\n  }\n}\n'
} >"$output/digests.json"

if [[ -n "$record" ]]; then
  mkdir -p "$record"
  cp "$output/union.json" "$output/union-plan.json" "$output/excluded-sources.json" "$output/pack.json" \
    "$output/costs.json" "$output/digests.json" "$record/"
  for arm in "${arms[@]}"; do
    mkdir -p "$record/$arm"
    cp "$output/$arm/model.json" "$output/$arm"/plan-*.json "$output/$arm"/evaluation-*.json "$record/$arm/"
  done
  cp "$output"/compare-*.json "$record/"
fi
jq -c '{units: (.units | length), sources: (.sources | length)}' "$output/candidates.json"
jq -c '{models: [.models[] | .model]}' "$output/pack.json"
for arm in "${arms[@]}"; do
  jq -c --arg a "$arm" '{arm: $a, task: .identity.task, training: (.partitions[0] | {rows: (.rows | length), classes, excluded})}' "$output/$arm/model.json"
  jq -c --arg a "$arm" '{arm: $a, partition: "final_test", eligible: .summary.micro.counts.eligible, covered: .summary.micro.counts.covered, recall: .summary.micro.recall, false_positive_rate: .summary.micro.false_positive_rate, brier: .summary.micro.brier, constant: .summary.micro.training_constant_brier}' "$output/$arm/evaluation-final_test.json"
done
