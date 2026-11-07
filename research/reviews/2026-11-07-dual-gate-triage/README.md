# Dual-gate triage for #349

This is an arithmetic audit of the [frozen 36-page editorial review](../2026-11-07-current-main-editorial-review/README.md),
not a detector experiment. It uses all 123 unchanged reference events and all
176 strict-profile finding decisions; the technical profile contains 155 of
those findings. The reviewer was a Codex assistant under ADR 0041. The
[analyzer](analyze.py) checks the archived review-input hashes and recomputes
every cohort count before writing the [result](result.json). No new source page,
model, or editorial label is used.

The existing #349 gates are at least 80% FULL-event recall and 85% accepted
delivered findings **in each complete cohort**. The table shows the current
technical profile and two deliberately optimistic acceptance scenarios:
every added finding is accepted, each newly FULL event needs only one added
finding, and no previously FULL event is lost. Actual changes may need more
findings and may also introduce false positives.

| Cohort | FULL now / required | Accepted now / delivered | Perfect additions needed if nothing is removed | Nonaccepted removals needed with exactly one perfect addition per missing FULL event |
| --- | ---: | ---: | ---: | ---: |
| Whole page | 24 / 46 | 78 / 122 | 172 | 27 |
| Contextual | 13 / 37 | 18 / 24 | 16 | 0 |
| Historical confirmation | 4 / 16 | 6 / 9 | 11 | 0 |

The strict profile needs 279 perfect additions with no removals on whole pages,
or 46 nonaccepted removals if exactly 22 perfect findings are added. Its current
whole-page stream is 78 accepted of 141 delivered. These are feasibility
equations, not a recommendation to add hundreds of warnings or hide existing
findings. Improving recall alone cannot plausibly repair the whole-page
acceptance rate; the existing warning stream also needs evidence-based triage.

The non-FULL event categories show why a metadiscourse-only approach cannot
complete #349. Even if every currently non-FULL `empty_framing` event became
FULL, the whole, contextual, and confirmation cohorts would still need at least
6, 11, and 3 new FULL events from other categories. The complete category
counts and event denominators are in `result.json`.

On whole pages, 38 of the 44 nonaccepted technical findings come from
`syntax.long-sentence`; that same rule has 46 accepted findings and is credited
for one existing FULL event. Removing it wholesale would discard real signal.
The strict-only `format.em-dash-density` rule contributes 19 nonaccepted
findings on whole pages and no accepted findings in the three cohorts, but this
exposed sample alone does not justify changing its default. The next experiment
should test source-local **selective** triage of length and punctuation warnings
with close legitimate controls, while separately adding operators for multiple
missed event categories. Any new stream needs complete source-context review,
original-byte evidence, and an untouched prospective confirmation. Existing
labels and thresholds remain fixed.

Reproduce without a model or network access:

```sh
python3 -B research/reviews/2026-11-07-dual-gate-triage/analyze.py --check
python3 -B -m unittest discover -s research/reviews/2026-11-07-dual-gate-triage -p 'test_*.py'
```
