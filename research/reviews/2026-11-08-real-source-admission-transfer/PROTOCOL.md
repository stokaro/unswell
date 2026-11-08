# Real-source admission transfer screen for #349

Status: frozen offline preparation until the packet and runner are committed.
This is an exposed development screen of diagnosis admission, not a detector
qualification or a change to the frozen 123 reference events.

## Hypothesis and input

The exact source-conditioned prompt and response schema from the synthetic
pilot in `../2026-11-08-evidence-flip-pilot` may distinguish a useful supplied
diagnosis from an unresolved one on complete real technical pages. The
synthetic pilot passed its registered 15/16 decision screen, but its sources
were constructed to flip one contextual fact at a time. Real pages contain
release voice, distant qualifications, protected syntax, and multiple valid
criticisms at the same span; this transfer is not established.

The preselected `packet/sample.json` contains 24 source-bound candidates from
16 of the original 36 exposed pages and 12 repository groups. A fixed SHA-256
sort selected one archived `accepted` and one archived `uncertain` candidate
per group; BuildKit has no archived `uncertain` candidate and is excluded.
Eligible rows came from six editorial categories, had one 20-300-byte target,
matched the original source hash, and avoided the nine previously reviewed
maintainer-pair target spans. The selection seed, archive hashes, item IDs,
original UTF-8 target offsets, and candidate diagnoses are retained in the
sample. `packet/key.json` preserves the old judgments; these are prior Codex
assistant grades, not independent human labels or fresh ground truth. They
remain unchanged even if this screen disagrees.

Each request has exactly one candidate, the complete original public source,
its SHA-256, exact target quote and byte offsets, and the supplied diagnosis.
The request does not contain the old grade, other sampled candidates, original
reference events, pair answers, or prior model responses. One fresh ephemeral
Codex invocation in an empty directory handles each request; tools, web,
plugins, MCP, agents, and repository instructions are disabled. The prompt,
JSON schema, and source-binding checks are reused unchanged from the synthetic
pilot and pinned by hash. A different legitimate defect at the same target
cannot justify admission of this diagnosis.

## Budget and failure policy

The user has authorized necessary model work without repeated questions.
Run at most **24 sequential GPT-6-Astra medium calls through the existing
Codex subscription**, one attempt per request, at most 600 seconds per call
and 10,800 seconds overall. No paid API route, retry, repair, or parallel
call. Stop at the first CLI, trace, schema, source-binding, timeout, or
resource error. Preserve complete raw requests, CLI events, stderr, model
responses, and the run ledger. Monetary subscription cost is unknown.

The complete packet, answer key, runner, scorer, exact model settings, stop
policy, and thresholds are committed before the first model request. Score
only after execution stops. Invalid and unattempted calls remain in the
24-item denominator.

## Registered decision and limitations

For an archived `accepted` candidate, count `admit` as retention. For an
archived `uncertain` candidate, count `reject` or `abstain` as withheld.
`uncertain` is a nonaccepted delivery grade, not a proven false diagnosis.
Report those two strata separately; do not infer prevalence, calibrated
precision, original-event recall, or product quality from their balanced
sample. Independently review every actual admitted edit in complete source
context under ADR 0041 and identify the reviewer; an assistant is never
recorded as a human annotator. Preserve technical meaning and protected
boundaries. An admitted edit with unsafe or unresolved meaning fails safety.

Advance only if all 24 calls and source bindings are valid, at least 21/24
archived-grade delivery decisions agree, at least 10/12 accepted candidates
are retained, at least 10/12 uncertain candidates are withheld, and every
admitted edit is assessed safe. A pass permits a separately frozen complete
36-page development run measuring all 123 unchanged reference events, all
actual delivered findings, abstentions, invalid pages, and suggestion safety.
The unchanged #349 qualification gates remain at least 80% FULL-event recall
and 85% accepted actual deliveries in each complete cohort, followed by
separate prospective confirmation. A failure retires this exact transfer
configuration; do not tune thresholds on the 24 seen cases and call it
confirmation.
