#!/usr/bin/env bash
# Run inside the pinned Skopeo container through mirror-containers.sh.
set +x
set -euo pipefail
umask 077
version=$1
namespace=$2
printf '{"auths":{}}\n' >/run/auth/anonymous.json
skopeo login --authfile /run/auth/auth.json --username "$namespace" --password-stdin docker.io
for name in unswell unswell-mcp; do
  skopeo inspect --raw --authfile /run/auth/anonymous.json "docker://ghcr.io/stokaro/$name:$version" >/run/auth/source.json
  digest=$(sha256sum /run/auth/source.json)
  digest=${digest%% *}
  skopeo copy --all --preserve-digests --authfile /run/auth/auth.json \
    "docker://ghcr.io/stokaro/$name@sha256:$digest" "docker://docker.io/$namespace/$name:$version"
done
skopeo logout --authfile /run/auth/auth.json docker.io
