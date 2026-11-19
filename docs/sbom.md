# Release SBOMs

Each CLI archive has a CycloneDX document for its operating system and
architecture, with the selected Go modules and the compiler. For the main
component, the release pipeline declares the MIT license from the project's root
`LICENSE`; dependency license evidence comes from the pinned generator and is
retained separately.

`make check` verifies this metadata step (it also rejects a missing or wrong
project license). The preparation script requires a new output path and preserves
the dependency graph and its license evidence.

## Alpha.1 metadata erratum

The original `0.1.0-alpha.1` SBOMs attach `Apache-2.0` as the main component's
detected license evidence. The cause: the generator scanned the root `LICENSE` and
bundled dependency notices in `licenses/`, then selected a third-party notice with
the same detection confidence. The project and archive `LICENSE` files declare
MIT, so this metadata error does not change the project's license. Since then the
`0.1.0-alpha.3` SBOMs declare MIT for the main component, as the
[alpha.3 audit](release/v0.1.0-alpha.3-audit.json) records.

[Issue #48](https://github.com/stokaro/unswell/issues/48) records the reproduction.
The original archives, checksum manifest, SBOMs and signed release tag are
retained. Corrected SBOMs are supplemental: they use `.corrected.cdx.json`
filenames and a separate checksum manifest. Each one adds the project's declared
MIT license and removes the misattributed main-component evidence (other metadata
and dependency evidence stay unchanged).

To prepare a correction from a downloaded original, choose a new output path:

```sh
bash scripts/prepare-sbom.sh path/to/original.cdx.json path/to/release.corrected.cdx.json
bash scripts/prepare-sbom.sh --check path/to/release.corrected.cdx.json
```

The six original alpha.1 SBOM module graphs match their binaries' build metadata,
and each archive also contains the root license and 21 dependency notice files or
notice indexes. Corrected metadata does not rebuild or replace those binaries: a
correction touches the metadata alone.
