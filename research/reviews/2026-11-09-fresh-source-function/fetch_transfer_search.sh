#!/usr/bin/env bash
set -euo pipefail

output_dir=${1:?pass a new output directory}
if [[ -e "$output_dir" ]]; then
  echo "output directory already exists" >&2
  exit 1
fi
mkdir -p "$output_dir"

gh api --method GET search/code -f q='"next generation" filename:README.md size:4000..20000' -f per_page=30 -f sort=indexed -f order=desc >"$output_dir/01.json"
gh api --method GET search/code -f q='"powerful and flexible" filename:README.md size:4000..20000' -f per_page=30 -f sort=indexed -f order=desc >"$output_dir/02.json"
gh api --method GET search/code -f q='"the following table" filename:README.md size:4000..20000' -f per_page=30 -f sort=indexed -f order=desc >"$output_dir/03.json"
