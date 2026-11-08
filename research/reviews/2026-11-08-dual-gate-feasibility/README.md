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

Reproduce the table and identity checks:

```sh
make check-dual-gate-feasibility
```

`result.json` is the machine-readable table. The check rejects a changed source digest or saved result.
