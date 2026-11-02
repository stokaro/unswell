#!/usr/bin/env bash
# Negative controls for the release binary reproducibility gate.
set -euo pipefail
cd "$(dirname "$0")/.."

workspace=$(mktemp -d)
trap 'rm -rf "$workspace"' EXIT
mkdir -p "$workspace/bin"
cat >"$workspace/bin/go" <<'FAKE_GO'
#!/usr/bin/env bash
set -euo pipefail
command_name=${1:?}
shift
case "$command_name" in
  build)
    case "${UNSWELL_REPRO_TEST_MODE:?}" in
      failed) exit 7 ;;
      missing) exit 0 ;;
      same | different)
        output=''
        while (($#)); do
          if [[ "$1" == -o ]]; then output=$2; break; fi
          shift
        done
        [[ -n "$output" ]] || exit 8
        content=identical
        if [[ "$UNSWELL_REPRO_TEST_MODE" == different && "$output" == */second/* ]]; then
          content=different
        fi
        printf '%s\n' "$content" >"$output"
        ;;
      *) exit 9 ;;
    esac
    ;;
  version) printf 'go version go-test darwin/arm64\n' ;;
  env)
    case "${1:-}" in
      GOHOSTOS) printf 'darwin\n' ;;
      GOHOSTARCH) printf 'arm64\n' ;;
      *) exit 10 ;;
    esac
    ;;
  *) exit 11 ;;
esac
FAKE_GO
chmod 700 "$workspace/bin/go"

check_failure() {
  local mode=$1 report="$workspace/$1.json"
  if UNSWELL_REPRO_TEST_MODE="$mode" PATH="$workspace/bin:$PATH" \
    bash scripts/verify-reproducible-build.sh --commit test-commit \
    --platforms linux/amd64 --output "$report" \
    >"$workspace/$mode.stdout" 2>"$workspace/$mode.stderr"; then
    printf 'Reproducibility gate accepted the %s negative control.\n' "$mode" >&2
    exit 1
  fi
  if [[ "$mode" != different && -e "$report" ]]; then
    printf 'Reproducibility gate published a report after %s.\n' "$mode" >&2
    exit 1
  fi
}

check_failure failed
check_failure missing
check_failure different
UNSWELL_REPRO_TEST_MODE=same PATH="$workspace/bin:$PATH" \
  bash scripts/verify-reproducible-build.sh --commit test-commit \
  --platforms linux/amd64 --output "$workspace/same.json" \
  >"$workspace/same.stdout" 2>"$workspace/same.stderr"
python3 - "$workspace/same.json" "$workspace/different.json" <<'PY'
import json
import sys

accepted = json.load(open(sys.argv[1], encoding="utf-8"))
rejected = json.load(open(sys.argv[2], encoding="utf-8"))
assert accepted["mismatched"] == 0
assert len(accepted["builds"]) == 1
assert accepted["builds"][0]["sha256"]
assert rejected["mismatched"] == 1
PY
printf 'Reproducibility gate rejected failed, missing, and mismatched builds.\n'
