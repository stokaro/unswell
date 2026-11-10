# Atomic-claim paired pilot: negative result

The nine-page paired pilot completed all 18 scheduled GPT-6-Astra medium calls
with valid schemas, exact source bindings, and no retry. The atomic ledger did
not improve the frozen positive-event result over free-form review. It also
failed the accepted-delivery and retained-control gates. **Retire this atomic
candidate; issue #349 remains open.**

| Measure | Free-form review | Atomic-claim ledger |
| --- | ---: | ---: |
| Complete source pages / valid calls | 9 / 9 | 9 / 9 |
| Frozen events: FULL / PARTIAL / MISSED | 13 / 2 / 0 | 13 / 2 / 0 |
| Delivered diagnoses | 137 | 134 |
| Accepted / uncertain / rejected | 80 / 38 / 19 | 80 / 36 / 18 |
| Accepted share of all delivered | 58.4% | 59.7% |
| Explicit controls hit / uncertain, of 21 | 2 / 5 | 2 / 4 |
| Explicit `retain` controls hit | 2 | 2 |

The two PARTIAL events are E006 and E010 in both arms. E006's diagnoses
challenge praise and repetition but leave several unsupported outcome claims
unaddressed. E010's diagnoses challenge unlimited size and ease but do not
fully address the universal text-dataset scope and later caveats. Both arms
challenged the same retained warnings: C004, a slow symbolic branch, and C007,
the camera's unique-hostname instruction. Those are reviewed control hits,
not mere byte-span overlaps.

The predeclared directional gate required **at least two additional FULL
events**, **at least 85% accepted diagnoses**, and **zero explicit `retain`
control hits**. The atomic arm achieved zero additional FULL events, 59.7%
accepted delivery, and two retained-control hits. It therefore cannot
advance. The unchanged engine's preceding baseline found no source-specific
FULL diagnosis for these 15 frozen events; this paired model result is a
separate research observation, not an integrated engine improvement.

The sample is cue-enriched and purposive. It contains nine complete public
README pages, 889 eligible units, 15 source-only positive events on seven
repositories, and 21 explicit controls across all nine. The root Codex
assistant was the actual source-only and post-run reviewer under ADR 0041;
there was no independent human annotation. `uncertain` diagnoses remain in
the delivery denominator and are not counted as accepted. Two pages with no
positive event remain in the denominator. Unlabeled text is unknown. The
diagnosis-only arms made no replacement edits, so suggestion safety was not
measured. These figures do not establish prevalence, general precision, or
qualification of #349's original 123-event cohorts.

The exact packet and raw execution are preserved in [`model-packet/`](model-packet/)
and [`model-run-v1/`](model-run-v1/). The run includes the 18 responses,
traces, stderr files, bound original-byte spans, complete assistant review,
and machine-scored result. `review.json` records every delivered diagnosis,
event, and explicit control decision. The committed packet was frozen before
the first call. The run took 2,216.303 seconds and reported 418,504 input and
104,444 output tokens across the existing subscription route.

To verify the packet, all raw responses against their traces and original
source spans, and the reviewed grade from the frozen scorer:

```sh
python3 research/reviews/2026-11-10-atomic-claim-pilot/verify_model_result.py
```

Key SHA-256 identities:

- Frozen packet: `e20e9864e9734b7ea89b1845f2bc62b47b87e297bd146ae0935a7fa9ab0be2de`
- Source-only reference: `9550fd43d25463c5e1d24d8e7d4806a5698321d40d59fdd4501053bef4258d39`
- Run journal: `3870a06e7b1f0209307ee85cd55e98b86c8f4c5cd4bf8666c8ea49ed449216b6`
- Assistant review: `10270451f95b0432308a7e32225a28e0e38bef2b66a48508e588e4e97487cfa8`
- Scored result: `34592931c752988b12845f45c40628bcc9b597b103c0bb8894dde278988073b5`

The next #349 candidate needs a source-specific reason to expect fewer
optional diagnoses while retaining all distinct material burdens. It should
be screened against the existing frozen cohorts and explicit controls before
another model run. This negative pilot does not justify lowering thresholds,
changing labels, or promoting either model arm into the library.
