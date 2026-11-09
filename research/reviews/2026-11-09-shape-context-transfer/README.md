# #349 source-disjoint sentence-shape transfer: stopped at design review

**Status: terminal negative design screen; no selector or product change.**
This study asks whether source context can turn the five candidate sentence
shapes in the [earlier screen](../2026-11-09-sentence-shape-screen/README.md)
into specific, safe editorial diagnoses on technical prose. It does not test
AI authorship. The earlier screen found 680 eligible locations in the exposed
#349 archive, including locations touching 28 previously non-FULL events, but
the matches did not themselves establish editorial need.

The [protocol](PROTOCOL.md) was committed as `c3dc1b2` before source
acquisition. The 12 repository README identities were pinned in `sources.json`
as `945f60e` before scoring. The collector and its tests were fixed as
`437e385`; `freeze.json` sealed their hashes in `75d64b5` before the one
offline candidate run. All pages were read from immutable Git commits, with
Git blob SHA-1, UTF-8 byte count, and source SHA-256 recorded. Full source
copies and complete candidate contexts remain in ignored local artifacts.
The first six repositories formed the design partition; the last six were
reserved for transfer. No source or label from the #349 prospective set was
opened, and no hosted model was called.

| Frozen measure | Design | Reserved transfer |
| --- | ---: | ---: |
| Complete pages | 6/6 | 6/6 |
| Eligible shape matches | 11 | 11 |
| Reviewed as warranted | **0** | Sealed; not reviewed |
| Reviewed as retain | 10 | Sealed; not reviewed |
| Reviewed as uncertain | 1 | Sealed; not reviewed |

The design prerequisite required four warranted matches from at least two
repositories and four retain controls. It failed. Ten matches express a
technical distinction, condition, consequence, or compatibility claim; one
reassurance is uncertain. `design-review.json` records a source-specific
reason for each match. The reviewer is a Codex assistant under ADR 0041,
never a human annotator. No edit was supplied, so suggestion safety was not
measured. As registered, the study stopped before defining a contextual
selector or inspecting transfer contexts. The transfer count is an inventory,
not an evaluation result.

This result argues against direct use of these five forms as editorial rules
on the fixed technical pages. It cannot rule out other source-grounded methods,
estimate general prevalence, or add FULL-event credit to #349. The unchanged
current reviewed baseline remains 28/57 whole-page, 16/46 contextual, and
4/20 historical FULL events, below the issue's complete-cohort target. The
next mechanism needs a source-bound hypothesis with editorially warranted
positive cases before tuning an admission rule. Retain the frozen development
denominators and separate prospective confirmation.

`candidate-summary.json` gives the page accounting; `candidate-ids.json`
binds all matches to pinned source hashes and original UTF-8 byte offsets.
`result.json` hashes the public inputs and records the stop decision. The
public verifier checks frozen hashes, source partitions, candidate counts,
review completeness, and the no-qualification claim:

```sh
python3 -B research/reviews/2026-11-09-shape-context-transfer/verify.py
python3 -B -m unittest discover \
  -s research/reviews/2026-11-09-shape-context-transfer -p 'test_*.py'
```

With the original ignored source copies and frozen exporter binary, replay
also verifies the Git blob hashes, source bytes, candidate IDs, and local
candidate packet hashes:

```sh
python3 -B research/reviews/2026-11-09-shape-context-transfer/verify.py \
  --exporter artifacts/research/2026-11-09-shape-context-transfer/bin/reviewtext \
  --sources artifacts/research/2026-11-09-shape-context-transfer/sources
```

The public packet contains source identities and review reasons but no full
source text. Public accounting alone cannot independently replay the source
locations without the pinned copies. This research check remains outside the
Makefile because an older frozen packet binds that file's SHA-256.
