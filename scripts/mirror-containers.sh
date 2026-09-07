#!/usr/bin/env bash
set +x
set -euo pipefail
cd "$(dirname "$0")/.."

if [[ $# != 3 ]]; then
  printf 'Usage: %s DOCKER_CONTEXT VERSION NAMESPACE <token\n' "$0" >&2
  exit 2
fi
mirror_context=$1
version=${2#v}
namespace=$3
bash scripts/check-image-mirrors.sh --validate "$version" "$namespace"
[[ "$mirror_context" =~ ^[A-Za-z0-9][A-Za-z0-9_.-]*$ ]] || {
  printf 'Invalid Docker context\n' >&2
  exit 2
}

# Skopeo 1.22.2 copies every manifest without rebuilding the release.
mirror_tool=quay.io/skopeo/stable@sha256:e5d9c4af8ec327785c7ca938d1e4f8452c6a05014850e58e2ff9456899ebd97c
mirror_temporary=$(mktemp -d)
mirror_container="unswell-mirror-$(basename "$mirror_temporary")"
mirror_tool_created=false
cleanup() {
  if docker --context "$mirror_context" container inspect "$mirror_container" >/dev/null 2>&1; then
    docker --context "$mirror_context" rm -f "$mirror_container"
  fi
  if [[ "$mirror_tool_created" == true ]] && docker --context "$mirror_context" image inspect "$mirror_tool" >/dev/null 2>&1; then
    docker --context "$mirror_context" image rm "$mirror_tool"
  fi
  rm -rf "$mirror_temporary"
}
trap cleanup EXIT
docker --context "$mirror_context" version >/dev/null </dev/null
if ! docker --context "$mirror_context" image inspect "$mirror_tool" >/dev/null 2>&1 </dev/null; then mirror_tool_created=true; fi
docker --context "$mirror_context" pull "$mirror_tool" </dev/null
docker --context "$mirror_context" create -i --name "$mirror_container" --entrypoint /bin/bash \
  --tmpfs /run/auth:rw,noexec,nosuid,size=1m "$mirror_tool" \
  /run/copy-containers.sh "$version" "$namespace" >/dev/null </dev/null
docker --context "$mirror_context" cp scripts/copy-containers.sh "$mirror_container:/run/copy-containers.sh" </dev/null
docker --context "$mirror_context" start --attach --interactive "$mirror_container"
mirror_exit=$(docker --context "$mirror_context" inspect --format '{{.State.ExitCode}}' "$mirror_container" </dev/null)
[[ "$mirror_exit" == 0 ]] || exit "$mirror_exit"
bash scripts/check-image-mirrors.sh "$version" "$namespace"
