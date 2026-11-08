# Evidence-flip admission pilot: synthetic preflight passed

A frozen GPT-6-Astra medium pilot admitted or rejected **one supplied diagnosis
per request** against a complete synthetic source. The 16 source variants were
merged in [PR #454](https://github.com/stokaro/unswell/pull/454). Each variant
was sent in a separate ephemeral, tool-free call with an opaque ID; no answer
key, condition name, rationale, neighboring variant, prior answer, original
#349 page, or maintainer grade was sent. The original target and its UTF-8 byte
coordinates were supplied, so this pilot tests *admission* rather than
independent source localization or diagnosis generation.

The [protocol](PROTOCOL.md), all request bytes, frozen labels, response schema,
runner, and scorer were committed as `ec51040282424c696487caddd150a4afe86b7bf2`
before the first model call. The packet SHA-256 is
`a7dc99f1cacc137eaf8425d8e495c3cffde9f550da19f2ab2ee8f8e99d6b8b14`.
The run completed all 16 sequential calls in one attempt each, with zero tool
calls, schema failures, source-binding failures, or retries. The raw request,
event, response, stderr, and bound-response files are in [run/](run/).

| Registered measure | Result | Requirement |
| --- | ---: | ---: |
| Correct fixed-diagnosis decisions | **15/16** | At least 14/16 |
| Support-flip pairs with both sides correct | **4/4** | 4/4 |
| Original-byte source bindings | **16/16** | 16/16 |
| Admitted edits separately reviewed safe | **9/9** | No unsafe or unresolved edit |
| Operational errors, retries, tool calls | **0** | 0 |

The sole decision error was the `unsupported-reader-need` variant where a
usability observation says zero readers missed the flag. The model admitted
the diagnosis "without evidence about readers" even though the source
contains *contradictory* reader evidence. Its deletion of the contrary need
claim was safe as an edit, but the supplied diagnosis was inaccurate. The
[assistant safety review](review.json) records this distinction. Under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md), the reviewer
is a maintainer-accepted Codex assistant, not an independent human annotator.
No edit was applied to a real document.

The [scored result](result.json) passes the *registered synthetic preflight*.
The [independent verifier](verify_result.py) reconstructs request/source
identity, packet hashes, tool-free raw traces, every model answer, original
byte offsets, the frozen key, all 16 decisions, four pairs, and safety
reviews. Seventeen positive and negative tests exercise the packet, scorer,
and verifier. Summed usage was 182,390 input tokens, 2,184 output tokens, and
265 reasoning output tokens through the Codex subscription; monetary cost is
unknown.

This is a narrow positive capability result on exposed artificial controls.
The model saw the proposed diagnosis and the exact target. It does not add any
FULL-event witness or delivered finding to the 36-page, 123-event #349
screen, establish transfer to natural technical documents, validate advice on
real pages, or qualify the product. The next treatment must freeze how it
generates and admits candidates over the **whole** source stream, then meet
the unchanged 80% FULL-event recall and 85% accepted-delivery thresholds in
each complete cohort, followed by separate prospective confirmation. The one
contradiction-versus-absence error remains a concrete transfer risk.

Run the offline verification with:

```sh
make check-evidence-flip-pilot
```
