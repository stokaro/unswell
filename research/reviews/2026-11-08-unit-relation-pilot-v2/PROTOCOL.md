# V2 continuation of the #349 mandatory unit verdict pilot

[V1](../2026-11-08-unit-relation-pilot/README.md) stopped after exactly one
GPT-6-Astra medium response. Its first Fastfetch inventory response returned
all 88 requested unit IDs, but used `Title` as a descriptive function. The V1
validator rejected that value because it expected an eight-label vocabulary
which the prompt did not specify. The V1 first-error journal, raw trace, and
signed pre-call packet remain unchanged and terminal.

V2 changes only that response contract: a function is now any nonempty,
descriptive string of at most 120 characters. The first response is copied
byte-for-byte into `inherited/`, along with its request, trace, V1 terminal
result, and V1 freeze manifest. The original source pages, 388 eligible units,
five source-only assistant events, counterexamples, prompts, schemas, request
bytes, candidate relations, verdict rubric, and editorial thresholds remain
unchanged. The inherited response passes the revised contract. V2 does not
repeat the first model call. Its new packet is frozen in a signed commit before
any of the remaining calls.

The new execution permits at most eight sequential GPT-6-Astra medium calls
through the Codex subscription, one attempt per request, no tools, 600 seconds
per call and 10,800 seconds overall. One inherited plus eight new responses
would complete the original nine-response sequence. Stop immediately on any
CLI, schema, source-binding, completeness, candidate-budget, or trace error.
Do not repair or retry an error inside V2. The first pass accounts for every
eligible unit; the second pass must give one diagnosis or abstention for each
unit marked narrower, contradicted, or unestablished. The 64-candidate-per-page
limit is a hard failure, never a silent truncation.

The unchanged assessment requires at least 4/5 FULL frozen events, at least
85% accepted actual delivered diagnoses, four valid complete pages, no
protected counterexample flagged, and no supplied edit. Every diagnosis gets
assistant review against the complete source; uncertainty remains in the
delivery denominator. Invalid pages, abstentions, and suggestion safety are
reported separately. This selected, exposed development pilot cannot qualify
#349 or establish AI authorship. An advance still requires a separate
repository-disjoint pilot, the full unchanged 123-event development accounting,
and untouched prospective confirmation.
