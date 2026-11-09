#!/usr/bin/env bash
set -euo pipefail

search_dir=${1:?usage: fetch-search.sh SEARCH_DIR}
if [[ -e $search_dir ]]; then
  echo "search directory already exists" >&2
  exit 1
fi
mkdir -p "$search_dir"

gh api --method GET search/code \
  -f q='"it is worth noting" filename:README.md size:4000..20000' \
  -f per_page=30 -f sort=indexed -f order=desc >"$search_dir/01.json"
gh api --method GET search/code \
  -f q='"powerful and flexible" filename:README.md size:4000..20000' \
  -f per_page=30 -f sort=indexed -f order=desc >"$search_dir/02.json"
gh api --method GET search/code \
  -f q='"the following table" filename:README.md size:4000..20000' \
  -f per_page=30 -f sort=indexed -f order=desc >"$search_dir/03.json"
