# Blinded evidence-flip admission pilot for #349

Status: frozen offline preparation until the packet is committed. This is an
exposed synthetic capability screen, not an original-event or prospective run.
No product rule, reference label, threshold, or source page changes.

## Hypothesis

A source-conditioned judge that receives one proposed diagnosis and the whole
short source at a time can distinguish a needed edit from a plausible criticism
whose supporting context has been removed, supplied, or contradicted. This
separates diagnosis admission from source localization and from author-origin
classification. The prior generic candidate selector, lexical retrieval,
source-unit cross-encoder, and maintainer-pair pilots failed; their outcomes are
not pooled with this pilot.

The input is the 16 source-hash-verified controls merged in PR #454. Four
quartets keep their target and *specific* diagnosis fixed while changing the
surrounding evidence. Eight controls support the diagnosis and eight do not.
Each call receives exactly one variant, an opaque ID, its complete source,
target quote, original UTF-8 byte offsets, and the proposed diagnosis. It does
not receive the case family, condition, label, rationale, another variant,
previous response, original #349 event, or maintainer grade. A new ephemeral
Codex invocation and empty working directory are used for every call; tools,
web access, MCP, plugins, and agents are disabled.

## Execution budget and stop policy

The user's standing instruction authorizes the needed work without repeated
permission questions. This run is bounded to **16 sequential GPT-6-Astra medium
calls through the existing Codex subscription**, one attempt each, at most
600 seconds per call and 7,200 seconds overall. No paid API route, automatic
retry, response repair, or parallel request is allowed. Stop at the first CLI,
trace, JSON Schema, source-binding, resource, or timeout error. Preserve raw
requests, CLI event streams, stderr, and response bytes. Monetary subscription
cost is unknown.

The packet, scorer, runner, source hashes, model, effort, limits, support-flip
pairs, and failure policy are committed before execution. The request schema
requires a decision (`admit`, `reject`, or `abstain`), exact echoed source
identity and target bytes, a concise basis, and a replacement for admitted
edits. An admitted empty replacement deletes the target. Abstentions and
invalid or unattempted calls remain in the 16-case denominator.

## Registered decision

Advance only if at least **14/16** fixed-diagnosis decisions are correct, both
members of each of the four registered support-flip pairs are correct, all
source coordinates are exact, and every admitted replacement has a separate
assistant review marking it safe rather than unsafe or unresolved. An
alternative valid criticism at the same target does not count. The assistant
reviewer is identified under ADR 0041 and is never called an independent human.

A pass permits only a separately frozen run on all 36 exposed pages and all
123 unchanged original events under the existing **80% FULL-event recall and
85% accepted delivered findings in each complete cohort**. Suggestion safety,
coverage, invalid pages, and full combined-stream quality remain separate.
Prospective pages retain their confirmation role. A failure retires this
configuration without retuning on the same controls or changing product code.
