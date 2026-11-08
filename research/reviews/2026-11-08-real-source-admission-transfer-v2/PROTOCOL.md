# Real-source admission transfer v2 for #349

Status: frozen offline preparation until this packet and runner are committed.
This is a distinct run with the same selected sources, old answer key, model,
prompt, schema, scoring thresholds, and stop policy as the terminal v1 packet.
It is an exposed development screen, not product qualification.

## Why v2 exists

The first v1 CLI process ended before any model response with
`failed to initialize in-process app-server client: Operation not permitted`.
Its stdout event stream is empty and no usage was recorded. V1 stopped at the
first failure as registered and is not retried. `codex --version` in the
restricted shell gave a PATH-alias permission warning; the same command with
system access succeeded without the warning. V2 changes only the host
execution permission: the runner is launched with the approved unsandboxed
shell permission so the Codex CLI can initialize its internal client. Model
tool access remains disabled through the inherited strict CLI configuration.
If this permission is denied or a request fails, v2 stops at once without a
retry or repair.

## Frozen transfer question

The deterministic `packet/sample.json` from v1 contains 24 source-hash-bound
candidates on 16 complete exposed pages across 12 repository groups. Its
SHA-256 is `46d9c604d83a023049af15bf0db1dd54fbc4cf13500f21d3fe4104bce9064206`.
`packet/key.json` contains 12 prior `accepted` and 12 prior `uncertain`
assistant grades. These are fixed comparison labels, not human ground truth.
Each fresh ephemeral Codex invocation receives one complete public source,
one target and diagnosis, and no grade, other candidate, earlier response,
reference event, local file path, tool, web access, MCP, plugin, or agent.
The exact prompt and JSON schema come from the merged synthetic pilot.

The user's standing authorization covers the bounded execution. V2 permits
at most **24 sequential GPT-6-Astra medium requests via the existing Codex
subscription**, one attempt each, 600 seconds per call and 10,800 seconds
overall. No paid API route or parallel call. Preserve the raw request bytes,
CLI event stream, stderr, response, source binding, and run journal. Stop at
the first operational, timeout, schema, trace, resource, or source-binding
failure. Invalid and unattempted items remain in the 24-item denominator.
The answer key, packet hashes, runner, scorer, settings, budget, and this
protocol are committed before the first v2 model request.

Count `admit` as retained for each archived accepted diagnosis and `reject`
or `abstain` as withheld for each archived uncertain diagnosis. Review every
actual admitted replacement separately against the complete source under
ADR 0041, naming the Codex assistant reviewer and never a human annotator.
Advance only with 24 valid calls and exact bindings, at least 21/24 agreements,
at least 10/12 accepted diagnoses retained, at least 10/12 uncertain diagnoses
withheld, and zero unsafe or unresolved admitted edits. The balanced sample
cannot estimate real delivery precision or event recall. Passing only permits
a separately frozen complete 36-page/123-event development screen at the
unchanged #349 gates: at least 80% FULL-event recall and 85% accepted actual
deliveries in each complete cohort, followed by separate prospective
confirmation. A failure retires v2 without tuning the seen cases or shipping
an unqualified candidate.
