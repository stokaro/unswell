#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

valid_parameters() {
  [[ "$1" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z]+(\.[0-9A-Za-z]+)*)?$ && ${#1} -le 80 &&
    "$2" =~ ^[a-z0-9][a-z0-9_-]*$ && ${#2} -le 64 ]]
}

same_images() {
  [[ -s "$1" && -s "$2" ]] || return 1
  cmp -s "$1" "$2" || return 1
  jq --exit-status '
    .schemaVersion == 2 and
    .mediaType == "application/vnd.oci.image.index.v1+json" and
    ([.manifests[] | select(.platform.os == "linux") | .platform.architecture] | sort) == ["amd64", "arm64"] and
    ([.manifests[] | select(.annotations["vnd.docker.reference.type"] == "attestation-manifest")] | length) == 2
  ' "$1" >/dev/null
}

self_test() {
  local status
  mirror_test_directory=$(mktemp -d)
  local temporary="$mirror_test_directory"
  trap 'rm -rf "$mirror_test_directory"' EXIT
  jq -n '{schemaVersion: 2, mediaType: "application/vnd.oci.image.index.v1+json", manifests: [
    {platform: {os: "linux", architecture: "amd64"}},
    {platform: {os: "linux", architecture: "arm64"}},
    {annotations: {"vnd.docker.reference.type": "attestation-manifest"}},
    {annotations: {"vnd.docker.reference.type": "attestation-manifest"}}
  ]}' >"$temporary/source.json"
  cp "$temporary/source.json" "$temporary/mirror.json"
  same_images "$temporary/source.json" "$temporary/mirror.json"
  for mutation in different missing malformed architecture attestation; do
    cp "$temporary/source.json" "$temporary/mirror.json"
    case "$mutation" in
      different) jq '.extra = true' "$temporary/source.json" >"$temporary/mirror.json" ;;
      missing) rm "$temporary/mirror.json" ;;
      malformed) printf 'invalid JSON\n' >"$temporary/mirror.json" ;;
      architecture) jq 'del(.manifests[1])' "$temporary/source.json" >"$temporary/mirror.json" ;;
      attestation) jq 'del(.manifests[3])' "$temporary/source.json" >"$temporary/mirror.json" ;;
    esac
    local source="$temporary/source.json"
    # Equal malformed or incomplete images must also fail the gate.
    if [[ "$mutation" == malformed || "$mutation" == architecture || "$mutation" == attestation ]]; then
      source="$temporary/mirror.json"
    fi
    status=0
    # same_images checks every command failure explicitly.
    # shellcheck disable=SC2310
    same_images "$source" "$temporary/mirror.json" 2>/dev/null || status=$?
    [[ "$status" != 0 ]] || {
      printf 'Mirror gate accepted %s images\n' "$mutation" >&2
      return 1
    }
  done
  valid_parameters 0.1.0-alpha.1 cabyrc
  for version in latest ../version 0.1; do
    # valid_parameters is a single Boolean expression.
    # shellcheck disable=SC2310
    if valid_parameters "$version" cabyrc; then return 1; fi
  done
  # valid_parameters is a single Boolean expression.
  # shellcheck disable=SC2310
  if valid_parameters 0.1.0 '../namespace'; then return 1; fi
  printf 'Mirror gate accepted equal images and rejected five broken mirrors and invalid identifiers.\n'
}

fetch_manifest() {
  local authentication=$1 manifest=$2 destination=$3 token
  token=$(curl --disable --fail --silent --show-error --retry 2 --connect-timeout 10 --max-time 60 \
    "$authentication" | jq --raw-output --exit-status '.token | strings | select(length > 0)')
  [[ "$token" =~ ^[A-Za-z0-9._=+/-]+$ ]] || {
    printf 'Invalid anonymous registry token\n' >&2
    return 1
  }
  printf 'header = "Authorization: Bearer %s"\n' "$token" |
    curl --disable --config - --fail --silent --show-error --retry 2 --connect-timeout 10 --max-time 60 \
      --header 'Accept: application/vnd.oci.image.index.v1+json' --output "$destination" "$manifest"
}

check_mirrors() {
  local version=$1 namespace=$2 directory name digest
  directory="artifacts/mirrors/$version/$namespace"
  mkdir -p "$directory"
  for name in unswell unswell-mcp; do
    fetch_manifest "https://ghcr.io/token?service=ghcr.io&scope=repository:stokaro/$name:pull" \
      "https://ghcr.io/v2/stokaro/$name/manifests/$version" "$directory/$name-ghcr.json"
    fetch_manifest "https://auth.docker.io/token?service=registry.docker.io&scope=repository:$namespace/$name:pull" \
      "https://registry-1.docker.io/v2/$namespace/$name/manifests/$version" "$directory/$name-hub.json"
    same_images "$directory/$name-ghcr.json" "$directory/$name-hub.json"
    digest=$(shasum -a 256 "$directory/$name-ghcr.json")
    printf '%s:%s sha256:%s\n' "$namespace/$name" "$version" "${digest%% *}"
  done
}

if [[ "${1:-}" == --self-test && $# == 1 ]]; then
  self_test
  exit 0
fi
validate_only=false
if [[ "${1:-}" == --validate ]]; then
  validate_only=true
  shift
fi
if [[ $# != 2 ]]; then
  printf 'Usage: %s [--validate] VERSION NAMESPACE | --self-test\n' "$0" >&2
  exit 2
fi
version=${1#v}
# valid_parameters is a single Boolean expression.
# shellcheck disable=SC2310
if ! valid_parameters "$version" "$2"; then
  printf 'Invalid release version or Docker Hub namespace\n' >&2
  exit 2
fi
if [[ "$validate_only" == false ]]; then check_mirrors "$version" "$2"; fi
