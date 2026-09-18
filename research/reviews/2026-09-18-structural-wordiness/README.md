# Structural wordiness: useful local repairs, low whole-page recall

The candidate improves full detection from **146/947 to 150/947** on 160 exposed
pages and from **1/63 to 2/63** on twelve new pages. On the six new Ptah pages,
**0/18** defects are fully diagnosed in either phase. These results do not meet
the 80% recall or 85% soft-diagnostic precision goals. A passing gate does not
establish that a page is clean.

The two new rules recognize nominal support around an action or modifier, and
repeated markers of the same grammatical relationship. The instruction rule can
also relate an adjacent nominal instrument to its action announcement. On the
new pages, one warning fully identifies a duplicate additive relationship and
another identifies only the nominal support inside a larger narrated instruction.
Neither rule establishes machine authorship.

An isolated subject-cleft rule was rejected during development: many matches
expressed defensible technical focus. Its source snapshot and every development
judgment are retained in [development-decisions.md](development-decisions.md) and
[development-screening.json](development-screening.json). It is absent from the
frozen confirmation candidate.

## Scope and review

The same Codex assistant selected sources, reviewed them, implemented the
candidate and reviewed diagnostics. This is maintainer-accepted assistant review
under ADR0041. It is not independent human annotation, model qualification or
population accuracy. Frozen labels cover all seven rubric categories and contain
63 defects, 16 uncertain events and 53 controls. Two Ptah pages have no defects
and were retained in the denominator of page-level outcomes.

The prior 160 source identities are exposed development data. The new sample has
four short and two medium Ptah pages, plus two short, two medium and two long
historical pages from six repositories. The pinned Ptah frame has no unused long
pages. Historical expansion and metadata ranking were fixed before source review;
new historical results cannot establish recall on fresh long Ptah pages. Exact
whole-block overlap checks found no prior matches of at least twelve whitespace
tokens. Shorter or paraphrased overlap remains possible. Dates do not prove human
origin.

All twenty sets were replayed in both profiles with the two recorded binaries.
Runs are complete, with no operational errors, skipped rules or abstentions.
Two older targeted sets retain their narrower review scope; every changed warning
still has a disposition. On exposed pages the candidate adds nineteen reviewed
warnings: four full frozen detections, four partial repairs and eleven additional
edits. One shorter instruction finding is replaced by the finding that also
includes its method. No existing fully detected event is lost.

| Set | Pages | Frozen defects | Baseline full | Candidate full |
| --- | --- | --- | --- | --- |
| development | 26 | 57 | 15 | 15 |
| exposed_repetition | 6 | 5 | 5 | 5 |
| exposed_framing | 9 | 11 | 3 | 3 |
| exposed_local | 12 | 78 | 11 | 11 |
| exposed_context | 11 | 42 | 9 | 10 |
| exposed_instruction | 12 | 99 | 30 | 30 |
| exposed_construction | 6 | 32 | 8 | 8 |
| exposed_purpose | 6 | 28 | 4 | 4 |
| exposed_verb | 6 | 43 | 3 | 5 |
| exposed_relations | 6 | 45 | 10 | 11 |
| exposed_projection | 6 | 34 | 6 | 6 |
| exposed_scope | 6 | 72 | 7 | 7 |
| exposed_proposition | 6 | 46 | 4 | 4 |
| exposed_clauses | 6 | 58 | 7 | 7 |
| exposed_action | 6 | 57 | 9 | 9 |
| exposed_roles | 6 | 44 | 3 | 3 |
| exposed_boundaries | 6 | 56 | 7 | 7 |
| exposed_grammar | 6 | 34 | 2 | 2 |
| exposed_discourse | 12 | 106 | 3 | 3 |
| confirmation | 12 | 63 | 1 | 2 |

## Confirmation outcomes

| Page | Stored source | Defects | Fully detected |
| --- | --- | --- | --- |
| c01 | docs/site/src/content/docs/inference/guides/migrate-a-live-table.md | 6 | 0 |
| c02 | docs/site/src/content/docs/inference/reference/evaluation-corpus.md | 6 | 0 |
| c03 | docs/site/src/content/docs/databases/overview.mdx | 0 | 0 |
| c04 | docs/site/src/content/docs/operate/seed-data.md | 2 | 0 |
| c05 | docs/site/src/content/docs/schema/serve.mdx | 4 | 0 |
| c06 | docs/site/src/content/docs/schema/orm-and-external.md | 0 | 0 |
| c07 | doc/xds-test-descriptions.md | 5 | 0 |
| c08 | design/draft-gobuild.md | 11 | 2 |
| c09 | contributors/devel/sig-node/test-suite.md | 12 | 0 |
| c10 | docs/entityRelationshipDiagram.md | 8 | 0 |
| c11 | PLUGINS.md | 1 | 0 |
| c12 | src/ch18-02-refutability.md | 8 | 0 |

Every finding in both phases and profiles was reviewed. Partial repairs do not
increase full-event recall. Additional edits discovered after diagnostics have
no frozen recall credit. Controls and uncertain labels are unchanged.

| Profile | Phase | Findings | Frozen actionable | Additional repairs | Uncertain | Nonactionable |
| --- | --- | --- | --- | --- | --- | --- |
| technical | before | 40 | 5 | 1 | 14 | 20 |
| technical | after | 42 | 7 | 1 | 14 | 20 |
| strict | before | 46 | 5 | 1 | 14 | 26 |
| strict | after | 48 | 7 | 1 | 14 | 26 |

The seven actionable strict findings include only two complete events; the others
partially address four events. Three existing duplicate-word warnings conflate
the Boolean operator name `AND` with the English conjunction `and`. These are
recorded as false positives, not counted as repeated prose. The frozen runtime
is unchanged after confirmation; [#333](https://github.com/stokaro/unswell/issues/333)
tracks this separate correctness repair.

[CHANGES.md](CHANGES.md) lists each diagnostic change. [MISSES.md](MISSES.md) keeps
all 61 remaining confirmation defects and their source-bound repairs.
The larger misses concern multi-part narration, indirect claims and repeated
explanation. More local cue additions alone have not shown sufficient transfer.
Future experiments must compare broader representations with this rules baseline
on separately grouped pages, preserving contextual controls and the origin/quality
distinction.

## Reproduce and validate

Baseline runtime: `f32dcd15f2ab9fac18fade9c840c0bf865e54277`.
Candidate runtime: `d15cc2055753c58d534072ec4e71e81e5de0b899`.
The candidate binary was built from the clean commit before the first
confirmation diagnostic output. `code-freeze.json` records all 491 source-input
hashes and the binary hash. The archive container was subsequently normalized to
remove directory entries; every frozen file digest stayed identical. Research
catalog revision 9 and its integration tests were added later and do not change
runtime source inputs.

Build either revision with `CGO_ENABLED=0`, `-trimpath`, and
`-ldflags '-X github.com/stokaro/unswell.BuildCommit=<full revision>'`, then run:

```sh
python3 tools/measure.py --binary /path/to/unswell --set confirmation --output /new/output
python3 tools/test_evidence.py
python3 tools/render.py
```

The runner restores sources outside any enclosing Git checkout, preventing its
index from silently excluding untracked replay inputs. Output directories must
be new. Fourteen evidence tests reject missing judgments, invented credit, source
reuse, changed runtime files and unresolved analysis failures. Public API tests
cover protected operands, quotations, legal restrictions, separate actions,
Unicode, mappings and paragraph boundaries. The annotated CLI fixture includes
revisions and controls with BOM and CRLF. Existing golden changes are catalog
identity hashes only. Race, active fuzzing and coverage remain deferred to #123.
Repository validation outcomes are retained in [validation.json](validation.json).

[COSTS.md](COSTS.md) reports complete local scan costs. They do not qualify the
separate 100,000-word production performance target.
