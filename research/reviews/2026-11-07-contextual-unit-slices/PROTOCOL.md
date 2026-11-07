# Exact raw unit slices for contextual review

Status: offline source-localization preparation for [#349](https://github.com/stokaro/unswell/issues/349),
November 7, 2026. This protocol was written before generating or checking the
new packets. It does not authorize or include model inference.

## Hypothesis and scope

The original 36-page contextual packet gives each engine paragraph an ID and
an original UTF-8 byte span, but leaves the model to reconstruct that slice
from the complete source. The terminal response quoted `**Deployment reports**`
from `whole/p009.mdx` although the relevant unit begins at
`Deployment reports**`. Giving the exact raw slice with each existing unit
should make that boundary visible and allow a response to quote only bytes the
existing binder can accept. This addresses source localization. It does not
predict whether the model will find the right editorial defect.

The only source is the previously frozen
`research/reviews/2026-09-26-contextual-evidence/evidence.tar.gz`, SHA-256
`95e04b63062b0e06aed325126f2726b69e4d6816872fb1eb9b62018832c3502b`.
Only its 36 `packet/requests/page-*.json` members and original request hashes
are used. Source bytes, source hashes, format, unit IDs, original byte spans,
protected ranges, and whole-document source text remain unchanged. Every unit
gains `raw`, the exact UTF-8 decoding of its byte slice, and `raw_sha256`.
No Markdown parser, tokenizer, normalization, or new unit boundary is used.

The 123 existing reference events, all prior responses and failures, review
labels, complete-cohort denominators, and reserved prospective pages remain
unchanged. The new packets are development inputs only. They are generated
locally from the committed archive; their committed manifest pins every
packet hash. No third-party source copy is added to the repository.

## Offline acceptance

The checker must reject a changed archive, source hash, original request,
unit ID/span, missing or altered raw slice, invalid UTF-8 boundary, or a
changed generated packet. It must verify all 36 sources and every unit, with
an exact file-set comparison against the manifest. The historical
`**Deployment reports**` quote must remain invalid in its owning unit;
`Deployment reports**` must bind to the original bytes there. Tests include
synthetic non-ASCII text and repeated quotes.

Passing this check proves only a sound source map. A later model experiment
needs its own frozen prompt, response contract, budget, failure policy, and
evaluation protocol before any call. It must report full-event recall,
accepted delivered findings, availability, and advice safety separately under
the unchanged #349 gates.
