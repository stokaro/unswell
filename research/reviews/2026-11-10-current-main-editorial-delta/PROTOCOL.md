# Current-main editorial delta for #349

## Frozen inputs

This is a paired replay on the same 36 complete pages and 123 labeled events
from the September #349 corpus. The September labels, source bytes, and
denominators stay frozen. The previous exposed source review is
2026-11-07-current-main-editorial-review, whose runtime was
5dd76191740c3d2f003b6ec90a9b381ccf1b772a.

The new runtime is exact main
3acba5eb4c13b95207caf238da17e9c0a72630e3. The binary SHA-256 is
e993feecba689b01014e440e29d9c7ee1ed55ca4ba37c5508fc0410529d5683d.
The committed runs.json identifies both complete technical and strict
reports. On the pinned commit with Go 1.27.3, rebuild using:

    CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags '-X github.com/stokaro/unswell.BuildCommit=3acba5eb4c13b95207caf238da17e9c0a72630e3' -o /tmp/unswell-349-rebuild ./cmd/unswell

This command reproduced the recorded binary SHA-256 from the research
worktree. Run the committed September replay.py on that binary to
independently reproduce the report bytes. The freeze.json hashes the two
reports, run record, and new review.

## Review transfer

Transfer a previous finding's source review only when the rule, original byte
location, message, evidence, advice, and all other report fields agree. The
normalization excludes only id, fingerprint, and rule_version. It does
not allow a changed diagnosis or suggestion to inherit an old judgment.
The verifier checks this for all 176 previous strict findings and confirms
that their technical/strict profile membership is unchanged. It also verifies
the previous review with its own frozen verifier.

The seven new findings are reviewed against their complete original source
pages, not their clipped snippets. new-review.json records each finding
disposition, repair action, event status, and rationale. summary.json
identifies the reviewer as the Codex assistant under ADR 0041, not a human annotator.
The seven findings each overlap one distinct, previously missed frozen event.
No frozen label was edited and no prior reviewed event was silently regraded.

An event is full only when an actionable technical finding covers every frozen
target and its diagnosis addresses the complete editorial burden. Partial
means a useful diagnosis with a remaining material gap. Conditional advice is
not a verified edit. Source review does not qualify model-origin detection.

## Validation and interpretation

Run python3 research/reviews/2026-11-10-current-main-editorial-delta/verify.py
and python3 research/reviews/2026-11-10-current-main-editorial-delta/test_verify.py.
The latter includes negative cases for a changed carried finding, an
unreviewed new finding, and a changed frozen target quote.

The result is a comparison of two exact runtime snapshots on identical pages.
It shows seven additional accepted diagnoses and four additional full events.
It does not isolate a single rule change, prove a safe replacement edit, or
meet the #349 full-event and accepted-finding targets. No new model calls
were used in this replay or review.
