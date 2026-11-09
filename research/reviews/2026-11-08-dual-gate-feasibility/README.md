# #349 dual-gate feasibility on the latest reviewed engine

This offline calculation uses the frozen [quality-claim screen](../2026-11-07-quality-claim-screen/README.md), whose source result SHA-256 is pinned in `analyze.py`. That screen is the latest complete 36-page, 123-event editorial accounting for the runtime rules on main. No runtime or policy files changed between its main commit `0b5433f` and main commit `b960c67`; later changes to research tools do not alter the measured engine. This calculation is exposed development arithmetic, not a detector experiment, a new editorial review, or prospective confirmation.

The fixed comparison supposes exactly one new **accepted** finding for each missing FULL event, no lost prior FULL witness, and no new nonaccepted finding. It then asks how many existing nonaccepted findings must be withheld to reach 85% accepted delivery. This assumption is deliberately favorable to candidate generation. It does not prove that such new findings exist, that a safe selector can identify nonaccepted findings, or that all suggestions preserve meaning.

| Profile and cohort | FULL now / required | Accepted now / delivered | New accepted findings in scenario | Existing nonaccepted findings to withhold |
| --- | ---: | ---: | ---: | ---: |
| Technical, whole page | 28 / 46 | 82 / 126 | 18 | **27 of 44** |
| Technical, contextual | 16 / 37 | 21 / 27 | 21 | 0 of 6 |
| Technical, historical confirmation | 4 / 16 | 6 / 9 | 12 | 0 of 3 |
| Strict, whole page | 28 / 46 | 82 / 145 | 18 | **46 of 63** |
| Strict, contextual | 16 / 37 | 21 / 28 | 21 | 0 of 7 |
| Strict, historical confirmation | 4 / 16 | 6 / 10 | 12 | 1 of 4 |

For example, technical whole-page delivery would become 100 accepted among 144 findings before filtering, or 69.4%. Withholding 27 nonaccepted findings would yield 100/117, or 85.5%. This is a conditional arithmetic requirement, **not** a measured improvement. Removing any accepted finding, generating rejected candidates, or failing to cover a complete event changes the numbers. A single finding may cover more than one event, and some events need multiple target spans; the scenario does not model either case.

The calculation makes the next end-to-end test concrete: source-bound semantic candidate generation must recover whole-page, contextual, and historical misses while delivery triage removes enough *nonaccepted* whole-page warnings without losing accepted FULL witnesses. The previous length-only [audit](../2026-11-08-length-triage-audit/README.md) cannot supply that triage. A separate page-disjoint pilot and then the unchanged complete development and prospective gates remain necessary. Issue #349 stays open.

## Category frontier of the frozen events

The [category frontier](category-frontier.json) joins the unchanged 123-event
review ledger to the seven later, individually reviewed FULL gains in the
scoped-degree and quality-claim screens. Source hashes and each cohort's
published FULL count are checked. This is an **optimistic upper bound** over
already exposed assistant labels, not a new detector result. It counts a
PARTIAL or UNCERTAIN event as recoverable only in the hypothetical ceiling;
neither receives FULL credit now.

| Hypothetical recovery | Whole-page FULL ceiling | Context FULL ceiling | Historical FULL ceiling |
| --- | ---: | ---: | ---: |
| Current reviewed baseline | 28/57 | 16/46 | 4/20 |
| Every non-FULL empty-framing event | 44/57 | 29/46 | 13/20 |
| Every non-FULL empty-framing and wordiness event | 48/57 | 39/46 | 16/20 |
| Required for 80% | 46/57 | 37/46 | 16/20 |

An empty-framing-only mechanism cannot reach the recall threshold in **any**
complete cohort, even if it is perfect. The two-category ceiling reaches all
three recall thresholds only if it recovers at least 18/20 selected non-FULL
whole-page events, 21/23 context events, and **all 12/12 historical events**.
That is an unusually brittle requirement for a new mechanism. It leaves the
whole-page delivery problem intact: under the separate perfect-new-finding
scenario above, 27/44 existing nonaccepted technical findings, or 46/63 strict
findings, still need safe suppression. The two calculations share frozen
denominators; neither assumes such recovery or suppression is achievable.

This rules out another empty-narration-only selector as the next complete
#349 candidate. The next hypothesis needs coverage across multiple editorial
relations, including wordiness or vague claims, and a separate treatment of
unaccepted delivery. It must be frozen and tested on new page-disjoint material
before the complete-stream and prospective checks. No origin score contributes
to these event counts.

Reproduce the table and identity checks:

```sh
make check-dual-gate-feasibility
```

`result.json` is the machine-readable delivery table, and
`category-frontier.json` is the category ceiling. The check rejects changed
source digests, missing events, duplicate events, changed reviewed totals, and
saved-result drift.
