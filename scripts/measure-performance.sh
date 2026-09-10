#!/usr/bin/env bash
# Measure one complete scan against the stage 4 target: 100,000 prose words
# within 10 seconds and 512 MiB. The corpus is synthetic and deterministic, so
# the numbers describe this tool on this host, not editorial quality.
set -euo pipefail
script_directory=$(dirname "$0")
cd "$script_directory/.."
root=$PWD

words=100000
output=""
label=""
self_test=false
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
    --self-test)
      self_test=true
      shift
      ;;
    *)
      printf 'Usage: %s [--words N] [--output FILE] [--label NAME] [--self-test]\n' "$0" >&2
      exit 2
      ;;
  esac
done

if ! [[ "$words" =~ ^[0-9]+$ ]] || ((words < 100)); then
  printf 'Requested word count must be an integer of at least 100.\n' >&2
  exit 2
fi

# generate_corpus writes deterministic technical prose across every scanned
# format. Sentences repeat by design: the measurement needs a fixed workload,
# and repetition rules must still see the candidates a real scan would.
generate_corpus() {
  local directory=$1 target=$2
  mkdir -p "$directory/docs" "$directory/service" "$directory/config" "$directory/scripts"
  # The corpus carries its own policy: the repository's own file selection is
  # written for the repository, not for a generated tree.
  cat >"$directory/.unswell.yaml" <<'POLICY'
version: 1
extends: [builtin:strict-v1]
language: en
files:
  include:
    - "**/*.md"
    - "**/*.go"
    - "**/*.yaml"
    - "**/*.sh"
POLICY
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
  "${time_tool[@]}" "$root/bin/unswell" check "$directory" \
    --project-root "$directory" --config "$directory/.unswell.yaml" \
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
  "$root/bin/unswell" check "$directory" --project-root "$directory" \
    --config "$directory/.unswell.yaml" --no-gate --report "json:$reports/count.json" >/dev/null 2>&1 || true
  awk '/"prose_words"/ { gsub(/[^0-9]/, "", $2); total += $2 } END { print total + 0 }' "$reports/count.json"
}

workspace=$(mktemp -d)
trap 'rm -rf "$workspace"' EXIT
corpus=$workspace/corpus

if [[ ! -x bin/unswell ]]; then
  printf 'Build the CLI first: make build.\n' >&2
  exit 2
fi

# Calibrate the generated size against the counted size, so the measurement
# covers at least the requested number of words rather than an estimate.
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

cold_result=$(measure "$corpus" "$workspace/cold" "$workspace/cold.time")
read -r cold_seconds cold_status <<<"$cold_result"
cold_peak=$(peak_bytes "$workspace/cold.time")
warm_result=$(measure "$corpus" "$workspace/warm" "$workspace/warm.time")
read -r warm_seconds warm_status <<<"$warm_result"
warm_peak=$(peak_bytes "$workspace/warm.time")

measured_words=$(awk '/"prose_words"/ { gsub(/[^0-9]/, "", $2); total += $2 } END { print total + 0 }' \
  "$workspace/cold/result.json")
# The text report states the counts the run itself reached.
summary=$(tail -20 "$workspace/cold/result.txt" | sed -n 's/^[A-Z]*: \([0-9]*\) documents, \([0-9]*\) findings.*/\1 \2/p' | tail -1)
documents=${summary%% *}
findings=${summary##* }
corpus_bytes=$(find "$corpus" -type f -exec cat {} + >"$workspace/corpus.bytes" && wc -c <"$workspace/corpus.bytes")
corpus_bytes=${corpus_bytes// /}
tool_version=$("$root/bin/unswell" --version 2>/dev/null || printf 'unknown')
tool_version=${tool_version%%$'\n'*}
system_name=$(uname -s)
machine_name=$(uname -m)
cpu_count=$(getconf _NPROCESSORS_ONLN 2>/dev/null || printf 0)

report=$(
  cat <<JSON
{
  "format": "unswell-scan-cost-observation-v1",
  "label": "${label:-unlabeled}",
  "scope": "Synthetic deterministic corpus; measures this build on this host, not editorial quality.",
  "tool_version": "$tool_version",
  "requested_words": $words,
  "measured_prose_words": $measured_words,
  "documents": $documents,
  "findings": $findings,
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
  "$root/bin/unswell" check "$corpus" --project-root "$corpus" --config "$corpus/bounded.yaml" \
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
