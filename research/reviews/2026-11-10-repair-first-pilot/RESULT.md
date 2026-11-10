# #349 repair-first pilot: stopped on its frozen edit contract

**Decision: terminal, unqualified contract failure.** The packet was committed
as 77883e15cc73eb4086a78e0dd9effe347f33cf70 before the first response.
Its freeze SHA-256 is
77b77c528d6e4e5ddfd0bd80cccf1e0b3b6a67b9e45e9ab7100ff25ce9f2f6a0.
The four public source pages, four source-only reference events, four retain
controls, model, effort, prompt, schema, binder, score policy, limits, and
failure rule stayed unchanged. No original #349 labels or prospective pages
were included in model requests or changed.

Two sequential GPT-6-Astra medium calls were made through the existing Codex
subscription. Nushell returned a schema-valid response with three edits; all
three bound to exact eligible original UTF-8 spans. Hyperfine returned a
schema-valid response with three edits, but its first replacement was 50 UTF-8
bytes against a 46-byte quote. The frozen binder requires the replacement to
be no longer than its quote, so it rejected the entire page. The runner stopped
at the first error. Eza and cargo-binstall were not called, and no retry or
response repair was made.

The rejected Hyperfine proposal changes "perform a certain number of program
executions" to "perform the specified number of program executions." Its
extra four bytes are the exact cause of the stop. This demonstrates that byte
length is an unsuitable hard proxy for minimality: it can reject a plausible
clarification before editorial safety review. It does not establish that the
proposal is accepted or safe, because the page was invalid and not scored.

The one valid Nushell response omitted the page's frozen wordiness event,
"From here one can then customize this file for their specific needs." None
of its three bound edit spans overlaps that target. This is a directional
location observation, not an accepted-finding fraction or a FULL-event score.
An offline post-stop inventory audit confirmed that all four frozen positive
events and all four retain controls occur verbatim in their frozen requests,
each wholly inside one eligible segment at its registered original byte span
and unit ordinal. The audit joined
[`reference.json`](../2026-11-10-additive-admission-pilot/reference.json) byte
spans to the frozen
request `source`, `raw_span`, and `eligible_segments` fields without changing
either input. The missed Nushell event is the entire raw slice of unit 33.
The target was available to the model in both the complete source and the exact
unit inventory; extraction and source eligibility did not hide it. This audit
does not establish why the model selected other edits or whether it attended
to the target.
No complete four-page review or combined detector claim is available. The
previous additive pilot's 0/4 FULL and 4/8 accepted result remains unchanged.

| Execution measure | Observed |
| --- | ---: |
| Planned pages | 4 |
| Attempted calls | 2 |
| Valid bound pages | 1 |
| Invalid pages | 1 |
| Uncalled pages | 2 |
| Bound Nushell edits | 3 |
| Nushell frozen event target overlaps | 0/1 |
| Reported input tokens | 58,853 |
| Reported output tokens | 1,401 |
| Reported reasoning output tokens | 466 |
| Reported monetary cost | Unknown; subscription route |

The run journal SHA-256 is
6d47583effc13e733943a713abad8d60f90ed0d11c8151dbde90c7373131eef6.
The Nushell bound output SHA-256 is
a30aff526bc117f26b9ed395142dd05a6c0f6cc5914510d37633f98cbcaa5872.
The invalid Hyperfine response SHA-256 is
4db80364985e6436300b43f6d23ba3a70b55fd052ac784bfeef15100edf50215.
Both raw traces and exact requests are preserved under run/. Both CLI stderr
files are empty. Rechecking the packet after the stop returned its original
freeze hash. The local contract and negative tests passed.

Do not relax the length limit and continue this packet on the remaining
pages. A future edit-based design would need a new, source-disjoint packet
and an explicit semantic retention check, not byte count alone. The first
valid page also missed its known target, so this small run provides no reason
to promote the current repair-first method. No runtime or scoring policy was
changed. Issue #349 remains open: its original 123-event complete-cohort
80% FULL / 85% accepted requirements and separate confirmation are unmet.
