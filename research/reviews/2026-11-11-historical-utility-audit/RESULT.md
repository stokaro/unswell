# Historical Markdown utility audit for #349

**Status: exposed retrospective assistant review, not product qualification.**
This audit asks whether the unchanged CLI supplied useful advice on the seven
complete Markdown sources from the [historical editorial PR screen](../2026-11-11-historical-editorial-pr-screen/RESULT.md), even when it missed that
screen's frozen diagnosis. It preserves the source-only event labels and
their denominators. The root Codex assistant is the actual reviewer under
ADR 0041, not a human annotator. The assistant had already seen the source-only
labels, baseline findings, and some historical patches; these decisions are
not blind confirmation.

All 13 pre-edit public files were fetched again at the frozen Git commits.
Their Git blob IDs and SHA-256 hashes matched `sources.json` exactly. The
seven Markdown CLI reports have `complete` status and no operational errors.
The review covers every one of their 12 delivered findings against its full
source. An **accepted** finding points to a source-specific editorial change
that would help the reader. **Optional** means the suggestion is plausible but
the source reads adequately as written. No replacement edits were supplied,
so suggestion safety was not assessed. The individual decisions and reasons
are in [`review.json`](review.json).

| Exposed measure | Result |
| --- | ---: |
| Complete Markdown pages | 7/7 |
| Delivered findings reviewed | 12/12 |
| Accepted independent diagnoses | **4/12** (33.3%) |
| Optional diagnoses | 8/12 |
| Pages with at least one accepted diagnosis | **2/7** |
| Frozen positive pages with any span overlap | 2/4 |
| Frozen positive pages with a FULL matching diagnosis | **0/4** |
| Operational errors | 0 |
| Supplied edit safety | Not assessed |

Three accepted findings occur in the Remix tutorial: two dense explanations
whose separate technical steps would read better apart, and one accidental
`a a` duplicate. The fourth is Coral's long scraper troubleshooting sentence,
which combines status, symptom, and menu action. Its split recommendation is
independently useful, but it does **not** identify the frozen double-`try`
problem in the same span. Umbraco's overlapping `in order to` finding likewise
does not identify the misleading `again` command sequence. A location-only
score would therefore overstate success on the four frozen events.

The Remix duplicate also shows a limit of the earlier source-only label
inventory: that page was labeled as having no clear six-purpose contextual
diagnosis before the CLI output was reviewed. The new accepted typo finding
does not retroactively change that frozen label. Pages without a frozen event
are **unlabeled for other possible findings**, not clean negative pages.

This selected sample is small, purpose-enriched, and reviewed by the same
assistant after seeing detector output. Its 4/12 accepted share is not a
general precision estimate. It does show some useful advice outside the four
frozen events, while the missing contextual diagnoses remain missing. The
80% FULL-event recall, 85% accepted-delivery, complete-cohort, safe-edit, and
separate prospective requirements for #349 remain unchanged. A future
utility scorecard should predeclare document-level accepted advice and review
burden on page-disjoint complete sources, without substituting those numbers
for #349 qualification.

The scorer pins the prior ledger, labels, and source manifest by SHA-256,
requires a decision for every finding, checks source identities and exact
span overlap, and rejects unbound FULL-event credit. Its output is
[`result.json`](result.json). Reproduce the public accounting and negative
checks from the repository root:

```sh
python3 -B research/reviews/2026-11-11-historical-utility-audit/score.py --check
python3 -B -m unittest discover \
  -s research/reviews/2026-11-11-historical-utility-audit -p 'test_*.py'
```

To repeat the full-source identity check, fetch the public sources with the
existing historical screen's `fetch_sources.py` and compare its generated
manifest to `sources.json`. The reviewed full source bytes remain outside Git;
the frozen commit, path, blob ID, and SHA-256 for each file are in that
manifest. No model call, runtime rule, original #349 event edit, or
prospective-page use occurred in this audit.
