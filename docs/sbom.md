# Release SBOMs

Each CLI archive has a CycloneDX document for its operating system and architecture.
It records the selected Go modules and compiler. The release pipeline declares the
main component's MIT license from the project's root `LICENSE`. Dependency license
evidence comes from the pinned generator and is retained separately.

`make check` verifies this metadata step, including rejection of a missing or wrong
project license. The preparation script requires a new output path and preserves
the dependency graph and its license evidence.

## Alpha.1 metadata erratum

The original `0.1.0-alpha.1` SBOMs attach `Apache-2.0` as the main component's
detected license evidence. The generator scanned the root `LICENSE` and bundled
dependency notices in `licenses/`, then selected a third-party notice with the
same detection confidence. The project and archive `LICENSE` files declare MIT.
This metadata error does not change the project's license.

[Issue #48](https://github.com/stokaro/unswell/issues/48) records the reproduction.
The original archives, checksum manifest, SBOMs and signed release tag are retained.
Any corrected supplemental SBOMs use `.corrected.cdx.json` filenames and a separate
checksum manifest. They add the project's declared MIT license and remove the
misattributed main-component evidence. Other metadata and dependency evidence stay
unchanged.

To prepare a correction from a downloaded original, choose a new output path:

```sh
bash scripts/prepare-sbom.sh path/to/original.cdx.json path/to/release.corrected.cdx.json
bash scripts/prepare-sbom.sh --check path/to/release.corrected.cdx.json
```

The six original alpha.1 SBOM module graphs match their binaries' build metadata.
Each archive also contains the root license and 21 dependency notice files or notice
indexes. Corrected metadata does not rebuild or replace those binaries.
