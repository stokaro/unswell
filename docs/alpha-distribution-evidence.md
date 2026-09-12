# Alpha distribution evidence

Verified on September 12, 2026. This record concerns the published
[v0.1.0-alpha.3](https://github.com/stokaro/unswell/releases/tag/v0.1.0-alpha.3),
source `2a2a6d441534cbc4be15d298898345d1a529c8b5`. Its tag signature verifies.
The alpha.2 tag exists at `208c7f50a555600670f6d542a3f31bd2ba0aa9ef`, and its
release run stopped in the mirror job, so no alpha.2 asset was published. The
alpha.1 record lives in the git history of this file and in the
[alpha.1 audit](release/v0.1.0-alpha.1-audit.json). Changes made later on main
are not part of these binaries.

## Published installation paths

| Path | Verified evidence |
| --- | --- |
| CLI archives | [The release audit](release/v0.1.0-alpha.3-audit.json) downloaded all six archives, both checksum files, and six SBOMs again. All twelve digests match GitHub's release records, and every binary rebuilt byte for byte from the tagged source on another machine. |
| Public Action | The three published-archive consumers in the [release run](https://github.com/stokaro/unswell/actions/runs/34686065825) passed on Linux, macOS, and Windows. Their JSON reports name the release source, version `0.1.0-alpha.3`, and a complete pass on 812 documents. |
| CLI and MCP images | Both public image checks in that release run passed on AMD64 and ARM64. The downloaded CLI and MCP reports name the release source, with 812 documents and the expected failure and control outcomes. |
| MCP Registry | The registry job's saved response shows `io.github.stokaro/unswell` version `0.1.0-alpha.3` as `active` and latest, published at 09:48 UTC, using the GHCR MCP image over stdio. |
| Homebrew | The release dispatched [tap PR #4](https://github.com/stokaro/homebrew-unswell/pull/4) with the versioned archive URLs and hashes. Its four install jobs, on Linux AMD64 and ARM64 and on macOS Intel and ARM64, passed at `3d688a3254d264f5e86607d17916cd98460649dd`. A maintainer merged it, and the tap's [main CI](https://github.com/stokaro/homebrew-unswell/actions/runs/34687164090) passed the same four jobs on `dc996273da72d2ddfcf11849d028eb4b5d8d397b`. |
| Go module | [The module check](release/v0.1.0-alpha.3-module-check.json) built and tested the consumer against the published `v0.1.0-alpha.3` with no local replacement and listed no public package missing from it. |

The Action's ordinary main CI uses a CLI built from a pinned source revision.
Those tests do not prove release downloads. The published-archive consumers
above provide that evidence. The release also dispatched
[Action PR #6](https://github.com/stokaro/unswell-action/pull/6), which moves
the default CLI version to `0.1.0-alpha.3`; its test and consumer jobs passed
on the three platforms. A maintainer merged it, main CI passed on
`b7e5d5ce69d69ca458879c938726ec0870457b35`, and the publish workflow created
the tag `v0.1.0-alpha.3` at that commit without moving `v0.1.0-alpha.1`. A
second request for the same version on each repository reported that the
target already selected 0.1.0-alpha.3 and opened nothing.

Anonymous reads confirm byte-identical GHCR and Docker Hub indexes for both
images, including Linux AMD64/ARM64 and their attestation manifests:

| Image | Published index digest |
| --- | --- |
| `ghcr.io/stokaro/unswell:0.1.0-alpha.3` and `cabyrc/unswell:0.1.0-alpha.3` | `sha256:d29ea808dd936762aa9f61d29890041e46fca069707e046c23f8a0ee19e68699` |
| `ghcr.io/stokaro/unswell-mcp:0.1.0-alpha.3` and `cabyrc/unswell-mcp:0.1.0-alpha.3` | `sha256:3365cb0570c30a57f165140f69256f066e46770cadec022ac8c5362a5729305c` |

The six SBOMs of this release declare MIT for the main component, so no
erratum accompanies them. The alpha.1 correction stays described in
[SBOM and notices](sbom.md).

## Remaining acceptance

The publishing app created both distribution pull requests, their required
checks ran, a maintainer merged each, a repeat request opened nothing, and the
Action tag followed. That is one full cycle of
[#66](https://github.com/stokaro/unswell/issues/66); the next release shows it
repeating.

This record does not establish stable-rule precision, a qualified probability
model, or final product acceptance. Those requirements remain in the
[roadmap](roadmap.md). Race detection, active fuzzing, and coverage collection for
ongoing alpha development remain deferred to final task
[#123](https://github.com/stokaro/unswell/issues/123).
