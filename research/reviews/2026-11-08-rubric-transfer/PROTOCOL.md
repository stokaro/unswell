# Repository-disjoint editorial rubric transfer for #349

Status: offline preparation. No model call may occur until the full packet,
runner, scorer, response schema, and stop policy are committed.

## Question

The source-conditioned admission request in the merged real-source transfer
screen retained 11/12 previously accepted diagnoses but withheld only 3/12
previously uncertain ones. It treated several optional rewrites as useful
edits and admitted one meaning-changing edit. This paired development pilot
tests whether showing prior editorial decisions from other repositories
improves selectivity on complete Ptah pages without losing accepted cases.
It does not change those earlier results or any frozen reference event.

The control request is byte-for-byte the merged admission prompt and schema
for the current source, target, and diagnosis. The treatment adds 22 archived
examples: one previously accepted and one previously uncertain candidate from
each of 11 non-Ptah repository groups, with a 240-byte source flank, target,
diagnosis, and the prior Codex assistant reason. The examples are development
grades under ADR 0041, never independent human labels. Their statuses and
reasons are visible to the treatment model; the 12 test-case grades are not.
Both arms receive the same complete original public source and exact original
UTF-8 target byte offsets. Each call uses a fresh empty working directory,
with tools, web, plugins, MCP, and agents disabled.

The 12 fixed test cases are selected by SHA-256 from 12 distinct Ptah pages
outside the 16 pages in the prior real-source transfer sample: six prior
`accepted` and six prior `uncertain` candidates, each with one 20–300-byte
source-matched target. `packet/key.json` holds their old grades separately.
The archived grade and source banks are SHA-pinned in `sample.py`. The sample
is exposed development material, not prospective confirmation. No case or
page is replaced after seeing a model response.

## Execution and failure policy

The user's standing authorization for necessary model work covers this
bounded pilot. At most 24 sequential GPT-6-Astra medium calls through the
existing Codex subscription are permitted: one control and one treatment
per case, one attempt each, no paid API route or parallel calls. Maximum
600 seconds per call and 10,800 seconds overall. Stop at the first CLI,
timeout, schema, trace, source-binding, or resource error. Never retry or
repair a response. Invalid and unattempted items remain in the 24-call
denominator. Preserve exact requests, events, stderr, bound response, and
run journal. Monetary subscription cost is unknown.

## Registered readout

For a prior `accepted` candidate, `admit` is retention; for a prior
`uncertain` candidate, `reject` or `abstain` is withheld. Report both strata
for each arm and the case-by-case paired changes. Review every admitted edit
in both arms separately against the complete source under ADR 0041, naming
the Codex assistant reviewer and recording safe, unsafe, or unresolved.
Never infer that an old uncertain grade proves the diagnosis false. An edit
that changes technical meaning, negation, conditions, identifiers, or a
protected boundary is unsafe; unresolved meaning is not safe.

Advance only if all 24 calls and source bindings are valid, treatment agrees
with at least 10/12 old grades, treatment improves by at least 3/12 over the
paired control, treatment retains at least 5/6 accepted diagnoses, treatment
withholds at least 5/6 uncertain diagnoses, and every treatment-admitted edit
is safe. Passing permits a separately frozen larger complete development
stream, not runtime integration or #349 qualification. A failed or invalid
pilot retires this exact prompt without threshold tuning on these pages.

The original #349 requirements remain at least 80% FULL-event recall and
85% accepted actual deliveries on each complete cohort, with abstentions,
invalid pages, and edit safety reported separately, followed by the
independently frozen prospective confirmation. Category-only selection is
not a substitute: on the archived exposed bank, retaining only four
high-precision categories covered just 6/57 whole-page events, versus the
required 46/57. That is a post hoc feasibility audit, not a fitted policy.
