# Current-engine baseline for the exposed #349 cohorts

Status: exposed development replay, November 7, 2026. An exploratory run before
this freeze showed that current-main findings differ from the September report.
This frozen replay can make that comparison reproducible; it cannot restore
independence or supply editorial judgments for new findings.

The base is merged commit `5dd76191740c3d2f003b6ec90a9b381ccf1b772a`.
Build the public CLI from that tree with Go 1.27.3 on darwin/arm64,
`CGO_ENABLED=0`, `-trimpath`, `-buildvcs=false`, and `BuildCommit` set to the
base SHA:

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false \
  -ldflags '-X github.com/stokaro/unswell.BuildCommit=5dd76191740c3d2f003b6ec90a9b381ccf1b772a' \
  -o /tmp/unswell-baseline ./cmd/unswell
```

The binary SHA-256 is recorded in `freeze.json`; another toolchain or host
needs a fresh binary identity and should not claim byte-identical replay.
Reuse the already committed
`2026-09-26-outcome-announcements/replay.py` to load and verify all 36 original
documents and run both `technical` and `strict` profiles with no gate. Preserve
both complete raw reports and the run ledger. The original 123 exposed events
and their three cohort denominators remain unchanged. No reserved prospective
material, model call, or new rule is used.

The analysis checks every original-source hash and report span. It compares
current findings to the older technical/strict reports by rule, source path,
and primary UTF-8 byte range, because finding IDs and rule versions may change.
It records old and new finding sets without transferring old editorial grades.
For each event, it computes a deliberately generous **span-overlap ceiling**:
every annotated target must overlap at least one current finding's primary or
related source range. This is necessary for full credit under the frozen #349
evaluation contract, not sufficient: a length warning or unrelated criticism
can overlap the right bytes and still fail the event. Report the ceiling and
the unchanged 80% requirement separately for all three complete cohorts.

Stop on a changed frozen input, binary/report hash mismatch, missing source,
incomplete report, skipped rule, malformed span, or changed event denominator.
No threshold is fit and no accepted-finding fraction, suggestion safety, or
qualified baseline is inferred from the overlap ceiling. Actual complete
stream review remains a separate step before any #349 qualification claim.
