# Bounded proposition repetition: three exposed gains, no new-page gain

This candidate detects three more frozen defects on previously exposed pages.
On six new complete pages, full detection stays at **0/46** in both profiles:
**0/14** in Ptah and **0/32** in historical documentation. The working 80% recall
and 85% soft-precision objectives remain unmet. The new result is evidence that
these bounded comparisons do not address the broader missed constructions.

## Runtime changes

`repetition.repeated-claim` version 2 compares an explicit adjacent reformulation:
an actor performs an action only on objects with a property, followed by the
same actor never performing that action on objects lacking that property.
The second sentence must start with `That is,` or `In other words,`. Actor,
action, tense, property and object must match. A separate bounded dependency-
licensing form preserves the owner and direct-and-transitive scope. It does
not equate different actors, modalities, quantities, conditions or protected
identifiers. This is a surface construction, not general semantic entailment.

`repetition.definition-echo` version 3 recognizes identical gerund actions and
an explicitly bounded both-objects shorthand across a copula.
`repetition.explanatory-restart` version 2 can include an abstract because clause
that repeats the same initial adjective instead of supplying a concrete cause.
A concrete cause is retained outside the warning. Both assertions, and all
three circular judgments when present, have original source spans.

There are still 53 rules and class manifest r8. Thresholds, weights and gates
are unchanged. Blackbox API and compiled CLI tests cover matching restrictions,
distinct actors, scope, tense, properties, conditions, quotes, hidden link
destinations, protected code, Markdown emphasis, Unicode, BOM, CRLF, Go comments,
Python strings, configuration, cancellation and the shared work budget.

## Frozen review and measured changes

The selector excluded all 112 exposed source identities and hashes and the
original historical study. Protocol, sources and notices were frozen before
prose review. The implementing assistant read 310 extracted prose blocks and
table cells, then froze 46 defects, seven uncertainties and 29 controls at
`2026-09-17T19:21:22.186975+00:00` before runtime changes or diagnostic output.

The reviewer is the implementing Codex assistant accepted under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md). These are
assistant development judgments, not human labels or independent qualification.
One page per cohort/length cell supports observed counts, not population recall.
All 118 source identities are now exposed and must be excluded from the next
fresh confirmation. No semantic matcher or frozen label changed after output.

| Set | Pages | Frozen defects | Before full | After full |
| --- | --- | --- | --- | --- |
| development | 26 | 57 | 14 | 14 |
| exposed_repetition | 6 | 5 | 3 | 5 |
| exposed_framing | 9 | 11 | 1 | 1 |
| exposed_local | 12 | 78 | 11 | 11 |
| exposed_context | 11 | 42 | 8 | 8 |
| exposed_instruction | 12 | 99 | 29 | 29 |
| exposed_construction | 6 | 32 | 7 | 7 |
| exposed_purpose | 6 | 28 | 4 | 4 |
| exposed_verb | 6 | 43 | 2 | 2 |
| exposed_relations | 6 | 45 | 8 | 8 |
| exposed_projection | 6 | 34 | 6 | 6 |
| exposed_scope | 6 | 72 | 6 | 7 |
| confirmation | 6 | 46 | 0 | 0 |

The three gains are the circular complex/reason/complex explanation, the
permissive-licenses/never-nonpermissive restatement, and typing both flags/still
typing both. The first replaces a partial diagnostic. No full detection is lost.
The Ptah glossary benefit paraphrase in #305 remains unsupported; that issue
must stay open. [CHANGES.md](CHANGES.md) retains all added and removed findings.
Older restricted repetition/framing review scopes remain explicit.

## New complete pages

| Page | Original path | Defects | Detected |
| --- | --- | --- | --- |
| c01 | docs/site/src/content/docs/inference/concepts/security-and-data-boundaries.md | 4 | 0 |
| c02 | docs/site/src/content/docs/atlas/adoption.md | 8 | 0 |
| c03 | docs/site/src/content/docs/extend/public-api.md | 2 | 0 |
| c04 | docs/solver.md | 14 | 0 |
| c05 | docs/getting-started.md | 16 | 0 |
| c06 | web/ui/react-app/README.md | 2 | 0 |

Technical emits 14 findings: one additional actionable wording edit absent from
frozen labels, two unresolved length judgments and 11 nonactionable findings.
Strict adds one nonactionable dash-density finding. The additional edit receives
no frozen recall credit. The instruction-scaffolding finding in the Prometheus
proxy explanation is a false positive against a frozen technical control.
All six pages pass both gates; that does not establish satisfactory detection.

The [miss ledger](CONFIRMATION-MISSES.md) records all 46 missed events: 16 empty
framing, 19 wordiness, eight unjustified intensifier, two vague claim and one
needless repetition events. There are no frozen formulaic-transition or
needless-complexity defects in this sample. Indirect action descriptions and
narrated tutorial steps recur among the misses. Another small equivalence
template alone is unlikely to deliver the required broad recall.

## Resource correction after the first replay

The first runtime commit `b963a42da231240e718b3e3f965ed9f51e546f5b`
charged both full sentences before checking whether a reformulation marker was
present. That added budget abstentions on `exposed_framing` page c06 and
`exposed_scope` page c03. A regression test reproduced the problem before the
fix. The corrected loop charges the bounded marker probe for every candidate
and the full comparison only for marked pairs. It keeps the existing budget
and does not change semantic matching or resource limits.

`prototype-reports` preserves the first replay. The evidence validator compares
all findings, source documents and gates byte-for-byte against the corrected
replay, in every set and both profiles. They are identical; only the two new
abstentions disappear. The existing #312 repeated-claim abstention on exposed
instruction page c05 remains explicit. No other abstention or error is accepted.

## Reproduction and limits

Before runtime: `d5ef492c15c67fcdcf2a5a7883f7f21d12e0340a`. After runtime: `e18f3dcc29bdc46cd5812979238d9e1e0aaa04d3`.
The after binary was built from a clean detached checkout. Exposed before
reports retain the preceding iteration; all after sets were replayed. New-page
before reports were generated with the pinned before binary. JSON records
source bytes, source mappings, engine and policy identities and resource costs.

Technical passes all sets. Strict fails only the preexisting forbidden
worth-noting phrase in the exposed instruction set. The historical CMake
text-mode limitation remains inherited. These changes are not published to the
playground and remote CI is not treated as verified by these local reports.

| Set | Profile | Seconds before → after | Peak MiB before → after |
| --- | --- | --- | --- |
| development | technical | 2.902 → 2.576 | 194.1 → 194.4 |
| development | strict | 2.119 → 2.015 | 196.6 → 200.8 |
| exposed_repetition | technical | 0.976 → 1.183 | 124.4 → 121.7 |
| exposed_repetition | strict | 0.903 → 0.695 | 120.3 → 113.6 |
| exposed_framing | technical | 1.187 → 1.255 | 144.7 → 141.5 |
| exposed_framing | strict | 1.192 → 1.223 | 134.0 → 131.0 |
| exposed_local | technical | 0.708 → 1.006 | 117.7 → 121.4 |
| exposed_local | strict | 0.694 → 1.078 | 120.1 → 121.0 |
| exposed_context | technical | 0.718 → 1.048 | 123.3 → 129.6 |
| exposed_context | strict | 0.733 → 1.066 | 120.4 → 124.9 |
| exposed_instruction | technical | 0.949 → 1.364 | 151.9 → 152.8 |
| exposed_instruction | strict | 0.947 → 1.329 | 167.3 → 163.8 |
| exposed_construction | technical | 0.498 → 0.879 | 107.7 → 103.9 |
| exposed_construction | strict | 0.502 → 0.764 | 101.7 → 97.9 |
| exposed_purpose | technical | 0.498 → 0.811 | 107.2 → 103.1 |
| exposed_purpose | strict | 0.502 → 0.834 | 106.1 → 106.9 |
| exposed_verb | technical | 0.553 → 0.895 | 97.8 → 104.6 |
| exposed_verb | strict | 0.554 → 0.871 | 107.2 → 99.2 |
| exposed_relations | technical | 0.586 → 0.904 | 125.4 → 115.6 |
| exposed_relations | strict | 0.581 → 0.856 | 108.5 → 117.6 |
| exposed_projection | technical | 0.492 → 0.871 | 108.7 → 110.2 |
| exposed_projection | strict | 0.486 → 0.838 | 103.6 → 111.3 |
| exposed_scope | technical | 0.735 → 1.135 | 165.3 → 164.8 |
| exposed_scope | strict | 0.736 → 0.772 | 153.1 → 166.2 |
| confirmation | technical | 0.814 → 0.387 | 87.7 → 84.0 |
| confirmation | strict | 0.736 → 0.386 | 90.5 → 79.7 |

Measurements overlapped ordinary checks; they are not a controlled performance
comparison or 2-vCPU qualification. Run `python3 tools/render.py` and
`python3 tools/test_evidence.py` here to validate and regenerate the artifacts.
`tools/measure.py` replays an explicit binary into a new output directory.
See [VALIDATION.md](VALIDATION.md) for product checks and actual status.
