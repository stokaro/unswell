# September 15 contextual-prose development evaluation

This record supports the [coverage repair](../../../docs/research/contextual-prose.md).
It is development evidence for #218 and #154. Both issues retain their outstanding
experimental requirements. No quality label, attribution probability, or editorial
false-positive rate is inferred from a rule count.

## Inputs and selection

- Ptah: `7b47e7cfb5d4ff32a38375345069f5533bff892f`, all Markdown/MDX in
  `docs/site/src/content/docs`, 133 pages. The repository is already exposed.
- Baseline engine: `f3d09fd3aeaeb355d6519b9cacb98598f149ef6d`.
- Corrected engine: the production-file hashes and binary identity in `engine.json`.
- Historical development sample: `comparison-manifest.json`, selected before the
  first corrected-rule scan. Its original hash is
  `ea01bb0101e588c64528216dd751f799548520048219ecbb91fa6f6ecefe548b`.
  The manifest retains rights, references, source hashes, and global groups.
- `comparison-source-bindings.json` maps the manifest to the isolated scan filenames.
  The sources belong to 17 groups in global dataset plan
  `7bee3144ebb406ce677ffb330f2d906a13bed5a8fde48d45f10bde4468948cd7`.

The historical selection used a fixed hash order, admitted README/documentation
roles and 400–6,000 prose words, and capped each global group at two pages and the
total at 40. It yielded 27 pages. An initial draw mistakenly used shard-local
groups. That unmeasured draw was superseded before comparison; no outcome informed
the corrected selection. Source files were verified against their manifest hashes.
Configuration files were excluded by passing the 27 source filenames explicitly.

No generator ran for this evaluation. All Ptah pages and selected historical pages
were processed; no failed page was silently removed. A scan with an outside-root
configuration path failed before analysis and was retried with a local policy copy.
One exploratory historical invocation also scanned the policy itself; its result
was discarded and replaced with the declared 27-file invocation.

## Results and interpretation

`technical.json`, `strict.json`, `all.json`, and `historical.json` summarize saved
CLI reports. Each retains report hashes, per-source identities, prose counts, and
per-rule counts before and after. `scripts/contextual-prose-report.py` derives
them without implementing or rerunning any rule. It rejects incomplete scans,
changed source bindings, and changed extraction counts.

The Ptah technical profile gains 150 contrast notes; existing counts stay fixed.
The all-rule profile additionally gains 341 passive-candidate findings. Historical
review load grows by one contrast note and 33 passive candidates across 31,488
words. Rates on these two differently selected corpora cannot establish causal
effects of model generation. In particular, zero historical contrast findings in
the baseline are not evidence of zero editorial false positives.

Disposition: fix the context omission, include comma-not alternatives in the
declared matcher, retain one allowed contrast, expose repeated contrasts as
zero-score notes, and keep passive candidates opt-in. Wider linguistic coverage
and generator association require further experiments. The rule neither detects
arbitrary semantic contrasts nor decides that a technical distinction is needless.

The three source passages and their agent-reviewed alternatives are committed in
`e2e/testdata/contextual_ptah`. Their provenance, rights, expected findings, and
preserved technical obligations are inspectable. Synthetic boundary cases in
`e2e/testdata/contextual_inline` cover the source transformations separately.

## Reproduction

Build the CLI at the baseline commit and at this change with `CGO_ENABLED=0`.
Check out Ptah at the pinned commit in an isolated directory. From that directory,
run each binary with these arguments, using separate report filenames:

```sh
unswell check --profile technical --include-source --no-gate \
  --report json:technical.json docs/site/src/content/docs
unswell check --profile strict --include-source --no-gate \
  --report json:strict.json docs/site/src/content/docs
```

For the all-rule scans, copy `research/acquisition/policy-e1.yaml` into the isolated
root and replace `--profile` with `--config` naming that copy. For historical scans,
obtain the exact source bytes named in `comparison-manifest.json`, verify SHA-256,
and copy them to the filenames in `comparison-source-bindings.json`. Pass those
filenames explicitly so the policy is not another input document. Use the same
all-rule policy with both binaries.

From the Unswell checkout, summarize each saved report pair:

```sh
python3 scripts/contextual-prose-report.py /path/to/before.json /path/to/after.json
go test ./e2e -run '^TestCLI/contextual_' -count=1
```

Report hashes vary with recorded build identities; source hashes and diagnostic
counts are the behavioral comparison. The normal fixture suite needs no network,
generator, or historical checkout. Full raw scans remain reproducible from the
pinned sources; they are not embedded because they include complete source text.
