# #349 additive generation and source-bound admission pilot

Date: November 10, 2026. Status: frozen source and review preparation; no new
model call has run. This is a small, source-disjoint, exposed development pilot,
not the complete 36-page/123-event evaluation or prospective confirmation.
The original #349 labels, denominators, product rules, and reserved pages remain
unchanged.

## Hypothesis

The prior joint pilot made one model response both propose diagnoses and withhold
existing findings. It lost four accepted baseline findings, and one proposed
finding crossed a protected source span. Here generation is additive: it may
only propose new source-bound diagnoses, and it cannot withhold existing engine
findings. A separate, ephemeral call judges each proposed diagnosis against
the complete source. The expected benefit is fewer unsupported proposals
without deleting earlier useful findings. This is a new test of that sequence,
not a replay of the retired joint packet or of the synthetic evidence-flip
preflight.

## Frozen sources and source-only review

`sources.json` pins four public README files by repository, Git commit, Git
blob SHA-1, original UTF-8 SHA-256, and byte count. The local `.source` files
match those blobs exactly. These repositories are absent from the prior
36-page #349 development corpus. They were named before any model response,
but this purposive set is not representative of documentation in general.
The complete `source-function-inventory` export uses the current technical-v1
extractor, protected regions, original byte spans, eligible segments, and raw
slices. `input.json` and `inventory.json` are frozen source artifacts. Their
hashes and the exporter binary identity must be recorded before execution.

`reference.json` records one Codex assistant's source-only full-page review:
four warranted editorial events across three pages and four explicit retain
controls across all four pages. The assistant is the actual reviewer under
ADR 0041, not a human annotator. The reference is deliberately small. It was
written without seeing outputs from this pilot. All four event and control
quotes are unique within their source and lie in eligible byte segments.
Do not add events or change judgments after a model response. New concerns
noticed during output review may be reported separately, never added to the
frozen recall denominator.

## Model and execution contract

Use GPT-6-Astra at medium reasoning effort through the existing Codex
subscription, with ephemeral tool-free sessions, read-only sandbox, and no
web access. The packet must freeze exact request bytes, JSON schemas, runner,
scorer, model, effort, and failure policy in a committed Git tree before the
first call. Total budget: at most 4 generation calls plus 24 admission calls,
sequential, one attempt each, no automatic retry, at most 3 hours total and
10 minutes per call. Stop on the first CLI, trace, schema, binding, source,
resource, or timeout error. Do not repair or replay a failed response under
the same packet. Record actual token usage; monetary cost through the
subscription is unknown.

Each generation request contains one complete source and its complete eligible
unit inventory, including ordinals, raw slices, exact byte spans, and eligible
segments. It contains no reference labels, prior model answer, or reviewer
grade. Request at most six diagnoses per page in `empty_framing`,
`wordiness`, `vague_claims`, or `needless_repetition`. The response must name
an eligible unit ordinal and quote exact source bytes inside one eligible
segment, plus a specific diagnosis and source-context basis. No replacement
text is requested or permitted. Grammar-only, factual-correction, AI-origin,
and generic style claims are outside the treatment. A repeated quote must be
unambiguous within the selected unit. Repeated or near-duplicate findings
remain in the raw delivered denominator; do not silently deduplicate them.

After all four generation responses bind, make one independent admission call
per generated diagnosis. Send that candidate, exact bound target, and the
complete source. The admission model must choose `admit`, `reject`, or
`abstain` for that *specific* diagnosis, with a source-specific basis. A
rejected or abstained candidate is not delivered. No candidate can suppress
an existing engine finding. If a page generates zero diagnoses, it makes no
admission calls and remains in the page and event denominator.

## Registered screen and accounting

The pilot advances only if all four pages are valid, at least 3/4 frozen events
receive FULL diagnosis credit, at least 85% of admitted diagnoses are accepted
on complete source review, no retain control is falsely diagnosed, and no
protected span is targeted. `FULL` requires both an eligible overlapping span
and the frozen source-specific defect, not location overlap alone. Report raw
generation count, accepted/rejected/abstained admission counts, every
admitted finding's review disposition, invalid and abstained page coverage,
and duplicate findings. Because the treatment supplies no replacement, edit
safety is `not applicable`; do not infer that future edits would be safe.

A pass authorizes only design of a larger independent development test. It
cannot establish a qualified baseline-plus-model system: that requires review
of the complete combined engine and model stream on all original 123 events,
80% FULL-event recall and 85% accepted delivered findings in each required
cohort, followed by separate reserved confirmation. The unchanged engine
baseline on these four pages may be reported as context, but is not removed or
credited to this additive treatment. If the pilot fails, publish the negative
result and retire this exact packet without changing its labels or thresholds.
