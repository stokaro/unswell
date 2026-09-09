# Alpha distribution evidence

Verified on September 9, 2026. This record concerns the published
[v0.1.0-alpha.1](https://github.com/stokaro/unswell/releases/tag/v0.1.0-alpha.1),
source `247fe6f673cb8cad08850c05694a874b71e538e7`. Its tag signature verifies.
Changes made later on main are not part of these binaries.

## Published installation paths

| Path | Verified evidence |
| --- | --- |
| Homebrew | [Tap main CI](https://github.com/stokaro/homebrew-unswell/actions/runs/34148425782) passed on Linux AMD64/ARM64 and macOS Intel/ARM64 at `ccaba48e5a497d8b176a8dd49e01ecc747e32f23`. |
| CLI archives | All six archives, both checksum files, twelve SBOM files, and the erratum were downloaded again. All 21 asset digests match GitHub's release records. Both checksum files match their named assets. |
| Public Action | The three published-archive consumers in the [release run](https://github.com/stokaro/unswell/actions/runs/34129248148) passed on Linux, macOS, and Windows. Their JSON reports identify the release source and a complete pass on 115 documents. |
| CLI and MCP images | Both public image checks in that release run passed. Downloaded CLI/MCP reports identify the release source, with 115 documents and the expected failure/control outcomes. |
| MCP Registry | A fresh public API read confirms active `io.github.stokaro/unswell` version `0.1.0-alpha.1`, using the GHCR MCP image over stdio. |

The tap's published formula uses versioned archive URLs and their verified hashes.
The four installation logs show the release formula being installed. Its tests
require clean prose to pass, a policy violation to return 1, and malformed C# to
return 2. The installed tool's retained reports agree on seven tap documents,
zero findings, version `0.1.0-alpha.1`, and the release source commit.
The formula installs dependency notices alongside the executable.

The Action's ordinary main CI uses a CLI built from a pinned source revision.
Those tests do not prove release downloads. The separate published-archive
consumers above provide that evidence.

Anonymous reads also confirm byte-identical GHCR and Docker Hub indexes for both
images, including Linux AMD64/ARM64 and their attestation manifests:

| Image | Published index digest |
| --- | --- |
| `ghcr.io/stokaro/unswell:0.1.0-alpha.1` and `cabyrc/unswell:0.1.0-alpha.1` | `sha256:40d95b3c853da4aa444defa495e94d5dc6605eca166a6c4396fd5bbd2ffe532a` |
| `ghcr.io/stokaro/unswell-mcp:0.1.0-alpha.1` and `cabyrc/unswell-mcp:0.1.0-alpha.1` | `sha256:80dfd8133833c2aac97d16cbab2ea9b08da9b60538262de39ff513f55fcefd16` |

Original archives and SBOMs remain unchanged. The six corrected SBOMs and
`SBOM-ERRATUM.md` preserve the correction described in
[#48](https://github.com/stokaro/unswell/issues/48).

## Remaining acceptance

The formula and Action documentation PRs that previously blocked distribution
have merged. This does not complete automatic updates:
[tap PR #2](https://github.com/stokaro/homebrew-unswell/pull/2) and
[Action PR #2](https://github.com/stokaro/unswell-action/pull/2) remain open.
[#66](https://github.com/stokaro/unswell/issues/66) still requires a real publishing
app cycle: PR creation, required checks, merge, repeat request, and Action tag
publication. The current tap workflow prepares a branch for maintainer review.

This record does not establish stable-rule precision, a qualified probability
model, or final product acceptance. Those requirements remain in the
[roadmap](roadmap.md). Race detection, active fuzzing, and coverage collection for
ongoing alpha development remain deferred to final task
[#123](https://github.com/stokaro/unswell/issues/123).
