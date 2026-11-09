# #349 joint coverage and delivery triage pilot

Status: design frozen before the first model response. The four pages are
pages 2, 6, 8, and 10 of the separately frozen 12-page transfer inventory in
`../2026-11-09-fresh-source-function/`. They were selected for this exploratory
screen because their existing source-only labels span repetition, absolute
ranking, unsupported intensifiers, and empty framing, while their complete
sources include technical controls and benchmark tables. This selection used
exposed labels; it is neither a fresh blind transfer nor prospective
confirmation. No source-only label is revised by this pilot.

## Hypothesis

A single complete-page pass that reviews **every** existing engine finding
and independently inventories missed source-specific defects can improve both
parts of #349's joint gate. The model must explain the function of the target
in the full page, the reader burden of keeping it, and the useful content a
repair must retain. It may withhold a baseline warning when the source shows
that the wording performs a necessary technical or editorial job. This differs
from prior cue admission, bag-of-words source context, and source-only
generation passes: the model receives both the original page and all actual
delivered baseline findings, but no reviewer grades or event labels.

The existing public CLI, built from actual main `dbfe252b505294db7826cc59fc8c6b797650aa74`,
completed all four pages. `baseline.json` pins each complete report by SHA-256
and contains all 11 delivered findings with original byte spans and raw slices.
The complete reports are copied byte-for-byte into `baseline-reports/` and
checked against those hashes.
`sources.json` pins the four public source revisions and original SHA-256s.
The complete sources are retained in `sources/`. The prompt includes the full
numbered source and every baseline finding; it omits frozen labels, previous
candidate output, reviewer judgments, and prospective pages.

## Execution contract

Before execution, `run.py prepare` writes exact request bytes, the supported
JSON schema, the execution policy, and a SHA-256 freeze manifest. Commit that
packet before any call. `run.py verify` must reject any changed source, prompt,
schema, request, baseline, runner, or scoring policy. A change after execution
requires a new packet version, not an edited replay.

- Route: existing Codex subscription, no paid API route; `gpt-6-astra` at
  medium reasoning. Four sequential, ephemeral, tool-free calls maximum.
- One attempt per page, ten minutes maximum per call, two hours total. Stop on
  the first CLI, schema, trace, source-binding, or resource error. No retry,
  response repair, or continuation after a stop.
- For every baseline finding, return exactly one `keep`, `withhold`, or
  `uncertain` decision with a source-specific reason. `uncertain` is withheld
  from delivery and reported separately. Never infer a decision from omission.
- Return at most 20 new diagnoses per page. Each quotes an exact contiguous
  original UTF-8 slice and its 1-based source line. The binder calculates byte
  offsets and rejects ambiguous quotes. No replacement edit is requested, so
  suggestion safety is explicitly not measured in this pilot.
- The model cannot read repository files, browse, start tools, or see labels.
  Its trace and final response are preserved. An invalid page remains in the
  denominator.

The user explicitly authorized necessary work and asked not to be repeatedly
queried. The bounded model budget above is the maximum for this new run; this
pilot does not reuse the budget of a prior experiment.

## Evaluation and stop rule

The frozen source-only labels contain eight clear positive events across these
four pages. Review the full combined delivery stream: every retained baseline
finding and every new diagnosis, including duplicates, against complete source
context. Keep `uncertain` and rejected findings in their declared denominators;
do not silently recode them as accepted. For each frozen positive, report
FULL, PARTIAL, or MISSED against the combined delivered stream, with exact
source binding. Separately report baseline withholds, accepted findings lost,
control hits, abstentions, invalid pages, and any supplied advice (none is
requested). A location overlap alone is not FULL.

The pilot advances only if all four pages are valid, at least 7/8 positives
receive FULL diagnoses, at least 85% of the actually delivered combined
findings are accepted, no `retain` control is flagged, and no previously
accepted baseline finding is withheld by triage. Otherwise retire this exact candidate without
tuning on these exposed pages. Even a positive pilot is not product
qualification: the unchanged 36-page/123-event development cohorts still
require 80% FULL and 85% accepted delivery in each complete cohort, full-stream
review, measured advice safety, and separate frozen prospective confirmation.

The root Codex assistant is the reviewer under ADR 0041, never a human
annotator. This experiment concerns editorial diagnoses, not AI authorship.
