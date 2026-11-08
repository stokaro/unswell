# #349 source-role admission hypothesis

Status: exposed offline development, November 9, 2026. Base commit:
`c4e9b4df241e339d67cb4309a8193793b09996ab`. The previous
[source-only packet](../2026-11-08-source-role-rescue/README.md) was frozen and
reviewed before this selector was written. The assistant has seen that packet,
the archived six-stream bank, and its grades. Neither the packet check nor an
archive replay can establish blind transfer or product qualification.

## Mechanism

The candidate operates only on existing `empty_framing/authorial_narration`
proposals that the frozen leave-one-repository-out operator policy would
otherwise withhold. It reads the exact target bytes and surrounding source
paragraph from the unchanged source snapshot. It returns `admit`, `withhold`,
or `abstain`; abstention never silently passes a gate. It must not read the
proposal's archived editorial grade or change a proposal's target span.

The source-role rule is conservative: retain explicit definitions, technical
concessions, safety qualifications, concrete instructions, links, code paths,
versions, and distinct comparative claims. Abstain on deliberate author voice.
Admit only generic document-outcome wishes, navigational announcements,
authorial process, or deictic guidance without a concrete action. Context
guards apply before admission cues. This is a bounded heuristic, not a
semantic entailment engine or an authorship classifier. Any selected
diagnosis remains diagnosis-only; archived edit suggestions cannot be
silently reused because five have unresolved safety reviews.

## Fixed evaluations

1. Run source-role unit cases from the prior four complete pages. The three
   narrow positives must be admitted, seven controls withheld, and the one
   unresolved author-voice case abstained. This is a development contract,
   not a held-out precision or recall estimate. Exact byte offsets and source
   hashes remain those in `2026-11-08-source-role-rescue/review.json` (SHA-256
   `25a680375a6ca26faa64b0c359ea5c642feb1dc698f0fb09af9291ce80dafba4`).
2. Apply the fixed selector without changes to the archived six-stream bank's
   excluded authorial-narration aliases. Preserve the original 36 pages,
   123 events, both policies, and the archived grade denominator. Compare
   against the 13-fold operator filter: whole 45/57 FULL and 1443/1700
   accepted, context 38/46 FULL and 389/413 accepted, historical confirmation
   17/20 FULL and 190/198 accepted. The first replay is exposed arithmetic;
   do not tune the selector on its result.
3. An archive replay advances only if all complete cohorts reach at least
   80% FULL and 85% accepted delivered diagnoses, with no unsafe supplied edit
   in the actual emitted form. It still cannot qualify #349 because the bank
   is cached model output, not a standalone complete generation run. A
   separate newly frozen page-disjoint confirmation and full end-to-end
   generation are required before product adoption. If the replay fails,
   record the negative result and retire this exact selector.

The old bank hashes and original labels are pinned in
[`audit.py`](../2026-11-08-operator-triage/audit.py). The selection code may
read `alias-metadata.json`, `cores.json`, `originals.json`, the original
selection alias list, and the already-published per-repository exclusion
   policy (SHA-256
   `22493865a6221276c08d945057829b9e455eccb50e0659d54f6df51c8d3232d8`).
   Only a separate scorer may read `reviews.json` and complete-event
witnesses. No model call, source rewrite, new label, or runtime rule is part of
this experiment.
