# #349 exhaustive source inventory pilot

Status: negative directional result. This is a three-page exploratory screen,
not qualification of the 36-page/123-event development requirement or the
separate prospective confirmation requirement in #349. No product detector
behavior changed.

## Question and frozen inputs

The preceding four-page source-function pilot had 16/17 accepted diagnoses
but fully recalled only 3/8 frozen events. This pilot asked whether giving the
model a complete, source-bound inventory of every eligible prose unit before
semantic judgment would recover omissions. The two pilots use different pages
and denominators; their recall fractions are not a paired comparison.

Three complete public README files came from repositories absent from the
previous source groups. [The source manifest](sources.json) pins repository
commits, Git blob identities, source SHA-256, and unit counts. Exact bytes are
retained under `sources/`. The existing Go extractor produced
[the inventory](inventory.json): 169 eligible units in total, including sentence
units and fragments at protected boundaries. The
[source-only review](labels.json) froze four independently useful editorial
events and counterexamples before any model output. Its reviewer was the Codex
assistant in this session, not a human annotator. Unit and byte spans were
verified against the pinned sources.

The [V2 packet](packet/) pins prompts, schemas, `gpt-6-astra` medium, the
subscription route, one attempt per request, no tools, six sequential calls
maximum, a ten-minute per-call timeout, a three-hour total limit, and immediate
stop on the first error. Each first pass must return exactly one record per
exported unit ID. The second pass sees the full page and complete validated
inventory, and returns source-specific semantic diagnoses only. Labels,
reviewer rationales, and grades never enter the requests. The packet SHA-256
is `7624796d13927bb06a197a206862a9661b4e3ed83b4f870715a19d58b71c7fb8`.

[V1](prior-v1/) stopped on its first CLI attempt before any model response
because the host sandbox denied in-process app-server initialization. The V1
freeze SHA-256 is `750474ef632d504485e921d0f35f7432db961fc01ed732146d1a2a82345d6956`.
There was no V1 retry. V2 used the host process permission needed by the CLI
and versioned the protocol and verifier; source bytes, labels, and inventory
requests stayed fixed.

## Delivered result

All six V2 calls were tool-free, schema-valid, source-bound, and completed.
The first passes covered 86/86 uv units, 34/34 Zed units, and 49/49 Atuin
units. The second passes delivered three diagnoses. The
[complete-stream review](review.json) accepted all three, including one
unanticipated but useful Zed license-compliance diagnosis. That new finding
remains outside the frozen recall denominator; the gold labels were not edited.

| Page | Eligible units | Frozen events | Delivered | FULL | Partial |
| --- | ---: | ---: | ---: | ---: | ---: |
| uv | 86 | 2 | 1 | 0 | 1 |
| Zed | 34 | 0 | 1 | 0 | 0 |
| Atuin | 49 | 2 | 1 | 0 | 1 |
| **Total** | **169** | **4** | **3** | **0** | **2** |

The uv diagnosis identified the universal workflow-preservation clause but
missed the same event's unconditional 10–100× speedup. The Atuin diagnosis
identified the broad secret-safety guarantee but its deletion test left the
separate unbounded `Everything` claim intact. Both are partial under the
predeclared FULL standard. The unsupported uv adoption claim and ambiguous
Atuin history-replacement claim were missed. No protected counterexample was
flagged. The model supplied no edits, so suggestion safety is not applicable.

[The scored result](result.json) is 0/4 FULL, 2 partial, 2 missed, and 3/3
accepted delivered diagnoses. It fails the pilot's directional advance rule;
it does not demonstrate improved recall. The single assistant review cannot
establish population accuracy or agreement between human raters.

## Reproduce and interpret

From the repository root, run:

```sh
python3 -B research/reviews/2026-11-08-exhaustive-inventory-pilot/verify_result.py
python3 -B -m unittest discover -s research/reviews/2026-11-08-exhaustive-inventory-pilot -p 'test_*.py'
```

The verifier checks the frozen packet, both attempt records, all six tool-free
traces and response hashes, every unit ID, dynamic second-pass request bytes,
original quote spans, complete-stream dispositions, and result arithmetic.
Negative tests prove that missing units, unbound quotes, changed results, and
a fabricated V1 response are rejected.

The result isolates a useful limit: exhaustive input accounting prevented
silent omission of source units, but the judgment pass still selected only one
diagnosis per page and left material defects unresolved. Repeating this prompt
on the same exposed pages would not provide a new blind test. Any next candidate
needs a distinct mechanism, new frozen repository-disjoint pages, and the full
#349 development and confirmation gates before a product quality claim.
