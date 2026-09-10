#!/usr/bin/env bash
# Audit release artifacts: archive digests, bundled notices, build records,
# per-platform SBOMs, and optionally whether each published binary rebuilds byte
# for byte from source. The assets are checked against their own manifest and
# records; that the source itself is correct is a separate question.
set -euo pipefail
script_directory=$(dirname "$0")
cd "$script_directory/.."

tag=""
dist=""
output=""
rebuild=false
rebuild_commit=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --tag)
      tag=$2
      shift 2
      ;;
    --dist)
      dist=$2
      shift 2
      ;;
    --output)
      output=$2
      shift 2
      ;;
    --rebuild)
      rebuild=true
      shift
      ;;
    --commit)
      rebuild_commit=$2
      shift 2
      ;;
    *)
      printf 'Usage: %s [--tag TAG | --dist DIR] [--rebuild --commit SHA] [--output FILE]\n' "$0" >&2
      exit 2
      ;;
  esac
done
if [[ -n "$tag" && -n "$dist" ]]; then
  printf 'Select either a published tag or a local distribution directory.\n' >&2
  exit 2
fi
if [[ "$rebuild" == true && -z "$rebuild_commit" ]]; then
  printf 'Rebuilding needs the commit the release embeds: pass --commit.\n' >&2
  exit 2
fi

workspace=$(mktemp -d)
trap 'rm -rf "$workspace"' EXIT
assets=$workspace/assets
mkdir -p "$assets"

source_label=""
if [[ -n "$dist" ]]; then
  cp "$dist"/* "$assets/"
  source_label="local distribution $dist"
else
  if [[ -z "$tag" ]]; then
    tag=$(gh release view --repo stokaro/unswell --json tagName --jq .tagName)
  fi
  gh release download "$tag" --repo stokaro/unswell --dir "$assets"
  source_label="published release $tag"
fi
if [[ ! -f "$assets/SHA256SUMS" ]]; then
  printf 'The assets do not include SHA256SUMS.\n' >&2
  exit 1
fi

digest() {
  local value
  if command -v sha256sum >/dev/null 2>&1; then
    value=$(sha256sum "$1")
  else
    value=$(shasum -a 256 "$1")
  fi
  printf '%s' "${value%% *}"
}

# record_value reads one build setting from a "go version -m" record.
record_value() {
  printf '%s' "$1" | sed -n "s/.*$2=\([A-Za-z0-9_.]*\).*/\1/p" | head -1
}

# Every asset the manifest names must exist and match its recorded digest.
checked_digests=0
while read -r expected name; do
  name=${name#./}
  [[ -n "$name" ]] || continue
  if [[ ! -f "$assets/$name" ]]; then
    printf 'The manifest names a missing asset: %s\n' "$name" >&2
    exit 1
  fi
  actual=$(digest "$assets/$name")
  if [[ "$actual" != "$expected" ]]; then
    printf 'Digest mismatch for %s\n' "$name" >&2
    exit 1
  fi
  checked_digests=$((checked_digests + 1))
done <"$assets/SHA256SUMS"

host_goos=$(go env GOHOSTOS)
host_goarch=$(go env GOHOSTARCH)
archives=0
notices=0
records=0
reproduced=0
host_version=""
compilers=()
unreproduced=()
for archive in "$assets"/*.tar.gz "$assets"/*.zip; do
  [[ -e "$archive" ]] || continue
  archives=$((archives + 1))
  name=$(basename "$archive")
  base=${name%.tar.gz}
  base=${base%.zip}
  extracted=$workspace/extracted/$base
  mkdir -p "$extracted"
  if [[ "$archive" == *.zip ]]; then
    unzip -q "$archive" -d "$extracted"
  else
    tar -xzf "$archive" -C "$extracted"
  fi
  package=$(find "$extracted" -maxdepth 1 -mindepth 1 -type d | head -1)
  for required in LICENSE THIRD_PARTY_NOTICES.md licenses; do
    if [[ ! -e "$package/$required" ]]; then
      printf '%s does not bundle %s\n' "$name" "$required" >&2
      exit 1
    fi
  done
  bundled=$(find "$package/licenses" -type f | wc -l)
  notices=$((notices + ${bundled// /}))
  binary=$package/unswell
  [[ -f "$binary" ]] || binary=$package/unswell.exe
  record=$(go version -m "$binary")
  for expected in "-trimpath=true" "CGO_ENABLED=0" "-compiler=gc"; do
    if [[ "$record" != *"$expected"* ]]; then
      printf '%s does not record %s\n' "$name" "$expected" >&2
      exit 1
    fi
  done
  records=$((records + 1))
  compilers+=("$(printf '%s' "$record" | sed -n '1s/.*: //p')")
  target=$(record_value "$record" GOOS)
  target_architecture=$(record_value "$record" GOARCH)
  if [[ "$name" != *"${target}_${target_architecture}"* ]]; then
    printf '%s records %s/%s, which its name does not name.\n' "$name" "$target" "$target_architecture" >&2
    exit 1
  fi
  if [[ "$rebuild" == true ]]; then
    rebuilt=$workspace/rebuilt/$base
    mkdir -p "$rebuilt"
    rebuilt_binary=$rebuilt/$(basename "$binary")
    CGO_ENABLED=0 GOOS=$target GOARCH=$target_architecture go build -trimpath -buildvcs=false \
      -ldflags "-s -w -X github.com/stokaro/unswell.BuildCommit=$rebuild_commit" \
      -o "$rebuilt_binary" ./cmd/unswell
    rebuilt_digest=$(digest "$rebuilt_binary")
    published_digest=$(digest "$binary")
    if [[ "$rebuilt_digest" == "$published_digest" ]]; then
      reproduced=$((reproduced + 1))
    else
      unreproduced+=("$name")
    fi
  fi
  if [[ "$name" == *"${host_goos}_${host_goarch}"* ]]; then
    host_version=$("$binary" --version)
  fi
done

unique_compilers=$(printf '%s\n' "${compilers[@]}" | sort -u | paste -sd, -)
sbom_count=$(find "$assets" -name '*.cdx.json' | wc -l)
sboms=${sbom_count// /}
if ((sboms < archives)); then
  printf 'Each archive needs a per-platform SBOM; found %s for %s archives.\n' "$sboms" "$archives" >&2
  exit 1
fi
project_licenses=$(python3 scripts/release-sbom-summary.py "$assets")
builder=$(go version)

report=$(
  cat <<JSON
{
  "format": "unswell-release-audit-v1",
  "source": "$source_label",
  "archives": $archives,
  "verified_digests": $checked_digests,
  "bundled_license_files": $notices,
  "build_records_checked": $records,
  "recorded_compilers": "$unique_compilers",
  "auditing_compiler": "$builder",
  "rebuilt_from_source": $rebuild,
  "rebuild_commit": "${rebuild_commit:-not requested}",
  "reproduced_binaries": $reproduced,
  "sboms": $sboms,
  "sbom_project_licenses": $project_licenses,
  "host_version": "${host_version:-not run on this host}",
  "scope": "Assets checked against their own manifest and records. Source correctness is separate."
}
JSON
)
if [[ -n "$output" ]]; then
  printf '%s\n' "$report" >"$output"
fi
printf '%s\n' "$report"
printf 'Audited %s: %s digests, %s archives, %s build records, %s SBOMs.\n' \
  "$source_label" "$checked_digests" "$archives" "$records" "$sboms" >&2
if [[ "$rebuild" == true && ${#unreproduced[@]} -ne 0 ]]; then
  printf 'These archives did not reproduce from source: %s\n' "${unreproduced[*]}" >&2
  exit 1
fi
