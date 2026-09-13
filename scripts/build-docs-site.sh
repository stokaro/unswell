#!/usr/bin/env bash
# Assembles the documentation site that GitHub Pages serves.
#
# The published site is one directory per version plus a small root. Edge comes
# from the current tree. Each released tag is built from its own worktree, so a
# version keeps the pages it shipped with. The root holds versions.json, the
# apex redirect, the CNAME and robots.txt.
#
# The workflow calls this, and so can a maintainer. Running the same steps by
# hand is the only way to see what a deploy would publish before it publishes.
set -euo pipefail
cd "$(dirname "$0")/.."
root=$PWD
site=$root/docs/site

output=$root/_site
max_tags=10
edge_only=false

while [[ $# != 0 ]]; do
  case "$1" in
    --output)
      output=$2
      shift 2
      ;;
    --max-tags)
      max_tags=$2
      shift 2
      ;;
    --edge-only)
      edge_only=true
      shift
      ;;
    *)
      printf 'Usage: bash scripts/build-docs-site.sh [--output DIR] [--max-tags N] [--edge-only]\n' >&2
      exit 2
      ;;
  esac
done

if [[ ! -d "$site" ]]; then
  printf 'No documentation site at %s.\n' "$site" >&2
  exit 2
fi

# build_version renders one version into its own folder of the assembly. It
# takes the version name and the site directory to build from, which is the tag
# worktree for a release and the current tree for edge.
build_version() {
  local version=$1
  local source=$2
  DOCS_VERSION=$version ASTRO_TELEMETRY_DISABLED=1 npm --prefix "$source" run build
  mkdir -p "$output/$version"
  cp -R "$source/dist/." "$output/$version/"
}

rm -rf "$output"
mkdir -p "$output"

printf 'Building edge from the current tree.\n'
build_version edge "$site"

if [[ $edge_only == false ]]; then
  tags=$(mktemp)
  trap 'rm -f "$tags"' EXIT
  git tag --list 'v*' | sort -V | tail -n "$max_tags" >"$tags"
  while read -r tag; do
    [[ -n $tag ]] || continue
    printf 'Building %s.\n' "$tag"
    workdir=$(mktemp -d)
    git worktree add --force "$workdir" "refs/tags/$tag" >/dev/null
    # A tag older than this site has no docs/site to build. Its folder is then
    # rendered from the current tree under that version's base. The two cases do
    # not make the same promise, so the log says which one happened.
    if [[ -f "$workdir/docs/site/package.json" ]]; then
      DOCS_VERSION=$tag ASTRO_TELEMETRY_DISABLED=1 npm --prefix "$workdir/docs/site" ci
      build_version "$tag" "$workdir/docs/site"
    else
      printf 'Tag %s has no documentation site of its own; building it from the current tree.\n' "$tag"
      build_version "$tag" "$site"
    fi
    git worktree remove --force "$workdir" || true
  done <"$tags"
fi

node "$site/scripts/gen-versions.mjs" "$output"
default=$(node -e 'process.stdout.write(require(process.argv[1]).default)' "$output/versions.json")
node "$site/scripts/publish-root-assets.mjs" "$output" "$default"
node "$site/scripts/check-pages-root.mjs" --dist "$output"

printf 'Assembled %s with default version %s.\n' "$output" "$default"
