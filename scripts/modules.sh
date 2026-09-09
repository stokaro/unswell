#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
operation=${1:?test or tidy required}
shift
while read -r directory role; do
  case "$operation:$role" in
    test:runtime | test:consumer) (cd "$directory" && go test -count=1 "$@" ./...) ;;
    test:tools) : ;; # Dependency-only module; tidy and pinned tool execution cover it.
    tidy:*) (cd "$directory" && go mod tidy "$@") ;;
    *)
      printf 'Unsupported module operation: %s:%s\n' "$operation" "$role" >&2
      exit 1
      ;;
  esac
done <.gomodules
