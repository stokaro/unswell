#!/usr/bin/env bash
# Measure one complete scan against the stage 4 target: 100,000 prose words
# within 10 seconds and 512 MiB. The default corpus is synthetic and
# deterministic, so the numbers describe this tool on this host, not editorial
# quality. With --corpus the same harness scans a real project tree instead.
set -euo pipefail
script_directory=$(dirname "$0")
cd "$script_directory/.."
root=$PWD
# sha prints the SHA-256 digest of a file with the tool this host has.
sha() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -c1-64
  else
    shasum -a 256 "$1" | cut -c1-64
  fi
}

words=100000
output=""
label=""
corpus_source=""
excludes=()
timeout=10m
self_test=false
binary=""
origin_model=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --words)
      words=$2
      shift 2
      ;;
    --output)
      output=$2
      shift 2
      ;;
    --label)
      label=$2
      shift 2
      ;;
    --corpus)
      corpus_source=$2
      shift 2
      ;;
    --exclude)
      excludes+=("$2")
      shift 2
      ;;
    --timeout)
      timeout=$2
      shift 2
      ;;
    --binary)
      binary=$2
      shift 2
      ;;
    --origin-model)
      origin_model=$2
      shift 2
      ;;
    --self-test)
      self_test=true
      shift
      ;;
    *)
      printf 'Usage: %s [--words N] [--corpus DIR] [--exclude GLOB]... [--timeout DURATION] [--binary FILE] [--origin-model FILE] [--output FILE] [--label NAME] [--self-test]\n' "$0" >&2
      exit 2
      ;;
  esac
done

if ! [[ "$words" =~ ^[0-9]+$ ]] || ((words < 100)); then
  printf 'Requested word count must be an integer of at least 100.\n' >&2
  exit 2
fi
if [[ -n "$corpus_source" && "$self_test" == true ]]; then
  printf 'The self-test measures the synthetic corpus; it does not take --corpus.\n' >&2
  exit 2
fi
if [[ -n "$corpus_source" && ! -d "$corpus_source" ]]; then
  printf 'The corpus directory %s does not exist.\n' "$corpus_source" >&2
  exit 2
fi

# write_policy writes the scan policy of the measured tree. The synthetic corpus
# selects the formats it generates. A real corpus keeps the default file
# selection, so its own format mix is what gets measured, and it repeats the
# default exclusions because an exclude list in a policy replaces them. With
# an origin model the policy enables the origin channel and accepts an
# experimental pack, so the scan pays for the estimate.
write_policy() {
  local directory=$1 pattern
  {
    printf 'version: 1\nextends: [builtin:strict-v1]\nlanguage: en\n'
    if [[ -n "$origin_model" ]]; then
      printf 'origin:\n  model: pack\n  accept_experimental: true\nfiles:\n'
    else
      printf 'files:\n'
    fi
    if [[ -z "$corpus_source" ]]; then
      printf '  include:\n    - "**/*.md"\n    - "**/*.go"\n    - "**/*.yaml"\n    - "**/*.sh"\n'
    else
      printf '  exclude:\n'
      for pattern in ".git/**" "vendor/**" "testdata/**" "artifacts/**" "dist/**" "rules/**" "**/*.generated.go" \
        ${excludes[@]+"${excludes[@]}"}; do
        printf '    - "%s"\n' "$pattern"
      done
    fi
  } >"$directory/.unswell.yaml"
}

# prepare_corpus copies a real project tree into the workspace without its
# version control metadata. The scan then reads the tree a checkout holds, and
# the policy the measurement adds never touches the source directory.
prepare_corpus() {
  local source=$1 directory=$2
  mkdir -p "$directory"
  cp -R "$source/." "$directory/"
  rm -rf "$directory/.git"
}

# generate_corpus writes deterministic technical prose across every scanned
# format. Sentences repeat by design: the measurement needs a fixed workload,
# and repetition rules must still see the candidates a real scan would.
generate_corpus() {
  local directory=$1 target=$2
  mkdir -p "$directory/docs" "$directory/service" "$directory/config" "$directory/scripts"
  # The corpus carries its own policy: the repository's own file selection is
  # written for the repository, not for a generated tree.
  write_policy "$directory"
  UNSWELL_CORPUS_DIRECTORY=$directory UNSWELL_CORPUS_WORDS=$target awk '
    BEGIN {
      directory = ENVIRON["UNSWELL_CORPUS_DIRECTORY"]
      target = ENVIRON["UNSWELL_CORPUS_WORDS"] + 0
      split("The cache retries after a transport failure and records the attempt|" \
        "Connection pools release idle sockets once the deadline passes|" \
        "The parser rejects a malformed header before it reaches the queue|" \
        "Workers drain their backlog when the scheduler reports pressure|" \
        "A failed write leaves the previous revision on disk untouched|" \
        "The client resolves each endpoint once and caches the result|" \
        "Metrics record the retry count separately from the failure count|" \
        "The reader stops at the first boundary that the policy declares", pool, "|")
      words_per_sentence = 11
      per_file = 2000
      file_index = 0
      written = 0
      sentence = 0
      while (written < target) {
        file_index++
        remaining = target - written
        chunk = remaining < per_file ? remaining : per_file
        kind = file_index % 4
        if (kind == 1) { path = sprintf("%s/docs/guide-%03d.md", directory, file_index) }
        else if (kind == 2) { path = sprintf("%s/service/handler_%03d.go", directory, file_index) }
        else if (kind == 3) { path = sprintf("%s/config/values-%03d.yaml", directory, file_index) }
        else { path = sprintf("%s/scripts/task-%03d.sh", directory, file_index) }
        if (kind == 1) { printf "# Handler notes %03d\n\n", file_index > path }
        else if (kind == 2) { printf "// Package service %03d documents one handler.\npackage service\n\n", file_index > path }
        else if (kind == 3) { printf "# Values %03d\nservice:\n", file_index > path }
        else { printf "#!/usr/bin/env bash\n# Task %03d prepares one queue.\nset -euo pipefail\n", file_index > path }
        emitted = 0
        while (emitted < chunk) {
          sentence++
          text = pool[(sentence % 8) + 1] "."
          if (kind == 1) {
            printf "%s\n", text > path
            if (sentence % 5 == 0) { printf "\n" > path }
          } else if (kind == 2) {
            printf "// %s\n", text > path
          } else if (kind == 3) {
            printf "  note_%06d: \"%s\"\n", sentence, text > path
          } else {
            printf "# %s\n", text > path
          }
          emitted += words_per_sentence
        }
        if (kind == 2) { printf "\nfunc Handle() {}\n" > path }
        close(path)
        written += emitted
      }
    }
  '
}

# measure runs one complete scan and reports its wall time, peak resident set
# and exit code. Peak memory covers the whole process, including report writing.
measure() {
  local directory=$1 reports=$2 stamp=$3
  local time_tool=() started ended status=0
  mkdir -p "$reports"
  if /usr/bin/time -v true >/dev/null 2>&1; then
    time_tool=(/usr/bin/time -v -o "$stamp")
  elif /usr/bin/time -l true >/dev/null 2>&1; then
    time_tool=(/usr/bin/time -l -o "$stamp")
  else
    printf 'No /usr/bin/time with resource reporting is available.\n' >&2
    return 2
  fi
  started=$(date +%s.%N)
  "${time_tool[@]}" "$tool" check "$directory" \
    ${root_flags[@]+"${root_flags[@]/PLACEHOLDER/$directory}"} --config "$directory/.unswell.yaml" --timeout "$timeout" \
    ${origin_flags[@]+"${origin_flags[@]}"} \
    --report "json:$reports/result.json" \
    --report "sarif:$reports/result.sarif" \
    --report "html:$reports/result.html" \
    --report "markdown:$reports/result.md" \
    --report "text:$reports/result.txt" >/dev/null 2>"$reports/stderr.txt" || status=$?
  ended=$(date +%s.%N)
  local elapsed
  elapsed=$(awk -v a="$started" -v b="$ended" 'BEGIN { printf "%.3f", b - a }')
  printf '%s %s\n' "$elapsed" "$status"
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

# count_prose_words reports what the tool itself counts in a corpus, which is
# what the target is stated in. The generator's own estimate is not the measure.
count_prose_words() {
  local directory=$1 reports=$2
  mkdir -p "$reports"
  "$tool" check "$directory" ${root_flags[@]+"${root_flags[@]/PLACEHOLDER/$directory}"} --timeout "$timeout" \
    --config "$directory/.unswell.yaml" --no-gate ${origin_flags[@]+"${origin_flags[@]}"} \
    --report "json:$reports/count.json" >/dev/null 2>&1 || true
  awk '/"prose_words"/ { gsub(/[^0-9]/, "", $2); total += $2 } END { print total + 0 }' "$reports/count.json"
}

workspace=$(mktemp -d)
trap 'rm -rf "$workspace"' EXIT
corpus=$workspace/corpus

# The measured tool is the repository build unless --binary names another
# executable, such as a published release asset; the record carries its
# digest either way.
# absolute prints the absolute path of a file that exists.
absolute() {
  local directory
  directory=$(dirname "$1")
  directory=$(cd "$directory" && pwd)
  printf '%s/%s' "$directory" "$(basename "$1")"
}
tool=$root/bin/unswell
if [[ -n "$binary" ]]; then
  tool=$(absolute "$binary")
fi
if [[ ! -x "$tool" ]]; then
  printf 'Build the CLI first (make build) or name an executable with --binary.\n' >&2
  exit 2
fi
# An older published binary may lack a flag this script passes. The record
# names every flag it left out, and a requested origin model without its flag
# is an error rather than a silent scan without the model.
supported_flags=$("$tool" check --help 2>&1 || true)
root_flags=(--project-root "PLACEHOLDER")
omitted_flags=""
if [[ "$supported_flags" != *"--project-root"* ]]; then
  root_flags=()
  omitted_flags="--project-root"
fi
if [[ -n "$origin_model" && "$supported_flags" != *"--origin-model"* ]]; then
  printf 'The measured binary does not accept --origin-model.\n' >&2
  exit 2
fi
origin_flags=()
origin_model_sha256=""
if [[ -n "$origin_model" ]]; then
  if [[ ! -f "$origin_model" ]]; then
    printf 'The origin model %s does not exist.\n' "$origin_model" >&2
    exit 2
  fi
  origin_model=$(absolute "$origin_model")
  origin_flags=(--origin-model "$origin_model")
  origin_model_sha256=$(sha "$origin_model")
fi
tool_sha256=$(sha "$tool")

if [[ -n "$corpus_source" ]]; then
  # A real corpus is measured as it is; nothing sizes it to the target.
  prepare_corpus "$corpus_source" "$corpus"
  write_policy "$corpus"
  requested_words=0
else
  # Calibrate the generated size against the counted size, so the measurement
  # covers at least the requested number of words rather than an estimate.
  requested_words=$words
  generated=$words
  for attempt in 1 2 3; do
    rm -rf "$corpus"
    generate_corpus "$corpus" "$generated"
    counted=$(count_prose_words "$corpus" "$workspace/calibration")
    if ((counted >= words)); then
      break
    fi
    if ((counted <= 0)); then
      printf 'The generated corpus produced no prose words.\n' >&2
      exit 1
    fi
    generated=$((generated * words * 102 / (counted * 100) + 1))
    if ((attempt == 3)); then
      printf 'Calibration reached %s prose words for a %s word request.\n' "$counted" "$words" >&2
    fi
  done
fi

cold_result=$(measure "$corpus" "$workspace/cold" "$workspace/cold.time")
read -r cold_seconds cold_status <<<"$cold_result"
cold_peak=$(peak_bytes "$workspace/cold.time")
warm_result=$(measure "$corpus" "$workspace/warm" "$workspace/warm.time")
read -r warm_seconds warm_status <<<"$warm_result"
warm_peak=$(peak_bytes "$workspace/warm.time")

# A scan the kernel killed leaves no report. The observation then records the
# exit code the time tool saw, with zero counts and an unknown outcome.
for report_file in result.json result.txt; do
  [[ -f "$workspace/cold/$report_file" ]] || : >"$workspace/cold/$report_file"
done
measured_words=$(awk '/"prose_words"/ { gsub(/[^0-9]/, "", $2); total += $2 } END { print total + 0 }' \
  "$workspace/cold/result.json")
# The text report states the counts and the outcome the run itself reached,
# and lists every operational error of a scan that did not complete.
summary=$(sed -n 's/^\([A-Z]*\): \([0-9]*\) documents, \([0-9]*\) findings.*/\1 \2 \3/p' "$workspace/cold/result.txt" | tail -1)
read -r outcome documents findings <<<"${summary:-unknown 0 0}"
errors=$(grep -c '^  error: ' "$workspace/cold/result.txt" || true)
if [[ -n "$corpus_source" ]]; then
  corpus_name=$(basename "$corpus_source")
  scope="Real project tree $corpus_name; measures this build on this host on that tree, not editorial quality."
else
  corpus_name=synthetic
  scope="Synthetic deterministic corpus; measures this build on this host, not editorial quality."
fi
exclude_list=""
for pattern in ${excludes[@]+"${excludes[@]}"}; do
  exclude_list="${exclude_list:+$exclude_list, }\"$pattern\""
done
corpus_bytes=$(find "$corpus" -type f -exec cat {} + >"$workspace/corpus.bytes" && wc -c <"$workspace/corpus.bytes")
corpus_bytes=${corpus_bytes// /}
tool_version=$("$tool" --version 2>/dev/null || printf 'unknown')
tool_version=${tool_version%%$'\n'*}
system_name=$(uname -s)
machine_name=$(uname -m)
cpu_count=$(getconf _NPROCESSORS_ONLN 2>/dev/null || printf 0)

report=$(
  cat <<JSON
{
  "format": "unswell-scan-cost-observation-v1",
  "label": "${label:-unlabeled}",
  "scope": "$scope",
  "corpus": "$corpus_name",
  "excludes": [$exclude_list],
  "analysis_timeout": "$timeout",
  "tool_version": "$tool_version",
  "tool_sha256": "$tool_sha256",
  "omitted_flags": "$omitted_flags",
  "origin_model_sha256": "$origin_model_sha256",
  "requested_words": $requested_words,
  "measured_prose_words": $measured_words,
  "documents": $documents,
  "findings": $findings,
  "errors": $errors,
  "outcome": "$outcome",
  "corpus_bytes": $corpus_bytes,
  "goos": "$system_name",
  "goarch": "$machine_name",
  "logical_cpus": $cpu_count,
  "gomaxprocs": "${GOMAXPROCS:-runtime default}",
  "reports": ["json", "sarif", "html", "markdown", "text"],
  "cold_seconds": $cold_seconds,
  "cold_exit_code": $cold_status,
  "cold_maximum_resident_bytes": $cold_peak,
  "warm_seconds": $warm_seconds,
  "warm_exit_code": $warm_status,
  "warm_maximum_resident_bytes": $warm_peak,
  "gate_evaluated": true
}
JSON
)

if [[ -n "$output" ]]; then
  printf '%s\n' "$report" >"$output"
fi
printf '%s\n' "$report"

if [[ "$self_test" == true ]]; then
  if ((measured_words < words / 2)); then
    printf 'The generated corpus produced %s prose words for a %s word request.\n' "$measured_words" "$words" >&2
    exit 1
  fi
  if ((cold_status > 1 || warm_status > 1)); then
    printf 'A complete scan must end in a policy decision, not an operational failure.\n' >&2
    exit 1
  fi
  if ((cold_peak <= 0 || warm_peak <= 0)); then
    printf 'Peak resident memory was not reported by the available time tool.\n' >&2
    exit 1
  fi
  # An exceeded analysis limit must fail the whole run, not pass a partial scan.
  bounded=$workspace/bounded
  mkdir -p "$bounded"
  sed 's/^extends:.*/&\nanalysis:\n  max_file_bytes: 64/' "$corpus/.unswell.yaml" >"$corpus/bounded.yaml"
  bounded_status=0
  "$tool" check "$corpus" ${root_flags[@]+"${root_flags[@]/PLACEHOLDER/$corpus}"} --config "$corpus/bounded.yaml" \
    --report "json:$bounded/result.json" >"$bounded/stdout.txt" 2>"$bounded/stderr.txt" || bounded_status=$?
  if ((bounded_status != 2)); then
    printf 'An exceeded max_file_bytes limit must exit 2; it exited %s.\n' "$bounded_status" >&2
    exit 1
  fi
  if grep -q '"status": "complete"' "$bounded/result.json" 2>/dev/null; then
    printf 'A run that exceeded a limit must not report a complete analysis.\n' >&2
    exit 1
  fi
  printf 'Self-test passed: %s prose words in %s documents, cold %ss, warm %ss; an exceeded limit exited 2.\n' \
    "$measured_words" "$documents" "$cold_seconds" "$warm_seconds"
fi
