# Plain-text C examples: extraction repair

Issue [#307](https://github.com/stokaro/unswell/issues/307) identified two long-sentence
warnings inflated by embedded C examples in libevent's plain-text release notes.
The candidate excludes structurally introduced, indented, syntax-validated C
regions. It preserves the prose on either side and the original UTF-8 byte ranges.
No rule, length threshold, profile, score formula, or gate policy changes.

## Frozen sources and review

[freeze.json](freeze.json) binds [labels.json](labels.json) and the permitted
sources and notices in [inputs.tar.gz](inputs.tar.gz). They were committed before
candidate execution. The actual reviewer is the OpenAI assistant, accepted under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md), not an independent
human annotator.

The issue's `whatsnew-2.1.txt` is retained at libevent commit
`5df3037d10556bfcb675bc73e516978b75fc7bc7`, SHA-256
`c101c12f88100aa4af50ab131d0b0932505d7bfe56c4d4da074fedd6830cd427`.
It is exposed development data. The additional `whatsnew-2.0.txt` was already used
by earlier Unswell studies and is also development, not fresh confirmation.

The separately selected complete `whatsnew-2.2.txt` is pinned at libevent commit
`d82464a277d0f42703702c4dfd9af6af38595a83`. Before execution, the review marked
three C regions and five prose controls, plus out-of-scope shell and Autoconf
examples. The confirmation scope is deliberately narrow: one page from the same
project does not establish transfer to other projects or languages.

## Matched replay

The baseline build records `888d513b1f750184d16bc4ccd0fc055373e0cd0c` (merged #356);
the candidate records `fab37cdfb2cf55b82acbdf1b2404e4ac7993d6a5`.
The replay records both binary hashes and every report hash.

Both `technical` and `strict` return 56 findings before and 54 afterward across
the three complete pages. Exactly two `syntax.long-sentence` findings disappear,
both from the original issue. There are no added findings or lost findings outside
those two code-inflated spans. Both profiles retain their previous passing gate.

- Lines 235–241 joined the two `event_del` signatures to surrounding prose.
- Lines 547–562 joined macros, a switch body, and replacement function definitions
  to the prose introducing them.

Review of the complete affected passages supports removing these warnings: the
separate prose statements no longer meet the unchanged length condition. The
conditions about blocking callbacks and the explanations around the optimization
remain eligible. All 12 frozen editorial-defect targets in the original page
remain outside the new exclusions; this repair does not claim to detect them.

| Complete page | Code regions excluded | Prose words before | Prose words after |
| --- | ---: | ---: | ---: |
| Development: 2.0 | 2 | 3,893 | 3,822 |
| Development: 2.1 | 6 | 4,707 | 4,652 |
| Confirmation: 2.2 | 3 | 1,910 | 1,871 |

All 11 new exclusions were reviewed as C examples. On the confirmation page, the
excluded bytes exactly match the three predeclared C regions. All five prose
controls remain eligible, including annotated API calls, compatibility and timing
conditions, DNS fallback rules, and the platform list. No other new bytes are
excluded. Other-language examples remain outside this recognizer.

## Bounds and regression checks

Recognition requires a colon-ended prose introduction, greater indentation,
a C declaration/preprocessor/call anchor, and a complete C syntax parse. Bare
calls require at least two complete signatures; a single call requires a
semicolon. The surrounding text is split at the excluded example. Ellipsis
placeholders are normalized only in a private parser buffer.

A candidate is limited to 128 lines and 16 KiB; a document permits at most 100
parser candidates. Malformed, oversized, mixed prose/code, or unrecognized regions
remain prose. Exceeding the candidate count returns an operational error. A
candidate is consumed once, preventing nested colons from triggering quadratic
reparsing. Indentation alone cannot hide prose.

Blackbox extractor tests cover exact ranges, unchanged input, Unicode, BOM,
CRLF, prose conditions, annotated identifiers, quoted signatures, ordinary
semicolon lists, malformed C, mixed prose/code, and both size limits. CLI golden
fixtures exercise LF and BOM/CRLF, verify text/JSON/SARIF locations and exclusions,
and keep configured warnings in the surrounding and indented prose. The complete `make check` passed on the candidate commit in a bounded remote
Linux container using Go 1.27.0, including all module tests, strict linters,
CLI/MCP dogfood equality, negative probes, and repeated pure-Go builds for six
platform/architecture combinations. The replay above used that same checkout.
No deferred race, fuzz, or coverage run was performed.

## Reproduce

Build the CLI before and after this change with `CGO_ENABLED=0`, recording each
source commit. Run from the repository root:

```sh
python3 -B research/reviews/2026-09-27-plaintext-code/replay.py \
  --before /absolute/path/to/before-unswell \
  --after /absolute/path/to/after-unswell \
  --output /absolute/path/to/replay-results
```

The script verifies the freeze, source hashes, and exact confirmation quotations;
runs both profiles without changing policy; saves complete compressed reports and
binary/report hashes; verifies unchanged old exclusions; and checks all original
editorial targets and fresh controls against the new exclusion ranges. The
[results](results/summary.json) retain the measured reports and removed findings.
