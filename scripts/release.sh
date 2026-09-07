#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
version=$(sed -n 's/^const Version = "\(.*\)"/\1/p' result.go)
[[ -n "$version" ]] || {
  printf 'Missing release version\n' >&2
  exit 1
}
commit=$(git rev-parse HEAD)
release_go=$(awk '/^toolchain / {print $2}' tools/go.mod)
compiler=$(go env GOVERSION)
if [[ "$compiler" != "$release_go" ]]; then
  printf 'Release builds require %s from tools/go.mod.\n' "$release_go" >&2
  exit 1
fi
if compgen -G 'dist/*' >/dev/null; then
  printf 'dist must be empty before a release build; preserve or remove old artifacts first.\n' >&2
  exit 1
fi
if [[ -n "${RELEASE_TAG:-}" && "$RELEASE_TAG" != "v$version" ]]; then
  printf 'Release tag and source version disagree\n' >&2
  exit 1
fi
sbom=$(cd tools && go tool -n cyclonedx-gomod)
mkdir -p dist
temporary=$(mktemp -d)
trap 'rm -rf "$temporary"' EXIT
for platform in linux darwin windows; do
  for architecture in amd64 arm64; do
    name="unswell_${version}_${platform}_${architecture}"
    package="$temporary/$name"
    mkdir "$package"
    binary=unswell
    if [[ "$platform" == windows ]]; then binary=unswell.exe; fi
    CGO_ENABLED=0 GOOS="$platform" GOARCH="$architecture" go build -trimpath -buildvcs=false \
      -ldflags "-s -w -X github.com/stokaro/unswell.BuildCommit=$commit" \
      -o "$package/$binary" ./cmd/unswell
    CGO_ENABLED=0 GOOS="$platform" GOARCH="$architecture" "$sbom" app -json -licenses -std \
      -noserial -notimestamp -main cmd/unswell -output "$temporary/$name.raw.cdx.json" .
    bash scripts/prepare-sbom.sh "$temporary/$name.raw.cdx.json" "dist/$name.cdx.json"
    cp LICENSE THIRD_PARTY_NOTICES.md "$package/"
    cp -R licenses "$package/"
    if [[ "$platform" == windows ]]; then
      destination="$PWD/dist/$name.zip"
      (cd "$temporary" && zip -qr "$destination" "$name")
    else
      tar -czf "dist/$name.tar.gz" -C "$temporary" "$name"
    fi
  done
done
(cd dist && shasum -a 256 ./*.tar.gz ./*.zip ./*.cdx.json >SHA256SUMS)
