#!/usr/bin/env bash
# Download pinned inputs and measure them on an explicitly selected Linux host.
set -euo pipefail
cd "$(dirname "$0")/../../.."
root=$PWD
context=${1:?Usage: reproduce.sh DOCKER_CONTEXT NEW_OUTPUT_DIRECTORY}
output=${2:?Usage: reproduce.sh DOCKER_CONTEXT NEW_OUTPUT_DIRECTORY}
case "$context" in
  diabolocom | remote-dev-container | vcluster-mac) ;;
  *)
    printf 'Choose an explicit remote Linux Docker context.\n' >&2
    exit 2
    ;;
esac
mkdir "$output"
output=$(cd "$output" && pwd)
temporary=$(mktemp -d)
name=unswell-perf219-$(date +%s)-$$
image=$name:local
created=false
built=false
cleanup() {
  local status=$?
  if [[ "$created" == true ]]; then
    docker --context "$context" cp "$name:/benchmark/results/." "$output/" || status=2
    docker --context "$context" rm -f "$name" || status=2
  fi
  if [[ "$built" == true ]]; then
    docker --context "$context" image rm "$image" || status=2
  fi
  rm -rf "$temporary"
  exit "$status"
}
trap cleanup EXIT
sha() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -c1-64
  else
    shasum -a 256 "$1" | cut -c1-64
  fi
}
asset=unswell_0.1.0-alpha.3_linux_amd64.tar.gz
curl -fsSL --retry 3 "https://github.com/stokaro/unswell/releases/download/v0.1.0-alpha.3/$asset" -o "$temporary/$asset"
actual=$(sha "$temporary/$asset")
test "$actual" = f24fb5ad49945b4898888470aee3a4af44b8907ee23978d8a06e022489dd4451
tar -xzf "$temporary/$asset" -C "$temporary"
mkdir -p "$temporary/payload/scripts" "$temporary/payload/sources"
cp "$temporary/unswell_0.1.0-alpha.3_linux_amd64/unswell" "$temporary/payload/unswell"
actual=$(sha "$temporary/payload/unswell")
test "$actual" = f165e953d41a09c2ab0490ff47f5e3d5729adfdaca8431f32faa0b56b9bc2e82
cp scripts/measure-performance.sh scripts/summarize-performance.py "$temporary/payload/scripts/"
sed "s/diabolocom-alpha3-/$context-alpha3-/" docs/performance/alpha3/run.sh >"$temporary/payload/run.sh"
jq -r '.[] | [.name, .repository, .commit, .archive_sha256] | @tsv' docs/performance/alpha3/inputs.json >"$temporary/inputs.tsv"
while IFS=$'\t' read -r corpus repository commit digest; do
  archive=$temporary/$corpus.tar.gz
  curl -fsSL --retry 3 "https://codeload.github.com/$repository/tar.gz/$commit" -o "$archive"
  actual=$(sha "$archive")
  test "$actual" = "$digest"
  mkdir "$temporary/payload/sources/$corpus"
  tar -xzf "$archive" --strip-components=1 -C "$temporary/payload/sources/$corpus"
done <"$temporary/inputs.tsv"
docker --context "$context" build --tag "$image" docs/performance/alpha3
built=true
docker --context "$context" create --name "$name" --cpus 2 --memory 512m --memory-swap 512m \
  --network none --env GOMAXPROCS=2 "$image"
created=true
docker --context "$context" cp "$temporary/payload/." "$name:/benchmark/"
docker --context "$context" inspect "$name" --format '{{json .HostConfig}}' >"$output/host-config.json"
docker --context "$context" image inspect "$image" --format '{{json .}}' >"$output/image.json"
cp "$root/docs/performance/alpha3/inputs.json" "$output/inputs.json"
docker --context "$context" start "$name"
docker --context "$context" exec "$name" bash /benchmark/run.sh
