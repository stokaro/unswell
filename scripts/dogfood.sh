#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p artifacts/dogfood
bin/unswell check . --config .unswell.yaml --include-source \
  --report text:- \
  --report json:artifacts/dogfood/result.json \
  --report sarif:artifacts/dogfood/result.sarif \
  --report html:artifacts/dogfood/result.html \
  --report markdown:artifacts/dogfood/result.md

# Exercise this exact policy through the built CLI. A disabled gate, ignored
# configuration, or an operational error must not make the negative probe pass.
status=0
printf '%s\n' 'Certainly! The client opens connections.' |
  bin/unswell check --stdin --filename dogfood-negative.md --config .unswell.yaml \
    --report json:artifacts/dogfood/negative.json || status=$?
if [[ "$status" != 1 ]]; then
  printf 'Dogfood policy probe returned %s; expected a complete policy failure (1).\n' "$status" >&2
  exit 1
fi
printf 'Dogfood policy rejected the negative probe with exit code 1.\n'
