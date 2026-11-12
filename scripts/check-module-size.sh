#!/usr/bin/env bash
# The Go module proxy refuses a module whose source tree exceeds 500 MB. This
# check keeps the root module far below that limit. It counts the tracked files
# outside the other classified modules, which is what a module archive holds.
set -euo pipefail
cd "$(dirname "$0")/.."

budget=$((64 * 1024 * 1024))
case "${1:-}" in
  --self-test)
    # The tree is larger than one byte, so the check must reject it.
    if bash "$0" --budget 1 >/dev/null 2>&1; then
      printf 'Module size check accepted a tree above its budget\n' >&2
      exit 1
    fi
    printf 'Module size check rejected a tree above its budget.\n'
    exit 0
    ;;
  --budget)
    budget=${2:?budget in bytes required}
    ;;
  '') ;;
  *)
    printf 'Usage: %s [--self-test | --budget BYTES]\n' "$0" >&2
    exit 2
    ;;
esac
[[ "$budget" =~ ^[0-9]+$ ]] || {
  printf 'Invalid budget: %s\n' "$budget" >&2
  exit 2
}

excludes=()
while read -r directory _; do
  if [[ "$directory" != . ]]; then excludes+=(":(exclude)$directory"); fi
done <.gomodules
size=$({ git ls-files -z -- . "${excludes[@]}" | xargs -0 cat 2>/dev/null || true; } | wc -c | tr -d ' ')
if ((size > budget)); then
  printf 'Root module source tree is %s bytes; the budget is %s bytes. Keep records in the research module.\n' "$size" "$budget" >&2
  exit 1
fi
printf 'Root module source tree: %s bytes of a %s byte budget.\n' "$size" "$budget"
