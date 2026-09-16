# September 16 Ptah rhetoric development review

The [evaluation](../../../docs/research/ptah-rhetoric.md) explains the decision
and its limits. This is exposed development material, not a new confirmatory
study or human qualification. No model-generation call was made.

`inputs.json` pins all 136 Ptah files to commit
`654eae5591392278e6c8bce8e54737f780766f19`, with file paths, byte counts, and
SHA-256. `technical.json` and `strict.json` are generated from complete saved
reports by `scripts/contextual-prose-report.py`. `new-findings.json` retains all
nine added warnings and their original locations. `engine.json` identifies the
development binary and changed production files; it is not a release identity.

`historical.json` retains all 27 source bindings, prose counts, report hashes,
and new findings for the reused historical comparison. The selection and rights
are in the [earlier record](../2026-09-15-contextual/README.md) and its
`comparison-manifest.json` and `comparison-source-bindings.json`. This exposed
sample was reused without selecting pages by the new diagnostics.

## Reproduction

Build the native CLI at `fd88e26b7afa8fd8e7543502d87df24914aeaba8` and at the
commit containing this review, with `CGO_ENABLED=0`. Obtain the pinned Ptah source
files and verify each hash against `inputs.json`. Put their documentation-relative
paths under `sources/` in an isolated directory without a project configuration.
Run both binaries, choosing separate report files, in that directory:

```sh
unswell check --profile technical --jobs 2 --include-source \
  --report json:technical.json sources
unswell check --profile strict --jobs 2 --include-source \
  --report json:strict.json sources
```

The original run passed every source filename explicitly. Directory traversal
reproduces that set when `sources/` contains only the 136 manifest entries.
For the historical comparison, reconstruct the 27 exact source bytes and
filenames in the earlier bindings file, then pass those filenames explicitly.
Do not include configuration files as prose inputs.

Generate each comparison from its two saved reports:

```sh
python3 scripts/contextual-prose-report.py before.json after.json
go test ./e2e -run '^TestPtahRhetoric$' -count=1
go test . ./builtin -run '^Test(LocalRhetoric|FrameRulesPropagate)' -count=1
```

Reports must be complete and preserve source hashes and extraction counts. The
nine new findings must match the source-bound casebook; existing findings retain
their rule IDs, fingerprints, and primary ranges. Raw report hashes depend on
build identities. Full raw reports are reproducible but omitted because they
include the complete source text.

The source examples are licensed under the MIT notice retained in
`e2e/rhetoricdata/LICENSE.ptah`. Their editorial judgments and revisions were
written by the implementation agent. No independent annotator agreed with them,
and the absence of warnings on the controls is not a human precision claim.
