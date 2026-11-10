# #349 additive generation and admission: negative real-source pilot

**Decision: terminal negative screen.** The source-only reference and runner
were frozen before inference. The complete model packet was committed as
`aac36b52f46ee4941bd808fa2045b7d61b9a242f` before its first call; packet
freeze SHA-256 is `78e1d02fe4b52871e3d208f021f73eaccf11ddaad5e0be84195e8f8809549b88`.
The sources, exact eligible units, original UTF-8 spans, source-only review,
prompt, response schemas, model, effort, limits, and failure policy were not
changed after the run. No earlier #349 reference label or prospective page was
used to select a response.

## Execution and reviewed outcome

Four sequential GPT-6-Astra medium generation calls and eight separate
admission calls completed through the existing Codex subscription. All calls
were ephemeral and tool-free, with one attempt each. The runner validated
all responses against their schemas, source hashes, eligible unit ordinals,
original byte spans, and protected boundaries. There were no operational
errors, retries, invalid pages, or protected targets. The 12 calls took
131.884 seconds in total. The CLI reported 227,337 input tokens, including
7,936 cached input tokens, 2,641 output tokens, and 490 reasoning output
tokens; monetary cost is unknown.

| Frozen pilot measure | Observed | Advance requirement |
| --- | ---: | ---: |
| Complete valid pages | 4/4 | 4/4 |
| Generated diagnoses | 8 | At most 24 |
| Admission decisions | 8 admit, 0 reject, 0 abstain | Separate review of every candidate |
| Accepted admitted diagnoses | **4/8 (50%)** | At least 85% |
| Uncertain / rejected admitted diagnoses | 3 / 1 | Count in delivery denominator |
| Frozen FULL events | **0/4** | At least 3/4 |
| Retain controls hit | 0/4 | 0 |
| Supplied edits | 0 | Report safety separately |

The post-response complete-source reviewer was a Codex assistant under
ADR 0041, not a human or independent annotator. [Review decisions](review.json)
cover all eight admitted findings. The admission stage admitted every
candidate, including three uncertain and one rejected diagnosis. It did not
improve delivery precision on these real pages. Zero edit proposals make
suggestion safety not applicable; they do not prove that a later edit would
preserve meaning.

The model missed the frozen Nushell customization sentence and eza's `one
single binary` redundancy. Its cargo-binstall finding quoted the frozen
sentence but diagnosed an unsupported reliability claim; the frozen event
concerned repeated hedges and anthropomorphic wording, and the reviewer found
the proposed reliability critique uncertain. The accepted eza diagnosis of
`Like seriously, we have a lot of good stuff` named the same repeated empty
promotional content as one frozen event, but used `needless_repetition` where
the frozen event and precommitted scorer require `empty_framing`. It receives
**no formal FULL credit**. Even a post-hoc category-equivalent sensitivity
would give only 1/4, below the registered 3/4 gate. No label, category,
response, or scorer was changed to improve this result.

## Interpretation and replay

This test supplies a clean source-binding result and a negative admission
result on four purposively chosen, repository-disjoint pages. It does not
measure the current engine's combined stream, transfer to the original
36-page/123-event cohorts, or separate prospective confirmation. It does not
show that source-conditioned admission in general is impossible. This exact
packet is retired: a next design needs independent real-source controls that
make unsupported diagnoses distinguishable, and any category-equivalence
rule must be frozen before scoring. Issue #349 remains open with its unchanged
80% FULL-event and 85% accepted-delivery requirements.

The raw requests, tool-free traces, responses, bound findings, and run journal
are in [run/](run/). The terminal journal SHA-256 is
`db144b02547dbb509beb5f20c9b015f5ac44700c71dcc7b9daf9d27d41865022`;
review SHA-256 is
`c4276d94aebb344e4e086ba380bfbd5ff49e96a6b863d0ce2af835429cc8f8b3`;
result SHA-256 is
`efe7cf92931dc40c5b081bf1eeb9604ae0d11e7050f7548c99a3fda26e7ff169`.
Recheck from the repository root:

```sh
make check-additive-admission-pilot
```
