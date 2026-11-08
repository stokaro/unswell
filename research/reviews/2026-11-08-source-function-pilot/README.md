# #349 source-function pilot: selective findings, insufficient recall

The source-only review and the four complete public README sources were frozen
before any model response. This was a directional pilot on four repositories
outside the 36-page development set, not prospective confirmation or a product
qualification. The packet SHA-256 is
`0cec9103172cf084dc277fb4110a92ae2764c5931823599eb21026fd97174a8f`.
The [source manifest](sources.json) records each upstream repository, commit,
Git blob, and SHA-256. The packet was locally frozen before execution; no
pre-execution Git commit proves that timing independently.
All four sequential GPT-6-Astra medium calls completed in one attempt, with
tool use disabled, valid source bindings, and no operational error. The run
journal SHA-256 is
`fa23ee46d207c9ac949797eebe71615626b86c56cf3f0cb4ff8ec02206a3da6b`.

| Measure | Result | Registered exploratory advance |
| --- | ---: | ---: |
| Fully recalled frozen events | **3/8 (37.5%)** | At least 7/8 |
| Partially addressed events, not recall credit | 2/8 | No credit |
| Accepted delivered diagnoses | **16/17 (94.1%)** | At least 85% |
| Uncertain delivered diagnoses | 1/17 | Remains in denominator |
| Valid complete pages | 4/4 | 4/4 |
| Supplied edits / edit safety | 0 / not applicable | Separately report |
| Protected technical counterexamples flagged | 0 | 0 |

Strict full-event recall by page was Caddy 1/2, restic 1/2, Gitea 0/2,
and OpenTelemetry Collector 1/2. The model located the unsupported Caddy
comparison, a vague restic endorsement, and a Collector agreement error. Its
second Caddy finding addressed the unsupported competitor comparison but left
the frozen event's intensifier; its Gitea screenshot finding corrected grammar
but left the frozen event's rhetorical lead. Both were recorded as partial,
not FULL. It missed the restic absolute user-behavior claim, Gitea's unsupported
superlatives, and the Collector performance tautology. The full 17-finding
assistant review is `review.json` (SHA-256
`cdc62695abfa0d3f38693e04709be6c6fb3a7d1ffdf598834057eb503a9e471f`).
The single uncertain finding concerns external sponsor copy in Caddy. These
are same-assistant judgments under ADR 0041, not human labels.

One frozen Gitea event is inside an HTML `<summary>` that the current engine
excludes. It stays in the preregistered 8-event denominator. The secondary
eligible-text count is 3/7, still far below the advance threshold. The raw
model saw the full source, while the engine CLI excludes that HTML block; this
limits the direct comparison. No post-result correction to the labels was made.

For an exploratory product reference, the exact `origin/main` snapshot
`61ea5ed12ca1e12aa14938ada443267d4a3e42f6` was archived and built.
The CLI technical profile scanned the same source bytes through stdin: four
Caddy findings, zero restic, two Gitea, and two Collector findings, all with
complete status and no errors. One Caddy finding spans both unsupported
promotion events. The Gitea long-sentence finding overlaps its superlative
sentence but does not diagnose the superlatives. This reference was added
after the model packet freeze and was not a preregistered paired efficacy
comparison; its eight findings were not separately editorially graded. It
does not establish a product improvement. Source SHA-256 values in the CLI
reports match all four frozen source files.

The four calls used 66,992 input tokens and 2,674 output tokens; monetary cost
is unknown. No source was edited. The result SHA-256 is
`474e54259847b6abbd73c6740085c8766011175e90270b0d01ea43d55f6b3417`.
The registered advance condition failed on recall. This exact prompt is
retired. The next hypothesis must prevent the model from selecting easy
grammar findings while missing source-specific semantic claims, and must be
tested on a new frozen page set before the required full-cohort screen.

The independent verifier pins the artifact hashes and reconstructs the packet, source hashes and Git blobs,
tool-free traces, response bindings, complete-stream review inventory, result
arithmetic, and CLI source identities. Five negative controls reject tampered
requests, a rehashed tampered packet, a missing review row, a fabricated pass,
and tool use in a trace.
Run it and its tests with:

```sh
make check-source-function-pilot
```
