# Discourse roles: development gains, confirmation regression

The new six-page result falls from **2/44 to 1/44** full frozen-event detections
in both profiles. The only new warning is a false positive on the timeout advice
for databases slow to accept connections. The stricter actor check removes a
valid configure-options diagnostic. This candidate has **not** met the 80% recall
or 85% soft-precision objectives and is not ready for acceptance.

On previously exposed pages, 13 new findings identify 11 full frozen events,
one additional wording repair and one false positive. Five removed findings
comprise two known false positives, one unresolved judgment, one full detection
and one partial detection. The result fixes the literal row-counts and checksum
errors, but its losses must remain visible. Development gains do not establish
quality on new prose.

## Runtime changes

The existing evaluative-closure rule, version 8, recognizes information subjects
with selecting importance or attention predicates, author-to-reader notices,
document-outcome endorsements and gerund subjects with abstract functional
clefts. It keeps a following independent clause outside a local judgment while
retaining attached conditions. Bare noun/counts compounds are excluded.

Unscoped-assurance version 5 admits ordinary copular quality and difficulty
predicates, with local condition, measurement and mechanism exclusions.
Instruction-scaffolding version 8 restricts relative support to operation or
actor antecedents. These are bounded surface roles over existing POS tokens,
not dependency edges or semantic entailment. No rule count, threshold, weight,
gate, model or class changed: the catalog remains 53 rules, class manifest r8.

The new negative evidence identifies specific limits:

- The quality matcher accepts a relative restriction inside an instruction as
  if it were the whole sentence's quality claim: the timeout recommendation is
  a concrete symptom/action relationship.
- A clock invocation comparison loses the surrounding coarse-resolution
  tradeoff and precise-timer opt-out. A missing numeric benchmark alone does
  not make that technical comparison an editorial defect.
- The relative actor filter loses a named utility function followed by a
  protected identifier, and a configure-options support description. A supplied
  operand and a named operation require different role checks.
- Most frozen wordiness, reader framing, qualitative claims and discourse-level
  repetition remain undiagnosed. More matches on familiar clauses do not meet
  the complete-page objective.

No runtime adjustment was made after opening these confirmation results. Keep
this candidate in draft until later work addresses those failures with separate
evidence. The archived negative result must not be rewritten as a success.

## Frozen review

Before source reading, the selector froze one short, medium and long page per
cohort, excluding 130 prior references and hashes. The implementing Codex
assistant reviewed all 792 extracted blocks and table cells, then froze 44
defects, seven uncertainties and 28 controls across all seven categories.
This follows [ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md):
maintainer-accepted assistant review, not human labels or independent review.

The source-only overlap audit found no complete extracted block of at least
12 whitespace tokens in earlier sources after whitespace normalization. It
compared all 130 previous sources. Shorter, partial and paraphrased reuse can
escape that audit; no known overlap is not proof of independent provenance.
All 136 selected source identities are now exposed. These six-page counts do
not estimate population recall, and one page per cohort/length cell cannot
support a meaningful within-cell bootstrap interval.

| Set | Pages | Frozen defects | Before full | After full |
| --- | --- | --- | --- | --- |
| development | 26 | 57 | 14 | 15 |
| exposed_repetition | 6 | 5 | 5 | 5 |
| exposed_framing | 9 | 11 | 1 | 3 |
| exposed_local | 12 | 78 | 11 | 11 |
| exposed_context | 11 | 42 | 8 | 9 |
| exposed_instruction | 12 | 99 | 30 | 30 |
| exposed_construction | 6 | 32 | 7 | 8 |
| exposed_purpose | 6 | 28 | 4 | 4 |
| exposed_verb | 6 | 43 | 2 | 3 |
| exposed_relations | 6 | 45 | 8 | 10 |
| exposed_projection | 6 | 34 | 6 | 6 |
| exposed_scope | 6 | 72 | 7 | 7 |
| exposed_proposition | 6 | 46 | 4 | 4 |
| exposed_clauses | 6 | 58 | 7 | 6 |
| exposed_action | 6 | 57 | 6 | 9 |
| confirmation | 6 | 44 | 2 | 1 |

The older repetition and framing studies retain their narrower original review
scope. Their counts do not support precision claims over all warnings. Every
changed finding is reviewed separately in [CHANGES.md](CHANGES.md), including
new findings outside those earlier scopes.

## New complete pages

| Page | Original path | Defects | Detected |
| --- | --- | --- | --- |
| c01 | docs/site/src/content/docs/operate/ai-agent-troubleshooting.md | 0 | 0 |
| c02 | docs/site/src/content/docs/direct/inspect.md | 10 | 1 |
| c03 | docs/site/src/content/docs/faq.md | 9 | 0 |
| c04 | docs/INSTALL.md | 13 | 0 |
| c05 | docs/usage/settings.md | 6 | 0 |
| c06 | docs/8.6.0_docs.md | 6 | 0 |

Ptah remains **1/19**: only the duplicated article is fully detected. Historical
pages fall from **1/25 to 0/25**. The agent diagnostic-code reference has no
required edits and no findings. Passing it is appropriate.

Technical has 27 findings and strict has 29, both before and after. After review,
one finding fully detects a frozen event and another partially detects a
purpose/mandatory wrapper. Five additional repairs were identified only after
diagnostics; they cannot increase frozen recall. Nine findings are unresolved.
The remaining 11 technical and 13 strict findings are nonactionable.

Thus the fraction providing a full frozen-event diagnosis is 1/27 technical and
1/29 strict. Including the partial and additional repairs yields 7/27 and 7/29,
still below 85%. These denominators count findings, not independent documents.
[CONFIRMATION-MISSES.md](CONFIRMATION-MISSES.md) retains all 43 events without
full coverage. All six pages pass both gates; no gate was lowered to force a
failure, and passing does not establish editorial quality.

## Reproduction and checks

Before runtime: `0e0941ae26dcfa89a4b5d92fcb3080cf889fc596`. After runtime: `241c65bb77b23551064bd7e45c2ea57de3434bf3`.
The latter binary was built from the clean semantic-freeze commit before
opening confirmation diagnostics. Exposed before reports are retained from the
preceding iteration; the fresh before run uses its pinned binary. The new runtime
replayed all 16 sets in both profiles with unchanged input and policy identities.
The known #312 repeated-claim budget abstention on exposed instruction c05 is
unchanged. No additional abstention, incomplete scan or operational error is
accepted by the evidence validator.

| Set | Profile | Seconds before → after | Peak MiB before → after |
| --- | --- | --- | --- |
| development | technical | 2.525 → 2.561 | 191.1 → 201.9 |
| development | strict | 1.492 → 1.507 | 188.5 → 183.6 |
| exposed_repetition | technical | 0.546 → 0.533 | 111.6 → 116.2 |
| exposed_repetition | strict | 0.540 → 0.534 | 129.0 → 116.8 |
| exposed_framing | technical | 0.736 → 0.729 | 129.9 → 129.4 |
| exposed_framing | strict | 0.723 → 0.725 | 136.1 → 140.7 |
| exposed_local | technical | 0.633 → 0.607 | 122.1 → 118.9 |
| exposed_local | strict | 0.656 → 0.616 | 127.0 → 120.3 |
| exposed_context | technical | 0.693 → 0.639 | 121.8 → 117.4 |
| exposed_context | strict | 0.671 → 0.643 | 121.1 → 127.0 |
| exposed_instruction | technical | 0.921 → 0.875 | 160.6 → 163.3 |
| exposed_instruction | strict | 0.961 → 0.866 | 148.4 → 165.8 |
| exposed_construction | technical | 0.485 → 0.439 | 102.7 → 98.0 |
| exposed_construction | strict | 0.491 → 0.434 | 105.9 → 103.3 |
| exposed_purpose | technical | 0.503 → 0.441 | 102.7 → 102.6 |
| exposed_purpose | strict | 0.504 → 0.448 | 102.8 → 103.1 |
| exposed_verb | technical | 0.551 → 0.480 | 105.2 → 97.0 |
| exposed_verb | strict | 0.548 → 0.483 | 96.6 → 104.8 |
| exposed_relations | technical | 0.578 → 0.520 | 118.0 → 121.2 |
| exposed_relations | strict | 0.573 → 0.516 | 112.5 → 114.6 |
| exposed_projection | technical | 0.499 → 0.438 | 111.8 → 103.8 |
| exposed_projection | strict | 0.493 → 0.437 | 103.2 → 107.4 |
| exposed_scope | technical | 0.733 → 0.659 | 155.9 → 169.4 |
| exposed_scope | strict | 0.735 → 0.657 | 155.6 → 148.1 |
| exposed_proposition | technical | 0.445 → 0.386 | 79.8 → 90.9 |
| exposed_proposition | strict | 0.468 → 0.383 | 87.2 → 82.0 |
| exposed_clauses | technical | 0.477 → 0.417 | 90.6 → 85.8 |
| exposed_clauses | strict | 0.483 → 0.417 | 90.0 → 86.0 |
| exposed_action | technical | 0.448 → 0.401 | 89.5 → 91.2 |
| exposed_action | strict | 0.444 → 0.387 | 87.7 → 90.1 |
| confirmation | technical | 0.497 → 0.438 | 98.5 → 91.9 |
| confirmation | strict | 0.437 → 0.433 | 95.2 → 89.8 |

These host observations are not a controlled performance comparison or a
2-vCPU qualification. Earlier before measurements overlapped other checks.
Run `python3 tools/render.py` and `python3 tools/test_evidence.py` to verify
and regenerate the report. `tools/measure.py` replays an explicit binary into
a new directory. [VALIDATION.md](VALIDATION.md) records actual product checks.
This experiment has not been merged or deployed to the playground.
