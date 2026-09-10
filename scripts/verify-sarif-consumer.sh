#!/usr/bin/env bash
# Import an Unswell SARIF report into a real consumer. The consumer is the
# reference implementation from the specification authors, run in a pinned
# container: passing its validation is evidence about this report, not a claim
# about every consumer.
set -euo pipefail
script_directory=$(dirname "$0")
cd "$script_directory/.."
root=$PWD

context=${UNSWELL_DOCKER_CONTEXT:-remote-dev-container}
image=mcr.microsoft.com/dotnet/sdk:8.0
tool_version=5.7.0
report=artifacts/sarif/result.sarif
output=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --context)
      context=$2
      shift 2
      ;;
    --report)
      report=$2
      shift 2
      ;;
    --output)
      output=$2
      shift 2
      ;;
    *)
      printf 'Usage: %s [--context NAME] [--report FILE] [--output FILE]\n' "$0" >&2
      exit 2
      ;;
  esac
done

if [[ ! -x bin/unswell ]]; then
  printf 'Build the CLI first: make build.\n' >&2
  exit 2
fi
mkdir -p "$(dirname "$report")"
if [[ ! -f "$report" ]]; then
  bin/unswell check . --config .unswell.yaml --no-gate --report "sarif:$report" >/dev/null
fi

container=unswell-sarif-consumer-$$
cleanup() {
  docker --context "$context" rm -f "$container" >/dev/null 2>&1 || true
}
trap cleanup EXIT

docker --context "$context" run -d --name "$container" "$image" sleep 600 >/dev/null
docker --context "$context" exec "$container" mkdir -p /work
docker --context "$context" cp "$report" "$container:/work/result.sarif"
docker --context "$context" exec "$container" bash -c "
set -euo pipefail
dotnet tool install --global Sarif.Multitool --version $tool_version >/dev/null
export PATH=\"\$PATH:/root/.dotnet/tools\"
sarif validate /work/result.sarif --output /work/validation.sarif --max-file-size-in-kb 200000
" >artifacts/sarif/consumer.log 2>&1 || {
  status=$?
  printf 'The consumer rejected the report; see artifacts/sarif/consumer.log.\n' >&2
  docker --context "$context" cp "$container:/work/validation.sarif" artifacts/sarif/validation.sarif >/dev/null 2>&1 || true
  exit "$status"
}
docker --context "$context" cp "$container:/work/validation.sarif" artifacts/sarif/validation.sarif

results=$(python3 scripts/sarif-consumer-summary.py artifacts/sarif/validation.sarif)

report_bytes=$(wc -c <"$report")
report_bytes=${report_bytes// /}
unswell_version=$("$root/bin/unswell" --version 2>/dev/null || printf unknown)
summary=$(
  cat <<JSON
{
  "format": "unswell-sarif-consumer-check-v1",
  "consumer": "Sarif.Multitool $tool_version",
  "consumer_image": "$image",
  "report": "$report",
  "report_bytes": $report_bytes,
  "tool_version": "$unswell_version",
  "validation_results_by_level": $results,
  "scope": "One report from this build imported by the reference consumer; not a claim about every consumer."
}
JSON
)
if [[ -n "$output" ]]; then
  printf '%s\n' "$summary" >"$output"
fi
printf '%s\n' "$summary"
