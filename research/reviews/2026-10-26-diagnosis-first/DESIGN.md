# Diagnosis before repair

## Evidence and hypothesis

The October 26 accounting run stopped at its tenth request after a TLS error.
On four complete context pages it detected 13 of 28 frozen defects, while 34 of
36 emitted findings were accepted. Of the 15 misses, 11 occurred in units marked
`retain` despite the existing instruction to inspect framing beside useful
facts. The observed result does not prove why the model omitted those defects.

Earlier source-only and edit-audit experiments also failed. The 176-call edit
screen explicitly reviewed useful facts and local wording; a contemporaneous
terminal note records an unsuccessful full-cohort screen. Its final replay
artifact is no longer present in this workspace, so this protocol does not use
those historical grades as a qualification input. More output or a generic
second audit is not an established fix. This pilot instead tests a
narrower interaction: it requires a separate technical-payload judgment and a
burden judgment for every owned unit, while removing the simultaneous request
for repair advice. A model may diagnose a burden without first solving a safe
rewrite. These two contract changes are evaluated together; the pilot cannot
attribute any result to one change alone or to model stochasticity.

## Source and evaluation boundaries

The input is derived mechanically from the previously frozen 36-page source
packet. Only the six previously exposed `context/c01.md` through `c06.md`
pages are sent in this pilot. Each request contains full-page source, protected
regions, every existing engine unit and its exact raw UTF-8 slice, and a disjoint
focus window of at most 64 units. The old source binder checks exact literal
occurrences, original byte spans, and eligible prose. No Markdown reparse,
normalization, label-guided selection, prior diagnosis, source edit, or repair
is introduced.

All 36 pages and 123 original reference events remain in the offline review
ledger. The pilot screen requires all six context pages, at least 37/46 FULL
events, and at least 85% accepted diagnoses across every emitted finding.
Events on pages without a complete validated response are recorded as
`unavailable`, not `missed`. Their frozen page and event counts remain visible;
cohort recall is undefined until every page in that cohort is available.
Partial-page emissions and uncertain diagnoses remain visible in their
denominators. The report separately counts failed and unattempted requests,
invalid and unattempted pages, unassessed units, and `uncertain` burden
decisions. An
assistant may perform the source review under ADR 0041, but is
identified as an assistant, not a human annotator. Suggestion safety is not
applicable because this pilot emits no suggestions. Any later repair stage must
be separately specified and evaluated before product use.

## Execution and decision

The proposed execution is seven sequential GPT-6-Astra medium requests through
the existing ChatGPT subscription, at most ten minutes per request and 90
minutes overall. It uses no tools, retry, repair, resume, fallback model, or
paid API. It stops at the first error, including a CLI reconnect error event,
and records hashes for every raw trace, stderr file, and response, including
failed calls. It is not authorized yet. A passing
pilot permits preparation of a separate full 36-page development experiment;
it does not qualify a product, establish a probability, or permit inspection
of the separately frozen prospective confirmation pages.
