#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
root=$PWD
golangci=$(cd tools && go tool -n golangci-lint)
qtlint=$(cd tools && go tool -n qtlint)
nolintguard=$(cd tools && go tool -n nolintguard)
"$golangci" config verify --config "$root/.golangci.yml"
while read -r directory role; do
  if [[ "$role" == tools ]]; then continue; fi
  (
    cd "$directory"
    "$golangci" run --config "$root/.golangci.yml"
    "$qtlint" -require-qt-c-receiver -require-data-rows -require-testing-run ./...
    "$nolintguard" -require-justification -forbidden-linters staticcheck,unused ./...
  )
done <.gomodules
