# Discourse constructions: candidate rejected

The candidate adds seven full detections on 841 previously exposed defect records,
but none on the twelve new pages: **2/106 before and 2/106 after** (1.9%).
Both full detections are in the historical Dear ImGui FAQ: a wordy causal phrase
and an accidentally repeated preposition. Ptah remains **0/31** on these six pages.
The only new confirmation warning concerns `Unfortunately` in kafka-go's
motivations section, which was already labeled uncertain before diagnostics.

The candidate is **not promoted into the runtime**. The final change restores
production code, tests and the generated rule catalog to merged main. It retains
this negative experiment and its missed-defect inventory. The goal of reliable
whole-page detection remains open; neither a low index nor a passing gate proves
that these pages are clean.

## Evidence and boundaries

The same Codex assistant selected the sample, reviewed sources, implemented the
candidate and reviewed diagnostics. This is maintainer-accepted assistant
agreement under ADR0041, not independent human annotation or population accuracy.
All seven rubric categories were reviewed before runtime changes or diagnostic
output. The frozen labels contain 106 defects, 26 uncertain events and 62 controls.

The final sample has two short, three medium and one long page per cohort.
Only one unused eligible long page remained per cohort; the metadata-only
selection amendment preceded source reading and is retained in
[source-selection-notes.md](source-selection-notes.md). All 148 earlier source
identities were excluded. The exact-block overlap check found no matches of at
least twelve whitespace tokens; it does not rule out shorter or paraphrased reuse.
Historical dates are source provenance, not proof of human authorship.

All 19 sets were replayed with fresh CLI processes for both profiles and both
revisions. Runs are complete, with no operational errors, skipped rules or
abstentions. The two oldest targeted sets have narrower review scope; they do not
establish precision for all their findings. These are observed counts without a
population interval; there is only one long page per cohort.

| Set | Pages | Frozen defects | Baseline full | Candidate full |
| --- | --- | --- | --- | --- |
| development | 26 | 57 | 15 | 15 |
| exposed_repetition | 6 | 5 | 5 | 5 |
| exposed_framing | 9 | 11 | 3 | 3 |
| exposed_local | 12 | 78 | 11 | 11 |
| exposed_context | 11 | 42 | 9 | 9 |
| exposed_instruction | 12 | 99 | 30 | 30 |
| exposed_construction | 6 | 32 | 8 | 8 |
| exposed_purpose | 6 | 28 | 4 | 4 |
| exposed_verb | 6 | 43 | 3 | 3 |
| exposed_relations | 6 | 45 | 10 | 10 |
| exposed_projection | 6 | 34 | 6 | 6 |
| exposed_scope | 6 | 72 | 7 | 9 |
| exposed_proposition | 6 | 46 | 4 | 5 |
| exposed_clauses | 6 | 58 | 7 | 7 |
| exposed_action | 6 | 57 | 9 | 9 |
| exposed_roles | 6 | 44 | 3 | 3 |
| exposed_boundaries | 6 | 56 | 7 | 9 |
| exposed_grammar | 6 | 34 | 2 | 4 |
| confirmation | 12 | 106 | 2 | 2 |

## New-page review

| Page | Stored source | Defects | Fully detected |
| --- | --- | --- | --- |
| c01 | docs/site/src/content/docs/start/quick-start-postgresql.mdx | 4 | 0 |
| c02 | docs/site/src/content/docs/reference/overview.mdx | 1 | 0 |
| c03 | docs/site/src/content/docs/databases/sqlserver.md | 9 | 0 |
| c04 | docs/site/src/content/docs/reference/hcl-schema.md | 3 | 0 |
| c05 | docs/site/src/content/docs/reference/yaml-schema.md | 3 | 0 |
| c06 | docs/site/src/content/docs/reference/lint-rules.md | 11 | 0 |
| c07 | docs/FAQ.md | 29 | 2 |
| c08 | documentation/internal_architecture.md | 12 | 0 |
| c09 | README.md | 6 | 0 |
| c10 | BUILDING.md | 11 | 0 |
| c11 | docs/NEW-PROTOCOL.md | 14 | 0 |
| c12 | docs/version_1_release_notes.md | 3 | 0 |

Every new-page finding in both phases and profiles has a retained disposition.
Three findings address frozen defects: two fully and one partially. Four other
small wording repairs were found after diagnostics; they cannot increase recall.
Sentence length, punctuation density and overlapping contrast markers receive no
credit for unrelated frozen defects. Controls and uncertainty were not relabeled
to make the candidate look better.

| Profile | Phase | Findings | Frozen actionable | Additional repairs | Uncertain | Nonactionable |
| --- | --- | --- | --- | --- | --- | --- |
| technical | before | 62 | 3 | 4 | 9 | 46 |
| technical | after | 63 | 3 | 4 | 10 | 46 |
| strict | before | 70 | 3 | 4 | 9 | 54 |
| strict | after | 71 | 3 | 4 | 10 | 54 |

The candidate introduces ten warnings on exposed sets: seven address frozen
misses and three identify additional tone edits. No old findings are removed.
On confirmation it adds one uncertain warning. The earlier development gains
do not justify a claim of improved transfer or the 80% recall / 85% soft-precision
targets. [CHANGES.md](CHANGES.md) records every changed warning;
[MISSES.md](MISSES.md) retains all 104 remaining new-page defects.

## Why the candidate was rejected

The candidate recognizes narrated tutorial actions, comma-delimited reader
inference and evaluative cues, and more copular quality modifiers. It still
requires narrow local surface structures. The new defects mostly concern longer
wording dependencies, unsupported scope, repeated explanation and indirect
claims. A finite verb mislabeled as a noun can suppress a valid candidate, while
a generic regret cue does not distinguish an unnecessary technical endorsement
from a relevant author reaction in a motivations section.

The miss inventory also contains a concrete separate boundary bug: the FAQ's
`when when` is missed in a sentence containing the quoted name `ImGui`, although
the duplicate is outside the quotation. This needs local quote scope, with
positive and negative source-mapping tests in [#329](https://github.com/stokaro/unswell/issues/329).
The post-diagnostic [minimal probe](quote-boundary-probe.json) records the three
cases and their complete CLI outcomes; it does not add frozen recall credit.

The next experiment should test substantially broader relations or resolve
measurable extraction/role failures. Extending another short cue list without
showing transfer would repeat this result. The new confirmation pages are now
exposed and may be used for development, not reused as fresh confirmation.

## Reproduction and validation

Baseline runtime: `49ed798d6ce4b00e387de93c8a6c4b490d2655e2`.
Candidate runtime: `068bbeba1f7e57b52f4ec6f50c1ca9eb7863d8b8`.
The candidate binary was built from the clean candidate commit before the
confirmation reports were opened. `code-freeze.json` and `runtime-inputs.tar.gz`
retain its runtime and focused tests. The final checkout intentionally contains
the baseline runtime; rebuilding HEAD does not reproduce the candidate.

Build the specified revision with `CGO_ENABLED=0`, `-trimpath`, and
`-ldflags '-X github.com/stokaro/unswell.BuildCommit=<full revision>'`.
Then run from this study directory:

```sh
python3 tools/measure.py --binary /path/to/unswell --set confirmation --output /new/output
python3 tools/test_evidence.py
python3 tools/render.py
```

The candidate's focused public API tests and strict lint passed. Its broader
root test run found stale rule-version expectations (`10` versus candidate `11`)
in existing tests; those failures are retained in `validation.json`. No candidate
release is claimed. Final repository checks run against the restored runtime and
research artifacts. Race, active fuzzing and coverage remain deferred to #123.
The fifteen evidence tests reject source drift, missing judgments, invented
credit, unexpected abstention and changes to the frozen runtime archive.

The following costs cover complete local CLI scans, including extraction,
analysis and reporting. They are not a qualification of the separate
100,000-word production target. Each run records host and binary identities.

| Set | Profile | Seconds baseline → candidate | Peak MiB baseline → candidate |
| --- | --- | --- | --- |
| development | technical | 2.74 → 2.16 | 192.4 → 184.3 |
| development | strict | 1.52 → 1.52 | 176.4 → 196.7 |
| exposed_repetition | technical | 0.54 → 0.53 | 124.7 → 112.4 |
| exposed_repetition | strict | 0.54 → 0.54 | 118.9 → 114.4 |
| exposed_framing | technical | 0.73 → 0.73 | 141.2 → 143.2 |
| exposed_framing | strict | 0.73 → 0.73 | 134.3 → 132.3 |
| exposed_local | technical | 0.62 → 0.62 | 128.0 → 124.2 |
| exposed_local | strict | 0.62 → 0.62 | 118.6 → 126.8 |
| exposed_context | technical | 0.65 → 0.64 | 122.8 → 117.7 |
| exposed_context | strict | 0.64 → 0.65 | 120.7 → 117.6 |
| exposed_instruction | technical | 0.87 → 0.87 | 146.9 → 142.5 |
| exposed_instruction | strict | 0.86 → 0.87 | 147.0 → 159.3 |
| exposed_construction | technical | 0.44 → 0.45 | 106.2 → 94.5 |
| exposed_construction | strict | 0.43 → 0.44 | 100.6 → 102.7 |
| exposed_purpose | technical | 0.45 → 0.44 | 105.3 → 105.1 |
| exposed_purpose | strict | 0.44 → 0.45 | 104.1 → 106.9 |
| exposed_verb | technical | 0.48 → 0.48 | 103.1 → 104.2 |
| exposed_verb | strict | 0.49 → 0.48 | 99.8 → 105.7 |
| exposed_relations | technical | 0.51 → 0.52 | 119.7 → 115.4 |
| exposed_relations | strict | 0.52 → 0.52 | 113.0 → 118.0 |
| exposed_projection | technical | 0.44 → 0.44 | 102.9 → 104.0 |
| exposed_projection | strict | 0.44 → 0.43 | 104.7 → 101.9 |
| exposed_scope | technical | 0.66 → 0.68 | 164.8 → 164.8 |
| exposed_scope | strict | 0.66 → 0.66 | 148.3 → 163.8 |
| exposed_proposition | technical | 0.39 → 0.40 | 91.2 → 82.9 |
| exposed_proposition | strict | 0.38 → 0.40 | 79.2 → 82.7 |
| exposed_clauses | technical | 0.41 → 0.41 | 94.1 → 86.7 |
| exposed_clauses | strict | 0.42 → 0.43 | 90.0 → 85.8 |
| exposed_action | technical | 0.39 → 0.39 | 93.2 → 88.2 |
| exposed_action | strict | 0.40 → 0.51 | 90.5 → 95.9 |
| exposed_roles | technical | 0.44 → 0.44 | 97.3 → 103.2 |
| exposed_roles | strict | 0.44 → 0.44 | 103.5 → 96.4 |
| exposed_boundaries | technical | 0.46 → 0.46 | 110.3 → 108.1 |
| exposed_boundaries | strict | 0.46 → 0.46 | 104.5 → 106.2 |
| exposed_grammar | technical | 0.43 → 0.42 | 100.5 → 107.4 |
| exposed_grammar | strict | 0.42 → 0.42 | 106.5 → 101.7 |
| confirmation | technical | 0.64 → 0.65 | 130.1 → 119.1 |
| confirmation | strict | 0.64 → 0.65 | 113.2 → 114.0 |
