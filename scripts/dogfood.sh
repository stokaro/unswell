#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p artifacts/dogfood
bin/unswell check . --config .unswell.yaml --include-source \
  --feature prose-words --feature type-token-ratio --feature activation/readability.long-paragraph \
  --feature activation/filler.announced-importance --feature activation/syntax.noun-stack \
  --report text:- \
  --report json:artifacts/dogfood/result.json \
  --report sarif:artifacts/dogfood/result.sarif \
  --report html:artifacts/dogfood/result.html \
  --report markdown:artifacts/dogfood/result.md

# Exercise this exact policy through the built CLI. A disabled gate, ignored
# configuration, or an operational error must not make the negative probe pass.
negative_prose='Certainly! The client opens connections.'
for context in markdown comment string csharp yaml; do
  case "$context" in
    markdown)
      probe_source=$negative_prose
      filename=dogfood-negative.md
      ;;
    comment)
      probe_source="# $negative_prose"
      filename=dogfood-negative.sh
      ;;
    string)
      probe_source="message=\"$negative_prose\""
      filename=dogfood-negative.sh
      ;;
    csharp)
      probe_source="class Sample { string message = \"$negative_prose\"; }"
      filename=dogfood-negative.cs
      ;;
    yaml)
      probe_source="message: \"$negative_prose\""
      filename=dogfood-negative.yaml
      ;;
    *)
      printf 'Unknown dogfood context: %s\n' "$context" >&2
      exit 1
      ;;
  esac
  status=0
  printf '%s\n' "$probe_source" |
    bin/unswell check --stdin --filename "$filename" --config .unswell.yaml \
      --report "json:artifacts/dogfood/negative-$context.json" || status=$?
  if [[ "$status" != 1 ]]; then
    printf 'Dogfood %s probe returned %s; expected a complete policy failure (1).\n' "$context" "$status" >&2
    exit 1
  fi
  printf 'Dogfood policy rejected the %s probe with exit code 1.\n' "$context"
done
