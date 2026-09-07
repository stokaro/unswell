#!/usr/bin/env bash
set -euo pipefail
script_directory=$(dirname "$0")
cd "$script_directory/.."
version=$(cat .shellcheck-version)
platform=$(uname -s)
architecture=$(uname -m)
case "$version:$platform:$architecture" in
  0.11.0:Linux:x86_64)
    target=linux.x86_64
    digest=b7af85e41cc99489dcc21d66c6d5f3685138f06d34651e6d34b42ec6d54fe6f6
    ;;
  0.11.0:Linux:aarch64)
    target=linux.aarch64
    digest=68a8133197a50beb8803f8d42f9908d1af1c5540d4bb05fdfca8c1fa47decefc
    ;;
  0.11.0:Darwin:arm64)
    target=darwin.aarch64
    digest=339b930feb1ea764467013cc1f72d09cd6b869ebf1013296ba9055ab2ffbd26f
    ;;
  0.11.0:Darwin:x86_64)
    target=darwin.x86_64
    digest=c2c15e08df0e8fbc374c335b230a7ee958c313fa5714817a59aa59f1aa594f51
    ;;
  *)
    printf 'No pinned ShellCheck archive for %s/%s/%s\n' "$version" "$platform" "$architecture" >&2
    exit 1
    ;;
esac
temporary=$(mktemp -d)
trap 'rm -rf "$temporary"' EXIT
curl --fail --location --silent --show-error --retry 3 \
  "https://github.com/koalaman/shellcheck/releases/download/v$version/shellcheck-v$version.$target.tar.gz" \
  --output "$temporary/shellcheck.tar.gz"
printf '%s  %s\n' "$digest" "$temporary/shellcheck.tar.gz" | shasum -a 256 --check --status
tar -xzf "$temporary/shellcheck.tar.gz" -C "$temporary" "shellcheck-v$version/shellcheck"
mkdir -p bin
install -m 755 "$temporary/shellcheck-v$version/shellcheck" bin/shellcheck
bin/shellcheck --version
