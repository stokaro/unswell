# Reproducible release binaries

A release publishes six executables. This page records how their reproducibility
is checked and what the check establishes.

## Protocol

`scripts/verify-reproducible-build.sh` builds each release target twice from the
same source, with the flags `scripts/release.sh` uses, and compares SHA-256
digests. The toolchain comes from the environment, so the caller decides which
compiler the check reports.

Released builds use the compiler pinned in `tools/go.mod`, which the release
workflow selects, and not the minimum compiler in `go.mod` that the native test
jobs use. Match the release by naming that toolchain:

```sh
GOTOOLCHAIN=go1.27.1 bash scripts/verify-reproducible-build.sh --output check.json
```

`make check` runs the same script for one target, which catches a regression
without building twelve binaries. `make reproducible` runs all six.

## The published alpha rebuilt from its tag

`scripts/verify-release-artifacts.sh --rebuild` rebuilds each published binary
from the source its tag names and compares the result with the archive. For
`v0.1.0-alpha.1` at `247fe6f673cb8cad08850c05694a874b71e538e7`, all six binaries
matched, rebuilt on a different machine from the one that released them. The
[recorded audit](release/v0.1.0-alpha.1-audit.json) holds that result together
with the digests, bundled notices, build records and SBOM licenses it checked.

This is the strongest form of the claim: a published artifact, its stated
source, and an independent machine produce the same bytes.

## Measured across two hosts

The six targets were also built from commit
`fee5caaf396e009ca01154caad8dd5a0c41fc5ef` on two hosts: macOS on ARM64, and
Debian 12 on AMD64 in a container. Both runs used Go 1.25.0, the lowest version
this project supports. All six digests matched across the two hosts, and each
host also repeated its own build byte for byte. These digests come from that
lower compiler, so they differ from the released binaries.

| Target | SHA-256 |
| --- | --- |
| linux/amd64 | `07431d083a073b73be525605039227e886a2e72d137afaebeba54d83e1eeaf40` |
| linux/arm64 | `2f3789fb8c2b0a910d41c46ed7ca3515b482466aa1eb2521893c34e1a52c860d` |
| darwin/amd64 | `c7f7122be9c4fa998c1bcbf0b8b72fc30d8bdb4b9d307623f7e550c03424fa26` |
| darwin/arm64 | `308af0cf05b518bb9369fbe9829ea0eaa0395d025c1fed65969f37acdaf18a93` |
| windows/amd64 | `14177be37a3d49944739b1315e52ca566371d1c2411b1d5fc19d44cf3d740ecb` |
| windows/arm64 | `38ee63a1ebcfe52be7c8399f085a6ccb64303921cb84f58002c08ac290d7b172` |

[The recorded checks](reproducible/) hold the full digests for both hosts and
the source commit they were built from. A digest depends on the Go sources and
on the commit string that `-ldflags` embeds, so passing `--commit` with the
recorded value reproduces these numbers even after documentation changes.

## What this establishes

The compiler and these flags do not embed the build host, its paths, or a
timestamp. Two different machines produced identical bytes from identical
source, so a third party with the same source and the same compiler can check a
published binary against its own build.

It does not show that a different compiler version reproduces those bytes. Go
does not promise that, which is why released builds pin the compiler and disable
automatic upgrades. The check also covers the executables alone; archives,
checksum manifests and SBOMs are produced by `scripts/release.sh` and audited
separately.
