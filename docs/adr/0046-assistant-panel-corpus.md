# ADR 0046: Assistant-panel corpus for a contextual diagnostic

Status: proposed on November 11, 2026. It takes effect when the maintainer
merges it.

## Context

[#349](https://github.com/stokaro/unswell/issues/349) asked for 80% complete-event
recall and 85% accepted findings on 123 reference events. One assistant
annotated those events on 36 selected pages. Between September 27 and
November 11, 2026, more than 60 frozen pilots failed that screen. The engine
reaches 45 of the 123 events.

Three facts explain the result.

- The reference is too small to learn from. The source-unit models were fitted
  on 122 to 140 units known to need an edit. About 4,300 other units on the
  same pages have no label.
- The reference is exposed. Each hypothesis received one frozen run on the same
  pages, and no attempt could be tuned afterwards.
- The planned corpus does not exist. [#22](https://github.com/stokaro/unswell/issues/22)
  requires 5,000 units labeled by two human raters, and no human rater is
  available. [ADR 0041](0041-assistant-review-acceptance.md) accepts assistant
  review for rules and excludes training data.

## Decision

Build a corpus annotated by a panel of assistants and fit a local contextual
diagnostic on it. Retire the 123-event screen as an acceptance gate.

**Annotation.** Two assistants from different model families annotate each
complete page independently under rubric `unswell-editorial-v1`. Each answer
labels a prepared unit as acceptable, needs revision, or uncertain, with
categories and an exact source quote. A unit is positive when both annotators
mark it as needing revision. It is negative when both mark it acceptable. Every
other unit is masked in training and evaluation. Publish raw agreement and
Krippendorff alpha on the original answers.

**Records.** Store each answer with the model identity, effort, prompt hash,
and packet hash. An assistant answer is never serialized as a human answer and
never fills a human slot of [ADR 0013](0013-annotation-protocol.md). A model
served from a private endpoint is recorded as `private-host/<model>`; records
name neither the endpoint nor its operator. The existing generation records
use the same form.

**Size and partitions.** Collect at least 5,000 units with a consensus label.
At least 1,000 of them are positive, and at least 1,000 belong to the final
test. Split by repository under [ADR 0014](0014-corpus-acquisition.md). The 36
exposed pages and their repositories stay out of the final test.

**Development.** Work on the training and development partitions is iterative.
An attempt needs no frozen protocol, pull request, or issue comment. Choose the
model and its threshold on development data. Open the final test once per
release candidate. An opened final test becomes development data.

**Audit.** A third assistant from another model family audits delivered
findings. It sees the original passage and a version with the flagged unit
removed or rewritten, in random order, without labels or model output. It
states which version serves the page better and whether information is lost.
The same packet contains placebo items: unflagged units edited the same way.
The audit is valid only when the auditor accepts at most 10% of placebo edits.
Also report how many sentences that maintainers removed or rewrote in public
editorial commits receive a finding.

**Acceptance.** The diagnostic ships as an advisory rule family when both
conditions hold on the final test at the frozen threshold:

- at least 85% of delivered findings are consensus positives;
- a valid audit confirms at least 85% of a sample of 300 or more delivered
  findings.

Recall at that threshold is reported with its interval. It is a progress
measure, not a release requirement.

**Stop rule.** Fit the model on 10%, 25%, 50%, and 100% of the training units.
Stop this path and record a negative result when development recall at 85%
precision is below 20% and grows by less than two points from 50% to 100%.
Enlarge the corpus when recall still grows.

## Consequences

The 123 events remain a regression replay. They no longer decide acceptance,
and #349 closes with its negative result recorded.

Metrics describe agreement with a named panel of assistants. They do not
measure agreement with a native-speaking editor. Models can share a preference
that readers do not have. The placebo items and the historical edits bound that
risk without removing it.

The fitted model runs offline inside the engine as a loaded pack. Its output
is a finding with an index. It is not a probability and never gates a build.
The human-labeled corpus of #22 and the calibrated probability of
[#25](https://github.com/stokaro/unswell/issues/25) stay deferred with their
requirements unchanged.

[ADR 0037](0037-diagnostics-not-authorship.md) stands. Origin detection is not
part of this path.
