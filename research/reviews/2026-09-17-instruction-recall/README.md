# Instruction wording: limited transfer, low overall recall

This iteration continues #309 and #299 with two experimental rules:
`filler.instruction-scaffolding` and `repetition.redundant-predicate`.
On twelve new complete pages, full event detections increase from 4/99 to
12/99 in both profiles. All eight gains occur on one Mermaid page. Ptah
stays at 2/35. The broad-recall objective remains unmet.

The implementing Codex assistant reviewed the source prose under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md).
These are maintainer-accepted editorial judgments, not human annotations,
independent agreement, authorship labels or population accuracy. Another
reviewer is not a prerequisite for continuing implementation.

## What changed

The first rule identifies bounded indirect capability instructions, reader-goal
prefaces followed by instructions, and adjacent method announcements. It keeps
source ranges for both clauses. The second identifies a path/location subject
with a redundant location predicate, or a reason subject with a redundant
because predicate. The suggested edits retain capability and optionality.
They do not convert a possibility into an obligation.

Negation, permission, failure, conditional applicability, quoted claims and
protected construction tokens prevent matches. Protected operands remain
opaque. The implementation uses existing NLP, clause windows, source maps,
budgets, cancellation and configuration. Both rules are experimental warnings
with `gate: none`; profile thresholds and existing rule weights are unchanged.
The catalog has 53 rules. Class manifest r8 retains r7 and adds two general-style
rules; neither has LLM-specific empirical qualification.

## Frozen reviews and separate denominators

Input identities were frozen at 14:25 UTC on September 17, 2026. Full-page
annotations were frozen at 14:38 UTC, before runtime edits or inspection of
confirmation diagnostics. The review recorded 99 defects, 17 uncertain events
and 43 acceptable controls across all seven categories.

The historical selection frame expands to the pinned long-prose selection
metadata; it excludes all 38 previously selected historical sources. The
Ptah frame remains at its existing snapshot. The deterministic selection
excludes all 64 previously reviewed pages and exact source hashes. Historical
selection admits one page per repository. The manifest preserves candidates,
selection order, source identities and licenses. This does not establish
repository-independent transfer: the new Mermaid page shares a documentation
family with the exposed sequence-diagram page.

| Set | Pages | Frozen defects | Before | After |
| --- | --- | --- | --- | --- |
| Original development | 26 | 57 | 13 (22.8%) | 13 (22.8%) |
| Exposed repetition targets | 6 | 5 | 3 (60.0%) | 3 (60.0%) |
| Exposed framing targets | 9 | 11 | 1 (9.1%) | 1 (9.1%) |
| Exposed local pages | 12 | 78 | 7 (9.0%) | 7 (9.0%) |
| Exposed context pages | 11 | 42 | 1 (2.4%) | 7 (16.7%) |
| New complete-page confirmation | 12 | 99 | 4 (4.0%) | 12 (12.1%) |

Keep these denominators separate. The repetition and framing reviews cover
specific targets; the other sets have complete-page annotations. Every new
and removed diagnostic is reviewed. Retained new-confirmation diagnostics
are reviewed too. Earlier full reviews are inherited by source identity.

On exposed context pages, six more events receive full credit and three
receive partial credit. One additional warning concerns an unlabelled note
instruction and remains uncertain. On new confirmation, the old wordy-phrase
warning covers only part of c09-d09; the new construction covers it fully.
Partial matches never increase full-event recall. [CHANGES.md](CHANGES.md)
records all additions and their source ranges. Nothing is removed.

## Review burden and applicability

| Set | Profile | Findings before → after | Actionable after | Uncertain | Nonactionable |
| --- | --- | --- | --- | --- | --- |
| Exposed context pages | technical | 28 → 38 | 10 | 7 | 21 |
| Exposed context pages | strict | 37 → 47 | 10 | 7 | 30 |
| New complete-page confirmation | technical | 129 → 137 | 13 | 18 | 106 |
| New complete-page confirmation | strict | 144 → 152 | 13 | 18 | 121 |

Actionable counts include partial findings; event counts deduplicate findings
for the same defect. Unlabelled plausible edits remain uncertain and receive
no primary recall credit. Nonactionable judgments distinguish necessary
technical conditions, reference consistency and incidental length/contrast
warnings from the actual frozen defect. This table is a review-burden measure
for these pages, not a universal false-positive rate.

The new rule adds eight actionable findings on one confirmation page and none
on the other eleven. That concentrated 8/8 result cannot qualify its default
precision or support a broad transfer claim. The redundant-predicate rule has
an exposed positive but no new-confirmation positive.

One existing rule, `repetition.repeated-claim`, exhausts its candidate budget
on c05, the 11,882-word Ptah migration reference. The same abstention occurs
in both engines and both profiles. It remains in every report and in the
summary; the page and its missed defects remain in the recall denominator.
There are no other abstentions or operational errors. Reports are complete
under the current optional-rule budget contract; completeness does not mean
every rule evaluated every page. A future budget repair needs separate evidence.

Technical passes on every set. Strict already fails on confirmation in the
baseline because curl's worth-noting phrase has `gate: forbid`; that result
is unchanged. Other sets pass. The new warnings do not create a gate failure.

| Confirmation page | Cohort | Defects | Detected after |
| --- | --- | --- | --- |
| c01 `docs/site/src/content/docs/inference/quick-start.md` | ptah | 0 | 0 |
| c02 `docs/site/src/content/docs/concepts/dialects-and-capabilities.md` | ptah | 2 | 0 |
| c03 `docs/site/src/content/docs/inference/reference/specification.md` | ptah | 6 | 0 |
| c04 `docs/site/src/content/docs/databases/oracle.md` | ptah | 2 | 0 |
| c05 `docs/site/src/content/docs/atlas/migrate-commands.md` | ptah | 17 | 1 |
| c06 `docs/site/src/content/docs/atlas/docs-coverage.md` | ptah | 8 | 1 |
| c07 `CONTRIBUTING.md` | historical | 12 | 0 |
| c08 `docs/TheArtOfHttpScripting.md` | historical | 20 | 1 |
| c09 `docs/diagrams-and-syntax-and-examples/flowchart.md` | historical | 21 | 9 |
| c10 `docs/querying/operators.md` | historical | 3 | 0 |
| c11 `docs/demo/quickref.md` | historical | 4 | 0 |
| c12 `SCOPE.md` | historical | 4 | 0 |

| Grouping | Group | Defects | Before | After |
| --- | --- | --- | --- | --- |
| cohort | historical | 64 | 2 | 10 |
| cohort | ptah | 35 | 2 | 2 |
| length_stratum | long | 57 | 3 | 3 |
| length_stratum | medium | 32 | 1 | 9 |
| length_stratum | short | 10 | 0 | 0 |

| Category | Confirmation defects | Detected after |
| --- | --- | --- |
| empty_framing | 18 | 1 |
| formulaic_transitions | 1 | 0 |
| needless_complexity | 6 | 1 |
| needless_repetition | 12 | 2 |
| unjustified_intensifiers | 10 | 0 |
| vague_claims | 6 | 0 |
| wordiness | 46 | 8 |

[CONFIRMATION-MISSES.md](CONFIRMATION-MISSES.md) retains all 87 missed events,
source quotes, proposed edits and context references. Broad wordiness,
indirect method statements, repeated explanations and vague claims remain
open work. The 80% recall / 85% soft-diagnostic precision working target is
not met. Lowering a gate would not identify those constructions.

Paired page bootstrap uses 2,000 draws within cohort/length cells and seed
917304. After-recall sensitivity is 2.7%–19.6%;
paired-gain sensitivity is 0.0%–15.2%.
These describe this small selected sample, not population confidence intervals.
Shared repositories and templates, genre selection and one assistant reviewer
limit the interpretation. No inference about who wrote a text is made.

## Evaluation corrections

The inherited semantic-credit table lacked `filler.announced-importance`.
This iteration explicitly maps it to `empty_framing` for both engines: the
exact frozen c08-d14 phrase and its diagnostic match. No source annotation or
runtime changes. A regression test rejects unrelated repetition credit.

The inherited replay validator also assumed zero abstentions. The new validator
retains all source, hash, policy, exit and completeness checks and requires
the exact observed c05 budget abstention in both phases. Negative probes reject
its deletion, duplication or appearance in an older set. This records a product
limitation instead of treating missing analysis as a clean result.

## Replay and resource records

Baseline runtime: `c1032187fcc1708617ac9aff53ce12c84c42f70c`.
Measured implementation: `296ebe84ee30c94ec0df46c04a1a8773e7db68cc`.
Earlier before reports are byte-identical copies of the preceding after reports.
The confirmation baseline and every after report were measured in this iteration.
Subsequent fixture, evidence and documentation changes do not change runtime.

| Set | Profile | Wall seconds before → after | Peak MiB before → after |
| --- | --- | --- | --- |
| Original development | technical | 2.133 → 2.194 | 181.9 → 183.3 |
| Original development | strict | 1.560 → 1.512 | 193.4 → 187.7 |
| Exposed repetition targets | technical | 0.527 → 0.540 | 113.6 → 112.7 |
| Exposed repetition targets | strict | 0.534 → 0.539 | 124.4 → 118.3 |
| Exposed framing targets | technical | 0.727 → 0.734 | 137.6 → 137.6 |
| Exposed framing targets | strict | 0.724 → 0.741 | 140.8 → 142.9 |
| Exposed local pages | technical | 0.667 → 0.618 | 120.0 → 121.6 |
| Exposed local pages | strict | 0.623 → 0.615 | 116.8 → 119.6 |
| Exposed context pages | technical | 0.690 → 0.654 | 119.7 → 117.9 |
| Exposed context pages | strict | 0.659 → 0.739 | 123.3 → 126.8 |
| New complete-page confirmation | technical | 0.911 → 0.886 | 162.5 → 145.1 |
| New complete-page confirmation | strict | 0.845 → 0.873 | 162.3 → 160.3 |

These are single macOS arm64 runs with `CGO_ENABLED=0`, not a speedup claim or
the separate two-vCPU Linux performance target. Records include CPU time, host,
binary hashes, configurations and commands. Authorized research reports include
source text; ordinary saved reports still omit it by default.

From the repository root:

```sh
python3 research/reviews/2026-09-17-instruction-recall/tools/render.py
python3 research/reviews/2026-09-17-instruction-recall/tools/test_evidence.py
```

To replay a scan, build the named revision and run `tools/measure.py` with
`--binary PATH --set SET --output NEW_DIR`. It requires a new directory and
uses the retained input archive without downloads. The renderer validates
freeze hashes, source quotes, all diagnostic dispositions and event credits,
then regenerates this page, the change ledger, misses and summary.

The protocol and both freeze manifests remain unchanged. These twelve pages
are now exposed development material; further tuning requires a separate
confirmation sample. #309 remains open for unsupported indirect method and
narrative-tail constructions. #299 remains open for broad useful recall.
