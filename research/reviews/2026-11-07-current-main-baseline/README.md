# Exact-main baseline and span-overlap ceiling for #349

This exposed, offline [#434](https://github.com/stokaro/unswell/issues/434)
follow-up replayed the unchanged public engine from merged commit
`5dd76191740c3d2f003b6ec90a9b381ccf1b772a` on the 36 original #349
development pages. The [protocol](PROTOCOL.md) and [freeze](freeze.json)
identify the source archives, original 123 reference events, prior comparison
reports, analysis code, and exact CLI binary. An exploratory replay had already
shown a changed finding stream; this is a reproducible exposed-data audit, not
an independent confirmation or a new model run. No reserved prospective page
was opened.

The [run ledger](runs.json), complete [technical](technical.json.gz) and
[strict](strict.json.gz) JSON reports, and [per-event analysis](result.json)
are retained. All 36 source hashes matched the frozen archives in both
profiles. Each report was complete, with zero operational errors and zero
skipped rules. The independent [analysis](analyze.py) verifies every source
identity and finding byte span, all 123 labels, report hashes, and the prior
comparison inputs. Its [negative test](test_analyze.py) rejects an altered
event-ceiling result.

| Profile | Current findings | Prior findings | Retained locations | Removed | Added |
| --- | ---: | ---: | ---: | ---: | ---: |
| Technical | 155 | 165 | 135 | 30 | 20 |
| Strict | 176 | 186 | 156 | 30 | 20 |

The 30 removed locations all belonged to
`syntax.paired-contrast-density`. The 20 additions came from seven later rule
families, including `filler.evaluative-closure` (9),
`filler.instruction-scaffolding` (3), and `filler.unscoped-assurance` (3).
All 21 locations that previously witnessed a fully credited event still have
a finding of the same rule at the same primary byte range. This is a location
continuity check; changing rule versions, messages, or surrounding findings
may change editorial judgments, so old accepted labels are not transferred.

The table below is a **generous upper bound based only on byte-range
intersection**, identical for technical and strict. It counts an event as
*possibly* full when every frozen target overlaps at least one delivered
primary or related finding range. Some overlaps address a different defect;
actual full-event recall can only be lower after source-context review.

| Complete cohort | Possible full by span | Required full for 80% | Events |
| --- | ---: | ---: | ---: |
| Whole page | 29 | 46 | 57 |
| Contextual | 13 | 37 | 46 |
| Historical confirmation | 4 | 16 | 20 |

Neither profile can reach #349's unchanged 80% full-event threshold with its
current delivered spans, even if every overlap is granted credit. No
accepted-finding fraction or suggestion-safety rate is asserted here: the 155
current technical findings and 176 strict findings have not received a fresh
complete-stream editorial review. The upper bound establishes a source-local
coverage gap, not the quality or cause of any individual finding. It does not
qualify the current engine, a new operator, or AI-origin detection.

The result JSON SHA-256 is
`bcefed9677464077490d0839f59ac5255e1f7d8e1309c6df4fb0b3f0533d9692`.
The compressed technical and strict report SHA-256 values are
`d45eaaaaf601cf66ee54ca649937c463f650b192d0a13f9de6807b4831e23aa1`
and `2be56f18dbc86f497f9c35b7520533d5fee49e5271cf49be0ab34c80cb3bd861`.

Verify this published record from the repository root without a model or
external service:

```sh
python3 -B research/reviews/2026-11-07-current-main-baseline/analyze.py \
  --run research/reviews/2026-11-07-current-main-baseline \
  --output research/reviews/2026-11-07-current-main-baseline/result.json --check
python3 -B -m unittest research/reviews/2026-11-07-current-main-baseline/test_analyze.py
```

This audit supports the #434 decision to keep searching for a genuinely new
source-local mechanism on exposed development pages, with a frozen hypothesis
and complete delivered-stream review before any admission claim. The two
short potential misses in the earlier
[external-control pass](../../methods/audits/2026-11-07-external-controls/README.md)
do not by themselves explain the much larger complete-page gap. Issue #349
remains open with its original labels and gates.

The subsequent [complete finding and event review](../2026-11-07-current-main-editorial-review/README.md)
adjudicates the delivered stream without treating span overlap as editorial
credit.
