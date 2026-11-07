# Exact contextual unit slices

This offline #349 preparation binds all 4,549 existing paragraph units on the
36 frozen pages to their exact original UTF-8 source slices. The [protocol](PROTOCOL.md)
and [manifest](manifest.json) pin the original archive and every generated
packet. Source text, unit boundaries, protected ranges, and reference labels
are unchanged. No model was called and no detector rule or response was changed.
The manifest SHA-256 is
`c7dcde652652bbbdbd3cb26c038308ebfd843b24f811d1d505915e1bdb8485ef`.

The earlier invalid quotation is reproduced at `page-019`, unit `u0038`:
the full document contains `**Deployment reports**`, while that unit starts
with `Deployment reports**`. The original binder rejects the former quotation
for that unit and accepts the latter at its original byte span. The new
`raw` field exposes that boundary directly to a future reviewer. Unit-local
`raw_sha256` values make accidental slice drift visible in packet reviews.

The checker verifies the archived request hashes, all source hashes, all
unit byte ranges, the generated packet hashes, and the complete 36-page file
set. Nine tests include changed-archive, changed-manifest, changed-raw,
shifted-span, protected-region, split-UTF-8, repeated-quotation, and
historical-boundary cases.

Run the offline check:

```sh
make check-contextual-unit-slices
```

To materialize the pinned packets in a new local directory:

```sh
python3 -B research/reviews/2026-11-07-contextual-unit-slices/unit_slices.py \
  --build /tmp/unswell-contextual-unit-slices
```

This is a source-map capability, not a semantic-diagnosis result. #349 still
needs a separately frozen contextual hypothesis, prompt, model budget, full
delivered-stream review, and untouched prospective confirmation under its
existing 80% FULL-event recall and 85% accepted-delivery gates. The new packet
does not authorize a retry of the stopped September run.
