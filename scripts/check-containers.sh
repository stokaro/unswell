#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
context=${1:?Docker context required}
cli_image=${2:?CLI image required}
mcp_image=${3:?MCP image required}
docker_command=(docker --context "$context")
temporary=$(mktemp -d)
prefix="unswell-check-$(basename "$temporary")"
volume="$prefix-source"
server="$prefix-mcp"
containers=("$prefix-copy" "$prefix-init" "$prefix-cli" "$prefix-negative" "$server")
mkdir -p artifacts/containers

cleanup() {
  local status=$?
  trap - EXIT
  local container
  for container in "${containers[@]}"; do
    if "${docker_command[@]}" container inspect "$container" >/dev/null 2>&1; then
      "${docker_command[@]}" rm -f "$container" >&2 || status=1
    fi
  done
  if "${docker_command[@]}" volume inspect "$volume" >/dev/null 2>&1; then
    "${docker_command[@]}" volume rm "$volume" >&2 || status=1
  fi
  rm -rf "$temporary"
  exit "$status"
}
trap cleanup EXIT
"${docker_command[@]}" volume create "$volume" >/dev/null

# Transfer only tracked project files. Bind mounts would refer to the remote
# daemon's filesystem, so this uses a private volume for local and CI contexts.
git ls-files -z --cached | COPYFILE_DISABLE=1 tar --no-xattrs --no-acls --null -T - -cf - |
  "${docker_command[@]}" run --rm -i --name "$prefix-copy" --network none --user 0:0 \
    --mount "type=volume,source=$volume,target=/work" --entrypoint tar "$cli_image" -xf - -C /work
"${docker_command[@]}" run --rm --name "$prefix-init" --network none --user 0:0 \
  --mount "type=volume,source=$volume,target=/work" --entrypoint /bin/sh "$cli_image" \
  -c 'git init -q && git add .'

"${docker_command[@]}" run --rm --name "$prefix-cli" --read-only --network none \
  --mount "type=volume,source=$volume,target=/work,readonly" "$cli_image" \
  check . --config .unswell.yaml --include-source --report json:- >artifacts/containers/cli.json

negative_prose='Certainly! The client opens connections.'
status=0
printf '%s\n' "$negative_prose" |
  "${docker_command[@]}" run --rm -i --name "$prefix-negative" --read-only --network none \
    --mount "type=volume,source=$volume,target=/work,readonly" "$cli_image" \
    check --stdin --filename container-negative.md --config .unswell.yaml --report json:- \
    >artifacts/containers/negative.json || status=$?
if [[ "$status" != 1 ]]; then
  printf 'CLI container returned %s for a policy violation; expected 1.\n' "$status" >&2
  exit 1
fi

(cd mcp && go run ./cmd/mcp-selfcheck --expected ../artifacts/containers/cli.json \
  --output ../artifacts/containers/mcp.json -- \
  docker --context "$context" run --rm -i --name "$server" --read-only --network none \
  --mount "type=volume,source=$volume,target=/work,readonly" "$mcp_image" --config /work/.unswell.yaml)
printf 'CLI and MCP container checks passed with read-only inputs and no runtime network.\n'
