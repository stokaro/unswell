#!/usr/bin/env bash
# Build the release binaries twice from the same source and compare digests.
# Equal digests show that this toolchain and these flags do not embed the build
# environment. They do not prove that a different compiler reproduces them.
set -euo pipefail
script_directory=$(dirname "$0")
cd "$script_directory/.."
root=$PWD

commit=""
output=""
platforms="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --commit)
      commit=$2
      shift 2
      ;;
    --output)
      output=$2
      shift 2
      ;;
    --platforms)
      platforms=$2
      shift 2
      ;;
    *)
      printf 'Usage: %s [--commit SHA] [--output FILE]\n' "$0" >&2
      exit 2
      ;;
  esac
done
if [[ -z "$commit" ]]; then
  commit=$(git rev-parse HEAD)
fi

workspace=$(mktemp -d)
trap 'rm -rf "$workspace"' EXIT

# build writes one release binary with the exact flags scripts/release.sh uses.
# The toolchain comes from the environment, so the caller decides which compiler
# the check reports; GOTOOLCHAIN=go1.25.0 matches the released builds.
build() {
  local platform=$1 architecture=$2 destination=$3 binary=unswell
  if [[ "$platform" == windows ]]; then binary=unswell.exe; fi
  mkdir -p "$destination"
  CGO_ENABLED=0 GOOS="$platform" GOARCH="$architecture" go build -trimpath -buildvcs=false \
    -ldflags "-s -w -X github.com/stokaro/unswell.BuildCommit=$commit" \
    -o "$destination/$binary" ./cmd/unswell
  printf '%s' "$destination/$binary"
}

digest() {
  local file=$1 value
  if command -v sha256sum >/dev/null 2>&1; then
    value=$(sha256sum "$file")
  else
    value=$(shasum -a 256 "$file")
  fi
  printf '%s' "${value%% *}"
}

entries=()
mismatched=0
for target in $platforms; do
  platform=${target%%/*}
  architecture=${target##*/}
  first=$(build "$platform" "$architecture" "$workspace/first/${platform}_${architecture}")
  second=$(build "$platform" "$architecture" "$workspace/second/${platform}_${architecture}")
  first_digest=$(digest "$first")
  second_digest=$(digest "$second")
  equal=true
  if [[ "$first_digest" != "$second_digest" ]]; then
    equal=false
    mismatched=$((mismatched + 1))
  fi
  entries+=("$(printf '{"platform": "%s", "architecture": "%s", "sha256": "%s", "repeated_sha256": "%s", "equal": %s}' \
    "$platform" "$architecture" "$first_digest" "$second_digest" "$equal")")
done

builds=$(
  IFS=,
  printf '[%s]' "${entries[*]}"
)
go_version=$(go version)
host_goos=$(go env GOHOSTOS)
host_goarch=$(go env GOHOSTARCH)
report=$(
  cat <<JSON
{
  "format": "unswell-reproducible-build-check-v1",
  "runtime_commit": "$commit",
  "go_version": "$go_version",
  "cgo_enabled": false,
  "flags": "-trimpath -buildvcs=false -ldflags '-s -w -X github.com/stokaro/unswell.BuildCommit=<commit>'",
  "host_goos": "$host_goos",
  "host_goarch": "$host_goarch",
  "builds": $builds,
  "mismatched": $mismatched,
  "scope": "Two builds of one source, one toolchain, one host. It makes no claim about other compilers."
}
JSON
)
if [[ -n "$output" ]]; then
  printf '%s\n' "$report" >"$output"
fi
printf '%s\n' "$report"
if ((mismatched != 0)); then
  printf 'Repeated builds produced %s different binaries.\n' "$mismatched" >&2
  exit 1
fi
printf 'Every requested release binary reproduced byte for byte from %s.\n' "$root" >&2
