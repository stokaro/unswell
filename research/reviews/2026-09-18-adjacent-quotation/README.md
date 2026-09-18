# Adjacent-word quotation scope

This follow-up fixes [#329](https://github.com/stokaro/unswell/issues/329).
The previous rule skipped an entire sentence containing a quotation mark.
A quoted library name therefore hid an accidental duplicate elsewhere in
that sentence. Version 2 protects the quoted tokens and retains quotation
state across sentences within one prose block.

The observed example is the `when when` span in the frozen ImGui FAQ
sample (`c07-d03`) from the
[discourse study](../2026-09-18-discourse-patterns/README.md).
Removing one `when` repairs the duplicate without changing the assertion.
The implementing Codex assistant reviewed this local repair. All pages
below were already exposed; this is regression evidence, not a new
confirmation study or an estimate of population recall.

## Regression result

[regression.json](regression.json) records 19 sets, 160 complete page scans
per profile, and both `technical` and `strict` results. The comparison uses
saved reports from `49ed798`, whose runtime is unchanged at the parent
`06dff08`.

- One additional duplicate is detected in each profile: `c07-d03`.
- No existing finding is removed; no other rule gains a finding.
- All 38 runs complete without errors, abstentions, or skipped rules.
- On the latest 12-page set, full matches increase from 2 to 3 of 106
  frozen defects. Ptah remains at 0 of 31 on those six pages.

This addresses a quotation boundary bug. It does not change the rejected
[discourse candidate decision](../2026-09-18-discourse-patterns/decision.json)
or establish good recall for contextual wording defects.

## Reproduction

Build the current CLI and use a new output directory:

```sh
CGO_ENABLED=0 go build -o /tmp/unswell-329 ./cmd/unswell
python3 research/reviews/2026-09-18-adjacent-quotation/tools/replay.py \
  --binary /tmp/unswell-329 --output /tmp/unswell-329-regression
```

The runner verifies the existing source/annotation freeze, runs both
profiles, rejects incomplete analysis, and compares exact source spans.
It writes complete reports and a compact delta. Report hashes vary with
build identity; compare the source-bound findings and counts when
reproducing with a different build.

Public API tests cover unrelated quoted names, quoted duplicates, straight
and curly quotes, multi-sentence quotations, contractions, possessives,
protected code, unmatched openers, block boundaries, term exemptions,
suppressions, budget exhaustion, and source spans with Unicode and CRLF.
CLI tests check the draft, its repair, and a quoted control.

Unclosed quotations protect the rest of their block. Ambiguous plural
possessives inside single quotes retain protection. Other rhetorical
rules keep their existing quotation behavior.
