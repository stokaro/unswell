# Source-disjoint sentence-shape function screen for #349

Status: protocol fixed before source acquisition and editorial review. This is
an exploratory mechanism screen, not a detector qualification or a replacement
for the 123 frozen #349 events.

## Question

Can source context distinguish an unnecessary rhetorical sentence shape from
a necessary technical contrast, condition, or consequence? The five candidate
expressions are exactly those in the merged
`2026-11-09-sentence-shape-screen/screen.py`. A match is a location to review,
never an AI-origin judgment or an automatic edit.

## Source selection

Acquire the default-branch `README.md` at one immutable commit for each of the
following repositories, in this order, without replacing a page after seeing
its content. The first six form a design partition; the last six form an
unseen transfer partition for any selector derived from the design pages.

1. `go-chi/chi`
2. `go-task/task`
3. `spf13/cobra`
4. `jackc/pgx`
5. `rs/zerolog`
6. `nats-io/nats.go`
7. `helm/helm`
8. `golang-migrate/migrate`
9. `jmoiron/sqlx`
10. `grpc/grpc-go`
11. `kubernetes-sigs/kustomize`
12. `go-playground/validator`

These repositories are absent from the 13 source groups in the #349 exposed
archive. Pin each commit, Git blob, original UTF-8 byte count, and SHA-256.
Keep full source copies only in ignored local artifacts. A missing, invalid,
or unparseable page remains in the 12-page denominator with an explicit error;
do not substitute another page. Do not open reserved #349 prospective pages.

## Candidate and review procedure

Build `research/annotation/cmd/reviewtext` from the frozen Unswell tree and
use its existing `--source-inventory` contract on each pinned page. Validate
every raw unit against the original UTF-8 bytes and every eligible segment
against protected ranges. A regex match qualifies only when all its bytes are
within one eligible segment. Preserve the original byte span, complete source
line, nearby heading, and adjacent lines for review. Do not add a Markdown
parser or treat raw Markdown syntax as editable prose.

Review all eligible matches on the six design pages before defining a
contextual selector. Record each match as `warranted`, `retain`, or
`uncertain`, with a source-specific reason and a minimal diagnosis if
warranted. A warranted label requires a concrete editorial burden and a
meaning-preserving direction for revision; merely resembling generated text
is insufficient. A retain label identifies the technical fact, distinction,
condition, or document function that the phrase supplies. Uncertain matches
remain in the denominator. The reviewer is the root Codex assistant under
ADR 0041, not a human annotator. No replacement edit is supplied in this
screen, so suggestion safety is not measured.

If the design pages yield at least four warranted matches from two or more
repositories and at least four close retain controls, write and freeze one
explicit source-context selector before viewing transfer-page match contexts
or labels. The selector may abstain. It must make a source-bound diagnosis,
not only assign an AI-origin or style score. Its transfer screen reports all
matches, all selector outputs, warranted/retain/uncertain dispositions,
coverage, and any unchanged-CLI overlap. An exploratory advance requires
at least two newly warranted transfer diagnoses from two repositories,
at least 85% warranted among actual delivered diagnoses, zero retain-control
hits, and no source or protected-span error. This is directional evidence,
not the #349 complete-cohort gate.

If the design prerequisite fails, stop without defining a selector. If a
selector fails its transfer screen, retire it without adjusting its threshold
or expressions on these pages. Keep the unchanged 36-page/123-event #349
development cohorts and prospective confirmation separate. Do not change
product rules, suggestions, scores, or defaults from this screen.

## Failure and resource policy

Use no hosted model calls. Acquisition makes read-only GitHub requests and
stops at the first transport or identity error, without retries. The offline
screen runs one exporter process per page with a 30-second timeout. A parser
error or timeout counts as an unavailable page and cannot pass a transfer
gate. Limit each source to 200,000 bytes; record larger pages as unavailable,
without trimming them. Limit the candidate packet to 200 eligible matches
across the 12 pages; if exceeded, record the overload and stop before review.
No active fuzzing, race testing, or coverage run is part of this research.
